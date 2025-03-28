// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
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
    user_id,
    service,
    blocklist,
    exclude_list,
    sync_all,
    settings,
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
             NOW(),
             NOW()
         ) RETURNING uuid, user_id, service, blocklist, exclude_list, sync_all, settings, created_at, updated_at
`

type CreateSyncPolicyParams struct {
	UUID        uuid.UUID   `json:"uuid"`
	UserID      pgtype.UUID `json:"user_id"`
	Service     string      `json:"service"`
	Blocklist   []string    `json:"blocklist"`
	ExcludeList []string    `json:"exclude_list"`
	SyncAll     bool        `json:"sync_all"`
	Settings    []byte      `json:"settings"`
}

func (q *Queries) CreateSyncPolicy(ctx context.Context, arg CreateSyncPolicyParams) (SyncPolicy, error) {
	row := q.db.QueryRow(ctx, createSyncPolicy,
		arg.UUID,
		arg.UserID,
		arg.Service,
		arg.Blocklist,
		arg.ExcludeList,
		arg.SyncAll,
		arg.Settings,
	)
	var i SyncPolicy
	err := row.Scan(
		&i.UUID,
		&i.UserID,
		&i.Service,
		&i.Blocklist,
		&i.ExcludeList,
		&i.SyncAll,
		&i.Settings,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteSyncPolicy = `-- name: DeleteSyncPolicy :exec
DELETE FROM sync_policy WHERE uuid = $1
`

func (q *Queries) DeleteSyncPolicy(ctx context.Context, argUuid uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteSyncPolicy, argUuid)
	return err
}

const getSyncPolicies = `-- name: GetSyncPolicies :many
WITH filtered_sync_policies AS (
    SELECT sp.uuid, sp.user_id, sp.service, sp.blocklist, sp.exclude_list, sp.sync_all, sp.settings, sp.created_at, sp.updated_at FROM sync_policy sp WHERE
        ($5::text IS NULL OR sp.service = $5) AND
        ($6::uuid IS NULL OR sp.uuid = $6) AND
        ($7::uuid IS NULL OR sp.user_id = $7) AND
        ($8::bool IS NULL OR sp.sync_all = $8)
)
SELECT
    uuid, user_id, service, blocklist, exclude_list, sync_all, settings, created_at, updated_at,
    (SELECT count(*) FROM filtered_sync_policies) as total_count
FROM filtered_sync_policies
ORDER BY
    CASE WHEN $1 = 'created_at' AND $2::text = 'asc' THEN created_at END ASC,
    CASE WHEN $1 = 'created_at' AND $2::text = 'desc' THEN created_at END DESC,
    CASE WHEN $1 = 'updated_at' AND $2::text = 'asc' THEN updated_at END ASC,
    CASE WHEN $1 = 'updated_at' AND $2::text = 'desc' THEN updated_at END DESC,
    CASE WHEN $1 = 'service' AND $2::text = 'asc' THEN service END ASC,
    CASE WHEN $1 = 'service' AND $2::text = 'desc' THEN service END DESC,
    created_at DESC
LIMIT NULLIF($4::int, 0) OFFSET $3
`

type GetSyncPoliciesParams struct {
	OrderBy        interface{} `json:"order_by"`
	OrderDirection string      `json:"order_direction"`
	Offset         int32       `json:"offset"`
	Limit          int32       `json:"limit"`
	Service        string      `json:"service"`
	UUID           pgtype.UUID `json:"uuid"`
	UserID         pgtype.UUID `json:"user_id"`
	SyncAll        bool        `json:"sync_all"`
}

type GetSyncPoliciesRow struct {
	UUID        uuid.UUID          `json:"uuid"`
	UserID      pgtype.UUID        `json:"user_id"`
	Service     string             `json:"service"`
	Blocklist   []string           `json:"blocklist"`
	ExcludeList []string           `json:"exclude_list"`
	SyncAll     bool               `json:"sync_all"`
	Settings    []byte             `json:"settings"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
	TotalCount  int64              `json:"total_count"`
}

func (q *Queries) GetSyncPolicies(ctx context.Context, arg GetSyncPoliciesParams) ([]GetSyncPoliciesRow, error) {
	rows, err := q.db.Query(ctx, getSyncPolicies,
		arg.OrderBy,
		arg.OrderDirection,
		arg.Offset,
		arg.Limit,
		arg.Service,
		arg.UUID,
		arg.UserID,
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
			&i.UserID,
			&i.Service,
			&i.Blocklist,
			&i.ExcludeList,
			&i.SyncAll,
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
    sync_policy.uuid, sync_policy.user_id, sync_policy.service, sync_policy.blocklist, sync_policy.exclude_list, sync_policy.sync_all, sync_policy.settings, sync_policy.created_at, sync_policy.updated_at
FROM sync_policy
ORDER BY created_at DESC
LIMIT CASE WHEN $2::int = 0 THEN NULL ELSE $2::int END
    OFFSET $1::int
`

type ListSyncPoliciesParams struct {
	OffsetRecords int32 `json:"offset_records"`
	LimitRecords  int32 `json:"limit_records"`
}

type ListSyncPoliciesRow struct {
	SyncPolicy SyncPolicy `json:"sync_policy"`
}

func (q *Queries) ListSyncPolicies(ctx context.Context, arg ListSyncPoliciesParams) ([]ListSyncPoliciesRow, error) {
	rows, err := q.db.Query(ctx, listSyncPolicies, arg.OffsetRecords, arg.LimitRecords)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListSyncPoliciesRow
	for rows.Next() {
		var i ListSyncPoliciesRow
		if err := rows.Scan(
			&i.SyncPolicy.UUID,
			&i.SyncPolicy.UserID,
			&i.SyncPolicy.Service,
			&i.SyncPolicy.Blocklist,
			&i.SyncPolicy.ExcludeList,
			&i.SyncPolicy.SyncAll,
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
                       user_id = $1,
                       service = $2,
                       blocklist = $3,
                       exclude_list = $4,
                       sync_all = $5,
                       settings = $6,
                       updated_at = NOW()
WHERE uuid = $7
`

type UpdateSyncPolicyParams struct {
	UserID      pgtype.UUID `json:"user_id"`
	Service     string      `json:"service"`
	Blocklist   []string    `json:"blocklist"`
	ExcludeList []string    `json:"exclude_list"`
	SyncAll     bool        `json:"sync_all"`
	Settings    []byte      `json:"settings"`
	UUID        uuid.UUID   `json:"uuid"`
}

func (q *Queries) UpdateSyncPolicy(ctx context.Context, arg UpdateSyncPolicyParams) error {
	_, err := q.db.Exec(ctx, updateSyncPolicy,
		arg.UserID,
		arg.Service,
		arg.Blocklist,
		arg.ExcludeList,
		arg.SyncAll,
		arg.Settings,
		arg.UUID,
	)
	return err
}
