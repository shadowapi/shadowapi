package rbac

import (
	"log/slog"
	"net/http"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
)

// OperationPermission maps operation names to required permissions.
type OperationPermission struct {
	Resource Resource
	Action   Action
}

// OperationPermissionMap maps operation IDs to their required permissions.
// Operations not in this map are allowed by default (no permission check).
var OperationPermissionMap = map[string]OperationPermission{
	// User operations (global)
	"listUsers":   {ResourceUser, ActionRead},
	"createUser":  {ResourceUser, ActionCreate},
	"getUser":     {ResourceUser, ActionRead},
	"updateUser":  {ResourceUser, ActionWrite},
	"deleteUser":  {ResourceUser, ActionDelete},

	// Workspace operations
	"listWorkspaces":  {ResourceWorkspace, ActionRead},
	"createWorkspace": {ResourceWorkspace, ActionCreate},
	"getWorkspace":    {ResourceWorkspace, ActionRead},
	"updateWorkspace": {ResourceWorkspace, ActionWrite},
	"deleteWorkspace": {ResourceWorkspace, ActionDelete},

	// Workspace member operations
	"listWorkspaceMembers":      {ResourceMember, ActionRead},
	"addWorkspaceMember":        {ResourceMember, ActionCreate},
	"updateWorkspaceMemberRole": {ResourceMember, ActionWrite},
	"removeWorkspaceMember":     {ResourceMember, ActionDelete},

	// Datasource operations
	"datasourceList":                {ResourceDatasource, ActionRead},
	"datasourceEmailCreate":         {ResourceDatasource, ActionCreate},
	"datasourceEmailGet":            {ResourceDatasource, ActionRead},
	"datasourceEmailUpdate":         {ResourceDatasource, ActionWrite},
	"datasourceEmailDelete":         {ResourceDatasource, ActionDelete},
	"datasourceTelegramCreate":      {ResourceDatasource, ActionCreate},
	"datasourceTelegramGet":         {ResourceDatasource, ActionRead},
	"datasourceTelegramUpdate":      {ResourceDatasource, ActionWrite},
	"datasourceTelegramDelete":      {ResourceDatasource, ActionDelete},
	"datasourceWhatsAppCreate":      {ResourceDatasource, ActionCreate},
	"datasourceWhatsAppGet":         {ResourceDatasource, ActionRead},
	"datasourceWhatsAppUpdate":      {ResourceDatasource, ActionWrite},
	"datasourceWhatsAppDelete":      {ResourceDatasource, ActionDelete},
	"datasourceLinkedinCreate":      {ResourceDatasource, ActionCreate},
	"datasourceLinkedinGet":         {ResourceDatasource, ActionRead},
	"datasourceLinkedinUpdate":      {ResourceDatasource, ActionWrite},
	"datasourceLinkedinDelete":      {ResourceDatasource, ActionDelete},

	// Pipeline operations
	"pipelineList":   {ResourcePipeline, ActionRead},
	"pipelineCreate": {ResourcePipeline, ActionCreate},
	"pipelineGet":    {ResourcePipeline, ActionRead},
	"pipelineUpdate": {ResourcePipeline, ActionWrite},
	"pipelineDelete": {ResourcePipeline, ActionDelete},

	// Mapper operations (pipeline field mapping)
	// NOTE: MapperSourceFieldsList and MapperTransformsList are excluded from RBAC
	// because they return static schema metadata and don't require workspace context.
	// Operations not in this map are allowed by default.
	"MapperValidate": {ResourcePipeline, ActionWrite},

	// Storage operations
	// NOTE: These use camelCase but ogen generates PascalCase (e.g., StoragePostgresUpdate)
	// so they don't actually match and storage operations pass through without RBAC check.
	"storageList":   {ResourceStorage, ActionRead},
	"storageCreate": {ResourceStorage, ActionCreate},
	"storageGet":    {ResourceStorage, ActionRead},
	"storageUpdate": {ResourceStorage, ActionWrite},
	"storageDelete": {ResourceStorage, ActionDelete},

	// Contact operations
	"contactList":   {ResourceContact, ActionRead},
	"contactCreate": {ResourceContact, ActionCreate},
	"contactGet":    {ResourceContact, ActionRead},
	"contactUpdate": {ResourceContact, ActionWrite},
	"contactDelete": {ResourceContact, ActionDelete},

	// Message operations
	"messageList":   {ResourceMessage, ActionRead},
	"messageGet":    {ResourceMessage, ActionRead},
	"messageSearch": {ResourceMessage, ActionRead},

	// Scheduler operations
	"schedulerList":   {ResourceScheduler, ActionRead},
	"schedulerCreate": {ResourceScheduler, ActionCreate},
	"schedulerGet":    {ResourceScheduler, ActionRead},
	"schedulerUpdate": {ResourceScheduler, ActionWrite},
	"schedulerDelete": {ResourceScheduler, ActionDelete},

	// RBAC operations (require super_admin)
	"listRoles":          {ResourceRole, ActionRead},
	"createRole":         {ResourceRole, ActionWrite},
	"getRole":            {ResourceRole, ActionRead},
	"updateRole":         {ResourceRole, ActionWrite},
	"deleteRole":         {ResourceRole, ActionWrite},
	"listPermissions":    {ResourceRole, ActionRead},
	"getUserRoles":       {ResourceRole, ActionRead},
	"assignRoleToUser":   {ResourceRole, ActionWrite},
	"removeRoleFromUser": {ResourceRole, ActionWrite},
	"checkPermission":    {ResourceRole, ActionRead},

	// Worker operations (global, require admin)
	"listRegisteredWorkers":        {ResourceWorker, ActionRead},
	"getRegisteredWorker":          {ResourceWorker, ActionRead},
	"updateRegisteredWorker":       {ResourceWorker, ActionWrite},
	"deleteRegisteredWorker":       {ResourceWorker, ActionDelete},
	"listWorkerEnrollmentTokens":   {ResourceWorker, ActionRead},
	"createWorkerEnrollmentToken":  {ResourceWorker, ActionCreate},
	"getWorkerEnrollmentToken":     {ResourceWorker, ActionRead},
	"deleteWorkerEnrollmentToken":  {ResourceWorker, ActionDelete},
}

