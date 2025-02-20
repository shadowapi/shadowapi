package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StoragePostgresCreate(ctx context.Context, req *api.StoragePostgres) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresCreate")

	storageUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}

	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:     storageUUID,
		Type:     "postgres",
		Settings: settings,
	})
	if err != nil {
		log.Error("failed to create storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create storage"))
	}

	resp := *req
	resp.UUID = api.OptString{Value: storage.UUID.String(), Set: true}

	return &resp, nil
}

func (h *Handler) StoragePostgresDelete(ctx context.Context, params api.StoragePostgresDeleteParams) error {
	log := h.log.With("handler", "StoragePostgresDelete")

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorage(ctx, storageUUID); err != nil {
		log.Error("failed to delete storage", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete storage"))
	}

	return nil
}

func (h *Handler) StoragePostgresGet(ctx context.Context, params api.StoragePostgresGetParams) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresGet")

	storages, err := query.New(h.dbp).GetStorages(ctx, query.GetStoragesParams{
		UUID:  params.UUID,
		Limit: 1,
	})
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	if len(storages) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
	}

	return QToStoragePostgres(storages[0])
}

func (h *Handler) StoragePostgresUpdate(
	ctx context.Context,
	req *api.StoragePostgres,
	params api.StoragePostgresUpdateParams,
) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresUpdate")

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StoragePostgres, error) {
		storages, err := query.New(tx).GetStorages(ctx, query.GetStoragesParams{
			UUID:  storageUUID.String(),
			Limit: 1,
		})
		if err != nil {
			log.Error("failed to get storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}

		update := storages[0]
		update.Name = req.Name
		update.Settings, err = json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}

		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:     storageUUID,
			Type:     "postgres",
			Name:     update.Name,
			Settings: update.Settings,
		}); err != nil {
			log.Error("failed to update storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		return h.StoragePostgresGet(ctx, api.StoragePostgresGetParams{UUID: params.UUID})
	})
}
