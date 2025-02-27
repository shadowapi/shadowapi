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

func (h *Handler) StorageS3Create(ctx context.Context, req *api.StorageS3) (*api.StorageS3, error) {
	storageUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}
	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:     storageUUID,
		Type:     "s3",
		Settings: settings,
	})
	if err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create storage"))
	}
	resp := *req
	resp.UUID = api.NewOptString(storage.UUID.String())
	return &resp, nil
}

func (h *Handler) StorageS3Delete(ctx context.Context, params api.StorageS3DeleteParams) error {
	s3UUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	if err := query.New(h.dbp).DeleteStorage(ctx, s3UUID); err != nil {
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete storage"))
	}
	return nil
}

func (h *Handler) StorageS3Get(ctx context.Context, params api.StorageS3GetParams) (*api.StorageS3, error) {
	log := h.log.With("handler", "StorageS3Get")
	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse s3 uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	rows, err := query.New(h.dbp).GetStorages(ctx, query.GetStoragesParams{
		UUID:  pgtype.UUID{Bytes: [16]byte(id.Bytes())},
		Limit: 1,
	})
	if err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}
	if len(rows) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
	}
	return QToStorageS3(rows[0])
}

func (h *Handler) StorageS3Update(ctx context.Context, req *api.StorageS3, params api.StorageS3UpdateParams) (*api.StorageS3, error) {
	s3UUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StorageS3, error) {
		rows, err := query.New(tx).GetStorages(ctx, query.GetStoragesParams{
			UUID:  pgtype.UUID{Bytes: [16]byte(s3UUID.Bytes())},
			Limit: 1,
		})
		if err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}
		if len(rows) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
		}
		updatedSettings, err := json.Marshal(req)
		if err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}
		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:     s3UUID,
			Type:     "s3",
			Name:     rows[0].Name,
			Settings: updatedSettings,
		}); err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}
		return h.StorageS3Get(ctx, api.StorageS3GetParams{UUID: params.UUID})
	})
}
