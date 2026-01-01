-- name: GetWorkspaceMember :one
SELECT * FROM workspace_member
WHERE workspace_uuid = $1 AND user_uuid = $2;

-- name: CreateWorkspaceMember :one
INSERT INTO workspace_member (uuid, workspace_uuid, user_uuid, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateWorkspaceMemberRole :one
UPDATE workspace_member
SET role = $3, updated_at = NOW()
WHERE workspace_uuid = $1 AND user_uuid = $2
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
