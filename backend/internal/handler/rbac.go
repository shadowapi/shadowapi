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

// ListPolicySets implements listPolicySets operation.
//
// List all policy sets.
//
// GET /rbac/policy-set
func (h *Handler) ListPolicySets(ctx context.Context, params api.ListPolicySetsParams) (api.ListPolicySetsRes, error) {
	q := query.New(h.dbp)

	var policySets []query.PolicySet
	var err error

	if params.Scope.IsSet() {
		policySets, err = q.ListPolicySetsByScope(ctx, string(params.Scope.Value))
	} else {
		policySets, err = q.ListPolicySets(ctx)
	}

	if err != nil {
		h.log.Error("failed to list policy sets", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list policy sets"))
	}

	result := make([]api.PolicySet, 0, len(policySets))
	for _, ps := range policySets {
		result = append(result, qPolicySetToAPI(ps))
	}

	res := api.ListPolicySetsOKApplicationJSON(result)
	return &res, nil
}

// CreatePolicySet implements createPolicySet operation.
//
// Create a new policy set.
//
// POST /rbac/policy-set
func (h *Handler) CreatePolicySet(ctx context.Context, req *api.PolicySet) (api.CreatePolicySetRes, error) {
	q := query.New(h.dbp)

	// Generate UUID
	policySetUUID := uuid.Must(uuid.NewV7())

	// Serialize permissions
	permsJSON := []byte("[]")
	if len(req.Permissions) > 0 {
		var err error
		permsJSON, err = json.Marshal(req.Permissions)
		if err != nil {
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid permissions format"))
		}
	}

	// Create policy set in database
	policySet, err := q.CreatePolicySet(ctx, query.CreatePolicySetParams{
		UUID:        pgtype.UUID{Bytes: policySetUUID, Valid: true},
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: pgtype.Text{String: req.Description.Value, Valid: req.Description.IsSet()},
		Scope:       string(req.Scope),
		IsSystem:    false, // User-created policy sets are not system policy sets
		Permissions: permsJSON,
	})
	if err != nil {
		h.log.Error("failed to create policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create policy set"))
	}

	// Add Ladon policies for the new policy set
	for _, perm := range req.Permissions {
		domain := "*"
		if req.Scope == api.PolicySetScopeGlobal {
			domain = "global"
		}
		if err := h.enforcer.AddPolicy(policySet.Name, domain, perm.Resource, string(perm.Action)); err != nil {
			h.log.Debug("policy may already exist", "policy_set", policySet.Name)
		}
	}

	result := qPolicySetToAPI(policySet)
	return &result, nil
}

// GetPolicySet implements getPolicySet operation.
//
// Get policy set details.
//
// GET /rbac/policy-set/{uuid}
func (h *Handler) GetPolicySet(ctx context.Context, params api.GetPolicySetParams) (api.GetPolicySetRes, error) {
	q := query.New(h.dbp)

	policySetUUID := pgtype.UUID{Bytes: params.UUID, Valid: true}
	policySet, err := q.GetPolicySetByUUID(ctx, policySetUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("policy set not found"))
		}
		h.log.Error("failed to get policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get policy set"))
	}

	result := qPolicySetToAPI(policySet)
	return &result, nil
}

// UpdatePolicySet implements updatePolicySet operation.
//
// Update a policy set.
//
// PUT /rbac/policy-set/{uuid}
func (h *Handler) UpdatePolicySet(ctx context.Context, req *api.PolicySet, params api.UpdatePolicySetParams) (api.UpdatePolicySetRes, error) {
	q := query.New(h.dbp)

	policySetUUID := pgtype.UUID{Bytes: params.UUID, Valid: true}

	// Check if policy set exists and is not a system policy set
	existingPolicySet, err := q.GetPolicySetByUUID(ctx, policySetUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("policy set not found"))
		}
		h.log.Error("failed to get policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get policy set"))
	}

	if existingPolicySet.IsSystem {
		return nil, ErrWithCode(http.StatusForbidden, E("cannot modify system policy set"))
	}

	// Serialize permissions
	permsJSON := []byte("[]")
	if len(req.Permissions) > 0 {
		permsJSON, err = json.Marshal(req.Permissions)
		if err != nil {
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid permissions format"))
		}
	}

	policySet, err := q.UpdatePolicySet(ctx, query.UpdatePolicySetParams{
		UUID:        policySetUUID,
		DisplayName: req.DisplayName,
		Description: pgtype.Text{String: req.Description.Value, Valid: req.Description.IsSet()},
		Permissions: permsJSON,
	})
	if err != nil {
		h.log.Error("failed to update policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update policy set"))
	}

	result := qPolicySetToAPI(policySet)
	return &result, nil
}

// DeletePolicySet implements deletePolicySet operation.
//
// Delete a policy set.
//
// DELETE /rbac/policy-set/{uuid}
func (h *Handler) DeletePolicySet(ctx context.Context, params api.DeletePolicySetParams) (api.DeletePolicySetRes, error) {
	q := query.New(h.dbp)

	policySetUUID := pgtype.UUID{Bytes: params.UUID, Valid: true}

	// Check if policy set exists and is not a system policy set
	policySet, err := q.GetPolicySetByUUID(ctx, policySetUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("policy set not found"))
		}
		h.log.Error("failed to get policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get policy set"))
	}

	if policySet.IsSystem {
		return nil, ErrWithCode(http.StatusForbidden, E("cannot delete system policy set"))
	}

	if err := q.DeletePolicySet(ctx, policySetUUID); err != nil {
		h.log.Error("failed to delete policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete policy set"))
	}

	return &api.DeletePolicySetNoContent{}, nil
}

