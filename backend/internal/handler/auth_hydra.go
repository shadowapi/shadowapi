package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// tenantInfo holds tenant data extracted from OAuth2 state
type tenantInfo struct {
	UUID        string
	Name        string
	RedirectURI string
}

// extractTenantFromRequestURL extracts tenant info from the OAuth2 request URL.
// The request_url contains our state parameter which maps to stored tenant info.
func (h *Handler) extractTenantFromRequestURL(requestURL string) *tenantInfo {
	if requestURL == "" || h.oauth2Svc == nil {
		return nil
	}

	// Parse the URL to extract the state parameter
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		h.log.Debug("failed to parse request_url", "url", requestURL, "error", err)
		return nil
	}

	state := parsedURL.Query().Get("state")
	if state == "" {
		h.log.Debug("no state parameter in request_url", "url", requestURL)
		return nil
	}

	// Look up the state in our store (don't delete it yet, we might need it in callback)
	h.oauth2Svc.stateMu.RLock()
	stateData, ok := h.oauth2Svc.stateStore[state]
	h.oauth2Svc.stateMu.RUnlock()

	if !ok {
		h.log.Debug("state not found in store", "state", state[:8]+"...")
		return nil
	}

	h.log.Debug("extracted tenant from OAuth2 state",
		"tenant_uuid", stateData.TenantUUID,
		"tenant_name", stateData.TenantName,
		"redirect_uri", stateData.RedirectURI,
	)

	return &tenantInfo{
		UUID:        stateData.TenantUUID,
		Name:        stateData.TenantName,
		RedirectURI: stateData.RedirectURI,
	}
}

// extractHostFromRedirectURI extracts the host (scheme + host) from a redirect URI
func extractHostFromRedirectURI(redirectURI string) string {
	if redirectURI == "" {
		return ""
	}

	parsedURL, err := url.Parse(redirectURI)
	if err != nil {
		return ""
	}

	return parsedURL.Scheme + "://" + parsedURL.Host
}

// AuthLogin handles the Hydra login redirect.
// GET /api/v1/auth/login?login_challenge=xxx
// If the user has a remembered session, accept immediately. Otherwise redirect to frontend login page.
func (h *Handler) AuthLogin(ctx context.Context, params api.AuthLoginParams) (*api.AuthLoginFound, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Get login request from Hydra
	loginReq, err := h.oauth2Svc.hydraClient.GetLoginRequest(ctx, params.LoginChallenge)
	if err != nil {
		h.log.Error("failed to get login request", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("invalid login challenge"))
	}

	// If skip is true, the user has a remembered session - accept immediately
	if loginReq.Skip {
		redirectURL, err := h.oauth2Svc.hydraClient.AcceptLoginRequest(
			ctx,
			params.LoginChallenge,
			loginReq.Subject,
			false, // don't extend remember
			0,
		)
		if err != nil {
			h.log.Error("failed to accept skipped login", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("login acceptance failed"))
		}

		h.log.Debug("login skipped (remembered session)", "subject", loginReq.Subject)
		return &api.AuthLoginFound{
			Location: api.NewOptString(redirectURL),
		}, nil
	}

	// Redirect to frontend login page with the challenge
	// Extract the original redirect host from the OAuth2 request URL (preserves tenant subdomain)
	var redirectHost string
	if tenantInfo := h.extractTenantFromRequestURL(loginReq.RequestURL); tenantInfo != nil {
		redirectHost = extractHostFromRedirectURI(tenantInfo.RedirectURI)
	}
	if redirectHost == "" {
		redirectHost = h.oauth2Svc.baseURL // fallback to configured baseURL
	}
	loginPageURL := fmt.Sprintf("%s/login?login_challenge=%s", redirectHost, params.LoginChallenge)

	h.log.Debug("redirecting to login page", "challenge", params.LoginChallenge[:8]+"...", "host", redirectHost)
	return &api.AuthLoginFound{
		Location: api.NewOptString(loginPageURL),
	}, nil
}

// AuthLoginSubmit handles login form submission.
// POST /api/v1/auth/login
// Authenticates user credentials and accepts the Hydra login request.
func (h *Handler) AuthLoginSubmit(ctx context.Context, req *api.AuthLoginSubmitReq) (*api.AuthLoginSubmitOK, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Authenticate user with dbauth
	user, err := h.userManager.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		h.log.Debug("authentication failed", "email", req.Email, "error", err)
		return nil, err // Returns 401 for invalid credentials
	}

	// Get subject (user UUID)
	subject := user.UUID.Value

	// Determine remember settings
	remember := req.Remember.Or(false)
	rememberFor := 0
	if remember {
		rememberFor = 3600 // 1 hour
	}

	// Accept login request with Hydra
	redirectURL, err := h.oauth2Svc.hydraClient.AcceptLoginRequest(
		ctx,
		req.LoginChallenge,
		subject,
		remember,
		rememberFor,
	)
	if err != nil {
		h.log.Error("failed to accept login request", "error", err, "subject", subject)
		return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("login acceptance failed"))
	}

	h.log.Info("user authenticated successfully", "email", req.Email, "subject", subject)

	return &api.AuthLoginSubmitOK{
		RedirectTo: redirectURL,
	}, nil
}

// AuthConsent handles the Hydra consent redirect.
// GET /api/v1/auth/consent?consent_challenge=xxx
// Auto-approves consent and redirects back to Hydra.
func (h *Handler) AuthConsent(ctx context.Context, params api.AuthConsentParams) (*api.AuthConsentFound, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Get consent request from Hydra
	consentReq, err := h.oauth2Svc.hydraClient.GetConsentRequest(ctx, params.ConsentChallenge)
	if err != nil {
		h.log.Error("failed to get consent request", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("invalid consent challenge"))
	}

	// Build access token extras with tenant information
	// Extract tenant from the original OAuth2 request URL (stored in state when flow started)
	var accessTokenExtras map[string]interface{}
	if tenantInfo := h.extractTenantFromRequestURL(consentReq.RequestURL); tenantInfo != nil && tenantInfo.UUID != "" {
		accessTokenExtras = map[string]interface{}{
			"tenant_uuid": tenantInfo.UUID,
			"tenant_name": tenantInfo.Name,
		}
		h.log.Debug("including tenant in consent from OAuth2 state", "tenant_uuid", tenantInfo.UUID, "tenant_name", tenantInfo.Name)
	} else {
		h.log.Debug("no tenant info found in OAuth2 state for consent")
	}

	// Auto-approve: Accept all requested scopes and audiences
	redirectURL, err := h.oauth2Svc.hydraClient.AcceptConsentRequest(
		ctx,
		params.ConsentChallenge,
		consentReq.RequestedScope,
		consentReq.RequestedAccessTokenAudience,
		consentReq.Subject,
		true,  // remember consent
		3600,  // remember for 1 hour
		accessTokenExtras,
	)
	if err != nil {
		h.log.Error("failed to accept consent request", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, fmt.Errorf("consent acceptance failed"))
	}

	h.log.Debug("consent auto-approved",
		"subject", consentReq.Subject,
		"scopes", consentReq.RequestedScope,
	)

	return &api.AuthConsentFound{
		Location: api.NewOptString(redirectURL),
	}, nil
}
