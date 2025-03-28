-- name: CreatePipeline :one
INSERT INTO pipeline (
  uuid, user_uuid, name, flow,
  created_at, updated_at
)
VALUES (
  @uuid,
  @user_uuid,        -- new
  @name,
  @flow,
  NOW(),
  NOW()
) RETURNING *;

-- name: UpdatePipeline :exec
UPDATE pipeline SET
  name = COALESCE(@name, name),
  flow = COALESCE(@flow, flow),
  updated_at = NOW()
WHERE uuid = @uuid
  AND (
    -- only the owning user can update
    user_uuid = @user_uuid
  );

-- name: DeletePipeline :exec
DELETE FROM pipeline WHERE uuid = @uuid
  AND user_uuid = @user_uuid;  -- multi-user safety check

-- name: GetPipelines :many
SELECT
*
FROM pipeline
WHERE 
  -- if all params are null, select all
  (sqlc.narg('uuid')::text IS NULL
   AND sqlc.narg('name')::text IS NULL
   AND sqlc.narg('created_at_from')::timestamp IS NULL
  ) OR
  -- if any param is not null, filter
  (@uuid IS NOT NULL AND "uuid"::text = @uuid::text)
  OR (@name IS NOT NULL AND "name"::text like @name::text)
  OR (@created_at_from IS NOT NULL AND "created_at" >= @created_at_from)
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: AddPipelineEntry :exec
INSERT INTO pipeline_entry (
  uuid, pipeline_uuid, parent_uuid, "type", params, created_at
) VALUES (
  @uuid, @pipeline_uuid, @parent_uuid, @type, @params, NOW()
) RETURNING *;

-- name: GetPipelineEntryByUUID :one
SELECT * FROM pipeline_entry
WHERE uuid = @uuid;

-- name: DeletePipelineEntry :exec
DELETE FROM pipeline_entry
WHERE uuid = @uuid;

-- name: GetPipelineEntries :many
SELECT * FROM pipeline_entry
WHERE
	(@pipeline_uuid <> '' AND "pipeline_uuid" = @pipeline_uuid) OR
	(@parent_uuid <> '' AND "parent_uuid" = @parent_uuid) OR
	(@type <> '' AND "type" = @type) OR
	(@name <> '' AND "name" like @name )
ORDER BY created_at DESC;

-- name: UpdatePipelineEntry :exec
UPDATE pipeline_entry SET
  params = COALESCE(@params, params),
  updated_at = NOW()
WHERE uuid = @uuid;
