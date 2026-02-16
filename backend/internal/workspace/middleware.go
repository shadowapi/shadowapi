package workspace

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Context key for workspace data
type contextKey string

const (
	WorkspaceSlugContextKey contextKey = "workspace_slug"
	WorkspaceUUIDContextKey contextKey = "workspace_uuid"
)

// Middleware handles workspace context extraction from cookie or URL path
type Middleware struct {
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool
}

// Provide creates a new workspace middleware for the dependency injector
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	return &Middleware{
		cfg: cfg,
		log: log,
		dbp: dbp,
	}, nil
}

// OgenMiddleware extracts workspace context from cookie first, falling back to URL path.
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	ctx := req.Context

	// First, check if workspace slug is in the cookie
	if slug, err := oauth2.GetWorkspaceSlugFromCookie(req.Raw); err == nil && slug != "" {
		// Look up workspace UUID from DB
		ws, err := query.New(m.dbp).GetWorkspaceBySlug(ctx, slug)
		if err == nil {
			ctx = context.WithValue(ctx, WorkspaceSlugContextKey, slug)
			ctx = context.WithValue(ctx, WorkspaceUUIDContextKey, ws.UUID.String())
			req.SetContext(ctx)
			m.log.Debug("workspace context set from cookie",
				"slug", slug,
				"uuid", ws.UUID.String(),
			)
			return next(req)
		}
		m.log.Debug("workspace cookie slug not found in DB", "slug", slug, "error", err)
	}

	// Fall back to extracting workspace slug from URL path like /api/v1/w/{slug}/...
	path := req.Raw.URL.Path
	slug := extractWorkspaceSlug(path)
	if slug != "" {
		ctx = context.WithValue(ctx, WorkspaceSlugContextKey, slug)
		req.SetContext(ctx)
		m.log.Debug("workspace context set from URL path", "slug", slug, "path", path)
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

// GetWorkspaceUUID retrieves the workspace UUID from context
func GetWorkspaceUUID(ctx context.Context) string {
	if uuid, ok := ctx.Value(WorkspaceUUIDContextKey).(string); ok {
		return uuid
	}
	return ""
}
