package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/email"
	"github.com/shadowapi/shadowapi/backend/internal/rbac"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

const (
	inviteTokenLength = 32
	inviteExpiry      = 24 * time.Hour
)

// ListWorkspaceInvites lists all pending invites for a workspace.
//
// GET /workspace/{uuid}/invites
func (h *Handler) ListWorkspaceInvites(ctx context.Context, params api.ListWorkspaceInvitesParams) (api.ListWorkspaceInvitesRes, error) {
	q := query.New(h.dbp)
	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)

	// Verify workspace exists
	_, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	invites, err := q.ListWorkspaceInvites(ctx, &workspaceUUID)
	if err != nil {
		h.log.Error("failed to list invites", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list invites"))
	}

	result := make([]api.UserInvite, 0, len(invites))
	for _, inv := range invites {
		result = append(result, dbInviteRowToAPI(inv))
	}

	res := api.ListWorkspaceInvitesOKApplicationJSON(result)
	return &res, nil
}

// CreateWorkspaceInvite creates a new invite and sends email.
//
// POST /workspace/{uuid}/invites
func (h *Handler) CreateWorkspaceInvite(ctx context.Context, req *api.UserInvite, params api.CreateWorkspaceInviteParams) (api.CreateWorkspaceInviteRes, error) {
	q := query.New(h.dbp)
	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)

	// Get current user
	userUUIDStr, err := getUserUUIDFromContext(ctx)
	if err != nil {
		return nil, ErrWithCode(http.StatusUnauthorized, E("authentication required"))
	}
	inviterUUID, err := uuid.FromString(userUUIDStr)
	if err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("invalid user UUID"))
	}

	// Get workspace for email template
	workspace, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	// Check if user already exists with this email
	existingUser, err := q.GetUserByEmail(ctx, req.Email)
	if err == nil {
		// User exists - check if already a member
		_, err := q.GetWorkspaceMember(ctx, query.GetWorkspaceMemberParams{
			WorkspaceUUID: &workspaceUUID,
			UserUUID:      &existingUser.UUID,
		})
		if err == nil {
			return nil, ErrWithCode(http.StatusConflict, E("user is already a member of this workspace"))
		}
	}

	// Check for existing active invite for this email
	_, err = q.GetUserInviteByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrWithCode(http.StatusConflict, E("an active invite already exists for this email"))
	}
	if err != pgx.ErrNoRows {
		h.log.Error("failed to check existing invite", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to check existing invite"))
	}

	// Generate secure token
	token, err := generateSecureToken(inviteTokenLength)
	if err != nil {
		h.log.Error("failed to generate token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to generate invite token"))
	}

	// Hash token for storage (SHA256 for fast lookup)
	tokenHash := hashToken(token)

	// Create invite
	inviteUUID := uuid.Must(uuid.NewV7())
	expiresAt := time.Now().Add(inviteExpiry)

	invite, err := q.CreateUserInvite(ctx, query.CreateUserInviteParams{
		UUID:              inviteUUID,
		WorkspaceUUID:     &workspaceUUID,
		Email:             req.Email,
		Role:              string(req.Role),
		TokenHash:         tokenHash,
		InvitedByUserUuid: &inviterUUID,
		ExpiresAt:         pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		h.log.Error("failed to create invite", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create invite"))
	}

	// Get inviter name for email
	inviterUser, _ := q.GetUser(ctx, converter.UuidToPgUUID(inviterUUID))
	inviterName := "A team member"
	if inviterUser.FirstName != "" {
		inviterName = inviterUser.FirstName + " " + inviterUser.LastName
	}

	// Build invite link
	inviteLink := fmt.Sprintf("%s/invite/%s", h.cfg.CSRBaseURL, token)

	// Send email
	if err := h.emailService.SendInviteEmail(ctx, email.InviteEmailParams{
		ToEmail:       req.Email,
		InviterName:   inviterName,
		WorkspaceName: workspace.DisplayName,
		Role:          string(req.Role),
		InviteLink:    inviteLink,
		ExpiresIn:     "24 hours",
	}); err != nil {
		h.log.Error("failed to send invite email", "error", err, "email", req.Email)
		// Don't fail the request - invite is created, email just didn't send
	}

	result := dbInviteToAPI(invite, inviterUser.Email, inviterUser.FirstName, inviterUser.LastName)
	return &result, nil
}

// DeleteWorkspaceInvite cancels/deletes an invite.
//
// DELETE /workspace/{uuid}/invites/{invite_uuid}
func (h *Handler) DeleteWorkspaceInvite(ctx context.Context, params api.DeleteWorkspaceInviteParams) (api.DeleteWorkspaceInviteRes, error) {
	q := query.New(h.dbp)
	inviteUUID := converter.GoogleToGofrsUUID(params.InviteUUID)

	// Verify invite exists
	_, err := q.GetUserInvite(ctx, inviteUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("invite not found"))
		}
		h.log.Error("failed to get invite", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get invite"))
	}

	if err := q.DeleteUserInvite(ctx, inviteUUID); err != nil {
		h.log.Error("failed to delete invite", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete invite"))
	}

	return &api.DeleteWorkspaceInviteNoContent{}, nil
}

// GetInviteByToken retrieves invite info for the accept page (public endpoint).
//
// GET /invite/{token}
func (h *Handler) GetInviteByToken(ctx context.Context, params api.GetInviteByTokenParams) (api.GetInviteByTokenRes, error) {
	q := query.New(h.dbp)

	// Hash the token and look up
	tokenHash := hashToken(params.Token)
	invite, err := q.GetValidInviteByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("invite not found or expired"))
		}
		h.log.Error("failed to get invite", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get invite"))
	}

	inviterName := invite.InviterFirstName + " " + invite.InviterLastName
	if invite.InviterFirstName == "" {
		inviterName = invite.InviterEmail
	}

	return &api.UserInviteInfo{
		Email:         invite.Email,
		WorkspaceName: invite.WorkspaceDisplayName,
		WorkspaceSlug: invite.WorkspaceSlug,
		Role:          api.UserInviteInfoRole(invite.Role),
		ExpiresAt:     invite.ExpiresAt.Time,
		InviterName:   api.NewOptString(inviterName),
	}, nil
}

