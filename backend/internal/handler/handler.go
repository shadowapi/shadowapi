package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Handler is the server handler
type Handler struct {
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool
	wbr *worker.Broker
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
	_, err = q.CreateUser(ctx, query.CreateUserParams{
		UUID:           pgtype.UUID{Bytes: uuid.Must(uuid.NewV7()).Bytes(), Valid: true},
		Email:          h.cfg.InitAdmin.Email,
		Password:       h.cfg.InitAdmin.Password,
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
		cfg: do.MustInvoke[*config.Config](i),
		log: do.MustInvoke[*slog.Logger](i),
		dbp: do.MustInvoke[*pgxpool.Pool](i),
		wbr: do.MustInvoke[*worker.Broker](i),
	}
	if err := h.ensureInitAdmin(context.Background()); err != nil {
		h.log.Error("init admin", "error", err)
	}
	return h, nil
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	statusCode := http.StatusInternalServerError
	if errors.Is(err, &errWraper{}) {
		err := err.(*errWraper)
		statusCode = err.status
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
