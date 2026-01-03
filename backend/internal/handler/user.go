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
func (h *Handler) CreateUser(ctx context.Context, req *api.User) (api.CreateUserRes, error) {
	user, err := h.userManager.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser implements deleteUser operation.
//
// Delete user.
//
// DELETE /user/{uuid}
func (h *Handler) DeleteUser(ctx context.Context, params api.DeleteUserParams) (api.DeleteUserRes, error) {
	if err := h.userManager.DeleteUser(ctx, params.UUID); err != nil {
		return nil, err
	}
	return &api.DeleteUserOK{}, nil
}

// GetUser implements getUser operation.
//
// Get user details.
//
// GET /user/{uuid}
func (h *Handler) GetUser(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
	user, err := h.userManager.GetUser(ctx, params.UUID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers implements listUsers operation.
//
// List all users.
//
// GET /user
func (h *Handler) ListUsers(ctx context.Context) (api.ListUsersRes, error) {
	users, err := h.userManager.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	res := api.ListUsersOKApplicationJSON(users)
	return &res, nil
}

// UpdateUser implements updateUser operation.
//
// Update user details.
//
// PUT /user/{uuid}
func (h *Handler) UpdateUser(ctx context.Context, req *api.User, params api.UpdateUserParams) (api.UpdateUserRes, error) {
	user, err := h.userManager.UpdateUser(ctx, req, params.UUID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUserSession implements createUserSession operation.
// This endpoint is no longer supported without Zitadel.
//
// POST /users/session
func (h *Handler) CreateUserSession(ctx context.Context) (api.CreateUserSessionRes, error) {
	h.log.Error("session token creation is not supported")
	return nil, ErrWithCode(400, E("session token creation is not supported"))
}
