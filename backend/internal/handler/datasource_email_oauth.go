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
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// DatasourceEmailOAuthCreate creates a new OAuth2‑based email datasource.
// POST /datasource/email_oauth
func (h *Handler) DatasourceEmailOAuthCreate(ctx context.Context, req *api.DatasourceEmailOAuth) (*api.DatasourceEmailOAuth, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthCreate")
	dsUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}
	isEnabled := req.IsEnabled.Or(false)
	pgUserUUID, err := converter.ConvertStringToPgUUID(req.UserUUID)
	if err != nil {
		log.Error("failed to convert user uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
	}
	ds, err := query.New(h.dbp).CreateDatasource(ctx, query.CreateDatasourceParams{
		UUID:      pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
		UserUUID:  pgUserUUID,
		Name:      req.Name,
		IsEnabled: isEnabled,
		Provider:  req.Provider,
		Settings:  settings,
		Type:      "email_oauth",
	})
	if err != nil {
		log.Error("failed to create datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create datasource"))
	}
	resp := *req
	resp.UUID = api.NewOptString(ds.UUID.String())
	return &resp, nil
}

// DatasourceEmailOAuthDelete deletes an OAuth2‑based email datasource.
// DELETE /datasource/email_oauth/{uuid}
func (h *Handler) DatasourceEmailOAuthDelete(ctx context.Context, params api.DatasourceEmailOAuthDeleteParams) error {
	log := h.log.With("handler", "DatasourceEmailOAuthDelete")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	if err := query.New(h.dbp).DeleteDatasource(ctx, pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true}); err != nil {
		log.Error("failed to delete datasource", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete datasource"))
	}
	return nil
}

// DatasourceEmailOAuthGet retrieves a single OAuth2‑based email datasource.
// GET /datasource/email_oauth/{uuid}
func (h *Handler) DatasourceEmailOAuthGet(ctx context.Context, params api.DatasourceEmailOAuthGetParams) (*api.DatasourceEmailOAuth, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthGet")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	// Use converter.UToBytes to get the 16-byte array
	dse, err := query.New(h.dbp).GetDatasource(ctx, pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true})
	if err != nil {
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
	}
	return QToDatasourceEmailOAuthRow(dse)
}

// DatasourceEmailOAuthList lists all OAuth2‑based email datasources.
// GET /datasource/email_oauth
func (h *Handler) DatasourceEmailOAuthList(ctx context.Context, params api.DatasourceEmailOAuthListParams) ([]api.DatasourceEmailOAuth, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthList")
	offset := params.Offset.Or(0)
	limit := params.Limit.Or(0)
	qp := query.GetDatasourcesParams{
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Offset:         offset,
		Limit:          limit,
		UUID:           "",            // no filtering by datasource uuid
		UserUUID:       "",            // no filtering by user uuid
		Name:           "",            // no filtering by name
		Type:           "email_oauth", // filter only email_oauth datasources
		Provider:       "",
		IsEnabled:      -1, // no filtering
		SyncAll:        -1, // no filtering
	}
	rows, err := query.New(h.dbp).GetDatasources(ctx, qp)
	if err != nil {
		log.Error("failed to get datasources", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list email_oauth datasources"))
	}
	var result []api.DatasourceEmailOAuth
	for _, row := range rows {
		out, err := QToDatasourceEmailOAuthRowMany(row)
		if err != nil {
			log.Error("failed to parse datasource row", "error", err)
			continue
		}
		result = append(result, *out)
	}
	return result, nil
}

// DatasourceEmailOAuthUpdate updates an existing OAuth2‑based email datasource.
// PUT /datasource/email_oauth/{uuid}
func (h *Handler) DatasourceEmailOAuthUpdate(ctx context.Context, req *api.DatasourceEmailOAuth, params api.DatasourceEmailOAuthUpdateParams) (*api.DatasourceEmailOAuth, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthUpdate")
	dsUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse datasource uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid datasource UUID"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.DatasourceEmailOAuth, error) {
		dse, err := query.New(tx).GetDatasource(ctx, pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true})
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
		pgUserUUID, err := converter.ConvertStringToPgUUID(req.UserUUID)
		if err != nil {
			log.Error("failed to convert user uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid user UUID"))
		}
		err = query.New(tx).UpdateDatasource(ctx, query.UpdateDatasourceParams{
			UUID:      pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
			UserUUID:  pgUserUUID,
			Name:      req.Name,
			IsEnabled: isEnabled,
			Provider:  req.Provider,
			Settings:  newSettings,
			Type:      "email_oauth",
		})
		if err != nil {
			log.Error("failed to update datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update datasource"))
		}
		return h.DatasourceEmailOAuthGet(ctx, api.DatasourceEmailOAuthGetParams{UUID: params.UUID})
	})
}

// QToDatasourceEmailOAuthRow converts a single GetDatasourceRow into an api.DatasourceEmailOAuth.
func QToDatasourceEmailOAuthRow(dse query.GetDatasourceRow) (*api.DatasourceEmailOAuth, error) {
	var out api.DatasourceEmailOAuth
	if err := json.Unmarshal(dse.Datasource.Settings, &out); err != nil {
		return nil, err
	}
	out.UUID = api.NewOptString(dse.Datasource.UUID.String())
	if dse.Datasource.UserUUID != nil {
		out.UserUUID = dse.Datasource.UserUUID.String()
	}
	out.Name = dse.Datasource.Name
	out.Provider = dse.Datasource.Provider
	out.IsEnabled = api.NewOptBool(dse.Datasource.IsEnabled)
	return &out, nil
}

// QToDatasourceEmailOAuthRowMany converts a GetDatasourcesRow into an api.DatasourceEmailOAuth.
func QToDatasourceEmailOAuthRowMany(r query.GetDatasourcesRow) (*api.DatasourceEmailOAuth, error) {
	var out api.DatasourceEmailOAuth
	if err := json.Unmarshal(r.Settings, &out); err != nil {
		return nil, err
	}
	out.UUID = api.NewOptString(r.UUID.String())
	if r.UserUUID != nil {
		out.UserUUID = r.UserUUID.String()
	}
	out.Name = r.Name
	out.Provider = r.Provider
	out.IsEnabled = api.NewOptBool(r.IsEnabled)
	return &out, nil
}
