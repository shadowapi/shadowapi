package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// ListWorkspaces implements listWorkspaces operation.
//
// List all workspaces accessible to the current user.
//
// GET /workspace
func (h *Handler) ListWorkspaces(ctx context.Context) (api.ListWorkspacesRes, error) {
	q := query.New(h.dbp)

	workspaces, err := q.ListWorkspaces(ctx)
	if err != nil {
		h.log.Error("failed to list workspaces", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list workspaces"))
	}

	result := make([]api.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		result = append(result, qWorkspaceToAPI(w))
	}

	res := api.ListWorkspacesOKApplicationJSON(result)
	return &res, nil
}

// GetWorkspace implements getWorkspace operation.
//
// Get workspace details.
//
// GET /workspace/{uuid}
func (h *Handler) GetWorkspace(ctx context.Context, params api.GetWorkspaceParams) (api.GetWorkspaceRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)
	workspace, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	result := qWorkspaceToAPI(workspace)
	return &result, nil
}

// CreateWorkspace implements createWorkspace operation.
//
// Create a new workspace.
//
// POST /workspace
func (h *Handler) CreateWorkspace(ctx context.Context, req *api.Workspace) (api.CreateWorkspaceRes, error) {
	q := query.New(h.dbp)

	// Generate UUID if not provided
	workspaceUUID := uuid.Must(uuid.NewV7())
	if req.UUID.IsSet() {
		workspaceUUID = converter.GoogleToGofrsUUID(req.UUID.Value)
	}

	// Serialize settings
	settings := []byte("{}")
	if req.Settings.IsSet() {
		var err error
		settings, err = json.Marshal(req.Settings.Value)
		if err != nil {
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid settings format"))
		}
	}

	isEnabled := true
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	workspace, err := q.CreateWorkspace(ctx, query.CreateWorkspaceParams{
		UUID:        workspaceUUID,
		Slug:        req.Slug,
		DisplayName: req.DisplayName,
		IsEnabled:   isEnabled,
		Settings:    settings,
	})
	if err != nil {
		h.log.Error("failed to create workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create workspace"))
	}

	result := qWorkspaceToAPI(workspace)
	return &result, nil
}

// UpdateWorkspace implements updateWorkspace operation.
//
// Update workspace details.
//
// PUT /workspace/{uuid}
func (h *Handler) UpdateWorkspace(ctx context.Context, req *api.Workspace, params api.UpdateWorkspaceParams) (api.UpdateWorkspaceRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)

	// Get existing workspace first
	existing, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	// Use existing values if not provided in request
	displayName := existing.DisplayName
	if req.DisplayName != "" {
		displayName = req.DisplayName
	}

	isEnabled := existing.IsEnabled
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	}

	settings := existing.Settings
	if req.Settings.IsSet() {
		var err error
		settings, err = json.Marshal(req.Settings.Value)
		if err != nil {
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid settings format"))
		}
	}

	workspace, err := q.UpdateWorkspace(ctx, query.UpdateWorkspaceParams{
		UUID:        workspaceUUID,
		DisplayName: displayName,
		IsEnabled:   isEnabled,
		Settings:    settings,
	})
	if err != nil {
		h.log.Error("failed to update workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update workspace"))
	}

	result := qWorkspaceToAPI(workspace)
	return &result, nil
}

// DeleteWorkspace implements deleteWorkspace operation.
//
// Delete a workspace.
//
// DELETE /workspace/{uuid}
func (h *Handler) DeleteWorkspace(ctx context.Context, params api.DeleteWorkspaceParams) (api.DeleteWorkspaceRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)

	// Check if workspace exists
	_, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	err = q.DeleteWorkspace(ctx, workspaceUUID)
	if err != nil {
		h.log.Error("failed to delete workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete workspace"))
	}

	return &api.DeleteWorkspaceNoContent{}, nil
}

// CheckWorkspaceExists implements checkWorkspaceExists operation.
//
// Check if a workspace slug exists.
//
// GET /workspace/check/{slug}
func (h *Handler) CheckWorkspaceExists(ctx context.Context, params api.CheckWorkspaceExistsParams) (api.CheckWorkspaceExistsRes, error) {
	q := query.New(h.dbp)

	workspace, err := q.GetWorkspaceBySlug(ctx, params.Slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &api.WorkspaceCheck{
				Exists: false,
			}, nil
		}
		h.log.Error("failed to check workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to check workspace"))
	}

	return &api.WorkspaceCheck{
		Exists:      true,
		DisplayName: api.NewOptString(workspace.DisplayName),
	}, nil
}