// ListPermissions implements listPermissions operation.
//
// List all permissions.
//
// GET /rbac/permission
func (h *Handler) ListPermissions(ctx context.Context, params api.ListPermissionsParams) (api.ListPermissionsRes, error) {
	q := query.New(h.dbp)

	var perms []query.Permission
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

	result := make([]api.Permission, 0, len(perms))
	for _, p := range perms {
		result = append(result, qPermissionToAPI(p))
	}

	res := api.ListPermissionsOKApplicationJSON(result)
	return &res, nil
}

// GetUserPolicySets implements getUserPolicySets operation.
//
// Get policy sets for a user.
//
// GET /rbac/user/{user_uuid}/policy-sets
func (h *Handler) GetUserPolicySets(ctx context.Context, params api.GetUserPolicySetsParams) (api.GetUserPolicySetsRes, error) {
	q := query.New(h.dbp)

	userUUID := params.UserUUID.String()
	var policySetsMap map[string][]string
	var err error

	if params.Domain.IsSet() {
		// Get policy sets for specific domain
		policySetNames := h.enforcer.GetPolicySetsForUser(userUUID, params.Domain.Value)
		policySetsMap = map[string][]string{params.Domain.Value: policySetNames}
	} else {
		// Get all policy sets across all domains
		policySetsMap, err = h.enforcer.GetAllPolicySetsForUser(userUUID)
		if err != nil {
			h.log.Error("failed to get user policy sets", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get user policy sets"))
		}
	}

	// Convert to API format
	var assignments []api.UserPolicySetAssignment
	for domain, policySetNames := range policySetsMap {
		for _, policySetName := range policySetNames {
			// Get policy set details from database
			_, err := q.GetPolicySetByName(ctx, policySetName)
			if err != nil {
				h.log.Debug("policy set not found in database", "policy_set", policySetName, "error", err)
				continue
			}

			assignments = append(assignments, api.UserPolicySetAssignment{
				UserUUID:  params.UserUUID,
				PolicySet: policySetName,
				Domain:    domain,
			})
		}
	}

	return &api.GetUserPolicySetsOK{
		UserUUID:   api.NewOptUUID(params.UserUUID),
		PolicySets: assignments,
	}, nil
}

// AssignPolicySetToUser implements assignPolicySetToUser operation.
//
// Assign a policy set to a user.
//
// POST /rbac/user/{user_uuid}/policy-sets
func (h *Handler) AssignPolicySetToUser(ctx context.Context, req *api.AssignPolicySetToUserReq, params api.AssignPolicySetToUserParams) (api.AssignPolicySetToUserRes, error) {
	q := query.New(h.dbp)

	// Verify policy set exists
	_, err := q.GetPolicySetByName(ctx, req.PolicySetName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("policy set not found"))
		}
		h.log.Error("failed to get policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get policy set"))
	}

	userUUID := params.UserUUID.String()
	if err := h.enforcer.AssignPolicySetToUser(userUUID, req.PolicySetName, req.Domain); err != nil {
		h.log.Error("failed to assign policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to assign policy set"))
	}

	return &api.AssignPolicySetToUserCreated{}, nil
}

// RemovePolicySetFromUser implements removePolicySetFromUser operation.
//
// Remove a policy set from a user.
//
// DELETE /rbac/user/{user_uuid}/policy-sets/{policy_set_name}
func (h *Handler) RemovePolicySetFromUser(ctx context.Context, params api.RemovePolicySetFromUserParams) (api.RemovePolicySetFromUserRes, error) {
	userUUID := params.UserUUID.String()

	if err := h.enforcer.RemovePolicySetFromUser(userUUID, params.PolicySetName, params.Domain); err != nil {
		h.log.Error("failed to remove policy set", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to remove policy set"))
	}

	return &api.RemovePolicySetFromUserNoContent{}, nil
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

func qPolicySetToAPI(ps query.PolicySet) api.PolicySet {
	// Parse permissions from JSON
	var perms []api.Permission
	if len(ps.Permissions) > 0 {
		_ = json.Unmarshal(ps.Permissions, &perms)
	}

	result := api.PolicySet{
		UUID:        api.NewOptUUID(googleuuid.UUID(ps.UUID)),
		Name:        ps.Name,
		DisplayName: ps.DisplayName,
		Scope:       api.PolicySetScope(ps.Scope),
		IsSystem:    api.NewOptBool(ps.IsSystem),
		Permissions: perms,
	}

	if ps.Description.Valid {
		result.Description = api.NewOptString(ps.Description.String)
	}
	if ps.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(ps.CreatedAt.Time)
	}
	if ps.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(ps.UpdatedAt.Time)
	}

	return result
}

func qPermissionToAPI(p query.Permission) api.Permission {
	result := api.Permission{
		UUID:        api.NewOptUUID(googleuuid.UUID(p.UUID)),
		Name:        p.Name,
		DisplayName: api.NewOptString(p.DisplayName),
		Resource:    p.Resource,
		Action:      api.PermissionAction(p.Action),
		Scope:       api.NewOptPermissionScope(api.PermissionScope(p.Scope)),
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
