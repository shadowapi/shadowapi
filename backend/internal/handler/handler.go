package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
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
		UUID:           pgtype.UUID{Bytes: uid, Valid: true},
		Email:          h.cfg.InitAdmin.Email,
		Password:       string(hashed),
		FirstName:      "Admin",
		LastName:       "User",
		IsEnabled:      true,
		IsAdmin:        true,
		ZitadelSubject: pgtype.Text{},
		Meta:           []byte(`{}`),
	})
	return err
}

// Provide API handler instance for the dependency injector
func Provide(i do.Injector) (*Handler, error) {
	h := &Handler{
		cfg:         do.MustInvoke[*config.Config](i),
		log:         do.MustInvoke[*slog.Logger](i),
		dbp:         do.MustInvoke[*pgxpool.Pool](i),
		wbr:         do.MustInvoke[*worker.Broker](i),
		userManager: do.MustInvoke[auth.UserManager](i),
	}
	if err := h.ensureInitAdmin(context.Background()); err != nil {
		h.log.Error("init admin", "error", err)
	}
	return h, nil
}

// CreateUserSession implements createUserSession operation.
//
// Create a session token for Zitadel authentication
//
// POST /users/session
func (h *Handler) CreateUserSession(ctx context.Context) (*api.UserSessionToken, error) {
	h.log.Info("creating user session token")

	// Check if we're using Zitadel user manager
	if h.cfg.Auth.UserManager != "zitadel" {
		h.log.Error("session token creation requires Zitadel user manager")
		return nil, ErrWithCode(http.StatusBadRequest, E("session token creation requires Zitadel user manager"))
	}

	// Get the service user token from ZitadelUserManager
	zitadelManager, ok := h.userManager.(interface{ GetAuthToken(context.Context) (string, error) })
	if !ok {
		h.log.Error("user manager doesn't support GetAuthToken method")
		return nil, ErrWithCode(http.StatusInternalServerError, E("Zitadel user manager not properly configured"))
	}

	token, err := zitadelManager.GetAuthToken(ctx)
	if err != nil {
		h.log.Error("failed to get Zitadel auth token", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get authentication token"))
	}

	response := &api.UserSessionToken{
		SessionToken: token,
		ZitadelURL:   h.cfg.Auth.Zitadel.InstanceURL,
		ExpiresIn:    3600, // 1 hour
	}

	h.log.Info("user session token created successfully", "zitadel_url", h.cfg.Auth.Zitadel.InstanceURL)
	return response, nil
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	statusCode := http.StatusInternalServerError
	if errors.Is(err, &errWraper{}) {
		err := err.(*errWraper)
		statusCode = err.status
	} else if sc, ok := err.(interface{ StatusCode() int }); ok {
		statusCode = sc.StatusCode()
	}
	return &api.ErrorStatusCode{
		StatusCode: statusCode,
		Response: api.Error{
			Status: api.OptInt64{
				Value: http.StatusInternalServerError,
				Set:   true,
			},
			Detail: api.OptString{
				Value: err.Error(),
				Set:   true,
			},
		},
	}
}
