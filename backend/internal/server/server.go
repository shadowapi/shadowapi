package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/internal/zitadel"
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
	auth         *auth.Auth
}

// Provide server instance for the dependency injector
func Provide(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	authService := do.MustInvoke[*auth.Auth](i)
	handlerService := do.MustInvoke[*handler.Handler](i)
	zitadelClient := zitadel.Provide(cfg)

	srv, err := api.NewServer(
		handlerService,
		authService,
		api.WithPathPrefix("/api/v1"),
		// api.WithMiddleware(authService.OgenMiddleware),
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

// ServeHTTP wraps the API server and also serves the frontend dist (SPA) with index.html fallback
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
