package handler

import (
	"encoding/json"
	"fmt"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
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
		ret.User = api.NewOptString(v)
	}
	if v, ok := raw["name"]; ok {
		ret.Name = v
	}
	if v, ok := raw["host"]; ok {
		ret.Host = v
	}
	if v, ok := raw["port"]; ok {
		ret.Port = api.NewOptString(v)
	}
	if v, ok := raw["options"]; ok {
		ret.Options = api.NewOptString(v)
	}

	return ret, nil
}

func QToStorageS3(row query.GetStoragesRow) (*api.StorageS3, error) {
	var s3 api.StorageS3
	if err := json.Unmarshal(row.Settings, &s3); err != nil {
		return nil, fmt.Errorf("failed to unmarshal s3 settings: %w", err)
	}
	s3.UUID = api.NewOptString(row.UUID.String())
	return &s3, nil
}

func QToStorageHostfiles(row query.GetStoragesRow) (*api.StorageHostfiles, error) {
	var raw map[string]string
	if err := json.Unmarshal(row.Settings, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hostfiles settings: %w", err)
	}
	ret := &api.StorageHostfiles{
		UUID: api.NewOptString(row.UUID.String()),
		Path: raw["path"],
	}
	return ret, nil
}
