-- name: CreateSyncPolicy :one
INSERT INTO sync_policy (
    uuid,
    user_uuid,
    service,
    blocklist,
    exclude_list,
    sync_all,
    settings,
    created_at,
    updated_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    NULLIF(sqlc.arg('user_uuid'), '')::uuid,
    NULLIF(sqlc.arg('service'), ''),
    sqlc.arg('blocklist'),
    sqlc.arg('exclude_list'),
    sqlc.arg('sync_all')::boolean,
    sqlc.arg('settings'),
             NOW(),
             NOW()
         ) RETURNING *;

-- name: ListSyncPolicies :many
SELECT
    sqlc.embed(sync_policy)
FROM sync_policy
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit_records')::int, 0)
OFFSET sqlc.arg('offset_records')::int;

-- name: GetSyncPolicies :many
WITH filtered_sync_policies AS (
    SELECT sp.*
    FROM sync_policy sp
    WHERE
        (NULLIF(sqlc.arg('service'), '') IS NULL OR sp.service = sqlc.arg('service')) AND
        (NULLIF(sqlc.arg('uuid'), '') IS NULL OR sp.uuid = sqlc.arg('uuid')::uuid) AND
        (NULLIF(sqlc.arg('user_uuid'), '') IS NULL OR sp.user_uuid = sqlc.arg('user_uuid')::uuid) AND
        (NULLIF(sqlc.arg('sync_all')::int, -1) IS NULL OR sp.sync_all = (sqlc.arg('sync_all')::int)::boolean)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_sync_policies) as total_count
FROM filtered_sync_policies
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'service' AND sqlc.arg('order_direction') = 'asc' THEN service END ASC,
    CASE WHEN sqlc.arg('order_by') = 'service' AND sqlc.arg('order_direction') = 'desc' THEN service END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset')::int;

-- name: UpdateSyncPolicy :exec
UPDATE sync_policy SET
    user_uuid = NULLIF(sqlc.arg('user_uuid'), '')::uuid,
    service = NULLIF(sqlc.arg('service'), ''),
    blocklist = sqlc.arg('blocklist'),
    exclude_list = sqlc.arg('exclude_list'),
    sync_all = sqlc.arg('sync_all')::boolean,
    settings = sqlc.arg('settings'),
                       updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteSyncPolicy :exec
DELETE FROM sync_policy WHERE uuid = sqlc.arg('uuid')::uuid;
