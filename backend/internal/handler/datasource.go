package handler

import (
	"context"
	"github.com/jackc/pgx/v5"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) DatasourceList(
	ctx context.Context,
	params api.DatasourceListParams, // Has Offset, Limit only
) ([]api.Datasource, error) {
	// Because ogen only defines offset/limit in DatasourceListParams,
	// we can unify them into a single function.
	log := h.log.With("handler", "DatasourceList")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.Datasource, error) {
		var offset, limit int32
		if params.Offset.IsSet() {
			offset = params.Offset.Value
		}
		if params.Limit.IsSet() {
			limit = params.Limit.Value
		}

		// Call our one "ListDatasources" query with offset & limit
		listArgs := query.ListDatasourcesParams{
			Offset: offset,
			Limit:  limit,
		}
		rows, err := query.New(h.dbp).ListDatasources(ctx, listArgs)
		if err != nil {
			log.Error("failed to list datasources", "error", err.Error())
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list datasources"))
		}

		// Convert each row from the query to our API representation
		var datasources []api.Datasource
		for _, row := range rows {
			datasources = append(datasources, QToDatasource(row.Datasource))
		}
		return datasources, nil
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
		c.UserUUID = ds.UserUUID.String()
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
