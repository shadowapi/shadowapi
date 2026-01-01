package mapper

import "github.com/shadowapi/shadowapi/backend/pkg/api"

// Message source fields - derived from spec/components/message.yaml
var messageFields = []api.SourceFieldDefinition{
	// Core identifiers
	{Name: "uuid", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Unique message identifier")},
	{Name: "type", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Message type (email, telegram, whatsapp, linkedin)")},
	{Name: "format", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Message format (text, media, system, notification, attachment)")},
	{Name: "chat_uuid", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Conversation/chat UUID")},
	{Name: "thread_uuid", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Sub-thread UUID")},
	{Name: "external_message_id", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Original system message ID")},

	// Sender and recipients
	{Name: "sender", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Message sender identifier")},
	{Name: "recipients", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeArray, Description: api.NewOptString("List of recipient identifiers")},

	// Content
	{Name: "subject", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Message subject/title")},
	{Name: "body", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Message body text content")},
	{Name: "reactions", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Reaction counts object")},
	{Name: "attachments", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeArray, Description: api.NewOptString("Array of file attachment references")},

	// Forwarding and replies
	{Name: "forward_from", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Original sender if forwarded")},
	{Name: "reply_to_message_uuid", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Parent message UUID if reply")},
	{Name: "forward_from_chat_uuid", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Original chat UUID if forwarded")},
	{Name: "forward_from_message_uuid", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Original message UUID if forwarded")},
	{Name: "forward_meta", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Additional forwarding context")},

	// Timestamps
	{Name: "created_at", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Message creation timestamp")},
	{Name: "updated_at", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Last update timestamp")},

	// Nested: body_parsed
	{Name: "body_parsed.subject_text", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Parsed subject as plain text"), IsNested: api.NewOptBool(true), Path: api.NewOptString("body_parsed.subject_text")},
	{Name: "body_parsed.body_text", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Parsed body as plain text"), IsNested: api.NewOptBool(true), Path: api.NewOptString("body_parsed.body_text")},
	{Name: "body_parsed.body_byte", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Base64-encoded binary content"), IsNested: api.NewOptBool(true), Path: api.NewOptString("body_parsed.body_byte")},
	{Name: "body_parsed.subject_slate", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Subject in Slate.js JSON format"), IsNested: api.NewOptBool(true), Path: api.NewOptString("body_parsed.subject_slate")},
	{Name: "body_parsed.body_slate", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Body in Slate.js JSON format"), IsNested: api.NewOptBool(true), Path: api.NewOptString("body_parsed.body_slate")},

	// Nested: meta
	{Name: "meta.has_raw_email", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeBoolean, Description: api.NewOptString("Whether raw email data is available"), IsNested: api.NewOptBool(true), Path: api.NewOptString("meta.has_raw_email")},
	{Name: "meta.is_incoming", Entity: api.SourceFieldDefinitionEntityMessage, Type: api.SourceFieldDefinitionTypeBoolean, Description: api.NewOptString("Whether message is inbound"), IsNested: api.NewOptBool(true), Path: api.NewOptString("meta.is_incoming")},
}

// Contact source fields - derived from spec/components/contact.yaml
var contactFields = []api.SourceFieldDefinition{
	// Core identifiers
	{Name: "uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Unique contact identifier")},
	{Name: "user_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Owner user UUID")},
	{Name: "instance_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Instance UUID")},
	{Name: "status", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Contact status")},

	// Names
	{Name: "first", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("First name")},
	{Name: "last", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Last name")},
	{Name: "middle", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Middle name")},
	{Name: "names", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Names object with all name fields")},
	{Name: "names_search", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Searchable names string")},

	// Emails
	{Name: "email1", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Primary email address")},
	{Name: "email1_type", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Primary email type")},
	{Name: "email2", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Secondary email address")},
	{Name: "email2_type", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Secondary email type")},
	{Name: "email3", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Tertiary email address")},
	{Name: "email4", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Fourth email address")},
	{Name: "email5", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Fifth email address")},
	{Name: "emails", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("All emails object")},
	{Name: "email_search", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Searchable emails string")},

	// Phones
	{Name: "phone1", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Primary phone number")},
	{Name: "phone1_type", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Primary phone type")},
	{Name: "phone1_country", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Primary phone country code")},
	{Name: "phone2", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Secondary phone number")},
	{Name: "phone3", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Tertiary phone number")},
	{Name: "phone4", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Fourth phone number")},
	{Name: "phone5", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Fifth phone number")},
	{Name: "phones", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("All phones object")},
	{Name: "phone_search", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Searchable phones string")},

	// Messengers
	{Name: "skype", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Skype username")},
	{Name: "skype_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Skype UUID")},
	{Name: "whatsapp", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("WhatsApp number")},
	{Name: "whatsapp_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("WhatsApp UUID")},
	{Name: "telegram", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Telegram username")},
	{Name: "telegram_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Telegram UUID")},
	{Name: "wechat", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("WeChat ID")},
	{Name: "wechat_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("WeChat UUID")},
	{Name: "line", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("LINE ID")},
	{Name: "line_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("LINE UUID")},
	{Name: "messengers", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("All messengers object")},
	{Name: "messengers_search", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Searchable messengers string")},

	// Social profiles
	{Name: "linkedin_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("LinkedIn UUID")},
	{Name: "linkedin_url", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("LinkedIn profile URL")},
	{Name: "facebook_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Facebook UUID")},
	{Name: "facebook_url", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Facebook profile URL")},
	{Name: "twitter_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Twitter UUID")},
	{Name: "twitter_url", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Twitter profile URL")},
	{Name: "github_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("GitHub UUID")},
	{Name: "github_url", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("GitHub profile URL")},
	{Name: "instagram_uuid", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Instagram UUID")},
	{Name: "instagram_url", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Instagram profile URL")},
	{Name: "socials", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("All social profiles object")},
	{Name: "socials_search", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Searchable socials string")},

	// Position/Employment
	{Name: "last_position_id", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Last position ID")},
	{Name: "last_position_company_id", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Last position company ID")},
	{Name: "last_position_company_name", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Last position company name")},
	{Name: "last_position_title", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Last position job title")},
	{Name: "last_position_start_date", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Last position start date")},
	{Name: "last_position_end_date", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Last position end date")},
	{Name: "last_position_description", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Last position description")},
	{Name: "last_positions", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("All positions object")},

	// Other
	{Name: "birthday", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Birthday date")},
	{Name: "birthday_type", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Birthday type")},
	{Name: "salary", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Salary value")},
	{Name: "salary_data", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Salary data object")},
	{Name: "note_search", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Searchable notes string")},
	{Name: "tracking_source", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Contact tracking source")},
	{Name: "tracking_slug", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Contact tracking slug")},
	{Name: "cached_img", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeString, Description: api.NewOptString("Cached image URL")},
	{Name: "cached_img_data", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Cached image data")},
	{Name: "crawl", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeObject, Description: api.NewOptString("Crawl data object")},

	// Timestamps
	{Name: "entry_date", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Contact entry date")},
	{Name: "edit_date", Entity: api.SourceFieldDefinitionEntityContact, Type: api.SourceFieldDefinitionTypeDatetime, Description: api.NewOptString("Last edit date")},
}

// GetAllSourceFields returns all available source fields for mapping.
func GetAllSourceFields() []api.SourceFieldDefinition {
	all := make([]api.SourceFieldDefinition, 0, len(messageFields)+len(contactFields))
	all = append(all, messageFields...)
	all = append(all, contactFields...)
	return all
}

// FilterByEntity filters source fields by entity type.
func FilterByEntity(fields []api.SourceFieldDefinition, entity api.MapperSourceFieldsListEntity) []api.SourceFieldDefinition {
	var entityType api.SourceFieldDefinitionEntity
	switch entity {
	case api.MapperSourceFieldsListEntityMessage:
		entityType = api.SourceFieldDefinitionEntityMessage
	case api.MapperSourceFieldsListEntityContact:
		entityType = api.SourceFieldDefinitionEntityContact
	default:
		return fields
	}

	result := make([]api.SourceFieldDefinition, 0)
	for _, f := range fields {
		if f.Entity == entityType {
			result = append(result, f)
		}
	}
	return result
}

// FilterByType filters source fields by data type.
func FilterByType(fields []api.SourceFieldDefinition, fieldType api.MapperSourceFieldsListType) []api.SourceFieldDefinition {
	var sourceType api.SourceFieldDefinitionType
	switch fieldType {
	case api.MapperSourceFieldsListTypeString:
		sourceType = api.SourceFieldDefinitionTypeString
	case api.MapperSourceFieldsListTypeInteger:
		sourceType = api.SourceFieldDefinitionTypeInteger
	case api.MapperSourceFieldsListTypeBoolean:
		sourceType = api.SourceFieldDefinitionTypeBoolean
	case api.MapperSourceFieldsListTypeDatetime:
		sourceType = api.SourceFieldDefinitionTypeDatetime
	case api.MapperSourceFieldsListTypeArray:
		sourceType = api.SourceFieldDefinitionTypeArray
	case api.MapperSourceFieldsListTypeObject:
		sourceType = api.SourceFieldDefinitionTypeObject
	default:
		return fields
	}

	result := make([]api.SourceFieldDefinition, 0)
	for _, f := range fields {
		if f.Type == sourceType {
			result = append(result, f)
		}
	}
	return result
}

// BuildSourceFieldIndex creates a lookup map for source fields.
func BuildSourceFieldIndex() map[string]api.SourceFieldDefinition {
	fields := GetAllSourceFields()
	index := make(map[string]api.SourceFieldDefinition, len(fields))
	for _, f := range fields {
		key := string(f.Entity) + "." + f.Name
		index[key] = f
	}
	return index
}
