-- ============================================================================
-- RBAC Role Queries
-- ============================================================================

-- name: ListRoles :many
SELECT * FROM rbac_role
ORDER BY scope, name;

-- name: ListRolesByScope :many
SELECT * FROM rbac_role
WHERE scope = sqlc.arg('scope')
ORDER BY name;

-- name: GetRoleByUUID :one
SELECT * FROM rbac_role
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetRoleByName :one
SELECT * FROM rbac_role
WHERE name = sqlc.arg('name');

-- name: CreateRole :one
INSERT INTO rbac_role (
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

-- name: UpdateRole :one
UPDATE rbac_role
SET
    display_name = sqlc.arg('display_name'),
    description = sqlc.arg('description'),
    permissions = sqlc.arg('permissions'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid AND is_system = FALSE
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM rbac_role
WHERE uuid = sqlc.arg('uuid')::uuid AND is_system = FALSE;

-- name: CountRoles :one
SELECT COUNT(*) FROM rbac_role;

-- name: RoleExists :one
SELECT EXISTS (
    SELECT 1 FROM rbac_role WHERE name = sqlc.arg('name')
) AS exists;

-- ============================================================================
-- RBAC Permission Queries
-- ============================================================================

-- name: ListPermissions :many
SELECT * FROM rbac_permission
ORDER BY resource, action;

-- name: ListPermissionsByScope :many
SELECT * FROM rbac_permission
WHERE scope = sqlc.arg('scope')
ORDER BY resource, action;

-- name: ListPermissionsByResource :many
SELECT * FROM rbac_permission
WHERE resource = sqlc.arg('resource')
ORDER BY action;

-- name: GetPermissionByUUID :one
SELECT * FROM rbac_permission
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetPermissionByName :one
SELECT * FROM rbac_permission
WHERE name = sqlc.arg('name');

-- name: CreatePermission :one
INSERT INTO rbac_permission (
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
DELETE FROM rbac_permission
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: CountPermissions :one
SELECT COUNT(*) FROM rbac_permission;

-- name: PermissionExists :one
SELECT EXISTS (
    SELECT 1 FROM rbac_permission WHERE name = sqlc.arg('name')
) AS exists;

-- ============================================================================
-- Casbin Rule Queries (for direct manipulation if needed)
-- ============================================================================

-- name: ListCasbinRules :many
SELECT * FROM casbin_rule
ORDER BY id;

-- name: ListCasbinRulesByPtype :many
SELECT * FROM casbin_rule
WHERE ptype = sqlc.arg('ptype')
ORDER BY id;

-- name: CountCasbinRules :one
SELECT COUNT(*) FROM casbin_rule;

-- name: DeleteAllCasbinRules :exec
DELETE FROM casbin_rule;
