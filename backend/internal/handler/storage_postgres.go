// storage_postgres.go
package handler

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Extract underlying values from optional fields.
	var isEnabled bool
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	} else {
		isEnabled = false
	}

	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:      storageUUID,
		Name:      req.Name,
		Type:      "postgres",
		IsEnabled: isEnabled,
		Settings:  settings,
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

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	fmt.Println("id", id)

	storages, err := query.New(h.dbp).GetStorage(ctx, [16]byte(id.Bytes()))
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	return QToStoragePostgres(storages)
}

func (h *Handler) StoragePostgresUpdate(ctx context.Context, req *api.StoragePostgres, params api.StoragePostgresUpdateParams) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresUpdate")

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StoragePostgres, error) {
		storage, err := query.New(tx).GetStorage(ctx, [16]byte(storageUUID.Bytes()))
		if err != nil {
			log.Error("failed to get storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}

		// For update, use new values if provided; otherwise fall back to current DB values.
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = storage.IsEnabled
		}

		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}

		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:      storageUUID,
			Type:      "postgres",
			Name:      req.Name,
			IsEnabled: isEnabled,
			Settings:  newSettings,
		}); err != nil {
			log.Error("failed to update storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		return h.StoragePostgresGet(ctx, api.StoragePostgresGetParams{UUID: params.UUID})
	})
}
