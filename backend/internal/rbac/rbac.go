package rbac

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxadapter "github.com/pckhoi/casbin-pgx-adapter/v2"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

//go:embed model.conf
var modelConf string

// Enforcer wraps the Casbin enforcer and provides convenience methods.
type Enforcer struct {
	casbin *casbin.Enforcer
	log    *slog.Logger
	mu     sync.RWMutex
}

// Provide creates a new Enforcer for the dependency injector.
func Provide(i do.Injector) (*Enforcer, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	log.Info("initializing RBAC enforcer")

	// Create PostgreSQL adapter
	adapter, err := pgxadapter.NewAdapter(
		cfg.DB.URI,
		pgxadapter.WithTableName("casbin_rule"),
	)
	if err != nil {
		log.Error("failed to create Casbin adapter", "error", err)
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Load model from embedded config
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		log.Error("failed to load Casbin model", "error", err)
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	// Create enforcer
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Error("failed to create Casbin enforcer", "error", err)
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load policies from database
	if err := e.LoadPolicy(); err != nil {
		log.Error("failed to load policies", "error", err)
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	enforcer := &Enforcer{
		casbin: e,
		log:    log,
	}

	// Initialize predefined roles and policies if they don't exist
	if err := enforcer.initializePredefinedPolicies(context.Background(), dbp); err != nil {
		log.Warn("failed to initialize predefined policies", "error", err)
		// Don't fail on this - policies can be added later
	}

	log.Info("RBAC enforcer initialized successfully")
	return enforcer, nil
}

// Enforce checks if a user can perform an action on a resource in a domain.
// sub: user UUID
// dom: domain (workspace slug or "global")
// obj: resource (e.g., "datasource", "pipeline")
// act: action (e.g., "read", "write", "delete")
func (e *Enforcer) Enforce(sub, dom, obj, act string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.casbin.Enforce(sub, dom, obj, act)
}

// EnforceContext is a convenience method that checks permission using typed parameters.
func (e *Enforcer) EnforceContext(ctx context.Context, userUUID string, domain string, resource Resource, action Action) (bool, error) {
	return e.Enforce(userUUID, domain, string(resource), string(action))
}

// AddRoleForUserInDomain assigns a role to a user in a specific domain.
// For global roles, use "global" as the domain.
func (e *Enforcer) AddRoleForUserInDomain(userUUID, role, domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, err := e.casbin.AddGroupingPolicy(userUUID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to add role: %w", err)
	}
	return nil
}

// RemoveRoleForUserInDomain removes a role from a user in a specific domain.
func (e *Enforcer) RemoveRoleForUserInDomain(userUUID, role, domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, err := e.casbin.RemoveGroupingPolicy(userUUID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}
	return nil
}

// GetRolesForUserInDomain returns all roles a user has in a specific domain.
func (e *Enforcer) GetRolesForUserInDomain(userUUID, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.casbin.GetRolesForUserInDomain(userUUID, domain)
}

// GetAllRolesForUser returns all roles a user has across all domains.
func (e *Enforcer) GetAllRolesForUser(userUUID string) (map[string][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Get all grouping policies (role assignments)
	policies, err := e.casbin.GetGroupingPolicy()
	if err != nil {
		return nil, err
	}
	result := make(map[string][]string)

	for _, policy := range policies {
		if len(policy) >= 3 && policy[0] == userUUID {
			role := policy[1]
			domain := policy[2]
			result[domain] = append(result[domain], role)
		}
	}

	return result, nil
}

// GetUsersForRoleInDomain returns all users with a specific role in a domain.
func (e *Enforcer) GetUsersForRoleInDomain(role, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.casbin.GetUsersForRoleInDomain(role, domain)
}

// AddPolicy adds a policy rule.
// sub: subject (role name)
// dom: domain (workspace slug or "global")
// obj: resource (e.g., "datasource", "pipeline", or "*" for all)
// act: action (e.g., "read", "write", "*" for all)
func (e *Enforcer) AddPolicy(sub, dom, obj, act string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, err := e.casbin.AddPolicy(sub, dom, obj, act)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}
	return nil
}

// RemovePolicy removes a policy rule.
func (e *Enforcer) RemovePolicy(sub, dom, obj, act string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, err := e.casbin.RemovePolicy(sub, dom, obj, act)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}
	return nil
}

// GetPolicies returns all policy rules.
func (e *Enforcer) GetPolicies() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.casbin.GetPolicy()
}

// GetGroupingPolicies returns all role assignments.
func (e *Enforcer) GetGroupingPolicies() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.casbin.GetGroupingPolicy()
}

// SavePolicy saves all policies to the database.
func (e *Enforcer) SavePolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.casbin.SavePolicy()
}

// LoadPolicy reloads all policies from the database.
func (e *Enforcer) LoadPolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.casbin.LoadPolicy()
}

// HasRoleForUserInDomain checks if a user has a specific role in a domain.
func (e *Enforcer) HasRoleForUserInDomain(userUUID, role, domain string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	roles := e.casbin.GetRolesForUserInDomain(userUUID, domain)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// RemoveAllRolesForUserInDomain removes all roles for a user in a specific domain.
func (e *Enforcer) RemoveAllRolesForUserInDomain(userUUID, domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get current roles
	roles := e.casbin.GetRolesForUserInDomain(userUUID, domain)
	for _, role := range roles {
		if _, err := e.casbin.RemoveGroupingPolicy(userUUID, role, domain); err != nil {
			return fmt.Errorf("failed to remove role %s: %w", role, err)
		}
	}
	return nil
}
