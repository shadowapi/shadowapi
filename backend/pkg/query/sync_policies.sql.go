// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: sync_policies.sql

package query

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createSyncPolicy = `-- name: CreateSyncPolicy :one
INSERT INTO sync_policy (
    uuid,
    pipeline_uuid,
    name,
    "type",
    blocklist,
    exclude_list,
    sync_all,
    settings,
    created_at,
    updated_at
) VALUES (
    $1::uuid,
    $2::uuid,
    $3,
    $4,
    $5,
    $6,
    $7::boolean,
    $8,
             NOW(),
             NOW()
         ) RETURNING uuid, pipeline_uuid, name, type, blocklist, exclude_list, sync_all, is_enabled, settings, created_at, updated_at
`

type CreateSyncPolicyParams struct {
	UUID         pgtype.UUID `json:"uuid"`
	PipelineUuid pgtype.UUID `json:"pipeline_uuid"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Blocklist    []string    `json:"blocklist"`
	ExcludeList  []string    `json:"exclude_list"`
	SyncAll      bool        `json:"sync_all"`
	Settings     []byte      `json:"settings"`
}

func (q *Queries) CreateSyncPolicy(ctx context.Context, arg CreateSyncPolicyParams) (SyncPolicy, error) {
	row := q.db.QueryRow(ctx, createSyncPolicy,
		arg.UUID,
		arg.PipelineUuid,
		arg.Name,
		arg.Type,
		arg.Blocklist,
		arg.ExcludeList,
		arg.SyncAll,
		arg.Settings,
	)
	var i SyncPolicy
	err := row.Scan(
		&i.UUID,
		&i.PipelineUuid,
		&i.Name,
		&i.Type,
		&i.Blocklist,
		&i.ExcludeList,
		&i.SyncAll,
		&i.IsEnabled,
		&i.Settings,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteSyncPolicy = `-- name: DeleteSyncPolicy :exec
DELETE FROM sync_policy WHERE uuid = $1::uuid
`

func (q *Queries) DeleteSyncPolicy(ctx context.Context, argUuid pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteSyncPolicy, argUuid)
	return err
}

const getPolicy = `-- name: GetPolicy :one
SELECT
    sync_policy.uuid, sync_policy.pipeline_uuid, sync_policy.name, sync_policy.type, sync_policy.blocklist, sync_policy.exclude_list, sync_policy.sync_all, sync_policy.is_enabled, sync_policy.settings, sync_policy.created_at, sync_policy.updated_at
FROM sync_policy
WHERE uuid = $1::uuid
`

type GetPolicyRow struct {
	SyncPolicy SyncPolicy `json:"sync_policy"`
}

func (q *Queries) GetPolicy(ctx context.Context, argUuid pgtype.UUID) (GetPolicyRow, error) {
	row := q.db.QueryRow(ctx, getPolicy, argUuid)
	var i GetPolicyRow
	err := row.Scan(
		&i.SyncPolicy.UUID,
		&i.SyncPolicy.PipelineUuid,
		&i.SyncPolicy.Name,
		&i.SyncPolicy.Type,
		&i.SyncPolicy.Blocklist,
		&i.SyncPolicy.ExcludeList,
		&i.SyncPolicy.SyncAll,
		&i.SyncPolicy.IsEnabled,
		&i.SyncPolicy.Settings,
		&i.SyncPolicy.CreatedAt,
		&i.SyncPolicy.UpdatedAt,
	)
	return i, err
}

const getSyncPolicies = `-- name: GetSyncPolicies :many
WITH filtered_sync_policies AS (
    SELECT sp.uuid, sp.pipeline_uuid, sp.name, sp.type, sp.blocklist, sp.exclude_list, sp.sync_all, sp.is_enabled, sp.settings, sp.created_at, sp.updated_at
    FROM sync_policy sp
    WHERE
        (NULLIF($5, '') IS NULL OR sp."type" = $5) AND
        (NULLIF($6, '') IS NULL OR sp.uuid = $6::uuid) AND
        (NULLIF($7::int, -1) IS NULL OR sp.sync_all = ($7::int)::boolean)
)
SELECT
    uuid, pipeline_uuid, name, type, blocklist, exclude_list, sync_all, is_enabled, settings, created_at, updated_at,
    (SELECT count(*) FROM filtered_sync_policies) as total_count
FROM filtered_sync_policies
ORDER BY
    CASE WHEN $1 = 'created_at' AND $2 = 'asc' THEN created_at END ASC,
    CASE WHEN $1 = 'created_at' AND $2 = 'desc' THEN created_at END DESC,
    CASE WHEN $1 = 'updated_at' AND $2 = 'asc' THEN updated_at END ASC,
    CASE WHEN $1 = 'updated_at' AND $2 = 'desc' THEN updated_at END DESC,
    created_at DESC
LIMIT NULLIF($4::int, 0)
OFFSET $3::int
`

type GetSyncPoliciesParams struct {
	OrderBy        interface{} `json:"order_by"`
	OrderDirection interface{} `json:"order_direction"`
	Offset         int32       `json:"offset"`
	Limit          int32       `json:"limit"`
	Type           interface{} `json:"type"`
	UUID           interface{} `json:"uuid"`
	SyncAll        int32       `json:"sync_all"`
}

type GetSyncPoliciesRow struct {
	UUID         uuid.UUID          `json:"uuid"`
	PipelineUuid *uuid.UUID         `json:"pipeline_uuid"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Blocklist    []string           `json:"blocklist"`
	ExcludeList  []string           `json:"exclude_list"`
	SyncAll      bool               `json:"sync_all"`
	IsEnabled    bool               `json:"is_enabled"`
	Settings     []byte             `json:"settings"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
	TotalCount   int64              `json:"total_count"`
}

func (q *Queries) GetSyncPolicies(ctx context.Context, arg GetSyncPoliciesParams) ([]GetSyncPoliciesRow, error) {
	rows, err := q.db.Query(ctx, getSyncPolicies,
		arg.OrderBy,
		arg.OrderDirection,
		arg.Offset,
		arg.Limit,
		arg.Type,
		arg.UUID,
		arg.SyncAll,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSyncPoliciesRow
	for rows.Next() {
		var i GetSyncPoliciesRow
		if err := rows.Scan(
			&i.UUID,
			&i.PipelineUuid,
			&i.Name,
			&i.Type,
			&i.Blocklist,
			&i.ExcludeList,
			&i.SyncAll,
			&i.IsEnabled,
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

const listSyncPolicies = `-- name: ListSyncPolicies :many
SELECT
    sync_policy.uuid, sync_policy.pipeline_uuid, sync_policy.name, sync_policy.type, sync_policy.blocklist, sync_policy.exclude_list, sync_policy.sync_all, sync_policy.is_enabled, sync_policy.settings, sync_policy.created_at, sync_policy.updated_at
FROM sync_policy
ORDER BY created_at DESC
LIMIT NULLIF($2::int, 0)
    OFFSET $1
`

type ListSyncPoliciesParams struct {
	Offset int32 `json:"offset"`
	Limit  int32 `json:"limit"`
}

type ListSyncPoliciesRow struct {
	SyncPolicy SyncPolicy `json:"sync_policy"`
}

func (q *Queries) ListSyncPolicies(ctx context.Context, arg ListSyncPoliciesParams) ([]ListSyncPoliciesRow, error) {
	rows, err := q.db.Query(ctx, listSyncPolicies, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListSyncPoliciesRow
	for rows.Next() {
		var i ListSyncPoliciesRow
		if err := rows.Scan(
			&i.SyncPolicy.UUID,
			&i.SyncPolicy.PipelineUuid,
			&i.SyncPolicy.Name,
			&i.SyncPolicy.Type,
			&i.SyncPolicy.Blocklist,
			&i.SyncPolicy.ExcludeList,
			&i.SyncPolicy.SyncAll,
			&i.SyncPolicy.IsEnabled,
			&i.SyncPolicy.Settings,
			&i.SyncPolicy.CreatedAt,
			&i.SyncPolicy.UpdatedAt,
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

const updateSyncPolicy = `-- name: UpdateSyncPolicy :exec
UPDATE sync_policy SET
    name = $1,
    blocklist = $2,
    exclude_list = $3,
    sync_all = $4::boolean,
    settings = $5,
    updated_at = NOW()
WHERE uuid = $6::uuid
`

type UpdateSyncPolicyParams struct {
	Name        string      `json:"name"`
	Blocklist   []string    `json:"blocklist"`
	ExcludeList []string    `json:"exclude_list"`
	SyncAll     bool        `json:"sync_all"`
	Settings    []byte      `json:"settings"`
	UUID        pgtype.UUID `json:"uuid"`
}

func (q *Queries) UpdateSyncPolicy(ctx context.Context, arg UpdateSyncPolicyParams) error {
	_, err := q.db.Exec(ctx, updateSyncPolicy,
		arg.Name,
		arg.Blocklist,
		arg.ExcludeList,
		arg.SyncAll,
		arg.Settings,
		arg.UUID,
	)
	return err
}
