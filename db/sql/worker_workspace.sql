-- name: AddWorkerWorkspace :exec
INSERT INTO worker_workspace (uuid, worker_uuid, workspace_uuid, created_at)
VALUES ($1, $2, $3, NOW());

-- name: RemoveWorkerWorkspace :exec
DELETE FROM worker_workspace
WHERE worker_uuid = $1 AND workspace_uuid = $2;

-- name: GetWorkerWorkspaces :many
SELECT w.* FROM workspace w
JOIN worker_workspace ww ON w.uuid = ww.workspace_uuid
WHERE ww.worker_uuid = $1;

-- name: GetWorkspaceWorkers :many
SELECT rw.* FROM registered_worker rw
JOIN worker_workspace ww ON rw.uuid = ww.worker_uuid
WHERE ww.workspace_uuid = $1;

-- name: ListWorkerWorkspaceLinks :many
SELECT * FROM worker_workspace
WHERE worker_uuid = $1;
