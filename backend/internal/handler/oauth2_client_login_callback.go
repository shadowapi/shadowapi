package handler

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuth2ClientLogin starts the OAuth2 login flow.
func (h *Handler) OAuth2ClientLogin(ctx context.Context, req *api.OAuth2ClientLoginReq) (*api.OAuth2ClientLoginOK, error) {
	log := h.log.With("handler", "OAuth2ClientLogin")
	q := query.New(h.dbp)

	// 1) Load the OAuth2Client row by req.ClientID:
	provider, err := oauthTools.GetClientConfig(ctx, h.dbp, req.ClientID)
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}

	// 2) Marshal any extra query values (e.g. datasource_uuid)
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

	// 3) Persist into oauth2_state:
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

	// 4) Build the AuthCodeURL and return it:
	authCodeURL := provider.AuthCodeURL(state.UUID.String(), oauth2.AccessTypeOffline)
	log.Info("redirecting to oauth2 provider", "clientUUID", provider.ClientID, "authCodeURL", authCodeURL)
	return &api.OAuth2ClientLoginOK{
		AuthCodeURL: authCodeURL,
	}, nil
}

// OAuth2ClientCallback handles the OAuth2 callback and processes the token exchange.
// defined in spec/paths/oauth2_callback.yaml
// Important! This needs to be reimplemented in API first mode
func (h *Handler) OAuth2ClientCallback(ctx context.Context, params api.OAuth2ClientCallbackParams) (*api.OAuth2ClientCallbackFound, error) {
	log := h.log.With("handler", "OAuth2ClientCallback")
	q := query.New(h.dbp)

	// 1) Parse the state parameter as a UUID.
	stateUUID, err := converter.ConvertStringToPgUUID(params.State.Value)
	if err != nil {
		log.Error("broken state parameter", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("broken state parameter"))
	}

	// Retrieve the state row using the generated sqlc method.
	// Note: The returned row now wraps the state fields in "Oauth2State".
	stateRow, err := q.GetOauth2State(ctx, stateUUID)
	if err != nil {
		log.Error("failed to query state object", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query state object"))
	}

	// 2) Unmarshal the original query blob:
	var stateQuery url.Values
	if err = json.Unmarshal(stateRow.Oauth2State.State, &stateQuery); err != nil {
		log.Error("failed to unmarshal state query", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal state query"))
	}

	// 3) Re‑build the oauth2.Config for this client:
	config, err := oauthTools.GetClientConfig(ctx, h.dbp, stateRow.Oauth2State.ClientUuid.String())
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}

	// 3) rebuild provider config
	token, err := config.Exchange(ctx, params.Code.Value)
	if err != nil {
		log.Error("token exchange failed", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("failed token exchange"))
	}

	// 4) exchange code → token
	token, err = config.TokenSource(ctx, token).Token()
	if err != nil {
		log.Error("failed to get token from token source", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get token from token source"))
	}

	// 5) provider‑specific persistence
	switch strings.ToLower(config.Provider) {
	case "gmail":
		// Handle Gmail-specific logic passing the ClientUuid.
		return h.handleGmailToken(ctx, log, token, stateQuery, stateRow.Oauth2State.ClientUuid.String())
	default:
		log.Error("unsupported provider", "provider", config.Provider)
		return nil, ErrWithCode(http.StatusInternalServerError, E("unknown provider"))
	}
}
