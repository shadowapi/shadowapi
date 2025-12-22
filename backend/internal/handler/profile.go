package handler

import (
	"context"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// GetProfile implements getProfile operation.
// Returns the current authenticated user's profile.
func (h *Handler) GetProfile(ctx context.Context) (*api.User, error) {
	// Get user claims from context (set by auth middleware)
	claims, ok := ctx.Value(auth.UserClaimsContextKey).(*oauth2.Claims)
	if !ok || claims == nil {
		return nil, ErrWithCode(http.StatusUnauthorized, E("not authenticated"))
	}

	// Get user UUID from claims subject
	userUUID := claims.Subject
	if userUUID == "" {
		return nil, ErrWithCode(http.StatusUnauthorized, E("invalid token: no subject"))
	}

	// Fetch user from database
	user, err := h.userManager.GetUser(ctx, userUUID)
	if err != nil {
		h.log.Error("failed to get user profile", "uuid", userUUID, "error", err)
		return nil, err
	}

	return user, nil
}

// UpdateProfile implements updateProfile operation.
func (h *Handler) UpdateProfile(ctx context.Context, req *api.UserProfile) (*api.User, error) {
	// Get user claims from context
	claims, ok := ctx.Value(auth.UserClaimsContextKey).(*oauth2.Claims)
	if !ok || claims == nil {
		return nil, ErrWithCode(http.StatusUnauthorized, E("not authenticated"))
	}

	userUUID := claims.Subject
	if userUUID == "" {
		return nil, ErrWithCode(http.StatusUnauthorized, E("invalid token: no subject"))
	}

	// Get current user
	currentUser, err := h.userManager.GetUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Update only allowed fields (first_name, last_name)
	currentUser.FirstName = req.FirstName
	currentUser.LastName = req.LastName

	// Save updates
	updatedUser, err := h.userManager.UpdateUser(ctx, currentUser, userUUID)
	if err != nil {
		h.log.Error("failed to update user profile", "uuid", userUUID, "error", err)
		return nil, err
	}

	return updatedUser, nil
}
