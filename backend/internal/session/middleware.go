package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// JWKSCache represents cached JWKS
type JWKSCache struct {
	Set    jwk.Set
	Expiry time.Time
}

// OIDCConfiguration represents OpenID Connect configuration
type OIDCConfiguration struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

// Middleware implements a pure Ogen middleware that checks for JWT tokens from Zitadel
type Middleware struct {
	log          *slog.Logger
	bearerSecret string
	sessions     map[string]string
	mu           sync.RWMutex

	// JWT validation
	cfg        *config.Config
	httpClient *http.Client
	jwksCache  *JWKSCache
	jwksMutex  sync.RWMutex
	oidcConfig *OIDCConfiguration
	oidcExpiry time.Time
}

// Provide session middleware instance for the dependency injector.
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return &Middleware{
		log:          do.MustInvoke[*slog.Logger](i),
		bearerSecret: cfg.Auth.BearerToken,
		sessions:     make(map[string]string),
		cfg:          cfg,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// getOIDCConfiguration fetches OpenID Connect configuration from Zitadel
func (m *Middleware) getOIDCConfiguration(ctx context.Context) (*OIDCConfiguration, error) {
	m.jwksMutex.RLock()
	if m.oidcConfig != nil && time.Now().Before(m.oidcExpiry) {
		config := m.oidcConfig
		m.jwksMutex.RUnlock()
		return config, nil
	}
	m.jwksMutex.RUnlock()

	m.jwksMutex.Lock()
	defer m.jwksMutex.Unlock()

	// Double check
	if m.oidcConfig != nil && time.Now().Before(m.oidcExpiry) {
		return m.oidcConfig, nil
	}

	if m.cfg.Auth.Zitadel.InstanceURL == "" {
		return nil, fmt.Errorf("not configured URL for Zitadel instance")
	}

	oidcURL := strings.TrimSuffix(m.cfg.Auth.Zitadel.InstanceURL, "/") + "/.well-known/openid-configuration"
	m.log.Debug("fetching OIDC configuration", "url", oidcURL)

	req, err := http.NewRequestWithContext(ctx, "GET", oidcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC config request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC config request failed with status %d", resp.StatusCode)
	}

	var config OIDCConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode OIDC config: %w", err)
	}

	m.oidcConfig = &config
	m.oidcExpiry = time.Now().Add(1 * time.Hour) // Cache for 1 hour

	m.log.Debug("cached OIDC configuration", "issuer", config.Issuer, "jwks_uri", config.JWKSURI)
	return &config, nil
}

// fetchJWKS fetches and caches JWKS from Zitadel
func (m *Middleware) fetchJWKS(ctx context.Context, jwksURI string) error {
	m.log.Debug("fetching JWKS", "uri", jwksURI)

	set, err := jwk.Fetch(ctx, jwksURI)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	m.jwksCache = &JWKSCache{
		Set:    set,
		Expiry: time.Now().Add(1 * time.Hour), // Cache for 1 hour
	}

	m.log.Debug("cached JWKS", "keys_count", set.Len())
	return nil
}

// getJWKS returns cached JWKS or fetches fresh if needed
func (m *Middleware) getJWKS(ctx context.Context) (jwk.Set, error) {
	m.jwksMutex.RLock()
	if m.jwksCache != nil && time.Now().Before(m.jwksCache.Expiry) {
		set := m.jwksCache.Set
		m.jwksMutex.RUnlock()
		return set, nil
	}
	m.jwksMutex.RUnlock()

	// Need to refresh JWKS
	oidcConfig, err := m.getOIDCConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC config: %w", err)
	}

	m.jwksMutex.Lock()
	defer m.jwksMutex.Unlock()

	// Double check after acquiring lock
	if m.jwksCache != nil && time.Now().Before(m.jwksCache.Expiry) {
		return m.jwksCache.Set, nil
	}

	// Fetch fresh JWKS
	if err := m.fetchJWKS(ctx, oidcConfig.JWKSURI); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return m.jwksCache.Set, nil
}

