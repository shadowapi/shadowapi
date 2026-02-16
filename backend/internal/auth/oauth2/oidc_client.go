package oauth2

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenResponse represents the OAuth2 token endpoint response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// OIDCClient communicates with the external OIDC provider (oxoauth.com)
type OIDCClient struct {
	issuerURL      string
	clientID       string
	callbackSecret string
	httpClient     *http.Client
	log            *slog.Logger
}

// NewOIDCClient creates a new OIDC client
func NewOIDCClient(issuerURL, clientID, callbackSecret string, log *slog.Logger) *OIDCClient {
	return &OIDCClient{
		issuerURL:      strings.TrimSuffix(issuerURL, "/"),
		clientID:       clientID,
		callbackSecret: callbackSecret,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		log:            log,
	}
}

// BuildAuthorizationURL constructs the authorization URL for the OAuth2 flow
func (c *OIDCClient) BuildAuthorizationURL(clientID, redirectURI, state, codeChallenge, scope string) string {
	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {scope},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	return c.issuerURL + "/authorize?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
func (c *OIDCClient) ExchangeCode(ctx context.Context, code, redirectURI, codeVerifier, clientID string) (*TokenResponse, error) {
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
func (c *OIDCClient) RefreshToken(ctx context.Context, refreshToken, clientID string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}

	return c.tokenRequest(ctx, data)
}

// RevokeToken revokes an access or refresh token
func (c *OIDCClient) RevokeToken(ctx context.Context, token, clientID string) error {
	data := url.Values{
		"token":     {token},
		"client_id": {clientID},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.issuerURL+"/oauth/revoke",
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revocation failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.log.Debug("token revoked successfully")
	return nil
}

// BuildLoginCallbackURL constructs an HMAC-signed callback URL for the OIDC login provider flow.
// Signature: hex(HMAC-SHA256(auth_request_id + "|" + subject + "|" + claims_b64, callback_secret))
func (c *OIDCClient) BuildLoginCallbackURL(authRequestID, subject string, claims map[string]interface{}) (string, error) {
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	claimsB64 := base64.URLEncoding.EncodeToString(claimsJSON)

	// Build signature: HMAC-SHA256(auth_request_id|subject|claims_b64, secret)
	message := authRequestID + "|" + subject + "|" + claimsB64
	mac := hmac.New(sha256.New, []byte(c.callbackSecret))
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	params := url.Values{
		"auth_request_id": {authRequestID},
		"subject":         {subject},
		"claims":          {claimsB64},
		"signature":       {signature},
	}

	callbackURL := c.issuerURL + "/login/callback?" + params.Encode()

	c.log.Debug("built login callback URL",
		"auth_request_id", authRequestID,
		"subject", subject,
	)

	return callbackURL, nil
}

// tokenRequest makes a POST request to the token endpoint
func (c *OIDCClient) tokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.issuerURL+"/oauth/token",
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
