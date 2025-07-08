package session

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/zitadel"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Middleware implements a pure Ogen middleware that checks for
// either a valid Bearer token or a valid Zitadel session.
type Middleware struct {
	log          *slog.Logger
	bearerSecret string
	zitadel      *zitadel.Client
	db           *pgxpool.Pool
	sessions     map[string]string
	sessionsMu   sync.RWMutex
}

// Provide session middleware instance for the dependency injector
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return &Middleware{
		log:          do.MustInvoke[*slog.Logger](i),
		bearerSecret: cfg.Auth.BearerToken,
		zitadel:      zitadel.Provide(cfg),
		db:           do.MustInvoke[*pgxpool.Pool](i),
		sessions:     make(map[string]string),
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

	// 2) Local session cookie
	if cookie, err := r.Cookie("sa_session"); err == nil {
		m.sessionsMu.RLock()
		if uid, ok := m.sessions[cookie.Value]; ok {
			m.sessionsMu.RUnlock()
			req.SetContext(WithIdentity(req.Context, Identity{ID: uid}))
			return next(req)
		}
		m.sessionsMu.RUnlock()
	}

	// 3) Validate Zitadel token if present
	if token := m.zitadelToken(r); token != "" {
		if info, err := m.zitadel.Introspect(req.Context, token); err == nil && info.Active {
			q := query.New(m.db)
			if user, err := q.GetUserByZitadelSubject(req.Context, pgtype.Text{String: info.Subject, Valid: true}); err == nil {
				req.SetContext(WithIdentity(req.Context, Identity{ID: user.UUID.String()}))
				return next(req)
			}
		}
	}

	if req.OperationID == "session-status" {
		// Unauthenticated session checks should still proceed
		return next(req)
	}

	return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, errors.New("unauthorized"))
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

// AddSession registers a new session token for the given user ID.
func (m *Middleware) AddSession(token, uid string) {
	m.sessionsMu.Lock()
	m.sessions[token] = uid
	m.sessionsMu.Unlock()
}

// DeleteSession removes the given session token.
func (m *Middleware) DeleteSession(token string) {
	m.sessionsMu.Lock()
	delete(m.sessions, token)
	m.sessionsMu.Unlock()
}
