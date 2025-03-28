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

// CreateContact implements createContact operation.
//
// Create a new contact record.
//
// POST /contact
func (h *Handler) CreateContact(ctx context.Context, req *api.Contact) (*api.Contact, error) {
	var r *api.Contact
	return r, ht.ErrNotImplemented
}

// DeleteContact implements deleteContact operation.
//
// Delete a contact record.
//
// DELETE /contact/{uuid}
func (h *Handler) DeleteContact(ctx context.Context, params api.DeleteContactParams) error {
	return ht.ErrNotImplemented
}

// GetContact implements getContact operation.
//
// Get contact details.
//
// GET /contact/{uuid}
func (h *Handler) GetContact(ctx context.Context, params api.GetContactParams) (*api.Contact, error) {
	var r *api.Contact
	return r, ht.ErrNotImplemented
}

// ListContacts implements listContacts operation.
//
// List all contacts.
//
// GET /contact
func (h *Handler) ListContacts(ctx context.Context) ([]api.Contact, error) {
	var r []api.Contact
	return r, ht.ErrNotImplemented
}

// UpdateContact implements updateContact operation.
//
// Update contact details.
//
// PUT /contact/{uuid}
func (h *Handler) UpdateContact(ctx context.Context, req *api.Contact, params api.UpdateContactParams) (*api.Contact, error) {
	var r *api.Contact
	return r, ht.ErrNotImplemented
}

// FileCreate implements file-create operation.
//
// Upload a new file and create its record.
//
// POST /file
func (h *Handler) FileCreate(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	var r *api.UploadFileResponse
	return r, ht.ErrNotImplemented
}

// FileDelete implements file-delete operation.
//
// Delete a stored file.
//
// DELETE /file/{uuid}
func (h *Handler) FileDelete(ctx context.Context, params api.FileDeleteParams) error {
	return ht.ErrNotImplemented
}

// FileGet implements file-get operation.
//
// Retrieve details of a stored file.
//
// GET /file/{uuid}
func (h *Handler) FileGet(ctx context.Context, params api.FileGetParams) (*api.FileObject, error) {
	var r *api.FileObject
	return r, ht.ErrNotImplemented
}

// FileList implements file-list operation.
//
// Retrieve a list of stored files.
//
// GET /file
func (h *Handler) FileList(ctx context.Context, params api.FileListParams) ([]api.FileObject, error) {
	var r []api.FileObject
	return r, ht.ErrNotImplemented
}

// FileUpdate implements file-update operation.
//
// Update metadata of a stored file.
//
// PUT /file/{uuid}
func (h *Handler) FileUpdate(ctx context.Context, req *api.FileUpdateReq, params api.FileUpdateParams) (*api.FileObject, error) {
	var r *api.FileObject
	return r, ht.ErrNotImplemented
}
