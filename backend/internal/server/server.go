package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type Server struct {
	cfg *config.Config
	log *slog.Logger

	api          *api.Server
	listener     net.Listener
	specsHandler http.Handler
}

// Provide server instance for the dependency injector
func Provide(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	authService := do.MustInvoke[*auth.Auth](i)
	handlerService := do.MustInvoke[*handler.Handler](i)
	authMiddleware := do.MustInvoke[*session.Middleware](i)

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
