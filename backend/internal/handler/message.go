package handler

import (
	"context"
	ht "github.com/ogen-go/ogen/http"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// MessageEmailQuery implements MessageEmailQuery operation.
//
// Execute a search query on email Message.
//
// POST /message/email/query
func (h *Handler) MessageEmailQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageEmailQueryOK, error) {
	return nil, ht.ErrNotImplemented
}

// MessageLinkedinQuery implements MessageLinkedinQuery operation.
//
// Execute a search query on LinkedIn Message.
//
// POST /message/linkedin/query
func (h *Handler) MessageLinkedinQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageLinkedinQueryOK, error) {
	return nil, ht.ErrNotImplemented
}

// MessageTelegramQuery implements MessageTelegramQuery operation.
//
// Execute a search query on Telegram Message.
//
// POST /message/telegram/query
func (h *Handler) MessageTelegramQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageTelegramQueryOK, error) {
	return nil, ht.ErrNotImplemented
}

// MessageWhatsappQuery implements MessageWhatsappQuery operation.
//
// Execute a search query on WhatsApp Message.
//
// POST /message/whatsapp/query
func (h *Handler) MessageWhatsappQuery(ctx context.Context, req *api.MessageQuery) (*api.MessageWhatsappQueryOK, error) {
	return nil, ht.ErrNotImplemented
}
