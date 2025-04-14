-- name: CreateOauth2State :one
INSERT INTO oauth2_state (
    uuid,
    client_uuid,
    state,
    created_at,
    updated_at,
    expired_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('client_uuid')::uuid,
             sqlc.arg('state'),
             NOW(),
             NOW(),
             sqlc.arg('expired_at')
         ) RETURNING *;

-- name: GetOauth2State :one
SELECT
    sqlc.embed(oauth2_state)
FROM oauth2_state
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListOauth2States :many
SELECT
    sqlc.embed(oauth2_state)
FROM oauth2_state
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: GetOauth2States :many
WITH filtered_oauth2_states AS (
    SELECT os.*
    FROM oauth2_state os
    WHERE
      (NULLIF(sqlc.arg('client_uuid'), '') IS NULL OR os.client_uuid = sqlc.arg('client_uuid')::uuid)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_oauth2_states) as total_count
FROM filtered_oauth2_states
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: UpdateOauth2State :exec
UPDATE oauth2_state SET
                        state = sqlc.arg('state'),
                        updated_at = NOW(),
                        expired_at = sqlc.arg('expired_at')
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteOauth2State :exec
DELETE FROM oauth2_state
WHERE uuid = sqlc.arg('uuid')::uuid;
