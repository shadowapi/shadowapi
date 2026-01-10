-- ============================================================================
-- Policy Set Queries
-- ============================================================================

-- name: ListPolicySets :many
SELECT * FROM policy_set
ORDER BY scope, name;

-- name: ListPolicySetsByScope :many
SELECT * FROM policy_set
WHERE scope = sqlc.arg('scope')
ORDER BY name;

-- name: GetPolicySetByUUID :one
SELECT * FROM policy_set
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetPolicySetByName :one
SELECT * FROM policy_set
WHERE name = sqlc.arg('name');

-- name: CreatePolicySet :one
INSERT INTO policy_set (
    uuid,
    name,
    display_name,
    description,
    scope,
    is_system,
    permissions,
    created_at,
    updated_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('name'),
    sqlc.arg('display_name'),
    sqlc.arg('description'),
    sqlc.arg('scope'),
    sqlc.arg('is_system')::boolean,
    sqlc.arg('permissions'),
    NOW(),
    NULL
) RETURNING *;

-- name: UpdatePolicySet :one
UPDATE policy_set
SET
    display_name = sqlc.arg('display_name'),
    description = sqlc.arg('description'),
    permissions = sqlc.arg('permissions'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid AND is_system = FALSE
RETURNING *;

-- name: DeletePolicySet :exec
DELETE FROM policy_set
WHERE uuid = sqlc.arg('uuid')::uuid AND is_system = FALSE;

-- name: CountPolicySets :one
SELECT COUNT(*) FROM policy_set;

-- name: PolicySetExists :one
SELECT EXISTS (
    SELECT 1 FROM policy_set WHERE name = sqlc.arg('name')
) AS exists;

-- ============================================================================
-- Permission Queries
-- ============================================================================

-- name: ListPermissions :many
SELECT * FROM permission
ORDER BY resource, action;

-- name: ListPermissionsByScope :many
SELECT * FROM permission
WHERE scope = sqlc.arg('scope')
ORDER BY resource, action;

-- name: ListPermissionsByResource :many
SELECT * FROM permission
WHERE resource = sqlc.arg('resource')
ORDER BY action;

-- name: GetPermissionByUUID :one
SELECT * FROM permission
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetPermissionByName :one
SELECT * FROM permission
WHERE name = sqlc.arg('name');

-- name: CreatePermission :one
INSERT INTO permission (
    uuid,
    name,
    display_name,
    description,
    resource,
    action,
    scope,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('name'),
    sqlc.arg('display_name'),
    sqlc.arg('description'),
    sqlc.arg('resource'),
    sqlc.arg('action'),
    sqlc.arg('scope'),
    NOW()
) RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permission
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: CountPermissions :one
SELECT COUNT(*) FROM permission;

-- name: PermissionExists :one
SELECT EXISTS (
    SELECT 1 FROM permission WHERE name = sqlc.arg('name')
) AS exists;

-- ============================================================================
-- Ladon Policy Queries
-- ============================================================================

-- name: CreateLadonPolicy :one
INSERT INTO ladon_policy (id, description, effect, conditions, meta, created_at)
VALUES (sqlc.arg('id'), sqlc.arg('description'), sqlc.arg('effect'), sqlc.arg('conditions'), sqlc.arg('meta'), NOW())
RETURNING *;

-- name: GetLadonPolicy :one
SELECT * FROM ladon_policy WHERE id = sqlc.arg('id');

-- name: UpdateLadonPolicy :one
UPDATE ladon_policy
SET description = sqlc.arg('description'),
    effect = sqlc.arg('effect'),
    conditions = sqlc.arg('conditions'),
    meta = sqlc.arg('meta'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteLadonPolicy :exec
DELETE FROM ladon_policy WHERE id = sqlc.arg('id');

-- name: ListLadonPolicies :many
SELECT * FROM ladon_policy ORDER BY id LIMIT sqlc.arg('limit_val') OFFSET sqlc.arg('offset_val');

-- name: CountLadonPolicies :one
SELECT COUNT(*) FROM ladon_policy;

-- name: LadonPolicyExists :one
SELECT EXISTS (SELECT 1 FROM ladon_policy WHERE id = sqlc.arg('id')) AS exists;

-- ============================================================================
-- Ladon Policy Subject Queries
-- ============================================================================

-- name: CreateLadonPolicySubject :exec
INSERT INTO ladon_policy_subject (policy_id, subject)
VALUES (sqlc.arg('policy_id'), sqlc.arg('subject'))
ON CONFLICT (policy_id, subject) DO NOTHING;

-- name: ListLadonPolicySubjects :many
SELECT subject FROM ladon_policy_subject WHERE policy_id = sqlc.arg('policy_id');

-- name: DeleteLadonPolicySubjects :exec
DELETE FROM ladon_policy_subject WHERE policy_id = sqlc.arg('policy_id');

-- name: FindPoliciesBySubject :many
SELECT DISTINCT policy_id FROM ladon_policy_subject WHERE subject = sqlc.arg('subject');

-- name: FindPoliciesBySubjects :many
SELECT DISTINCT policy_id FROM ladon_policy_subject WHERE subject = ANY(sqlc.arg('subjects')::text[]);

-- ============================================================================
-- Ladon Policy Resource Queries
-- ============================================================================

-- name: CreateLadonPolicyResource :exec
INSERT INTO ladon_policy_resource (policy_id, resource)
VALUES (sqlc.arg('policy_id'), sqlc.arg('resource'))
ON CONFLICT (policy_id, resource) DO NOTHING;

-- name: ListLadonPolicyResources :many
SELECT resource FROM ladon_policy_resource WHERE policy_id = sqlc.arg('policy_id');

-- name: DeleteLadonPolicyResources :exec
DELETE FROM ladon_policy_resource WHERE policy_id = sqlc.arg('policy_id');

-- name: FindPoliciesByResource :many
SELECT DISTINCT policy_id FROM ladon_policy_resource WHERE resource = sqlc.arg('resource');

-- ============================================================================
-- Ladon Policy Action Queries
-- ============================================================================

-- name: CreateLadonPolicyAction :exec
INSERT INTO ladon_policy_action (policy_id, action)
VALUES (sqlc.arg('policy_id'), sqlc.arg('action'))
ON CONFLICT (policy_id, action) DO NOTHING;

-- name: ListLadonPolicyActions :many
SELECT action FROM ladon_policy_action WHERE policy_id = sqlc.arg('policy_id');

-- name: DeleteLadonPolicyActions :exec
DELETE FROM ladon_policy_action WHERE policy_id = sqlc.arg('policy_id');

-- ============================================================================
-- User Policy Set Assignment Queries
-- ============================================================================

-- name: CreateUserPolicySetAssignment :exec
INSERT INTO user_policy_set (user_uuid, policy_set_name, workspace_slug)
VALUES (sqlc.arg('user_uuid')::uuid, sqlc.arg('policy_set_name'), sqlc.arg('workspace_slug'))
ON CONFLICT (user_uuid, policy_set_name, workspace_slug) DO NOTHING;

-- name: DeleteUserPolicySetAssignment :exec
DELETE FROM user_policy_set
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND policy_set_name = sqlc.arg('policy_set_name')
  AND (workspace_slug = sqlc.arg('workspace_slug') OR (sqlc.arg('workspace_slug') IS NULL AND workspace_slug IS NULL));

-- name: ListUserPolicySetAssignmentsByUser :many
SELECT * FROM user_policy_set WHERE user_uuid = sqlc.arg('user_uuid')::uuid;

-- name: ListUserPolicySetAssignmentsByUserAndWorkspace :many
SELECT * FROM user_policy_set
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND (workspace_slug = sqlc.arg('workspace_slug') OR workspace_slug IS NULL);

-- name: ListGlobalPolicySetAssignmentsByUser :many
SELECT * FROM user_policy_set
WHERE user_uuid = sqlc.arg('user_uuid')::uuid AND workspace_slug IS NULL;

-- name: HasUserPolicySetAssignment :one
SELECT EXISTS (
    SELECT 1 FROM user_policy_set
    WHERE user_uuid = sqlc.arg('user_uuid')::uuid
      AND policy_set_name = sqlc.arg('policy_set_name')
      AND (workspace_slug = sqlc.arg('workspace_slug') OR (sqlc.arg('workspace_slug') IS NULL AND workspace_slug IS NULL))
) AS exists;

-- name: CountUserPolicySetAssignments :one
SELECT COUNT(*) FROM user_policy_set;

-- name: DeleteAllUserPolicySetAssignmentsInWorkspace :exec
DELETE FROM user_policy_set
WHERE user_uuid = sqlc.arg('user_uuid')::uuid
  AND workspace_slug = sqlc.arg('workspace_slug');
