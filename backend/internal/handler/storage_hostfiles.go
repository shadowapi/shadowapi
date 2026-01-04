// storage_hostfiles.go
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

func (h *Handler) StorageHostfilesCreate(ctx context.Context, req *api.StorageHostfiles) (api.StorageHostfilesCreateRes, error) {
	log := h.log.With("handler", "StorageHostfilesCreate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

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
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
		Name:          req.Name,
		Type:          "hostfiles",
		IsEnabled:     isEnabled,
		Settings:      settings,
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

func (h *Handler) StorageHostfilesDelete(ctx context.Context, params api.StorageHostfilesDeleteParams) (api.StorageHostfilesDeleteRes, error) {
	log := h.log.With("handler", "StorageHostfilesDelete")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorageByWorkspace(ctx, query.DeleteStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(hostfilesUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
	}); err != nil {
		log.Error("failed to delete hostfiles storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete hostfiles storage"))
	}

	return &api.StorageHostfilesDeleteOK{}, nil
}

func (h *Handler) StorageHostfilesGet(ctx context.Context, params api.StorageHostfilesGetParams) (api.StorageHostfilesGetRes, error) {
	log := h.log.With("handler", "StorageHostfilesGet")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	storage, err := query.New(h.dbp).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(id), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get hostfiles storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	return QToStorageHostfileByWorkspace(storage)
}

func (h *Handler) StorageHostfilesUpdate(ctx context.Context, req *api.StorageHostfiles, params api.StorageHostfilesUpdateParams) (api.StorageHostfilesUpdateRes, error) {
	log := h.log.With("handler", "StorageHostfilesUpdate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	hostfilesUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid hostfiles UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.StorageHostfilesUpdateRes, error) {
		storage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(hostfilesUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
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

		if err := query.New(tx).UpdateStorageByWorkspace(ctx, query.UpdateStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(hostfilesUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
			Type:          "hostfiles",
			Name:          req.Name,
			IsEnabled:     isEnabled,
			Settings:      updatedSettings,
		}); err != nil {
			log.Error("failed to update hostfiles storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		// Re-fetch and return the updated storage
		updatedStorage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(hostfilesUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get updated hostfiles storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated storage"))
		}
		return QToStorageHostfileByWorkspace(updatedStorage)
	})
}

func QToStorageHostfile(row query.GetStorageRow) (*api.StorageHostfiles, error) {
	return storageToHostfile(row.Storage)
}

func QToStorageHostfileByWorkspace(row query.GetStorageByWorkspaceRow) (*api.StorageHostfiles, error) {
	return storageToHostfile(row.Storage)
}

func storageToHostfile(s query.Storage) (*api.StorageHostfiles, error) {
	var out api.StorageHostfiles
	if err := json.Unmarshal(s.Settings, &out); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal hostfiles settings: %w", err))
	}
	out.UUID = api.NewOptString(s.UUID.String())
	out.Name = s.Name
	out.IsEnabled = api.NewOptBool(s.IsEnabled)
	return &out, nil
}
