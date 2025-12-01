-- name: CreateTenant :one
INSERT INTO tenant (
    uuid,
    name,
    display_name,
    is_enabled,
    settings,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('name'),
    sqlc.arg('display_name'),
    sqlc.arg('is_enabled')::boolean,
    sqlc.arg('settings'),
    NOW()
) RETURNING *;

-- name: GetTenant :one
SELECT * FROM tenant WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: GetTenantByName :one
SELECT * FROM tenant WHERE name = sqlc.arg('name');

-- name: ListTenants :many
SELECT * FROM tenant
ORDER BY display_name ASC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset');

-- name: ListEnabledTenants :many
SELECT * FROM tenant
WHERE is_enabled = true
ORDER BY display_name ASC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset');

-- name: UpdateTenant :exec
UPDATE tenant SET
    display_name = sqlc.arg('display_name'),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    settings = sqlc.arg('settings'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteTenant :exec
DELETE FROM tenant WHERE uuid = sqlc.arg('uuid')::uuid;
