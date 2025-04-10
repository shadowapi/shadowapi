-- name: CreateMessage :one
INSERT INTO message (
    uuid,
    format,
    type,
    chat_uuid,
    thread_uuid,
    sender,
    recipients,
    subject,
    body,
    body_parsed,
    reactions,
    attachments,
    forward_from,
    reply_to_message_uuid,
    forward_from_chat_uuid,
    forward_from_message_uuid,
    forward_meta,
    meta,
    created_at,
    updated_at
) VALUES (
             sqlc.arg('uuid')::uuid,
             sqlc.arg('format'),
             sqlc.arg('type'),
             sqlc.arg('chat_uuid')::uuid,
             sqlc.arg('thread_uuid')::uuid,
             sqlc.arg('sender'),
             sqlc.arg('recipients'),
             sqlc.arg('subject'),
             sqlc.arg('body'),
             sqlc.arg('body_parsed'),
             sqlc.arg('reactions'),
             sqlc.arg('attachments'),
             sqlc.arg('forward_from'),
             sqlc.arg('reply_to_message_uuid')::uuid,
             sqlc.arg('forward_from_chat_uuid')::uuid,
             sqlc.arg('forward_from_message_uuid')::uuid,
             sqlc.arg('forward_meta'),
             sqlc.arg('meta'),
             NOW(),
             NOW()
         ) RETURNING *;

-- name: GetMessage :one
SELECT
    sqlc.embed(message)
FROM message
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListMessages :many
SELECT
    sqlc.embed(message)
FROM message
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset')::int;

-- name: GetMessages :many
WITH filtered_messages AS (
    SELECT m.*
    FROM message m
    WHERE
        (NULLIF(sqlc.arg('type'), '') IS NULL OR m.type = sqlc.arg('type')) AND
        (NULLIF(sqlc.arg('format'), '') IS NULL OR m.format = sqlc.arg('format')) AND
        (sqlc.arg('chat_uuid')::uuid IS NULL OR m.chat_uuid = sqlc.arg('chat_uuid')::uuid) AND
        (sqlc.arg('thread_uuid')::uuid IS NULL OR m.thread_uuid = sqlc.arg('thread_uuid')::uuid) AND
        (NULLIF(sqlc.arg('sender'), '') IS NULL OR m.sender = sqlc.arg('sender'))
)
SELECT
    *,
    (SELECT count(*) FROM filtered_messages) as total_count
FROM filtered_messages
ORDER BY
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'created_at' AND sqlc.arg('order_direction') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'asc' THEN updated_at END ASC,
    CASE WHEN sqlc.arg('order_by') = 'updated_at' AND sqlc.arg('order_direction') = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
OFFSET sqlc.arg('offset')::int;

-- name: UpdateMessage :exec
UPDATE message
SET
    format                    = sqlc.arg('format'),
    type                      = sqlc.arg('type'),
    chat_uuid                 = sqlc.arg('chat_uuid')::uuid,
    thread_uuid               = sqlc.arg('thread_uuid')::uuid,
    sender                    = sqlc.arg('sender'),
    recipients                = sqlc.arg('recipients'),
    subject                   = sqlc.arg('subject'),
    body                      = sqlc.arg('body'),
    body_parsed               = sqlc.arg('body_parsed'),
    reactions                 = sqlc.arg('reactions'),
    attachments               = sqlc.arg('attachments'),
    forward_from              = sqlc.arg('forward_from'),
    reply_to_message_uuid     = sqlc.arg('reply_to_message_uuid')::uuid,
    forward_from_chat_uuid    = sqlc.arg('forward_from_chat_uuid')::uuid,
    forward_from_message_uuid = sqlc.arg('forward_from_message_uuid')::uuid,
    forward_meta              = sqlc.arg('forward_meta'),
    meta                      = sqlc.arg('meta'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteMessage :exec
DELETE FROM message
WHERE uuid = sqlc.arg('uuid')::uuid;

