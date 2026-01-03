package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	googleuuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// ListRoles implements listRoles operation.
//
// List all roles.
//
// GET /rbac/role
func (h *Handler) ListRoles(ctx context.Context, params api.ListRolesParams) (api.ListRolesRes, error) {
	q := query.New(h.dbp)

	var roles []query.RbacRole
	var err error

	if params.Scope.IsSet() {
		roles, err = q.ListRolesByScope(ctx, string(params.Scope.Value))
	} else {
		roles, err = q.ListRoles(ctx)
	}

	if err != nil {
		h.log.Error("failed to list roles", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list roles"))
	}

	result := make([]api.RbacRole, 0, len(roles))
	for _, r := range roles {
		result = append(result, qRoleToAPI(r))
	}

	res := api.ListRolesOKApplicationJSON(result)
	return &res, nil
}

// CreateRole implements createRole operation.
//
// Create a new role.
//
// POST /rbac/role
func (h *Handler) CreateRole(ctx context.Context, req *api.RbacRole) (api.CreateRoleRes, error) {
	q := query.New(h.dbp)

	// Generate UUID
	roleUUID := uuid.Must(uuid.NewV7())

	// Serialize permissions
	permsJSON := []byte("[]")
	if len(req.Permissions) > 0 {
		var err error
		permsJSON, err = json.Marshal(req.Permissions)
		if err != nil {
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid permissions format"))
		}
	}

	// Create role in database
	role, err := q.CreateRole(ctx, query.CreateRoleParams{
		UUID:        pgtype.UUID{Bytes: roleUUID, Valid: true},
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: pgtype.Text{String: req.Description.Value, Valid: req.Description.IsSet()},
		Scope:       string(req.Scope),
		IsSystem:    false, // User-created roles are not system roles
		Permissions: permsJSON,
	})
	if err != nil {
		h.log.Error("failed to create role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create role"))
	}

	// Add Casbin policies for the new role
	for _, perm := range req.Permissions {
		domain := "*"
		if req.Scope == api.RbacRoleScopeGlobal {
			domain = "global"
		}
		if err := h.enforcer.AddPolicy(role.Name, domain, perm.Resource, string(perm.Action)); err != nil {
			h.log.Debug("policy may already exist", "role", role.Name)
		}
	}

	result := qRoleToAPI(role)
	return &result, nil
}

// GetRole implements getRole operation.
//
// Get role details.
//
// GET /rbac/role/{uuid}
func (h *Handler) GetRole(ctx context.Context, params api.GetRoleParams) (api.GetRoleRes, error) {
	q := query.New(h.dbp)

	roleUUID := pgtype.UUID{Bytes: params.UUID, Valid: true}
	role, err := q.GetRoleByUUID(ctx, roleUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("role not found"))
		}
		h.log.Error("failed to get role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get role"))
	}

	result := qRoleToAPI(role)
	return &result, nil
}

// UpdateRole implements updateRole operation.
//
// Update a role.
//
// PUT /rbac/role/{uuid}
func (h *Handler) UpdateRole(ctx context.Context, req *api.RbacRole, params api.UpdateRoleParams) (api.UpdateRoleRes, error) {
	q := query.New(h.dbp)

	roleUUID := pgtype.UUID{Bytes: params.UUID, Valid: true}

	// Check if role exists and is not a system role
	existingRole, err := q.GetRoleByUUID(ctx, roleUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("role not found"))
		}
		h.log.Error("failed to get role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get role"))
	}

	if existingRole.IsSystem {
		return nil, ErrWithCode(http.StatusForbidden, E("cannot modify system role"))
	}

	// Serialize permissions
	permsJSON := []byte("[]")
	if len(req.Permissions) > 0 {
		permsJSON, err = json.Marshal(req.Permissions)
		if err != nil {
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid permissions format"))
		}
	}

	role, err := q.UpdateRole(ctx, query.UpdateRoleParams{
		UUID:        roleUUID,
		DisplayName: req.DisplayName,
		Description: pgtype.Text{String: req.Description.Value, Valid: req.Description.IsSet()},
		Permissions: permsJSON,
	})
	if err != nil {
		h.log.Error("failed to update role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update role"))
	}

	result := qRoleToAPI(role)
	return &result, nil
}

// DeleteRole implements deleteRole operation.
//
// Delete a role.
//
// DELETE /rbac/role/{uuid}
func (h *Handler) DeleteRole(ctx context.Context, params api.DeleteRoleParams) (api.DeleteRoleRes, error) {
	q := query.New(h.dbp)

	roleUUID := pgtype.UUID{Bytes: params.UUID, Valid: true}

	// Check if role exists and is not a system role
	role, err := q.GetRoleByUUID(ctx, roleUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("role not found"))
		}
		h.log.Error("failed to get role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get role"))
	}

	if role.IsSystem {
		return nil, ErrWithCode(http.StatusForbidden, E("cannot delete system role"))
	}

	if err := q.DeleteRole(ctx, roleUUID); err != nil {
		h.log.Error("failed to delete role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete role"))
	}

	return &api.DeleteRoleNoContent{}, nil
}

