package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	zitadellog "github.com/shadowapi/shadowapi/backend/internal/auth/zitadel"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/internal/zitadel"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

type Server struct {
	cfg *config.Config
	log *slog.Logger

	api          *api.Server
	listener     net.Listener
	specsHandler http.Handler
	zitadel      *zitadel.Client
	handler      *handler.Handler
	sessions     *session.Middleware
	auth         *auth.Auth
}

// ----- helper for PKCE -------------------------------------------------------

func newCodeVerifier() (string, string, error) {
	b := make([]byte, 43) // 256-bit entropy
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// -----------------------------------------------------------------------------

// Provide server instance for the dependency injector
func Provide(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	authService := do.MustInvoke[*auth.Auth](i)
	handlerService := do.MustInvoke[*handler.Handler](i)
	authMiddleware := do.MustInvoke[*session.Middleware](i)
	zitadelClient := zitadel.Provide(cfg)

	srv, err := api.NewServer(
		handlerService,
		authService,
		api.WithPathPrefix("/api/v1"),
		api.WithMiddleware(authMiddleware.OgenMiddleware),
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
		zitadel:      zitadelClient,
		handler:      handlerService,
		sessions:     authMiddleware,
		auth:         authService,
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

// ServeHTTP implements the http.Handler interface to wrap the API server
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "ok"}); err != nil {
			s.log.Error("failed to encode JSON response", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.URL.Path == "/login/zitadel" {
		s.handleZitadelLogin(w, r)
		return
	}

	if r.URL.Path == "/auth/callback" {
		s.handleAuthCallback(w, r)
		return
	}

	if r.URL.Path == "/login" && r.Method == http.MethodPost {
		s.handlePlainLogin(w, r)
		return
	}

	if r.URL.Path == "/logout" {
		s.handleLogout(w, r)
		return
	}

	if r.URL.Path == "/logout/callback" {
		s.handleLogoutCallback(w, r)
		return
	}

	// catch the API static specs requests, handle them separately
	if s.specsHandler != nil && strings.HasPrefix(r.URL.Path, "/assets/docs/api") {
		s.specsHandler.ServeHTTP(w, r)
		return
	}

	s.log.Debug("request", "method", r.Method, "url", r.URL.Path)
	s.api.ServeHTTP(w, r)
}

// Shutdown stops the server
func (s *Server) Shutdown() error {
	return s.listener.Close()
}

// ---------------- ZITADEL PKCE login flow ------------------------------------

func (s *Server) handleZitadelLogin(w http.ResponseWriter, r *http.Request) {
	verifier, challenge, err := newCodeVerifier()
	if err != nil {
		s.log.Error("pkce generation failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sa_pkce",
		Value:    verifier,
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   300,
	})
	state := fmt.Sprintf("%d", time.Now().UnixNano())
	u := s.cfg.Auth.Zitadel.InstanceURL + "/oauth/v2/authorize?" + url.Values{
		"client_id":             {s.cfg.Auth.Zitadel.Audience},
		"response_type":         {"code"},
		"scope":                 {"openid profile email"},
		"redirect_uri":          {s.cfg.Auth.Zitadel.RedirectURI},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}.Encode()
	http.Redirect(w, r, u, http.StatusFound)
}

// handleAuthCallback exchanges the code for a token and sets session cookie.
func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	verCookie, _ := r.Cookie("sa_pkce")
	verifier := ""
	if verCookie != nil {
		verifier = verCookie.Value
	}

	tok, err := s.zitadel.ExchangeCode(r.Context(), code, verifier)
	if err != nil {
		if s.auth.IgnoreHttpsError {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				s.log.Info("exchange code", "code", code, "url", urlErr.URL, "err", urlErr.Err)
				http.Error(w, "exchange failed", http.StatusInternalServerError)
				return
			}
		}
		zitadellog.LogExchangeError(s.log, err)
		s.log.Error("exchange code",
			"code", code,
			"query", r.URL.RawQuery,
			"token_url", fmt.Sprintf("%s/oauth/v2/token", s.cfg.Auth.Zitadel.InstanceURL))
		http.Error(w, "exchange failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "sa_pkce",
		Value:  "",
		Path:   "/auth",
		MaxAge: -1,
	})

	// Associate Zitadel subject with a user record and create local session
	var userID string
	info, err := s.zitadel.Introspect(r.Context(), tok.AccessToken)
	if err == nil && info.Active {
		q := query.New(s.handler.DB())
		user, errUser := q.GetUserByZitadelSubject(r.Context(), pgtype.Text{String: info.Subject, Valid: true})
		if errors.Is(errUser, pgx.ErrNoRows) {
			uid := uuid.Must(uuid.NewV7())
			user, errUser = q.CreateUser(r.Context(), query.CreateUserParams{
				UUID:           pgtype.UUID{Bytes: uid, Valid: true},
				Email:          fmt.Sprintf("zitadel_%s@example.com", uid.String()),
				Password:       "",
				FirstName:      "",
				LastName:       "",
				IsEnabled:      true,
				IsAdmin:        false,
				ZitadelSubject: pgtype.Text{String: info.Subject, Valid: true},
				Meta:           []byte(`{}`),
			})
			if errUser != nil {
				s.log.Error("create user", "error", errUser)
			}
		} else if errUser != nil {
			s.log.Error("lookup user", "error", errUser)
		}
		if errUser == nil {
			userID = user.UUID.String()
		}
	}
	if userID != "" {
		token := uuid.Must(uuid.NewV7()).String()
		s.sessions.AddSession(token, userID)
		http.SetCookie(w, &http.Cookie{
			Name:     "sa_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "zitadel_access_token",
		Value:    tok.AccessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

// handlePlainLogin verifies email/password and sets session cookie.
func (s *Server) handlePlainLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userID, err := s.handler.PlainLogin(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token := uuid.Must(uuid.NewV7()).String()
	s.sessions.AddSession(token, userID)
	http.SetCookie(w, &http.Cookie{
		Name:     "sa_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	_ = json.NewEncoder(w).Encode(map[string]bool{"active": true})
}

// handleLogout invalidates local session and redirects to the appropriate logout flow.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	sessCookie, errSess := r.Cookie("sa_session")
	if errSess == nil {
		s.sessions.DeleteSession(sessCookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "sa_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	_, zitadelCookieErr := r.Cookie("zitadel_access_token")
	http.SetCookie(w, &http.Cookie{
		Name:   "zitadel_access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	if zitadelCookieErr == nil {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		redirect := fmt.Sprintf("%s://%s/logout/callback", scheme, r.Host)
		target := fmt.Sprintf("%s/oidc/v1/end_session?post_logout_redirect_uri=%s",
			s.cfg.Auth.Zitadel.InstanceURL, url.QueryEscape(redirect))
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/logout/callback", http.StatusFound)
}

// handleLogoutCallback clears session cookie.
func (s *Server) handleLogoutCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sa_session")
	if err == nil {
		s.sessions.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "sa_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "zitadel_access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
