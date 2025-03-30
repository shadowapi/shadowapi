// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: tg_cached_chats.sql

package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const tgCreateCachedChat = `-- name: TgCreateCachedChat :exec
INSERT INTO tg_cached_chats (
    id,
    title,
    raw,
    raw_full,
    fk_session_id
) VALUES (
             $1,          -- id
             $2,       -- title
             $3,         -- raw
             $4,    -- raw_full
             $5   -- fk_session_id
         )
ON CONFLICT (id, fk_session_id) DO UPDATE
    SET title = COALESCE(EXCLUDED.title, tg_cached_chats.title),
        raw = COALESCE(EXCLUDED.raw, tg_cached_chats.raw),
        raw_full = COALESCE(EXCLUDED.raw_full, tg_cached_chats.raw_full)
`

type TgCreateCachedChatParams struct {
	ID        int64       `json:"id"`
	Title     pgtype.Text `json:"title"`
	Raw       []byte      `json:"raw"`
	RawFull   []byte      `json:"raw_full"`
	SessionID int64       `json:"session_id"`
}

func (q *Queries) TgCreateCachedChat(ctx context.Context, arg TgCreateCachedChatParams) error {
	_, err := q.db.Exec(ctx, tgCreateCachedChat,
		arg.ID,
		arg.Title,
		arg.Raw,
		arg.RawFull,
		arg.SessionID,
	)
	return err
}

const tgGetCachedChat = `-- name: TgGetCachedChat :one
SELECT
    id,
    title,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_chats
WHERE
    fk_session_id = $1
  AND id = $2
`

type TgGetCachedChatParams struct {
	SessionID int64 `json:"session_id"`
	ID        int64 `json:"id"`
}

func (q *Queries) TgGetCachedChat(ctx context.Context, arg TgGetCachedChatParams) (TgCachedChat, error) {
	row := q.db.QueryRow(ctx, tgGetCachedChat, arg.SessionID, arg.ID)
	var i TgCachedChat
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Raw,
		&i.RawFull,
		&i.FkSessionID,
	)
	return i, err
}

const tgListCachedChats = `-- name: TgListCachedChats :many
SELECT
    id,
    title,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_chats
WHERE
    fk_session_id = $1
ORDER BY id
LIMIT $3 OFFSET $2
`

type TgListCachedChatsParams struct {
	SessionID int64 `json:"session_id"`
	Offset    int32 `json:"offset"`
	Limit     int32 `json:"limit"`
}

func (q *Queries) TgListCachedChats(ctx context.Context, arg TgListCachedChatsParams) ([]TgCachedChat, error) {
	rows, err := q.db.Query(ctx, tgListCachedChats, arg.SessionID, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TgCachedChat
	for rows.Next() {
		var i TgCachedChat
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Raw,
			&i.RawFull,
			&i.FkSessionID,
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
