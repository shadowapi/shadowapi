package rbac

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/ladon"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Enforcer wraps Ory Ladon warden for access control decisions.
type Enforcer struct {
	warden  *ladon.Ladon
	manager *PostgresManager
	log     *slog.Logger
	mu      sync.RWMutex
}

// Provide creates a new Enforcer for the dependency injector.
func Provide(i do.Injector) (*Enforcer, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	_ = cfg // Reserved for future configuration options

	log.Info("initializing Ladon policy-based authorization enforcer")

	manager := NewPostgresManager(dbp, log)
	warden := &ladon.Ladon{
		Manager: manager,
	}

	enforcer := &Enforcer{
		warden:  warden,
		manager: manager,
		log:     log,
	}

	// Initialize predefined policy sets and policies if they don't exist
	if err := enforcer.initializePredefinedPolicies(context.Background(), dbp); err != nil {
		log.Warn("failed to initialize predefined policies", "error", err)
		// Don't fail on this - policies can be added later
	}

	log.Info("Ladon policy-based authorization enforcer initialized successfully")
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

	req := &ladon.Request{
		Subject:  sub,
		Resource: e.buildResourceURN(dom, obj),
		Action:   act,
		Context: ladon.Context{
			"workspace": dom,
		},
	}

	err := e.warden.IsAllowed(context.Background(), req)
	if err != nil {
		// Ladon returns an error when access is denied
		// We treat any error as denial for safety
		return false, nil
	}
	return true, nil
}

// EnforceContext is a convenience method that checks permission using typed parameters.
func (e *Enforcer) EnforceContext(ctx context.Context, userUUID string, domain string, resource Resource, action Action) (bool, error) {
	return e.Enforce(userUUID, domain, string(resource), string(action))
}

// buildResourceURN converts flat resource to URN format for Ladon.
// Global resources: shadowapi:global:{resource}:*
// Workspace resources: shadowapi:workspace:{domain}:{resource}:*
func (e *Enforcer) buildResourceURN(domain, resource string) string {
	if domain == "global" {
		return fmt.Sprintf("shadowapi:global:%s:*", resource)
	}
	return fmt.Sprintf("shadowapi:workspace:%s:%s:*", domain, resource)
}

// AssignPolicySetToUser assigns a policy set to a user in a specific domain.
// For global policy sets, use "global" as the domain.
func (e *Enforcer) AssignPolicySetToUser(userUUID, policySet, domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.manager.AddPolicySetAssignment(context.Background(), userUUID, policySet, domain)
}

// RemovePolicySetFromUser removes a policy set from a user in a specific domain.
func (e *Enforcer) RemovePolicySetFromUser(userUUID, policySet, domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.manager.RemovePolicySetAssignment(context.Background(), userUUID, policySet, domain)
}

// GetPolicySetsForUser returns all policy sets a user has in a specific domain.
func (e *Enforcer) GetPolicySetsForUser(userUUID, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	policySets, err := e.manager.GetPolicySetsForUser(context.Background(), userUUID, domain)
	if err != nil {
		e.log.Debug("failed to get policy sets for user", "user", userUUID, "domain", domain, "error", err)
		return nil
	}
	return policySets
}

// GetAllPolicySetsForUser returns all policy sets a user has across all domains.
func (e *Enforcer) GetAllPolicySetsForUser(userUUID string) (map[string][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.manager.GetAllPolicySetsForUser(context.Background(), userUUID)
}

// GetUsersForPolicySetInDomain returns all users with a specific policy set in a domain.
// Note: This requires iterating through all policy set assignments, which may be slow for large datasets.
func (e *Enforcer) GetUsersForPolicySetInDomain(policySet, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Query users with this policy set in the domain
	var workspaceSlug *string
	if domain != "global" && domain != "" {
		workspaceSlug = &domain
	}

	rows, err := e.manager.db.Query(context.Background(), `
		SELECT user_uuid::text FROM user_policy_set
		WHERE policy_set_name = $1 AND (workspace_slug = $2 OR ($2 IS NULL AND workspace_slug IS NULL))
	`, policySet, workspaceSlug)
	if err != nil {
		e.log.Debug("failed to get users for policy set", "policy_set", policySet, "domain", domain, "error", err)
		return nil
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var userUUID string
		if err := rows.Scan(&userUUID); err != nil {
			continue
		}
		users = append(users, userUUID)
	}
	return users
}

// AddPolicy adds a policy rule for Ladon.
// sub: subject (policy set name, will be prefixed with "policy_set:")
// dom: domain (workspace slug or "global")
// obj: resource (e.g., "datasource", "pipeline", or "*" for all)
// act: action (e.g., "read", "write", "*" for all)
func (e *Enforcer) AddPolicy(sub, dom, obj, act string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

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

// RemovePolicy removes a policy rule.
func (e *Enforcer) RemovePolicy(sub, dom, obj, act string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	policyID := fmt.Sprintf("policy-%s-%s-%s-%s", sub, dom, obj, act)
	return e.manager.Delete(context.Background(), policyID)
}

// GetPolicies returns all policy rules.
func (e *Enforcer) GetPolicies() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies, err := e.manager.GetAll(context.Background(), 10000, 0)
	if err != nil {
		return nil, err
	}

	var result [][]string
	for _, p := range policies {
		for _, subj := range p.GetSubjects() {
			for _, res := range p.GetResources() {
				for _, act := range p.GetActions() {
					result = append(result, []string{subj, res, act, p.GetEffect()})
				}
			}
		}
	}
	return result, nil
}

// GetGroupingPolicies returns all policy set assignments.
func (e *Enforcer) GetGroupingPolicies() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rows, err := e.manager.db.Query(context.Background(), `
		SELECT user_uuid::text, policy_set_name, COALESCE(workspace_slug, 'global')
		FROM user_policy_set
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result [][]string
	for rows.Next() {
		var userUUID, policySet, domain string
		if err := rows.Scan(&userUUID, &policySet, &domain); err != nil {
			continue
		}
		result = append(result, []string{userUUID, policySet, domain})
	}
	return result, nil
}

// SavePolicy is a no-op for Ladon as policies are immediately persisted.
func (e *Enforcer) SavePolicy() error {
	return nil
}

// LoadPolicy is a no-op for Ladon as policies are loaded on demand.
func (e *Enforcer) LoadPolicy() error {
	return nil
}

// HasPolicySet checks if a user has a specific policy set in a domain.
func (e *Enforcer) HasPolicySet(userUUID, policySet, domain string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	hasPolicySet, err := e.manager.HasPolicySet(context.Background(), userUUID, policySet, domain)
	if err != nil {
		e.log.Debug("failed to check policy set", "user", userUUID, "policy_set", policySet, "domain", domain, "error", err)
		return false
	}
	return hasPolicySet
}

// RemoveAllPolicySetsFromUser removes all policy sets for a user in a specific domain.
func (e *Enforcer) RemoveAllPolicySetsFromUser(userUUID, domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.manager.RemoveAllPolicySetsForUserInDomain(context.Background(), userUUID, domain)
}
