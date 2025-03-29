package handler

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// CreateContact implements createContact operation.
//
// Create a new contact record.
//
// POST /contact
func (h *Handler) CreateContact(ctx context.Context, req *api.Contact) (*api.Contact, error) {
	log := h.log.With("handler", "CreateContact")

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Contact, error) {
		contactUUID := uuid.Must(uuid.NewV7())

		params, err := mapAPIContactToCreateParams(req, contactUUID)
		if err != nil {
			log.Error("failed to map contact", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid contact data"))
		}

		row, err := query.New(tx).CreateContact(ctx, params)
		if err != nil {
			log.Error("failed to create contact", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create contact"))
		}

		// Convert DB row -> API model
		out, err := mapDBContactToAPI(row)
		if err != nil {
			log.Error("failed to map DB contact to API", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map contact"))
		}

		return out, nil
	})
}

// DeleteContact implements deleteContact operation.
//
// Delete a contact record.
//
// DELETE /contact/{uuid}
func (h *Handler) DeleteContact(ctx context.Context, params api.DeleteContactParams) error {
	log := h.log.With("handler", "DeleteContact")

	contactUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid contact uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid contact UUID"))
	}

	err = query.New(h.dbp).DeleteContact(ctx, contactUUID)
	switch {
	case err == pgx.ErrNoRows:
		// Possibly return 404 if you want the caller to know it's missing
		log.Warn("no contact found to delete", "contact_uuid", contactUUID)
		return nil
	case err != nil:
		log.Error("failed to delete contact", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete contact"))
	}
	return nil
}

// GetContact implements getContact operation.
//
// Get contact details.
//
// GET /contact/{uuid}
func (h *Handler) GetContact(ctx context.Context, params api.GetContactParams) (*api.Contact, error) {
	log := h.log.With("handler", "GetContact")

	contactUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid contact uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid contact UUID"))
	}

	c, err := query.New(h.dbp).GetContact(ctx, contactUUID)
	if err == pgx.ErrNoRows {
		return nil, ErrWithCode(http.StatusNotFound, E("contact not found"))
	} else if err != nil {
		log.Error("failed to get contact", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get contact"))
	}

	out, err := mapDBContactToAPI(c)
	if err != nil {
		log.Error("failed to map DB contact to API", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map contact"))
	}

	return out, nil
}

// ListContacts implements listContacts operation.
//
// List all contacts.
//
// GET /contact
func (h *Handler) ListContacts(ctx context.Context) ([]api.Contact, error) {
	log := h.log.With("handler", "ListContacts")

	// For demonstration, we'll set an offset=0, limit=100.
	// In real usage, you might pass query parameters for pagination.
	rows, err := query.New(h.dbp).ListContacts(ctx, query.ListContactsParams{
		OffsetRecords: 0,
		LimitRecords:  100,
	})
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to list contacts", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list contacts"))
	}

	var results []api.Contact
	for _, row := range rows {
		apiC, err := mapDBContactToAPI(row)
		if err != nil {
			log.Error("failed to map DB contact to API", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map contact"))
		}
		results = append(results, *apiC)
	}

	return results, nil
}

// UpdateContact implements updateContact operation.
//
// Update contact details.
//
// PUT /contact/{uuid}
func (h *Handler) UpdateContact(ctx context.Context, req *api.Contact, params api.UpdateContactParams) (*api.Contact, error) {
	log := h.log.With("handler", "UpdateContact")

	contactUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid contact uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid contact UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.Contact, error) {
		// fetch existing contact
		existing, err := query.New(tx).GetContact(ctx, contactUUID)
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("contact not found"))
		} else if err != nil {
			log.Error("failed to get contact", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get contact"))
		}

		uParams, err := mapAPIContactToUpdateParams(req, existing)
		if err != nil {
			log.Error("failed to map contact for update", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid contact data for update"))
		}
		uParams.UUID = contactUUID

		err = query.New(tx).UpdateContact(ctx, uParams)
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("contact not found to update"))
		} else if err != nil {
			log.Error("failed to update contact", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update contact"))
		}

		updated, err := query.New(tx).GetContact(ctx, contactUUID)
		if err != nil {
			log.Error("failed to get updated contact", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated contact"))
		}

		apiC, err := mapDBContactToAPI(updated)
		if err != nil {
			log.Error("failed to map updated DB contact", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map contact"))
		}

		return apiC, nil
	})
}

/* ------------------------------------------------------------------
   Below are helper functions that map between the DB query layer and
   the api.Contact struct. We cover all fields so no code is skipped!
------------------------------------------------------------------ */

