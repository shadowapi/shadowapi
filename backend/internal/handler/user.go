package handler

import (
	"context"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// CreateUser implements createUser operation.
//
// Create a new user.
//
// POST /user
func (h *Handler) CreateUser(ctx context.Context, req *api.User) (*api.User, error) {
	return h.userManager.CreateUser(ctx, req)
}

// DeleteUser implements deleteUser operation.
//
// Delete user.
//
// DELETE /user/{uuid}
func (h *Handler) DeleteUser(ctx context.Context, params api.DeleteUserParams) error {
	return h.userManager.DeleteUser(ctx, params.UUID)
}

// GetUser implements getUser operation.
//
// Get user details.
//
// GET /user/{uuid}
func (h *Handler) GetUser(ctx context.Context, params api.GetUserParams) (*api.User, error) {
	return h.userManager.GetUser(ctx, params.UUID)
}

// ListUsers implements listUsers operation.
//
// List all users.
//
// GET /user
func (h *Handler) ListUsers(ctx context.Context) ([]api.User, error) {
	return h.userManager.ListUsers(ctx)
}

// UpdateUser implements updateUser operation.
//
// Update user details.
//
// PUT /user/{uuid}
func (h *Handler) UpdateUser(ctx context.Context, req *api.User, params api.UpdateUserParams) (*api.User, error) {
	return h.userManager.UpdateUser(ctx, req, params.UUID)
}

// CreateUserSession implements createUserSession operation.
// This endpoint is no longer supported without Zitadel.
//
// POST /users/session
func (h *Handler) CreateUserSession(ctx context.Context) (*api.UserSessionToken, error) {
	h.log.Error("session token creation is not supported")
	return nil, ErrWithCode(400, E("session token creation is not supported"))
}
