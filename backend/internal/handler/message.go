package handler

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"net/http"
	"time"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// MessageEmailQuery implements MessageEmailQuery operation.
// Execute a search query on email Message.
// POST /message/email/query
func (h *Handler) MessageEmailQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageEmailQueryOK, error) {
	log := h.log.With("handler", "MessageEmailQuery")
	params := convertMessageQueryToParams(req, "email")
	rows, err := query.New(h.dbp).GetMessages(ctx, params)
	if err != nil {
		log.Error("failed to query email messages", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query email messages"))
	}
	var messages []api.Message
	for _, row := range rows {
		m, err := qToApiMessage(row)
		if err != nil {
			log.Error("failed to map email message", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map email message"))
		}
		messages = append(messages, m)
	}
	return &api.MessageEmailQueryOK{Messages: messages}, nil
}

// MessageLinkedinQuery implements MessageLinkedinQuery operation.
// Execute a search query on LinkedIn Message.
// POST /message/linkedin/query
func (h *Handler) MessageLinkedinQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageLinkedinQueryOK, error) {
	log := h.log.With("handler", "MessageLinkedinQuery")
	params := convertMessageQueryToParams(req, "linkedin")
	rows, err := query.New(h.dbp).GetMessages(ctx, params)
	if err != nil {
		log.Error("failed to query linkedin messages", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query linkedin messages"))
	}
	var messages []api.Message
	for _, row := range rows {
		m, err := qToApiMessage(row)
		if err != nil {
			log.Error("failed to map linkedin message", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map linkedin message"))
		}
		messages = append(messages, m)
	}
	return &api.MessageLinkedinQueryOK{Messages: messages}, nil
}

// MessageTelegramQuery implements MessageTelegramQuery operation.
// Execute a search query on Telegram Message.
// POST /message/telegram/query
func (h *Handler) MessageTelegramQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageTelegramQueryOK, error) {
	log := h.log.With("handler", "MessageTelegramQuery")
	params := convertMessageQueryToParams(req, "telegram")
	rows, err := query.New(h.dbp).GetMessages(ctx, params)
	if err != nil {
		log.Error("failed to query telegram messages", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query telegram messages"))
	}
	var messages []api.Message
	for _, row := range rows {
		m, err := qToApiMessage(row)
		if err != nil {
			log.Error("failed to map telegram message", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map telegram message"))
		}
		messages = append(messages, m)
	}
	return &api.MessageTelegramQueryOK{Messages: messages}, nil
}

// MessageWhatsappQuery implements MessageWhatsappQuery operation.
// Execute a search query on WhatsApp Message.
// POST /message/whatsapp/query
func (h *Handler) MessageWhatsappQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageWhatsappQueryOK, error) {
	log := h.log.With("handler", "MessageWhatsappQuery")
	params := convertMessageQueryToParams(req, "whatsapp")
	rows, err := query.New(h.dbp).GetMessages(ctx, params)
	if err != nil {
		log.Error("failed to query whatsapp messages", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query whatsapp messages"))
	}
	var messages []api.Message
	for _, row := range rows {
		m, err := qToApiMessage(row)
		if err != nil {
			log.Error("failed to map whatsapp message", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map whatsapp message"))
		}
		messages = append(messages, m)
	}
	return &api.MessageWhatsappQueryOK{Messages: messages}, nil
}

// convertMessageQueryToParams converts an API MessageQuery into query.GetMessagesParams.
// It uses the provided msgType (e.g. "email", "linkedin", "telegram", "whatsapp") as filter.
// convertMessageQueryToParams converts an API MessageQuery into query.GetMessagesParams.
// It uses the provided msgType (e.g. "email", "linkedin", "telegram", "whatsapp") as filter.
func convertMessageQueryToParams(req *api.MessageQuery, msgType string) query.GetMessagesParams {
	orderBy := "created_at"
	orderDirection := "desc"
	if req.Order.IsSet() {
		orderDirection = string(req.Order.Value) // MessageQueryOrder underlying type is string
	}
	offset := int32(0)
	if req.Offset.IsSet() {
		offset = int32(req.Offset.Value)
	}
	limit := int32(50)
	if req.Limit.IsSet() {
		limit = int32(req.Limit.Value)
	}

	var chatUUID pgtype.UUID
	if req.ChatID.IsSet() && req.ChatID.Value != "" {
		if u, err := uuid.FromString(req.ChatID.Value); err == nil {
			chatUUID = pgtype.UUID{Bytes: converter.UToBytes(u), Valid: true}
		} else {
			chatUUID = pgtype.UUID{Valid: false}
		}
	} else {
		chatUUID = pgtype.UUID{Valid: false}
	}

	var threadUUID pgtype.UUID
	if req.ThreadID.IsSet() && req.ThreadID.Value != "" {
		if u, err := uuid.FromString(req.ThreadID.Value); err == nil {
			threadUUID = pgtype.UUID{Bytes: converter.UToBytes(u), Valid: true}
		} else {
			threadUUID = pgtype.UUID{Valid: false}
		}
	} else {
		threadUUID = pgtype.UUID{Valid: false}
	}

	return query.GetMessagesParams{
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
		Offset:         offset,
		Limit:          limit,
		Type:           msgType,
		Format:         "",
		ChatUuid:       chatUUID,
		ThreadUuid:     threadUUID,
		Sender:         "",
	}
}

// qToApiMessage converts a query.GetMessagesRow into an API Message.
func qToApiMessage(r query.GetMessagesRow) (api.Message, error) {
	var msg api.Message
	msg.UUID = api.NewOptString(r.UUID.String())
	msg.Format = r.Format
	msg.Type = r.Type
	if r.ChatUuid != nil {
		msg.ChatUUID = api.NewOptString(r.ChatUuid.String())
	}
	if r.ThreadUuid != nil {
		msg.ThreadUUID = api.NewOptString(r.ThreadUuid.String())
	}
	msg.Sender = r.Sender
	msg.Recipients = r.Recipients
	msg.Subject = api.NewOptString(r.Subject.String)
	msg.Body = r.Body
	if len(r.BodyParsed) > 0 {
		var bp api.MessageBodyParsed
		if err := json.Unmarshal(r.BodyParsed, &bp); err != nil {
			return msg, err
		}
		msg.BodyParsed.SetTo(bp)
	}
	if len(r.Reactions) > 0 {
		var react map[string]int
		if err := json.Unmarshal(r.Reactions, &react); err != nil {
			return msg, err
		}
		msg.Reactions.SetTo(react)
	}
	if len(r.Attachments) > 0 {
		var atts []api.FileObject
		if err := json.Unmarshal(r.Attachments, &atts); err != nil {
			return msg, err
		}
		msg.Attachments = atts
	}
	msg.ForwardFrom = api.NewOptString(r.ForwardFrom.String)
	if r.ReplyToMessageUuid != nil {
		msg.ReplyToMessageUUID = api.NewOptString(r.ReplyToMessageUuid.String())
	}
	if r.ForwardFromChatUuid != nil {
		msg.ForwardFromChatUUID = api.NewOptString(r.ForwardFromChatUuid.String())
	}
	if r.ForwardFromMessageUuid != nil {
		msg.ForwardFromMessageUUID = api.NewOptString(r.ForwardFromMessageUuid.String())
	}
	if len(r.ForwardMeta) > 0 {
		var fm map[string]interface{}
		if err := json.Unmarshal(r.ForwardMeta, &fm); err != nil {
			return msg, err
		}
		b, err := json.Marshal(fm)
		if err != nil {
			return msg, err
		}
		msg.ForwardMeta.SetTo(api.MessageForwardMeta{"data": b})
	}
	if len(r.Meta) > 0 {
		var meta map[string]interface{}
		if err := json.Unmarshal(r.Meta, &meta); err != nil {
			return msg, err
		}
		b, err := json.Marshal(meta)
		if err != nil {
			return msg, err
		}
		msg.Meta.SetTo(api.MessageMeta{"data": b})
	}
	if r.CreatedAt.Valid {
		msg.CreatedAt = api.NewOptDateTime(r.CreatedAt.Time)
	} else {
		msg.CreatedAt = api.NewOptDateTime(time.Time{})
	}
	if r.UpdatedAt.Valid {
		msg.UpdatedAt = api.NewOptDateTime(r.UpdatedAt.Time)
	} else {
		msg.UpdatedAt = api.NewOptDateTime(time.Time{})
	}
	return msg, nil
}