// mapDBContactToAPI transforms the DB row into a full `api.Contact`
// covering all fields from the table.
func mapDBContactToAPI(dbRow query.Contact) (*api.Contact, error) {
	out := &api.Contact{}

	// Field: UUID
	out.UUID = api.NewOptString(dbRow.UUID.String())

	// Field: UserUUID
	if dbRow.UserUUID != nil {
		out.UserUUID = api.NewOptString(dbRow.UserUUID.String())
	}

	// Field: InstanceUUID
	if dbRow.InstanceUuid != nil {
		out.InstanceUUID = api.NewOptString(dbRow.InstanceUuid.String())
	}

	// Field: Status
	if dbRow.Status.Valid {
		out.Status = api.NewOptString(dbRow.Status.String)
	}

	// Field: Names (json)
	if len(dbRow.Names) > 0 {
		var names api.ContactNames
		if err := json.Unmarshal(dbRow.Names, &names); err != nil {
			return nil, err
		}
		out.Names = &names
	}

	// Field: NamesSearch
	if dbRow.NamesSearch.Valid {
		out.NamesSearch = api.NewOptString(dbRow.NamesSearch.String)
	}

	// Field: Last
	if dbRow.Last.Valid {
		out.Last = api.NewOptString(dbRow.Last.String)
	}

	// Field: First
	if dbRow.First.Valid {
		out.First = api.NewOptString(dbRow.First.String)
	}

	// Field: Middle
	if dbRow.Middle.Valid {
		out.Middle = api.NewOptString(dbRow.Middle.String)
	}

	// Field: Birthday
	if dbRow.Birthday.Valid {
		out.Birthday = api.NewOptDateTime(dbRow.Birthday.Time)
	}

	// Field: BirthdayType
	if dbRow.BirthdayType.Valid {
		out.BirthdayType = api.NewOptString(dbRow.BirthdayType.String)
	}

	// Field: Salary
	if dbRow.Salary.Valid {
		out.Salary = api.NewOptString(dbRow.Salary.String)
	}

	// Field: SalaryData (json)
	if len(dbRow.SalaryData) > 0 {
		var s api.ContactSalaryData
		if err := json.Unmarshal(dbRow.SalaryData, &s); err != nil {
			return nil, err
		}
		out.SalaryData = &s
	}

	// Field: LastPositions (json)
	if len(dbRow.LastPositions) > 0 {
		var lp api.ContactLastPositions
		if err := json.Unmarshal(dbRow.LastPositions, &lp); err != nil {
			return nil, err
		}
		out.LastPositions = &lp
	}

	// Field: LastPositionID
	if dbRow.LastPositionID.Valid {
		out.LastPositionID = api.NewOptInt(int(dbRow.LastPositionID.Int32))
	}

	// Field: LastPositionCompanyID
	if dbRow.LastPositionCompanyID.Valid {
		out.LastPositionCompanyID = api.NewOptInt(int(dbRow.LastPositionCompanyID.Int32))
	}

	// Field: LastPositionCompanyName
	if dbRow.LastPositionCompanyName.Valid {
		out.LastPositionCompanyName = api.NewOptString(dbRow.LastPositionCompanyName.String)
	}

	// Field: LastPositionTitle
	if dbRow.LastPositionTitle.Valid {
		out.LastPositionTitle = api.NewOptString(dbRow.LastPositionTitle.String)
	}

	// Field: LastPositionStartDate
	if dbRow.LastPositionStartDate.Valid {
		out.LastPositionStartDate = api.NewOptDateTime(dbRow.LastPositionStartDate.Time)
	}

	// Field: LastPositionEndDate
	if dbRow.LastPositionEndDate.Valid {
		out.LastPositionEndDate = api.NewOptDateTime(dbRow.LastPositionEndDate.Time)
	}

	// Field: LastPositionEndNow
	if dbRow.LastPositionEndNow.Valid {
		out.LastPositionEndNow = api.NewOptBool(dbRow.LastPositionEndNow.Bool)
	}

	// Field: LastPositionDescription
	if dbRow.LastPositionDescription.Valid {
		out.LastPositionDescription = api.NewOptString(dbRow.LastPositionDescription.String)
	}

	// Field: NoteSearch
	if dbRow.NoteSearch.Valid {
		out.NoteSearch = api.NewOptString(dbRow.NoteSearch.String)
	}

	// Field: NoteKpiID (json or other type)
	if len(dbRow.NoteKpiID) > 0 {
		var nk api.ContactNoteKpiID
		if err := json.Unmarshal(dbRow.NoteKpiID, &nk); err != nil {
			return nil, err
		}
		out.NoteKpiID = &nk
	}

	// Field: Phones (json)
	if len(dbRow.Phones) > 0 {
		var ph api.ContactPhones
		if err := json.Unmarshal(dbRow.Phones, &ph); err != nil {
			return nil, err
		}
		out.Phones = &ph
	}

	// Field: PhoneSearch
	if dbRow.PhoneSearch.Valid {
		out.PhoneSearch = api.NewOptString(dbRow.PhoneSearch.String)
	}

	// Field: phone1 .. phone5 + type/country
	if dbRow.Phone1.Valid {
		out.Phone1 = api.NewOptString(dbRow.Phone1.String)
	}
	if dbRow.Phone1Type.Valid {
		out.Phone1Type = api.NewOptString(dbRow.Phone1Type.String)
	}
	if dbRow.Phone1Country.Valid {
		out.Phone1Country = api.NewOptString(dbRow.Phone1Country.String)
	}
	if dbRow.Phone2.Valid {
		out.Phone2 = api.NewOptString(dbRow.Phone2.String)
	}
	if dbRow.Phone2Type.Valid {
		out.Phone2Type = api.NewOptString(dbRow.Phone2Type.String)
	}
	if dbRow.Phone2Country.Valid {
		out.Phone2Country = api.NewOptString(dbRow.Phone2Country.String)
	}
	if dbRow.Phone3.Valid {
		out.Phone3 = api.NewOptString(dbRow.Phone3.String)
	}
	if dbRow.Phone3Type.Valid {
		out.Phone3Type = api.NewOptString(dbRow.Phone3Type.String)
	}
	if dbRow.Phone3Country.Valid {
		out.Phone3Country = api.NewOptString(dbRow.Phone3Country.String)
	}
	if dbRow.Phone4.Valid {
		out.Phone4 = api.NewOptString(dbRow.Phone4.String)
	}
	if dbRow.Phone4Type.Valid {
		out.Phone4Type = api.NewOptString(dbRow.Phone4Type.String)
	}
	if dbRow.Phone4Country.Valid {
		out.Phone4Country = api.NewOptString(dbRow.Phone4Country.String)
	}
	if dbRow.Phone5.Valid {
		out.Phone5 = api.NewOptString(dbRow.Phone5.String)
	}
	if dbRow.Phone5Type.Valid {
		out.Phone5Type = api.NewOptString(dbRow.Phone5Type.String)
	}
	if dbRow.Phone5Country.Valid {
		out.Phone5Country = api.NewOptString(dbRow.Phone5Country.String)
	}

	// Field: Emails (json)
	if len(dbRow.Emails) > 0 {
		var em api.ContactEmails
		if err := json.Unmarshal(dbRow.Emails, &em); err != nil {
			return nil, err
		}
		out.Emails = &em
	}

	// Field: EmailSearch
	if dbRow.EmailSearch.Valid {
		out.EmailSearch = api.NewOptString(dbRow.EmailSearch.String)
	}
	// Field: email1..5 + type
	if dbRow.Email1.Valid {
		out.Email1 = api.NewOptString(dbRow.Email1.String)
	}
	if dbRow.Email1Type.Valid {
		out.Email1Type = api.NewOptString(dbRow.Email1Type.String)
	}
	if dbRow.Email2.Valid {
		out.Email2 = api.NewOptString(dbRow.Email2.String)
	}
	if dbRow.Email2Type.Valid {
		out.Email2Type = api.NewOptString(dbRow.Email2Type.String)
	}
	if dbRow.Email3.Valid {
		out.Email3 = api.NewOptString(dbRow.Email3.String)
	}
	if dbRow.Email3Type.Valid {
		out.Email3Type = api.NewOptString(dbRow.Email3Type.String)
	}
	if dbRow.Email4.Valid {
		out.Email4 = api.NewOptString(dbRow.Email4.String)
	}
	if dbRow.Email4Type.Valid {
		out.Email4Type = api.NewOptString(dbRow.Email4Type.String)
	}
	if dbRow.Email5.Valid {
		out.Email5 = api.NewOptString(dbRow.Email5.String)
	}
	if dbRow.Email5Type.Valid {
		out.Email5Type = api.NewOptString(dbRow.Email5Type.String)
	}

	// Field: Messengers (json)
	if len(dbRow.Messengers) > 0 {
		var m api.ContactMessengers
		if err := json.Unmarshal(dbRow.Messengers, &m); err != nil {
			return nil, err
		}
		out.Messengers = &m
	}

	// Field: MessengersSearch
	if dbRow.MessengersSearch.Valid {
		out.MessengersSearch = api.NewOptString(dbRow.MessengersSearch.String)
	}

	// Field: SkypeUuid
	if dbRow.SkypeUuid != nil {
		out.SkypeUUID = api.NewOptString(dbRow.SkypeUuid.String())
	}
	// Field: Skype
	if dbRow.Skype.Valid {
		out.Skype = api.NewOptString(dbRow.Skype.String)
	}

	// Field: WhatsappUuid
	if dbRow.WhatsappUuid != nil {
		out.WhatsappUUID = api.NewOptString(dbRow.WhatsappUuid.String())
	}
	// Field: Whatsapp
	if dbRow.Whatsapp.Valid {
		out.Whatsapp = api.NewOptString(dbRow.Whatsapp.String)
	}

	// Field: TelegramUuid
	if dbRow.TelegramUuid != nil {
		out.TelegramUUID = api.NewOptString(dbRow.TelegramUuid.String())
	}
	// Field: Telegram
	if dbRow.Telegram.Valid {
		out.Telegram = api.NewOptString(dbRow.Telegram.String)
	}

	// Field: WechatUuid
	if dbRow.WechatUuid != nil {
		out.WechatUUID = api.NewOptString(dbRow.WechatUuid.String())
	}
	// Field: Wechat
	if dbRow.Wechat.Valid {
		out.Wechat = api.NewOptString(dbRow.Wechat.String)
	}

	// Field: LineUuid
	if dbRow.LineUuid != nil {
		out.LineUUID = api.NewOptString(dbRow.LineUuid.String())
	}
	// Field: Line
	if dbRow.Line.Valid {
		out.Line = api.NewOptString(dbRow.Line.String)
	}

	// Field: Socials (json)
	if len(dbRow.Socials) > 0 {
		var s api.ContactSocials
		if err := json.Unmarshal(dbRow.Socials, &s); err != nil {
			return nil, err
		}
		out.Socials = &s
	}

	// Field: SocialsSearch
	if dbRow.SocialsSearch.Valid {
		out.SocialsSearch = api.NewOptString(dbRow.SocialsSearch.String)
	}

	// Field: LinkedinUuid
	if dbRow.LinkedinUuid != nil {
		out.LinkedinUUID = api.NewOptString(dbRow.LinkedinUuid.String())
	}
	// Field: LinkedinUrl
	if dbRow.LinkedinUrl.Valid {
		out.LinkedinURL = api.NewOptString(dbRow.LinkedinUrl.String)
	}

	// Field: FacebookUuid
	if dbRow.FacebookUuid != nil {
		out.FacebookUUID = api.NewOptString(dbRow.FacebookUuid.String())
	}
	// Field: FacebookUrl
	if dbRow.FacebookUrl.Valid {
		out.FacebookURL = api.NewOptString(dbRow.FacebookUrl.String)
	}

	// Field: TwitterUuid
	if dbRow.TwitterUuid != nil {
		out.TwitterUUID = api.NewOptString(dbRow.TwitterUuid.String())
	}
	// Field: TwitterUrl
	if dbRow.TwitterUrl.Valid {
		out.TwitterURL = api.NewOptString(dbRow.TwitterUrl.String)
	}

	// Field: GithubUuid
	if dbRow.GithubUuid != nil {
		out.GithubUUID = api.NewOptString(dbRow.GithubUuid.String())
	}
	// Field: GithubUrl
	if dbRow.GithubUrl.Valid {
		out.GithubURL = api.NewOptString(dbRow.GithubUrl.String)
	}

	// Field: VkUuid
	if dbRow.VkUuid != nil {
		out.VkUUID = api.NewOptString(dbRow.VkUuid.String())
	}
	// Field: VkUrl
	if dbRow.VkUrl.Valid {
		out.VkURL = api.NewOptString(dbRow.VkUrl.String)
	}

	// Field: OdnoUuid
	if dbRow.OdnoUuid != nil {
		out.OdnoUUID = api.NewOptString(dbRow.OdnoUuid.String())
	}
	// Field: OdnoUrl
	if dbRow.OdnoUrl.Valid {
		out.OdnoURL = api.NewOptString(dbRow.OdnoUrl.String)
	}

	// Field: HhruUuid
	if dbRow.HhruUuid != nil {
		out.HhruUUID = api.NewOptString(dbRow.HhruUuid.String())
	}
	// Field: HhruUrl
	if dbRow.HhruUrl.Valid {
		out.HhruURL = api.NewOptString(dbRow.HhruUrl.String)
	}

	// Field: HabrUuid
	if dbRow.HabrUuid != nil {
		out.HabrUUID = api.NewOptString(dbRow.HabrUuid.String())
	}
	// Field: HabrUrl
	if dbRow.HabrUrl.Valid {
		out.HabrURL = api.NewOptString(dbRow.HabrUrl.String)
	}

	// Field: MoikrugUuid
	if dbRow.MoikrugUuid != nil {
		out.MoikrugUUID = api.NewOptString(dbRow.MoikrugUuid.String())
	}
	// Field: MoikrugUrl
	if dbRow.MoikrugUrl.Valid {
		out.MoikrugURL = api.NewOptString(dbRow.MoikrugUrl.String)
	}

	// Field: InstagramUuid
	if dbRow.InstagramUuid != nil {
		out.InstagramUUID = api.NewOptString(dbRow.InstagramUuid.String())
	}
	// Field: InstagramUrl
	if dbRow.InstagramUrl.Valid {
		out.InstagramURL = api.NewOptString(dbRow.InstagramUrl.String)
	}

	// Field: Social1Uuid
	if dbRow.Social1Uuid != nil {
		out.Social1UUID = api.NewOptString(dbRow.Social1Uuid.String())
	}
	// Field: Social1Url
	if dbRow.Social1Url.Valid {
		out.Social1URL = api.NewOptString(dbRow.Social1Url.String)
	}
	// Field: Social1Type
	if dbRow.Social1Type.Valid {
		out.Social1Type = api.NewOptString(dbRow.Social1Type.String)
	}

	// Field: Social2Uuid
	if dbRow.Social2Uuid != nil {
		out.Social2UUID = api.NewOptString(dbRow.Social2Uuid.String())
	}
	// Field: Social2Url
	if dbRow.Social2Url.Valid {
		out.Social2URL = api.NewOptString(dbRow.Social2Url.String)
	}
	// Field: Social2Type
	if dbRow.Social2Type.Valid {
		out.Social2Type = api.NewOptString(dbRow.Social2Type.String)
	}

	// Field: Social3Uuid
	if dbRow.Social3Uuid != nil {
		out.Social3UUID = api.NewOptString(dbRow.Social3Uuid.String())
	}
	// Field: Social3Url
	if dbRow.Social3Url.Valid {
		out.Social3URL = api.NewOptString(dbRow.Social3Url.String)
	}
	// Field: Social3Type
	if dbRow.Social3Type.Valid {
		out.Social3Type = api.NewOptString(dbRow.Social3Type.String)
	}

	// Field: Social4Uuid
	if dbRow.Social4Uuid != nil {
		out.Social4UUID = api.NewOptString(dbRow.Social4Uuid.String())
	}
	// Field: Social4Url
	if dbRow.Social4Url.Valid {
		out.Social4URL = api.NewOptString(dbRow.Social4Url.String)
	}
	// Field: Social4Type
	if dbRow.Social4Type.Valid {
		out.Social4Type = api.NewOptString(dbRow.Social4Type.String)
	}

	// Field: Social5Uuid
	if dbRow.Social5Uuid != nil {
		out.Social5UUID = api.NewOptString(dbRow.Social5Uuid.String())
	}
	// Field: Social5Url
	if dbRow.Social5Url.Valid {
		out.Social5URL = api.NewOptString(dbRow.Social5Url.String)
	}
	// Field: Social5Type
	if dbRow.Social5Type.Valid {
		out.Social5Type = api.NewOptString(dbRow.Social5Type.String)
	}

	// Field: Social6Uuid
	if dbRow.Social6Uuid != nil {
		out.Social6UUID = api.NewOptString(dbRow.Social6Uuid.String())
	}
	// Field: Social6Url
	if dbRow.Social6Url.Valid {
		out.Social6URL = api.NewOptString(dbRow.Social6Url.String)
	}
	// Field: Social6Type
	if dbRow.Social6Type.Valid {
		out.Social6Type = api.NewOptString(dbRow.Social6Type.String)
	}

	// Field: Social7Uuid
	if dbRow.Social7Uuid != nil {
		out.Social7UUID = api.NewOptString(dbRow.Social7Uuid.String())
	}
	// Field: Social7Url
	if dbRow.Social7Url.Valid {
		out.Social7URL = api.NewOptString(dbRow.Social7Url.String)
	}
	// Field: Social7Type
	if dbRow.Social7Type.Valid {
		out.Social7Type = api.NewOptString(dbRow.Social7Type.String)
	}

	// Field: Social8Uuid
	if dbRow.Social8Uuid != nil {
		out.Social8UUID = api.NewOptString(dbRow.Social8Uuid.String())
	}
	// Field: Social8Url
	if dbRow.Social8Url.Valid {
		out.Social8URL = api.NewOptString(dbRow.Social8Url.String)
	}
	// Field: Social8Type
	if dbRow.Social8Type.Valid {
		out.Social8Type = api.NewOptString(dbRow.Social8Type.String)
	}

	// Field: Social9Uuid
	if dbRow.Social9Uuid != nil {
		out.Social9UUID = api.NewOptString(dbRow.Social9Uuid.String())
	}
	// Field: Social9Url
	if dbRow.Social9Url.Valid {
		out.Social9URL = api.NewOptString(dbRow.Social9Url.String)
	}
	// Field: Social9Type
	if dbRow.Social9Type.Valid {
		out.Social9Type = api.NewOptString(dbRow.Social9Type.String)
	}

	// Field: TrackingSource
	if dbRow.TrackingSource.Valid {
		out.TrackingSource = api.NewOptString(dbRow.TrackingSource.String)
	}
	// Field: TrackingSlug
	if dbRow.TrackingSlug.Valid {
		out.TrackingSlug = api.NewOptString(dbRow.TrackingSlug.String)
	}
	// Field: CachedImg
	if dbRow.CachedImg.Valid {
		out.CachedImg = api.NewOptString(dbRow.CachedImg.String)
	}
	// Field: CachedImgData (json, maybe)
	if len(dbRow.CachedImgData) > 0 {
		var ccd api.ContactCachedImgData
		if err := json.Unmarshal(dbRow.CachedImgData, &ccd); err != nil {
			return nil, err
		}
		out.CachedImgData = &ccd
	}

	// Field: Crawl (json?)
	if len(dbRow.Crawl) > 0 {
		var cc api.ContactCrawl
		if err := json.Unmarshal(dbRow.Crawl, &cc); err != nil {
			return nil, err
		}
		out.Crawl = &cc
	}

	// Field: DuplicateUserID
	if dbRow.DuplicateUserID.Valid {
		out.DuplicateUserID = api.NewOptString(dbRow.DuplicateUserID.String)
	}
	// Field: DuplicateAlternativeID
	if dbRow.DuplicateAlternativeID.Valid {
		out.DuplicateAlternativeID = api.NewOptString(dbRow.DuplicateAlternativeID.String)
	}
	// Field: DuplicateReportDate
	if dbRow.DuplicateReportDate.Valid {
		out.DuplicateReportDate = api.NewOptDateTime(dbRow.DuplicateReportDate.Time)
	}
	// Field: EntryDate
	if dbRow.EntryDate.Valid {
		out.EntryDate = api.NewOptDateTime(dbRow.EntryDate.Time)
	}
	// Field: EditDate
	if dbRow.EditDate.Valid {
		out.EditDate = api.NewOptDateTime(dbRow.EditDate.Time)
	}
	// Field: LastKpiEntryDate
	if dbRow.LastKpiEntryDate.Valid {
		out.LastKpiEntryDate = api.NewOptDateTime(dbRow.LastKpiEntryDate.Time)
	}

	return out, nil
}

