-- name: TgGetSession :one
SELECT * FROM tg_sessions WHERE id = sqlc.arg('id');

-- name: TgGetSessionList :many
SELECT * FROM tg_sessions WHERE account_id = sqlc.arg('account_id');

-- name: TgGetSessionByPhone :one
SELECT * FROM tg_sessions WHERE phone = sqlc.arg('phone');

-- name: TgCreateSession :one
INSERT INTO tg_sessions (account_id, session, phone)
VALUES (sqlc.arg('account_id'), sqlc.arg('session'), sqlc.arg('phone')) RETURNING id;

-- name: TgUpdateSession :exec
UPDATE tg_sessions
SET session = COALESCE(session, sqlc.arg('session')),
    contacts_hash = COALESCE(contacts_hash, sqlc.arg('contacts_hash')),
    updated_at = FLOOR(EXTRACT(EPOCH FROM CURRENT_TIMESTAMP))
WHERE id = sqlc.arg('id');
