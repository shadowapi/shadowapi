package config

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v9"
	"github.com/gofrs/uuid"
	"github.com/samber/do/v2"
)

// Config holds configuration for the grpc2nats service
type Config struct {
	// InstanceID uniquely identifies this grpc2nats instance (for distributed state)
	InstanceID string `env:"G2N_INSTANCE_ID"`

	// Log settings
	Log struct {
		Level string `env:"G2N_LOG_LEVEL" envDefault:"info"`
	}

	// GRPC server settings
	GRPC struct {
		Host string `env:"G2N_GRPC_HOST" envDefault:"0.0.0.0"`
		Port int    `env:"G2N_GRPC_PORT" envDefault:"9090"`
	}

	// Database settings
	DB struct {
		URI string `env:"G2N_DB_URI"`
	}

	// NATS settings
	NATS struct {
		URL      string `env:"G2N_NATS_URL" envDefault:"nats://localhost:4222"`
		Username string `env:"G2N_NATS_USERNAME"`
		Password string `env:"G2N_NATS_PASSWORD"`
		Prefix   string `env:"G2N_NATS_PREFIX" envDefault:"shadowapi"`
	}

	// KV store settings
	KV struct {
		// WorkerBucket is the NATS KV bucket name for worker state
		WorkerBucket string `env:"G2N_KV_WORKER_BUCKET" envDefault:"workers"`
	}
}

// Provide creates a Config instance for dependency injection
func Provide(i do.Injector) (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Generate instance ID if not provided
	if cfg.InstanceID == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = uuid.Must(uuid.NewV7()).String()[:8]
		}
		cfg.InstanceID = "g2n-" + hostname
	}

	slog.Info("grpc2nats config loaded",
		"instance_id", cfg.InstanceID,
		"grpc_host", cfg.GRPC.Host,
		"grpc_port", cfg.GRPC.Port,
		"nats_url", cfg.NATS.URL,
	)

	return cfg, nil
}

// Subjects returns the NATS subject patterns for job routing
func (c *Config) Subjects() SubjectConfig {
	return SubjectConfig{
		Prefix: c.NATS.Prefix,
	}
}

// SubjectConfig holds NATS subject patterns
type SubjectConfig struct {
	Prefix string
}

// JobsGlobal returns the subject for global jobs: {prefix}.jobs.global.>
func (s SubjectConfig) JobsGlobal() string {
	return s.Prefix + ".jobs.global.>"
}

// JobsWorkspace returns the subject for workspace jobs: {prefix}.jobs.workspace.{slug}.>
func (s SubjectConfig) JobsWorkspace(slug string) string {
	return s.Prefix + ".jobs.workspace." + slug + ".>"
}

// JobsAll returns the subject pattern for all jobs: {prefix}.jobs.>
func (s SubjectConfig) JobsAll() string {
	return s.Prefix + ".jobs.>"
}

// Results returns the subject for job results: {prefix}.results.{jobID}
func (s SubjectConfig) Results(jobID string) string {
	return s.Prefix + ".results." + jobID
}

// ResultsAll returns the subject pattern for all results: {prefix}.results.>
func (s SubjectConfig) ResultsAll() string {
	return s.Prefix + ".results.>"
}

// WorkerStatus returns the subject for worker status: {prefix}.workers.status.{workerID}
func (s SubjectConfig) WorkerStatus(workerID string) string {
	return s.Prefix + ".workers.status." + workerID
}

// WorkerStatusAll returns the subject pattern for all worker status: {prefix}.workers.status.>
func (s SubjectConfig) WorkerStatusAll() string {
	return s.Prefix + ".workers.status.>"
}

// DataAll returns the subject pattern for all data records: {prefix}.data.>
func (s SubjectConfig) DataAll() string {
	return s.Prefix + ".data.>"
}