// mapAPIContactToCreateParams transforms a full `api.Contact` into DB Insert params.
func mapAPIContactToCreateParams(apiC *api.Contact, newUUID uuid.UUID) (query.CreateContactParams, error) {
	p := query.CreateContactParams{
		UUID: newUUID,
	}

	// user_uuid
	if apiC.UserUUID.IsSet() && apiC.UserUUID.Value != "" {
		u, err := uuid.FromString(apiC.UserUUID.Value)
		if err == nil {
			p.UserUUID = &u
		}
	}

	// instance_uuid
	if apiC.InstanceUUID.IsSet() && apiC.InstanceUUID.Value != "" {
		u, err := uuid.FromString(apiC.InstanceUUID.Value)
		if err == nil {
			p.InstanceUuid = &u
		}
	}

	// status
	p.Status = pgtype.Text{
		String: apiC.Status.Or(""),
		Valid:  apiC.Status.IsSet(),
	}

	// names (json)
	if apiC.Names != nil {
		b, err := json.Marshal(apiC.Names)
		if err != nil {
			return p, err
		}
		p.Names = b
	} else {
		p.Names = nil
	}

	// names_search
	p.NamesSearch = pgtype.Text{
		String: apiC.NamesSearch.Or(""),
		Valid:  apiC.NamesSearch.IsSet(),
	}

	// last
	p.Last = pgtype.Text{
		String: apiC.Last.Or(""),
		Valid:  apiC.Last.IsSet(),
	}
	// first
	p.First = pgtype.Text{
		String: apiC.First.Or(""),
		Valid:  apiC.First.IsSet(),
	}
	// middle
	p.Middle = pgtype.Text{
		String: apiC.Middle.Or(""),
		Valid:  apiC.Middle.IsSet(),
	}
	// birthday
	if apiC.Birthday.IsSet() {
		p.Birthday = pgtype.Timestamptz{
			Time:  apiC.Birthday.Value,
			Valid: true,
		}
	} else {
		p.Birthday = pgtype.Timestamptz{}
	}
	// birthday_type
	p.BirthdayType = pgtype.Text{
		String: apiC.BirthdayType.Or(""),
		Valid:  apiC.BirthdayType.IsSet(),
	}
	// salary
	p.Salary = pgtype.Text{
		String: apiC.Salary.Or(""),
		Valid:  apiC.Salary.IsSet(),
	}
	// salary_data (json)
	if apiC.SalaryData != nil {
		b, err := json.Marshal(apiC.SalaryData)
		if err != nil {
			return p, err
		}
		p.SalaryData = b
	}
	// last_positions (json)
	if apiC.LastPositions != nil {
		b, err := json.Marshal(apiC.LastPositions)
		if err != nil {
			return p, err
		}
		p.LastPositions = b
	}
	// last_position_id
	if apiC.LastPositionID.IsSet() {
		p.LastPositionID = pgtype.Int4{
			Int32: int32(apiC.LastPositionID.Value),
			Valid: true,
		}
	}
	// last_position_company_id
	if apiC.LastPositionCompanyID.IsSet() {
		p.LastPositionCompanyID = pgtype.Int4{
			Int32: int32(apiC.LastPositionCompanyID.Value),
			Valid: true,
		}
	}
	// last_position_company_name
	p.LastPositionCompanyName = pgtype.Text{
		String: apiC.LastPositionCompanyName.Or(""),
		Valid:  apiC.LastPositionCompanyName.IsSet(),
	}
	// last_position_title
	p.LastPositionTitle = pgtype.Text{
		String: apiC.LastPositionTitle.Or(""),
		Valid:  apiC.LastPositionTitle.IsSet(),
	}
	// last_position_start_date
	if apiC.LastPositionStartDate.IsSet() {
		p.LastPositionStartDate = pgtype.Timestamptz{
			Time:  apiC.LastPositionStartDate.Value,
			Valid: true,
		}
	}
	// last_position_end_date
	if apiC.LastPositionEndDate.IsSet() {
		p.LastPositionEndDate = pgtype.Timestamptz{
			Time:  apiC.LastPositionEndDate.Value,
			Valid: true,
		}
	}
	// last_position_end_now
	if apiC.LastPositionEndNow.IsSet() {
		p.LastPositionEndNow = pgtype.Bool{
			Bool:  apiC.LastPositionEndNow.Value,
			Valid: true,
		}
	}
	// last_position_description
	p.LastPositionDescription = pgtype.Text{
		String: apiC.LastPositionDescription.Or(""),
		Valid:  apiC.LastPositionDescription.IsSet(),
	}
	// note_search
	p.NoteSearch = pgtype.Text{
		String: apiC.NoteSearch.Or(""),
		Valid:  apiC.NoteSearch.IsSet(),
	}
	// note_kpi_id (json)
	if apiC.NoteKpiID != nil {
		b, err := json.Marshal(apiC.NoteKpiID)
		if err != nil {
			return p, err
		}
		p.NoteKpiID = b
	}
	// phones (json)
	if apiC.Phones != nil {
		b, err := json.Marshal(apiC.Phones)
		if err != nil {
			return p, err
		}
		p.Phones = b
	}
	// phone_search
	p.PhoneSearch = pgtype.Text{
		String: apiC.PhoneSearch.Or(""),
		Valid:  apiC.PhoneSearch.IsSet(),
	}
	// phone1..phone5 + type/country
	p.Phone1 = textOpt(apiC.Phone1)
	p.Phone1Type = textOpt(apiC.Phone1Type)
	p.Phone1Country = textOpt(apiC.Phone1Country)
	p.Phone2 = textOpt(apiC.Phone2)
	p.Phone2Type = textOpt(apiC.Phone2Type)
	p.Phone2Country = textOpt(apiC.Phone2Country)
	p.Phone3 = textOpt(apiC.Phone3)
	p.Phone3Type = textOpt(apiC.Phone3Type)
	p.Phone3Country = textOpt(apiC.Phone3Country)
	p.Phone4 = textOpt(apiC.Phone4)
	p.Phone4Type = textOpt(apiC.Phone4Type)
	p.Phone4Country = textOpt(apiC.Phone4Country)
	p.Phone5 = textOpt(apiC.Phone5)
	p.Phone5Type = textOpt(apiC.Phone5Type)
	p.Phone5Country = textOpt(apiC.Phone5Country)

	// emails (json)
	if apiC.Emails != nil {
		b, err := json.Marshal(apiC.Emails)
		if err != nil {
			return p, err
		}
		p.Emails = b
	}
	// email_search
	p.EmailSearch = textOpt(apiC.EmailSearch)
	// email1..email5 + type
	p.Email1 = textOpt(apiC.Email1)
	p.Email1Type = textOpt(apiC.Email1Type)
	p.Email2 = textOpt(apiC.Email2)
	p.Email2Type = textOpt(apiC.Email2Type)
	p.Email3 = textOpt(apiC.Email3)
	p.Email3Type = textOpt(apiC.Email3Type)
	p.Email4 = textOpt(apiC.Email4)
	p.Email4Type = textOpt(apiC.Email4Type)
	p.Email5 = textOpt(apiC.Email5)
	p.Email5Type = textOpt(apiC.Email5Type)

	// messengers (json)
	if apiC.Messengers != nil {
		b, err := json.Marshal(apiC.Messengers)
		if err != nil {
			return p, err
		}
		p.Messengers = b
	}
	// messengers_search
	p.MessengersSearch = textOpt(apiC.MessengersSearch)

	// skype_uuid
	if apiC.SkypeUUID.IsSet() && apiC.SkypeUUID.Value != "" {
		su, err := uuid.FromString(apiC.SkypeUUID.Value)
		if err == nil {
			p.SkypeUuid = &su
		}
	}
	// skype
	p.Skype = textOpt(apiC.Skype)

	// whatsapp_uuid
	if apiC.WhatsappUUID.IsSet() && apiC.WhatsappUUID.Value != "" {
		wu, err := uuid.FromString(apiC.WhatsappUUID.Value)
		if err == nil {
			p.WhatsappUuid = &wu
		}
	}
	p.Whatsapp = textOpt(apiC.Whatsapp)

	// telegram_uuid
	if apiC.TelegramUUID.IsSet() && apiC.TelegramUUID.Value != "" {
		tu, err := uuid.FromString(apiC.TelegramUUID.Value)
		if err == nil {
			p.TelegramUuid = &tu
		}
	}
	p.Telegram = textOpt(apiC.Telegram)

	// wechat_uuid
	if apiC.WechatUUID.IsSet() && apiC.WechatUUID.Value != "" {
		wu, err := uuid.FromString(apiC.WechatUUID.Value)
		if err == nil {
			p.WechatUuid = &wu
		}
	}
	p.Wechat = textOpt(apiC.Wechat)

	// line_uuid
	if apiC.LineUUID.IsSet() && apiC.LineUUID.Value != "" {
		lu, err := uuid.FromString(apiC.LineUUID.Value)
		if err == nil {
			p.LineUuid = &lu
		}
	}
	p.Line = textOpt(apiC.Line)

	// socials (json)
	if apiC.Socials != nil {
		b, err := json.Marshal(apiC.Socials)
		if err != nil {
			return p, err
		}
		p.Socials = b
	}
	// socials_search
	p.SocialsSearch = textOpt(apiC.SocialsSearch)

	// linkedin_uuid
	if apiC.LinkedinUUID.IsSet() && apiC.LinkedinUUID.Value != "" {
		li, err := uuid.FromString(apiC.LinkedinUUID.Value)
		if err == nil {
			p.LinkedinUuid = &li
		}
	}
	p.LinkedinUrl = textOpt(apiC.LinkedinURL)

	// facebook_uuid
	if apiC.FacebookUUID.IsSet() && apiC.FacebookUUID.Value != "" {
		fu, err := uuid.FromString(apiC.FacebookUUID.Value)
		if err == nil {
			p.FacebookUuid = &fu
		}
	}
	p.FacebookUrl = textOpt(apiC.FacebookURL)

	// twitter_uuid
	if apiC.TwitterUUID.IsSet() && apiC.TwitterUUID.Value != "" {
		tu, err := uuid.FromString(apiC.TwitterUUID.Value)
		if err == nil {
			p.TwitterUuid = &tu
		}
	}
	p.TwitterUrl = textOpt(apiC.TwitterURL)

	// github_uuid
	if apiC.GithubUUID.IsSet() && apiC.GithubUUID.Value != "" {
		gu, err := uuid.FromString(apiC.GithubUUID.Value)
		if err == nil {
			p.GithubUuid = &gu
		}
	}
	p.GithubUrl = textOpt(apiC.GithubURL)

	// vk_uuid
	if apiC.VkUUID.IsSet() && apiC.VkUUID.Value != "" {
		vku, err := uuid.FromString(apiC.VkUUID.Value)
		if err == nil {
			p.VkUuid = &vku
		}
	}
	p.VkUrl = textOpt(apiC.VkURL)

	// odno_uuid
	if apiC.OdnoUUID.IsSet() && apiC.OdnoUUID.Value != "" {
		ou, err := uuid.FromString(apiC.OdnoUUID.Value)
		if err == nil {
			p.OdnoUuid = &ou
		}
	}
	p.OdnoUrl = textOpt(apiC.OdnoURL)

	// hhru_uuid
	if apiC.HhruUUID.IsSet() && apiC.HhruUUID.Value != "" {
		hu, err := uuid.FromString(apiC.HhruUUID.Value)
		if err == nil {
			p.HhruUuid = &hu
		}
	}
	p.HhruUrl = textOpt(apiC.HhruURL)

	// habr_uuid
	if apiC.HabrUUID.IsSet() && apiC.HabrUUID.Value != "" {
		habu, err := uuid.FromString(apiC.HabrUUID.Value)
		if err == nil {
			p.HabrUuid = &habu
		}
	}
	p.HabrUrl = textOpt(apiC.HabrURL)

	// moikrug_uuid
	if apiC.MoikrugUUID.IsSet() && apiC.MoikrugUUID.Value != "" {
		mu, err := uuid.FromString(apiC.MoikrugUUID.Value)
		if err == nil {
			p.MoikrugUuid = &mu
		}
	}
	p.MoikrugUrl = textOpt(apiC.MoikrugURL)

	// instagram_uuid
	if apiC.InstagramUUID.IsSet() && apiC.InstagramUUID.Value != "" {
		iu, err := uuid.FromString(apiC.InstagramUUID.Value)
		if err == nil {
			p.InstagramUuid = &iu
		}
	}
	p.InstagramUrl = textOpt(apiC.InstagramURL)

	// social1_uuid..social9_type
	p.Social1Uuid = parseUUID(apiC.Social1UUID)
	p.Social1Url = textOpt(apiC.Social1URL)
	p.Social1Type = textOpt(apiC.Social1Type)

	p.Social2Uuid = parseUUID(apiC.Social2UUID)
	p.Social2Url = textOpt(apiC.Social2URL)
	p.Social2Type = textOpt(apiC.Social2Type)

	p.Social3Uuid = parseUUID(apiC.Social3UUID)
	p.Social3Url = textOpt(apiC.Social3URL)
	p.Social3Type = textOpt(apiC.Social3Type)

	p.Social4Uuid = parseUUID(apiC.Social4UUID)
	p.Social4Url = textOpt(apiC.Social4URL)
	p.Social4Type = textOpt(apiC.Social4Type)

	p.Social5Uuid = parseUUID(apiC.Social5UUID)
	p.Social5Url = textOpt(apiC.Social5URL)
	p.Social5Type = textOpt(apiC.Social5Type)

	p.Social6Uuid = parseUUID(apiC.Social6UUID)
	p.Social6Url = textOpt(apiC.Social6URL)
	p.Social6Type = textOpt(apiC.Social6Type)

	p.Social7Uuid = parseUUID(apiC.Social7UUID)
	p.Social7Url = textOpt(apiC.Social7URL)
	p.Social7Type = textOpt(apiC.Social7Type)

	p.Social8Uuid = parseUUID(apiC.Social8UUID)
	p.Social8Url = textOpt(apiC.Social8URL)
	p.Social8Type = textOpt(apiC.Social8Type)

	p.Social9Uuid = parseUUID(apiC.Social9UUID)
	p.Social9Url = textOpt(apiC.Social9URL)
	p.Social9Type = textOpt(apiC.Social9Type)

	// tracking_source, tracking_slug
	p.TrackingSource = textOpt(apiC.TrackingSource)
	p.TrackingSlug = textOpt(apiC.TrackingSlug)

	// cached_img
	p.CachedImg = textOpt(apiC.CachedImg)
	// cached_img_data
	if apiC.CachedImgData != nil {
		b, err := json.Marshal(apiC.CachedImgData)
		if err != nil {
			return p, err
		}
		p.CachedImgData = b
	}
	// crawl
	if apiC.Crawl != nil {
		b, err := json.Marshal(apiC.Crawl)
		if err != nil {
			return p, err
		}
		p.Crawl = b
	}

	// duplicate_user_id
	p.DuplicateUserID = textOpt(apiC.DuplicateUserID)
	// duplicate_alternative_id
	p.DuplicateAlternativeID = textOpt(apiC.DuplicateAlternativeID)
	// duplicate_report_date
	if apiC.DuplicateReportDate.IsSet() {
		p.DuplicateReportDate = pgtype.Timestamptz{
			Time:  apiC.DuplicateReportDate.Value,
			Valid: true,
		}
	}
	// entry_date
	if apiC.EntryDate.IsSet() {
		p.EntryDate = pgtype.Timestamptz{
			Time:  apiC.EntryDate.Value,
			Valid: true,
		}
	}
	// edit_date
	if apiC.EditDate.IsSet() {
		p.EditDate = pgtype.Timestamptz{
			Time:  apiC.EditDate.Value,
			Valid: true,
		}
	}
	// last_kpi_entry_date
	if apiC.LastKpiEntryDate.IsSet() {
		p.LastKpiEntryDate = pgtype.Timestamptz{
			Time:  apiC.LastKpiEntryDate.Value,
			Valid: true,
		}
	}

	return p, nil
}

