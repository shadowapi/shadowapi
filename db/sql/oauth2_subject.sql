-- name: CreateOauth2Subject :one
INSERT INTO oauth2_subject (
    uuid,
    user_uuid,
    client_uuid,
    token,
    created_at,
    updated_at,
    expired_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('user_uuid')::uuid,
             sqlc.arg('client_uuid')::uuid,
             sqlc.arg('token'),
             NOW(),
             NOW(),
             sqlc.arg('expired_at')
         ) RETURNING *;

-- name: GetOauth2Subject :one
SELECT
    sqlc.embed(oauth2_subject)
FROM oauth2_subject
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListOauth2Subjects :many
SELECT
    sqlc.embed(oauth2_subject)
FROM oauth2_subject
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: GetOauth2Subjects :many
WITH filtered_oauth2_subjects AS (
    SELECT os.*
    FROM oauth2_subject os
    WHERE
        (NULLIF(sqlc.arg('user_uuid'), '') IS NULL OR os.user_uuid = sqlc.arg('user_uuid')::uuid)
      AND (NULLIF(sqlc.arg('client_uuid'), '') IS NULL OR os.client_uuid = sqlc.arg('client_uuid')::uuid)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_oauth2_subjects) as total_count
FROM filtered_oauth2_subjects
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset')::int;

-- name: UpdateOauth2Subject :exec
UPDATE oauth2_subject SET
                          token = sqlc.arg('token'),
                          updated_at = NOW(),
                          expired_at = sqlc.arg('expired_at')
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteOauth2Subject :exec
DELETE FROM oauth2_subject
WHERE uuid = sqlc.arg('uuid')::uuid;
