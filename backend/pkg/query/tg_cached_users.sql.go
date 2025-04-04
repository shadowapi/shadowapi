// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: tg_cached_users.sql

package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const tgCreateCachedUser = `-- name: TgCreateCachedUser :exec
INSERT INTO tg_cached_users (
    id,
    first_name,
    last_name,
    username,
    phone,
    raw,
    raw_full,
    fk_session_id
) VALUES (
             $1,          -- id
             $2,  -- first_name
             $3,   -- last_name
             $4,    -- username
             $5,       -- phone
             $6,         -- raw
             $7,    -- raw_full
             $8   -- fk_session_id
         )
ON CONFLICT (id, fk_session_id) DO UPDATE
    SET first_name =  COALESCE(tg_cached_users.first_name, EXCLUDED.first_name),
        last_name =   COALESCE(tg_cached_users.last_name, EXCLUDED.last_name),
        username =    COALESCE(tg_cached_users.username, EXCLUDED.username),
        phone =       COALESCE(tg_cached_users.phone, EXCLUDED.phone),
        raw =         COALESCE(tg_cached_users.raw, EXCLUDED.raw),
        raw_full =    COALESCE(tg_cached_users.raw_full, EXCLUDED.raw_full)
`

type TgCreateCachedUserParams struct {
	ID        int64       `json:"id"`
	FirstName pgtype.Text `json:"first_name"`
	LastName  pgtype.Text `json:"last_name"`
	Username  pgtype.Text `json:"username"`
	Phone     pgtype.Text `json:"phone"`
	Raw       []byte      `json:"raw"`
	RawFull   []byte      `json:"raw_full"`
	SessionID int64       `json:"session_id"`
}

func (q *Queries) TgCreateCachedUser(ctx context.Context, arg TgCreateCachedUserParams) error {
	_, err := q.db.Exec(ctx, tgCreateCachedUser,
		arg.ID,
		arg.FirstName,
		arg.LastName,
		arg.Username,
		arg.Phone,
		arg.Raw,
		arg.RawFull,
		arg.SessionID,
	)
	return err
}

const tgGetCachedUser = `-- name: TgGetCachedUser :one
SELECT
    id,
    first_name,
    last_name,
    username,
    phone,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_users
WHERE
    fk_session_id = $1
  AND id = $2
`

type TgGetCachedUserParams struct {
	SessionID int64 `json:"session_id"`
	ID        int64 `json:"id"`
}

func (q *Queries) TgGetCachedUser(ctx context.Context, arg TgGetCachedUserParams) (TgCachedUser, error) {
	row := q.db.QueryRow(ctx, tgGetCachedUser, arg.SessionID, arg.ID)
	var i TgCachedUser
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Username,
		&i.Phone,
		&i.Raw,
		&i.RawFull,
		&i.FkSessionID,
	)
	return i, err
}

const tgListCachedUsers = `-- name: TgListCachedUsers :many
SELECT
    id,
    first_name,
    last_name,
    username,
    phone,
    raw,
    raw_full,
    fk_session_id
FROM
    tg_cached_users
WHERE
    fk_session_id = $1
ORDER BY id
LIMIT $3 OFFSET $2
`

type TgListCachedUsersParams struct {
	SessionID int64 `json:"session_id"`
	Offset    int32 `json:"offset"`
	Limit     int32 `json:"limit"`
}

func (q *Queries) TgListCachedUsers(ctx context.Context, arg TgListCachedUsersParams) ([]TgCachedUser, error) {
	rows, err := q.db.Query(ctx, tgListCachedUsers, arg.SessionID, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TgCachedUser
	for rows.Next() {
		var i TgCachedUser
		if err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Username,
			&i.Phone,
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
