-- name: TgGetState :one
SELECT * FROM tg_sessions_states WHERE id = sqlc.arg('id');

-- name: TgUpsertState :exec
INSERT INTO tg_sessions_states (id, pts, qts, "date", seq)
VALUES (sqlc.arg('id'),
        sqlc.arg('pts'),
        sqlc.arg('qts'),
        sqlc.arg('date'),
        sqlc.arg('seq'))
ON CONFLICT (id)
    DO UPDATE SET pts = EXCLUDED.pts,
                  qts = EXCLUDED.qts,
                  date = EXCLUDED.date,
                  seq = EXCLUDED.seq;

-- name: TgUpdatePts :exec
UPDATE tg_sessions_states SET pts = sqlc.arg('pts') WHERE id = sqlc.arg('id');

-- name: TgUpdateQts :exec
UPDATE tg_sessions_states SET qts = sqlc.arg('qts') WHERE id = sqlc.arg('id');

-- name: TgUpdateDate :exec
UPDATE tg_sessions_states SET date = sqlc.arg('date') WHERE id = sqlc.arg('id');

-- name: TgUpdateSeq :exec
UPDATE tg_sessions_states SET seq = sqlc.arg('seq') WHERE id = sqlc.arg('id');
