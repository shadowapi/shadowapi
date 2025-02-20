package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// Handler is the server handler
type Handler struct {
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool
	wbr *worker.Broker
}

// Provide API handler instance for the dependency injector
func Provide(i do.Injector) (*Handler, error) {
	return &Handler{
		cfg: do.MustInvoke[*config.Config](i),
		log: do.MustInvoke[*slog.Logger](i),
		dbp: do.MustInvoke[*pgxpool.Pool](i),
		wbr: do.MustInvoke[*worker.Broker](i),
	}, nil
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
