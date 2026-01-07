-- name: CreatePasswordReset :one
INSERT INTO password_reset (uuid, user_uuid, email, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetValidPasswordResetByTokenHash :one
SELECT pr.*, u.first_name, u.last_name
FROM password_reset pr
JOIN "user" u ON u.uuid = pr.user_uuid
WHERE pr.token_hash = $1
AND pr.used_at IS NULL
AND pr.expires_at > NOW();

-- name: GetLatestPasswordResetByEmail :one
SELECT * FROM password_reset
WHERE email = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkPasswordResetUsed :exec
UPDATE password_reset
SET used_at = NOW()
WHERE uuid = $1;

-- name: InvalidatePasswordResetsForUser :exec
UPDATE password_reset
SET used_at = NOW()
WHERE user_uuid = $1
AND used_at IS NULL;
