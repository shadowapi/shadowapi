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

	gofrsUUID "github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/tenant"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
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
	TenantUUID   string // Tenant from original request context
	TenantName   string // Tenant from original request context
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

	// Get tenant from context if present (this runs on tenant subdomain)
	var tenantUUID, tenantName string
	if t, ok := tenant.FromContext(ctx); ok {
		tenantUUID = t.UUID
		tenantName = t.Name
		h.log.Debug("capturing tenant in OAuth2 state", "tenant_uuid", tenantUUID, "tenant_name", tenantName)
	}

	// Store state, verifier, and tenant
	h.oauth2Svc.stateMu.Lock()
	h.oauth2Svc.stateStore[state] = &oauth2AuthState{
		CodeVerifier: verifier,
		RedirectURI:  req.RedirectURI,
		TenantUUID:   tenantUUID,
		TenantName:   tenantName,
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

	// Get or generate session ID for cross-subdomain session tracking
	sessionID, _ := ctx.Value(oauth2.SharedSessionContextKey).(string)
	if sessionID == "" {
		sessionID = gofrsUUID.Must(gofrsUUID.NewV7()).String()
		h.log.Debug("generated new shared session ID", "session_id", sessionID[:8]+"...")
	} else {
		h.log.Debug("using existing shared session ID", "session_id", sessionID[:8]+"...")
	}

	// Build cookie headers (separate headers for each cookie)
	accessTTL := time.Duration(tokenResp.ExpiresIn) * time.Second
	refreshTTL := 720 * time.Hour // Match Hydra's refresh token TTL

	cookieHeaders := buildCookieHeaders(
		h.oauth2Svc.cookieConfig,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		sessionID,
		accessTTL,
		refreshTTL,
	)

	// Create tenant session record for cross-subdomain session tracking
	if stateData.TenantUUID != "" && h.oauth2Svc.jwtValidator != nil {
		// Decode access token to get user UUID (subject claim)
		claims, err := h.oauth2Svc.jwtValidator.Validate(ctx, tokenResp.AccessToken)
		if err != nil {
			h.log.Warn("failed to decode access token for tenant session", "error", err)
		} else if claims.Subject != "" {
			// Parse UUIDs
			tenantUUID, err := gofrsUUID.FromString(stateData.TenantUUID)
			if err != nil {
				h.log.Warn("invalid tenant UUID in state", "tenant_uuid", stateData.TenantUUID, "error", err)
			} else {
				userUUID, err := gofrsUUID.FromString(claims.Subject)
				if err != nil {
					h.log.Warn("invalid user UUID in token subject", "subject", claims.Subject, "error", err)
				} else {
					// Create or update tenant session record
					sessionRecordUUID := gofrsUUID.Must(gofrsUUID.NewV7())
					expiresAt := time.Now().Add(refreshTTL)

					_, err = query.New(h.dbp).UpsertTenantSession(ctx, query.UpsertTenantSessionParams{
						UUID:       pgtype.UUID{Bytes: sessionRecordUUID, Valid: true},
						SessionID:  sessionID,
						TenantUuid: pgtype.UUID{Bytes: tenantUUID, Valid: true},
						UserUUID:   pgtype.UUID{Bytes: userUUID, Valid: true},
						ExpiresAt:  pgtype.Timestamptz{Time: expiresAt, Valid: true},
					})
					if err != nil {
						h.log.Warn("failed to create tenant session", "error", err)
					} else {
						h.log.Debug("created tenant session",
							"session_id", sessionID[:8]+"...",
							"tenant", stateData.TenantName,
							"user", claims.Subject[:8]+"...",
						)
					}
				}
			}
		}
	}

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
	// Don't include session ID on refresh - the shared session cookie is already set
	accessTTL := time.Duration(tokenResp.ExpiresIn) * time.Second
	refreshTTL := 720 * time.Hour

	cookieHeaders := buildCookieHeaders(
		h.oauth2Svc.cookieConfig,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		"", // No session ID on refresh
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
// For tenant subdomains, validates that the token belongs to the current tenant.
func (h *Handler) AuthOAuth2Session(ctx context.Context) (*api.AuthOAuth2SessionOK, error) {
	if h.oauth2Svc == nil {
		return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
	}

	// Get access token from context (set by middleware)
	accessToken, _ := ctx.Value(auth.AccessTokenContextKey).(string)

	// If we have a JWT validator and an access token, validate tenant
	if h.oauth2Svc.jwtValidator != nil && accessToken != "" {
		claims, err := h.oauth2Svc.jwtValidator.Validate(ctx, accessToken)
		if err != nil {
			h.log.Debug("session check: JWT validation failed", "error", err)
			return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
		}

		// Get current tenant from context (set by tenant middleware)
		currentTenant, hasTenant := tenant.FromContext(ctx)

		// If we're on a tenant subdomain, verify the token is for this tenant
		if hasTenant && claims.TenantUUID() != "" {
			if claims.TenantUUID() != currentTenant.UUID {
				h.log.Debug("session check: tenant mismatch",
					"token_tenant", claims.TenantUUID(),
					"current_tenant", currentTenant.UUID,
				)
				return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
			}
		}

		// Token is valid and tenant matches (or no tenant context)
		expiresIn := 3600 // Default
		if claims.ExpiresAt != nil {
			expiresIn = int(time.Until(claims.ExpiresAt.Time).Seconds())
		}

		return &api.AuthOAuth2SessionOK{
			Authenticated: true,
			ExpiresIn:     api.NewOptInt(expiresIn),
		}, nil
	}

	// Fallback: check refresh token (for backwards compatibility or when JWT validator not configured)
	refreshToken, ok := ctx.Value(auth.RefreshTokenContextKey).(string)
	if !ok || refreshToken == "" {
		return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
	}

	// We have a refresh token but couldn't validate tenant via JWT
	// For security, if we're on a tenant subdomain without JWT validation, require re-login
	if _, hasTenant := tenant.FromContext(ctx); hasTenant {
		h.log.Debug("session check: on tenant subdomain but cannot validate token tenant")
		return &api.AuthOAuth2SessionOK{Authenticated: false}, nil
	}

	// Root domain: having a refresh token means authenticated
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

func buildCookieHeaders(cfg oauth2.CookieConfig, accessToken, refreshToken, sessionID string, accessTTL, refreshTTL time.Duration) []string {
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

	cookies := []string{accessCookie, refreshCookie}

	// Add shared session cookie if sessionID is provided
	if sessionID != "" {
		// Use leading dot for domain to include all subdomains
		sharedDomain := cfg.Domain
		if sharedDomain != "" && sharedDomain[0] != '.' {
			sharedDomain = "." + sharedDomain
		}

		sharedSessionCookie := fmt.Sprintf("%s=%s; Path=/; Domain=%s; Max-Age=%d; HttpOnly; SameSite=Lax%s",
			oauth2.SharedSessionCookie,
			sessionID,
			sharedDomain,
			int(refreshTTL.Seconds()),
			secure,
		)
		cookies = append(cookies, sharedSessionCookie)
	}

	return cookies
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
