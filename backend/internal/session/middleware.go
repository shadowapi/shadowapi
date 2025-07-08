package session

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Middleware implements a pure Ogen middleware that checks for a bearer token
// or a previously established local session.
type Middleware struct {
	log          *slog.Logger
	bearerSecret string
	sessions     map[string]string
	mu           sync.RWMutex
}

// Provide session middleware instance for the dependency injector.
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return &Middleware{
		log:          do.MustInvoke[*slog.Logger](i),
		bearerSecret: cfg.Auth.BearerToken,
		sessions:     make(map[string]string),
	}, nil
}

// OgenMiddleware satisfies Ogen's middleware.Middleware signature.
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	r := req.Raw

	// 1) machine-to-machine bearer
	if m.validBearer(r) {
		m.log.Debug("auth bearer ok")
		req.SetContext(WithIdentity(req.Context, Identity{ID: "0"}))
		return next(req)
	}

	// 2) first-class local session
	if c, err := r.Cookie("sa_session"); err == nil {
		m.mu.RLock()
		uid, ok := m.sessions[c.Value]
		m.mu.RUnlock()
		if ok {
			m.log.Debug("auth session ok", "uid", uid)
			req.SetContext(WithIdentity(req.Context, Identity{ID: uid}))
			return next(req)
		}
		m.log.Debug("session cookie miss", "token", c.Value)
		req.SetContext(context.WithValue(req.Context, "auth_reason", "session cookie miss"))
	}

	// 3) anonymous `/session` probe is allowed so front-end can learn the reason
	if req.OperationID == "session-status" {
		return next(req)
	}

	req.SetContext(context.WithValue(req.Context, "auth_reason", "unauthorized"))
	return middleware.Response{}, ErrWithCode(http.StatusUnauthorized, errors.New("unauthorized"))
}

func (m *Middleware) validBearer(r *http.Request) bool {
	h := r.Header.Get("Authorization")
	if h == "" {
		return false
	}
	p := strings.SplitN(h, " ", 2)
	return len(p) == 2 && strings.EqualFold(p[0], "Bearer") && p[1] == m.bearerSecret
}

func (m *Middleware) AddSession(token, uid string) {
	m.mu.Lock()
	m.sessions[token] = uid
	m.mu.Unlock()
	m.log.Debug("session added", "token", token, "uid", uid)
}

func (m *Middleware) DeleteSession(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
	m.log.Debug("session deleted", "token", token)
}
