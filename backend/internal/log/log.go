package log

import (
	"log/slog"
	"os"
	"strings"

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

	var handler slog.Handler
	if strings.ToLower(c.Log.Format) == "json" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel.Level()})
	} else {
		handler = console.NewHandler(os.Stderr, &console.HandlerOptions{Level: logLevel.Level()})
	}

	l := slog.New(handler)
	slog.Info("set default log level", "set_level", logLevel.String(), "format", c.Log.Format)
	return l, nil
}
