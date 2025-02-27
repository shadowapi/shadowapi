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
	storageUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}
	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:     storageUUID,
		Type:     "hostfiles",
		Settings: settings,
	})
	if err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create storage"))
	}
	resp := *req
	resp.UUID = api.NewOptString(storage.UUID.String())
	return &resp, nil
}

func (h *Handler) StorageHostfilesDelete(ctx context.Context, params api.StorageHostfilesDeleteParams) error {
	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	if err := query.New(h.dbp).DeleteStorage(ctx, hostfilesUUID); err != nil {
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete storage"))
	}
	return nil
}

func (h *Handler) StorageHostfilesGet(ctx context.Context, params api.StorageHostfilesGetParams) (*api.StorageHostfiles, error) {
	log := h.log.With("handler", "StorageHostfilesGet")
	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse hostfile uuid", "error", err)
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
	return QToStorageHostfiles(rows[0])
}

func (h *Handler) StorageHostfilesUpdate(ctx context.Context, req *api.StorageHostfiles, params api.StorageHostfilesUpdateParams) (*api.StorageHostfiles, error) {
	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StorageHostfiles, error) {
		rows, err := query.New(tx).GetStorages(ctx, query.GetStoragesParams{
			UUID:  pgtype.UUID{Bytes: [16]byte(hostfilesUUID.Bytes())},
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
			UUID:     hostfilesUUID,
			Type:     "hostfiles",
			Name:     rows[0].Name,
			Settings: updatedSettings,
		}); err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}
		return h.StorageHostfilesGet(ctx, api.StorageHostfilesGetParams{UUID: params.UUID})
	})
}
