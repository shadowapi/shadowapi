package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"

	ht "github.com/ogen-go/ogen/http"
)

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

// SyncpolicyCreate implements syncpolicy-create operation.
//
// Create a new sync policy.
//
// POST /syncpolicy
func (h *Handler) SyncpolicyCreate(ctx context.Context, req *api.SyncPolicy) (*api.SyncPolicy, error) {
	return nil, ht.ErrNotImplemented
}

// SyncpolicyDelete implements syncpolicy-delete operation.
//
// Delete a sync policy by uuid.
//
// DELETE /syncpolicy/{uuid}
func (h *Handler) SyncpolicyDelete(ctx context.Context, params api.SyncpolicyDeleteParams) error {
	return ht.ErrNotImplemented
}

// SyncpolicyGet implements syncpolicy-get operation.
//
// Retrieve a specific sync policy by uuid.
//
// GET /syncpolicy/{uuid}
func (h *Handler) SyncpolicyGet(ctx context.Context, params api.SyncpolicyGetParams) (*api.SyncPolicy, error) {
	return nil, ht.ErrNotImplemented
}

// SyncpolicyList implements syncpolicy-list operation.
//
// Retrieve a list of sync policies for the authenticated user.
//
// GET /syncpolicy
func (h *Handler) SyncpolicyList(ctx context.Context, params api.SyncpolicyListParams) (*api.SyncpolicyListOK, error) {
	return nil, ht.ErrNotImplemented
}

// SyncpolicyUpdate implements syncpolicy-update operation.
//
// Update a sync policy by uuid.
//
// PUT /syncpolicy/{uuid}
func (h *Handler) SyncpolicyUpdate(ctx context.Context, req *api.SyncPolicy, params api.SyncpolicyUpdateParams) (*api.SyncPolicy, error) {
	return nil, ht.ErrNotImplemented
}
