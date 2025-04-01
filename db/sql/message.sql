-- name: CreateMessage :one
INSERT INTO message (
    uuid,
    source,
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
             NULLIF(sqlc.arg('source'), '')::uuid,
             NULLIF(sqlc.arg('type'), ''),
             NULLIF(sqlc.arg('chat_uuid'), '')::uuid,
             NULLIF(sqlc.arg('thread_uuid'), '')::uuid,
             sqlc.arg('sender'),
             sqlc.arg('recipients'),
             sqlc.arg('subject'),
             sqlc.arg('body'),
             sqlc.arg('body_parsed'),
             sqlc.arg('reactions'),
             sqlc.arg('attachments'),
             sqlc.arg('forward_from'),
             NULLIF(sqlc.arg('reply_to_message_uuid'), '')::uuid,
             NULLIF(sqlc.arg('forward_from_chat_uuid'), '')::uuid,
             NULLIF(sqlc.arg('forward_from_message_uuid'), '')::uuid,
             sqlc.arg('forward_meta'),
             sqlc.arg('meta'),
             NOW(),
             NOW()
         ) RETURNING *;

-- name: GetMessage :one
SELECT
    *
FROM message
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: ListMessages :many
SELECT
    *
FROM message
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg('limit')::int, 0)
    OFFSET sqlc.arg('offset');

-- name: UpdateMessage :exec
UPDATE message
SET
    source                    = NULLIF(sqlc.arg('source'), '')::uuid,
    type                      = NULLIF(sqlc.arg('type'), ''),
    chat_uuid                 = NULLIF(sqlc.arg('chat_uuid'), '')::uuid,
    thread_uuid               = NULLIF(sqlc.arg('thread_uuid'), '')::uuid,
    sender                    =  sqlc.arg('sender'),
    recipients                = sqlc.arg('recipients'),
    subject                   = sqlc.arg('subject'),
    body                      = sqlc.arg('body'),
    body_parsed               =  sqlc.arg('body_parsed'),
    reactions                 = sqlc.arg('reactions'),
    attachments               =  sqlc.arg('attachments'),
    forward_from              = sqlc.arg('forward_from'),
    reply_to_message_uuid     = NULLIF(sqlc.arg('reply_to_message_uuid'), '')::uuid,
    forward_from_chat_uuid    = NULLIF(sqlc.arg('forward_from_chat_uuid'), '')::uuid,
    forward_from_message_uuid = NULLIF(sqlc.arg('forward_from_message_uuid'), '')::uuid,
    forward_meta              = sqlc.arg('forward_meta'),
    meta                      = sqlc.arg('meta'),
    updated_at = NOW()
WHERE uuid = sqlc.arg('uuid')::uuid;

-- name: DeleteMessage :exec
DELETE FROM message WHERE uuid = sqlc.arg('uuid')::uuid;

