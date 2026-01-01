package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrLoginChallengeExpired is returned when the login challenge has expired
var ErrLoginChallengeExpired = errors.New("login challenge expired")

// TokenResponse represents the OAuth2 token endpoint response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// HydraClient communicates with Ory Hydra
type HydraClient struct {
	publicURL  string
	adminURL   string
	httpClient *http.Client
	log        *slog.Logger
}

// NewHydraClient creates a new Hydra client
func NewHydraClient(publicURL, adminURL string, log *slog.Logger) *HydraClient {
	return &HydraClient{
		publicURL:  strings.TrimSuffix(publicURL, "/"),
		adminURL:   strings.TrimSuffix(adminURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		log:        log,
	}
}

// ExchangeCode exchanges an authorization code for tokens
func (c *HydraClient) ExchangeCode(ctx context.Context, code, redirectURI, codeVerifier, clientID string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {codeVerifier},
	}

	return c.tokenRequest(ctx, data)
}

// RefreshToken uses a refresh token to get new tokens
func (c *HydraClient) RefreshToken(ctx context.Context, refreshToken, clientID string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}

	return c.tokenRequest(ctx, data)
}

// RevokeToken revokes an access or refresh token
func (c *HydraClient) RevokeToken(ctx context.Context, token, clientID string) error {
	data := url.Values{
		"token":     {token},
		"client_id": {clientID},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.publicURL+"/oauth2/revoke",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	defer resp.Body.Close()

	// Revocation returns 200 on success (even if token was already revoked)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revocation failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.log.Debug("token revoked successfully")
	return nil
}

// tokenRequest makes a POST request to the token endpoint
func (c *HydraClient) tokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.publicURL+"/oauth2/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.log.Error("token request failed", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.log.Debug("token exchange successful",
		"expires_in", tokenResp.ExpiresIn,
		"scope", tokenResp.Scope,
		"has_refresh_token", tokenResp.RefreshToken != "",
	)

	return &tokenResp, nil
}

// BuildAuthorizationURL constructs the authorization URL for the OAuth2 flow
func (c *HydraClient) BuildAuthorizationURL(clientID, redirectURI, state, codeChallenge, scope string) string {
	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {scope},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	return c.publicURL + "/oauth2/auth?" + params.Encode()
}

// LoginRequest represents Hydra's login request info from admin API
type LoginRequest struct {
	Challenge      string   `json:"challenge"`
	RequestedScope []string `json:"requested_scope"`
	Subject        string   `json:"subject"`
	Skip           bool     `json:"skip"`
	Client         struct {
		ClientID string `json:"client_id"`
	} `json:"client"`
	RequestURL string `json:"request_url"`
}

