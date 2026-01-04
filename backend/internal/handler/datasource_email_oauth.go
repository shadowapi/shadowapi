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

// DatasourceEmailOAuthCreate creates a new OAuth2‑based email datasource.
// POST /datasource/email_oauth
func (h *Handler) DatasourceEmailOAuthCreate(ctx context.Context, req *api.DatasourceEmailOAuth) (api.DatasourceEmailOAuthCreateRes, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthCreate")

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
		Type:          "email_oauth",
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

// DatasourceEmailOAuthDelete deletes an OAuth2‑based email datasource.
// DELETE /datasource/email_oauth/{uuid}
func (h *Handler) DatasourceEmailOAuthDelete(ctx context.Context, params api.DatasourceEmailOAuthDeleteParams) (api.DatasourceEmailOAuthDeleteRes, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthDelete")

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
	return &api.DatasourceEmailOAuthDeleteOK{}, nil
}

// DatasourceEmailOAuthGet retrieves a single OAuth2‑based email datasource.
// GET /datasource/email_oauth/{uuid}
func (h *Handler) DatasourceEmailOAuthGet(ctx context.Context, params api.DatasourceEmailOAuthGetParams) (api.DatasourceEmailOAuthGetRes, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthGet")

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
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
	}
	return QToDatasourceEmailOAuthRowByWorkspace(dse)
}

// DatasourceEmailOAuthList lists all OAuth2‑based email datasources.
// GET /datasource/email_oauth
func (h *Handler) DatasourceEmailOAuthList(ctx context.Context, params api.DatasourceEmailOAuthListParams) (api.DatasourceEmailOAuthListRes, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthList")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	offset := params.Offset.Or(0)
	limit := params.Limit.Or(0)
	rows, err := query.New(h.dbp).ListDatasourcesByWorkspace(ctx, query.ListDatasourcesByWorkspaceParams{
		WorkspaceUUID: workspaceUUID,
		Offset:        offset,
		Limit:         limit,
	})
	if err != nil {
		log.Error("failed to get datasources", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list email_oauth datasources"))
	}
	var result []api.DatasourceEmailOAuth
	for _, row := range rows {
		// Filter for email_oauth type only
		if row.Datasource.Type != "email_oauth" {
			continue
		}
		out, err := datasourceToEmailOAuth(row.Datasource)
		if err != nil {
			log.Error("failed to parse datasource row", "error", err)
			continue
		}
		result = append(result, *out)
	}
	res := api.DatasourceEmailOAuthListOKApplicationJSON(result)
	return &res, nil
}

// DatasourceEmailOAuthUpdate updates an existing OAuth2‑based email datasource.
// PUT /datasource/email_oauth/{uuid}
func (h *Handler) DatasourceEmailOAuthUpdate(ctx context.Context, req *api.DatasourceEmailOAuth, params api.DatasourceEmailOAuthUpdateParams) (api.DatasourceEmailOAuthUpdateRes, error) {
	log := h.log.With("handler", "DatasourceEmailOAuthUpdate")

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
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.DatasourceEmailOAuthUpdateRes, error) {
		dse, err := query.New(tx).GetDatasourceByWorkspace(ctx, query.GetDatasourceByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get datasource"))
		}
		isEnabled := req.IsEnabled.Or(dse.Datasource.IsEnabled)

		// Parse existing settings to preserve oauth2_token_uuid (set by OAuth callback)
		var existingSettings struct {
			OAuth2TokenUUID string `json:"oauth2_token_uuid,omitempty"`
		}
		_ = json.Unmarshal(dse.Datasource.Settings, &existingSettings)

		// Marshal request to get base settings
		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}

		// If oauth2_token_uuid exists in existing settings, preserve it
		if existingSettings.OAuth2TokenUUID != "" {
			var settingsMap map[string]any
			if err := json.Unmarshal(newSettings, &settingsMap); err == nil {
				settingsMap["oauth2_token_uuid"] = existingSettings.OAuth2TokenUUID
				newSettings, _ = json.Marshal(settingsMap)
			}
		}

		// Preserve the existing user_uuid from the database record
		err = query.New(tx).UpdateDatasourceByWorkspace(ctx, query.UpdateDatasourceByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(dsUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
			UserUUID:      converter.UuidPtrToPgUUID(dse.Datasource.UserUUID),
			Name:          req.Name,
			IsEnabled:     isEnabled,
			Provider:      string(req.Provider),
			Settings:      newSettings,
			Type:          "email_oauth",
		})
		if err != nil {
			log.Error("failed to update datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update datasource"))
		}
		return QToDatasourceEmailOAuthRowByWorkspace(dse)
	})
}

// QToDatasourceEmailOAuthRow converts a single GetDatasourceRow into an api.DatasourceEmailOAuth.
func QToDatasourceEmailOAuthRow(dse query.GetDatasourceRow) (*api.DatasourceEmailOAuth, error) {
	return datasourceToEmailOAuth(dse.Datasource)
}

// QToDatasourceEmailOAuthRowByWorkspace converts a workspace-filtered row into an api.DatasourceEmailOAuth.
func QToDatasourceEmailOAuthRowByWorkspace(dse query.GetDatasourceByWorkspaceRow) (*api.DatasourceEmailOAuth, error) {
	return datasourceToEmailOAuth(dse.Datasource)
}

func datasourceToEmailOAuth(ds query.Datasource) (*api.DatasourceEmailOAuth, error) {
	var out api.DatasourceEmailOAuth
	if err := json.Unmarshal(ds.Settings, &out); err != nil {
		return nil, err
	}
	out.UUID = api.NewOptString(ds.UUID.String())
	if ds.UserUUID != nil {
		out.UserUUID = api.NewOptString(ds.UserUUID.String())
	}
	out.Name = ds.Name
	out.Provider = api.DatasourceEmailOAuthProvider(ds.Provider)
	out.IsEnabled = api.NewOptBool(ds.IsEnabled)
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
		out.UserUUID = api.NewOptString(r.UserUUID.String())
	}
	out.Name = r.Name
	out.Provider = api.DatasourceEmailOAuthProvider(r.Provider)
	out.IsEnabled = api.NewOptBool(r.IsEnabled)
	return &out, nil
}