// mapAPIContactToUpdateParams merges the new data from `api.Contact` with the existing DB row.
func mapAPIContactToUpdateParams(apiC *api.Contact, existing query.Contact) (query.UpdateContactParams, error) {
	u := query.UpdateContactParams{}

	// We fill every field by either using the new input (if set) or the existing DB value.

	// user_uuid
	if apiC.UserUUID.IsSet() && apiC.UserUUID.Value != "" {
		uid, err := uuid.FromString(apiC.UserUUID.Value)
		if err == nil {
			u.UserUUID = pgNullableUUID(&uid)
		} else {
			u.UserUUID = pgNullableUUID(existing.UserUUID)
		}
	} else {
		u.UserUUID = pgNullableUUID(existing.UserUUID)
	}

	// instance_uuid
	if apiC.InstanceUUID.IsSet() && apiC.InstanceUUID.Value != "" {
		iu, err := uuid.FromString(apiC.InstanceUUID.Value)
		if err == nil {
			u.InstanceUuid = pgNullableUUID(&iu)
		} else {
			u.InstanceUuid = pgNullableUUID(existing.InstanceUuid)
		}
	} else {
		u.InstanceUuid = pgNullableUUID(existing.InstanceUuid)
	}

	u.Status = pickText(apiC.Status, existing.Status)
	// names
	if apiC.Names != nil {
		b, err := json.Marshal(apiC.Names)
		if err != nil {
			return u, err
		}
		u.Names = b
	} else {
		u.Names = existing.Names
	}
	u.NamesSearch = pickText(apiC.NamesSearch, existing.NamesSearch)
	u.Last = pickText(apiC.Last, existing.Last)
	u.First = pickText(apiC.First, existing.First)
	u.Middle = pickText(apiC.Middle, existing.Middle)
	u.Birthday = pickTimestamptz(apiC.Birthday, existing.Birthday)
	u.BirthdayType = pickText(apiC.BirthdayType, existing.BirthdayType)
	u.Salary = pickText(apiC.Salary, existing.Salary)

	// salary_data
	if apiC.SalaryData != nil {
		b, err := json.Marshal(apiC.SalaryData)
		if err != nil {
			return u, err
		}
		u.SalaryData = b
	} else {
		u.SalaryData = existing.SalaryData
	}
	// last_positions
	if apiC.LastPositions != nil {
		b, err := json.Marshal(apiC.LastPositions)
		if err != nil {
			return u, err
		}
		u.LastPositions = b
	} else {
		u.LastPositions = existing.LastPositions
	}

	if apiC.LastPositionID.IsSet() {
		u.LastPositionID = pgtype.Int4{
			Int32: int32(apiC.LastPositionID.Value),
			Valid: true,
		}
	} else {
		u.LastPositionID = existing.LastPositionID
	}
	if apiC.LastPositionCompanyID.IsSet() {
		u.LastPositionCompanyID = pgtype.Int4{
			Int32: int32(apiC.LastPositionCompanyID.Value),
			Valid: true,
		}
	} else {
		u.LastPositionCompanyID = existing.LastPositionCompanyID
	}
	u.LastPositionCompanyName = pickText(apiC.LastPositionCompanyName, existing.LastPositionCompanyName)
	u.LastPositionTitle = pickText(apiC.LastPositionTitle, existing.LastPositionTitle)
	u.LastPositionStartDate = pickTimestamptz(apiC.LastPositionStartDate, existing.LastPositionStartDate)
	u.LastPositionEndDate = pickTimestamptz(apiC.LastPositionEndDate, existing.LastPositionEndDate)
	u.LastPositionEndNow = pickBool(apiC.LastPositionEndNow, existing.LastPositionEndNow)
	u.LastPositionDescription = pickText(apiC.LastPositionDescription, existing.LastPositionDescription)
	u.NoteSearch = pickText(apiC.NoteSearch, existing.NoteSearch)

	// note_kpi_id
	if apiC.NoteKpiID != nil {
		b, err := json.Marshal(apiC.NoteKpiID)
		if err != nil {
			return u, err
		}
		u.NoteKpiID = b
	} else {
		u.NoteKpiID = existing.NoteKpiID
	}

	// phones
	if apiC.Phones != nil {
		b, err := json.Marshal(apiC.Phones)
		if err != nil {
			return u, err
		}
		u.Phones = b
	} else {
		u.Phones = existing.Phones
	}
	u.PhoneSearch = pickText(apiC.PhoneSearch, existing.PhoneSearch)
	u.Phone1 = pickText(apiC.Phone1, existing.Phone1)
	u.Phone1Type = pickText(apiC.Phone1Type, existing.Phone1Type)
	u.Phone1Country = pickText(apiC.Phone1Country, existing.Phone1Country)
	u.Phone2 = pickText(apiC.Phone2, existing.Phone2)
	u.Phone2Type = pickText(apiC.Phone2Type, existing.Phone2Type)
	u.Phone2Country = pickText(apiC.Phone2Country, existing.Phone2Country)
	u.Phone3 = pickText(apiC.Phone3, existing.Phone3)
	u.Phone3Type = pickText(apiC.Phone3Type, existing.Phone3Type)
	u.Phone3Country = pickText(apiC.Phone3Country, existing.Phone3Country)
	u.Phone4 = pickText(apiC.Phone4, existing.Phone4)
	u.Phone4Type = pickText(apiC.Phone4Type, existing.Phone4Type)
	u.Phone4Country = pickText(apiC.Phone4Country, existing.Phone4Country)
	u.Phone5 = pickText(apiC.Phone5, existing.Phone5)
	u.Phone5Type = pickText(apiC.Phone5Type, existing.Phone5Type)
	u.Phone5Country = pickText(apiC.Phone5Country, existing.Phone5Country)

	// emails
	if apiC.Emails != nil {
		b, err := json.Marshal(apiC.Emails)
		if err != nil {
			return u, err
		}
		u.Emails = b
	} else {
		u.Emails = existing.Emails
	}
	u.EmailSearch = pickText(apiC.EmailSearch, existing.EmailSearch)
	u.Email1 = pickText(apiC.Email1, existing.Email1)
	u.Email1Type = pickText(apiC.Email1Type, existing.Email1Type)
	u.Email2 = pickText(apiC.Email2, existing.Email2)
	u.Email2Type = pickText(apiC.Email2Type, existing.Email2Type)
	u.Email3 = pickText(apiC.Email3, existing.Email3)
	u.Email3Type = pickText(apiC.Email3Type, existing.Email3Type)
	u.Email4 = pickText(apiC.Email4, existing.Email4)
	u.Email4Type = pickText(apiC.Email4Type, existing.Email4Type)
	u.Email5 = pickText(apiC.Email5, existing.Email5)
	u.Email5Type = pickText(apiC.Email5Type, existing.Email5Type)

	// messengers
	if apiC.Messengers != nil {
		b, err := json.Marshal(apiC.Messengers)
		if err != nil {
			return u, err
		}
		u.Messengers = b
	} else {
		u.Messengers = existing.Messengers
	}
	u.MessengersSearch = pickText(apiC.MessengersSearch, existing.MessengersSearch)

	// skype_uuid
	u.SkypeUuid = mergeUUID(apiC.SkypeUUID, existing.SkypeUuid)
	u.Skype = pickText(apiC.Skype, existing.Skype)

	// whatsapp_uuid
	u.WhatsappUuid = mergeUUID(apiC.WhatsappUUID, existing.WhatsappUuid)
	u.Whatsapp = pickText(apiC.Whatsapp, existing.Whatsapp)

	// telegram_uuid
	u.TelegramUuid = mergeUUID(apiC.TelegramUUID, existing.TelegramUuid)
	u.Telegram = pickText(apiC.Telegram, existing.Telegram)

	// wechat_uuid
	u.WechatUuid = mergeUUID(apiC.WechatUUID, existing.WechatUuid)
	u.Wechat = pickText(apiC.Wechat, existing.Wechat)

	// line_uuid
	u.LineUuid = mergeUUID(apiC.LineUUID, existing.LineUuid)
	u.Line = pickText(apiC.Line, existing.Line)

	// socials
	if apiC.Socials != nil {
		b, err := json.Marshal(apiC.Socials)
		if err != nil {
			return u, err
		}
		u.Socials = b
	} else {
		u.Socials = existing.Socials
	}
	u.SocialsSearch = pickText(apiC.SocialsSearch, existing.SocialsSearch)

	// linkedin_uuid
	u.LinkedinUuid = mergeUUID(apiC.LinkedinUUID, existing.LinkedinUuid)
	u.LinkedinUrl = pickText(apiC.LinkedinURL, existing.LinkedinUrl)

	// facebook_uuid
	u.FacebookUuid = mergeUUID(apiC.FacebookUUID, existing.FacebookUuid)
	u.FacebookUrl = pickText(apiC.FacebookURL, existing.FacebookUrl)

	// twitter_uuid
	u.TwitterUuid = mergeUUID(apiC.TwitterUUID, existing.TwitterUuid)
	u.TwitterUrl = pickText(apiC.TwitterURL, existing.TwitterUrl)

	// github_uuid
	u.GithubUuid = mergeUUID(apiC.GithubUUID, existing.GithubUuid)
	u.GithubUrl = pickText(apiC.GithubURL, existing.GithubUrl)

	// vk_uuid
	u.VkUuid = mergeUUID(apiC.VkUUID, existing.VkUuid)
	u.VkUrl = pickText(apiC.VkURL, existing.VkUrl)

	// odno_uuid
	u.OdnoUuid = mergeUUID(apiC.OdnoUUID, existing.OdnoUuid)
	u.OdnoUrl = pickText(apiC.OdnoURL, existing.OdnoUrl)

	// hhru_uuid
	u.HhruUuid = mergeUUID(apiC.HhruUUID, existing.HhruUuid)
	u.HhruUrl = pickText(apiC.HhruURL, existing.HhruUrl)

	// habr_uuid
	u.HabrUuid = mergeUUID(apiC.HabrUUID, existing.HabrUuid)
	u.HabrUrl = pickText(apiC.HabrURL, existing.HabrUrl)

	// moikrug_uuid
	u.MoikrugUuid = mergeUUID(apiC.MoikrugUUID, existing.MoikrugUuid)
	u.MoikrugUrl = pickText(apiC.MoikrugURL, existing.MoikrugUrl)

	// instagram_uuid
	u.InstagramUuid = mergeUUID(apiC.InstagramUUID, existing.InstagramUuid)
	u.InstagramUrl = pickText(apiC.InstagramURL, existing.InstagramUrl)

	// social1..9
	u.Social1Uuid = mergeUUID(apiC.Social1UUID, existing.Social1Uuid)
	u.Social1Url = pickText(apiC.Social1URL, existing.Social1Url)
	u.Social1Type = pickText(apiC.Social1Type, existing.Social1Type)

	u.Social2Uuid = mergeUUID(apiC.Social2UUID, existing.Social2Uuid)
	u.Social2Url = pickText(apiC.Social2URL, existing.Social2Url)
	u.Social2Type = pickText(apiC.Social2Type, existing.Social2Type)

	u.Social3Uuid = mergeUUID(apiC.Social3UUID, existing.Social3Uuid)
	u.Social3Url = pickText(apiC.Social3URL, existing.Social3Url)
	u.Social3Type = pickText(apiC.Social3Type, existing.Social3Type)

	u.Social4Uuid = mergeUUID(apiC.Social4UUID, existing.Social4Uuid)
	u.Social4Url = pickText(apiC.Social4URL, existing.Social4Url)
	u.Social4Type = pickText(apiC.Social4Type, existing.Social4Type)

	u.Social5Uuid = mergeUUID(apiC.Social5UUID, existing.Social5Uuid)
	u.Social5Url = pickText(apiC.Social5URL, existing.Social5Url)
	u.Social5Type = pickText(apiC.Social5Type, existing.Social5Type)

	u.Social6Uuid = mergeUUID(apiC.Social6UUID, existing.Social6Uuid)
	u.Social6Url = pickText(apiC.Social6URL, existing.Social6Url)
	u.Social6Type = pickText(apiC.Social6Type, existing.Social6Type)

	u.Social7Uuid = mergeUUID(apiC.Social7UUID, existing.Social7Uuid)
	u.Social7Url = pickText(apiC.Social7URL, existing.Social7Url)
	u.Social7Type = pickText(apiC.Social7Type, existing.Social7Type)

	u.Social8Uuid = mergeUUID(apiC.Social8UUID, existing.Social8Uuid)
	u.Social8Url = pickText(apiC.Social8URL, existing.Social8Url)
	u.Social8Type = pickText(apiC.Social8Type, existing.Social8Type)

	u.Social9Uuid = mergeUUID(apiC.Social9UUID, existing.Social9Uuid)
	u.Social9Url = pickText(apiC.Social9URL, existing.Social9Url)
	u.Social9Type = pickText(apiC.Social9Type, existing.Social9Type)

	// tracking_source
	u.TrackingSource = pickText(apiC.TrackingSource, existing.TrackingSource)
	// tracking_slug
	u.TrackingSlug = pickText(apiC.TrackingSlug, existing.TrackingSlug)
	// cached_img
	u.CachedImg = pickText(apiC.CachedImg, existing.CachedImg)

	// cached_img_data
	if apiC.CachedImgData != nil {
		b, err := json.Marshal(apiC.CachedImgData)
		if err != nil {
			return u, err
		}
		u.CachedImgData = b
	} else {
		u.CachedImgData = existing.CachedImgData
	}

	// crawl
	if apiC.Crawl != nil {
		b, err := json.Marshal(apiC.Crawl)
		if err != nil {
			return u, err
		}
		u.Crawl = b
	} else {
		u.Crawl = existing.Crawl
	}

	u.DuplicateUserID = pickText(apiC.DuplicateUserID, existing.DuplicateUserID)
	u.DuplicateAlternativeID = pickText(apiC.DuplicateAlternativeID, existing.DuplicateAlternativeID)

	u.DuplicateReportDate = pickTimestamptz(apiC.DuplicateReportDate, existing.DuplicateReportDate)
	u.EntryDate = pickTimestamptz(apiC.EntryDate, existing.EntryDate)
	u.EditDate = pickTimestamptz(apiC.EditDate, existing.EditDate)
	u.LastKpiEntryDate = pickTimestamptz(apiC.LastKpiEntryDate, existing.LastKpiEntryDate)

	return u, nil
}

