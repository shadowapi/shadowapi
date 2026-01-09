package handler

import (
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// QToDatasourceEmail converts a query datasource row into an API DatasourceEmail.
func QToDatasourceEmail(row query.GetDatasourcesRow) (*api.DatasourceEmail, error) {
	var ds api.DatasourceEmail
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	if row.UserUUID != nil {
		ds.UserUUID = api.NewOptString(row.UserUUID.String())
	}

	// ds.Name is a plain string:
	ds.Name = row.Name
	ds.IsEnabled = api.NewOptBool(row.IsEnabled)
	ds.Provider = row.Provider
	if row.CreatedAt.Valid {
		ds.CreatedAt = api.NewOptDateTime(row.CreatedAt.Time)
	}
	if row.UpdatedAt.Valid {
		ds.UpdatedAt = api.NewOptDateTime(row.UpdatedAt.Time)
	}
	return &ds, nil
}

func QToDatasourceEmailRow(row query.GetDatasourceRow) (*api.DatasourceEmail, error) {
	return datasourceToEmail(row.Datasource)
}

func QToDatasourceEmailRowByWorkspace(row query.GetDatasourceByWorkspaceRow) (*api.DatasourceEmail, error) {
	return datasourceToEmail(row.Datasource)
}

func datasourceToEmail(ds query.Datasource) (*api.DatasourceEmail, error) {
	var out api.DatasourceEmail
	if err := json.Unmarshal(ds.Settings, &out); err != nil {
		return nil, err
	}
	out.UUID = api.NewOptString(ds.UUID.String())
	if ds.UserUUID != nil {
		out.UserUUID = api.NewOptString(ds.UserUUID.String())
	}
	out.Name = ds.Name
	out.IsEnabled = api.NewOptBool(ds.IsEnabled)
	out.Provider = ds.Provider
	if ds.CreatedAt.Valid {
		out.CreatedAt = api.NewOptDateTime(ds.CreatedAt.Time)
	}
	if ds.UpdatedAt.Valid {
		out.UpdatedAt = api.NewOptDateTime(ds.UpdatedAt.Time)
	}
	return &out, nil
}

func QToStorage(row query.GetStoragesRow) api.Storage {
	return api.Storage{
		UUID:      row.UUID.String(),
		Name:      api.NewOptString(row.Name),
		Type:      row.Type,
		IsEnabled: row.IsEnabled,
		CreatedAt: api.NewOptDateTime(row.CreatedAt.Time),
		UpdatedAt: api.NewOptDateTime(row.UpdatedAt.Time),
	}
}
