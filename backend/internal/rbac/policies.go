package rbac

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/ladon"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Policy set names
const (
	PolicySetSuperAdmin      = "super_admin"
	PolicySetWorkspaceOwner  = "workspace_owner"
	PolicySetWorkspaceAdmin  = "workspace_admin"
	PolicySetWorkspaceMember = "workspace_member"
)

// PredefinedPolicySets contains all system policy sets that are created on startup.
var PredefinedPolicySets = []PredefinedPolicySet{
	{
		Name:        PolicySetSuperAdmin,
		DisplayName: "Super Admin",
		Description: "Full system access including user management and all workspaces",
		Scope:       ScopeGlobal,
		Permissions: []Permission{
			{Resource: "*", Action: "*"}, // Full access to everything
		},
	},
	{
		Name:        PolicySetWorkspaceOwner,
		DisplayName: "Workspace Owner",
		Description: "Full control over the workspace including member management",
		Scope:       ScopeWorkspace,
		Permissions: []Permission{
			{Resource: ResourceWorkspace, Action: ActionAdmin},
			{Resource: ResourceDatasource, Action: "*"},
			{Resource: ResourcePipeline, Action: "*"},
			{Resource: ResourceStorage, Action: "*"},
			{Resource: ResourceContact, Action: "*"},
			{Resource: ResourceMessage, Action: "*"},
			{Resource: ResourceScheduler, Action: "*"},
			{Resource: ResourceMember, Action: "*"},
		},
	},
	{
		Name:        PolicySetWorkspaceAdmin,
		DisplayName: "Workspace Admin",
		Description: "Can manage workspace resources but not members",
		Scope:       ScopeWorkspace,
		Permissions: []Permission{
			{Resource: ResourceWorkspace, Action: ActionRead},
			{Resource: ResourceDatasource, Action: "*"},
			{Resource: ResourcePipeline, Action: "*"},
			{Resource: ResourceStorage, Action: "*"},
			{Resource: ResourceContact, Action: "*"},
			{Resource: ResourceMessage, Action: "*"},
			{Resource: ResourceScheduler, Action: "*"},
			{Resource: ResourceMember, Action: ActionRead},
		},
	},
	{
		Name:        PolicySetWorkspaceMember,
		DisplayName: "Workspace Member",
		Description: "Read-only access to workspace resources",
		Scope:       ScopeWorkspace,
		Permissions: []Permission{
			{Resource: ResourceWorkspace, Action: ActionRead},
			{Resource: ResourceDatasource, Action: ActionRead},
			{Resource: ResourcePipeline, Action: ActionRead},
			{Resource: ResourceStorage, Action: ActionRead},
			{Resource: ResourceContact, Action: ActionRead},
			{Resource: ResourceMessage, Action: ActionRead},
			{Resource: ResourceScheduler, Action: ActionRead},
			{Resource: ResourceMember, Action: ActionRead},
		},
	},
}

