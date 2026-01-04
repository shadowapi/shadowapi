package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) DatasourceList(
	ctx context.Context,
	params api.DatasourceListParams, // Has Offset, Limit only
) (api.DatasourceListRes, error) {
	log := h.log.With("handler", "DatasourceList")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.DatasourceListRes, error) {
		var offset, limit int32
		if params.Offset.IsSet() {
			offset = params.Offset.Value
		}
		if params.Limit.IsSet() {
			limit = params.Limit.Value
		}

		// Call the workspace-filtered query that includes OAuth authentication status
		listArgs := query.ListDatasourcesWithOAuthStatusByWorkspaceParams{
			WorkspaceUUID: workspaceUUID,
			Offset:        offset,
			Limit:         limit,
		}
		rows, err := query.New(h.dbp).ListDatasourcesWithOAuthStatusByWorkspace(ctx, listArgs)
		if err != nil {
			log.Error("failed to list datasources", "error", err.Error())
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list datasources"))
		}

		// Convert each row from the query to our API representation
		var datasources []api.Datasource
		for _, row := range rows {
			ds := QToDatasource(row.Datasource)
			ds.IsOAuthAuthenticated = api.NewOptBool(row.IsOauthAuthenticated)
			datasources = append(datasources, ds)
		}
		res := api.DatasourceListOKApplicationJSON(datasources)
		return &res, nil
	})
}

// QToDatasource maps a single query.Datasource to api.Datasource.
// Notice it does NOT reference any 'Status' or 'IsEnabled' param from your endpoint –
// it just reads the actual fields in query.Datasource.
func QToDatasource(ds query.Datasource) api.Datasource {
	c := api.Datasource{
		UUID:      api.NewOptString(ds.UUID.String()),
		Name:      ds.Name,
		Type:      ds.Type,
		IsEnabled: api.NewOptBool(ds.IsEnabled),
	}
	if ds.UserUUID != nil {
		c.UserUUID = api.NewOptString(ds.UserUUID.String())
	}
	if ds.CreatedAt.Valid {
		c.CreatedAt = api.NewOptDateTime(ds.CreatedAt.Time)
	}
	if ds.UpdatedAt.Valid {
		c.UpdatedAt = api.NewOptDateTime(ds.UpdatedAt.Time)
	}
	// Feel free to parse ds.Provider or ds.Settings if needed
	return c
}
