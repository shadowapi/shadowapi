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

func (h *Handler) DatasourceTelegramCreate(ctx context.Context, req *api.DatasourceTelegram) (api.DatasourceTelegramCreateRes, error) {
	log := h.log.With("handler", "DatasourceTelegramCreate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Get user UUID from authenticated session
	userUUIDStr, err := getUserUUIDFromContext(ctx)
	if err != nil {
		log.Error("failed to get user from context", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("authentication required"))
	}
	pgUserUUID, err := converter.ConvertStringToPgUUID(userUUIDStr)
	if err != nil {
		log.Error("failed to convert user uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
	}

	dsUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}
	isEnabled := req.IsEnabled.Or(false)
	ds, err := query.New(h.dbp).CreateDatasource(ctx, query.CreateDatasourceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
		UserUUID:      pgUserUUID,
		Name:          req.Name,
		IsEnabled:     isEnabled,
		Provider:      string(req.Provider),
		Settings:      settings,
		Type:          "telegram",
	})
	if err != nil {
		log.Error("failed to create datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create datasource"))
	}
	resp := *req
	resp.UUID = api.NewOptString(ds.UUID.String())
	resp.UserUUID = api.NewOptString(userUUIDStr)
	return &resp, nil
}

func (h *Handler) DatasourceTelegramDelete(ctx context.Context, params api.DatasourceTelegramDeleteParams) (api.DatasourceTelegramDeleteRes, error) {
	log := h.log.With("handler", "DatasourceTelegramDelete")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	if err := query.New(h.dbp).DeleteDatasourceByWorkspace(ctx, query.DeleteDatasourceByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
	}); err != nil {
		log.Error("failed to delete datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete datasource"))
	}
	return &api.DatasourceTelegramDeleteOK{}, nil
}

func (h *Handler) DatasourceTelegramGet(ctx context.Context, params api.DatasourceTelegramGetParams) (api.DatasourceTelegramGetRes, error) {
	log := h.log.With("handler", "DatasourceTelegramGet")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	dse, err := query.New(h.dbp).GetDatasourceByWorkspace(ctx, query.GetDatasourceByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: [16]byte(dsUUID.Bytes()), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
	}
	return QToDatasourceTelegramRowByWorkspace(dse)
}

func (h *Handler) DatasourceTelegramUpdate(ctx context.Context, req *api.DatasourceTelegram, params api.DatasourceTelegramUpdateParams) (api.DatasourceTelegramUpdateRes, error) {
	log := h.log.With("handler", "DatasourceTelegramUpdate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.DatasourceTelegramUpdateRes, error) {
		dse, err := query.New(tx).GetDatasourceByWorkspace(ctx, query.GetDatasourceByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: [16]byte(dsUUID.Bytes()), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
		}
		isEnabled := req.IsEnabled.Or(dse.Datasource.IsEnabled)
		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}
		// Preserve the existing user_uuid from the database record
		if err := query.New(tx).UpdateDatasourceByWorkspace(ctx, query.UpdateDatasourceByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
			UserUUID:      converter.UuidPtrToPgUUID(dse.Datasource.UserUUID),
			Name:          req.Name,
			IsEnabled:     isEnabled,
			Provider:      string(req.Provider),
			Settings:      newSettings,
			Type:          "telegram",
		}); err != nil {
			log.Error("failed to update datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update datasource"))
		}
		return QToDatasourceTelegramRowByWorkspace(dse)
	})
}