// PredefinedPermissions contains all permissions in the system.
var PredefinedPermissions = []struct {
	Name        string
	DisplayName string
	Description string
	Resource    Resource
	Action      Action
	Scope       Scope
}{
	// Global permissions
	{"user:read", "View Users", "View user list and details", ResourceUser, ActionRead, ScopeGlobal},
	{"user:create", "Create Users", "Create new users", ResourceUser, ActionCreate, ScopeGlobal},
	{"user:write", "Edit Users", "Edit user details", ResourceUser, ActionWrite, ScopeGlobal},
	{"user:delete", "Delete Users", "Delete users", ResourceUser, ActionDelete, ScopeGlobal},
	{"workspace:create", "Create Workspace", "Create new workspaces", ResourceWorkspace, ActionCreate, ScopeGlobal},
	{"policy_set:read", "View Policy Sets", "View policy sets and permissions", ResourcePolicySet, ActionRead, ScopeGlobal},
	{"policy_set:write", "Manage Policy Sets", "Create and modify policy sets", ResourcePolicySet, ActionWrite, ScopeGlobal},
	{"rbac:admin", "RBAC Admin", "Full RBAC administration", ResourceRBAC, ActionAdmin, ScopeGlobal},

	// Workspace permissions
	{"workspace:read", "View Workspace", "View workspace details", ResourceWorkspace, ActionRead, ScopeWorkspace},
	{"workspace:write", "Edit Workspace", "Edit workspace settings", ResourceWorkspace, ActionWrite, ScopeWorkspace},
	{"workspace:delete", "Delete Workspace", "Delete workspace", ResourceWorkspace, ActionDelete, ScopeWorkspace},
	{"workspace:admin", "Admin Workspace", "Full workspace administration", ResourceWorkspace, ActionAdmin, ScopeWorkspace},

	{"datasource:read", "View Datasources", "View datasource list and details", ResourceDatasource, ActionRead, ScopeWorkspace},
	{"datasource:create", "Create Datasources", "Create new datasources", ResourceDatasource, ActionCreate, ScopeWorkspace},
	{"datasource:write", "Edit Datasources", "Edit datasource settings", ResourceDatasource, ActionWrite, ScopeWorkspace},
	{"datasource:delete", "Delete Datasources", "Delete datasources", ResourceDatasource, ActionDelete, ScopeWorkspace},

	{"pipeline:read", "View Pipelines", "View pipeline list and details", ResourcePipeline, ActionRead, ScopeWorkspace},
	{"pipeline:create", "Create Pipelines", "Create new pipelines", ResourcePipeline, ActionCreate, ScopeWorkspace},
	{"pipeline:write", "Edit Pipelines", "Edit pipeline settings", ResourcePipeline, ActionWrite, ScopeWorkspace},
	{"pipeline:delete", "Delete Pipelines", "Delete pipelines", ResourcePipeline, ActionDelete, ScopeWorkspace},

	{"storage:read", "View Storages", "View storage list and details", ResourceStorage, ActionRead, ScopeWorkspace},
	{"storage:create", "Create Storages", "Create new storages", ResourceStorage, ActionCreate, ScopeWorkspace},
	{"storage:write", "Edit Storages", "Edit storage settings", ResourceStorage, ActionWrite, ScopeWorkspace},
	{"storage:delete", "Delete Storages", "Delete storages", ResourceStorage, ActionDelete, ScopeWorkspace},

	{"contact:read", "View Contacts", "View contact list and details", ResourceContact, ActionRead, ScopeWorkspace},
	{"contact:create", "Create Contacts", "Create new contacts", ResourceContact, ActionCreate, ScopeWorkspace},
	{"contact:write", "Edit Contacts", "Edit contact details", ResourceContact, ActionWrite, ScopeWorkspace},
	{"contact:delete", "Delete Contacts", "Delete contacts", ResourceContact, ActionDelete, ScopeWorkspace},

	{"message:read", "View Messages", "View messages", ResourceMessage, ActionRead, ScopeWorkspace},
	{"message:create", "Create Messages", "Send messages", ResourceMessage, ActionCreate, ScopeWorkspace},
	{"message:delete", "Delete Messages", "Delete messages", ResourceMessage, ActionDelete, ScopeWorkspace},

	{"scheduler:read", "View Schedulers", "View scheduler list and details", ResourceScheduler, ActionRead, ScopeWorkspace},
	{"scheduler:create", "Create Schedulers", "Create new schedulers", ResourceScheduler, ActionCreate, ScopeWorkspace},
	{"scheduler:write", "Edit Schedulers", "Edit scheduler settings", ResourceScheduler, ActionWrite, ScopeWorkspace},
	{"scheduler:delete", "Delete Schedulers", "Delete schedulers", ResourceScheduler, ActionDelete, ScopeWorkspace},

	{"member:read", "View Members", "View workspace members", ResourceMember, ActionRead, ScopeWorkspace},
	{"member:create", "Add Members", "Add members to workspace", ResourceMember, ActionCreate, ScopeWorkspace},
	{"member:write", "Edit Members", "Edit member policy sets", ResourceMember, ActionWrite, ScopeWorkspace},
	{"member:delete", "Remove Members", "Remove members from workspace", ResourceMember, ActionDelete, ScopeWorkspace},

	// Worker permissions (global)
	{"worker:read", "View Workers", "View registered workers and enrollment tokens", ResourceWorker, ActionRead, ScopeGlobal},
	{"worker:create", "Create Workers", "Create enrollment tokens for workers", ResourceWorker, ActionCreate, ScopeGlobal},
	{"worker:write", "Edit Workers", "Edit worker settings", ResourceWorker, ActionWrite, ScopeGlobal},
	{"worker:delete", "Delete Workers", "Delete workers and enrollment tokens", ResourceWorker, ActionDelete, ScopeGlobal},
}

