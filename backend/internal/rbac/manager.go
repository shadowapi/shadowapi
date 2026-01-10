package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/ladon"
)

// ErrPolicyNotFound is returned when a policy is not found.
var ErrPolicyNotFound = errors.New("policy not found")

// PostgresManager implements ladon.Manager for PostgreSQL storage.
type PostgresManager struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

// NewPostgresManager creates a new PostgreSQL-backed policy manager.
func NewPostgresManager(db *pgxpool.Pool, log *slog.Logger) *PostgresManager {
	return &PostgresManager{db: db, log: log}
}

// Create stores a new policy.
func (m *PostgresManager) Create(ctx context.Context, policy ladon.Policy) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	condJSON, _ := json.Marshal(policy.GetConditions())
	metaJSON := policy.GetMeta()
	if metaJSON == nil {
		metaJSON = []byte("{}")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ladon_policy (id, description, effect, conditions, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO NOTHING
	`, policy.GetID(), policy.GetDescription(), policy.GetEffect(), condJSON, metaJSON)
	if err != nil {
		return err
	}

	for _, subj := range policy.GetSubjects() {
		_, err = tx.Exec(ctx, `
			INSERT INTO ladon_policy_subject (policy_id, subject) VALUES ($1, $2)
			ON CONFLICT (policy_id, subject) DO NOTHING
		`, policy.GetID(), subj)
		if err != nil {
			return err
		}
	}

	for _, res := range policy.GetResources() {
		_, err = tx.Exec(ctx, `
			INSERT INTO ladon_policy_resource (policy_id, resource) VALUES ($1, $2)
			ON CONFLICT (policy_id, resource) DO NOTHING
		`, policy.GetID(), res)
		if err != nil {
			return err
		}
	}

	for _, act := range policy.GetActions() {
		_, err = tx.Exec(ctx, `
			INSERT INTO ladon_policy_action (policy_id, action) VALUES ($1, $2)
			ON CONFLICT (policy_id, action) DO NOTHING
		`, policy.GetID(), act)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Get retrieves a policy by ID.
func (m *PostgresManager) Get(ctx context.Context, id string) (ladon.Policy, error) {
	var p ladon.DefaultPolicy
	var condJSON []byte

	err := m.db.QueryRow(ctx, `
		SELECT id, description, effect, conditions, meta
		FROM ladon_policy WHERE id = $1
	`, id).Scan(&p.ID, &p.Description, &p.Effect, &condJSON, &p.Meta)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}

	// Load subjects
	rows, err := m.db.Query(ctx, `SELECT subject FROM ladon_policy_subject WHERE policy_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		p.Subjects = append(p.Subjects, s)
	}

	// Load resources
	rows, err = m.db.Query(ctx, `SELECT resource FROM ladon_policy_resource WHERE policy_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		p.Resources = append(p.Resources, r)
	}

	// Load actions
	rows, err = m.db.Query(ctx, `SELECT action FROM ladon_policy_action WHERE policy_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		p.Actions = append(p.Actions, a)
	}

	// Unmarshal conditions
	if len(condJSON) > 0 {
		p.Conditions = make(ladon.Conditions)
		json.Unmarshal(condJSON, &p.Conditions)
	}

	return &p, nil
}

// Delete removes a policy.
func (m *PostgresManager) Delete(ctx context.Context, id string) error {
	_, err := m.db.Exec(ctx, `DELETE FROM ladon_policy WHERE id = $1`, id)
	return err
}

