-- name: CreateDatasource :one
INSERT INTO datasource (
    uuid,
    user_uuid,
    name,
    "type",
    is_enabled,
    provider,
    settings,
    created_at,
    updated_at
) VALUES (
             @uuid,
             @user_uuid,
             @name,
             @type,
             @is_enabled,
             @provider,
             @settings,
             NOW(),
             NOW()
         ) RETURNING *;

-- name: GetDatasource :one
SELECT
    sqlc.embed(datasource)
FROM datasource
WHERE datasource.uuid = @uuid
LIMIT 1;

-- name: ListDatasources :many
SELECT
    sqlc.embed(datasource)
FROM datasource
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
    OFFSET @offset_records::int;

-- name: GetDatasources :many
WITH filtered_datasource AS (
    SELECT d.*
    FROM datasource d
    WHERE
        (@uuid::uuid IS NULL OR d.uuid = @uuid)
      AND (@user_uuid::uuid IS NULL OR d.user_uuid = @user_uuid)
      AND (@type::text IS NULL OR d."type" = @type)
      AND (@provider::text IS NULL OR d.provider ILIKE @provider)
      AND (@is_enabled::bool IS NULL OR d.is_enabled = @is_enabled)
      AND (@name::text IS NULL OR d.name ILIKE @name)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_datasource) AS total_count
FROM filtered_datasource
ORDER BY
    CASE WHEN @order_by = 'created_at' AND @order_direction::text = 'asc' THEN created_at END ASC,
    CASE WHEN @order_by = 'created_at' AND @order_direction::text = 'desc' THEN created_at END DESC,
    CASE WHEN @order_by = 'updated_at' AND @order_direction::text = 'asc' THEN updated_at END ASC,
    CASE WHEN @order_by = 'updated_at' AND @order_direction::text = 'desc' THEN updated_at END DESC,
    CASE WHEN @order_by = 'name' AND @order_direction::text = 'asc' THEN name END ASC,
    CASE WHEN @order_by = 'name' AND @order_direction::text = 'desc' THEN name END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: UpdateDatasource :exec
UPDATE datasource
SET
    user_uuid  = @user_uuid,
    "type"     = @type,
    name       = @name,
    is_enabled = @is_enabled,
    provider   = @provider,
    settings   = @settings,
    updated_at = NOW()
WHERE uuid = @uuid;

-- name: DeleteDatasource :exec
DELETE FROM datasource
WHERE uuid = @uuid;
