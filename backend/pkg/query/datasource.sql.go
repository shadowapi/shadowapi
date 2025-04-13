// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: datasource.sql

package query

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createDatasource = `-- name: CreateDatasource :one
INSERT INTO datasource (
    uuid,
    user_uuid,
    name,
    "type",
    is_enabled,
    provider,
    settings,
    created_at,
    updated_at
) VALUES (
             $1::uuid,
             $2::uuid,
             NULLIF($3, ''),
             NULLIF($4, ''),
             $5::boolean,
             $6,
             $7,
             NOW(),
             NOW()
         ) RETURNING uuid, user_uuid, name, type, is_enabled, provider, settings, created_at, updated_at
`

type CreateDatasourceParams struct {
	UUID      pgtype.UUID `json:"uuid"`
	UserUUID  pgtype.UUID `json:"user_uuid"`
	Name      interface{} `json:"name"`
	Type      interface{} `json:"type"`
	IsEnabled bool        `json:"is_enabled"`
	Provider  string      `json:"provider"`
	Settings  []byte      `json:"settings"`
}

func (q *Queries) CreateDatasource(ctx context.Context, arg CreateDatasourceParams) (Datasource, error) {
	row := q.db.QueryRow(ctx, createDatasource,
		arg.UUID,
		arg.UserUUID,
		arg.Name,
		arg.Type,
		arg.IsEnabled,
		arg.Provider,
		arg.Settings,
	)
	var i Datasource
	err := row.Scan(
		&i.UUID,
		&i.UserUUID,
		&i.Name,
		&i.Type,
		&i.IsEnabled,
		&i.Provider,
		&i.Settings,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteDatasource = `-- name: DeleteDatasource :exec
DELETE FROM datasource WHERE uuid = $1::uuid
`

func (q *Queries) DeleteDatasource(ctx context.Context, argUuid pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteDatasource, argUuid)
	return err
}

const getDatasource = `-- name: GetDatasource :one
SELECT
    datasource.uuid, datasource.user_uuid, datasource.name, datasource.type, datasource.is_enabled, datasource.provider, datasource.settings, datasource.created_at, datasource.updated_at
FROM datasource
WHERE uuid = $1::uuid
`

type GetDatasourceRow struct {
	Datasource Datasource `json:"datasource"`
}

func (q *Queries) GetDatasource(ctx context.Context, argUuid pgtype.UUID) (GetDatasourceRow, error) {
	row := q.db.QueryRow(ctx, getDatasource, argUuid)
	var i GetDatasourceRow
	err := row.Scan(
		&i.Datasource.UUID,
		&i.Datasource.UserUUID,
		&i.Datasource.Name,
		&i.Datasource.Type,
		&i.Datasource.IsEnabled,
		&i.Datasource.Provider,
		&i.Datasource.Settings,
		&i.Datasource.CreatedAt,
		&i.Datasource.UpdatedAt,
	)
	return i, err
}

const getDatasources = `-- name: GetDatasources :many
WITH filtered_datasource AS (
    SELECT d.uuid, d.user_uuid, d.name, d.type, d.is_enabled, d.provider, d.settings, d.created_at, d.updated_at
    FROM datasource d
    WHERE
        (NULLIF($5, '') IS NULL OR sp.uuid = $5::uuid) AND
        (NULLIF($6, '') IS NULL OR sp.uuid = $6::uuid) AND
        (NULLIF($7, '') IS NULL OR sp."type" = $7) AND
        (NULLIF($8, '') IS NULL OR sp."type" = $8) AND
        (NULLIF($9, '') IS NULL OR sp."type" = $9) AND
        (NULLIF($10::int, -1) IS NULL OR sp.sync_all = ($11::int)::boolean)
)
SELECT
    uuid, user_uuid, name, type, is_enabled, provider, settings, created_at, updated_at,
    (SELECT count(*) FROM filtered_datasource) AS total_count
FROM filtered_datasource
ORDER BY
    CASE WHEN $1 = 'created_at' AND $2::text = 'asc' THEN created_at END ASC,
    CASE WHEN $1 = 'created_at' AND $2::text = 'desc' THEN created_at END DESC,
    CASE WHEN $1 = 'updated_at' AND $2::text = 'asc' THEN updated_at END ASC,
    CASE WHEN $1 = 'updated_at' AND $2::text = 'desc' THEN updated_at END DESC,
    CASE WHEN $1 = 'name' AND $2::text = 'asc' THEN name END ASC,
    CASE WHEN $1 = 'name' AND $2::text = 'desc' THEN name END DESC,
    created_at DESC
LIMIT NULLIF($4::int, 0)
    OFFSET $3
`

type GetDatasourcesParams struct {
	OrderBy        interface{} `json:"order_by"`
	OrderDirection string      `json:"order_direction"`
	Offset         int32       `json:"offset"`
	Limit          int32       `json:"limit"`
	UUID           interface{} `json:"uuid"`
	UserUUID       interface{} `json:"user_uuid"`
	Name           interface{} `json:"name"`
	Type           interface{} `json:"type"`
	Provider       interface{} `json:"provider"`
	IsEnabled      int32       `json:"is_enabled"`
	SyncAll        int32       `json:"sync_all"`
}

type GetDatasourcesRow struct {
	UUID       uuid.UUID          `json:"uuid"`
	UserUUID   *uuid.UUID         `json:"user_uuid"`
	Name       string             `json:"name"`
	Type       string             `json:"type"`
	IsEnabled  bool               `json:"is_enabled"`
	Provider   string             `json:"provider"`
	Settings   []byte             `json:"settings"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
	TotalCount int64              `json:"total_count"`
}

func (q *Queries) GetDatasources(ctx context.Context, arg GetDatasourcesParams) ([]GetDatasourcesRow, error) {
	rows, err := q.db.Query(ctx, getDatasources,
		arg.OrderBy,
		arg.OrderDirection,
		arg.Offset,
		arg.Limit,
		arg.UUID,
		arg.UserUUID,
		arg.Name,
		arg.Type,
		arg.Provider,
		arg.IsEnabled,
		arg.SyncAll,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDatasourcesRow
	for rows.Next() {
		var i GetDatasourcesRow
		if err := rows.Scan(
			&i.UUID,
			&i.UserUUID,
			&i.Name,
			&i.Type,
			&i.IsEnabled,
			&i.Provider,
			&i.Settings,
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

const listDatasources = `-- name: ListDatasources :many
SELECT
    datasource.uuid, datasource.user_uuid, datasource.name, datasource.type, datasource.is_enabled, datasource.provider, datasource.settings, datasource.created_at, datasource.updated_at
FROM datasource
ORDER BY created_at DESC
LIMIT NULLIF($2::int, 0)
    OFFSET $1
`

type ListDatasourcesParams struct {
	Offset int32 `json:"offset"`
	Limit  int32 `json:"limit"`
}

type ListDatasourcesRow struct {
	Datasource Datasource `json:"datasource"`
}

func (q *Queries) ListDatasources(ctx context.Context, arg ListDatasourcesParams) ([]ListDatasourcesRow, error) {
	rows, err := q.db.Query(ctx, listDatasources, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListDatasourcesRow
	for rows.Next() {
		var i ListDatasourcesRow
		if err := rows.Scan(
			&i.Datasource.UUID,
			&i.Datasource.UserUUID,
			&i.Datasource.Name,
			&i.Datasource.Type,
			&i.Datasource.IsEnabled,
			&i.Datasource.Provider,
			&i.Datasource.Settings,
			&i.Datasource.CreatedAt,
			&i.Datasource.UpdatedAt,
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

const updateDatasource = `-- name: UpdateDatasource :exec
UPDATE datasource
SET
    user_uuid  = $1::uuid,
    "type"     = NULLIF($2, ''),
    name       =  NULLIF($3, ''),
    is_enabled = $4::boolean,
    provider   = $5,
    settings   = $6,
    updated_at = NOW()
WHERE uuid = $7::uuid
`

type UpdateDatasourceParams struct {
	UserUUID  pgtype.UUID `json:"user_uuid"`
	Type      interface{} `json:"type"`
	Name      interface{} `json:"name"`
	IsEnabled bool        `json:"is_enabled"`
	Provider  string      `json:"provider"`
	Settings  []byte      `json:"settings"`
	UUID      pgtype.UUID `json:"uuid"`
}

func (q *Queries) UpdateDatasource(ctx context.Context, arg UpdateDatasourceParams) error {
	_, err := q.db.Exec(ctx, updateDatasource,
		arg.UserUUID,
		arg.Type,
		arg.Name,
		arg.IsEnabled,
		arg.Provider,
		arg.Settings,
		arg.UUID,
	)
	return err
}
