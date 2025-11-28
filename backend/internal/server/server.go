package server

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/samber/do/v2"

    "github.com/shadowapi/shadowapi/backend/internal/auth"
    "github.com/shadowapi/shadowapi/backend/internal/config"
    "github.com/shadowapi/shadowapi/backend/internal/handler"
    "github.com/shadowapi/shadowapi/backend/internal/idp"
    "github.com/shadowapi/shadowapi/backend/pkg/api"
    "golang.org/x/oauth2"
)

type Server struct {
    cfg *config.Config
    log *slog.Logger

    api          *api.Server
    listener     net.Listener
    specsHandler http.Handler
    idp          idp.Provider
    handler      *handler.Handler
    auth         *auth.Auth

    // simple in-memory state store for OIDC (login flow)
    oidcStates map[string]time.Time
}

// Provide server instance for the dependency injector
func Provide(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	authService := do.MustInvoke[*auth.Auth](i)
	handlerService := do.MustInvoke[*handler.Handler](i)
    idpProvider := idp.NewProvider(cfg)

	srv, err := api.NewServer(
		handlerService,
		authService,
		api.WithPathPrefix("/api/v1"),
		api.WithMiddleware(authService.OgenMiddleware),
		api.WithNotFound(func(w http.ResponseWriter, r *http.Request) {
			log.Info("no ogen route matched, returning 404")
			http.NotFound(w, r)
		}),
	)
	if err != nil {
		log.Error("failed to create server", "error", err.Error())
		return nil, err
	}

	var specsHandler http.Handler
	if cfg.API.SpecsDir != "" {
		log.Info("serving API specs", "dir", cfg.API.SpecsDir, "url", "/assets/docs/api/openapi.yaml")
		specsHandler = http.StripPrefix("/assets/docs/api", http.FileServer(http.Dir(cfg.API.SpecsDir)))
	}

    return &Server{
        cfg:          cfg,
        log:          do.MustInvoke[*slog.Logger](i),
        api:          srv,
        specsHandler: specsHandler,
        idp:          idpProvider,
        handler:      handlerService,
        auth:         authService,
        oidcStates:   make(map[string]time.Time),
    }, nil
}

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.log.Error("failed to listen", "error", err.Error())
		return err
	}
	s.listener = listener

	return http.Serve(listener, s)
}

// ServeHTTP wraps the API server and also serves the frontend dist (SPA) with index.html fallback
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // OIDC login endpoints handled before ogen router
    if r.URL.Path == "/api/v1/auth/login" {
        s.handleOIDCLogin(w, r)
        return
    }
    if r.URL.Path == "/api/v1/auth/callback" || r.URL.Path == "/auth/callback" {
        s.handleOIDCCallback(w, r)
        return
    }
    if r.URL.Path == "/api/v1/auth/logout" && r.Method == http.MethodPost {
        s.handleOIDCLogout(w, r)
        return
    }

    // catch the API static specs requests, handle them separately
    if s.specsHandler != nil && strings.HasPrefix(r.URL.Path, "/assets/docs/api") {
        s.specsHandler.ServeHTTP(w, r)
        return
    }

	// ogen api
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/api") {
		s.log.Debug("api request", "method", r.Method, "url", r.URL.Path)
		s.api.ServeHTTP(w, r)
		return
	}

	// try to serve frontend assets (and SPA index.html fallback)
	if s.tryServeFrontend(w, r) {
		return
	}

	// default to ogen (will 404 via WithNotFound)
	s.log.Debug("request", "method", r.Method, "url", r.URL.Path)
	s.api.ServeHTTP(w, r)
}

// Shutdown stops the server
func (s *Server) Shutdown() error {
	return s.listener.Close()
}

func (s *Server) tryServeFrontend(w http.ResponseWriter, r *http.Request) bool {
	dir := s.cfg.FrontendAssetsDir
	if dir == "" {
		// not configured, tell the caller to continue
		return false
	}

	// root → strict index.html check
	if r.URL.Path == "/" {
		indexPath := filepath.Join(dir, "index.html")
		if !isReadableFile(indexPath) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "dist folder is missing or index.html not found"})
			return true
		}
		http.ServeFile(w, r, indexPath)
		return true
	}

	// prevent path traversal
	cleanPath := filepath.Clean(r.URL.Path)
	full := filepath.Join(dir, cleanPath)
	if !strings.HasPrefix(full, filepath.Clean(dir)+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return true
	}

	// if the requested asset exists, serve it
	if isReadableFile(full) {
		http.ServeFile(w, r, full)
		return true
	}

	// SPA fallback: serve index.html for non-file routes (e.g. /app/settings)
	indexPath := filepath.Join(dir, "index.html")
	if isReadableFile(indexPath) {
		http.ServeFile(w, r, indexPath)
		return true
	}

	// index missing → let caller continue
	return false
}

func isReadableFile(p string) bool {
    info, err := os.Stat(p)
    return err == nil && !info.IsDir()
}

