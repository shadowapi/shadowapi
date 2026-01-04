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

// QToDatasourceLinkedin converts a query datasource row into an API DatasourceLinkedin.
func QToDatasourceLinkedin(row query.GetDatasourcesRow) (*api.DatasourceLinkedin, error) {
	var ds api.DatasourceLinkedin
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	if row.UserUUID != nil {
		ds.UserUUID = api.NewOptString(row.UserUUID.String())
	}

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

func QToDatasourceLinkedinRow(row query.GetDatasourceRow) (*api.DatasourceLinkedin, error) {
	return datasourceToLinkedin(row.Datasource)
}

func QToDatasourceLinkedinRowByWorkspace(row query.GetDatasourceByWorkspaceRow) (*api.DatasourceLinkedin, error) {
	return datasourceToLinkedin(row.Datasource)
}

func datasourceToLinkedin(ds query.Datasource) (*api.DatasourceLinkedin, error) {
	var out api.DatasourceLinkedin
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

// QToDatasourceTelegram converts a query datasource row into an API DatasourceTelegram.
func QToDatasourceTelegram(row query.GetDatasourcesRow) (*api.DatasourceTelegram, error) {
	var ds api.DatasourceTelegram
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	if row.UserUUID != nil {
		ds.UserUUID = api.NewOptString(row.UserUUID.String())
	}

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
func QToDatasourceTelegramRow(row query.GetDatasourceRow) (*api.DatasourceTelegram, error) {
	return datasourceToTelegram(row.Datasource)
}

func QToDatasourceTelegramRowByWorkspace(row query.GetDatasourceByWorkspaceRow) (*api.DatasourceTelegram, error) {
	return datasourceToTelegram(row.Datasource)
}

func datasourceToTelegram(ds query.Datasource) (*api.DatasourceTelegram, error) {
	var out api.DatasourceTelegram
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

// QToDatasourceWhatsapp converts a query datasource row into an API DatasourceWhatsapp.
func QToDatasourceWhatsapp(row query.GetDatasourcesRow) (*api.DatasourceWhatsapp, error) {
	var ds api.DatasourceWhatsapp
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	if row.UserUUID != nil {
		ds.UserUUID = api.NewOptString(row.UserUUID.String())
	}

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

func QToDatasourceWhatsappRow(row query.GetDatasourceRow) (*api.DatasourceWhatsapp, error) {
	return datasourceToWhatsapp(row.Datasource)
}

func QToDatasourceWhatsappRowByWorkspace(row query.GetDatasourceByWorkspaceRow) (*api.DatasourceWhatsapp, error) {
	return datasourceToWhatsapp(row.Datasource)
}

func datasourceToWhatsapp(ds query.Datasource) (*api.DatasourceWhatsapp, error) {
	var out api.DatasourceWhatsapp
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
