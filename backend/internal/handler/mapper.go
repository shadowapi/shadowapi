package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"net/http"
)

// ConvertUUID converts a non-empty  UUID string to a pointer to uuid.UUID.
// If the input is empty or invalid, it returns nil.
func ConvertUUID(originalUUID string) *uuid.UUID {
	if originalUUID == "" {
		return nil
	}
	u, err := uuid.FromString(originalUUID)
	if err != nil {
		return nil
	}
	return &u
}

// ConvertOptStringToUUID converts an api.OptString to a uuid.UUID.
// It returns an error if the OptString is not set or if the inner value is invalid.
func ConvertOptStringToUUID(opt api.OptString) (uuid.UUID, error) {
	if !opt.IsSet() || opt.Value == "" {
		return uuid.Nil, fmt.Errorf("opt string is not set")
	}
	return uuid.FromString(opt.Value)
}

// QToDatasource converts a query.Datasource to an api.Datasource
func QToDatasource(row query.GetDatasourcesRow) api.Datasource {
	c := api.Datasource{
		UUID:      row.UUID.String(),
		Name:      row.Name,
		Type:      row.Type,
		IsEnabled: row.IsEnabled,
	}
	if row.UserUUID != nil {
		c.UserUUID = row.UserUUID.String()
	}
	if row.CreatedAt.Valid {
		c.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		c.UpdatedAt = row.UpdatedAt.Time
	}
	return c
}

// QToDatasourceEmail converts a query datasource row into an API DatasourceEmail.
func QToDatasourceEmail(row query.GetDatasourcesRow) (*api.DatasourceEmail, error) {
	var ds api.DatasourceEmail
	if err := json.Unmarshal(row.Settings, &ds); err != nil {
		return nil, err
	}
	ds.UUID = api.NewOptString(row.UUID.String())
	ds.UserUUID = row.UserUUID.String()
	ds.Name = row.Name
	ds.IsEnabled = api.NewOptBool(row.IsEnabled)
	ds.Provider = row.Provider
	if row.CreatedAt.Valid {
		ds.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		ds.UpdatedAt = row.UpdatedAt.Time
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
	// (If your settings include an OAuth2 token UUID, it should have been unmarshaled into ds.OAuth2TokenUUID.)
	if row.Datasource.CreatedAt.Valid {
		ds.CreatedAt = row.Datasource.CreatedAt.Time
	}
	if row.Datasource.UpdatedAt.Valid {
		ds.UpdatedAt = row.Datasource.UpdatedAt.Time
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
	ds.UserUUID = row.UserUUID.String()
	ds.Name = row.Name
	ds.IsEnabled = api.NewOptBool(row.IsEnabled)
	ds.Provider = row.Provider
	if row.CreatedAt.Valid {
		ds.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		ds.UpdatedAt = row.UpdatedAt.Time
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
	ds.UserUUID = row.UserUUID.String()
	ds.Name = row.Name
	ds.IsEnabled = api.NewOptBool(row.IsEnabled)
	ds.Provider = row.Provider
	if row.CreatedAt.Valid {
		ds.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		ds.UpdatedAt = row.UpdatedAt.Time
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
	ds.UserUUID = row.UserUUID.String()
	ds.Name = row.Name
	ds.IsEnabled = api.NewOptBool(row.IsEnabled)
	ds.Provider = row.Provider
	if row.CreatedAt.Valid {
		ds.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		ds.UpdatedAt = row.UpdatedAt.Time
	}
	return &ds, nil
}

func QToStorage(row query.GetStoragesRow) api.Storage {
	r := api.Storage{
		UUID:      row.UUID.String(),
		Name:      api.NewOptString(row.Name),
		Type:      row.Type,
		IsEnabled: row.IsEnabled,
	}
	if row.CreatedAt.Valid {
		r.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		r.UpdatedAt = row.UpdatedAt.Time
	}
	return r
}

func QToStoragePostgres(row query.GetStoragesRow) (*api.StoragePostgres, error) {
	// The JSON in row.Settings has the entire Postgres object
	var s api.StoragePostgres
	if err := json.Unmarshal(row.Settings, &s); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal postgres settings", err.Error()))
	}
	s.UUID = api.NewOptString(row.UUID.String())
	s.Name = row.Name
	s.IsEnabled = api.NewOptBool(row.IsEnabled)
	return &s, nil
}

func QToStorageS3(row query.GetStoragesRow) (*api.StorageS3, error) {
	// The JSON in row.Settings has the entire S3 object
	var stored api.StorageS3
	if err := json.Unmarshal(row.Settings, &stored); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal s3 settings", err.Error()))
	}

	stored.UUID = api.NewOptString(row.UUID.String())
	stored.Name = row.Name
	stored.IsEnabled = api.NewOptBool(row.IsEnabled)

	return &stored, nil
}

func QToStorageHostfiles(row query.GetStoragesRow) (*api.StorageHostfiles, error) {
	// The JSON stored in row.Settings has the entire original api.StorageHostfiles object.
	var stored api.StorageHostfiles
	if err := json.Unmarshal(row.Settings, &stored); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal hostfiles settings", err.Error()))
	}
	stored.UUID = api.NewOptString(row.UUID.String())
	stored.Name = row.Name
	stored.IsEnabled = api.NewOptBool(row.IsEnabled)

	return &stored, nil
}
