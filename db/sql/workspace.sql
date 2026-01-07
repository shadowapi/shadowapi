-- name: GetWorkspaceBySlug :one
SELECT * FROM workspace
WHERE slug = $1;

-- name: GetWorkspaceByUUID :one
SELECT * FROM workspace
WHERE uuid = $1;

-- name: CreateWorkspace :one
INSERT INTO workspace (uuid, slug, display_name, is_enabled, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateWorkspace :one
UPDATE workspace
SET display_name = $2, is_enabled = $3, settings = $4, updated_at = NOW()
WHERE uuid = $1
RETURNING *;

-- name: ListWorkspaces :many
SELECT * FROM workspace
ORDER BY created_at DESC;

-- name: ListWorkspacesByUser :many
SELECT w.* FROM workspace w
INNER JOIN workspace_member wm ON w.uuid = wm.workspace_uuid
WHERE wm.user_uuid = $1
ORDER BY w.created_at DESC;

-- name: DeleteWorkspace :exec
DELETE FROM workspace
WHERE uuid = $1;
