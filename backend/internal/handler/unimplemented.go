package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"

	ht "github.com/ogen-go/ogen/http"
)

// TODO @reactima consider removing

// DatasourceEmailList implements datasource-email-list operation.
//
// List email datasources.
//
// GET /datasource/email
func (h *Handler) DatasourceEmailList(ctx context.Context, params api.DatasourceEmailListParams) ([]api.DatasourceEmail, error) {
	var pl []api.DatasourceEmail

	return pl, ht.ErrNotImplemented
}

// DatasourceLinkedinList implements datasource-linkedin-list operation.
//
// List all LinkedIn datasources.
//
// GET /datasource/linkedin
func (h *Handler) DatasourceLinkedinList(ctx context.Context, params api.DatasourceLinkedinListParams) ([]api.DatasourceLinkedin, error) {
	var pl []api.DatasourceLinkedin

	return pl, ht.ErrNotImplemented
}

// DatasourceSetOAuth2Client implements datasource-set-oauth2-client operation.
//
// Set OAuth2 client to the datasource.
//
// PUT /datasource/{uuid}/oauth2/client
func (h *Handler) DatasourceSetOAuth2Client(ctx context.Context, req *api.DatasourceSetOAuth2ClientReq, params api.DatasourceSetOAuth2ClientParams) error {
	return ht.ErrNotImplemented
}

// DatasourceTelegramList implements datasource-telegram-list operation.
//
// List all Telegram datasources.
//
// GET /datasource/telegram
func (h *Handler) DatasourceTelegramList(ctx context.Context, params api.DatasourceTelegramListParams) ([]api.DatasourceTelegram, error) {
	var pl []api.DatasourceTelegram

	return pl, ht.ErrNotImplemented
}

// DatasourceWhatsappList implements datasource-whatsapp-list operation.
//
// List all WhatsApp datasources.
//
// GET /datasource/whatsapp
func (h *Handler) DatasourceWhatsappList(ctx context.Context, params api.DatasourceWhatsappListParams) ([]api.DatasourceWhatsapp, error) {
	var pl []api.DatasourceWhatsapp

	return pl, ht.ErrNotImplemented
}
