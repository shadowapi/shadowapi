package session

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/internal/zitadel"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Middleware implements a pure Ogen middleware that checks for
// either a valid Bearer token or a valid Zitadel session.
type Middleware struct {
	log          *slog.Logger
	bearerSecret string
	zitadel      *zitadel.Client
	handler      *handler.Handler
}

// Provide session middleware instance for the dependency injector
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return &Middleware{
		log:          do.MustInvoke[*slog.Logger](i),
		bearerSecret: cfg.Auth.BearerToken,
		zitadel:      zitadel.Provide(cfg),
		handler:      do.MustInvoke[*handler.Handler](i),
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
		newCtx := WithIdentity(req.Context, Identity{ID: "0"})
		req.SetContext(newCtx)

		return next(req)
	}

	// Check for Zitadel token either in header or cookie
	if token := m.zitadelToken(r); token != "" {
		info, err := m.zitadel.Introspect(req.Context, token)
		if err == nil && info.Active {
			q := query.New(m.handler.DB())
			user, err := q.GetUserByZitadelSubject(req.Context, pgtype.Text{String: info.Subject, Valid: true})
			if err == nil {
				newCtx := WithIdentity(req.Context, Identity{ID: user.UUID.String()})
				req.SetContext(newCtx)
				return next(req)
			}
		}
	}

	if cookie, err := r.Cookie("sa_session"); err == nil {
		m.handler.SessionsMu.Lock()
		uid, ok := m.handler.Sessions[cookie.Value]
		m.handler.SessionsMu.Unlock()
		if ok {
			newCtx := WithIdentity(req.Context, Identity{ID: uid})
			req.SetContext(newCtx)
			return next(req)
		}
	}

	return middleware.Response{}, errors.New("unauthorized")
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

// zitadelToken extracts bearer token or cookie that may contain a Zitadel token
func (m *Middleware) zitadelToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if parts[1] != m.bearerSecret {
				return parts[1]
			}
		}
	}
	if cookie, err := r.Cookie("zitadel_access_token"); err == nil {
		return cookie.Value
	}
	return ""
}
