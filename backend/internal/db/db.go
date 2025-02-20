package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// Provide database connection pool for the dependency injector
func Provide(i do.Injector) (*pgxpool.Pool, error) {
	ctx := do.MustInvoke[context.Context](i)
	config := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)

	if config.DB.URI == "" {
		log.Error("database URI is empty")
		return nil, fmt.Errorf("failed to connect to database: database URI is empty")
	}

	cfg, err := pgxpool.ParseConfig(config.DB.URI)
	if err != nil {
		log.Error("parse config", "error", err)
		return nil, err
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newLogger(log),
		LogLevel: tracelog.LogLevelDebug,
	}

	slog.Debug("connecting to database", "uri", config.DB.URI)
	dbpool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Error("unable to create connection pool", "error", err)
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		log.Error("failed to ping database", "error", err)
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return dbpool, nil
}
