-- name: CreatePipeline :one
INSERT INTO pipeline (
  uuid,
  workspace_uuid,
  datasource_uuid,
  storage_uuid,
  worker_uuid,
  name,
  type,
  is_enabled,
  flow,
  created_at,
  updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('workspace_uuid')::uuid,
             sqlc.arg('datasource_uuid')::uuid,
             sqlc.arg('storage_uuid')::uuid,
             sqlc.narg('worker_uuid')::uuid,
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

-- name: GetPipelineByWorkspace :one
SELECT
    sqlc.embed(pipeline)
FROM pipeline
WHERE uuid = sqlc.arg('uuid')::uuid
  AND workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: ListPipelines :many
SELECT
    sqlc.embed(pipeline)
FROM pipeline
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: ListPipelinesByWorkspace :many
SELECT
    sqlc.embed(pipeline)
FROM pipeline
WHERE workspace_uuid = sqlc.arg('workspace_uuid')::uuid
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: GetPipelines :many
WITH filtered_pipelines AS (
    SELECT p.*
    FROM pipeline p
    WHERE
        (sqlc.arg('workspace_uuid')::uuid IS NULL OR p.workspace_uuid = sqlc.arg('workspace_uuid')::uuid) AND
        (NULLIF(sqlc.arg('uuid'), '') IS NULL OR p.uuid = sqlc.arg('uuid')::uuid) AND
        (NULLIF(sqlc.arg('datasource_uuid'), '') IS NULL OR p.datasource_uuid = sqlc.arg('datasource_uuid')::uuid) AND
        (NULLIF(sqlc.arg('storage_uuid'), '') IS NULL OR p.storage_uuid = sqlc.arg('storage_uuid')::uuid) AND
        (NULLIF(sqlc.arg('type'), '') IS NULL OR p.type = sqlc.arg('type')) AND
        (NULLIF(sqlc.arg('is_enabled')::int, -1) IS NULL OR p.is_enabled = sqlc.arg('is_enabled')::boolean) AND
        (NULLIF(sqlc.arg('name'), '') IS NULL OR p.name ILIKE '%' || sqlc.arg('name') || '%')
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
  worker_uuid = sqlc.narg('worker_uuid')::uuid,
  is_enabled = sqlc.arg('is_enabled')::boolean,
  flow = sqlc.arg('flow'),
  updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: UpdatePipelineByWorkspace :exec
UPDATE pipeline SET
  "name" = NULLIF(sqlc.arg('name'), ''),
  "type" = NULLIF(sqlc.arg('type'), ''),
  datasource_uuid = sqlc.arg('datasource_uuid')::uuid,
  storage_uuid = sqlc.arg('storage_uuid')::uuid,
  worker_uuid = sqlc.narg('worker_uuid')::uuid,
  is_enabled = sqlc.arg('is_enabled')::boolean,
  flow = sqlc.arg('flow'),
  updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
  AND workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: DeletePipeline :exec
DELETE FROM pipeline WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeletePipelineByWorkspace :exec
DELETE FROM pipeline
WHERE uuid = sqlc.arg('uuid')::uuid
  AND workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: GetPipelineWorkspaceSlug :one
SELECT w.slug
FROM pipeline p
JOIN workspace w ON p.workspace_uuid = w.uuid
WHERE p.uuid = sqlc.arg('pipeline_uuid')::uuid;

-- name: GetPipelineWithDetails :one
-- Get a pipeline with its associated datasource and storage settings for scheduler
SELECT
    p.uuid as pipeline_uuid,
    p.workspace_uuid,
    p.name as pipeline_name,
    p.type as pipeline_type,
    p.is_enabled as pipeline_enabled,
    p.flow as pipeline_flow,
    w.slug as workspace_slug,
    d.uuid as datasource_uuid,
    d.name as datasource_name,
    d.type as datasource_type,
    d.provider as datasource_provider,
    d.settings as datasource_settings,
    s.uuid as storage_uuid,
    s.name as storage_name,
    s.type as storage_type,
    s.settings as storage_settings
FROM pipeline p
INNER JOIN workspace w ON p.workspace_uuid = w.uuid
INNER JOIN datasource d ON p.datasource_uuid = d.uuid
INNER JOIN storage s ON p.storage_uuid = s.uuid
WHERE p.uuid = sqlc.arg('pipeline_uuid')::uuid;
