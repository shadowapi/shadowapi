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
		ds.UserUUID = row.UserUUID.String()
	} else {
		ds.UserUUID = ""
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

func QToDatasourceEmail2(row query.GetDatasourceRow) (*api.DatasourceEmail, error) {
	var ds api.DatasourceEmail
	// Unmarshal the settings JSON into the DatasourceEmail fields.
	if err := json.Unmarshal(row.Datasource.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.Datasource.UUID.String())
	if row.Datasource.UserUUID != nil {
		ds.UserUUID = row.Datasource.UserUUID.String()
	} else {
		ds.UserUUID = ""
	}
	ds.Name = row.Datasource.Name
	ds.IsEnabled = api.NewOptBool(row.Datasource.IsEnabled)
	ds.Provider = row.Datasource.Provider
	if row.Datasource.CreatedAt.Valid {
		ds.CreatedAt = api.NewOptDateTime(row.Datasource.CreatedAt.Time)
	}
	if row.Datasource.UpdatedAt.Valid {
		ds.UpdatedAt = api.NewOptDateTime(row.Datasource.UpdatedAt.Time)
	}
	return &ds, nil
}

// QToDatasourceLinkedin converts a query datasource row into an API DatasourceLinkedin.
func QToDatasourceLinkedin(row query.GetDatasourcesRow) (*api.DatasourceLinkedin, error) {
	var ds api.DatasourceLinkedin
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	// ds.UserUUID is a plain string
	if row.UserUUID != nil {
		ds.UserUUID = row.UserUUID.String()
	} else {
		ds.UserUUID = ""
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

// QToDatasourceTelegram converts a query datasource row into an API DatasourceTelegram.
func QToDatasourceTelegram(row query.GetDatasourcesRow) (*api.DatasourceTelegram, error) {
	var ds api.DatasourceTelegram
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	if row.UserUUID != nil {
		ds.UserUUID = row.UserUUID.String()
	} else {
		ds.UserUUID = ""
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

// QToDatasourceWhatsapp converts a query datasource row into an API DatasourceWhatsapp.
func QToDatasourceWhatsapp(row query.GetDatasourcesRow) (*api.DatasourceWhatsapp, error) {
	var ds api.DatasourceWhatsapp
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())

	if row.UserUUID != nil {
		ds.UserUUID = row.UserUUID.String()
	} else {
		ds.UserUUID = ""
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
