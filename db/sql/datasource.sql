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
             sqlc.arg('uuid')::uuid,
             sqlc.arg('user_uuid')::uuid,
             NULLIF(sqlc.arg('name'), ''),
             NULLIF(sqlc.arg('type'), ''),
             sqlc.arg('is_enabled')::boolean,
             sqlc.arg('provider'),
             sqlc.arg('settings'),
             NOW(),
             NOW()
         ) RETURNING *;

-- name: GetDatasource :one
SELECT
    sqlc.embed(datasource)
FROM datasource
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListDatasources :many
SELECT
    sqlc.embed(datasource)
FROM datasource
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

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
    user_uuid  = sqlc.arg('user_uuid')::uuid,
    "type"     = NULLIF(sqlc.arg('type'), ''),
    name       =  NULLIF(sqlc.arg('name'), ''),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    provider   = sqlc.arg('provider'),
    settings   = sqlc.arg('settings'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteDatasource :exec
DELETE FROM datasource WHERE uuid = sqlc.arg('uuid')::uuid;
