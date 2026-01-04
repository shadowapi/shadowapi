package mapper

import (
	"encoding/base64"
	"encoding/json"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/transforms"
)

// Executor applies mapper configurations to source data.
type Executor struct {
	config api.MapperConfig
}

// NewExecutor creates a new mapper executor.
func NewExecutor(config api.MapperConfig) *Executor {
	return &Executor{config: config}
}

// Execute applies the mapper config to source data and returns mapped values by table.
// Returns map[tableName]map[fieldName]value
func (e *Executor) Execute(message *api.Message, contact *api.Contact) (map[string]map[string]any, error) {
	result := make(map[string]map[string]any)
	sourceData := buildSourceData(message, contact)

	for _, mapping := range e.config.Mappings {
		if !mapping.IsEnabled.Or(true) {
			continue
		}

		// Extract source value
		fieldKey := string(mapping.SourceEntity) + "." + mapping.SourceField
		value, exists := sourceData[fieldKey]
		if !exists {
			continue // Skip if field not found in source data
		}

		// Apply transform if specified
		var err error
		if mapping.Transform.IsSet() {
			value, err = applyTransform(value, mapping.Transform.Value, sourceData)
			if err != nil {
				continue // Skip on transform error
			}
		}

		// Store in result by table
		if result[mapping.TargetTable] == nil {
			result[mapping.TargetTable] = make(map[string]any)
		}
		result[mapping.TargetTable][mapping.TargetField] = value
	}

	return result, nil
}

