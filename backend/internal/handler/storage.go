package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// StorageList handles the "GET /storages" endpoint.
func (h *Handler) StorageList(
	ctx context.Context, params api.StorageListParams,
) ([]api.Storage, error) {
	log := h.log.With("handler", "StorageList")

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.Storage, error) {
		// Start with zero-values for each field:
		arg := query.GetStoragesParams{}

		// If user passes ?limit=, set it; else leave at 0
		if params.Limit.IsSet() {
			arg.Limit = params.Limit.Value
		}
		// If user passes ?offset=, set it; else leave at 0
		if params.Offset.IsSet() {
			arg.Offset = params.Offset.Value
		}
		// If user passes ?type=, set it; else remains ""
		if params.Type.IsSet() {
			arg.Type = params.Type.Value
		}
		// If user passes ?is_enabled=, set it; else remains false
		if params.IsEnabled.IsSet() {
			arg.IsEnabled = params.IsEnabled.Value
		}
		// If user passes ?name=, set it; else remains ""
		if params.Name.IsSet() {
			arg.Name = params.Name.Value
		}
		// If user passes ?order_by=, set it; else remains nil interface{}
		if params.OrderBy.IsSet() {
			// Convert custom type to plain string:
			arg.OrderBy = string(params.OrderBy.Value)
		}
		// If user passes ?order_direction=, set it; else remains ""
		if params.OrderDirection.IsSet() {
			arg.OrderDirection = string(params.OrderDirection.Value)
		}

		rows, err := query.New(h.dbp).GetStorages(ctx, arg)
		if err != nil {
			log.Error("failed to list storage", "error", err.Error())
			return nil, ErrWithCode(
				http.StatusInternalServerError,
				E("failed to list storage"),
			)
		}

		var storages []api.Storage
		for _, row := range rows {
			storages = append(storages, QToStorage(row))
		}
		return storages, nil
	})
}
