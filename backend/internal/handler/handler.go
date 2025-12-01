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
}

func (h *Handler) DB() *pgxpool.Pool {
	return h.dbp
}

// ensureInitTenantAndAdmin creates the default tenants and first admin user if they don't exist.
func (h *Handler) ensureInitTenantAndAdmin(ctx context.Context) error {
	if h.cfg.InitAdmin.Email == "" || h.cfg.InitAdmin.Password == "" {
		return nil
	}

	q := query.New(h.dbp)

	// Define tenants to create
	tenants := []struct {
		name        string
		displayName string
	}{
		{"internal", "Internal"},
		{"demo", "Demo"},
	}

	// Hash password once for all tenants
	hashed, err := bcrypt.GenerateFromPassword([]byte(h.cfg.InitAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	for _, t := range tenants {
		if err := h.ensureTenantWithAdmin(ctx, q, t.name, t.displayName, string(hashed)); err != nil {
			return err
		}
	}

	return nil
}

// ensureTenantWithAdmin creates a tenant and admin user if they don't exist.
func (h *Handler) ensureTenantWithAdmin(ctx context.Context, q *query.Queries, name, displayName, hashedPassword string) error {
	// Step 1: Ensure tenant exists
	var tenantUUID pgtype.UUID
	tenant, err := q.GetTenantByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create tenant
			tUUID := uuid.Must(uuid.NewV7())
			tenantUUID = pgtype.UUID{Bytes: tUUID, Valid: true}
			_, err = q.CreateTenant(ctx, query.CreateTenantParams{
				UUID:        tenantUUID,
				Name:        name,
				DisplayName: displayName,
				IsEnabled:   true,
				Settings:    []byte(`{}`),
			})
			if err != nil {
				return errors.New("failed to create " + name + " tenant: " + err.Error())
			}
			h.log.Info("created tenant", "name", name)
		} else {
			return err
		}
	} else {
		var bytes [16]byte
		copy(bytes[:], tenant.UUID.Bytes())
		tenantUUID = pgtype.UUID{Bytes: bytes, Valid: true}
	}

	// Step 2: Ensure admin user exists in tenant
	users, err := q.ListUsersByTenant(ctx, query.ListUsersByTenantParams{
		TenantUuid: tenantUUID,
		Offset:     0,
		Limit:      1,
	})
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if len(users) > 0 {
		return nil // Admin already exists
	}

	// Create admin user
	uid := uuid.Must(uuid.NewV7())
	_, err = q.CreateUser(ctx, query.CreateUserParams{
		UUID:       pgtype.UUID{Bytes: uid, Valid: true},
		TenantUuid: tenantUUID,
		Email:      h.cfg.InitAdmin.Email,
		Password:   hashedPassword,
		FirstName:  "Admin",
		LastName:   "User",
		IsEnabled:  true,
		IsAdmin:    true,
		Meta:       []byte(`{}`),
	})
	if err != nil {
		return err
	}
	h.log.Info("created admin user in tenant", "tenant", name, "email", h.cfg.InitAdmin.Email)
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
			cfg.BaseURL, // Hydra issuer is the base URL
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
			cfg.BaseURL,
		)

		log.Info("OAuth2 service initialized",
			"client_id", cfg.OAuth2.SPAClientID,
			"hydra_url", cfg.OAuth2.HydraPublicURL,
		)
	}

	if err := h.ensureInitTenantAndAdmin(context.Background()); err != nil {
		h.log.Error("init tenant and admin", "error", err)
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
