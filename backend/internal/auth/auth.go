package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// Context keys for auth data
type contextKey string

const (
	UserClaimsContextKey   contextKey = "user_claims"
	RefreshTokenContextKey contextKey = "refresh_token"
)

type Auth struct {
	log   *slog.Logger
	cfg   *config.Config
	mutex sync.RWMutex
	allow map[string]string

	httpClient       *http.Client
	bearerSecret     string
	IgnoreHttpsError bool

	// JWT validation
	jwtValidator *oauth2.JWTValidator
}

// Provide returns the authenticator instance
func Provide(i do.Injector) (*Auth, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)

	auth := &Auth{
		log: log,
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		IgnoreHttpsError: cfg.Auth.IgnoreHttpsError,
		allow: map[string]string{
			"/health":                       http.MethodGet,
			"/ready":                        http.MethodGet,
			"/api/v1/user":                  http.MethodPost,
			"/api/v1/user/session":          http.MethodPost,
			"/api/v1/auth/login":            http.MethodGet,
			"/api/v1/auth/callback":         http.MethodGet,
			"/api/v1/auth/logout":           http.MethodPost,
			"/auth/callback":                http.MethodGet,
			"/api/v1/auth/oauth2/authorize": http.MethodPost,
			"/api/v1/auth/oauth2/callback":  http.MethodGet,
			"/api/v1/auth/oauth2/refresh":   http.MethodPost,
			"/api/v1/auth/oauth2/logout":    http.MethodPost,
		},
	}

	// Initialize JWT validator if OAuth2 is configured
	if cfg.OAuth2.SPAClientID != "" {
		jwksURL := cfg.OAuth2.HydraPublicURL + "/.well-known/jwks.json"
		jwksCache := oauth2.NewJWKSCache(
			jwksURL,
			time.Duration(cfg.OAuth2.JWKSCacheDuration)*time.Second,
			log,
		)
		auth.jwtValidator = oauth2.NewJWTValidator(jwksCache, cfg.BaseURL, log)
		log.Info("JWT validator initialized", "jwks_url", jwksURL)
	}

	return auth, nil
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
	if token == "" {
		return nil, errResult
	}

	// If JWT validator is configured, validate the token
	if a.jwtValidator != nil {
		claims, err := a.jwtValidator.Validate(ctx, token)
		if err != nil {
			a.log.Debug("JWT validation failed", "error", err)
			return nil, errResult
		}
		// Add claims to context
		ctx = context.WithValue(ctx, UserClaimsContextKey, claims)
		a.log.Debug("JWT validated successfully", "subject", claims.Subject)
		return ctx, nil
	}

	// Fallback: accept any non-empty bearer token (for development)
	a.log.Debug("accepting bearer token (no JWT validator)", "token_prefix", token[:min(len(token), 20)])
	return ctx, nil
}

func (a *Auth) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	// Check if path is in allow list
	if m, ok := a.allow[req.Raw.URL.Path]; ok && m == req.Raw.Method {
		// For OAuth2 refresh/logout, add refresh token to context
		if strings.HasPrefix(req.Raw.URL.Path, "/api/v1/auth/oauth2/") {
			if cookie, err := req.Raw.Cookie(oauth2.RefreshTokenCookie); err == nil {
				ctx := context.WithValue(req.Context, RefreshTokenContextKey, cookie.Value)
				req.SetContext(ctx)
			}
		}
		return next(req)
	}

	// Try to get token from Authorization header first
	authHeader := req.Raw.Header.Get("Authorization")
	var token string

	if authHeader != "" {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid authorization header format"))
		}
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Try to get token from cookie (for browser requests)
		if cookie, err := req.Raw.Cookie(oauth2.AccessTokenCookie); err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authorization required"))
	}

	// Check bearer secret (for service-to-service auth)
	if a.bearerSecret != "" && token == a.bearerSecret {
		return next(req)
	}

	// Validate JWT token if validator is configured
	if a.jwtValidator != nil {
		claims, err := a.jwtValidator.Validate(req.Context, token)
		if err != nil {
			a.log.Debug("JWT validation failed in middleware", "error", err)
			return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("invalid token"))
		}
		// Add claims to context
		ctx := context.WithValue(req.Context, UserClaimsContextKey, claims)
		req.SetContext(ctx)
		return next(req)
	}

	// Fallback: accept any non-empty bearer token (for development)
	a.log.Debug("accepting bearer token (no JWT validator)", "token_prefix", token[:min(len(token), 20)])
	return next(req)
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
