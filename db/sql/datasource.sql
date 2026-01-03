-- name: CreateDatasource :one
INSERT INTO datasource (
    uuid,
    workspace_uuid,
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
             sqlc.arg('workspace_uuid')::uuid,
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

-- name: GetDatasourceByWorkspace :one
SELECT
    sqlc.embed(datasource)
FROM datasource
WHERE uuid = sqlc.arg('uuid')::uuid
  AND workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: ListDatasources :many
SELECT
    sqlc.embed(datasource)
FROM datasource
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: ListDatasourcesByWorkspace :many
SELECT
    sqlc.embed(datasource)
FROM datasource
WHERE workspace_uuid = sqlc.arg('workspace_uuid')::uuid
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: GetDatasources :many
WITH filtered_datasource AS (
    SELECT d.*
    FROM datasource d
    WHERE
        (sqlc.arg('workspace_uuid')::uuid IS NULL OR d.workspace_uuid = sqlc.arg('workspace_uuid')::uuid) AND
        (NULLIF(sqlc.arg('uuid'), '') IS NULL OR d.uuid = sqlc.arg('uuid')::uuid) AND
        (NULLIF(sqlc.arg('user_uuid'), '') IS NULL OR d.user_uuid = sqlc.arg('user_uuid')::uuid) AND
        (NULLIF(sqlc.arg('name'), '') IS NULL OR d.name ILIKE '%' || sqlc.arg('name') || '%') AND
        (NULLIF(sqlc.arg('type'), '') IS NULL OR d."type" = sqlc.arg('type')) AND
        (NULLIF(sqlc.arg('provider'), '') IS NULL OR d.provider = sqlc.arg('provider')) AND
        (NULLIF(sqlc.arg('is_enabled')::int, -1) IS NULL OR d.is_enabled = (sqlc.arg('is_enabled')::int)::boolean)
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

-- name: UpdateDatasourceByWorkspace :exec
UPDATE datasource
SET
    user_uuid  = sqlc.arg('user_uuid')::uuid,
    "type"     = NULLIF(sqlc.arg('type'), ''),
    name       =  NULLIF(sqlc.arg('name'), ''),
    is_enabled = sqlc.arg('is_enabled')::boolean,
    provider   = sqlc.arg('provider'),
    settings   = sqlc.arg('settings'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid
  AND workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: DeleteDatasource :exec
DELETE FROM datasource WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteDatasourceByWorkspace :exec
DELETE FROM datasource
WHERE uuid = sqlc.arg('uuid')::uuid
  AND workspace_uuid = sqlc.arg('workspace_uuid')::uuid;

-- name: ListDatasourcesWithOAuthStatus :many
SELECT
    sqlc.embed(datasource),
    CASE
        WHEN datasource.type = 'email_oauth'
             AND datasource.settings->>'oauth2_client_uuid' IS NOT NULL
             AND ot.uuid IS NOT NULL
        THEN true
        ELSE false
    END AS is_oauth_authenticated
FROM datasource
LEFT JOIN oauth2_token ot ON ot.client_uuid = (datasource.settings->>'oauth2_client_uuid')::uuid
ORDER BY datasource.created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset');

-- name: ListDatasourcesWithOAuthStatusByWorkspace :many
SELECT
    sqlc.embed(datasource),
    CASE
        WHEN datasource.type = 'email_oauth'
             AND datasource.settings->>'oauth2_client_uuid' IS NOT NULL
             AND ot.uuid IS NOT NULL
        THEN true
        ELSE false
    END AS is_oauth_authenticated
FROM datasource
LEFT JOIN oauth2_token ot ON ot.client_uuid = (datasource.settings->>'oauth2_client_uuid')::uuid
WHERE datasource.workspace_uuid = sqlc.arg('workspace_uuid')::uuid
ORDER BY datasource.created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset');
