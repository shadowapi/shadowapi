package workspace

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Context key for workspace data
type contextKey string

const (
	WorkspaceSlugContextKey contextKey = "workspace_slug"
	WorkspaceUUIDContextKey contextKey = "workspace_uuid"
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

// OgenMiddleware extracts workspace context from JWT claims first, falling back to URL path.
// When workspace is present in JWT claims, that takes precedence over URL path.
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	ctx := req.Context

	// First, check if workspace info is in JWT claims
	if claims, ok := ctx.Value(auth.UserClaimsContextKey).(*oauth2.Claims); ok && claims != nil {
		if claims.WorkspaceSlug != "" {
			ctx = context.WithValue(ctx, WorkspaceSlugContextKey, claims.WorkspaceSlug)
			ctx = context.WithValue(ctx, WorkspaceUUIDContextKey, claims.WorkspaceID)
			req.SetContext(ctx)
			m.log.Debug("workspace context set from JWT claims",
				"slug", claims.WorkspaceSlug,
				"uuid", claims.WorkspaceID,
			)
			return next(req)
		}
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

// GetWorkspaceUUID retrieves the workspace UUID from context (only available when set from JWT claims)
func GetWorkspaceUUID(ctx context.Context) string {
	if uuid, ok := ctx.Value(WorkspaceUUIDContextKey).(string); ok {
		return uuid
	}
	return ""
}
