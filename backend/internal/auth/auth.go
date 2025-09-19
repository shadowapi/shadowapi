package auth

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
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
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

// SessionInfo represents validated session information
type SessionInfo struct {
	SessionID string
	UserID    string
	Email     string
	Factors   map[string]interface{}
}

type Auth struct {
	log              *slog.Logger
	IgnoreHttpsError bool
	cfg              *config.Config
	bearerSecret     string
	sessions         map[string]string
	mu               sync.RWMutex

	// JWT validation
	httpClient *http.Client
	jwksCache  *JWKSCache
	jwksMutex  sync.RWMutex
	oidcConfig *OIDCConfiguration
	oidcExpiry time.Time
}

// Provide returns the authenticator instance
func Provide(i do.Injector) (*Auth, error) {
	cfg := do.MustInvoke[*config.Config](i)
	// keep log case of debugging ogen
	return &Auth{
		log:              do.MustInvoke[*slog.Logger](i),
		IgnoreHttpsError: cfg.Auth.IgnoreHttpsError,
		cfg:              cfg,
		bearerSecret:     cfg.Auth.BearerToken,
		sessions:         make(map[string]string),
		httpClient:       &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// HandleBearerAuth checks the Bearer token
func (a *Auth) HandleBearerAuth(
	ctx context.Context,
	op api.OperationName,
	t api.BearerAuth,
) (context.Context, error) {
	token := t.GetToken()

	// Check legacy bearer secret first
	if a.bearerSecret != "" && token == a.bearerSecret {
		a.log.Debug("legacy bearer token authenticated")
		return ctx, nil
	}

	// Validate JWT token from Zitadel
	if a.cfg.Auth.UserManager == "zitadel" && a.cfg.Auth.Zitadel.InstanceURL != "" {
		jwtToken, err := a.validateJWT(ctx, token)
		if err != nil {
			a.log.Debug("JWT validation failed", "error", err)
			return ctx, &ogenerrors.SecurityError{
				OperationContext: ogenerrors.OperationContext{
					Name: string(op),
					ID:   string(op),
				},
				Security: "BearerAuth",
				Err:      fmt.Errorf("invalid JWT token: %w", err),
			}
		}

		scopeClaim, _ := jwtToken.PrivateClaims()["scope"].(string)
		a.log.Debug("JWT authenticated", "sub", jwtToken.Subject(), "scope", scopeClaim)

		// TODO: Add user context to request
		// You can store user info in request context here
		return ctx, nil
	}

	return ctx, &ogenerrors.SecurityError{
		OperationContext: ogenerrors.OperationContext{
			Name: string(op),
			ID:   string(op),
		},
		Security: "BearerAuth",
		Err:      fmt.Errorf("authentication failed"),
	}
}

// getOIDCConfiguration fetches OpenID Connect configuration from Zitadel
func (a *Auth) getOIDCConfiguration(ctx context.Context) (*OIDCConfiguration, error) {
	a.jwksMutex.RLock()
	if a.oidcConfig != nil && time.Now().Before(a.oidcExpiry) {
		config := a.oidcConfig
		a.jwksMutex.RUnlock()
		return config, nil
	}
	a.jwksMutex.RUnlock()

	a.jwksMutex.Lock()
	defer a.jwksMutex.Unlock()

	// Double check
	if a.oidcConfig != nil && time.Now().Before(a.oidcExpiry) {
		return a.oidcConfig, nil
	}

	if a.cfg.Auth.Zitadel.InstanceURL == "" {
		return nil, fmt.Errorf("not configured URL for Zitadel instance")
	}

	// Use external Zitadel URL through Traefik
	zitadelBaseURL := strings.TrimSuffix(a.cfg.Auth.Zitadel.InstanceURL, "/")
	oidcURL := zitadelBaseURL + "/.well-known/openid-configuration"
	a.log.Debug("fetching OIDC configuration", "url", oidcURL)

	req, err := http.NewRequestWithContext(ctx, "GET", oidcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC config request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
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

	a.oidcConfig = &config
	a.oidcExpiry = time.Now().Add(1 * time.Hour) // Cache for 1 hour

	a.log.Debug("cached OIDC configuration", "issuer", config.Issuer, "jwks_uri", config.JWKSURI)
	return &config, nil
}

// fetchJWKS fetches and caches JWKS from Zitadel
func (a *Auth) fetchJWKS(ctx context.Context, jwksURI string) error {
	a.log.Debug("fetching JWKS", "uri", jwksURI)

	set, err := jwk.Fetch(ctx, jwksURI)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	a.jwksCache = &JWKSCache{
		Set:    set,
		Expiry: time.Now().Add(1 * time.Hour), // Cache for 1 hour
	}

	a.log.Debug("cached JWKS", "keys_count", set.Len())
	return nil
}

// getJWKS returns cached JWKS or fetches fresh if needed
func (a *Auth) getJWKS(ctx context.Context) (jwk.Set, error) {
	a.jwksMutex.RLock()
	if a.jwksCache != nil && time.Now().Before(a.jwksCache.Expiry) {
		set := a.jwksCache.Set
		a.jwksMutex.RUnlock()
		return set, nil
	}
	a.jwksMutex.RUnlock()

	// Need to refresh JWKS
	oidcConfig, err := a.getOIDCConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC config: %w", err)
	}

	a.jwksMutex.Lock()
	defer a.jwksMutex.Unlock()

	// Double check after acquiring lock
	if a.jwksCache != nil && time.Now().Before(a.jwksCache.Expiry) {
		return a.jwksCache.Set, nil
	}

	// Fetch fresh JWKS
	if err := a.fetchJWKS(ctx, oidcConfig.JWKSURI); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return a.jwksCache.Set, nil
}

// ValidateSessionToken validates a session token using Zitadel Session API v2
func (a *Auth) validateSessionToken(ctx context.Context, sessionToken string) (*SessionInfo, error) {
	if a.cfg.Auth.Zitadel.InstanceURL == "" {
		return nil, fmt.Errorf("Zitadel instance URL not configured")
	}

	// For now, we'll use OIDC userinfo endpoint to validate the session token
	// This approach uses the userinfo endpoint which should work with session tokens
	// Use external Zitadel URL through Traefik
	zitadelURL := strings.TrimSuffix(a.cfg.Auth.Zitadel.InstanceURL, "/")
	userinfoURL := zitadelURL + "/oidc/v1/userinfo"
	req, err := http.NewRequestWithContext(ctx, "GET", userinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session via userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("session validation failed with status %d", resp.StatusCode)
	}

	var userinfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	a.log.Debug("session validation successful via userinfo",
		"user_id", userinfo.Sub,
		"email", userinfo.Email,
		"name", userinfo.Name)

	return &SessionInfo{
		SessionID: "", // We don't have session ID from userinfo endpoint
		UserID:    userinfo.Sub,
		Email:     userinfo.Email,
		Factors:   map[string]interface{}{"validated": true},
	}, nil
}

// ValidateJWT validates a JWT token from Zitadel (public method for auth handlers)
func (a *Auth) ValidateJWT(ctx context.Context, tokenString string) (jwt.Token, error) {
	return a.validateJWT(ctx, tokenString)
}

// validateJWT validates a JWT token from Zitadel
func (a *Auth) validateJWT(ctx context.Context, tokenString string) (jwt.Token, error) {
	// Get JWKS for verification
	jwksSet, err := a.getJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	// Parse and verify the token
	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(jwksSet))
	if err != nil {
		return nil, fmt.Errorf("failed to parse and verify JWT: %w", err)
	}

	// Get OIDC configuration for validation
	oidcConfig, err := a.getOIDCConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC config: %w", err)
	}

	// Validate issuer
	if token.Issuer() != oidcConfig.Issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", oidcConfig.Issuer, token.Issuer())
	}

	// Validate audience (if configured)
	if a.cfg.Auth.Zitadel.Audience != "" {
		if !slices.Contains(token.Audience(), a.cfg.Auth.Zitadel.Audience) {
			return nil, fmt.Errorf("invalid audience: expected %s", a.cfg.Auth.Zitadel.Audience)
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
			a.log.Debug("JWT validated successfully", "sub", token.Subject(), "scope", scopeStr)
		} else {
			a.log.Warn("scope claim is not a string")
		}
	}

	return token, nil
}