// Middleware provides RBAC enforcement for ogen handlers.
type Middleware struct {
	enforcer *Enforcer
	log      *slog.Logger
}

// ProvideMiddleware creates a new RBAC middleware for the dependency injector.
func ProvideMiddleware(i do.Injector) (*Middleware, error) {
	enforcer := do.MustInvoke[*Enforcer](i)
	log := do.MustInvoke[*slog.Logger](i)

	return &Middleware{
		enforcer: enforcer,
		log:      log,
	}, nil
}

// OgenMiddleware is the ogen middleware function for RBAC enforcement.
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	opName := req.OperationName

	// Check if this operation requires permission
	perm, needsAuth := OperationPermissionMap[opName]
	if !needsAuth {
		// No permission required for this operation
		return next(req)
	}

	// Get user claims from context (set by auth middleware)
	claims, ok := req.Context.Value(auth.UserClaimsContextKey).(*oauth2.Claims)
	if !ok || claims == nil {
		m.log.Debug("RBAC: no user claims in context", "operation", opName)
		return middleware.Response{}, &ForbiddenError{Message: "authentication required"}
	}
	userUUID := claims.Subject

	// Determine the domain (workspace or global)
	domain := "global"
	if !IsGlobalResource(perm.Resource) {
		// For workspace-scoped resources, get the workspace slug from context
		workspaceSlug := workspace.GetWorkspaceSlug(req.Context)
		if workspaceSlug != "" {
			domain = workspaceSlug
		}
	}

	// Enforce permission
	allowed, err := m.enforcer.Enforce(userUUID, domain, string(perm.Resource), string(perm.Action))
	if err != nil {
		m.log.Error("RBAC: enforcement error",
			"operation", opName,
			"user", userUUID,
			"domain", domain,
			"resource", perm.Resource,
			"action", perm.Action,
			"error", err,
		)
		return middleware.Response{}, &ForbiddenError{Message: "permission check failed"}
	}

	if !allowed {
		m.log.Debug("RBAC: access denied",
			"operation", opName,
			"user", userUUID,
			"domain", domain,
			"resource", perm.Resource,
			"action", perm.Action,
		)
		return middleware.Response{}, &ForbiddenError{
			Message: "access denied: insufficient permissions",
		}
	}

	m.log.Debug("RBAC: access granted",
		"operation", opName,
		"user", userUUID,
		"domain", domain,
		"resource", perm.Resource,
		"action", perm.Action,
	)

	return next(req)
}

// ForbiddenError represents a 403 Forbidden error.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code for this error.
func (e *ForbiddenError) StatusCode() int {
	return http.StatusForbidden
}
