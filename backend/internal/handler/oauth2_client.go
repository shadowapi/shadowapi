package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"

	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) OAuth2ClientCallback(
	ctx context.Context, params api.OAuth2ClientCallbackParams,
) (*api.OAuth2ClientCallbackFound, error) {
	log := h.log.With("handler", "OAuth2ClientCallback")
	tx := query.New(h.dbp)
	stateUUID, err := uuid.FromString(params.State.Value)
	if stateUUID.IsNil() {
		log.Error("broken state parameter", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("broken state parameter"))
	}

	state, err := tx.GetOauth2State(ctx, stateUUID)
	if err != nil {
		log.Error("query state object", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed query state object"))
	}

	// we expect that state contains all query parameters
	var stateQuery url.Values
	if err = json.Unmarshal(state.State, &stateQuery); err != nil {
		log.Error("unmarshal state query", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed unmarshal state query"))
	}

	config, err := oauthTools.GetClientConfig(ctx, h.dbp, state.ClientID)
	if err != nil {
		log.Error("get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed get client config"))
	}

	token, err := config.Exchange(ctx, params.Code.Value)
	if err != nil {
		log.Error("wrong exchange token", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("wrong exchange token"))
	}

	token, err = config.TokenSource(ctx, token).Token()
	if err != nil {
		log.Error("get token source", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed get token source"))
	}

	switch strings.ToLower(config.Provider) {
	case "gmail":
		return h.handleGmailToken(ctx, log, token, stateQuery, state.ClientID)
	}
	return nil, ErrWithCode(http.StatusInternalServerError, E("unknown provider"))
}

func (h *Handler) OAuth2ClientCreate(
	ctx context.Context, req *api.OAuth2ClientCreateReq,
) (*api.OAuth2Client, error) {
	log := h.log.With("handler", "OAuth2ClientCreate")
	create := query.CreateOauth2ClientParams{
		ID:       req.ID,
		Name:     req.Name,
		Secret:   req.Secret,
		Provider: req.Provider,
	}
	obj, err := query.New(h.dbp).CreateOauth2Client(ctx, create)
	if err != nil {
		log.Error("failed create oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	out := api.OAuth2Client{
		ID:       obj.ID,
		Name:     obj.Name,
		Provider: obj.Provider,
	}
	if obj.CreatedAt.Valid {
		out.CreatedAt = obj.CreatedAt.Time
	}
	if obj.UpdatedAt.Valid {
		out.UpdatedAt = obj.UpdatedAt.Time
	}
	return &out, nil
}

func (h *Handler) OAuth2ClientDelete(ctx context.Context, params api.OAuth2ClientDeleteParams) error {
	log := h.log.With("handler", "OAuth2ClientDelete")
	err := query.New(h.dbp).DeleteOauth2Client(ctx, params.ID)
	if err == pgx.ErrNoRows {
		log.Error("no such oauth2 client", "uuid", params.ID)
		return ErrWithCode(http.StatusNotFound, E("no such oauth2 client"))
	} else if err != nil {
		log.Error("failed delete oauth2 client", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	return nil
}

func (h *Handler) OAuth2ClientGet(ctx context.Context, params api.OAuth2ClientGetParams) (*api.OAuth2Client, error) {
	log := h.log.With("handler", "OAuth2ClientGet")
	tx := query.New(h.dbp)
	details, err := tx.GetOauth2Client(ctx, params.ID)
	if err != nil {
		log.Error("failed to get client details", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client details"))
	}
	result := &api.OAuth2Client{
		ID:       details.ID,
		Name:     details.Name,
		Provider: details.Provider,
	}
	if details.CreatedAt.Valid {
		result.CreatedAt = details.CreatedAt.Time
	}
	if details.UpdatedAt.Valid {
		result.UpdatedAt = details.UpdatedAt.Time
	}
	return result, nil
}

func (h *Handler) OAuth2ClientList(
	ctx context.Context, params api.OAuth2ClientListParams,
) (*api.OAuth2ClientListOK, error) {
	log := h.log.With("handler", "OAuth2ClientList")
	clients, err := query.New(h.dbp).ListOauth2Clients(ctx)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("no oauth2 clients")
		return &api.OAuth2ClientListOK{}, nil
	} else if err != nil {
		log.Error("list oauth2 clients", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	out := &api.OAuth2ClientListOK{}
	for _, c := range clients {
		a := api.OAuth2Client{
			ID:       c.ID,
			Name:     c.Name,
			Provider: c.Provider,
		}
		if c.CreatedAt.Valid {
			a.CreatedAt = c.CreatedAt.Time
		}
		if c.UpdatedAt.Valid {
			a.UpdatedAt = c.UpdatedAt.Time
		}

		out.Clients = append(out.Clients, a)
	}
	return out, nil
}

func (h *Handler) OAuth2ClientLogin(
	ctx context.Context, req *api.OAuth2ClientLoginReq,
) (*api.OAuth2ClientLoginOK, error) {
	log := h.log.With("handler", "OAuth2ClientLogin")
	tx := query.New(h.dbp)
	provider, err := oauthTools.GetClientConfig(ctx, h.dbp, req.ClientID)
	if err != nil {
		log.Error("get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}

	queryData, err := json.Marshal(req.Query)
	if err != nil {
		log.Error("marshal query values", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal query values"))
	}

	state, err := tx.CreateOauth2State(ctx, query.CreateOauth2StateParams{
		UUID:       uuid.Must(uuid.NewV7()),
		ClientName: provider.Name,
		ClientID:   provider.ClientID,
		State:      queryData,
	})
	if err != nil {
		log.Error("create oauth2 state", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create oauth2 state"))
	}

	h.log.Info("redirecting to oauth2 provider", "clientID", provider.ClientID)
	return &api.OAuth2ClientLoginOK{
		AuthCodeURL: provider.AuthCodeURL(state.UUID.String(), oauth2.AccessTypeOffline),
	}, nil
}

func (h *Handler) OAuth2ClientTokenDelete(
	ctx context.Context, params api.OAuth2ClientTokenDeleteParams,
) error {
	log := h.log.With("handler", "OAuth2ClientTokenDelete", "tokenUUID", params.UUID)
	tx := query.New(h.dbp)
	tokenUUID := uuid.Must(uuid.FromString(params.UUID))
	if err := tx.DeleteOauth2Token(ctx, tokenUUID); err == pgx.ErrNoRows {
		log.Error("no such oauth2 token")
		return ErrWithCode(http.StatusNotFound, E("no such oauth2 token"))
	} else if err != nil {
		log.Error("delete oauth2 token", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	return nil
}

func (h *Handler) OAuth2ClientTokenList(
	ctx context.Context, params api.OAuth2ClientTokenListParams,
) ([]api.OAuth2ClientToken, error) {
	log := h.log.With("handler", "OAuth2ClientTokenList", "datasourceUUID", params.DatasourceUUID)
	tx := query.New(h.dbp)

	client, err := tx.GetOauth2Client(ctx, params.DatasourceUUID)
	if err == pgx.ErrNoRows {
		log.Error("no oauth2 client")
		return nil, ErrWithCode(http.StatusNotFound, E("client not found"))
	} else if err != nil {
		log.Error("get oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}

	tokens, err := tx.GetOauth2ClientTokens(ctx, client.ID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("no client tokens")
		return []api.OAuth2ClientToken{}, nil
	} else if err != nil {
		log.Error("get client tokens", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	out := []api.OAuth2ClientToken{}
	for _, t := range tokens {
		a := api.OAuth2ClientToken{
			UUID:     t.UUID.String(),
			ClientID: t.ClientID,
			Name:     t.Name.String,
			Token:    string(t.Token),
		}
		if t.CreatedAt.Valid {
			a.CreatedAt = t.CreatedAt.Time
		}
		if t.UpdatedAt.Valid {
			a.UpdatedAt = t.UpdatedAt.Time
		}

		out = append(out, a)
	}
	return out, nil
}

func (h *Handler) OAuth2ClientUpdate(
	ctx context.Context, req *api.OAuth2ClientUpdateReq, params api.OAuth2ClientUpdateParams,
) (*api.OAuth2Client, error) {
	log := h.log.With("handler", "OAuth2ClientUpdate", "clientID", params.ID)
	tx := query.New(h.dbp)
	update := query.UpdateOauth2ClientParams{
		ID:       params.ID,
		Name:     req.Name,
		Provider: req.Provider,
		Secret:   req.Secret,
	}
	if err := tx.UpdateOauth2Client(ctx, update); err == pgx.ErrNoRows {
		log.Error("no such oauth2 client")
		return nil, ErrWithCode(http.StatusNotFound, E("no such oauth2 client"))
	} else if err != nil {
		log.Error("update oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	raw, err := tx.GetOauth2Client(ctx, params.ID)
	if err != nil {
		log.Error("get oauth2 client", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("internal server error"))
	}
	out := api.OAuth2Client{
		ID:       raw.ID,
		Name:     raw.Name,
		Provider: raw.Provider,
	}
	if raw.CreatedAt.Valid {
		out.CreatedAt = raw.CreatedAt.Time
	}
	if raw.UpdatedAt.Valid {
		out.UpdatedAt = raw.UpdatedAt.Time
	}

	return &out, nil
}
