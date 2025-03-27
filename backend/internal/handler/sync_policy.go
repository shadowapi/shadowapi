package handler

import (
	"context"
	ht "github.com/ogen-go/ogen/http"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

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
