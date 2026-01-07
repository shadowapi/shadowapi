package handler

import (
	"context"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// GetProfile implements getProfile operation.
// Returns the current authenticated user's profile.
func (h *Handler) GetProfile(ctx context.Context) (api.GetProfileRes, error) {
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

	// Fetch user's roles from RBAC enforcer
	rolesMap, err := h.enforcer.GetAllRolesForUser(userUUID)
	if err != nil {
		h.log.Warn("failed to get user roles", "uuid", userUUID, "error", err)
		// Don't fail the request, just return empty roles
		rolesMap = make(map[string][]string)
	}

	// Convert roles map to API format
	user.Roles = convertRolesToAPI(rolesMap)

	// Add current workspace info from JWT claims if present
	if claims.WorkspaceID != "" && claims.WorkspaceSlug != "" {
		user.CurrentWorkspace = api.NewOptUserCurrentWorkspace(api.UserCurrentWorkspace{
			UUID: api.NewOptString(claims.WorkspaceID),
			Slug: api.NewOptString(claims.WorkspaceSlug),
		})
	}

	return user, nil
}

// convertRolesToAPI converts a map of domain->roles to API format.
func convertRolesToAPI(rolesMap map[string][]string) []api.UserRolesItem {
	var roles []api.UserRolesItem
	for domain, roleNames := range rolesMap {
		for _, roleName := range roleNames {
			roles = append(roles, api.UserRolesItem{
				Role:   api.NewOptString(roleName),
				Domain: api.NewOptString(domain),
			})
		}
	}
	return roles
}

// UpdateProfile implements updateProfile operation.
func (h *Handler) UpdateProfile(ctx context.Context, req *api.UserProfile) (api.UpdateProfileRes, error) {
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

// ChangePassword implements changePassword operation.
// Allows the authenticated user to change their own password.
func (h *Handler) ChangePassword(ctx context.Context, req *api.PasswordChange) (api.ChangePasswordRes, error) {
	// Get user claims from context (set by auth middleware)
	claims, ok := ctx.Value(auth.UserClaimsContextKey).(*oauth2.Claims)
	if !ok || claims == nil {
		return nil, ErrWithCode(http.StatusUnauthorized, E("not authenticated"))
	}

	userUUID := claims.Subject
	if userUUID == "" {
		return nil, ErrWithCode(http.StatusUnauthorized, E("invalid token: no subject"))
	}

	// Fetch user from database to get current password hash
	user, err := h.userManager.GetUser(ctx, userUUID)
	if err != nil {
		h.log.Error("failed to get user for password change", "uuid", userUUID, "error", err)
		return nil, err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		h.log.Debug("password change failed: current password incorrect", "uuid", userUUID)
		return nil, ErrWithCode(http.StatusUnauthorized, E("current password is incorrect"))
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("failed to hash new password", "uuid", userUUID, "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to process password"))
	}

	// Update password in database
	pgUUID, err := converter.ConvertStringToPgUUID(userUUID)
	if err != nil {
		h.log.Error("failed to convert user UUID", "uuid", userUUID, "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid user UUID"))
	}

	if err := query.New(h.dbp).UpdateUserPassword(ctx, query.UpdateUserPasswordParams{
		Password: string(hashedPassword),
		UUID:     pgUUID,
	}); err != nil {
		h.log.Error("failed to update password", "uuid", userUUID, "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update password"))
	}

	h.log.Info("password changed successfully", "uuid", userUUID)

	return &api.ChangePasswordNoContent{}, nil
}