// ListWorkspaceMembers implements listWorkspaceMembers operation.
//
// List all members of a workspace.
//
// GET /workspace/{uuid}/members
func (h *Handler) ListWorkspaceMembers(ctx context.Context, params api.ListWorkspaceMembersParams) (api.ListWorkspaceMembersRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)

	// Check if workspace exists
	_, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	members, err := q.ListWorkspaceMembersByWorkspace(ctx, &workspaceUUID)
	if err != nil {
		h.log.Error("failed to list workspace members", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list workspace members"))
	}

	result := make([]api.WorkspaceMember, 0, len(members))
	for _, m := range members {
		apiMember, err := h.qWorkspaceMemberToAPI(ctx, q, m)
		if err != nil {
			h.log.Warn("failed to convert workspace member", "error", err)
			continue
		}
		result = append(result, apiMember)
	}

	res := api.ListWorkspaceMembersOKApplicationJSON(result)
	return &res, nil
}

// AddWorkspaceMember implements addWorkspaceMember operation.
//
// Add a member to a workspace.
//
// POST /workspace/{uuid}/members
func (h *Handler) AddWorkspaceMember(ctx context.Context, req *api.WorkspaceMember, params api.AddWorkspaceMemberParams) (api.AddWorkspaceMemberRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)
	userUUID := converter.GoogleToGofrsUUID(req.UserUUID)

	// Check if workspace exists
	_, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	// Check if user exists
	userPgUUID := pgtype.UUID{Bytes: userUUID, Valid: true}
	_, err = q.GetUser(ctx, userPgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("user not found"))
		}
		h.log.Error("failed to get user", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get user"))
	}

	// Check if member already exists
	_, err = q.GetWorkspaceMember(ctx, query.GetWorkspaceMemberParams{
		WorkspaceUUID: &workspaceUUID,
		UserUUID:      &userUUID,
	})
	if err == nil {
		return nil, ErrWithCode(http.StatusConflict, E("user is already a member of this workspace"))
	}
	if err != pgx.ErrNoRows {
		h.log.Error("failed to check workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to check workspace member"))
	}

	// Create member
	memberUUID := uuid.Must(uuid.NewV7())
	member, err := q.CreateWorkspaceMember(ctx, query.CreateWorkspaceMemberParams{
		UUID:          memberUUID,
		WorkspaceUUID: &workspaceUUID,
		UserUUID:      &userUUID,
		Role:          string(req.Role),
	})
	if err != nil {
		h.log.Error("failed to add workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to add workspace member"))
	}

	apiMember, err := h.qWorkspaceMemberToAPI(ctx, q, member)
	if err != nil {
		h.log.Error("failed to convert workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get member details"))
	}

	return &apiMember, nil
}

// RemoveWorkspaceMember implements removeWorkspaceMember operation.
//
// Remove a member from a workspace.
//
// DELETE /workspace/{uuid}/members/{user_uuid}
func (h *Handler) RemoveWorkspaceMember(ctx context.Context, params api.RemoveWorkspaceMemberParams) (api.RemoveWorkspaceMemberRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)
	userUUID := converter.GoogleToGofrsUUID(params.UserUUID)

	// Check if workspace exists
	_, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	// Check if member exists
	member, err := q.GetWorkspaceMember(ctx, query.GetWorkspaceMemberParams{
		WorkspaceUUID: &workspaceUUID,
		UserUUID:      &userUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("member not found"))
		}
		h.log.Error("failed to get workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace member"))
	}

	// Prevent removing the last owner
	if member.Role == "owner" {
		members, err := q.ListWorkspaceMembersByWorkspace(ctx, &workspaceUUID)
		if err != nil {
			h.log.Error("failed to list workspace members", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list workspace members"))
		}

		ownerCount := 0
		for _, m := range members {
			if m.Role == "owner" {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return nil, ErrWithCode(http.StatusBadRequest, E("cannot remove the last owner of a workspace"))
		}
	}

	err = q.DeleteWorkspaceMember(ctx, query.DeleteWorkspaceMemberParams{
		WorkspaceUUID: &workspaceUUID,
		UserUUID:      &userUUID,
	})
	if err != nil {
		h.log.Error("failed to remove workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to remove workspace member"))
	}

	return &api.RemoveWorkspaceMemberNoContent{}, nil
}

