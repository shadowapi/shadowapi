package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	"log/slog"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// handleGmailToken persists a newly issued Gmail OAuth2 token, links it with the datasource
// that initiated the flow and stores an oauth2_subject row so that we know which user
// granted which client access.
//
// Tables layout (relevant columns):
//
//	oauth2_token   – uuid (PK), client_uuid, user_uuid, token (jsonb)
//	oauth2_subject – uuid (PK), user_uuid, token_uuid
//	datasource     – settings -> oauth2_token_uuid
//
// The oauth2_subject table does **not** have a client_uuid column – it references token_uuid.
// sqlc generated query CreateOauth2Subject is therefore unusable (it inserts client_uuid).
// We do a raw `INSERT` instead.
func (h *Handler) handleGmailToken(
	ctx context.Context,
	log *slog.Logger,
	token *oauth2.Token,
	stateQuery url.Values,
	clientID string,
) (*api.OAuth2ClientCallbackFound, error) {

	log.Info("0. handleGmailToken", "client_id", clientID, "state_query", stateQuery)

	// 1) Marshal token
	tokenData, err := json.Marshal(token)
	if err != nil {
		log.Error("1. failed to marshal token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed marshal token"))
	}

	// 2) Extract datasource_uuid from state
	dsID := stateQuery.Get("datasource_uuid")
	if dsID == "" {
		log.Error("2.1 missing datasource_uuid in state query")
		return nil, ErrWithCode(http.StatusBadRequest, E("missing datasource_uuid parameter"))
	}
	log.Info("2.1 handleGmailToken", "dsID", dsID)

	goDSUUID, err := uuid.FromString(dsID)
	if err != nil {
		log.Error("2.2 invalid datasource_uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource_uuid parameter"))
	}

	// 3) Resolve client & user UUIDs
	pgClientUUID, err := converter.ConvertStringToPgUUID(clientID)
	if err != nil {
		log.Error("3. failed to convert client UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid client UUID"))
	}

	// user_uuid precedence: explicit in state → datasource.owner
	var pgUserUUID pgtype.UUID
	if u := stateQuery.Get("user_uuid"); u != "" {
		if goUser, err := uuid.FromString(u); err == nil {
			pgUserUUID = converter.UuidToPgUUID(goUser)
		}
	}

	// fallback – owner of datasource
	if !pgUserUUID.Valid {
		// light read using readonly pool outside tx
		dsRow, err := query.New(h.dbp).GetDatasource(ctx, converter.UuidToPgUUID(goDSUUID))
		if err != nil {
			log.Error("3.1 cannot fetch datasource to resolve user", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to resolve user"))
		}
		pgUserUUID = converter.UuidPtrToPgUUID(dsRow.Datasource.UserUUID)
	}

	if !pgUserUUID.Valid {
		log.Error("3.2 resolved user_uuid is still NULL – can't proceed")
		return nil, ErrWithCode(http.StatusInternalServerError, E("user_uuid not resolved"))
	}

	// 4) Start transaction
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.OAuth2ClientCallbackFound, error) {
		q := query.New(tx)

		// 4.1) Remove previous token for this client (single‑token semantics)
		_ = q.DeleteOauth2TokenByClientUUID(ctx, pgClientUUID)

		// 4.2) Create new oauth2_token row
		goTokenUUID := uuid.Must(uuid.NewV7())
		pgTokenUUID := converter.UuidToPgUUID(goTokenUUID)

		log.Info("4.1 handleGmailToken", "query.CreateOauth2TokenParams", query.CreateOauth2TokenParams{
			UUID:       pgTokenUUID,
			ClientUuid: pgClientUUID,
			UserUUID:   pgUserUUID,
			Token:      tokenData,
		})

		if _, err := q.CreateOauth2Token(ctx, query.CreateOauth2TokenParams{
			UUID:       pgTokenUUID,
			ClientUuid: pgClientUUID,
			UserUUID:   pgUserUUID,
			Token:      tokenData,
		}); err != nil {
			log.Error("4.2 create oauth2 token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("can't add oauth2 token"))
		}

		// 4.3) Link datasource → token in settings.oauth2_token_uuid
		if err := q.LinkDatasourceWithToken(ctx, query.LinkDatasourceWithTokenParams{
			UUID:            goDSUUID,
			OAuth2TokenUUID: goTokenUUID.String(),
		}); err != nil {
			log.Error("4.3 link datasource with token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to link datasource with token"))
		}

		// !!
		// 4.4) Insert oauth2_subject referencing **token_uuid** (raw SQL because sqlc file is wrong)
		goSubjectUUID := uuid.Must(uuid.NewV7())
		pgSubjectUUID := converter.UuidToPgUUID(goSubjectUUID)
		if _, err := q.CreateOauth2Subject(ctx, query.CreateOauth2SubjectParams{
			UUID:      pgSubjectUUID,
			UserUUID:  pgUserUUID,
			TokenUuid: pgTokenUUID,
		}); err != nil {
			log.Error("4.4 failed to add oauth2 subject", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("can't add oauth2 subject"))
		}

		// TODO: schedule token refresh via worker/pipeline – skipped per current scope

		// 4.5) Build redirect location back to datasource page
		location, err := url.Parse("/datasources/" + dsID)
		if err != nil {
			log.Error("4.5 can't parse location", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("can't parse location"))
		}

		return &api.OAuth2ClientCallbackFound{
			Location: api.OptURI{Value: *location, Set: true},
		}, nil
	})
}

