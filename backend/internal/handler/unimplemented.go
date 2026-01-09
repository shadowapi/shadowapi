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
func (h *Handler) DatasourceEmailList(ctx context.Context, params api.DatasourceEmailListParams) (api.DatasourceEmailListRes, error) {
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
