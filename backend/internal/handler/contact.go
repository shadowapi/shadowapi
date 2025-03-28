package handler

import (
	"context"
	ht "github.com/ogen-go/ogen/http"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

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
