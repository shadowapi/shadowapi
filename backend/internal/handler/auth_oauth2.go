package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// OAuth2Service provides OAuth2 functionality for the handlers
type OAuth2Service struct {
	hydraClient  *oauth2.HydraClient
	jwtValidator *oauth2.JWTValidator
	cookieConfig oauth2.CookieConfig
	clientID     string
	redirectURI  string
	baseURL      string

	// In-memory state storage (for production, use Redis or DB)
	stateMu    sync.RWMutex
	stateStore map[string]*oauth2AuthState
}

type oauth2AuthState struct {
	CodeVerifier string
	RedirectURI  string
	CreatedAt    time.Time
}

// NewOAuth2Service creates a new OAuth2 service
func NewOAuth2Service(
	hydraClient *oauth2.HydraClient,
	jwtValidator *oauth2.JWTValidator,
	cookieConfig oauth2.CookieConfig,
	clientID, baseURL string,
) *OAuth2Service {
	svc := &OAuth2Service{
		hydraClient:  hydraClient,
		jwtValidator: jwtValidator,
		cookieConfig: cookieConfig,
		clientID:     clientID,
		redirectURI:  baseURL + "/api/v1/auth/oauth2/callback",
		baseURL:      baseURL,
		stateStore:   make(map[string]*oauth2AuthState),
	}

	// Start cleanup goroutine for expired states
	go svc.cleanupExpiredStates()

	return svc
}

// cleanupExpiredStates removes states older than 10 minutes
func (s *OAuth2Service) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.stateMu.Lock()
		now := time.Now()
		for state, data := range s.stateStore {
			if now.Sub(data.CreatedAt) > 10*time.Minute {
				delete(s.stateStore, state)
			}
		}
		s.stateMu.Unlock()
	}
}

// generateState creates a random state string
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generatePKCE creates a code verifier and challenge
func generatePKCE() (verifier, challenge string, err error) {
	// Generate 32 random bytes for verifier
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)

	// Create S256 challenge
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	return verifier, challenge, nil
}

// AuthOAuth2Authorize initiates the OAuth2 authorization flow
func (h *Handler) AuthOAuth2Authorize(ctx context.Context, req *api.AuthOAuth2AuthorizeReq) (*api.AuthOAuth2AuthorizeOK, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Generate state and PKCE
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	verifier, challenge, err := generatePKCE()
	if err != nil {
		return nil, fmt.Errorf("generate PKCE: %w", err)
	}

	// Store state and verifier
	h.oauth2Svc.stateMu.Lock()
	h.oauth2Svc.stateStore[state] = &oauth2AuthState{
		CodeVerifier: verifier,
		RedirectURI:  req.RedirectURI,
		CreatedAt:    time.Now(),
	}
	h.oauth2Svc.stateMu.Unlock()

	// Build authorization URL
	authURL := h.oauth2Svc.hydraClient.BuildAuthorizationURL(
		h.oauth2Svc.clientID,
		h.oauth2Svc.redirectURI,
		state,
		challenge,
		"openid offline_access profile email",
	)

	h.log.Debug("OAuth2 authorization initiated",
		"state", state[:8]+"...",
		"redirect_uri", req.RedirectURI,
	)

	return &api.AuthOAuth2AuthorizeOK{
		AuthorizationURL: authURL,
		State:            state,
	}, nil
}

// AuthOAuth2Callback handles the OAuth2 callback from Hydra
func (h *Handler) AuthOAuth2Callback(ctx context.Context, params api.AuthOAuth2CallbackParams) (*api.AuthOAuth2CallbackFound, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Validate and retrieve state
	h.oauth2Svc.stateMu.Lock()
	stateData, ok := h.oauth2Svc.stateStore[params.State]
	if ok {
		delete(h.oauth2Svc.stateStore, params.State)
	}
	h.oauth2Svc.stateMu.Unlock()

	if !ok {
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("invalid or expired state"))
	}

	// Check state age
	if time.Since(stateData.CreatedAt) > 10*time.Minute {
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("state expired"))
	}

	// Exchange code for tokens
	tokenResp, err := h.oauth2Svc.hydraClient.ExchangeCode(
		ctx,
		params.Code,
		h.oauth2Svc.redirectURI,
		stateData.CodeVerifier,
		h.oauth2Svc.clientID,
	)
	if err != nil {
		h.log.Error("token exchange failed", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, fmt.Errorf("token exchange failed: %w", err))
	}

	h.log.Debug("token exchange successful",
		"expires_in", tokenResp.ExpiresIn,
		"has_refresh_token", tokenResp.RefreshToken != "",
	)

	// Build cookie headers (separate headers for each cookie)
	accessTTL := time.Duration(tokenResp.ExpiresIn) * time.Second
	refreshTTL := 720 * time.Hour // Match Hydra's refresh token TTL

	cookieHeaders := buildCookieHeaders(
		h.oauth2Svc.cookieConfig,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		accessTTL,
		refreshTTL,
	)

	// Redirect to the frontend
	redirectURL := stateData.RedirectURI
	if redirectURL == "" {
		redirectURL = h.oauth2Svc.baseURL + "/"
	}

	return &api.AuthOAuth2CallbackFound{
		Location:  api.NewOptString(redirectURL),
		SetCookie: cookieHeaders,
	}, nil
}