// buildSourceData extracts all field values from message and contact into a flat map.
func buildSourceData(message *api.Message, contact *api.Contact) map[string]any {
	data := make(map[string]any)

	if message != nil {
		// Core fields - some are direct strings, some are OptString
		data["message.uuid"] = message.UUID.Or("")
		data["message.type"] = message.Type            // string, not optional
		data["message.format"] = message.Format        // string, not optional
		data["message.chat_uuid"] = message.ChatUUID.Or("")
		data["message.thread_uuid"] = message.ThreadUUID.Or("")
		data["message.external_message_id"] = message.ExternalMessageID.Or("")
		data["message.sender"] = message.Sender        // string, not optional
		data["message.subject"] = message.Subject.Or("")
		data["message.body"] = message.Body            // string, not optional
		data["message.forward_from"] = message.ForwardFrom.Or("")
		data["message.reply_to_message_uuid"] = message.ReplyToMessageUUID.Or("")
		data["message.forward_from_chat_uuid"] = message.ForwardFromChatUUID.Or("")
		data["message.forward_from_message_uuid"] = message.ForwardFromMessageUUID.Or("")

		// Recipients - direct slice
		if len(message.Recipients) > 0 {
			data["message.recipients"] = message.Recipients
		}

		// Attachments - direct slice
		if len(message.Attachments) > 0 {
			data["message.attachments"] = message.Attachments
		}

		// Timestamps
		if message.CreatedAt.IsSet() {
			data["message.created_at"] = message.CreatedAt.Value
		}
		if message.UpdatedAt.IsSet() {
			data["message.updated_at"] = message.UpdatedAt.Value
		}

		// Body parsed nested fields
		if message.BodyParsed.IsSet() {
			bp := message.BodyParsed.Value
			data["message.body_parsed.subject_text"] = bp.SubjectText.Or("")
			data["message.body_parsed.body_text"] = bp.BodyText   // string, not optional
			data["message.body_parsed.body_byte"] = base64.StdEncoding.EncodeToString(bp.BodyByte) // []byte -> base64 string
			if bp.SubjectSlate.IsSet() {
				data["message.body_parsed.subject_slate"] = bp.SubjectSlate.Value
			}
			if bp.BodySlate.IsSet() {
				data["message.body_parsed.body_slate"] = bp.BodySlate.Value
			}
		}

		// Meta nested fields
		if message.Meta.IsSet() {
			m := message.Meta.Value
			data["message.meta.has_raw_email"] = m.HasRawEmail.Or(false)
			data["message.meta.is_incoming"] = m.IsIncoming.Or(false)
		}

		// Complex fields as JSON
		if message.Reactions.IsSet() {
			data["message.reactions"] = message.Reactions.Value
		}
		if message.ForwardMeta.IsSet() {
			data["message.forward_meta"] = message.ForwardMeta.Value
		}
	}

	if contact != nil {
		// Core fields
		data["contact.uuid"] = contact.UUID.Or("")
		data["contact.user_uuid"] = contact.UserUUID.Or("")
		data["contact.instance_uuid"] = contact.InstanceUUID.Or("")
		data["contact.status"] = contact.Status.Or("")

		// Names
		data["contact.first"] = contact.First.Or("")
		data["contact.last"] = contact.Last.Or("")
		data["contact.middle"] = contact.Middle.Or("")
		data["contact.names_search"] = contact.NamesSearch.Or("")

		// Emails
		data["contact.email1"] = contact.Email1.Or("")
		data["contact.email1_type"] = contact.Email1Type.Or("")
		data["contact.email2"] = contact.Email2.Or("")
		data["contact.email2_type"] = contact.Email2Type.Or("")
		data["contact.email3"] = contact.Email3.Or("")
		data["contact.email4"] = contact.Email4.Or("")
		data["contact.email5"] = contact.Email5.Or("")
		data["contact.email_search"] = contact.EmailSearch.Or("")

		// Phones
		data["contact.phone1"] = contact.Phone1.Or("")
		data["contact.phone1_type"] = contact.Phone1Type.Or("")
		data["contact.phone1_country"] = contact.Phone1Country.Or("")
		data["contact.phone2"] = contact.Phone2.Or("")
		data["contact.phone3"] = contact.Phone3.Or("")
		data["contact.phone4"] = contact.Phone4.Or("")
		data["contact.phone5"] = contact.Phone5.Or("")
		data["contact.phone_search"] = contact.PhoneSearch.Or("")

		// Messengers
		data["contact.skype"] = contact.Skype.Or("")
		data["contact.skype_uuid"] = contact.SkypeUUID.Or("")
		data["contact.whatsapp"] = contact.Whatsapp.Or("")
		data["contact.whatsapp_uuid"] = contact.WhatsappUUID.Or("")
		data["contact.telegram"] = contact.Telegram.Or("")
		data["contact.telegram_uuid"] = contact.TelegramUUID.Or("")
		data["contact.wechat"] = contact.Wechat.Or("")
		data["contact.wechat_uuid"] = contact.WechatUUID.Or("")
		data["contact.line"] = contact.Line.Or("")
		data["contact.line_uuid"] = contact.LineUUID.Or("")
		data["contact.messengers_search"] = contact.MessengersSearch.Or("")

		// Social profiles
		data["contact.linkedin_uuid"] = contact.LinkedinUUID.Or("")
		data["contact.linkedin_url"] = contact.LinkedinURL.Or("")
		data["contact.facebook_uuid"] = contact.FacebookUUID.Or("")
		data["contact.facebook_url"] = contact.FacebookURL.Or("")
		data["contact.twitter_uuid"] = contact.TwitterUUID.Or("")
		data["contact.twitter_url"] = contact.TwitterURL.Or("")
		data["contact.github_uuid"] = contact.GithubUUID.Or("")
		data["contact.github_url"] = contact.GithubURL.Or("")
		data["contact.instagram_uuid"] = contact.InstagramUUID.Or("")
		data["contact.instagram_url"] = contact.InstagramURL.Or("")
		data["contact.socials_search"] = contact.SocialsSearch.Or("")

		// Position - OptInt fields need special handling
		if contact.LastPositionID.IsSet() {
			data["contact.last_position_id"] = contact.LastPositionID.Value
		}
		if contact.LastPositionCompanyID.IsSet() {
			data["contact.last_position_company_id"] = contact.LastPositionCompanyID.Value
		}
		data["contact.last_position_company_name"] = contact.LastPositionCompanyName.Or("")
		data["contact.last_position_title"] = contact.LastPositionTitle.Or("")
		data["contact.last_position_description"] = contact.LastPositionDescription.Or("")

		// Other
		data["contact.birthday_type"] = contact.BirthdayType.Or("")
		data["contact.salary"] = contact.Salary.Or("")
		data["contact.note_search"] = contact.NoteSearch.Or("")
		data["contact.tracking_source"] = contact.TrackingSource.Or("")
		data["contact.tracking_slug"] = contact.TrackingSlug.Or("")
		data["contact.cached_img"] = contact.CachedImg.Or("")

		// Date fields
		if contact.Birthday.IsSet() {
			data["contact.birthday"] = contact.Birthday.Value
		}
		if contact.LastPositionStartDate.IsSet() {
			data["contact.last_position_start_date"] = contact.LastPositionStartDate.Value
		}
		if contact.LastPositionEndDate.IsSet() {
			data["contact.last_position_end_date"] = contact.LastPositionEndDate.Value
		}
		if contact.EntryDate.IsSet() {
			data["contact.entry_date"] = contact.EntryDate.Value
		}
		if contact.EditDate.IsSet() {
			data["contact.edit_date"] = contact.EditDate.Value
		}

		// Complex fields - pointers (check for nil, not IsSet)
		if contact.Names != nil {
			data["contact.names"] = contact.Names
		}
		if contact.Emails != nil {
			data["contact.emails"] = contact.Emails
		}
		if contact.Phones != nil {
			data["contact.phones"] = contact.Phones
		}
		if contact.Messengers != nil {
			data["contact.messengers"] = contact.Messengers
		}
		if contact.Socials != nil {
			data["contact.socials"] = contact.Socials
		}
		if contact.LastPositions != nil {
			data["contact.last_positions"] = contact.LastPositions
		}
		if contact.SalaryData != nil {
			data["contact.salary_data"] = contact.SalaryData
		}
		if contact.CachedImgData != nil {
			data["contact.cached_img_data"] = contact.CachedImgData
		}
		if contact.Crawl != nil {
			data["contact.crawl"] = contact.Crawl
		}
	}

	return data
}

// applyTransform applies a transformation to a value using the shared transforms package.
func applyTransform(value any, transform api.MapperTransform, sourceData map[string]any) (any, error) {
	t := transforms.Transform{
		Type:   string(transform.Type),
		Params: extractParams(transform.Params),
	}
	return transforms.Apply(value, t, sourceData)
}

// extractParams converts API transform params to a simple map.
func extractParams(opt api.OptMapperTransformParams) map[string]any {
	if !opt.IsSet() {
		return nil
	}
	result := make(map[string]any)
	for k, v := range opt.Value {
		var parsed any
		if err := json.Unmarshal(v, &parsed); err == nil {
			result[k] = parsed
		}
	}
	return result
}
