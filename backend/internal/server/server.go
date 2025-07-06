package server

import (
        "context"
        "encoding/json"
        "errors"
        "fmt"
        "log/slog"
        "net"
        "net/http"
        "strings"

        "github.com/gofrs/uuid"
        "github.com/jackc/pgx/v5"
        "github.com/jackc/pgx/v5/pgtype"

        "github.com/samber/do/v2"

        "github.com/shadowapi/shadowapi/backend/internal/auth"
        "github.com/shadowapi/shadowapi/backend/internal/config"
        "github.com/shadowapi/shadowapi/backend/internal/handler"
        "github.com/shadowapi/shadowapi/backend/internal/session"
        "github.com/shadowapi/shadowapi/backend/internal/zitadel"
        "github.com/shadowapi/shadowapi/backend/pkg/query"
        "github.com/shadowapi/shadowapi/backend/pkg/api"
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
}

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

	if r.URL.Path == "/auth/callback" {
		s.handleAuthCallback(w, r)
		return
	}

	if r.URL.Path == "/login" && r.Method == http.MethodPost {
		s.handlePlainLogin(w, r)
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

// handleAuthCallback exchanges the code for a token and sets session cookie.
func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
        tok, err := s.zitadel.ExchangeCode(r.Context(), code)
        if err != nil {
                s.log.Error("exchange code", "error", err)
                http.Error(w, "exchange failed", http.StatusInternalServerError)
                return
        }

        // Associate Zitadel subject with a user record
        info, err := s.zitadel.Introspect(r.Context(), tok.AccessToken)
        if err == nil && info.Active {
                q := query.New(s.handler.DB())
                _, errUser := q.GetUserByZitadelSubject(r.Context(), pgtype.Text{String: info.Subject, Valid: true})
                if errors.Is(errUser, pgx.ErrNoRows) {
                        uid := uuid.Must(uuid.NewV7())
                        _, errUser = q.CreateUser(r.Context(), query.CreateUserParams{
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
	http.Redirect(w, r, "/", http.StatusFound)
}
