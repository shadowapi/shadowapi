-- name: CreateOauth2Client :one
INSERT INTO oauth2_client (
    uuid,
    name,
    provider,
    client_id,
    secret,
    created_at,
    updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('name'),
             sqlc.arg('provider'),
             sqlc.arg('client_id'),
             sqlc.arg('secret'),
             NOW(),
             NOW()
         ) RETURNING *;

-- name: GetOauth2Client :one
SELECT
    sqlc.embed(oauth2_client)
FROM oauth2_client
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListOauth2Clients :many
SELECT
    sqlc.embed(oauth2_client)
FROM oauth2_client
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: GetOauth2Clients :many
WITH filtered_oauth2_clients AS (
    SELECT oc.*
    FROM oauth2_client oc
    WHERE
        (NULLIF(sqlc.arg('name'), '') IS NULL OR oc.name = sqlc.arg('name'))
      AND (NULLIF(sqlc.arg('provider'), '') IS NULL OR oc.provider = sqlc.arg('provider'))
)
SELECT
    *,
    (SELECT count(*) FROM filtered_oauth2_clients) as total_count
FROM filtered_oauth2_clients
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: UpdateOauth2Client :exec
UPDATE oauth2_client SET
                         name = sqlc.arg('name'),
                         provider = sqlc.arg('provider'),
                         client_id = sqlc.arg('client_id'),
                         secret = sqlc.arg('secret'),
                         updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteOauth2Client :exec
DELETE FROM oauth2_client
WHERE uuid = sqlc.arg('uuid')::uuid;
