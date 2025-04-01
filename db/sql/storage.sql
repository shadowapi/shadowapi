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
 sqlc.arg('uuid')::uuid,
 NULLIF(sqlc.arg('name'), ''),
 NULLIF(sqlc.arg('type'), ''),
 sqlc.arg('is_enabled')::boolean,
  sqlc.arg('settings'),
  NOW(),
  NOW()
) RETURNING *;

-- name: GetStorage :one
SELECT
    sqlc.embed(storage)
FROM storage
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListStorages :many
SELECT
    sqlc.embed(storage)
FROM storage
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

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
    OFFSET sqlc.arg('offset')::int;


-- name: UpdateStorage :exec
UPDATE storage SET
  name = NULLIF(sqlc.arg('name'), ''),
  "type" = NULLIF(sqlc.arg('type'), ''),
  is_enabled = sqlc.arg('is_enabled')::boolean,
  settings = sqlc.arg('settings'),
  updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteStorage :exec
DELETE FROM storage WHERE uuid = sqlc.arg('uuid')::uuid;
