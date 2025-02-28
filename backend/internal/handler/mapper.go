package handler

import (
	"encoding/json"
	"fmt"
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
	raw := map[string]string{}
	if err := json.Unmarshal(row.Settings, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	ret := &api.StoragePostgres{
		UUID: api.NewOptString(row.UUID.String()),
	}

	if v, ok := raw["user"]; ok {
		ret.User = v
	}
	if v, ok := raw["name"]; ok {
		ret.Name = v
	}
	if v, ok := raw["host"]; ok {
		ret.Host = v
	}
	if v, ok := raw["port"]; ok {
		ret.Port = v
	}
	if v, ok := raw["options"]; ok {
		ret.Options = api.NewOptString(v)
	}

	return ret, nil
}

func QToStorageS3(row query.GetStoragesRow) (*api.StorageS3, error) {
	// The JSON in row.Settings has the entire S3 object
	var stored api.StorageS3
	if err := json.Unmarshal(row.Settings, &stored); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal s3 settings"))
	}

	stored.UUID = api.NewOptString(row.UUID.String())
	// If you want to reflect the DB name/is_enabled, override here:
	stored.Name = row.Name
	stored.IsEnabled = api.NewOptBool(row.IsEnabled)

	return &stored, nil
}

func QToStorageHostfiles(row query.GetStoragesRow) (*api.StorageHostfiles, error) {
	// The JSON stored in row.Settings has the entire original api.StorageHostfiles object.
	var stored api.StorageHostfiles
	if err := json.Unmarshal(row.Settings, &stored); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal hostfiles settings"))
	}

	// Overwrite the UUID from the DB, just in case
	stored.UUID = api.NewOptString(row.UUID.String())

	// If we want to use the name/is_enabled from the top-level columns, we can overwrite here:
	stored.Name = row.Name
	stored.IsEnabled = api.NewOptBool(row.IsEnabled)

	return &stored, nil
}
