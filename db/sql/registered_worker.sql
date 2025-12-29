-- name: CreateRegisteredWorker :one
INSERT INTO registered_worker (
    uuid, name, secret_hash, status, is_global, version, labels, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW()
) RETURNING *;

-- name: GetRegisteredWorker :one
SELECT * FROM registered_worker
WHERE uuid = $1;

-- name: GetRegisteredWorkerByName :one
SELECT * FROM registered_worker
WHERE name = $1;

-- name: UpdateWorkerHeartbeat :exec
UPDATE registered_worker
SET last_heartbeat = NOW(), status = $2, updated_at = NOW()
WHERE uuid = $1;

-- name: UpdateWorkerConnected :exec
UPDATE registered_worker
SET status = 'online', last_connected_at = NOW(), connected_from = $2, updated_at = NOW()
WHERE uuid = $1;

-- name: UpdateWorkerDisconnected :exec
UPDATE registered_worker
SET status = 'offline', updated_at = NOW()
WHERE uuid = $1;

-- name: UpdateRegisteredWorker :one
UPDATE registered_worker
SET name = $2, is_global = $3, version = $4, labels = $5, updated_at = NOW()
WHERE uuid = $1
RETURNING *;

-- name: ListRegisteredWorkers :many
SELECT * FROM registered_worker
ORDER BY created_at DESC
LIMIT NULLIF($1::int, 0)
OFFSET $2;

-- name: DeleteRegisteredWorker :exec
DELETE FROM registered_worker WHERE uuid = $1;
