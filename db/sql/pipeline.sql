-- name: CreatePipeline :one
INSERT INTO pipeline (
  uuid,
  datasource_uuid,
  storage_uuid,
  name,
  type,
  is_enabled,
  flow,
  created_at,
  updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('datasource_uuid')::uuid,
             sqlc.arg('storage_uuid')::uuid,
             NULLIF(sqlc.arg('name'), ''),
             NULLIF(sqlc.arg('type'), ''),
  sqlc.arg('is_enabled')::boolean,
              sqlc.arg('flow'),
             NOW(),
  NOW()
) RETURNING *;

-- name: GetPipeline :one
SELECT
    sqlc.embed(pipeline)
FROM pipeline
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListPipelines :many
SELECT
    sqlc.embed(pipeline)
FROM pipeline
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: GetPipelines :many
WITH filtered_pipelines AS (
    SELECT p.*
    FROM pipeline p
    WHERE
        (NULLIF(sqlc.arg('uuid'), '') IS NULL OR p.uuid = sqlc.arg('uuid')::uuid)
      AND (NULLIF(sqlc.arg('datasource_uuid'), '') IS NULL OR p.datasource_uuid = sqlc.arg('datasource_uuid')::uuid)
      AND (NULLIF(sqlc.arg('storage_uuid'), '') IS NULL OR p.storage_uuid = sqlc.arg('storage_uuid')::uuid)
      AND (NULLIF(sqlc.arg('type'), '') IS NULL OR p.type = sqlc.arg('type'))
      AND (NULLIF(sqlc.arg('is_enabled')::int, -1) IS NULL OR p.is_enabled = sqlc.arg('is_enabled')::boolean)
      AND (NULLIF(sqlc.arg('name'), '') IS NULL OR p.name ILIKE '%' || sqlc.arg('name') || '%')
)
SELECT
    *,
    (SELECT count(*) FROM filtered_pipelines) AS total_count
FROM filtered_pipelines
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'name' AND sqlc.arg('order_direction') = 'asc' THEN name END ASC,
    CASE WHEN sqlc.arg('order_by') = 'name' AND sqlc.arg('order_direction') = 'desc' THEN name END DESC,
    CASE WHEN sqlc.arg('order_by') = 'type' AND sqlc.arg('order_direction') = 'asc' THEN type END ASC,
    CASE WHEN sqlc.arg('order_by') = 'type' AND sqlc.arg('order_direction') = 'desc' THEN type END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;



-- name: UpdatePipeline :exec
UPDATE pipeline SET
  "name" = NULLIF(sqlc.arg('name'), ''),
  "type" = NULLIF(sqlc.arg('type'), ''),
  datasource_uuid = sqlc.arg('datasource_uuid')::uuid,
  storage_uuid = sqlc.arg('storage_uuid')::uuid,
  is_enabled = sqlc.arg('is_enabled')::boolean,
  flow = sqlc.arg('flow'),
  updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeletePipeline :exec
DELETE FROM pipeline WHERE uuid = sqlc.arg('uuid')::uuid;