// AuthOAuth2Refresh refreshes the access token
func (h *Handler) AuthOAuth2Refresh(ctx context.Context) (*api.AuthOAuth2RefreshOKHeaders, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Get refresh token from context (set by middleware)
	refreshToken, ok := ctx.Value(auth.RefreshTokenContextKey).(string)
	if !ok || refreshToken == "" {
		return nil, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("no refresh token"))
	}

	// Exchange refresh token for new tokens
	tokenResp, err := h.oauth2Svc.hydraClient.RefreshToken(ctx, refreshToken, h.oauth2Svc.clientID)
	if err != nil {
		h.log.Error("token refresh failed", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("token refresh failed"))
	}

	// Build cookie headers (separate headers for each cookie)
	accessTTL := time.Duration(tokenResp.ExpiresIn) * time.Second
	refreshTTL := 720 * time.Hour

	cookieHeaders := buildCookieHeaders(
		h.oauth2Svc.cookieConfig,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		accessTTL,
		refreshTTL,
	)

	return &api.AuthOAuth2RefreshOKHeaders{
		SetCookie: cookieHeaders,
		Response: api.AuthOAuth2RefreshOK{
			ExpiresIn: tokenResp.ExpiresIn,
		},
	}, nil
}

// AuthOAuth2Session checks if the user has a valid session without triggering token refresh.
// Always returns 200 to avoid console errors for unauthenticated users.
func (h *Handler) AuthOAuth2Session(ctx context.Context) (*api.AuthOAuth2SessionOK, error) {
	if h.oauth2Svc == nil {
		return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
	}

	// Get refresh token from context (set by middleware)
	refreshToken, ok := ctx.Value(auth.RefreshTokenContextKey).(string)
	if !ok || refreshToken == "" {
		return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
	}

	// We have a refresh token, so the user has a session.
	// Note: We don't validate the token here to avoid the overhead.
	// The actual token refresh will validate it when needed.
	return &api.AuthOAuth2SessionOK{
		Authenticated: true,
		ExpiresIn:     api.NewOptInt(3600), // Approximate TTL
	}, nil
}

// AuthOAuth2Logout logs out the user
func (h *Handler) AuthOAuth2Logout(ctx context.Context) (*api.AuthOAuth2LogoutOKHeaders, error) {
	if h.oauth2Svc == nil {
		return nil, fmt.Errorf("OAuth2 service not configured")
	}

	// Get refresh token from context to revoke it
	refreshToken, _ := ctx.Value(auth.RefreshTokenContextKey).(string)
	if refreshToken != "" {
		if err := h.oauth2Svc.hydraClient.RevokeToken(ctx, refreshToken, h.oauth2Svc.clientID); err != nil {
			h.log.Warn("failed to revoke token", "error", err)
			// Continue with logout even if revocation fails
		}
	}

	// Build cookie headers to clear all cookies (separate headers for each cookie)
	cookieHeaders := buildClearCookieHeaders(h.oauth2Svc.cookieConfig)

	return &api.AuthOAuth2LogoutOKHeaders{
		SetCookie: cookieHeaders,
		Response: api.AuthOAuth2LogoutOK{
			Success: true,
		},
	}, nil
}

// Helper functions for building cookie headers

func buildCookieHeaders(cfg oauth2.CookieConfig, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) []string {
	secure := ""
	if cfg.Secure {
		secure = "; Secure"
	}

	accessCookie := fmt.Sprintf("%s=%s; Path=/api; Domain=%s; Max-Age=%d; HttpOnly; SameSite=Lax%s",
		oauth2.AccessTokenCookie,
		accessToken,
		cfg.Domain,
		int(accessTTL.Seconds()),
		secure,
	)

	refreshCookie := fmt.Sprintf("%s=%s; Path=/api/v1/auth/oauth2; Domain=%s; Max-Age=%d; HttpOnly; SameSite=Lax%s",
		oauth2.RefreshTokenCookie,
		refreshToken,
		cfg.Domain,
		int(refreshTTL.Seconds()),
		secure,
	)

	return []string{accessCookie, refreshCookie}
}

func buildClearCookieHeaders(cfg oauth2.CookieConfig) []string {
	secure := ""
	if cfg.Secure {
		secure = "; Secure"
	}

	accessCookie := fmt.Sprintf("%s=; Path=/api; Domain=%s; Max-Age=0; HttpOnly; SameSite=Lax%s",
		oauth2.AccessTokenCookie,
		cfg.Domain,
		secure,
	)

	refreshCookie := fmt.Sprintf("%s=; Path=/api/v1/auth/oauth2; Domain=%s; Max-Age=0; HttpOnly; SameSite=Lax%s",
		oauth2.RefreshTokenCookie,
		cfg.Domain,
		secure,
	)

	return []string{accessCookie, refreshCookie}
}
