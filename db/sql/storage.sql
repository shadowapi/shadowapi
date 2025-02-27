-- name: CreateStorage :one
INSERT INTO storage (
  uuid,
  "type",
  settings,
  created_at,
  updated_at
) VALUES (
  @uuid,
  @type,
  @settings,
  NOW(),
  NOW()
) RETURNING *;

-- name: GetStorages :many
WITH filtered_storages AS (
  SELECT d.* FROM storage d WHERE 
    (@type::text IS NULL OR "type" = @type) AND
    (@uuid::uuid IS NULL OR "uuid" = @uuid) AND
    (@user_uuid::uuid IS NULL OR user_uuid = @user_uuid) AND
    (@is_enabled::bool IS NULL OR is_enabled = @is_enabled) AND
    (@name::text IS NULL OR d.name ILIKE @name))
SELECT
  *,
  (SELECT count(*) FROM filtered_storages) as total_count
FROM filtered_storages
ORDER BY 
  CASE WHEN @order_by = 'created_at' AND @order_direction::text = 'asc' THEN created_at END ASC,
  CASE WHEN @order_by = 'created_at' AND @order_direction::text = 'desc' THEN created_at END DESC,
  CASE WHEN @order_by = 'updated_at' AND @order_direction::text = 'asc' THEN updated_at END ASC,
  CASE WHEN @order_by = 'updated_at' AND @order_direction::text = 'desc' THEN updated_at END DESC,
  CASE WHEN @order_by = 'name' AND @order_direction::text = 'asc' THEN name END ASC,
  CASE WHEN @order_by = 'name' AND @order_direction::text = 'desc' THEN name END DESC,
  created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0) OFFSET sqlc.arg('offset');

-- name: UpdateStorage :exec
UPDATE storage SET
  "type" = @type,
  name = @name,
  settings = @settings,
  updated_at = NOW()
WHERE uuid = @uuid;

-- name: DeleteStorage :exec
DELETE FROM storage WHERE uuid = @uuid;
