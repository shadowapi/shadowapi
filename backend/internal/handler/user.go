package handler

import (
	"context"
	"net/http"

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
//
// # Create a session token for Zitadel authentication
//
// POST /users/session
func (h *Handler) CreateUserSession(ctx context.Context) (*api.UserSessionToken, error) {
	h.log.Info("creating user session token")

	// Check if we're using Zitadel user manager
	if h.cfg.Auth.UserManager != "zitadel" {
		h.log.Error("session token creation requires Zitadel user manager")
		return nil, ErrWithCode(http.StatusBadRequest, E("session token creation requires Zitadel user manager"))
	}

	// Get the service user token from ZitadelUserManager
	zitadelManager, ok := h.userManager.(interface {
		GetAuthToken(context.Context) (string, error)
	})
	if !ok {
		h.log.Error("user manager doesn't support GetAuthToken method")
		return nil, ErrWithCode(http.StatusInternalServerError, E("Zitadel user manager not properly configured"))
	}

	token, err := zitadelManager.GetAuthToken(ctx)
	if err != nil {
		h.log.Error("failed to get Zitadel auth token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get authentication token"))
	}

	response := &api.UserSessionToken{
		SessionToken: token,
		ZitadelURL:   h.cfg.Auth.Zitadel.InstanceURL,
		ExpiresIn:    3600, // 1 hour
	}

	h.log.Info("user session token created successfully", "zitadel_url", h.cfg.Auth.Zitadel.InstanceURL)
	return response, nil
}
