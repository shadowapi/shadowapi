package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/auth/oauth2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/rbac"
	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Handler is the server handler
type Handler struct {
	cfg         *config.Config
	log         *slog.Logger
	dbp         *pgxpool.Pool
	wbr         *worker.Broker
	userManager auth.UserManager
	oauth2Svc   *OAuth2Service
	enforcer    *rbac.Enforcer
}

func (h *Handler) DB() *pgxpool.Pool {
	return h.dbp
}

// ensureInitWorkspaceAndAdmin creates the default workspaces and admin user if they don't exist.
func (h *Handler) ensureInitWorkspaceAndAdmin(ctx context.Context) error {
	if h.cfg.InitAdmin.Email == "" || h.cfg.InitAdmin.Password == "" {
		return nil
	}

	q := query.New(h.dbp)

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(h.cfg.InitAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Step 1: Ensure admin user exists (global user, not workspace-specific)
	var userUUID pgtype.UUID
	user, err := q.GetUserByEmail(ctx, h.cfg.InitAdmin.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create admin user
			uid := uuid.Must(uuid.NewV7())
			userUUID = pgtype.UUID{Bytes: uid, Valid: true}
			_, err = q.CreateUser(ctx, query.CreateUserParams{
				UUID:      userUUID,
				Email:     h.cfg.InitAdmin.Email,
				Password:  string(hashed),
				FirstName: "Admin",
				LastName:  "User",
				IsEnabled: true,
				Meta:      []byte(`{}`),
			})
			if err != nil {
				return errors.New("failed to create admin user: " + err.Error())
			}
			h.log.Info("created admin user", "email", h.cfg.InitAdmin.Email)
		} else {
			return err
		}
	} else {
		userUUID = pgtype.UUID{Bytes: user.UUID, Valid: true}
	}

	// Step 2: Assign super_admin role (global scope)
	userUUIDStr := uuid.UUID(userUUID.Bytes).String()
	if !h.enforcer.HasRoleForUserInDomain(userUUIDStr, rbac.RoleSuperAdmin, "global") {
		if err := h.enforcer.AddRoleForUserInDomain(userUUIDStr, rbac.RoleSuperAdmin, "global"); err != nil {
			h.log.Warn("failed to assign super_admin role", "email", h.cfg.InitAdmin.Email, "error", err)
		} else {
			h.log.Info("assigned super_admin role to admin user", "email", h.cfg.InitAdmin.Email)
		}
	}

	// Step 3: Ensure default workspaces exist
	workspaces := []struct {
		slug        string
		displayName string
	}{
		{"internal", "Internal"},
		{"demo", "Demo"},
	}

	for _, w := range workspaces {
		if err := h.ensureWorkspaceWithOwner(ctx, q, w.slug, w.displayName, userUUID); err != nil {
			return err
		}
	}

	return nil
}

// ensureWorkspaceWithOwner creates a workspace and adds the user as owner if it doesn't exist.
func (h *Handler) ensureWorkspaceWithOwner(ctx context.Context, q *query.Queries, slug, displayName string, userUUID pgtype.UUID) error {
	// Convert pgtype.UUID to gofrs/uuid
	userUUIDVal := uuid.UUID(userUUID.Bytes)

	// Check if workspace exists
	workspace, err := q.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create workspace
			workspaceUUID := uuid.Must(uuid.NewV7())
			_, err = q.CreateWorkspace(ctx, query.CreateWorkspaceParams{
				UUID:        workspaceUUID,
				Slug:        slug,
				DisplayName: displayName,
				IsEnabled:   true,
				Settings:    []byte(`{}`),
			})
			if err != nil {
				return errors.New("failed to create " + slug + " workspace: " + err.Error())
			}
			h.log.Info("created workspace", "slug", slug)

			// Add user as owner
			memberUUID := uuid.Must(uuid.NewV7())
			_, err = q.CreateWorkspaceMember(ctx, query.CreateWorkspaceMemberParams{
				UUID:          memberUUID,
				WorkspaceUUID: &workspaceUUID,
				UserUUID:      &userUUIDVal,
				Role:          "owner",
			})
			if err != nil {
				return errors.New("failed to add workspace owner: " + err.Error())
			}
			h.log.Info("added user as workspace owner", "workspace", slug)
		} else {
			return err
		}
	} else {
		// Workspace exists, ensure user is a member
		workspaceUUID := workspace.UUID
		_, err := q.GetWorkspaceMember(ctx, query.GetWorkspaceMemberParams{
			WorkspaceUUID: &workspaceUUID,
			UserUUID:      &userUUIDVal,
		})
		if err == pgx.ErrNoRows {
			// Add user as owner
			memberUUID := uuid.Must(uuid.NewV7())
			_, err = q.CreateWorkspaceMember(ctx, query.CreateWorkspaceMemberParams{
				UUID:          memberUUID,
				WorkspaceUUID: &workspaceUUID,
				UserUUID:      &userUUIDVal,
				Role:          "owner",
			})
			if err != nil {
				h.log.Warn("failed to add workspace member", "workspace", slug, "error", err)
			} else {
				h.log.Info("added user as workspace owner", "workspace", slug)
			}
		}
	}

	return nil
}

