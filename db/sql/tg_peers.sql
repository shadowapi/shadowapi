-- name: TgSavePeer :exec
INSERT INTO tg_peers (id, fk_session_id, peer_type, access_hash)
VALUES (sqlc.arg('id'), sqlc.arg('session_id'), sqlc.arg('peer_type'), sqlc.arg('access_hash'))
ON CONFLICT (id, fk_session_id)
    DO UPDATE SET peer_type = EXCLUDED.peer_type,
                  access_hash = COALESCE(EXCLUDED.access_hash, tg_peers.access_hash);

-- name: TgFindPeer :one
SELECT * FROM tg_peers WHERE fk_session_id = sqlc.arg('session_id') AND id = sqlc.arg('id') AND peer_type = sqlc.arg('peer_type');

-- name: TgSavePeerUserPhone :exec
INSERT INTO tg_peers_users (id, fk_session_id, phone)
VALUES (sqlc.arg('id'), sqlc.arg('session_id'), sqlc.arg('phone'))
ON CONFLICT (id, fk_session_id)
    DO UPDATE SET phone = EXCLUDED.phone;

-- name: TgFindPeerUserByPhone :one
SELECT p.id, access_hash
FROM tg_peers_users pu
         JOIN tg_peers p ON p.id = tg_peers_users.id AND p.fk_session_id = tg_peers_users.fk_session_id
WHERE p.fk_session_id = sqlc.arg('session_id') AND phone = sqlc.arg('phone');

-- name: TgGetPeerChannel :one
SELECT * FROM tg_peers_channels WHERE fk_session_id = sqlc.arg('session_id') AND id = sqlc.arg('id');

-- name: TgSetPeerChannelState :exec
INSERT INTO tg_peers_channels (id, fk_session_id, pts)
VALUES (sqlc.arg('id'), sqlc.arg('session_id'), sqlc.arg('pts'))
ON CONFLICT (id, fk_session_id)
    DO UPDATE SET pts = EXCLUDED.pts;

-- name: TgGetPeersChannels :many
SELECT * FROM tg_peers_channels WHERE fk_session_id = sqlc.arg('session_id');
