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
    SELECT d.*
    FROM storage d
    WHERE
        (NULLIF(sqlc.arg('type'), '') IS NULL OR d."type" = sqlc.arg('type'))
      AND (sqlc.arg('uuid')::uuid IS NULL OR d.uuid = sqlc.arg('uuid'))
      AND (NULLIF(sqlc.arg('is_enabled')::int, -1) IS NULL OR d.is_enabled = sqlc.arg('is_enabled')::boolean)
      AND (NULLIF(sqlc.arg('name'), '') IS NULL OR d.name ILIKE sqlc.arg('name'))
)
SELECT
    *,
    (SELECT count(*) FROM filtered_storages) AS total_count
FROM filtered_storages
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');


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
