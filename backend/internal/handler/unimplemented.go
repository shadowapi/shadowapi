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
func (h *Handler) DatasourceEmailList(ctx context.Context, params api.DatasourceEmailListParams) (api.DatasourceEmailListRes, error) {
	return nil, ht.ErrNotImplemented
}

// DatasourceLinkedinList implements datasource-linkedin-list operation.
//
// List all LinkedIn datasources.
//
// GET /datasource/linkedin
func (h *Handler) DatasourceLinkedinList(ctx context.Context, params api.DatasourceLinkedinListParams) (api.DatasourceLinkedinListRes, error) {
	return nil, ht.ErrNotImplemented
}

// DatasourceSetOAuth2Client implements datasource-set-oauth2-client operation.
//
// Set OAuth2 client to the datasource.
//
// PUT /datasource/{uuid}/oauth2/client
func (h *Handler) DatasourceSetOAuth2Client(ctx context.Context, req *api.DatasourceSetOAuth2ClientReq, params api.DatasourceSetOAuth2ClientParams) (api.DatasourceSetOAuth2ClientRes, error) {
	return nil, ht.ErrNotImplemented
}

// DatasourceTelegramList implements datasource-telegram-list operation.
//
// List all Telegram datasources.
//
// GET /datasource/telegram
func (h *Handler) DatasourceTelegramList(ctx context.Context, params api.DatasourceTelegramListParams) (api.DatasourceTelegramListRes, error) {
	return nil, ht.ErrNotImplemented
}

// DatasourceWhatsappList implements datasource-whatsapp-list operation.
//
// List all WhatsApp datasources.
//
// GET /datasource/whatsapp
func (h *Handler) DatasourceWhatsappList(ctx context.Context, params api.DatasourceWhatsappListParams) (api.DatasourceWhatsappListRes, error) {
	return nil, ht.ErrNotImplemented
}

// GeneratePresignedUploadUrl implements generatePresignedUploadUrl operation.
//
// Generate a pre-signed URL for file upload.
//
// POST /storage/upload-url
func (h *Handler) GeneratePresignedUploadUrl(ctx context.Context, req *api.UploadPresignedUrlRequest) (api.GeneratePresignedUploadUrlRes, error) {
	return nil, ht.ErrNotImplemented
}
