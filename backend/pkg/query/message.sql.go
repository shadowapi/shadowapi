// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: message.sql

package query

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createMessage = `-- name: CreateMessage :one
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
             $1::uuid,
             $2,
             $3,
             $4::uuid,
             $5::uuid,
             $6,
             $7,
             $8,
             $9,
             $10,
             $11,
             $12,
             $13,
             $14::uuid,
             $15::uuid,
             $16::uuid,
             $17,
             $18,
             NOW(),
             NOW()
         ) RETURNING uuid, format, type, chat_uuid, thread_uuid, sender, recipients, subject, body, body_parsed, reactions, attachments, forward_from, reply_to_message_uuid, forward_from_chat_uuid, forward_from_message_uuid, forward_meta, meta, created_at, updated_at
`

type CreateMessageParams struct {
	UUID                   pgtype.UUID `json:"uuid"`
	Format                 string      `json:"format"`
	Type                   string      `json:"type"`
	ChatUuid               pgtype.UUID `json:"chat_uuid"`
	ThreadUuid             pgtype.UUID `json:"thread_uuid"`
	Sender                 string      `json:"sender"`
	Recipients             []string    `json:"recipients"`
	Subject                pgtype.Text `json:"subject"`
	Body                   string      `json:"body"`
	BodyParsed             []byte      `json:"body_parsed"`
	Reactions              []byte      `json:"reactions"`
	Attachments            []byte      `json:"attachments"`
	ForwardFrom            pgtype.Text `json:"forward_from"`
	ReplyToMessageUuid     pgtype.UUID `json:"reply_to_message_uuid"`
	ForwardFromChatUuid    pgtype.UUID `json:"forward_from_chat_uuid"`
	ForwardFromMessageUuid pgtype.UUID `json:"forward_from_message_uuid"`
	ForwardMeta            []byte      `json:"forward_meta"`
	Meta                   []byte      `json:"meta"`
}

