// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: tg_cached_channels.sql

package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const tgCreateCachedChannel = `-- name: TgCreateCachedChannel :exec
INSERT INTO tg_cached_channels (
    id,
    title,
    username,
    broadcast,
    forum,
    megagroup,
    raw,
    raw_full,
    fk_session_id
) VALUES (
             $1,          -- id
             $2,       -- title
             $3,    -- username
             $4,   -- broadcast
             $5,       -- forum
             $6,   -- megagroup
             $7,         -- raw
             $8,    -- raw_full
             $9   -- fk_session_id
         )
ON CONFLICT (id, fk_session_id) DO UPDATE
    SET title = COALESCE(EXCLUDED.title, tg_cached_channels.title),
        username = COALESCE(EXCLUDED.username, tg_cached_channels.username),
        broadcast = COALESCE(EXCLUDED.broadcast, tg_cached_channels.broadcast),
        forum = COALESCE(EXCLUDED.forum, tg_cached_channels.forum),
        megagroup = COALESCE(EXCLUDED.megagroup, tg_cached_channels.megagroup),
        raw = COALESCE(EXCLUDED.raw, tg_cached_channels.raw),
        raw_full = COALESCE(EXCLUDED.raw_full, tg_cached_channels.raw_full)
`

type TgCreateCachedChannelParams struct {
	ID        int64       `json:"id"`
	Title     pgtype.Text `json:"title"`
	Username  pgtype.Text `json:"username"`
	Broadcast pgtype.Bool `json:"broadcast"`
	Forum     pgtype.Bool `json:"forum"`
	Megagroup pgtype.Bool `json:"megagroup"`
	Raw       []byte      `json:"raw"`
	RawFull   []byte      `json:"raw_full"`
	SessionID int64       `json:"session_id"`
}

func (q *Queries) TgCreateCachedChannel(ctx context.Context, arg TgCreateCachedChannelParams) error {
	_, err := q.db.Exec(ctx, tgCreateCachedChannel,
		arg.ID,
		arg.Title,
		arg.Username,
		arg.Broadcast,
		arg.Forum,
		arg.Megagroup,
		arg.Raw,
		arg.RawFull,
		arg.SessionID,
	)
	return err
}

const tgGetCachedChannel = `-- name: TgGetCachedChannel :one
SELECT
    id,
    title,
    username,
    broadcast,
    forum,
    megagroup,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_channels
WHERE
    fk_session_id = $1
  AND id = $2
`

type TgGetCachedChannelParams struct {
	SessionID int64 `json:"session_id"`
	ID        int64 `json:"id"`
}

func (q *Queries) TgGetCachedChannel(ctx context.Context, arg TgGetCachedChannelParams) (TgCachedChannel, error) {
	row := q.db.QueryRow(ctx, tgGetCachedChannel, arg.SessionID, arg.ID)
	var i TgCachedChannel
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Username,
		&i.Broadcast,
		&i.Forum,
		&i.Megagroup,
		&i.Raw,
		&i.RawFull,
		&i.FkSessionID,
	)
	return i, err
}

const tgListCachedChannels = `-- name: TgListCachedChannels :many
SELECT
    id,
    title,
    username,
    broadcast,
    forum,
    megagroup,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_channels
WHERE
    fk_session_id = $1
ORDER BY
    id
LIMIT $3 OFFSET $2
`

type TgListCachedChannelsParams struct {
	SessionID int64 `json:"session_id"`
	Offset    int32 `json:"offset"`
	Limit     int32 `json:"limit"`
}

func (q *Queries) TgListCachedChannels(ctx context.Context, arg TgListCachedChannelsParams) ([]TgCachedChannel, error) {
	rows, err := q.db.Query(ctx, tgListCachedChannels, arg.SessionID, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TgCachedChannel
	for rows.Next() {
		var i TgCachedChannel
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Username,
			&i.Broadcast,
			&i.Forum,
			&i.Megagroup,
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
