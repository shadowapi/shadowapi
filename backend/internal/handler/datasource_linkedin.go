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

func (h *Handler) DatasourceLinkedinCreate(ctx context.Context, req *api.DatasourceLinkedin) (*api.DatasourceLinkedin, error) {
	log := h.log.With("handler", "DatasourceLinkedinCreate")
	dsUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}
	isEnabled := req.IsEnabled.Or(false)
	ds, err := query.New(h.dbp).CreateDatasource(ctx, query.CreateDatasourceParams{
		UUID:      dsUUID,
		UserUUID:  ConvertUUID(req.UserUUID),
		Name:      req.Name,
		IsEnabled: isEnabled,
		Provider:  string(req.Provider),
		Settings:  settings,
	})
	if err != nil {
		log.Error("failed to create datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create datasource"))
	}
	resp := *req
	resp.UUID = api.NewOptString(ds.UUID.String())
	return &resp, nil
}

func (h *Handler) DatasourceLinkedinDelete(ctx context.Context, params api.DatasourceLinkedinDeleteParams) error {
	log := h.log.With("handler", "DatasourceLinkedinDelete")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	if err := query.New(h.dbp).DeleteDatasource(ctx, dsUUID); err != nil {
		log.Error("failed to delete datasource", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete datasource"))
	}
	return nil
}

func (h *Handler) DatasourceLinkedinGet(ctx context.Context, params api.DatasourceLinkedinGetParams) (*api.DatasourceLinkedin, error) {
	log := h.log.With("handler", "DatasourceLinkedinGet")
	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	dses, err := query.New(h.dbp).GetDatasources(ctx, query.GetDatasourcesParams{
		UUID:   pgtype.UUID{Bytes: [16]byte(id.Bytes()), Valid: true},
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
	return QToDatasourceLinkedin(dses[0])
}

func (h *Handler) DatasourceLinkedinUpdate(ctx context.Context, req *api.DatasourceLinkedin, params api.DatasourceLinkedinUpdateParams) (*api.DatasourceLinkedin, error) {
	log := h.log.With("handler", "DatasourceLinkedinUpdate")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.DatasourceLinkedin, error) {
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
		if err := query.New(tx).UpdateDatasource(ctx, query.UpdateDatasourceParams{
			UUID:      dsUUID,
			UserUUID:  ConvertUUID(req.UserUUID),
			Name:      req.Name,
			IsEnabled: isEnabled,
			Provider:  string(req.Provider),
			Settings:  newSettings,
		}); err != nil {
			log.Error("failed to update datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update datasource"))
		}
		return h.DatasourceLinkedinGet(ctx, api.DatasourceLinkedinGetParams{UUID: params.UUID})
	})
}
