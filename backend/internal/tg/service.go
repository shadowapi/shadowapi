package tg

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
)

type Service struct {
	ctx context.Context
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool
}

func Provide(i do.Injector) (*Service, error) {
	ctx := do.MustInvoke[context.Context](i)
	cfg := do.MustInvoke[*config.Config](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	log := do.MustInvoke[*slog.Logger](i).With("service", "tg")

	return &Service{
		ctx: ctx,
		cfg: cfg,
		log: log,
		dbp: dbp,
	}, nil
}
