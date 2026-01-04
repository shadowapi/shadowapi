// storage_s3.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StorageS3Create(ctx context.Context, req *api.StorageS3) (api.StorageS3CreateRes, error) {
	log := h.log.With("handler", "StorageS3Create")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

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
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
		Name:          req.Name,
		Type:          "s3",
		IsEnabled:     isEnabled,
		Settings:      settings,
	})
	if err != nil {
		log.Error("failed to create s3 storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create s3 storage"))
	}

	resp := *req
	resp.UUID = api.NewOptString(storage.UUID.String())
	return &resp, nil
}

func (h *Handler) StorageS3Delete(ctx context.Context, params api.StorageS3DeleteParams) (api.StorageS3DeleteRes, error) {
	log := h.log.With("handler", "StorageS3Delete")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	s3UUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid s3 UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid s3 storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorageByWorkspace(ctx, query.DeleteStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(s3UUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
	}); err != nil {
		log.Error("failed to delete s3 storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete s3 storage"))
	}
	return &api.StorageS3DeleteOK{}, nil
}

func (h *Handler) StorageS3Get(ctx context.Context, params api.StorageS3GetParams) (api.StorageS3GetRes, error) {
	log := h.log.With("handler", "StorageS3Get")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid s3 UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid s3 storage UUID"))
	}

	storage, err := query.New(h.dbp).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(id), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get s3 storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get s3 storage"))
	}
	return QToStorageS3ByWorkspace(storage)
}

func (h *Handler) StorageS3Update(ctx context.Context, req *api.StorageS3, params api.StorageS3UpdateParams) (api.StorageS3UpdateRes, error) {
	log := h.log.With("handler", "StorageS3Update")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	s3UUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid s3 UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid s3 storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.StorageS3UpdateRes, error) {
		storage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(s3UUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get s3 storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get s3 storage"))
		}
		// For update, use new values if provided; fallback to existing DB values.
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = storage.Storage.IsEnabled
		}

		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal s3 updated settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal s3 updated settings"))
		}

		if err := query.New(tx).UpdateStorageByWorkspace(ctx, query.UpdateStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(s3UUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
			Type:          "s3",
			Name:          req.Name,
			IsEnabled:     isEnabled,
			Settings:      newSettings,
		}); err != nil {
			log.Error("failed to update s3 storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update s3 storage"))
		}

		// Re-fetch and return the updated storage
		updatedStorage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(s3UUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get updated s3 storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated storage"))
		}
		return QToStorageS3ByWorkspace(updatedStorage)
	})
}

func QToStorageS3(row query.GetStorageRow) (*api.StorageS3, error) {
	return storageToS3(row.Storage)
}

func QToStorageS3ByWorkspace(row query.GetStorageByWorkspaceRow) (*api.StorageS3, error) {
	return storageToS3(row.Storage)
}

func storageToS3(s query.Storage) (*api.StorageS3, error) {
	var out api.StorageS3
	if err := json.Unmarshal(s.Settings, &out); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal s3 settings: %w", err))
	}
	out.UUID = api.NewOptString(s.UUID.String())
	out.Name = s.Name
	out.IsEnabled = api.NewOptBool(s.IsEnabled)
	return &out, nil
}
