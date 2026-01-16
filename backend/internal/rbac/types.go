package rbac

// ContextKey is a type for context keys used by this package.
type contextKey string

const (
	// EnforcerContextKey is the context key for the RBAC enforcer.
	EnforcerContextKey contextKey = "rbac_enforcer"
)

// Scope represents the scope of a policy set or permission.
type Scope string

const (
	// ScopeGlobal represents system-wide permissions.
	ScopeGlobal Scope = "global"
	// ScopeWorkspace represents workspace-scoped permissions.
	ScopeWorkspace Scope = "workspace"
)

// Action represents an action that can be performed on a resource.
type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionCreate Action = "create"
	ActionDelete Action = "delete"
	ActionAdmin  Action = "admin"
)

// Resource represents a resource type in the system.
type Resource string

const (
	// Global resources
	ResourceUser      Resource = "user"
	ResourceWorkspace Resource = "workspace"
	ResourcePolicySet Resource = "policy_set"
	ResourceRBAC      Resource = "rbac"
	ResourceWorker    Resource = "worker"

	// Workspace-scoped resources
	ResourceDatasource Resource = "datasource"
	ResourcePipeline   Resource = "pipeline"
	ResourceStorage    Resource = "storage"
	ResourceScheduler  Resource = "scheduler"
	ResourceMember     Resource = "member"
)

// Permission represents a permission to perform an action on a resource.
type Permission struct {
	Resource Resource
	Action   Action
}

// PredefinedPolicySet represents a system policy set with its permissions.
type PredefinedPolicySet struct {
	Name        string
	DisplayName string
	Description string
	Scope       Scope
	Permissions []Permission
}

// GlobalResources is a set of resources that don't require workspace context.
var GlobalResources = map[Resource]bool{
	ResourceUser:      true,
	ResourceWorkspace: true,
	ResourcePolicySet: true,
	ResourceRBAC:      true,
	ResourceWorker:    true,
}

// IsGlobalResource checks if a resource is global (not workspace-scoped).
func IsGlobalResource(r Resource) bool {
	return GlobalResources[r]
}
