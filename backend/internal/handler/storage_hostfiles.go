// storage_hostfiles.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StorageHostfilesCreate(ctx context.Context, req *api.StorageHostfiles) (*api.StorageHostfiles, error) {
	log := h.log.With("handler", "StorageHostfilesCreate")

	storageUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal hostfiles settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal hostfiles settings"))
	}

	// Extract underlying values from the optional types.
	var isEnabled bool
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	} else {
		isEnabled = false
	}

	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:      storageUUID,
		Name:      req.Name,
		Type:      "hostfiles",
		IsEnabled: isEnabled,
		Settings:  settings,
	})
	if err != nil {
		log.Error("failed to create hostfiles storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create hostfiles storage"))
	}

	// Return the same object the caller sent, but attach the newly generated UUID.
	resp := *req
	resp.UUID = api.NewOptString(storage.UUID.String())

	return &resp, nil
}

func (h *Handler) StorageHostfilesDelete(ctx context.Context, params api.StorageHostfilesDeleteParams) error {
	log := h.log.With("handler", "StorageHostfilesDelete")

	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorage(ctx, hostfilesUUID); err != nil {
		log.Error("failed to delete hostfiles storage", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete hostfiles storage"))
	}

	return nil
}

func (h *Handler) StorageHostfilesGet(ctx context.Context, params api.StorageHostfilesGetParams) (*api.StorageHostfiles, error) {
	log := h.log.With("handler", "StorageHostfilesGet")

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	storages, err := query.New(h.dbp).GetStorages(ctx, query.GetStoragesParams{
		UUID:  pgtype.UUID{Bytes: [16]byte(id.Bytes())},
		Limit: 1,
	})
	if err != nil {
		log.Error("failed to get hostfiles storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}
	if len(storages) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
	}

	return QToStorageHostfiles(storages[0])
}

func (h *Handler) StorageHostfilesUpdate(ctx context.Context, req *api.StorageHostfiles, params api.StorageHostfilesUpdateParams) (*api.StorageHostfiles, error) {
	log := h.log.With("handler", "StorageHostfilesUpdate")

	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StorageHostfiles, error) {
		storages, err := query.New(tx).GetStorages(ctx, query.GetStoragesParams{
			UUID:  pgtype.UUID{Bytes: [16]byte(hostfilesUUID.Bytes())},
			Limit: 1,
		})
		if err != nil {
			log.Error("failed to get hostfiles storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}
		if len(storages) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
		}

		// For update, use new values if provided; otherwise fallback to current DB value.
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = storages[0].IsEnabled
		}

		updatedSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal updated hostfiles settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal hostfiles updated settings"))
		}

		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:      hostfilesUUID,
			Type:      "hostfiles",
			Name:      req.Name,
			IsEnabled: isEnabled,
			Settings:  updatedSettings,
		}); err != nil {
			log.Error("failed to update hostfiles storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		return h.StorageHostfilesGet(ctx, api.StorageHostfilesGetParams{UUID: params.UUID})
	})
}
