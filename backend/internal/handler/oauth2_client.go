package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Standard code generate CRUD

// OAuth2ClientCreate creates a new OAuth2 client.
func (h *Handler) OAuth2ClientCreate(ctx context.Context, req *api.OAuth2ClientCreateReq) (api.OAuth2ClientCreateRes, error) {
	log := h.log.With("handler", "OAuth2ClientCreate")

	clientUUID := uuid.Must(uuid.NewV7())
	create := query.CreateOauth2ClientParams{
		UUID:     converter.UuidToPgUUID(clientUUID),
		Name:     req.Name,
		Provider: req.Provider,
		// New required field.
		ClientID: req.ClientID,
		Secret:   req.Secret,
	}
	obj, err := query.New(h.dbp).CreateOauth2Client(ctx, create)
	if err != nil {
		log.Error("failed to create oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	out := api.OAuth2Client{
		UUID:     api.NewOptString(obj.UUID.String()),
		Name:     obj.Name,
		Provider: obj.Provider,
		ClientID: obj.ClientID,
	}
	if obj.CreatedAt.Valid {
		out.CreatedAt = api.NewOptDateTime(obj.CreatedAt.Time)
	}
	if obj.UpdatedAt.Valid {
		out.UpdatedAt = api.NewOptDateTime(obj.UpdatedAt.Time)
	}
	return &out, nil
}

// OAuth2ClientDelete deletes an OAuth2 client.
func (h *Handler) OAuth2ClientDelete(ctx context.Context, params api.OAuth2ClientDeleteParams) (api.OAuth2ClientDeleteRes, error) {
	log := h.log.With("handler", "OAuth2ClientDelete")
	clientUUID, err := converter.ConvertStringToPgUUID(params.UUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}
	err = query.New(h.dbp).DeleteOauth2Client(ctx, clientUUID)
	if err == pgx.ErrNoRows {
		log.Error("no such oauth2 client", "uuid", params.UUID)
		return nil, ErrWithCode(http.StatusNotFound, E("no such oauth2 client"))
	} else if err != nil {
		log.Error("failed to delete oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	return &api.OAuth2ClientDeleteOK{}, nil
}

// OAuth2ClientGet retrieves OAuth2 client details.
func (h *Handler) OAuth2ClientGet(ctx context.Context, params api.OAuth2ClientGetParams) (api.OAuth2ClientGetRes, error) {
	log := h.log.With("handler", "OAuth2ClientGet")
	clientUUID, err := converter.ConvertStringToPgUUID(params.UUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}
	details, err := query.New(h.dbp).GetOauth2Client(ctx, clientUUID)
	if err != nil {
		log.Error("failed to get client details", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client details"))
	}
	result := &api.OAuth2Client{
		UUID:     api.NewOptString(details.Oauth2Client.UUID.String()),
		Name:     details.Oauth2Client.Name,
		Provider: details.Oauth2Client.Provider,
		ClientID: details.Oauth2Client.ClientID,
		Secret:   details.Oauth2Client.Secret,
	}
	if details.Oauth2Client.CreatedAt.Valid {
		result.CreatedAt = api.NewOptDateTime(details.Oauth2Client.CreatedAt.Time)
	}
	if details.Oauth2Client.UpdatedAt.Valid {
		result.UpdatedAt = api.NewOptDateTime(details.Oauth2Client.UpdatedAt.Time)
	}
	return result, nil
}

// OAuth2ClientList lists all OAuth2 clients.
func (h *Handler) OAuth2ClientList(ctx context.Context, params api.OAuth2ClientListParams) (api.OAuth2ClientListRes, error) {
	log := h.log.With("handler", "OAuth2ClientList")

	offset := params.Offset.Or(0)
	limit := params.Limit.Or(0)
	qParams := query.ListOauth2ClientsParams{
		Offset: offset,
		Limit:  limit,
	}
	clients, err := query.New(h.dbp).ListOauth2Clients(ctx, qParams)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to list oauth2 clients", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	out := &api.OAuth2ClientListOK{}
	for _, c := range clients {
		a := api.OAuth2Client{
			UUID:     api.NewOptString(c.Oauth2Client.UUID.String()),
			Name:     c.Oauth2Client.Name,
			Provider: c.Oauth2Client.Provider,
			ClientID: c.Oauth2Client.ClientID,
			Secret:   c.Oauth2Client.Secret,
		}
		if c.Oauth2Client.CreatedAt.Valid {
			a.CreatedAt = api.NewOptDateTime(c.Oauth2Client.CreatedAt.Time)
		}
		if c.Oauth2Client.UpdatedAt.Valid {
			a.UpdatedAt = api.NewOptDateTime(c.Oauth2Client.UpdatedAt.Time)
		}
		out.Clients = append(out.Clients, a)
	}
	return out, nil
}

// OAuth2ClientTokenList lists tokens for the OAuth2 client associated with a datasource.
func (h *Handler) OAuth2ClientTokenList(ctx context.Context, params api.OAuth2ClientTokenListParams) (api.OAuth2ClientTokenListRes, error) {
	log := h.log.With("handler", "OAuth2ClientTokenList", "datasourceUUID", params.DatasourceUUID)
	q := query.New(h.dbp)

	dsUUID, err := converter.ConvertStringToPgUUID(params.DatasourceUUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	// First, get the datasource to find the oauth2_client_uuid from its settings
	dsRow, err := q.GetDatasource(ctx, dsUUID)
	if err == pgx.ErrNoRows {
		log.Error("datasource not found")
		return nil, ErrWithCode(http.StatusNotFound, E("datasource not found"))
	} else if err != nil {
		log.Error("failed to get datasource", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}

	// Parse settings to extract oauth2_client_uuid
	var settings struct {
		OAuth2ClientUUID string `json:"oauth2_client_uuid"`
	}
	if err := json.Unmarshal(dsRow.Datasource.Settings, &settings); err != nil {
		log.Error("failed to parse datasource settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to parse datasource settings"))
	}
	if settings.OAuth2ClientUUID == "" {
		log.Error("datasource has no oauth2_client_uuid")
		return nil, ErrWithCode(http.StatusBadRequest, E("datasource not linked to oauth2 client"))
	}

	// Now get the OAuth2 client using the correct UUID from settings
	clientUUID, err := converter.ConvertStringToPgUUID(settings.OAuth2ClientUUID)
	if err != nil {
		log.Error("invalid oauth2_client_uuid in settings", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid oauth2 client UUID"))
	}

	tokens, err := q.GetOauth2TokensByClientUUID(ctx, clientUUID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to get client tokens", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}

	var out []api.OAuth2ClientToken
	for _, t := range tokens {
		// Convert the raw JSONB token field into a string.
		tokenStr := string(t.Oauth2Token.Token)
		// Unmarshal the token JSON into tokenObj.
		var tokenObj api.OAuth2ClientTokenObj
		if err := tokenObj.UnmarshalJSON([]byte(tokenStr)); err != nil {
			log.Error("failed to unmarshal oauth token", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to decode oauth token"))
		}
		token := api.OAuth2ClientToken{
			UUID:       api.NewOptString(t.Oauth2Token.UUID.String()),
			ClientUUID: t.Oauth2Token.ClientUuid.String(),
			Token:      tokenObj,
		}
		if t.Oauth2Token.CreatedAt.Valid {
			token.CreatedAt = api.NewOptDateTime(t.Oauth2Token.CreatedAt.Time)
		}
		if t.Oauth2Token.UpdatedAt.Valid {
			token.UpdatedAt = api.NewOptDateTime(t.Oauth2Token.UpdatedAt.Time)
		}
		out = append(out, token)
	}
	res := api.OAuth2ClientTokenListOKApplicationJSON(out)
	return &res, nil
}

// OAuth2ClientUpdate updates an OAuth2 client.
func (h *Handler) OAuth2ClientUpdate(ctx context.Context, req *api.OAuth2ClientUpdateReq, params api.OAuth2ClientUpdateParams) (api.OAuth2ClientUpdateRes, error) {
	log := h.log.With("handler", "OAuth2ClientUpdate", "clientUUID", params.UUID)
	q := query.New(h.dbp)
	clientUUID, err := converter.ConvertStringToPgUUID(params.UUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid client id"))
	}
	update := query.UpdateOauth2ClientParams{
		Name:     req.Name,
		Provider: req.Provider,
		Secret:   req.Secret,
		ClientID: req.ClientID,
		UUID:     clientUUID,
	}
	if err := q.UpdateOauth2Client(ctx, update); err == pgx.ErrNoRows {
		log.Error("no such oauth2 client")
		return nil, ErrWithCode(http.StatusNotFound, E("no such oauth2 client"))
	} else if err != nil {
		log.Error("failed to update oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	raw, err := q.GetOauth2Client(ctx, clientUUID)
	if err != nil {
		log.Error("failed to get oauth2 client after update", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}

	out := api.OAuth2Client{
		UUID:      api.NewOptString(raw.Oauth2Client.UUID.String()),
		Name:      raw.Oauth2Client.Name,
		Provider:  raw.Oauth2Client.Provider,
		ClientID:  raw.Oauth2Client.ClientID,
		Secret:    raw.Oauth2Client.Secret,
		CreatedAt: api.NewOptDateTime(raw.Oauth2Client.CreatedAt.Time),
		UpdatedAt: api.NewOptDateTime(raw.Oauth2Client.UpdatedAt.Time),
	}
	if raw.Oauth2Client.CreatedAt.Valid {
		out.CreatedAt = api.NewOptDateTime(raw.Oauth2Client.CreatedAt.Time)
	}
	if raw.Oauth2Client.UpdatedAt.Valid {
		out.UpdatedAt = api.NewOptDateTime(raw.Oauth2Client.UpdatedAt.Time)
	}
	return &out, nil
}
