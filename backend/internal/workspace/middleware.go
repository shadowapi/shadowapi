package workspace

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Context key for workspace data
type contextKey string

const (
	WorkspaceSlugContextKey contextKey = "workspace_slug"
)

// Middleware handles workspace context extraction from URL path
type Middleware struct {
	cfg *config.Config
	log *slog.Logger
}

// Provide creates a new workspace middleware for the dependency injector
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)

	return &Middleware{
		cfg: cfg,
		log: log,
	}, nil
}

// OgenMiddleware extracts workspace slug from URL path (/w/{slug}/...)
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	path := req.Raw.URL.Path

	// Extract workspace slug from path like /api/v1/w/{slug}/...
	slug := extractWorkspaceSlug(path)
	if slug != "" {
		ctx := context.WithValue(req.Context, WorkspaceSlugContextKey, slug)
		req.SetContext(ctx)
		m.log.Debug("workspace context set", "slug", slug, "path", path)
	}

	return next(req)
}

// extractWorkspaceSlug extracts the workspace slug from URL path
// Expected format: /api/v1/w/{slug}/... or /w/{slug}/...
func extractWorkspaceSlug(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "w" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// GetWorkspaceSlug retrieves the workspace slug from context
func GetWorkspaceSlug(ctx context.Context) string {
	if slug, ok := ctx.Value(WorkspaceSlugContextKey).(string); ok {
		return slug
	}
	return ""
}
