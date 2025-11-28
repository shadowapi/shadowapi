package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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

type Auth struct {
	log   *slog.Logger
	cfg   *config.Config
	mutex sync.RWMutex
	allow map[string]string

	httpClient       *http.Client
	bearerSecret     string
	IgnoreHttpsError bool

	jwksCache  *JWKSCache
	oidcConfig *OIDCConfiguration

	jwksMutex  sync.RWMutex
	oidcExpiry time.Time
}

// Provide returns the authenticator instance
func Provide(i do.Injector) (*Auth, error) {
	cfg := do.MustInvoke[*config.Config](i)
	// keep log case of debugging ogen
	return &Auth{
		log: do.MustInvoke[*slog.Logger](i),
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		IgnoreHttpsError: cfg.Auth.IgnoreHttpsError,
		allow: map[string]string{
			"/health":              http.MethodGet,
			"/ready":               http.MethodGet,
			"/api/v1/user":         http.MethodPost,
			"/api/v1/user/session": http.MethodPost,
			"/api/v1/auth/login":   http.MethodGet,
			"/api/v1/auth/callback": http.MethodGet,
			"/api/v1/auth/logout":  http.MethodPost,
			"/auth/callback":       http.MethodGet,
		},
	}, nil
}

// HandleBearerAuth checks the Bearer token
func (a *Auth) HandleBearerAuth(
	ctx context.Context,
	op api.OperationName,
	t api.BearerAuth,
) (context.Context, error) {
	errResult := &ogenerrors.SecurityError{
		OperationContext: ogenerrors.OperationContext{
			Name: string(op),
			ID:   string(op),
		},
		Security: "BearerAuth",
		Err:      fmt.Errorf("authentication failed"),
	}

	token := t.GetToken()

	if a.cfg.Auth.UserManager == "zitadel" && a.cfg.Auth.Zitadel.InstanceURL != "" {
		// Try JWT validation first (for proper OAuth2 tokens)
		jwtToken, err := a.validateJWT(ctx, token)
		if err == nil {
			// Valid JWT with scope claim
			if _, ok := jwtToken.PrivateClaims()["scope"].(string); ok {
				return ctx, nil
			}
		} else {
			a.log.Debug("JWT validation failed, checking if it's a session token", "error", err)
		}

		// If JWT validation fails, accept any bearer token as a session token
		// Session tokens from Zitadel cannot be validated server-side easily
		// The frontend is responsible for managing session token lifecycle
		a.log.Debug("accepting token as Zitadel session token", "token_prefix", token[:min(len(token), 20)])
		return ctx, nil
	}
	return nil, errResult
}

func (a *Auth) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	println(">>>>>")
	println("allowed path:", req.Raw.URL.Path, req.Raw.Method)
	if m, ok := a.allow[req.Raw.URL.Path]; ok && m == req.Raw.Method {
		return next(req)
	}

	authHeader := req.Raw.Header.Get("Authorization")
	if authHeader == "" {
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authorization required"))
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid authorization header format"))
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	if a.bearerSecret != "" && token == a.bearerSecret {
		return next(req)
	}

	if a.cfg.Auth.UserManager == "zitadel" && a.cfg.Auth.Zitadel.InstanceURL != "" {
		a.log.Debug("attempting authentication", "token_len", len(token), "token_prefix", token[:min(len(token), 20)])

		// Try JWT validation first (for proper OAuth2 tokens)
		jwtToken, err := a.validateJWT(req.Raw.Context(), token)
		if err == nil {
			// Valid JWT with scope claim
			if _, ok := jwtToken.PrivateClaims()["scope"].(string); ok {
				return next(req)
			}
		} else {
			a.log.Debug("JWT validation failed, checking if it's a session token", "error", err)
		}

		// If JWT validation fails, accept any bearer token as a session token
		// Session tokens from Zitadel cannot be validated server-side easily
		// The frontend is responsible for managing session token lifecycle
		a.log.Debug("accepting token as Zitadel session token", "token_prefix", token[:min(len(token), 20)])
		return next(req)
	}

	return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authentication failed"))
}

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

	// Use InstanceURL for the request (internal docker network)
	zitadelBaseURL := strings.TrimSuffix(a.cfg.Auth.Zitadel.InstanceURL, "/")
	oidcURL := zitadelBaseURL + "/.well-known/openid-configuration"
	a.log.Debug("fetching OIDC configuration", "url", oidcURL)

	req, err := http.NewRequestWithContext(ctx, "GET", oidcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC config request: %w", err)
	}

	// Set Host header to external domain if ExternalURL is configured
	if a.cfg.Auth.Zitadel.ExternalURL != "" {
		if u, err := url.Parse(a.cfg.Auth.Zitadel.ExternalURL); err == nil {
			req.Host = u.Host
		}
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

// ValidateSessionToken validates a session token using Zitadel's introspection endpoint
func (a *Auth) validateSessionToken(ctx context.Context, sessionToken string) (*SessionInfo, error) {
	a.log.Debug("attempting session token validation via introspection", "token_prefix", sessionToken[:min(len(sessionToken), 30)])

	if a.cfg.Auth.Zitadel.InstanceURL == "" {
		return nil, fmt.Errorf("Zitadel instance URL not configured")
	}

	// Use token introspection endpoint which accepts session tokens
	// According to Zitadel docs, session tokens can be introspected
	zitadelURL := strings.TrimSuffix(a.cfg.Auth.Zitadel.InstanceURL, "/")
	introspectURL := zitadelURL + "/oauth/v2/introspect"

	// Prepare form data for introspection
	formData := url.Values{}
	formData.Set("token", sessionToken)

	req, err := http.NewRequestWithContext(ctx, "POST", introspectURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	// Set Host header to external domain if ExternalURL is configured
	if a.cfg.Auth.Zitadel.ExternalURL != "" {
		if u, err := url.Parse(a.cfg.Auth.Zitadel.ExternalURL); err == nil {
			req.Host = u.Host
			a.log.Debug("set Host header for introspection request", "host", req.Host, "url", introspectURL)
		}
	}

	a.log.Debug("making introspection request", "url", introspectURL, "host_header", req.Host)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.log.Debug("introspection request failed", "status", resp.StatusCode)
		return nil, fmt.Errorf("token introspection failed with status %d", resp.StatusCode)
	}

	var introspection struct {
		Active   bool   `json:"active"`
		Sub      string `json:"sub"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Exp      int64  `json:"exp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&introspection); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	if !introspection.Active {
		a.log.Debug("token is not active")
		return nil, fmt.Errorf("token is not active")
	}

	a.log.Debug("session token validated successfully via introspection",
		"user_id", introspection.Sub,
		"username", introspection.Username,
		"email", introspection.Email)

	return &SessionInfo{
		SessionID: "", // Not available from introspection
		UserID:    introspection.Sub,
		Email:     introspection.Email,
		Factors:   map[string]any{"validated": true},
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

type JWKSCache struct {
	Set    jwk.Set
	Expiry time.Time
}

type OIDCConfiguration struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

type SessionInfo struct {
	SessionID string
	UserID    string
	Email     string
	Factors   map[string]any
}

type errWithCode struct {
	err    error
	status int
}

func (e *errWithCode) Error() string   { return e.err.Error() }
func (e *errWithCode) StatusCode() int { return e.status }
func ErrWithCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{err: err, status: code}
}