// Provide API handler instance for the dependency injector
func Provide(i do.Injector) (*Handler, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)

	h := &Handler{
		cfg:         cfg,
		log:         log,
		dbp:         do.MustInvoke[*pgxpool.Pool](i),
		wbr:         do.MustInvoke[*worker.Broker](i),
		userManager: do.MustInvoke[auth.UserManager](i),
		enforcer:    do.MustInvoke[*rbac.Enforcer](i),
	}

	// Initialize OAuth2 service if configured
	if cfg.OAuth2.SPAClientID != "" {
		jwksURL := cfg.OAuth2.HydraPublicURL + "/.well-known/jwks.json"
		jwksCache := oauth2.NewJWKSCache(
			jwksURL,
			time.Duration(cfg.OAuth2.JWKSCacheDuration)*time.Second,
			log,
		)

		jwtValidator := oauth2.NewJWTValidator(
			jwksCache,
			cfg.OAuth2.HydraPublicURL, // Issuer matches Hydra's self.issuer URL
			log,
		)

		hydraClient := oauth2.NewHydraClient(
			cfg.OAuth2.HydraPublicURL,
			cfg.OAuth2.HydraAdminURL,
			log,
		)

		cookieConfig := oauth2.CookieConfig{
			Domain:   cfg.OAuth2.CookieDomain,
			Secure:   cfg.OAuth2.CookieSecure,
			SameSite: http.SameSiteLaxMode,
		}

		h.oauth2Svc = NewOAuth2Service(
			hydraClient,
			jwtValidator,
			cookieConfig,
			cfg.OAuth2.SPAClientID,
			cfg.CSRBaseURL, // Login page is on app subdomain
			cfg.APIBaseURL,
		)

		log.Info("OAuth2 service initialized",
			"client_id", cfg.OAuth2.SPAClientID,
			"hydra_url", cfg.OAuth2.HydraPublicURL,
		)
	}

	if err := h.ensureInitWorkspaceAndAdmin(context.Background()); err != nil {
		h.log.Error("init workspace and admin", "error", err)
	}
	return h, nil
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	statusCode := http.StatusInternalServerError

	// Handle SecurityError specifically - return 401 Unauthorized for authentication failures
	if _, ok := err.(*ogenerrors.SecurityError); ok {
		statusCode = http.StatusUnauthorized
	} else if errors.Is(err, &errWraper{}) {
		err := err.(*errWraper)
		statusCode = err.status
	} else if sc, ok := err.(interface{ StatusCode() int }); ok {
		statusCode = sc.StatusCode()
	}
	return &api.ErrorStatusCode{
		StatusCode: statusCode,
		Response: api.Error{
			Status: api.OptInt64{
				Value: int64(statusCode),
				Set:   true,
			},
			Detail: api.OptString{
				Value: err.Error(),
				Set:   true,
			},
		},
	}
}

// getUserUUIDFromContext extracts the authenticated user's UUID from the JWT claims in context.
// Returns an error if the user is not authenticated or claims are missing.
func getUserUUIDFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(auth.UserClaimsContextKey).(*oauth2.Claims)
	if !ok || claims == nil {
		return "", errors.New("authentication required")
	}
	if claims.Subject == "" {
		return "", errors.New("invalid user claims: missing subject")
	}
	return claims.Subject, nil
}
