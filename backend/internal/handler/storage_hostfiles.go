// storage_hostfiles.go
package handler

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
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
		UUID:      pgtype.UUID{Bytes: uToBytes(storageUUID), Valid: true},
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

	if err := query.New(h.dbp).DeleteStorage(ctx, pgtype.UUID{Bytes: uToBytes(hostfilesUUID), Valid: true}); err != nil {
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

	storage, err := query.New(h.dbp).GetStorage(ctx, pgtype.UUID{Bytes: uToBytes(id), Valid: true})
	if err != nil {
		log.Error("failed to get hostfiles storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	return QToStorageHostfile(storage)
}

func (h *Handler) StorageHostfilesUpdate(ctx context.Context, req *api.StorageHostfiles, params api.StorageHostfilesUpdateParams) (*api.StorageHostfiles, error) {
	log := h.log.With("handler", "StorageHostfilesUpdate")

	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StorageHostfiles, error) {
		storage, err := query.New(tx).GetStorage(ctx, pgtype.UUID{Bytes: uToBytes(hostfilesUUID), Valid: true})
		if err != nil {
			log.Error("failed to get hostfiles storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}

		// For update, use new values if provided; otherwise fallback to current DB value.
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = storage.Storage.IsEnabled
		}

		updatedSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal updated hostfiles settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal hostfiles updated settings"))
		}

		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:      pgtype.UUID{Bytes: uToBytes(hostfilesUUID), Valid: true},
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

func QToStorageHostfile(row query.GetStorageRow) (*api.StorageHostfiles, error) {
	var stored api.StorageHostfiles
	if err := json.Unmarshal(row.Storage.Settings, &stored); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal hostfiles settings", err.Error()))
	}
	stored.UUID = api.NewOptString(row.Storage.UUID.String())
	stored.Name = row.Storage.Name
	stored.IsEnabled = api.NewOptBool(row.Storage.IsEnabled)
	return &stored, nil
}
