package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/tracelog"
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
