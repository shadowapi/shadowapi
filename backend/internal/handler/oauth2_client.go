package handler

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"

	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// OAuth2ClientCallback handles the OAuth2 callback and processes the token exchange.
func (h *Handler) OAuth2ClientCallback(ctx context.Context, params api.OAuth2ClientCallbackParams) (*api.OAuth2ClientCallbackFound, error) {
	log := h.log.With("handler", "OAuth2ClientCallback")
	q := query.New(h.dbp)

	// Parse the state parameter as a UUID.
	stateUUID, err := converter.ConvertStringToPgUUID(params.State.Value)
	if err != nil {
		log.Error("broken state parameter", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("broken state parameter"))
	}

	// Retrieve the state row using generated sqlc method.
	stateRow, err := q.GetOauth2State(ctx, stateUUID)
	if err != nil {
		log.Error("failed to query state object", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query state object"))
	}

	// Unmarshal the JSON state field.
	var stateQuery url.Values
	if err = json.Unmarshal(stateRow.Oauth2State.State, &stateQuery); err != nil {
		log.Error("failed to unmarshal state query", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal state query"))
	}

	// Use the ClientUuid from the state row to get client config.
	config, err := oauthTools.GetClientConfig(ctx, h.dbp, stateRow.Oauth2State.ClientUuid.String())
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}

	// Exchange the code for a token.
	token, err := config.Exchange(ctx, params.Code.Value)
	if err != nil {
		log.Error("token exchange failed", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("failed token exchange"))
	}

	// Refresh the token to ensure up-to-date details.
	token, err = config.TokenSource(ctx, token).Token()
	if err != nil {
		log.Error("failed to get token from token source", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get token from token source"))
	}

	switch strings.ToLower(config.Provider) {
	case "gmail":
		// Handle Gmail-specific logic passing the ClientUuid.
		return h.handleGmailToken(ctx, log, token, stateQuery, stateRow.Oauth2State.ClientUuid.String())
	default:
		return nil, ErrWithCode(http.StatusInternalServerError, E("unknown provider"))
	}
}

// OAuth2ClientCreate creates a new OAuth2 client.
func (h *Handler) OAuth2ClientCreate(ctx context.Context, req *api.OAuth2ClientCreateReq) (*api.OAuth2Client, error) {
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
func (h *Handler) OAuth2ClientDelete(ctx context.Context, params api.OAuth2ClientDeleteParams) error {
	log := h.log.With("handler", "OAuth2ClientDelete")
	clientUUID, err := converter.ConvertStringToPgUUID(params.UUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}
	err = query.New(h.dbp).DeleteOauth2Client(ctx, clientUUID)
	if err == pgx.ErrNoRows {
		log.Error("no such oauth2 client", "uuid", params.UUID)
		return ErrWithCode(http.StatusNotFound, E("no such oauth2 client"))
	} else if err != nil {
		log.Error("failed to delete oauth2 client", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	return nil
}

// OAuth2ClientGet retrieves OAuth2 client details.
func (h *Handler) OAuth2ClientGet(ctx context.Context, params api.OAuth2ClientGetParams) (*api.OAuth2Client, error) {
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
func (h *Handler) OAuth2ClientList(ctx context.Context, params api.OAuth2ClientListParams) (*api.OAuth2ClientListOK, error) {
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

// OAuth2ClientLogin starts the OAuth2 login flow.
func (h *Handler) OAuth2ClientLogin(ctx context.Context, req *api.OAuth2ClientLoginReq) (*api.OAuth2ClientLoginOK, error) {
	log := h.log.With("handler", "OAuth2ClientLogin")
	q := query.New(h.dbp)
	// Look up the client using the provided client_id.
	provider, err := oauthTools.GetClientConfig(ctx, h.dbp, req.ClientID)
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}
	queryData, err := json.Marshal(req.Query)
	if err != nil {
		log.Error("failed to marshal query values", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal query values"))
	}

	oauth2StateUUID := uuid.Must(uuid.NewV7())

	clientUuid, err := converter.ConvertStringToPgUUID(provider.ClientID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	// Create state using the provider's client id.
	state, err := q.CreateOauth2State(ctx, query.CreateOauth2StateParams{
		UUID:       converter.UuidToPgUUID(oauth2StateUUID),
		ClientUuid: clientUuid,
		State:      queryData,
		ExpiredAt:  pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	})
	if err != nil {
		log.Error("failed to create oauth2 state", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create oauth2 state"))
	}
	log.Info("redirecting to oauth2 provider", "clientUUID", provider.ClientID)
	return &api.OAuth2ClientLoginOK{
		AuthCodeURL: provider.AuthCodeURL(state.UUID.String(), oauth2.AccessTypeOffline),
	}, nil
}

// OAuth2ClientTokenDelete deletes an OAuth2 token.
func (h *Handler) OAuth2ClientTokenDelete(ctx context.Context, params api.OAuth2ClientTokenDeleteParams) error {
	log := h.log.With("handler", "OAuth2ClientTokenDelete", "tokenUUID", params.UUID)
	q := query.New(h.dbp)

	tokenUUID, err := converter.ConvertStringToPgUUID(params.UUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	if err := q.DeleteOauth2Token(ctx, tokenUUID); err == pgx.ErrNoRows {
		log.Error("no such oauth2 token")
		return ErrWithCode(http.StatusNotFound, E("no such oauth2 token"))
	} else if err != nil {
		log.Error("failed to delete oauth2 token", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	return nil
}

// OAuth2ClientTokenList lists tokens for the OAuth2 client associated with a datasource.
func (h *Handler) OAuth2ClientTokenList(ctx context.Context, params api.OAuth2ClientTokenListParams) ([]api.OAuth2ClientToken, error) {
	log := h.log.With("handler", "OAuth2ClientTokenList", "datasourceUUID", params.DatasourceUUID)
	q := query.New(h.dbp)

	dsUUID, err := converter.ConvertStringToPgUUID(params.DatasourceUUID)
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	// Retrieve the client based on datasource UUID.
	client, err := q.GetOauth2Client(ctx, dsUUID)
	if err == pgx.ErrNoRows {
		log.Error("no oauth2 client")
		return nil, ErrWithCode(http.StatusNotFound, E("client not found"))
	} else if err != nil {
		log.Error("failed to get oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}

	clientUUID, err := converter.ConvertStringToPgUUID(client.Oauth2Client.UUID.String())
	if err != nil {
		log.Error("invalid UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	tokens, err := q.GetOauth2ClientTokens(ctx, clientUUID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to get client tokens", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	var out []api.OAuth2ClientToken
	for _, t := range tokens {
		token := api.OAuth2ClientToken{
			UUID:       api.NewOptString(t.UUID.String()),
			ClientUUID: t.ClientUuid.String(),
			// TODO @reactima !!! fix first
			// Fixed conversion: convert []byte to string then to OAuth2ClientTokenToken.
			//Token: api.OAuth2ClientTokenToken(string(t.Token)),
		}
		if t.CreatedAt.Valid {
			token.CreatedAt = api.NewOptDateTime(t.CreatedAt.Time)
		}
		if t.UpdatedAt.Valid {
			token.UpdatedAt = api.NewOptDateTime(t.UpdatedAt.Time)
		}
		out = append(out, token)
	}
	return out, nil
}

// OAuth2ClientUpdate updates an OAuth2 client.
func (h *Handler) OAuth2ClientUpdate(ctx context.Context, req *api.OAuth2ClientUpdateReq, params api.OAuth2ClientUpdateParams) (*api.OAuth2Client, error) {
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
		Name:     raw.Oauth2Client.Name,
		Provider: raw.Oauth2Client.Provider,
		ClientID: raw.Oauth2Client.ClientID,
	}
	if raw.Oauth2Client.CreatedAt.Valid {
		out.CreatedAt = api.NewOptDateTime(raw.Oauth2Client.CreatedAt.Time)
	}
	if raw.Oauth2Client.UpdatedAt.Valid {
		out.UpdatedAt = api.NewOptDateTime(raw.Oauth2Client.UpdatedAt.Time)
	}
	return &out, nil
}
