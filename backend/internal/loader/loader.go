package loader

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"log/slog"
)

type Loader struct {
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool
}

// Provide loader instance for the dependency injector
func Provide(i do.Injector) (*Loader, error) {
	log := do.MustInvoke[*slog.Logger](i)
	log.Debug("Registering loader")
	return &Loader{
		cfg: do.MustInvoke[*config.Config](i),
		log: log,
		dbp: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

// Run starts the loader
func (s *Loader) Run(ctx context.Context) error {
	print("Run loader")
	return nil
}
