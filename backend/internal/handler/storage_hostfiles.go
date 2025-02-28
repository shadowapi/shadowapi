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
	// We store the entire request in the "settings" JSON blob.
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal hostfiles settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal hostfiles settings"))
	}

	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:     storageUUID,
		Type:     "hostfiles",
		Settings: settings,
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

		// The row from the DB; we'll only update .Name field if we want to store it in the column.
		update := storages[0]
		// If you want to reflect "req.Name" in the top-level "name" column, do something like:
		updateName := update.Name
		if req.Name.IsSet() {
			updateName = req.Name.Or(update.Name)
		}

		// Store entire request in settings JSON
		updatedSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal updated hostfiles settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal hostfiles updated settings"))
		}

		err = query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			Type:     "hostfiles",
			Name:     updateName,
			Settings: updatedSettings,
			UUID:     hostfilesUUID,
		})
		if err != nil {
			log.Error("failed to update hostfiles storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		return h.StorageHostfilesGet(ctx, api.StorageHostfilesGetParams{UUID: params.UUID})
	})
}
