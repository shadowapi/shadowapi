package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/db"

	// "github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StorageList(
	ctx context.Context, params api.StorageListParams,
) ([]api.Storage, error) {
	log := h.log.With("handler", "StorageList")

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.Storage, error) {
		arg := query.GetStoragesParams{}
		if params.Limit.IsSet() {
			arg.Limit = params.Limit.Value
		}
		if params.Offset.IsSet() {
			arg.Offset = params.Offset.Value
		}

		rows, err := query.New(h.dbp).GetStorages(ctx, arg)
		if err != nil {
			log.Error("failed to list storage", "error", err.Error())
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list storage"))
		}

		var storages []api.Storage
		for _, row := range rows {
			storages = append(storages, QToStorage(row))
		}
		return storages, nil
	})
}
