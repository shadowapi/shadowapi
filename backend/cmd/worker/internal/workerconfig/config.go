package workerconfig

import (
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// Config holds the worker configuration
type Config struct {
	// Server is the gRPC server address to connect to
	Server string `env:"WORKER_SERVER" envDefault:"localhost:9090"`

	// TLS enables TLS for gRPC connections (required for external connections via rpc.meshpump.com)
	TLS bool `env:"WORKER_TLS" envDefault:"false"`

	// WorkerID is the unique identifier for this worker (set after enrollment)
	WorkerID string `env:"WORKER_ID"`

	// WorkerSecret is the secret for authenticating (set after enrollment)
	WorkerSecret string `env:"WORKER_SECRET"`

	// LogLevel is the logging level
	LogLevel string `env:"WORKER_LOG_LEVEL" envDefault:"info"`

	// LogFormat is the log format ("console" or "json")
	LogFormat string `env:"WORKER_LOG_FORMAT" envDefault:"console"`

	// HeartbeatInterval is the interval between heartbeats in seconds
	HeartbeatInterval int `env:"WORKER_HEARTBEAT_INTERVAL" envDefault:"30"`

	// Capacity is the maximum number of concurrent jobs
	Capacity int `env:"WORKER_CAPACITY" envDefault:"10"`
}

// Provide creates a new worker config for DI
func Provide(i do.Injector) (*Config, error) {
	cfg := &Config{}

	// Parse environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Override from viper (command line flags)
	if v := viper.GetString("server"); v != "" {
		cfg.Server = v
	}
	if v := viper.GetString("worker_id"); v != "" {
		cfg.WorkerID = v
	}
	if v := viper.GetString("worker_secret"); v != "" {
		cfg.WorkerSecret = v
	}
	if v := viper.GetString("log.level"); v != "" {
		cfg.LogLevel = v
	}

	slog.Info("worker config loaded",
		"server", cfg.Server,
		"worker_id", maskSecret(cfg.WorkerID),
		"log_level", cfg.LogLevel,
	)

	return cfg, nil
}

// ProvideLogger creates a logger for DI based on config
func ProvideLogger(i do.Injector) (*slog.Logger, error) {
	cfg := do.MustInvoke[*Config](i)

	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if strings.ToLower(cfg.LogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler), nil
}

// maskSecret returns a masked version of a secret for logging
func maskSecret(s string) string {
	if len(s) == 0 {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
