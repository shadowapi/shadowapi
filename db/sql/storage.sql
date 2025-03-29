-- name: CreateStorage :one
INSERT INTO storage (
  uuid,
  name,
  "type",
  is_enabled,
  settings,
  created_at,
  updated_at
) VALUES (
  @uuid,
  @name,
  @type,
  @is_enabled,
  @settings,
  NOW(),
  NOW()
) RETURNING *;

-- name: ListStorages :many
SELECT
    sqlc.embed(storage)
FROM storage
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
    OFFSET @offset_records::int;

-- name: GetStorages :many
WITH filtered_storages AS (
  SELECT d.* FROM storage d WHERE 
    (@type::text IS NULL OR "type" = @type) AND
    (@uuid::uuid IS NULL OR "uuid" = @uuid) AND
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

-- name: GetStorage :one
SELECT
    uuid,
    name,
    "type",
    is_enabled,
    settings,
    created_at,
    updated_at
FROM storage
WHERE uuid = $1
LIMIT 1;

-- name: UpdateStorage :exec
UPDATE storage SET
  "type" = @type,
  name = @name,
  is_enabled = @is_enabled,
  settings = @settings,
  updated_at = NOW()
WHERE uuid = @uuid;

-- name: DeleteStorage :exec
DELETE FROM storage WHERE uuid = @uuid;
