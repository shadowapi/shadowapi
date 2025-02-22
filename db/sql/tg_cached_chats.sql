-- name: TgCreateCachedChat :exec
INSERT INTO tg_cached_chats (
    id,
    title,
    raw,
    raw_full,
    fk_session_id
) VALUES (
             sqlc.arg('id'),          -- id
             sqlc.arg('title'),       -- title
             sqlc.arg('raw'),         -- raw
             sqlc.arg('raw_full'),    -- raw_full
             sqlc.arg('session_id')   -- fk_session_id
         )
ON CONFLICT (id, fk_session_id) DO UPDATE
    SET title = COALESCE(EXCLUDED.title, tg_cached_chats.title),
        raw = COALESCE(EXCLUDED.raw, tg_cached_chats.raw),
        raw_full = COALESCE(EXCLUDED.raw_full, tg_cached_chats.raw_full);

-- name: TgGetCachedChat :one
SELECT
    id,
    title,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_chats
WHERE
    fk_session_id = sqlc.arg('session_id')
  AND id = sqlc.arg('id');

-- name: TgListCachedChats :many
SELECT
    id,
    title,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_chats
WHERE
    fk_session_id = sqlc.arg('session_id')
ORDER BY id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
