package handler

import (
	"encoding/json"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"net/http"
)

// QToDatasource converts a query.Datasource to an api.Datasource
func QToDatasource(row query.Datasource) api.Datasource {
	c := api.Datasource{
		UUID:      row.UUID.String(),
		Name:      row.Name,
		Type:      row.Type,
		IsEnabled: row.IsEnabled,
	}
	if row.UserUUID != nil {
		c.UserUUID = api.OptString{Value: row.UserUUID.String(), Set: true}
	}
	if row.OAuth2TokenUUID != nil {
		c.OAuth2TokenUUID = api.OptString{Value: row.OAuth2TokenUUID.String(), Set: true}
	}
	if row.Oauth2ClientID.Valid {
		c.OAuth2ClientID = api.OptString{Value: row.Oauth2ClientID.String, Set: true}
	}
	if row.CreatedAt.Valid {
		c.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		c.UpdatedAt = row.UpdatedAt.Time
	}
	return c
}

// QToDatasourceEmail extracts query.DatasourceEmail fields and set them to an api.Datasource
func QToDatasourceEmail(c *api.Datasource, row query.DatasourceEmail) {
	c.Email = api.OptString{Value: row.Email, Set: true}
	c.Provider = api.OptString{Value: row.Provider, Set: true}
	if row.IMAPServer.Valid {
		c.ImapServer = api.OptString{Value: row.IMAPServer.String, Set: true}
	}
	if row.SMTPServer.Valid {
		c.SMTPServer = api.OptString{Value: row.SMTPServer.String, Set: true}
	}
	if row.SMTPTLS.Valid {
		c.SMTPTLS = api.OptBool{Value: row.SMTPTLS.Bool, Set: true}
	}
}

func QToStorageOld(row query.GetStoragesRow) api.Storage {
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