// OgenMiddleware satisfies Ogen's middleware.Middleware signature.
func (a *Auth) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
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
			a.mu.RLock()
			if userID, exists := a.sessions[cookie.Value]; exists {
				a.mu.RUnlock()
				a.log.Debug("session authenticated", "user_id", userID)
				return next(req)
			}
			a.mu.RUnlock()
		}
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authorization required"))
	}

	// Check for Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid authorization header format"))
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Check legacy bearer secret first
	if a.bearerSecret != "" && token == a.bearerSecret {
		a.log.Debug("legacy bearer token authenticated")
		return next(req)
	}

	// Validate tokens from Zitadel
	if a.cfg.Auth.UserManager == "zitadel" && a.cfg.Auth.Zitadel.InstanceURL != "" {
		tokenPrefix := token
		if len(token) > 10 {
			tokenPrefix = token[:10]
		}
		a.log.Debug("Starting Zitadel authentication", "token_prefix", tokenPrefix)
		// Try Session Token validation first (v2 Sessions API)
		if sessionInfo, err := a.validateSessionToken(req.Raw.Context(), token); err == nil {
			a.log.Debug("Session token authenticated", "session_id", sessionInfo.SessionID, "user_id", sessionInfo.UserID, "email", sessionInfo.Email)
			return next(req)
		} else {
			a.log.Debug("Session token validation failed", "error", err)
		}

		// Fallback to JWT token validation
		jwtToken, err := a.validateJWT(req.Raw.Context(), token)
		if err != nil {
			a.log.Debug("Both session and JWT validation failed", "error", err)
			return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid token: %w", err))
		}

		scopeClaim, _ := jwtToken.PrivateClaims()["scope"].(string)
		a.log.Debug("JWT authenticated", "sub", jwtToken.Subject(), "scope", scopeClaim)

		// TODO: Add user context to request
		// You can store user info in request context here
		return next(req)
	}

	return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authentication failed"))
}

// errWithCode wraps an error with an associated HTTP status code.
type errWithCode struct {
	err    error
	status int
}

func (e *errWithCode) Error() string { return e.err.Error() }

// StatusCode returns the associated HTTP status code.
func (e *errWithCode) StatusCode() int { return e.status }

// ErrWithCode creates a new error annotated with an HTTP status code.
func ErrWithCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{err: err, status: code}
}
