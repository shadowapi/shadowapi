package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// OAuth2ClientLogin starts the OAuth2 login flow.
// It supports two different calling patterns:
//  1. Deprecated: body contains `client_id` pointing directly to oauth2_client.uuid.
//  2. Preferred:  body contains `query.datasource_uuid` – we first resolve the datasource, then its attached OAuth2 client.
//
// The second form is what the current React UI issues.
func (h *Handler) OAuth2ClientLogin(ctx context.Context, req *api.OAuth2ClientLoginReq) (*api.OAuth2ClientLoginOK, error) {
	log := h.log.With("handler", "OAuth2ClientLogin")
	q := query.New(h.dbp)

	// ---------------------------------------------------------------------
	// 1. Figure‑out which oauth2_client we should use
	// ---------------------------------------------------------------------
	var clientUUIDStr string

	// Preferred path – resolve from datasource_uuid
	if dsIDs, ok := req.Query["datasource_uuid"]; ok && len(dsIDs) > 0 {
		dsUUIDPg, err := converter.ConvertStringToPgUUID(dsIDs[0])
		if err != nil {
			log.Error("invalid datasource UUID", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
		}

		// Load datasource – we only need settings json → oauth2_client_uuid
		dsRow, err := q.GetDatasource(ctx, dsUUIDPg)
		if err == pgx.ErrNoRows {
			log.Error("datasource not found")
			return nil, ErrWithCode(http.StatusNotFound, E("datasource not found"))
		} else if err != nil {
			log.Error("failed to load datasource", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("internal error"))
		}

		// We expect settings column to contain JSON; we unmarshal just to grab the field.
		var settings struct {
			OAuth2ClientUUID string `json:"oauth2_client_uuid"`
		}
		if err := json.Unmarshal(dsRow.Datasource.Settings, &settings); err != nil {
			log.Error("broken datasource.settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("broken datasource settings"))
		}
		if settings.OAuth2ClientUUID == "" {
			log.Error("datasource missing oauth2_client_uuid")
			return nil, ErrWithCode(http.StatusBadRequest, E("datasource not linked to oauth2 client"))
		}
		clientUUIDStr = settings.OAuth2ClientUUID
	}

	// ------------------------------------------------------------------
	// 2. Build provider‑specific oauth2.Config
	// ------------------------------------------------------------------
	provider, err := oauthTools.GetClientConfig(ctx, h.dbp, clientUUIDStr)
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}

	// ------------------------------------------------------------------
	// 3. Persist state to DB (10‑minute expiration)
	// ------------------------------------------------------------------
	queryData, _ := json.Marshal(req.Query) // never fails, query is simple map
	stateID := uuid.Must(uuid.NewV7())

	clientPgUUID, err := converter.ConvertStringToPgUUID(clientUUIDStr)
	if err != nil {
		log.Error("invalid oauth2 client uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	if _, err := q.CreateOauth2State(ctx, query.CreateOauth2StateParams{
		UUID:       converter.UuidToPgUUID(stateID),
		ClientUuid: clientPgUUID,
		State:      queryData,
		ExpiredAt:  pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	}); err != nil {
		log.Error("failed to create oauth2 state", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create oauth2 state"))
	}

	// ------------------------------------------------------------------
	// 4. Produce AuthCodeURL and hand it back to UI
	// ------------------------------------------------------------------
	return &api.OAuth2ClientLoginOK{
		AuthCodeURL: provider.AuthCodeURL(stateID.String(), oauth2.AccessTypeOffline, oauth2.ApprovalForce),
	}, nil
}

// OAuth2ClientCallback exchanges the `code` for tokens, persists them and redirects back to UI.
func (h *Handler) OAuth2ClientCallback(ctx context.Context, params api.OAuth2ClientCallbackParams) (*api.OAuth2ClientCallbackFound, error) {
	log := h.log.With("handler", "OAuth2ClientCallback")
	q := query.New(h.dbp)

	// 1. Parse state
	stateUUID, err := converter.ConvertStringToPgUUID(params.State.Value)
	if err != nil {
		log.Error("broken state parameter", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("broken state parameter"))
	}
	stateRow, err := q.GetOauth2State(ctx, stateUUID)
	if err != nil {
		log.Error("failed to query state object", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to query state object"))
	}

	// Expired?
	if stateRow.Oauth2State.ExpiredAt.Valid && time.Now().After(stateRow.Oauth2State.ExpiredAt.Time) {
		return nil, ErrWithCode(http.StatusBadRequest, E("state expired"))
	}

	// 2. Unmarshal original query values so we know where to return later
	var stateQuery url.Values
	_ = json.Unmarshal(stateRow.Oauth2State.State, &stateQuery) // fail‑open – original code already validated

	// 3. Build provider config again from client_uuid
	config, err := oauthTools.GetClientConfig(ctx, h.dbp, stateRow.Oauth2State.ClientUuid.String())
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get client config"))
	}

	// 4. Exchange code → token
	tok, err := config.Exchange(ctx, params.Code.Value)
	if err != nil {
		log.Error("token exchange failed", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("failed token exchange"))
	}
	// Force fill expiry / refresh etc.
	tok, _ = config.TokenSource(ctx, tok).Token()

	switch strings.ToLower(config.Provider) {
	case "gmail":
		return h.handleGmailToken(ctx, log, tok, stateQuery, stateRow.Oauth2State.ClientUuid.String())
	default:
		return nil, ErrWithCode(http.StatusInternalServerError, E("unknown provider"))
	}
}

// OAuth2ClientTokenDelete revokes the token at the provider and removes it from DB.
func (h *Handler) OAuth2ClientTokenDelete(ctx context.Context, params api.OAuth2ClientTokenDeleteParams) error {
	log := h.log.With("handler", "OAuth2ClientTokenDelete", "tokenUUID", params.UUID)
	q := query.New(h.dbp)

	tokenPgUUID, err := converter.ConvertStringToPgUUID(params.UUID)
	if err != nil {
		return ErrWithCode(http.StatusBadRequest, E("invalid UUID"))
	}

	// Retrieve token row so we know which provider to revoke against.
	tokRow, err := q.GetOauth2TokenByUUID(ctx, tokenPgUUID)
	if err == pgx.ErrNoRows {
		return ErrWithCode(http.StatusNotFound, E("token not found"))
	} else if err != nil {
		log.Error("failed to fetch token", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal error"))
	}

	clientCfg, err := oauthTools.GetClientConfig(ctx, h.dbp, tokRow.Oauth2Token.ClientUuid.String())
	if err != nil {
		log.Error("failed to resolve client", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal error"))
	}

	// Unmarshal stored token JSON so we can pick refresh/access tokens
	var stored oauth2.Token
	if err := json.Unmarshal(tokRow.Oauth2Token.Token, &stored); err != nil {
		log.Warn("broken stored token json", "error", err)
	}

	// Provider‑specific revoke
	switch strings.ToLower(clientCfg.Provider) {
	case "gmail":
		_ = revokeGoogleToken(ctx, stored.AccessToken)
	}

	// Finally delete from DB
	if err := q.DeleteOauth2Token(ctx, tokenPgUUID); err != nil {
		log.Error("failed to delete oauth2 token", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("internal error"))
	}
	return nil
}

// revokeGoogleToken calls https://oauth2.googleapis.com/revoke
func revokeGoogleToken(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return nil // nothing to revoke
	}
	form := url.Values{}
	form.Set("token", accessToken)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

/*

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
*/

/*
// OLD CODE
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
*/