// UpdateWorkspaceMemberRole implements updateWorkspaceMemberRole operation.
//
// Update a member's role in a workspace.
//
// PUT /workspace/{uuid}/members/{user_uuid}
func (h *Handler) UpdateWorkspaceMemberRole(ctx context.Context, req *api.UpdateWorkspaceMemberRoleReq, params api.UpdateWorkspaceMemberRoleParams) (api.UpdateWorkspaceMemberRoleRes, error) {
	q := query.New(h.dbp)

	workspaceUUID := converter.GoogleToGofrsUUID(params.UUID)
	userUUID := converter.GoogleToGofrsUUID(params.UserUUID)

	// Check if workspace exists
	_, err := q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("workspace not found"))
		}
		h.log.Error("failed to get workspace", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace"))
	}

	// Check if member exists and get current role
	currentMember, err := q.GetWorkspaceMember(ctx, query.GetWorkspaceMemberParams{
		WorkspaceUUID: &workspaceUUID,
		UserUUID:      &userUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("member not found"))
		}
		h.log.Error("failed to get workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get workspace member"))
	}

	// Prevent demoting the last owner
	if currentMember.Role == "owner" && string(req.Role) != "owner" {
		members, err := q.ListWorkspaceMembersByWorkspace(ctx, &workspaceUUID)
		if err != nil {
			h.log.Error("failed to list workspace members", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list workspace members"))
		}

		ownerCount := 0
		for _, m := range members {
			if m.Role == "owner" {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return nil, ErrWithCode(http.StatusBadRequest, E("cannot demote the last owner of a workspace"))
		}
	}

	member, err := q.UpdateWorkspaceMemberRole(ctx, query.UpdateWorkspaceMemberRoleParams{
		WorkspaceUUID: &workspaceUUID,
		UserUUID:      &userUUID,
		Role:          string(req.Role),
	})
	if err != nil {
		h.log.Error("failed to update workspace member role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update workspace member role"))
	}

	apiMember, err := h.qWorkspaceMemberToAPI(ctx, q, member)
	if err != nil {
		h.log.Error("failed to convert workspace member", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get member details"))
	}

	return &apiMember, nil
}

// Helper function to convert query.Workspace to api.Workspace
func qWorkspaceToAPI(w query.Workspace) api.Workspace {
	result := api.Workspace{
		UUID:        api.NewOptUUID(converter.GofrsToGoogleUUID(w.UUID)),
		Slug:        w.Slug,
		DisplayName: w.DisplayName,
		IsEnabled:   api.NewOptBool(w.IsEnabled),
	}

	// Parse settings
	if len(w.Settings) > 0 {
		var settings api.WorkspaceSettings
		if err := json.Unmarshal(w.Settings, &settings); err == nil {
			result.Settings = api.NewOptWorkspaceSettings(settings)
		}
	}

	if w.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(w.CreatedAt.Time)
	}
	if w.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(w.UpdatedAt.Time)
	}

	return result
}

// Helper function to convert query.WorkspaceMember to api.WorkspaceMember
func (h *Handler) qWorkspaceMemberToAPI(ctx context.Context, q *query.Queries, m query.WorkspaceMember) (api.WorkspaceMember, error) {
	result := api.WorkspaceMember{
		UUID: api.NewOptUUID(converter.GofrsToGoogleUUID(m.UUID)),
		Role: api.WorkspaceMemberRole(m.Role),
	}

	if m.WorkspaceUUID != nil {
		result.WorkspaceUUID = converter.GofrsToGoogleUUID(*m.WorkspaceUUID)
	}

	if m.UserUUID != nil {
		result.UserUUID = converter.GofrsToGoogleUUID(*m.UserUUID)

		// Fetch user details
		userPgUUID := pgtype.UUID{Bytes: converter.GofrsToGoogleUUID(*m.UserUUID), Valid: true}
		user, err := q.GetUser(ctx, userPgUUID)
		if err == nil {
			result.UserEmail = api.NewOptString(user.Email)
			result.UserFirstName = api.NewOptString(user.FirstName)
			result.UserLastName = api.NewOptString(user.LastName)
		}
	}

	if m.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(m.CreatedAt.Time)
	}
	if m.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(m.UpdatedAt.Time)
	}

	return result, nil
}
