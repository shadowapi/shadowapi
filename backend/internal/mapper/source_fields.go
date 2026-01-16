package mapper

import "github.com/shadowapi/shadowapi/backend/pkg/api"

// DatasourceType represents the type of data source for field filtering.
type DatasourceType string

const (
	DatasourceTypeEmail      DatasourceType = "email"
	DatasourceTypeEmailOAuth DatasourceType = "email_oauth"
	DatasourceTypeTelegram   DatasourceType = "telegram"
	DatasourceTypeWhatsapp   DatasourceType = "whatsapp"
	DatasourceTypeLinkedin   DatasourceType = "linkedin"
)

// datasourceMessageFields maps datasource types to available message field names.
// NOTE: Only includes fields that are actually populated by the current implementation.
// body_parsed.* and meta.* fields are NOT populated (storage always sets BodyParsed: nil).
var datasourceMessageFields = map[DatasourceType][]string{
	DatasourceTypeEmail: {
		// Core identifiers
		"uuid", "type", "format", "chat_uuid", "thread_uuid", "external_message_id",
		// Sender and recipients
		"sender", "recipients",
		// Content
		"subject", "body", "attachments",
		// Timestamps
		"created_at", "updated_at",
	},
	DatasourceTypeEmailOAuth: {
		// Same as email - Gmail uses the same message structure
		"uuid", "type", "format", "chat_uuid", "thread_uuid", "external_message_id",
		// Sender and recipients
		"sender", "recipients",
		// Content
		"subject", "body", "attachments",
		// Timestamps
		"created_at", "updated_at",
	},
	DatasourceTypeTelegram: {
		// Core identifiers (no thread_uuid or subject in Telegram)
		"uuid", "type", "format", "chat_uuid", "external_message_id",
		// Sender and recipients
		"sender", "recipients",
		// Content
		"body", "reactions", "attachments",
		// Forwarding and replies (Telegram supports full forwarding)
		"forward_from", "reply_to_message_uuid",
		"forward_from_chat_uuid", "forward_from_message_uuid", "forward_meta",
		// Timestamps
		"created_at", "updated_at",
	},
	DatasourceTypeWhatsapp: {
		// Core identifiers (no thread_uuid or subject in WhatsApp)
		"uuid", "type", "format", "chat_uuid", "external_message_id",
		// Sender and recipients
		"sender", "recipients",
		// Content
		"body", "reactions", "attachments",
		// Forwarding and replies (limited forwarding info in WhatsApp)
		"forward_from", "reply_to_message_uuid",
		// Timestamps
		"created_at", "updated_at",
	},
	DatasourceTypeLinkedin: {
		// Core identifiers
		"uuid", "type", "format", "chat_uuid", "thread_uuid", "external_message_id",
		// Sender and recipients
		"sender", "recipients",
		// Content (LinkedIn has simple text messages)
		"body", "attachments",
		// Timestamps
		"created_at", "updated_at",
	},
}

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

// GetAllSourceFields returns all available source fields for mapping.
func GetAllSourceFields() []api.SourceFieldDefinition {
	return messageFields
}

// FilterByEntity filters source fields by entity type.
func FilterByEntity(fields []api.SourceFieldDefinition, entity api.MapperSourceFieldsListEntity) []api.SourceFieldDefinition {
	var entityType api.SourceFieldDefinitionEntity
	switch entity {
	case api.MapperSourceFieldsListEntityMessage:
		entityType = api.SourceFieldDefinitionEntityMessage
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

// FilterByDatasourceType filters source fields by datasource type.
// If the datasource type is empty or unknown, returns all fields.
func FilterByDatasourceType(fields []api.SourceFieldDefinition, dsType string) []api.SourceFieldDefinition {
	if dsType == "" {
		return fields
	}

	dt := DatasourceType(dsType)

	// Build lookup set of allowed field names for messages
	allowedMessageFields := make(map[string]bool)

	if msgFields, ok := datasourceMessageFields[dt]; ok {
		for _, f := range msgFields {
			allowedMessageFields[f] = true
		}
	}

	// If unknown datasource type, return all fields
	if len(allowedMessageFields) == 0 {
		return fields
	}

	result := make([]api.SourceFieldDefinition, 0)
	for _, f := range fields {
		if f.Entity == api.SourceFieldDefinitionEntityMessage {
			if allowedMessageFields[f.Name] {
				result = append(result, f)
			}
		}
	}
	return result
}