/* ---------------------------------------------------------
   Small helper methods for bridging pgtype.* with Opt*
--------------------------------------------------------- */

// merges an incoming OptString with an existing *uuid.UUID
func mergeUUID(in api.OptString, existing *uuid.UUID) *uuid.UUID {
	if in.IsSet() && in.Value != "" {
		u, err := uuid.FromString(in.Value)
		if err == nil {
			return &u
		}
	}
	return existing
}

// parseUUID is used in "create" logic
func parseUUID(in api.OptString) *uuid.UUID {
	if in.IsSet() && in.Value != "" {
		uu, err := uuid.FromString(in.Value)
		if err == nil {
			return &uu
		}
	}
	return nil
}

// pickText picks the new text if set, else old value
func pickText(in api.OptString, old pgtype.Text) pgtype.Text {
	if in.IsSet() {
		return pgtype.Text{
			String: in.Value,
			Valid:  true,
		}
	}
	return old
}

// pickTimestamptz picks the new time if set, else old
func pickTimestamptz(in api.OptDateTime, old pgtype.Timestamptz) pgtype.Timestamptz {
	if in.IsSet() {
		return pgtype.Timestamptz{
			Time:  in.Value,
			Valid: true,
		}
	}
	return old
}

// pickBool picks new bool if set, else old
func pickBool(in api.OptBool, old pgtype.Bool) pgtype.Bool {
	if in.IsSet() {
		return pgtype.Bool{
			Bool:  in.Value,
			Valid: true,
		}
	}
	return old
}

// textOpt is used for "create" logic
func textOpt(in api.OptString) pgtype.Text {
	if in.IsSet() {
		return pgtype.Text{String: in.Value, Valid: true}
	}
	return pgtype.Text{}
}

// pgNullableUUID is used to transform a *uuid.UUID -> *uuid.UUID for sqlc
func pgNullableUUID(u *uuid.UUID) *uuid.UUID {
	if u == nil {
		return nil
	}
	return u
}
