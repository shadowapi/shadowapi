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

func (h *Handler) DatasourceTelegramCreate(ctx context.Context, req *api.DatasourceTelegram) (*api.DatasourceTelegram, error) {
	log := h.log.With("handler", "DatasourceTelegramCreate")
	dsUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}
	isEnabled := req.IsEnabled.Or(false)
	pgUserUUID, err := ConvertStringToPgUUID(req.UserUUID)
	if err != nil {
		log.Error("failed to convert user uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
	}
	ds, err := query.New(h.dbp).CreateDatasource(ctx, query.CreateDatasourceParams{
		UUID:      pgtype.UUID{Bytes: uToBytes(dsUUID), Valid: true},
		UserUUID:  pgUserUUID,
		Name:      req.Name,
		IsEnabled: isEnabled,
		Provider:  string(req.Provider),
		Settings:  settings,
		Type:      "telegram",
	})
	if err != nil {
		log.Error("failed to create datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create datasource"))
	}
	resp := *req
	resp.UUID = api.NewOptString(ds.UUID.String())
	return &resp, nil
}

func (h *Handler) DatasourceTelegramDelete(ctx context.Context, params api.DatasourceTelegramDeleteParams) error {
	log := h.log.With("handler", "DatasourceTelegramDelete")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	if err := query.New(h.dbp).DeleteDatasource(ctx, pgtype.UUID{Bytes: uToBytes(dsUUID), Valid: true}); err != nil {
		log.Error("failed to delete datasource", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete datasource"))
	}
	return nil
}

func (h *Handler) DatasourceTelegramGet(ctx context.Context, params api.DatasourceTelegramGetParams) (*api.DatasourceTelegram, error) {
	log := h.log.With("handler", "DatasourceTelegramGet")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	dses, err := query.New(h.dbp).GetDatasources(ctx, query.GetDatasourcesParams{
		UUID:   pgtype.UUID{Bytes: [16]byte(dsUUID.Bytes()), Valid: true},
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
	}
	if len(dses) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("datasource not found"))
	}
	return QToDatasourceTelegram(dses[0])
}

func (h *Handler) DatasourceTelegramUpdate(ctx context.Context, req *api.DatasourceTelegram, params api.DatasourceTelegramUpdateParams) (*api.DatasourceTelegram, error) {
	log := h.log.With("handler", "DatasourceTelegramUpdate")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.DatasourceTelegram, error) {
		dses, err := query.New(tx).GetDatasources(ctx, query.GetDatasourcesParams{
			UUID:  pgtype.UUID{Bytes: [16]byte(dsUUID.Bytes()), Valid: true},
			Limit: 1,
		})
		if err != nil {
			log.Error("failed to get datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
		}
		if len(dses) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("datasource not found"))
		}
		isEnabled := req.IsEnabled.Or(dses[0].IsEnabled)
		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}
		pgUserUUID, err := ConvertStringToPgUUID(req.UserUUID)
		if err != nil {
			log.Error("failed to convert user uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
		}

		if err := query.New(tx).UpdateDatasource(ctx, query.UpdateDatasourceParams{
			UUID:      pgtype.UUID{Bytes: uToBytes(dsUUID), Valid: true},
			UserUUID:  pgUserUUID,
			Name:      req.Name,
			IsEnabled: isEnabled,
			Provider:  string(req.Provider),
			Settings:  newSettings,
		}); err != nil {
			log.Error("failed to update datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update datasource"))
		}
		return h.DatasourceTelegramGet(ctx, api.DatasourceTelegramGetParams{UUID: params.UUID})
	})
}
