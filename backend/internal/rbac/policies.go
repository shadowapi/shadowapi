package rbac

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Role names
const (
	RoleSuperAdmin      = "super_admin"
	RoleWorkspaceOwner  = "workspace_owner"
	RoleWorkspaceAdmin  = "workspace_admin"
	RoleWorkspaceMember = "workspace_member"
)

// PredefinedRoles contains all system roles that are created on startup.
var PredefinedRoles = []PredefinedRole{
	{
		Name:        RoleSuperAdmin,
		DisplayName: "Super Admin",
		Description: "Full system access including user management and all workspaces",
		Scope:       ScopeGlobal,
		Permissions: []Permission{
			{Resource: "*", Action: "*"}, // Full access to everything
		},
	},
	{
		Name:        RoleWorkspaceOwner,
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
		Name:        RoleWorkspaceAdmin,
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
		Name:        RoleWorkspaceMember,
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
	{"role:read", "View Roles", "View roles and permissions", ResourceRole, ActionRead, ScopeGlobal},
	{"role:write", "Manage Roles", "Create and modify roles", ResourceRole, ActionWrite, ScopeGlobal},
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
	{"member:write", "Edit Members", "Edit member roles", ResourceMember, ActionWrite, ScopeWorkspace},
	{"member:delete", "Remove Members", "Remove members from workspace", ResourceMember, ActionDelete, ScopeWorkspace},
}

// initializePredefinedPolicies creates the predefined roles and permissions in the database.
func (e *Enforcer) initializePredefinedPolicies(ctx context.Context, dbp *pgxpool.Pool) error {
	q := query.New(dbp)

	// Initialize predefined roles
	for _, role := range PredefinedRoles {
		exists, err := q.RoleExists(ctx, role.Name)
		if err != nil {
			e.log.Error("failed to check role existence", "role", role.Name, "error", err)
			continue
		}

		if !exists {
			permJSON, _ := json.Marshal(role.Permissions)
			roleUUID := uuid.Must(uuid.NewV7())

			_, err := q.CreateRole(ctx, query.CreateRoleParams{
				UUID:        pgtype.UUID{Bytes: roleUUID, Valid: true},
				Name:        role.Name,
				DisplayName: role.DisplayName,
				Description: pgtype.Text{String: role.Description, Valid: true},
				Scope:       string(role.Scope),
				IsSystem:    true,
				Permissions: permJSON,
			})
			if err != nil {
				e.log.Error("failed to create role", "role", role.Name, "error", err)
				continue
			}
			e.log.Info("created predefined role", "role", role.Name)
		}

		// Add Casbin policies for this role
		for _, perm := range role.Permissions {
			domain := "*" // Workspace roles use wildcard domain, actual domain is determined at runtime
			if role.Scope == ScopeGlobal {
				domain = "global"
			}
			if err := e.AddPolicy(role.Name, domain, string(perm.Resource), string(perm.Action)); err != nil {
				e.log.Debug("policy may already exist", "role", role.Name, "resource", perm.Resource, "action", perm.Action)
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

// MapWorkspaceMemberRoleToCasbinRole maps legacy workspace_member.role to Casbin role names.
func MapWorkspaceMemberRoleToCasbinRole(legacyRole string) string {
	switch legacyRole {
	case "owner":
		return RoleWorkspaceOwner
	case "admin":
		return RoleWorkspaceAdmin
	case "member":
		return RoleWorkspaceMember
	default:
		return RoleWorkspaceMember
	}
}
