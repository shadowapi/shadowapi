package session

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	ory "github.com/ory/kratos-client-go"
	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Middleware implements a pure Ogen middleware that checks for
// either a valid Bearer token or a valid Ory Kratos session.
type Middleware struct {
	ory          *ory.APIClient
	log          *slog.Logger
	bearerSecret string
}

// Provide session middleware instance for the dependency injector
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	oryCfg := ory.NewConfiguration()
	oryCfg.Servers = []ory.ServerConfiguration{
		{
			URL: cfg.Auth.Ory.KratosUserAPI,
		},
	}

	return &Middleware{
		ory:          ory.NewAPIClient(oryCfg),
		log:          do.MustInvoke[*slog.Logger](i),
		bearerSecret: cfg.Auth.BearerToken,
	}, nil
}

// OgenMiddleware satisfies Ogen's middleware.Middleware signature
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {

	// 'req.Raw' is the original *http.Request
	r := req.Raw

	// 1) Check for Bearer token
	if m.validateBearer(r) {
		m.log.Debug("validated bearer token successfully")
		// Proceed to the next middleware/handler

		// TODO find machine user
		newCtx := WithIdentity(req.Context, Identity{Id: "0"})
		req.SetContext(newCtx)

		return next(req)
	}

	// 2) Fallback to session validation
	session, err := m.validateSession(r)
	if err != nil {
		m.log.Debug("session validation failed", "error", err)
		// Return an error to Ogen (could also return a redirect error if desired)
		return middleware.Response{}, errors.New("session validation failed")
	}
	if session == nil || session.Active == nil || !*session.Active {
		m.log.Debug("session is not active or nil")
		return middleware.Response{}, errors.New("invalid session")
	}

	// 3) Attach identity to context
	identity := session.GetIdentity()
	if identity.Id == "" {
		m.log.Debug("no Identity Id")
		return middleware.Response{}, errors.New("no Identity Id")
	}

	newCtx := WithIdentity(req.Context, Identity{Id: identity.Id})
	req.SetContext(newCtx)

	return next(req)
}

// validateBearer checks if the Authorization header has a valid Bearer token
func (m *Middleware) validateBearer(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}
	return parts[1] == m.bearerSecret
}

// validateSession ensures we have a valid cookie-based Ory Kratos session
func (m *Middleware) validateSession(r *http.Request) (*ory.Session, error) {
	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		m.log.Debug("error getting cookie", "error", err)
		return nil, err
	}
	if cookie == nil {
		return nil, errors.New("no session found in cookie")
	}
	resp, _, err := m.ory.FrontendAPI.ToSession(r.Context()).Cookie(r.Header.Get("Cookie")).Execute()
	if err != nil {
		m.log.Debug("error validating session", "error", err)
		return nil, err
	}
	return resp, nil
}