// LoginAcceptResponse represents the response from accepting a login request
type LoginAcceptResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// GetLoginRequest fetches login request info from Hydra Admin API
func (c *HydraClient) GetLoginRequest(ctx context.Context, challenge string) (*LoginRequest, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.adminURL+"/admin/oauth2/auth/requests/login?login_challenge="+url.QueryEscape(challenge),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.log.Error("get login request failed", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("get login request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(resp.Body).Decode(&loginReq); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.log.Debug("got login request", "challenge", challenge, "skip", loginReq.Skip, "subject", loginReq.Subject)
	return &loginReq, nil
}

// AcceptLoginRequest accepts a login request and returns the redirect URL
func (c *HydraClient) AcceptLoginRequest(ctx context.Context, challenge, subject string, remember bool, rememberFor int) (string, error) {
	body := map[string]interface{}{
		"subject":      subject,
		"remember":     remember,
		"remember_for": rememberFor,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.adminURL+"/admin/oauth2/auth/requests/login/accept?login_challenge="+url.QueryEscape(challenge),
		strings.NewReader(string(bodyBytes)),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("accept login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("accept login request failed", "status", resp.StatusCode, "body", string(respBody))
		// Check for expired login challenge (401 with request_unauthorized)
		if resp.StatusCode == http.StatusUnauthorized && strings.Contains(string(respBody), "request_unauthorized") {
			return "", ErrLoginChallengeExpired
		}
		return "", fmt.Errorf("accept login request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var acceptResp LoginAcceptResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	c.log.Debug("login request accepted", "challenge", challenge, "redirect_to", acceptResp.RedirectTo)
	return acceptResp.RedirectTo, nil
}

// RejectLoginRequest rejects a login request and returns the redirect URL
func (c *HydraClient) RejectLoginRequest(ctx context.Context, challenge, errorID, errorDescription string) (string, error) {
	body := map[string]interface{}{
		"error":             errorID,
		"error_description": errorDescription,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.adminURL+"/admin/oauth2/auth/requests/login/reject?login_challenge="+url.QueryEscape(challenge),
		strings.NewReader(string(bodyBytes)),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("reject login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("reject login request failed", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("reject login request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var rejectResp LoginAcceptResponse
	if err := json.NewDecoder(resp.Body).Decode(&rejectResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	c.log.Debug("login request rejected", "challenge", challenge, "redirect_to", rejectResp.RedirectTo)
	return rejectResp.RedirectTo, nil
}

// ConsentRequest represents Hydra's consent request info from admin API
type ConsentRequest struct {
	Challenge                    string   `json:"challenge"`
	RequestedScope               []string `json:"requested_scope"`
	RequestedAccessTokenAudience []string `json:"requested_access_token_audience"`
	Subject                      string   `json:"subject"`
	Skip                         bool     `json:"skip"`
	Client                       struct {
		ClientID string `json:"client_id"`
	} `json:"client"`
	RequestURL string `json:"request_url"` // Original OAuth2 authorize URL with state parameter
}

// ConsentAcceptResponse represents the response from accepting a consent request
type ConsentAcceptResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// GetConsentRequest fetches consent request info from Hydra Admin API
func (c *HydraClient) GetConsentRequest(ctx context.Context, challenge string) (*ConsentRequest, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.adminURL+"/admin/oauth2/auth/requests/consent?consent_challenge="+url.QueryEscape(challenge),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get consent request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.log.Error("get consent request failed", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("get consent request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var consentReq ConsentRequest
	if err := json.NewDecoder(resp.Body).Decode(&consentReq); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.log.Debug("got consent request", "challenge", challenge, "skip", consentReq.Skip, "subject", consentReq.Subject)
	return &consentReq, nil
}

// AcceptConsentRequest accepts a consent request and returns the redirect URL
// accessTokenExtras contains additional claims to include in the access token (e.g., tenant info)
func (c *HydraClient) AcceptConsentRequest(ctx context.Context, challenge string, grantScope []string, grantAudience []string, subject string, remember bool, rememberFor int, accessTokenExtras map[string]interface{}) (string, error) {
	session := map[string]interface{}{
		"id_token": map[string]interface{}{
			"sub": subject,
		},
	}

	// Add extra claims to access token if provided
	if accessTokenExtras != nil && len(accessTokenExtras) > 0 {
		session["access_token"] = accessTokenExtras
	}

	body := map[string]interface{}{
		"grant_scope":                 grantScope,
		"grant_access_token_audience": grantAudience,
		"remember":                    remember,
		"remember_for":                rememberFor,
		"session":                     session,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.adminURL+"/admin/oauth2/auth/requests/consent/accept?consent_challenge="+url.QueryEscape(challenge),
		strings.NewReader(string(bodyBytes)),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("accept consent request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("accept consent request failed", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("accept consent request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var acceptResp ConsentAcceptResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	c.log.Debug("consent request accepted", "challenge", challenge, "redirect_to", acceptResp.RedirectTo)
	return acceptResp.RedirectTo, nil
}

// RevokeLoginSession revokes all login sessions for a subject
// This should be called during logout to prevent automatic re-authentication
func (c *HydraClient) RevokeLoginSession(ctx context.Context, subject string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.adminURL+"/admin/oauth2/auth/sessions/login?subject="+url.QueryEscape(subject),
		nil,
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke login session: %w", err)
	}
	defer resp.Body.Close()

	// 204 No Content on success, 404 if no session exists (both are OK)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		c.log.Error("revoke login session failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("revoke login session failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.log.Debug("login session revoked", "subject", subject)
	return nil
}
