-- name: CreateTenantSession :one
INSERT INTO tenant_session (
    uuid,
    session_id,
    tenant_uuid,
    user_uuid,
    is_active,
    expires_at,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('session_id'),
    sqlc.arg('tenant_uuid')::uuid,
    sqlc.arg('user_uuid')::uuid,
    sqlc.arg('is_active')::boolean,
    sqlc.arg('expires_at'),
    NOW()
) RETURNING *;

-- name: GetTenantSessionsBySessionID :many
SELECT
    ts.uuid,
    ts.session_id,
    ts.tenant_uuid,
    ts.user_uuid,
    ts.is_active,
    ts.last_accessed_at,
    ts.expires_at,
    ts.created_at,
    t.name as tenant_name,
    t.display_name as tenant_display_name,
    u.email as user_email
FROM tenant_session ts
JOIN tenant t ON ts.tenant_uuid = t.uuid
JOIN "user" u ON ts.user_uuid = u.uuid
WHERE ts.session_id = sqlc.arg('session_id')
  AND ts.is_active = true
  AND ts.expires_at > NOW()
  AND t.is_enabled = true;

-- name: UpsertTenantSession :one
INSERT INTO tenant_session (
    uuid,
    session_id,
    tenant_uuid,
    user_uuid,
    is_active,
    expires_at,
    created_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('session_id'),
    sqlc.arg('tenant_uuid')::uuid,
    sqlc.arg('user_uuid')::uuid,
    true,
    sqlc.arg('expires_at'),
    NOW()
) ON CONFLICT (session_id, tenant_uuid) DO UPDATE SET
    user_uuid = EXCLUDED.user_uuid,
    is_active = true,
    expires_at = EXCLUDED.expires_at,
    last_accessed_at = NOW()
RETURNING *;

-- name: UpdateTenantSessionAccess :exec
UPDATE tenant_session SET
    last_accessed_at = NOW()
WHERE session_id = sqlc.arg('session_id')
  AND tenant_uuid = sqlc.arg('tenant_uuid')::uuid;

-- name: DeleteTenantSession :exec
DELETE FROM tenant_session
WHERE session_id = sqlc.arg('session_id')
  AND tenant_uuid = sqlc.arg('tenant_uuid')::uuid;

-- name: DeleteExpiredTenantSessions :exec
DELETE FROM tenant_session WHERE expires_at < NOW();

-- name: DeactivateTenantSession :exec
UPDATE tenant_session SET
    is_active = false
WHERE session_id = sqlc.arg('session_id')
  AND tenant_uuid = sqlc.arg('tenant_uuid')::uuid;
