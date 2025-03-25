-- name: CreateSyncPolicy :one
INSERT INTO sync_policy (
    uuid,
    user_id,
    service,
    blocklist,
    exclude_list,
    sync_all,
    settings,
    created_at,
    updated_at
) VALUES (
             @uuid,
             @user_id,
             @service,
             @blocklist,
             @exclude_list,
             @sync_all,
             @settings,
             NOW(),
             NOW()
         ) RETURNING *;

-- name: ListSyncPolicies :many
SELECT
    sqlc.embed(sync_policy)
FROM sync_policy
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
    OFFSET @offset_records::int;

-- name: GetSyncPolicies :many
WITH filtered_sync_policies AS (
    SELECT sp.* FROM sync_policy sp WHERE
        (@service::text IS NULL OR sp.service = @service) AND
        (@uuid::uuid IS NULL OR sp.uuid = @uuid) AND
        (@user_id::uuid IS NULL OR sp.user_id = @user_id) AND
        (@sync_all::bool IS NULL OR sp.sync_all = @sync_all)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_sync_policies) as total_count
FROM filtered_sync_policies
ORDER BY
    CASE WHEN @order_by = 'created_at' AND @order_direction::text = 'asc' THEN created_at END ASC,
    CASE WHEN @order_by = 'created_at' AND @order_direction::text = 'desc' THEN created_at END DESC,
    CASE WHEN @order_by = 'updated_at' AND @order_direction::text = 'asc' THEN updated_at END ASC,
    CASE WHEN @order_by = 'updated_at' AND @order_direction::text = 'desc' THEN updated_at END DESC,
    CASE WHEN @order_by = 'service' AND @order_direction::text = 'asc' THEN service END ASC,
    CASE WHEN @order_by = 'service' AND @order_direction::text = 'desc' THEN service END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0) OFFSET sqlc.arg('offset');

-- name: UpdateSyncPolicy :exec
UPDATE sync_policy SET
                       user_id = @user_id,
                       service = @service,
                       blocklist = @blocklist,
                       exclude_list = @exclude_list,
                       sync_all = @sync_all,
                       settings = @settings,
                       updated_at = NOW()
WHERE uuid = @uuid;

-- name: DeleteSyncPolicy :exec
DELETE FROM sync_policy WHERE uuid = @uuid;