/*
func (h *Handler) handleGmailTokenOld(
	ctx context.Context,
	log *slog.Logger,
	token *oauth2.Token,
	stateQuery url.Values,
	clientID string,
) (*api.OAuth2ClientCallbackFound, error) {
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))

		var tokenData []byte

		tokenData, err := json.Marshal(token)
		if err != nil {
			log.Error("marshal api token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed marshal api token"))
		}

		datasourceUUID := uuid.FromStringOrNil(stateQuery.Get("datasource_uuid"))
		if datasourceUUID.IsNil() {
			log.Error("missing datasource_uuid parameter")
			return nil, ErrWithCode(http.StatusBadRequest, E("missing datasource_uuid parameter"))
		}

		return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.OAuth2ClientCallbackFound, error) {
			db := query.New(tx)


			err := db.DeleteOauth2TokenByClientID(ctx, clientID)
			if err != nil && err != pgx.ErrNoRows {
				log.Error("can't delete oauth2 token by client id", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("can't delete oauth2 token by client id"))
			}

			tokenUUID := uuid.Must(uuid.NewV7())
			err = db.AddOauth2Token(ctx, query.AddOauth2TokenParams{
				UUID:     tokenUUID,
				ClientID: clientID,
				Token:    tokenData,
			})
			if err != nil {
				log.Error("can't add oauth2 token", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("can't add oauth2 token"))
			}

			err = db.LinkDatasourceWithToken(ctx, query.LinkDatasourceWithTokenParams{
				UUID:            datasourceUUID,
				OAuth2TokenUUID: tokenUUID.String(),
			})
			if err != nil {
				log.Error("link datasource with token", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("link datasource with token"))
			}

			// TODO @reactima token needs to be sent and processed as a message
			//log.Info("add token refresh job to worker pool", "token_uuid", tokenUUID, "expiry", token.Expiry)
			//if err = h.wbr.ScheduleRefresh(ctx, tokenUUID, token.Expiry); err != nil {
			//	log.Error("fail add job to worker pool", "error", err)
			//	return nil, ErrWithCode(http.StatusInternalServerError, E("fail add job to worker pool"))
			//}

			location, err := url.Parse("/datasources/" + datasourceUUID.String())
			if err != nil {
				log.Error("can't parse location", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("can't parse location"))
			}

			return &api.OAuth2ClientCallbackFound{
				Location: api.OptURI{Value: *location, Set: true},
			}, nil
		})


}
*/
