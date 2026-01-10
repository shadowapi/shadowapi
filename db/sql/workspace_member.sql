-- name: GetWorkspaceMember :one
SELECT * FROM workspace_member
WHERE workspace_uuid = $1 AND user_uuid = $2;

-- name: CreateWorkspaceMember :one
INSERT INTO workspace_member (uuid, workspace_uuid, user_uuid)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListWorkspaceMembersByWorkspace :many
SELECT * FROM workspace_member
WHERE workspace_uuid = $1
ORDER BY created_at;

-- name: ListWorkspaceMembersByUser :many
SELECT * FROM workspace_member
WHERE user_uuid = $1
ORDER BY created_at;

-- name: DeleteWorkspaceMember :exec
DELETE FROM workspace_member
WHERE workspace_uuid = $1 AND user_uuid = $2;
