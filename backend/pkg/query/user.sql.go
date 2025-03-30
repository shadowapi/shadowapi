// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user.sql

package query

import (
	"context"

	"github.com/gofrs/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO "user" (
    uuid,
    email,
    password,
    first_name,
    last_name,
    is_enabled,
    is_admin,
    meta,
    created_at,
    updated_at
) VALUES (
             $1,
             $2,
             $3,
             $4,
             $5,
             $6,
             $7,
             $8,
             NOW(),
             NULL
         ) RETURNING uuid, email, password, first_name, last_name, is_enabled, is_admin, meta, created_at, updated_at
`

type CreateUserParams struct {
	UUID      uuid.UUID `json:"uuid"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsEnabled bool      `json:"is_enabled"`
	IsAdmin   bool      `json:"is_admin"`
	Meta      []byte    `json:"meta"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.UUID,
		arg.Email,
		arg.Password,
		arg.FirstName,
		arg.LastName,
		arg.IsEnabled,
		arg.IsAdmin,
		arg.Meta,
	)
	var i User
	err := row.Scan(
		&i.UUID,
		&i.Email,
		&i.Password,
		&i.FirstName,
		&i.LastName,
		&i.IsEnabled,
		&i.IsAdmin,
		&i.Meta,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM "user"
WHERE uuid = $1
`

func (q *Queries) DeleteUser(ctx context.Context, argUuid uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, argUuid)
	return err
}

const getUser = `-- name: GetUser :one
SELECT uuid, email, password, first_name, last_name, is_enabled, is_admin, meta, created_at, updated_at
FROM "user"
WHERE uuid = $1
LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, argUuid uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUser, argUuid)
	var i User
	err := row.Scan(
		&i.UUID,
		&i.Email,
		&i.Password,
		&i.FirstName,
		&i.LastName,
		&i.IsEnabled,
		&i.IsAdmin,
		&i.Meta,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT uuid, email, password, first_name, last_name, is_enabled, is_admin, meta, created_at, updated_at
FROM "user"
ORDER BY created_at DESC
LIMIT CASE WHEN $2::int = 0 THEN NULL ELSE $2::int END
    OFFSET $1::int
`

type ListUsersParams struct {
	OffsetRecords int32 `json:"offset_records"`
	LimitRecords  int32 `json:"limit_records"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers, arg.OffsetRecords, arg.LimitRecords)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.UUID,
			&i.Email,
			&i.Password,
			&i.FirstName,
			&i.LastName,
			&i.IsEnabled,
			&i.IsAdmin,
			&i.Meta,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const updateUser = `-- name: UpdateUser :exec
UPDATE "user"
SET
    email = COALESCE($1, email),
    password = COALESCE($2, password),
    first_name = COALESCE($3, first_name),
    last_name = COALESCE($4, last_name),
    is_enabled = COALESCE($5, is_enabled),
    is_admin = COALESCE($6, is_admin),
    meta = COALESCE($7, meta),
    updated_at = NOW()
WHERE uuid = $8
`

type UpdateUserParams struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsEnabled bool      `json:"is_enabled"`
	IsAdmin   bool      `json:"is_admin"`
	Meta      []byte    `json:"meta"`
	UUID      uuid.UUID `json:"uuid"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.Exec(ctx, updateUser,
		arg.Email,
		arg.Password,
		arg.FirstName,
		arg.LastName,
		arg.IsEnabled,
		arg.IsAdmin,
		arg.Meta,
		arg.UUID,
	)
	return err
}
