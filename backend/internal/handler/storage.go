package handler

import (
	"context"
	"fmt"
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
			Limit:          10000,
			Type:           "",
			UUID:           pgtype.UUID{}, // Will be treated as NULL if not set.
			IsEnabled:      -1,            // -1 = unset, 0 = false, 1 = true
			Name:           "",
		}
		if params.Limit.IsSet() {
			fmt.Println("Limit is set")
			arg.Limit = params.Limit.Value
		}
		if params.Offset.IsSet() {
			fmt.Println("Offset is set")
			arg.Offset = params.Offset.Value
		}
		if params.Type.IsSet() {
			fmt.Println("Type is set")
			arg.Type = params.Type.Value
		}
		if params.IsEnabled.IsSet() {
			fmt.Println("IsEnabled is set")
			arg.IsEnabled = 1
		}
		if params.Name.IsSet() {
			fmt.Println("Name is set")
			// Wrap with wildcards for ILIKE filtering
			arg.Name = "%" + params.Name.Value + "%"
		}
		if params.OrderBy.IsSet() {
			fmt.Println("OrderBy is set")
			arg.OrderBy = params.OrderBy.Value
		}
		if params.OrderDirection.IsSet() {
			fmt.Println("OrderDirection is set")
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
