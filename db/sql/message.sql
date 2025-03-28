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
             @uuid,
             @source,
             @type,
             @chat_uuid,
             @thread_uuid,
             @sender,
             @recipients,
             @subject,
             @body,
             @body_parsed,
             @reactions,
             @attachments,
             @forward_from,
             @reply_to_message_uuid,
             @forward_from_chat_uuid,
             @forward_from_message_uuid,
             @forward_meta,
             @meta,
             NOW(),
             NULL
         ) RETURNING *;

-- name: GetMessage :one
SELECT
    *
FROM message
WHERE
    uuid = @uuid
LIMIT 1;

-- name: ListMessages :many
SELECT
    *
FROM message
ORDER BY created_at DESC
LIMIT CASE WHEN @limit_records::int = 0 THEN NULL ELSE @limit_records::int END
    OFFSET @offset_records::int;

-- name: UpdateMessage :exec
UPDATE message
SET
    source                    = COALESCE(@source, source),
    type                      = COALESCE(@type, type),
    chat_uuid                 = COALESCE(@chat_uuid, chat_uuid),
    thread_uuid               = COALESCE(@thread_uuid, thread_uuid),
    sender                    = COALESCE(@sender, sender),
    recipients                = COALESCE(@recipients, recipients),
    subject                   = COALESCE(@subject, subject),
    body                      = COALESCE(@body, body),
    body_parsed               = COALESCE(@body_parsed, body_parsed),
    reactions                 = COALESCE(@reactions, reactions),
    attachments               = COALESCE(@attachments, attachments),
    forward_from              = COALESCE(@forward_from, forward_from),
    reply_to_message_uuid     = COALESCE(@reply_to_message_uuid, reply_to_message_uuid),
    forward_from_chat_uuid    = COALESCE(@forward_from_chat_uuid, forward_from_chat_uuid),
    forward_from_message_uuid = COALESCE(@forward_from_message_uuid, forward_from_message_uuid),
    forward_meta              = COALESCE(@forward_meta, forward_meta),
    meta                      = COALESCE(@meta, meta),
    updated_at                = NOW()
WHERE uuid = @uuid;

-- name: DeleteMessage :exec
DELETE FROM message
WHERE uuid = @uuid;