// ListPermissions implements listPermissions operation.
//
// List all permissions.
//
// GET /rbac/permission
func (h *Handler) ListPermissions(ctx context.Context, params api.ListPermissionsParams) (api.ListPermissionsRes, error) {
	q := query.New(h.dbp)

	var perms []query.RbacPermission
	var err error

	if params.Scope.IsSet() {
		perms, err = q.ListPermissionsByScope(ctx, string(params.Scope.Value))
	} else if params.Resource.IsSet() {
		perms, err = q.ListPermissionsByResource(ctx, params.Resource.Value)
	} else {
		perms, err = q.ListPermissions(ctx)
	}

	if err != nil {
		h.log.Error("failed to list permissions", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list permissions"))
	}

	result := make([]api.RbacPermission, 0, len(perms))
	for _, p := range perms {
		result = append(result, qPermissionToAPI(p))
	}

	res := api.ListPermissionsOKApplicationJSON(result)
	return &res, nil
}

// GetUserRoles implements getUserRoles operation.
//
// Get roles for a user.
//
// GET /rbac/user/{user_uuid}/roles
func (h *Handler) GetUserRoles(ctx context.Context, params api.GetUserRolesParams) (api.GetUserRolesRes, error) {
	q := query.New(h.dbp)

	userUUID := params.UserUUID.String()
	var roles map[string][]string
	var err error

	if params.Domain.IsSet() {
		// Get roles for specific domain
		roleNames := h.enforcer.GetRolesForUserInDomain(userUUID, params.Domain.Value)
		roles = map[string][]string{params.Domain.Value: roleNames}
	} else {
		// Get all roles across all domains
		roles, err = h.enforcer.GetAllRolesForUser(userUUID)
		if err != nil {
			h.log.Error("failed to get user roles", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get user roles"))
		}
	}

	// Convert to API format
	var assignments []api.RbacRoleAssignment
	for domain, roleNames := range roles {
		for _, roleName := range roleNames {
			// Get role details from database
			role, err := q.GetRoleByName(ctx, roleName)
			if err != nil {
				h.log.Debug("role not found in database", "role", roleName, "error", err)
				continue
			}

			assignments = append(assignments, api.RbacRoleAssignment{
				UserUUID: params.UserUUID,
				Role:     qRoleToAPI(role),
				Domain:   domain,
			})
		}
	}

	return &api.GetUserRolesOK{
		UserUUID: api.NewOptUUID(params.UserUUID),
		Roles:    assignments,
	}, nil
}

// AssignRoleToUser implements assignRoleToUser operation.
//
// Assign a role to a user.
//
// POST /rbac/user/{user_uuid}/roles
func (h *Handler) AssignRoleToUser(ctx context.Context, req *api.AssignRoleToUserReq, params api.AssignRoleToUserParams) (api.AssignRoleToUserRes, error) {
	q := query.New(h.dbp)

	// Verify role exists
	_, err := q.GetRoleByName(ctx, req.RoleName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("role not found"))
		}
		h.log.Error("failed to get role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get role"))
	}

	userUUID := params.UserUUID.String()
	if err := h.enforcer.AddRoleForUserInDomain(userUUID, req.RoleName, req.Domain); err != nil {
		h.log.Error("failed to assign role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to assign role"))
	}

	return &api.AssignRoleToUserCreated{}, nil
}

// RemoveRoleFromUser implements removeRoleFromUser operation.
//
// Remove a role from a user.
//
// DELETE /rbac/user/{user_uuid}/roles/{role_name}
func (h *Handler) RemoveRoleFromUser(ctx context.Context, params api.RemoveRoleFromUserParams) (api.RemoveRoleFromUserRes, error) {
	userUUID := params.UserUUID.String()

	if err := h.enforcer.RemoveRoleForUserInDomain(userUUID, params.RoleName, params.Domain); err != nil {
		h.log.Error("failed to remove role", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to remove role"))
	}

	return &api.RemoveRoleFromUserNoContent{}, nil
}

// CheckPermission implements checkPermission operation.
//
// Check if a user has permission.
//
// POST /rbac/check
func (h *Handler) CheckPermission(ctx context.Context, req *api.CheckPermissionReq) (api.CheckPermissionRes, error) {
	allowed, err := h.enforcer.Enforce(
		req.UserUUID.String(),
		req.Domain,
		req.Resource,
		req.Action,
	)
	if err != nil {
		h.log.Error("failed to check permission", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("permission check failed"))
	}

	return &api.CheckPermissionOK{
		Allowed: api.NewOptBool(allowed),
	}, nil
}

// Helper functions

func qRoleToAPI(r query.RbacRole) api.RbacRole {
	// Parse permissions from JSON
	var perms []api.RbacPermission
	if len(r.Permissions) > 0 {
		_ = json.Unmarshal(r.Permissions, &perms)
	}

	result := api.RbacRole{
		UUID:        api.NewOptUUID(googleuuid.UUID(r.UUID)),
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Scope:       api.RbacRoleScope(r.Scope),
		IsSystem:    api.NewOptBool(r.IsSystem),
		Permissions: perms,
	}

	if r.Description.Valid {
		result.Description = api.NewOptString(r.Description.String)
	}
	if r.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(r.CreatedAt.Time)
	}
	if r.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(r.UpdatedAt.Time)
	}

	return result
}

func qPermissionToAPI(p query.RbacPermission) api.RbacPermission {
	result := api.RbacPermission{
		UUID:        api.NewOptUUID(googleuuid.UUID(p.UUID)),
		Name:        p.Name,
		DisplayName: api.NewOptString(p.DisplayName),
		Resource:    p.Resource,
		Action:      api.RbacPermissionAction(p.Action),
		Scope:       api.NewOptRbacPermissionScope(api.RbacPermissionScope(p.Scope)),
	}

	if p.Description.Valid {
		result.Description = api.NewOptString(p.Description.String)
	}
	if p.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(p.CreatedAt.Time)
	}

	return result
}

// Ensure converter is used (for other functions that might need it)
var _ = converter.GoogleToGofrsUUID
