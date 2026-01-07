-- name: CreateUserInvite :one
INSERT INTO user_invite (
    uuid,
    workspace_uuid,
    email,
    role,
    token_hash,
    invited_by_user_uuid,
    expires_at,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW()
) RETURNING *;

-- name: GetUserInvite :one
SELECT * FROM user_invite
WHERE uuid = $1;

-- name: GetUserInviteByEmail :one
SELECT * FROM user_invite
WHERE email = $1
AND accepted_at IS NULL;

-- name: GetValidInviteByTokenHash :one
SELECT
    ui.*,
    w.slug as workspace_slug,
    w.display_name as workspace_display_name,
    u.email as inviter_email,
    u.first_name as inviter_first_name,
    u.last_name as inviter_last_name
FROM user_invite ui
JOIN workspace w ON w.uuid = ui.workspace_uuid
JOIN "user" u ON u.uuid = ui.invited_by_user_uuid
WHERE ui.token_hash = $1
AND ui.accepted_at IS NULL
AND ui.expires_at > NOW();

-- name: ListWorkspaceInvites :many
SELECT
    ui.*,
    u.email as inviter_email,
    u.first_name as inviter_first_name,
    u.last_name as inviter_last_name
FROM user_invite ui
JOIN "user" u ON u.uuid = ui.invited_by_user_uuid
WHERE ui.workspace_uuid = $1
AND ui.accepted_at IS NULL
ORDER BY ui.created_at DESC;

-- name: MarkInviteAccepted :exec
UPDATE user_invite
SET accepted_at = NOW()
WHERE uuid = $1;

-- name: DeleteUserInvite :exec
DELETE FROM user_invite WHERE uuid = $1;

-- name: DeleteExpiredInvites :exec
DELETE FROM user_invite
WHERE expires_at < NOW() AND accepted_at IS NULL;
