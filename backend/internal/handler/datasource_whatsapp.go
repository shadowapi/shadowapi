package handler

import (
	"context"
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) DatasourceWhatsappCreate(ctx context.Context, req *api.DatasourceWhatsapp) (api.DatasourceWhatsappCreateRes, error) {
	log := h.log.With("handler", "DatasourceWhatsappCreate")

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
		UUID:      pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
		UserUUID:  pgUserUUID,
		Name:      req.Name,
		IsEnabled: isEnabled,
		Provider:  string(req.Provider),
		Settings:  settings,
		Type:      "whatsapp",
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

func (h *Handler) DatasourceWhatsappDelete(ctx context.Context, params api.DatasourceWhatsappDeleteParams) (api.DatasourceWhatsappDeleteRes, error) {
	log := h.log.With("handler", "DatasourceWhatsappDelete")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	if err := query.New(h.dbp).DeleteDatasource(ctx, pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true}); err != nil {
		log.Error("failed to delete datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete datasource"))
	}
	return &api.DatasourceWhatsappDeleteOK{}, nil
}

func (h *Handler) DatasourceWhatsappGet(ctx context.Context, params api.DatasourceWhatsappGetParams) (api.DatasourceWhatsappGetRes, error) {
	log := h.log.With("handler", "DatasourceWhatsappGet")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	dse, err := query.New(h.dbp).GetDatasource(ctx, pgtype.UUID{Bytes: [16]byte(dsUUID.Bytes()), Valid: true})
	if err != nil {
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
	}
	return QToDatasourceWhatsappRow(dse)
}

func (h *Handler) DatasourceWhatsappUpdate(ctx context.Context, req *api.DatasourceWhatsapp, params api.DatasourceWhatsappUpdateParams) (api.DatasourceWhatsappUpdateRes, error) {
	log := h.log.With("handler", "DatasourceWhatsappUpdate")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.DatasourceWhatsappUpdateRes, error) {
		dse, err := query.New(tx).GetDatasource(ctx, pgtype.UUID{Bytes: [16]byte(dsUUID.Bytes()), Valid: true})
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
		if err := query.New(tx).UpdateDatasource(ctx, query.UpdateDatasourceParams{
			UUID:      pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
			UserUUID:  converter.UuidPtrToPgUUID(dse.Datasource.UserUUID),
			Name:      req.Name,
			IsEnabled: isEnabled,
			Provider:  string(req.Provider),
			Settings:  newSettings,
			Type:      "whatsapp",
		}); err != nil {
			log.Error("failed to update datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update datasource"))
		}
		return QToDatasourceWhatsappRow(dse)
	})
}
