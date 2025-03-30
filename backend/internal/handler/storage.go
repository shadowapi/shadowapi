package handler

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StorageList(ctx context.Context, params api.StorageListParams) ([]api.Storage, error) {
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.Storage, error) {
		// Build GetStoragesParams with default values.
		arg := query.GetStoragesParams{
			OrderBy:        "created_at",
			OrderDirection: "desc",
			Offset:         0,
			Limit:          0,
			// Use empty string for text filters which the SQL now converts to NULL using NULLIF.
			Type:      "",
			UUID:      pgtype.UUID{}, // Will be treated as NULL if not set.
			IsEnabled: false,         // Must be provided explicitly to filter.
			Name:      "",
		}
		if params.Limit.IsSet() {
			arg.Limit = params.Limit.Value
		}
		if params.Offset.IsSet() {
			arg.Offset = params.Offset.Value
		}
		if params.Type.IsSet() {
			arg.Type = params.Type.Value
		}
		if params.IsEnabled.IsSet() {
			arg.IsEnabled = params.IsEnabled.Value
		}
		if params.Name.IsSet() {
			// Wrap with wildcards for ILIKE filtering
			arg.Name = "%" + params.Name.Value + "%"
		}
		if params.OrderBy.IsSet() {
			arg.OrderBy = params.OrderBy.Value
		}
		if params.OrderDirection.IsSet() {
			arg.OrderDirection = params.OrderDirection.Value
		}

		rows, err := query.New(h.dbp).GetStorages(ctx, arg)
		if err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list storage"))
		}

		var storages []api.Storage
		for _, row := range rows {
			storages = append(storages, QToStorage(row))
		}
		return storages, nil
	})
}

func QToStorage(row query.GetStoragesRow) api.Storage {
	return api.Storage{
		UUID:      row.UUID.String(),
		Name:      api.NewOptString(row.Name),
		Type:      row.Type,
		IsEnabled: row.IsEnabled,
		CreatedAt: api.NewOptDateTime(row.CreatedAt.Time),
		UpdatedAt: api.NewOptDateTime(row.UpdatedAt.Time),
	}
}
