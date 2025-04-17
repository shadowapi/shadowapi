-- name: CreateOauth2Token :one
INSERT INTO oauth2_token (
    uuid,
    client_uuid,
    user_uuid,
    token,
    created_at,
    updated_at
) VALUES (
    sqlc.arg('uuid')::uuid,
    sqlc.arg('client_uuid')::uuid,
    sqlc.arg('user_uuid')::uuid,
    sqlc.arg('token'),
    NOW(),
    NOW()
) RETURNING *;

-- name: GetOauth2TokensByClientUUID :many
SELECT
    sqlc.embed(oauth2_token)
FROM oauth2_token
WHERE client_uuid = sqlc.arg('client_uuid')::uuid;

-- name: GetTokensToRefresh :many
SELECT
    sqlc.embed(oauth2_token)
FROM oauth2_token
WHERE
   (NULLIF(sqlc.arg('client_uuid'), '') IS NULL OR os.client_uuid = sqlc.arg('client_uuid')::uuid) AND
    expires_at < NOW() AND
    (updated_at IS NULL OR updated_at < NOW() - INTERVAL '1 hour');

-- name: GetOauth2TokenByUUID :one
SELECT
    sqlc.embed(oauth2_token)
FROM oauth2_token
WHERE client_uuid = sqlc.arg('uuid')::uuid;

-- name: GetOauth2ClientTokens :many
SELECT
    ot.*,
    c.name
FROM oauth2_token AS ot
         LEFT JOIN datasource AS c ON c.oauth2_token_uuid = ot.uuid
WHERE ot.client_uuid = sqlc.arg('client_uuid')::uuid;

-- name: GetOauth2Tokens :many
WITH filtered_oauth2_tokens AS (
    SELECT ot.*
    FROM oauth2_token ot
    WHERE
        (NULLIF(sqlc.arg('client_uuid'), '') IS NULL OR ot.client_uuid = sqlc.arg('client_uuid')::uuid)
)
SELECT
    *,
    (SELECT count(*) FROM filtered_oauth2_tokens) as total_count
FROM filtered_oauth2_tokens
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset')::int;

-- name: UpdateOauth2Token :exec
UPDATE oauth2_token SET
                        client_uuid = sqlc.arg('client_uuid')::uuid,
    user_uuid = sqlc.arg('user_uuid')::uuid,
    token = sqlc.arg('token'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteOauth2Token :exec
DELETE FROM oauth2_token
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteOauth2TokenByClientUUID :exec
DELETE FROM oauth2_token
WHERE client_uuid = sqlc.arg('client_uuid')::uuid;