// ValidateJWT validates a JWT token from Zitadel (public method for auth handlers)
func (m *Middleware) ValidateJWT(ctx context.Context, tokenString string) (jwt.Token, error) {
	return m.validateJWT(ctx, tokenString)
}

// validateJWT validates a JWT token from Zitadel
func (m *Middleware) validateJWT(ctx context.Context, tokenString string) (jwt.Token, error) {
	// Get JWKS for verification
	jwksSet, err := m.getJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	// Parse and verify the token
	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(jwksSet))
	if err != nil {
		return nil, fmt.Errorf("failed to parse and verify JWT: %w", err)
	}

	// Get OIDC configuration for validation
	oidcConfig, err := m.getOIDCConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC config: %w", err)
	}

	// Validate issuer
	if token.Issuer() != oidcConfig.Issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", oidcConfig.Issuer, token.Issuer())
	}

	// Validate audience (if configured)
	if m.cfg.Auth.Zitadel.Audience != "" {
		if !slices.Contains(token.Audience(), m.cfg.Auth.Zitadel.Audience) {
			return nil, fmt.Errorf("invalid audience: expected %s", m.cfg.Auth.Zitadel.Audience)
		}
	}

	// Validate required scopes
	if scopeClaim, exists := token.PrivateClaims()["scope"]; exists {
		if scopeStr, ok := scopeClaim.(string); ok {
			scopes := strings.Fields(scopeStr)
			requiredScopes := []string{"openid"}
			for _, required := range requiredScopes {
				if !slices.Contains(scopes, required) {
					return nil, fmt.Errorf("missing required scope: %s", required)
				}
			}
			m.log.Debug("JWT validated successfully", "sub", token.Subject(), "scope", scopeStr)
		} else {
			m.log.Warn("scope claim is not a string")
		}
	}

	return token, nil
}

// OgenMiddleware satisfies Ogen's middleware.Middleware signature.
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	// Skip auth for health check and public endpoints
	path := req.Raw.URL.Path
	if path == "/health" || path == "/" || strings.HasPrefix(path, "/assets/") {
		return next(req)
	}

	// Extract Authorization header
	authHeader := req.Raw.Header.Get("Authorization")
	if authHeader == "" {
		// Check for session cookie as fallback
		if cookie, err := req.Raw.Cookie("sa_session"); err == nil {
			m.mu.RLock()
			if userID, exists := m.sessions[cookie.Value]; exists {
				m.mu.RUnlock()
				m.log.Debug("session authenticated", "user_id", userID)
				return next(req)
			}
			m.mu.RUnlock()
		}
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authorization required"))
	}

	// Check for Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid authorization header format"))
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Check legacy bearer secret first
	if m.bearerSecret != "" && token == m.bearerSecret {
		m.log.Debug("legacy bearer token authenticated")
		return next(req)
	}

	// Validate JWT token from Zitadel
	if m.cfg.Auth.UserManager == "zitadel" && m.cfg.Auth.Zitadel.InstanceURL != "" {
		jwtToken, err := m.validateJWT(req.Raw.Context(), token)
		if err != nil {
			m.log.Debug("JWT validation failed", "error", err)
			return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid JWT token: %w", err))
		}

		scopeClaim, _ := jwtToken.PrivateClaims()["scope"].(string)
		m.log.Debug("JWT authenticated", "sub", jwtToken.Subject(), "scope", scopeClaim)

		// TODO: Add user context to request
		// You can store user info in request context here
		return next(req)
	}

	return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authentication failed"))
}

// AddSession adds a session for local authentication
func (m *Middleware) AddSession(sessionID, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[sessionID] = userID
	m.log.Debug("session added", "session_id", sessionID, "user_id", userID)
}

// RemoveSession removes a session
func (m *Middleware) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	m.log.Debug("session removed", "session_id", sessionID)
}
