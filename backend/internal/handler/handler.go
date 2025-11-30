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

// ensureInitAdmin creates the first admin user if the DB has no users yet.
func (h *Handler) ensureInitAdmin(ctx context.Context) error {
	if h.cfg.InitAdmin.Email == "" || h.cfg.InitAdmin.Password == "" {
		return nil
	}
	q := query.New(h.dbp)
	users, err := q.ListUsers(ctx, query.ListUsersParams{Offset: 0, Limit: 1})
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if len(users) > 0 {
		return nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(h.cfg.InitAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	uid := uuid.Must(uuid.NewV7())
	_, err = q.CreateUser(ctx, query.CreateUserParams{
		UUID:      pgtype.UUID{Bytes: uid, Valid: true},
		Email:     h.cfg.InitAdmin.Email,
		Password:  string(hashed),
		FirstName: "Admin",
		LastName:  "User",
		IsEnabled: true,
		IsAdmin:   true,
		Meta:      []byte(`{}`),
	})
	return err
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

	if err := h.ensureInitAdmin(context.Background()); err != nil {
		h.log.Error("init admin", "error", err)
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
