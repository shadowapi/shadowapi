package log

import (
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Provide logger instance for the dependency injector
func Provide(i do.Injector) (*slog.Logger, error) {
	c := do.MustInvoke[*config.Config](i)
	logLevel := slog.LevelVar{}
	if err := logLevel.UnmarshalText([]byte(c.Log.Level)); err != nil {
		logLevel.Set(slog.LevelError)
	}

	l := slog.New(console.NewHandler(os.Stderr, &console.HandlerOptions{Level: slog.LevelDebug}))
	slog.Info("set default log level", "set_level", logLevel.String())
	return l, nil
}
