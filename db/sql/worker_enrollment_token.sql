-- name: CreateEnrollmentToken :one
INSERT INTO worker_enrollment_token (
    uuid, token_hash, name, is_global, workspace_uuids, expires_at, created_by_user_uuid, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW()
) RETURNING *;

-- name: GetEnrollmentToken :one
SELECT * FROM worker_enrollment_token
WHERE uuid = $1;

-- name: GetValidEnrollmentTokenByHash :one
SELECT * FROM worker_enrollment_token
WHERE token_hash = $1
AND used_at IS NULL
AND expires_at > NOW();

-- name: MarkTokenUsed :exec
UPDATE worker_enrollment_token
SET used_at = NOW(), used_by_worker_uuid = $2
WHERE uuid = $1;

-- name: ListEnrollmentTokens :many
SELECT * FROM worker_enrollment_token
ORDER BY created_at DESC
LIMIT NULLIF($1::int, 0)
OFFSET $2;

-- name: DeleteEnrollmentToken :exec
DELETE FROM worker_enrollment_token WHERE uuid = $1;