// AcceptInvite accepts an invite and creates the user account.
//
// POST /invite/accept
func (h *Handler) AcceptInvite(ctx context.Context, req *api.UserInviteAccept) (api.AcceptInviteRes, error) {
	q := query.New(h.dbp)

	// Find and validate invite by token
	tokenHash := hashToken(req.Token)
	invite, err := q.GetValidInviteByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("invite not found or expired"))
		}
		h.log.Error("failed to get invite", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get invite"))
	}

	// Check if user already exists with this email
	existingUser, err := q.GetUserByEmail(ctx, invite.Email)
	if err == nil {
		// User exists - just add to workspace
		return h.addExistingUserToWorkspace(ctx, q, existingUser, invite)
	}
	if err != pgx.ErrNoRows {
		h.log.Error("failed to check user", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to check user"))
	}

	// Create new user
	userUUID := uuid.Must(uuid.NewV7())
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("failed to hash password", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to hash password"))
	}

	_, err = q.CreateUser(ctx, query.CreateUserParams{
		UUID:      converter.UuidToPgUUID(userUUID),
		Email:     invite.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsEnabled: true,
		Meta:      []byte("{}"),
	})
	if err != nil {
		h.log.Error("failed to create user", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create user"))
	}

	// Add user to workspace
	memberUUID := uuid.Must(uuid.NewV7())
	_, err = q.CreateWorkspaceMember(ctx, query.CreateWorkspaceMemberParams{
		UUID:          memberUUID,
		WorkspaceUUID: invite.WorkspaceUUID,
		UserUUID:      &userUUID,
		Role:          invite.Role,
	})
	if err != nil {
		h.log.Error("failed to add workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to add to workspace"))
	}

	// Mark invite as accepted
	if err := q.MarkInviteAccepted(ctx, invite.UUID); err != nil {
		h.log.Warn("failed to mark invite accepted", "error", err)
	}

	// Assign workspace role via RBAC
	roleName := rbac.RoleWorkspaceMember
	if invite.Role == "admin" {
		roleName = rbac.RoleWorkspaceAdmin
	}
	if err := h.enforcer.AddRoleForUserInDomain(userUUID.String(), roleName, invite.WorkspaceSlug); err != nil {
		h.log.Warn("failed to assign workspace role", "error", err)
	}

	return &api.AcceptInviteOK{
		RedirectURL: api.NewOptString(h.cfg.CSRBaseURL + "/login"),
	}, nil
}