func (q *Queries) CreateMessage(ctx context.Context, arg CreateMessageParams) (Message, error) {
	row := q.db.QueryRow(ctx, createMessage,
		arg.UUID,
		arg.Format,
		arg.Type,
		arg.ChatUuid,
		arg.ThreadUuid,
		arg.Sender,
		arg.Recipients,
		arg.Subject,
		arg.Body,
		arg.BodyParsed,
		arg.Reactions,
		arg.Attachments,
		arg.ForwardFrom,
		arg.ReplyToMessageUuid,
		arg.ForwardFromChatUuid,
		arg.ForwardFromMessageUuid,
		arg.ForwardMeta,
		arg.Meta,
	)
	var i Message
	err := row.Scan(
		&i.UUID,
		&i.Format,
		&i.Type,
		&i.ChatUuid,
		&i.ThreadUuid,
		&i.Sender,
		&i.Recipients,
		&i.Subject,
		&i.Body,
		&i.BodyParsed,
		&i.Reactions,
		&i.Attachments,
		&i.ForwardFrom,
		&i.ReplyToMessageUuid,
		&i.ForwardFromChatUuid,
		&i.ForwardFromMessageUuid,
		&i.ForwardMeta,
		&i.Meta,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteMessage = `-- name: DeleteMessage :exec
DELETE FROM message
WHERE uuid = $1::uuid
`

func (q *Queries) DeleteMessage(ctx context.Context, argUuid pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteMessage, argUuid)
	return err
}

const getMessage = `-- name: GetMessage :one
SELECT
    message.uuid, message.format, message.type, message.chat_uuid, message.thread_uuid, message.sender, message.recipients, message.subject, message.body, message.body_parsed, message.reactions, message.attachments, message.forward_from, message.reply_to_message_uuid, message.forward_from_chat_uuid, message.forward_from_message_uuid, message.forward_meta, message.meta, message.created_at, message.updated_at
FROM message
WHERE uuid = $1::uuid
`

type GetMessageRow struct {
	Message Message `json:"message"`
}

func (q *Queries) GetMessage(ctx context.Context, argUuid pgtype.UUID) (GetMessageRow, error) {
	row := q.db.QueryRow(ctx, getMessage, argUuid)
	var i GetMessageRow
	err := row.Scan(
		&i.Message.UUID,
		&i.Message.Format,
		&i.Message.Type,
		&i.Message.ChatUuid,
		&i.Message.ThreadUuid,
		&i.Message.Sender,
		&i.Message.Recipients,
		&i.Message.Subject,
		&i.Message.Body,
		&i.Message.BodyParsed,
		&i.Message.Reactions,
		&i.Message.Attachments,
		&i.Message.ForwardFrom,
		&i.Message.ReplyToMessageUuid,
		&i.Message.ForwardFromChatUuid,
		&i.Message.ForwardFromMessageUuid,
		&i.Message.ForwardMeta,
		&i.Message.Meta,
		&i.Message.CreatedAt,
		&i.Message.UpdatedAt,
	)
	return i, err
}

const getMessages = `-- name: GetMessages :many
WITH filtered_messages AS (
    SELECT m.uuid, m.format, m.type, m.chat_uuid, m.thread_uuid, m.sender, m.recipients, m.subject, m.body, m.body_parsed, m.reactions, m.attachments, m.forward_from, m.reply_to_message_uuid, m.forward_from_chat_uuid, m.forward_from_message_uuid, m.forward_meta, m.meta, m.created_at, m.updated_at
    FROM message m
    WHERE
        (NULLIF($5, '') IS NULL OR m.type = $5) AND
        (NULLIF($6, '') IS NULL OR m.format = $6) AND
        ($7::uuid IS NULL OR m.chat_uuid = $7::uuid) AND
        ($8::uuid IS NULL OR m.thread_uuid = $8::uuid) AND
        (NULLIF($9, '') IS NULL OR m.sender = $9)
)
SELECT
    uuid, format, type, chat_uuid, thread_uuid, sender, recipients, subject, body, body_parsed, reactions, attachments, forward_from, reply_to_message_uuid, forward_from_chat_uuid, forward_from_message_uuid, forward_meta, meta, created_at, updated_at,
    (SELECT count(*) FROM filtered_messages) as total_count
FROM filtered_messages
ORDER BY
    CASE WHEN $1 = 'created_at' AND $2 = 'asc' THEN created_at END ASC,
    CASE WHEN $1 = 'created_at' AND $2 = 'desc' THEN created_at END DESC,
    CASE WHEN $1 = 'updated_at' AND $2 = 'asc' THEN updated_at END ASC,
    CASE WHEN $1 = 'updated_at' AND $2 = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF($4::int, 0)
OFFSET $3::int
`

type GetMessagesParams struct {
	OrderBy        interface{} `json:"order_by"`
	OrderDirection interface{} `json:"order_direction"`
	Offset         int32       `json:"offset"`
	Limit          int32       `json:"limit"`
	Type           interface{} `json:"type"`
	Format         interface{} `json:"format"`
	ChatUuid       pgtype.UUID `json:"chat_uuid"`
	ThreadUuid     pgtype.UUID `json:"thread_uuid"`
	Sender         interface{} `json:"sender"`
}

type GetMessagesRow struct {
	UUID                   uuid.UUID          `json:"uuid"`
	Format                 string             `json:"format"`
	Type                   string             `json:"type"`
	ChatUuid               *uuid.UUID         `json:"chat_uuid"`
	ThreadUuid             *uuid.UUID         `json:"thread_uuid"`
	Sender                 string             `json:"sender"`
	Recipients             []string           `json:"recipients"`
	Subject                pgtype.Text        `json:"subject"`
	Body                   string             `json:"body"`
	BodyParsed             []byte             `json:"body_parsed"`
	Reactions              []byte             `json:"reactions"`
	Attachments            []byte             `json:"attachments"`
	ForwardFrom            pgtype.Text        `json:"forward_from"`
	ReplyToMessageUuid     *uuid.UUID         `json:"reply_to_message_uuid"`
	ForwardFromChatUuid    *uuid.UUID         `json:"forward_from_chat_uuid"`
	ForwardFromMessageUuid *uuid.UUID         `json:"forward_from_message_uuid"`
	ForwardMeta            []byte             `json:"forward_meta"`
	Meta                   []byte             `json:"meta"`
	CreatedAt              pgtype.Timestamptz `json:"created_at"`
	UpdatedAt              pgtype.Timestamptz `json:"updated_at"`
	TotalCount             int64              `json:"total_count"`
}

func (q *Queries) GetMessages(ctx context.Context, arg GetMessagesParams) ([]GetMessagesRow, error) {
	rows, err := q.db.Query(ctx, getMessages,
		arg.OrderBy,
		arg.OrderDirection,
		arg.Offset,
		arg.Limit,
		arg.Type,
		arg.Format,
		arg.ChatUuid,
		arg.ThreadUuid,
		arg.Sender,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMessagesRow
	for rows.Next() {
		var i GetMessagesRow
		if err := rows.Scan(
			&i.UUID,
			&i.Format,
			&i.Type,
			&i.ChatUuid,
			&i.ThreadUuid,
			&i.Sender,
			&i.Recipients,
			&i.Subject,
			&i.Body,
			&i.BodyParsed,
			&i.Reactions,
			&i.Attachments,
			&i.ForwardFrom,
			&i.ReplyToMessageUuid,
			&i.ForwardFromChatUuid,
			&i.ForwardFromMessageUuid,
			&i.ForwardMeta,
			&i.Meta,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.TotalCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listMessages = `-- name: ListMessages :many
SELECT
    message.uuid, message.format, message.type, message.chat_uuid, message.thread_uuid, message.sender, message.recipients, message.subject, message.body, message.body_parsed, message.reactions, message.attachments, message.forward_from, message.reply_to_message_uuid, message.forward_from_chat_uuid, message.forward_from_message_uuid, message.forward_meta, message.meta, message.created_at, message.updated_at
FROM message
ORDER BY created_at DESC
LIMIT NULLIF($2::int, 0)
OFFSET $1::int
`

type ListMessagesParams struct {
	Offset int32 `json:"offset"`
	Limit  int32 `json:"limit"`
}

type ListMessagesRow struct {
	Message Message `json:"message"`
}

func (q *Queries) ListMessages(ctx context.Context, arg ListMessagesParams) ([]ListMessagesRow, error) {
	rows, err := q.db.Query(ctx, listMessages, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListMessagesRow
	for rows.Next() {
		var i ListMessagesRow
		if err := rows.Scan(
			&i.Message.UUID,
			&i.Message.Format,
			&i.Message.Type,
			&i.Message.ChatUuid,
			&i.Message.ThreadUuid,
			&i.Message.Sender,
			&i.Message.Recipients,
			&i.Message.Subject,
			&i.Message.Body,
			&i.Message.BodyParsed,
			&i.Message.Reactions,
			&i.Message.Attachments,
			&i.Message.ForwardFrom,
			&i.Message.ReplyToMessageUuid,
			&i.Message.ForwardFromChatUuid,
			&i.Message.ForwardFromMessageUuid,
			&i.Message.ForwardMeta,
			&i.Message.Meta,
			&i.Message.CreatedAt,
			&i.Message.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateMessage = `-- name: UpdateMessage :exec
UPDATE message
SET
    format                    = $1,
    type                      = $2,
    chat_uuid                 = $3::uuid,
    thread_uuid               = $4::uuid,
    sender                    = $5,
    recipients                = $6,
    subject                   = $7,
    body                      = $8,
    body_parsed               = $9,
    reactions                 = $10,
    attachments               = $11,
    forward_from              = $12,
    reply_to_message_uuid     = $13::uuid,
    forward_from_chat_uuid    = $14::uuid,
    forward_from_message_uuid = $15::uuid,
    forward_meta              = $16,
    meta                      = $17,
    updated_at = NOW()
WHERE uuid = $18::uuid
`

type UpdateMessageParams struct {
	Format                 string      `json:"format"`
	Type                   string      `json:"type"`
	ChatUuid               pgtype.UUID `json:"chat_uuid"`
	ThreadUuid             pgtype.UUID `json:"thread_uuid"`
	Sender                 string      `json:"sender"`
	Recipients             []string    `json:"recipients"`
	Subject                pgtype.Text `json:"subject"`
	Body                   string      `json:"body"`
	BodyParsed             []byte      `json:"body_parsed"`
	Reactions              []byte      `json:"reactions"`
	Attachments            []byte      `json:"attachments"`
	ForwardFrom            pgtype.Text `json:"forward_from"`
	ReplyToMessageUuid     pgtype.UUID `json:"reply_to_message_uuid"`
	ForwardFromChatUuid    pgtype.UUID `json:"forward_from_chat_uuid"`
	ForwardFromMessageUuid pgtype.UUID `json:"forward_from_message_uuid"`
	ForwardMeta            []byte      `json:"forward_meta"`
	Meta                   []byte      `json:"meta"`
	UUID                   pgtype.UUID `json:"uuid"`
}

func (q *Queries) UpdateMessage(ctx context.Context, arg UpdateMessageParams) error {
	_, err := q.db.Exec(ctx, updateMessage,
		arg.Format,
		arg.Type,
		arg.ChatUuid,
		arg.ThreadUuid,
		arg.Sender,
		arg.Recipients,
		arg.Subject,
		arg.Body,
		arg.BodyParsed,
		arg.Reactions,
		arg.Attachments,
		arg.ForwardFrom,
		arg.ReplyToMessageUuid,
		arg.ForwardFromChatUuid,
		arg.ForwardFromMessageUuid,
		arg.ForwardMeta,
		arg.Meta,
		arg.UUID,
	)
	return err
}
