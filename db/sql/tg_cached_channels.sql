-- name: TgCreateCachedChannel :exec
INSERT INTO tg_cached_channels (
    id,
    title,
    username,
    broadcast,
    forum,
    megagroup,
    raw,
    raw_full,
    fk_session_id
) VALUES (
             sqlc.arg('id'),          -- id
             sqlc.arg('title'),       -- title
             sqlc.arg('username'),    -- username
             sqlc.arg('broadcast'),   -- broadcast
             sqlc.arg('forum'),       -- forum
             sqlc.arg('megagroup'),   -- megagroup
             sqlc.arg('raw'),         -- raw
             sqlc.arg('raw_full'),    -- raw_full
             sqlc.arg('session_id')   -- fk_session_id
         )
ON CONFLICT (id, fk_session_id) DO UPDATE
    SET title = COALESCE(EXCLUDED.title, tg_cached_channels.title),
        username = COALESCE(EXCLUDED.username, tg_cached_channels.username),
        broadcast = COALESCE(EXCLUDED.broadcast, tg_cached_channels.broadcast),
        forum = COALESCE(EXCLUDED.forum, tg_cached_channels.forum),
        megagroup = COALESCE(EXCLUDED.megagroup, tg_cached_channels.megagroup),
        raw = COALESCE(EXCLUDED.raw, tg_cached_channels.raw),
        raw_full = COALESCE(EXCLUDED.raw_full, tg_cached_channels.raw_full);

-- name: TgGetCachedChannel :one
SELECT
    id,
    title,
    username,
    broadcast,
    forum,
    megagroup,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_channels
WHERE
    fk_session_id = sqlc.arg('session_id')
  AND id = sqlc.arg('id');

-- name: TgListCachedChannels :many
SELECT
    id,
    title,
    username,
    broadcast,
    forum,
    megagroup,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_channels
WHERE
    fk_session_id = sqlc.arg('session_id')
ORDER BY
    id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