// GetAll retrieves all policies with pagination.
func (m *PostgresManager) GetAll(ctx context.Context, limit, offset int64) (ladon.Policies, error) {
	rows, err := m.db.Query(ctx, `
		SELECT id FROM ladon_policy ORDER BY id LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies ladon.Policies
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		p, err := m.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// FindRequestCandidates finds policies that could potentially match the request.
// This is used by Ladon to narrow down policies before detailed matching.
func (m *PostgresManager) FindRequestCandidates(ctx context.Context, r *ladon.Request) (ladon.Policies, error) {
	// Get workspace from context for policy set lookup
	workspace := ""
	if ws, ok := r.Context["workspace"].(string); ok {
		workspace = ws
	}

	// Get user's policy sets
	policySets, err := m.GetPolicySetsForUser(ctx, r.Subject, workspace)
	if err != nil {
		m.log.Debug("failed to get policy sets for user", "user", r.Subject, "error", err)
		policySets = []string{}
	}

	// Build subjects to search: user UUID + policy set subjects
	subjects := []string{r.Subject}
	for _, ps := range policySets {
		subjects = append(subjects, "policy_set:"+ps)
	}

	// Find policies with matching subjects
	rows, err := m.db.Query(ctx, `
		SELECT DISTINCT p.id
		FROM ladon_policy p
		JOIN ladon_policy_subject ps ON p.id = ps.policy_id
		WHERE ps.subject = ANY($1)
		   OR ps.subject LIKE '<%'
	`, subjects)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies ladon.Policies
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		p, err := m.Get(ctx, id)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// FindPoliciesForSubject finds policies for a specific subject.
func (m *PostgresManager) FindPoliciesForSubject(ctx context.Context, subject string) (ladon.Policies, error) {
	rows, err := m.db.Query(ctx, `
		SELECT DISTINCT policy_id FROM ladon_policy_subject WHERE subject = $1
	`, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies ladon.Policies
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		p, err := m.Get(ctx, id)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// FindPoliciesForResource finds policies for a specific resource.
func (m *PostgresManager) FindPoliciesForResource(ctx context.Context, resource string) (ladon.Policies, error) {
	rows, err := m.db.Query(ctx, `
		SELECT DISTINCT policy_id FROM ladon_policy_resource WHERE resource = $1
	`, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies ladon.Policies
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		p, err := m.Get(ctx, id)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// Update updates an existing policy.
func (m *PostgresManager) Update(ctx context.Context, policy ladon.Policy) error {
	if err := m.Delete(ctx, policy.GetID()); err != nil {
		return err
	}
	return m.Create(ctx, policy)
}

// Policy set assignment methods

// AddPolicySetAssignment assigns a policy set to a user in a domain.
func (m *PostgresManager) AddPolicySetAssignment(ctx context.Context, userUUID, policySet, domain string) error {
	var workspaceSlug *string
	if domain != "global" && domain != "" {
		workspaceSlug = &domain
	}

	_, err := m.db.Exec(ctx, `
		INSERT INTO user_policy_set (user_uuid, policy_set_name, workspace_slug)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_uuid, policy_set_name, workspace_slug) DO NOTHING
	`, userUUID, policySet, workspaceSlug)
	return err
}

// RemovePolicySetAssignment removes a policy set from a user in a domain.
func (m *PostgresManager) RemovePolicySetAssignment(ctx context.Context, userUUID, policySet, domain string) error {
	var workspaceSlug *string
	if domain != "global" && domain != "" {
		workspaceSlug = &domain
	}

	_, err := m.db.Exec(ctx, `
		DELETE FROM user_policy_set
		WHERE user_uuid = $1 AND policy_set_name = $2
		AND (workspace_slug = $3 OR ($3 IS NULL AND workspace_slug IS NULL))
	`, userUUID, policySet, workspaceSlug)
	return err
}

// GetPolicySetsForUser returns policy sets for a user in a domain.
// If domain is "global" or empty, returns only global policy sets.
// Otherwise, returns both workspace-specific policy sets and global policy sets.
func (m *PostgresManager) GetPolicySetsForUser(ctx context.Context, userUUID, domain string) ([]string, error) {
	var rows pgx.Rows
	var err error

	if domain == "global" || domain == "" {
		rows, err = m.db.Query(ctx, `
			SELECT policy_set_name FROM user_policy_set
			WHERE user_uuid = $1 AND workspace_slug IS NULL
		`, userUUID)
	} else {
		// Include both workspace-specific and global policy sets
		rows, err = m.db.Query(ctx, `
			SELECT policy_set_name FROM user_policy_set
			WHERE user_uuid = $1 AND (workspace_slug = $2 OR workspace_slug IS NULL)
		`, userUUID, domain)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policySets []string
	for rows.Next() {
		var policySet string
		if err := rows.Scan(&policySet); err != nil {
			return nil, err
		}
		policySets = append(policySets, policySet)
	}
	return policySets, nil
}

// GetAllPolicySetsForUser returns all policy sets across all domains.
func (m *PostgresManager) GetAllPolicySetsForUser(ctx context.Context, userUUID string) (map[string][]string, error) {
	rows, err := m.db.Query(ctx, `
		SELECT policy_set_name, COALESCE(workspace_slug, 'global')
		FROM user_policy_set
		WHERE user_uuid = $1
	`, userUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var policySet, domain string
		if err := rows.Scan(&policySet, &domain); err != nil {
			return nil, err
		}
		result[domain] = append(result[domain], policySet)
	}
	return result, nil
}

// RemoveAllPolicySetsForUserInDomain removes all policy sets for a user in a domain.
func (m *PostgresManager) RemoveAllPolicySetsForUserInDomain(ctx context.Context, userUUID, domain string) error {
	var err error
	if domain == "global" || domain == "" {
		_, err = m.db.Exec(ctx, `
			DELETE FROM user_policy_set
			WHERE user_uuid = $1 AND workspace_slug IS NULL
		`, userUUID)
	} else {
		_, err = m.db.Exec(ctx, `
			DELETE FROM user_policy_set
			WHERE user_uuid = $1 AND workspace_slug = $2
		`, userUUID, domain)
	}
	return err
}

// HasPolicySet checks if a user has a specific policy set in a domain.
func (m *PostgresManager) HasPolicySet(ctx context.Context, userUUID, policySet, domain string) (bool, error) {
	var workspaceSlug *string
	if domain != "global" && domain != "" {
		workspaceSlug = &domain
	}

	var exists bool
	err := m.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_policy_set
			WHERE user_uuid = $1
			  AND policy_set_name = $2
			  AND (workspace_slug = $3 OR ($3 IS NULL AND workspace_slug IS NULL))
		)
	`, userUUID, policySet, workspaceSlug).Scan(&exists)
	return exists, err
}