// addExistingUserToWorkspace adds an existing user to the workspace from an invite.
func (h *Handler) addExistingUserToWorkspace(ctx context.Context, q *query.Queries, user query.User, invite query.GetValidInviteByTokenHashRow) (api.AcceptInviteRes, error) {
	// Check if already a member
	_, err := q.GetWorkspaceMember(ctx, query.GetWorkspaceMemberParams{
		WorkspaceUUID: invite.WorkspaceUUID,
		UserUUID:      &user.UUID,
	})
	if err == nil {
		return nil, ErrWithCode(http.StatusConflict, E("you are already a member of this workspace"))
	}
	if err != pgx.ErrNoRows {
		h.log.Error("failed to check workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to check workspace membership"))
	}

	// Add to workspace
	memberUUID := uuid.Must(uuid.NewV7())
	_, err = q.CreateWorkspaceMember(ctx, query.CreateWorkspaceMemberParams{
		UUID:          memberUUID,
		WorkspaceUUID: invite.WorkspaceUUID,
		UserUUID:      &user.UUID,
		Role:          invite.Role,
	})
	if err != nil {
		h.log.Error("failed to add workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to add to workspace"))
	}

	// Mark invite as accepted
	if err := q.MarkInviteAccepted(ctx, invite.UUID); err != nil {
		h.log.Warn("failed to mark invite accepted", "error", err)
	}

	// Assign workspace role via RBAC
	roleName := rbac.RoleWorkspaceMember
	if invite.Role == "admin" {
		roleName = rbac.RoleWorkspaceAdmin
	}
	if err := h.enforcer.AddRoleForUserInDomain(user.UUID.String(), roleName, invite.WorkspaceSlug); err != nil {
		h.log.Warn("failed to assign workspace role", "error", err)
	}

	return &api.AcceptInviteOK{
		RedirectURL: api.NewOptString(h.cfg.CSRBaseURL + "/login"),
	}, nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashToken creates a SHA256 hash of the token for storage.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// dbInviteToAPI converts a database UserInvite to API UserInvite.
func dbInviteToAPI(inv query.UserInvite, inviterEmail, inviterFirstName, inviterLastName string) api.UserInvite {
	result := api.UserInvite{
		UUID:      api.NewOptUUID(converter.GofrsToGoogleUUID(inv.UUID)),
		Email:     inv.Email,
		Role:      api.UserInviteRole(inv.Role),
		ExpiresAt: api.NewOptDateTime(inv.ExpiresAt.Time),
	}
	if inv.WorkspaceUUID != nil {
		result.WorkspaceUUID = api.NewOptUUID(converter.GofrsToGoogleUUID(*inv.WorkspaceUUID))
	}
	if inv.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(inv.CreatedAt.Time)
	}
	if inv.AcceptedAt.Valid {
		result.AcceptedAt = api.NewOptDateTime(inv.AcceptedAt.Time)
	}
	if inviterEmail != "" {
		result.InvitedByEmail = api.NewOptString(inviterEmail)
	}
	if inviterFirstName != "" {
		result.InvitedByName = api.NewOptString(inviterFirstName + " " + inviterLastName)
	}
	return result
}

// dbInviteRowToAPI converts a ListWorkspaceInvitesRow to API UserInvite.
func dbInviteRowToAPI(inv query.ListWorkspaceInvitesRow) api.UserInvite {
	result := api.UserInvite{
		UUID:      api.NewOptUUID(converter.GofrsToGoogleUUID(inv.UUID)),
		Email:     inv.Email,
		Role:      api.UserInviteRole(inv.Role),
		ExpiresAt: api.NewOptDateTime(inv.ExpiresAt.Time),
	}
	if inv.WorkspaceUUID != nil {
		result.WorkspaceUUID = api.NewOptUUID(converter.GofrsToGoogleUUID(*inv.WorkspaceUUID))
	}
	if inv.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(inv.CreatedAt.Time)
	}
	if inv.AcceptedAt.Valid {
		result.AcceptedAt = api.NewOptDateTime(inv.AcceptedAt.Time)
	}
	if inv.InviterEmail != "" {
		result.InvitedByEmail = api.NewOptString(inv.InviterEmail)
		result.InvitedByName = api.NewOptString(inv.InviterFirstName + " " + inv.InviterLastName)
	}
	return result
}
