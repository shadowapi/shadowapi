-- name: TgCreateCachedUser :exec
INSERT INTO tg_cached_users (
    id,
    first_name,
    last_name,
    username,
    phone,
    raw,
    raw_full,
    fk_session_id
) VALUES (
             sqlc.arg('id'),          -- id
             sqlc.arg('first_name'),  -- first_name
             sqlc.arg('last_name'),   -- last_name
             sqlc.arg('username'),    -- username
             sqlc.arg('phone'),       -- phone
             sqlc.arg('raw'),         -- raw
             sqlc.arg('raw_full'),    -- raw_full
             sqlc.arg('session_id')   -- fk_session_id
         )
ON CONFLICT (id, fk_session_id) DO UPDATE
    SET first_name =  COALESCE(tg_cached_users.first_name, EXCLUDED.first_name),
        last_name =   COALESCE(tg_cached_users.last_name, EXCLUDED.last_name),
        username =    COALESCE(tg_cached_users.username, EXCLUDED.username),
        phone =       COALESCE(tg_cached_users.phone, EXCLUDED.phone),
        raw =         COALESCE(tg_cached_users.raw, EXCLUDED.raw),
        raw_full =    COALESCE(tg_cached_users.raw_full, EXCLUDED.raw_full);

-- name: TgGetCachedUser :one
SELECT
    id,
    first_name,
    last_name,
    username,
    phone,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_users
WHERE
    fk_session_id = sqlc.arg('session_id')
  AND id = sqlc.arg('id');

-- name: TgListCachedUsers :many
SELECT
    id,
    first_name,
    last_name,
    username,
    phone,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_users
WHERE
    fk_session_id = sqlc.arg('session_id')
ORDER BY id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