// handleOIDCLogin starts the OIDC redirect flow using configured provider (Zitadel by default)
func (s *Server) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
    if s.cfg.Auth.UserManager != "zitadel" {
        http.Error(w, "OIDC login is only supported with Zitadel user manager", http.StatusBadRequest)
        return
    }

    // generate state and remember it for 10 minutes
    state := fmt.Sprintf("%d_%d", time.Now().UnixNano(), os.Getpid())
    s.oidcStates[state] = time.Now().Add(10 * time.Minute)

    // ensure redirect_uri matches current BaseURL if configured
    redirectURI := s.cfg.Auth.Zitadel.RedirectURI
    if redirectURI == "" {
        redirectURI = strings.TrimSuffix(s.cfg.BaseURL, "/") + "/api/v1/auth/callback"
    }

    // build auth URL; override redirect_uri if we computed one
    authURL := s.idp.AuthCodeURL(state, oauth2.SetAuthURLParam("redirect_uri", redirectURI))
    s.log.Info("redirecting to IDP for login", "url", authURL)
    http.Redirect(w, r, authURL, http.StatusFound)
}

// handleOIDCCallback exchanges the code for tokens and persists them to sessionStorage via a small HTML page
func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
    if s.cfg.Auth.UserManager != "zitadel" {
        http.Error(w, "OIDC login is only supported with Zitadel user manager", http.StatusBadRequest)
        return
    }

    q := r.URL.Query()
    state := q.Get("state")
    code := q.Get("code")
    if state == "" || code == "" {
        http.Error(w, "missing state or code", http.StatusBadRequest)
        return
    }

    // validate state (and cleanup expired states)
    now := time.Now()
    for k, exp := range s.oidcStates {
        if now.After(exp) {
            delete(s.oidcStates, k)
        }
    }
    exp, ok := s.oidcStates[state]
    if !ok || now.After(exp) {
        http.Error(w, "invalid or expired state", http.StatusBadRequest)
        return
    }
    delete(s.oidcStates, state)

    // ensure redirect_uri matches what we used for login
    redirectURI := s.cfg.Auth.Zitadel.RedirectURI
    if redirectURI == "" {
        redirectURI = strings.TrimSuffix(s.cfg.BaseURL, "/") + "/api/v1/auth/callback"
    }

    // exchange code
    tok, err := s.idp.Exchange(r.Context(), code, redirectURI)
    if err != nil {
        s.log.Error("OIDC code exchange failed", "error", err)
        http.Error(w, "failed to exchange code", http.StatusBadRequest)
        return
    }

    accessToken := tok.AccessToken
    expiresIn := int(time.Until(tok.Expiry).Seconds())
    if expiresIn < 0 {
        expiresIn = 0
    }
    var idToken string
    if v := tok.Extra("id_token"); v != nil {
        if s, ok := v.(string); ok {
            idToken = s
        }
    }

    // respond with a tiny HTML that stores the token in sessionStorage (shadowapi_auth) used by the frontend
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    // Basic escaping for embedding strings in JS
    esc := func(in string) string { return strings.ReplaceAll(strings.ReplaceAll(in, "\\", "\\\\"), "'", "\\'") }
    fmt.Fprintf(w, `<!doctype html>
<html><head><meta charset="utf-8"><title>Signing in…</title></head>
<body>
<script>
(function(){
  try {
    var data = {
      email: '',
      accessToken: '%s',
      idToken: '%s',
      refreshToken: '',
      expiresAt: Date.now() + (%d * 1000)
    };
    sessionStorage.setItem('shadowapi_auth', JSON.stringify(data));
  } catch(e) { console.error('auth store failed', e); }
  window.location = '/';
})();
</script>
</body></html>`, esc(accessToken), esc(idToken), expiresIn)
}

// handleOIDCLogout attempts to revoke the provided access token at the IDP and returns 204 regardless of outcome
func (s *Server) handleOIDCLogout(w http.ResponseWriter, r *http.Request) {
    authz := r.Header.Get("Authorization")
    if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
        // no token provided; nothing to revoke
        w.WriteHeader(http.StatusNoContent)
        return
    }
    token := strings.TrimPrefix(authz, "Bearer ")

    // Build revoke URL against InstanceURL
    base := strings.TrimSuffix(s.cfg.Auth.Zitadel.InstanceURL, "/")
    revokeURL := base + "/oauth/v2/revoke"

    form := url.Values{}
    form.Set("token", token)
    form.Set("token_type_hint", "access_token")

    req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, revokeURL, strings.NewReader(form.Encode()))
    if err != nil {
        s.log.Warn("logout: build request", "err", err)
        w.WriteHeader(http.StatusNoContent)
        return
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    // Optional: set Host header for proxies
    if s.cfg.Auth.Zitadel.ExternalURL != "" {
        if u, err := url.Parse(s.cfg.Auth.Zitadel.ExternalURL); err == nil {
            req.Host = u.Host
        }
    }

    // best-effort call
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        s.log.Warn("logout: revoke call failed", "err", err)
        w.WriteHeader(http.StatusNoContent)
        return
    }
    _ = resp.Body.Close()
    w.WriteHeader(http.StatusNoContent)
}
