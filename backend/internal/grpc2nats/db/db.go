package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
)

// Logger is a tracelog.Logger that logs to a slog.Logger.
type Logger struct {
	logger *slog.Logger
}

// newLogger returns a new Logger that logs to the given slog.Logger.
func newLogger(logger *slog.Logger) *Logger {
	return &Logger{logger: logger}
}

// Log a message at the given level with data key/value pairs. data may be nil.
func (pl *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	var fields []interface{}
	fields = make([]interface{}, 0, len(data)*2)
	for k, v := range data {
		fields = append(fields, k, v)
	}

	switch level {
	case tracelog.LogLevelInfo:
		pl.logger.Info(msg, fields...)
	case tracelog.LogLevelDebug:
		pl.logger.Debug(msg, fields...)
	case tracelog.LogLevelError:
		pl.logger.Error(msg, fields...)
	}
}

// Provide database connection pool for grpc2nats service
func Provide(i do.Injector) (*pgxpool.Pool, error) {
	ctx := do.MustInvoke[context.Context](i)
	conf := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)

	if conf.DB.URI == "" {
		log.Error("database URI is empty")
		return nil, fmt.Errorf("failed to connect to database: database URI is empty")
	}

	cfg, err := pgxpool.ParseConfig(conf.DB.URI)
	if err != nil {
		log.Error("parse config", "error", err)
		return nil, err
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newLogger(log),
		LogLevel: tracelog.LogLevelDebug,
	}

	slog.Debug("connecting to database", "uri", conf.DB.URI)
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
