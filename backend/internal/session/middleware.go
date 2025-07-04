package session

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Middleware implements a pure Ogen middleware that checks for
// either a valid Bearer token or a valid ZITADEL session.
type Middleware struct {
	httpClient    *http.Client
	log           *slog.Logger
	bearerSecret  string
	introspectURL string
	clientID      string
	clientSecret  string
	cookieName    string
}

// Provide session middleware instance for the dependency injector
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)

	return &Middleware{
		httpClient:    http.DefaultClient,
		log:           do.MustInvoke[*slog.Logger](i),
		bearerSecret:  cfg.Auth.BearerToken,
		introspectURL: cfg.Auth.Zitadel.IntrospectURL,
		clientID:      cfg.Auth.Zitadel.ClientID,
		clientSecret:  cfg.Auth.Zitadel.ClientSecret,
		cookieName:    cfg.Auth.Zitadel.CookieName,
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

	// 2) Fallback to session validation
	subject, err := m.validateSession(req)
	if err != nil {
		m.log.Debug("session validation failed", "error", err)
		return middleware.Response{}, errors.New("session validation failed")
	}
	if subject == "" {
		m.log.Debug("empty subject from session")
		return middleware.Response{}, errors.New("invalid session")
	}

	// 3) Attach identity to context
	newCtx := WithIdentity(req.Context, Identity{ID: subject})
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

// validateSession ensures we have a valid ZITADEL session cookie and returns the subject
func (m *Middleware) validateSession(req middleware.Request) (string, error) {
	cookie, err := req.Raw.Cookie(m.cookieName)
	if err != nil {
		m.log.Debug("error getting cookie", "error", err)
		return "", err
	}
	if cookie == nil {
		return "", errors.New("no session found in cookie")
	}

	data := url.Values{}
	data.Set("token", cookie.Value)

	r, err := http.NewRequestWithContext(req.Context, http.MethodPost, m.introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if m.clientID != "" || m.clientSecret != "" {
		r.SetBasicAuth(m.clientID, m.clientSecret)
	}

	resp, err := m.httpClient.Do(r)
	if err != nil {
		m.log.Debug("error validating session", "error", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("introspection failed")
	}
	var out struct {
		Active  bool   `json:"active"`
		Subject string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if !out.Active {
		return "", errors.New("inactive session")
	}
	return out.Subject, nil
}
