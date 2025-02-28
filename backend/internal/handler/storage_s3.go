// storage_s3.go
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

func (h *Handler) StorageS3Create(ctx context.Context, req *api.StorageS3) (*api.StorageS3, error) {
	log := h.log.With("handler", "StorageS3Create")

	storageUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal s3 settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal s3 settings"))
	}

	var isEnabled bool
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	} else {
		isEnabled = false
	}

	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:      storageUUID,
		Name:      req.Name,
		Type:      "s3",
		IsEnabled: isEnabled,
		Settings:  settings,
	})
	if err != nil {
		log.Error("failed to create s3 storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create s3 storage"))
	}

	resp := *req
	resp.UUID = api.NewOptString(storage.UUID.String())
	return &resp, nil
}

func (h *Handler) StorageS3Delete(ctx context.Context, params api.StorageS3DeleteParams) error {
	log := h.log.With("handler", "StorageS3Delete")

	s3UUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid s3 UUID", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid s3 storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorage(ctx, s3UUID); err != nil {
		log.Error("failed to delete s3 storage", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete s3 storage"))
	}
	return nil
}

func (h *Handler) StorageS3Get(ctx context.Context, params api.StorageS3GetParams) (*api.StorageS3, error) {
	log := h.log.With("handler", "StorageS3Get")

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid s3 UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid s3 storage UUID"))
	}

	storages, err := query.New(h.dbp).GetStorages(ctx, query.GetStoragesParams{
		UUID:  pgtype.UUID{Bytes: [16]byte(id.Bytes())},
		Limit: 1,
	})
	if err != nil {
		log.Error("failed to get s3 storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get s3 storage"))
	}
	if len(storages) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("s3 storage not found"))
	}

	return QToStorageS3(storages[0])
}

func (h *Handler) StorageS3Update(ctx context.Context, req *api.StorageS3, params api.StorageS3UpdateParams) (*api.StorageS3, error) {
	log := h.log.With("handler", "StorageS3Update")

	s3UUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid s3 UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid s3 storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StorageS3, error) {
		storages, err := query.New(tx).GetStorages(ctx, query.GetStoragesParams{
			UUID:  pgtype.UUID{Bytes: [16]byte(s3UUID.Bytes())},
			Limit: 1,
		})
		if err != nil {
			log.Error("failed to get s3 storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get s3 storage"))
		}
		if len(storages) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("s3 storage not found"))
		}

		// For update, use new values if provided; fallback to existing DB values.
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = storages[0].IsEnabled
		}

		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal s3 updated settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal s3 updated settings"))
		}

		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:      s3UUID,
			Type:      "s3",
			Name:      req.Name,
			IsEnabled: isEnabled,
			Settings:  newSettings,
		}); err != nil {
			log.Error("failed to update s3 storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update s3 storage"))
		}

		return h.StorageS3Get(ctx, api.StorageS3GetParams{UUID: params.UUID})
	})
}