// initializePredefinedPolicies creates the predefined policy sets and permissions in the database.
// It creates both the rbac_policy_set/rbac_permission records for API display and the
// Ladon policies for actual enforcement.
func (e *Enforcer) initializePredefinedPolicies(ctx context.Context, dbp *pgxpool.Pool) error {
	q := query.New(dbp)

	// Initialize predefined policy sets
	for _, policySet := range PredefinedPolicySets {
		exists, err := q.PolicySetExists(ctx, policySet.Name)
		if err != nil {
			e.log.Error("failed to check policy set existence", "policy_set", policySet.Name, "error", err)
			continue
		}

		if !exists {
			permJSON, _ := json.Marshal(policySet.Permissions)
			policySetUUID := uuid.Must(uuid.NewV7())

			_, err := q.CreatePolicySet(ctx, query.CreatePolicySetParams{
				UUID:        pgtype.UUID{Bytes: policySetUUID, Valid: true},
				Name:        policySet.Name,
				DisplayName: policySet.DisplayName,
				Description: pgtype.Text{String: policySet.Description, Valid: true},
				Scope:       string(policySet.Scope),
				IsSystem:    true,
				Permissions: permJSON,
			})
			if err != nil {
				e.log.Error("failed to create policy set", "policy_set", policySet.Name, "error", err)
				continue
			}
			e.log.Info("created predefined policy set", "policy_set", policySet.Name)
		}

		// Add Ladon policies for this policy set
		for _, perm := range policySet.Permissions {
			domain := "*" // Workspace policy sets use wildcard domain, actual domain is determined at runtime
			if policySet.Scope == ScopeGlobal {
				domain = "global"
			}
			if err := e.addPolicyUnlocked(policySet.Name, domain, string(perm.Resource), string(perm.Action)); err != nil {
				e.log.Debug("policy may already exist", "policy_set", policySet.Name, "resource", perm.Resource, "action", perm.Action)
			}
		}
	}

	// Initialize predefined permissions
	for _, perm := range PredefinedPermissions {
		exists, err := q.PermissionExists(ctx, perm.Name)
		if err != nil {
			e.log.Error("failed to check permission existence", "permission", perm.Name, "error", err)
			continue
		}

		if !exists {
			permUUID := uuid.Must(uuid.NewV7())

			_, err := q.CreatePermission(ctx, query.CreatePermissionParams{
				UUID:        pgtype.UUID{Bytes: permUUID, Valid: true},
				Name:        perm.Name,
				DisplayName: perm.DisplayName,
				Description: pgtype.Text{String: perm.Description, Valid: true},
				Resource:    string(perm.Resource),
				Action:      string(perm.Action),
				Scope:       string(perm.Scope),
			})
			if err != nil {
				e.log.Error("failed to create permission", "permission", perm.Name, "error", err)
				continue
			}
			e.log.Debug("created predefined permission", "permission", perm.Name)
		}
	}

	return nil
}

// addPolicyUnlocked is an internal method that adds a policy without locking.
// Used during initialization when no concurrent access is possible.
func (e *Enforcer) addPolicyUnlocked(sub, dom, obj, act string) error {
	policyID := fmt.Sprintf("policy-%s-%s-%s-%s", sub, dom, obj, act)

	// Build resource pattern
	var resourcePattern string
	if obj == "*" {
		if dom == "global" {
			resourcePattern = "shadowapi:global:<.+>:.*"
		} else {
			resourcePattern = "shadowapi:workspace:<.+>:<.+>:.*"
		}
	} else {
		if dom == "global" {
			resourcePattern = fmt.Sprintf("shadowapi:global:%s:.*", obj)
		} else {
			resourcePattern = fmt.Sprintf("shadowapi:workspace:<.+>:%s:.*", obj)
		}
	}

	// Build action pattern
	actionPattern := act
	if act == "*" {
		actionPattern = "<.+>"
	}

	policy := &ladon.DefaultPolicy{
		ID:          policyID,
		Description: fmt.Sprintf("Policy for %s on %s in %s", sub, obj, dom),
		Subjects:    []string{"policy_set:" + sub},
		Resources:   []string{resourcePattern},
		Actions:     []string{actionPattern},
		Effect:      ladon.AllowAccess,
	}

	// Add workspace condition for workspace-scoped policies
	if dom != "global" {
		policy.Conditions = ladon.Conditions{
			"workspace": &WorkspaceCondition{Workspaces: []string{"*"}},
		}
	}

	return e.manager.Create(context.Background(), policy)
}
