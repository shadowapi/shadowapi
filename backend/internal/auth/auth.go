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
}

// Provide returns the authenticator instance
func Provide(i do.Injector) (*Auth, error) {
	cfg := do.MustInvoke[*config.Config](i)
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

	// TODO: Implement proper token validation
	// For now, accept any non-empty bearer token
	if token != "" {
		a.log.Debug("accepting bearer token", "token_prefix", token[:min(len(token), 20)])
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

	// TODO: Implement proper token validation
	// For now, accept any non-empty bearer token
	if token != "" {
		a.log.Debug("accepting bearer token", "token_prefix", token[:min(len(token), 20)])
		return next(req)
	}

	return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, fmt.Errorf("authentication failed"))
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
