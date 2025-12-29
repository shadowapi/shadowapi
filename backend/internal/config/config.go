// backend/internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/samber/do/v2"
	"gopkg.in/yaml.v3"
)

// Config is the main configuration structure
// No envDefault is specified so values from the config file remain if the environment variable is unset.
type Config struct {
	FrontendAssetsDir string `json:"frontend_assets_dir" yaml:"frontend_assets_dir" env:"BE_FRONTEND_ASSETS_DIR" envDefault:"./dist"`

	// BaseURL root path for the system (SSR/public pages on root domain)
	BaseURL string `json:"base_url" yaml:"base_url" env:"BE_BASE_URL" envDefault:"http://localtest.me"`

	// CSRBaseURL is the URL for the CSR app (app subdomain) - used for login redirects
	CSRBaseURL string `json:"csr_base_url" yaml:"csr_base_url" env:"BE_CSR_BASE_URL" envDefault:"http://app.localtest.me"`

	// APIBaseURL is the base URL for the API (used for OAuth2 callback)
	APIBaseURL string `json:"api_base_url" yaml:"api_base_url" env:"BE_API_BASE_URL" envDefault:"http://api.localtest.me"`

	// Log settings
	Log struct {
		// Level is the log level. Valid values are "debug", "info", "warn", "error".
		Level string `json:"level" yaml:"level" env:"BE_LOG_LEVEL"`
	} `json:"log" yaml:"log"`

	// Server configuration for local UI and API requests
	Server struct {
		Host string `yaml:"host" json:"host" env:"BE_HOST"`
		Port int    `yaml:"port" json:"port" env:"BE_PORT"`
	} `yaml:"server" json:"server"`

	// DB is a database configuration
	DB struct {
		URI string `yaml:"uri,omitempty" json:"uri,omitempty" env:"BE_DB_URI"`
	} `yaml:"db" json:"db"`

	// API settings
	API struct {
		// SpecsDir is the directory where the API specs are stored to serve them
		SpecsDir string `json:"specs_dir" yaml:"specs_dir" env:"BE_API_SPECS_DIR"`
	} `yaml:"api" json:"api"`

	// JWT is a struct that holds all the JWT related methods
	JWT struct {
		// PrivateKey is the secret key used to sign the JWT tokens
		PrivateKey string `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	} `yaml:"jwt" json:"jwt"`

	// InitAdmin configures the first administrator account
	InitAdmin struct {
		Email    string `yaml:"email" json:"email" env:"BE_INIT_ADMIN_EMAIL"`
		Password string `yaml:"password" json:"password" env:"BE_INIT_ADMIN_PASSWORD"`
	} `yaml:"init_admin" json:"init_admin"`

	// Auth is a struct that holds all the authentication settings
	Auth struct {
		// IgnoreHttpsError disables logging OAuth2 HTTPS errors. Useful for development
		IgnoreHttpsError bool `yaml:"ignore_https_error" json:"ignore_https_error" env:"BE_AUTH_IGNORE_HTTPS_ERROR"`
	} `yaml:"auth" json:"auth"`

	// Worker settings
	Worker struct {
		// MaxCount is the maximum number of workers that can be started
		MaxCount int `yaml:"max_count" json:"max_count" env:"BE_WORKER_MAX_COUNT" envDefault:"100"`
	} `yaml:"worker" json:"worker"`

	// Queue settings for the NATS queue
	Queue struct {
		URL      string `yaml:"url" json:"url" env:"BE_QUEUE_URL"`
		Prefix   string `yaml:"prefix" json:"prefix" env:"BE_QUEUE_PREFIX"`
		Username string `yaml:"username" json:"username" env:"BE_QUEUE_USERNAME"`
		Password string `yaml:"password" json:"password" env:"BE_QUEUE_PASSWORD"`
	} `yaml:"queue" json:"queue"`

	// Add cfg.Telegram.AppID, cfg.Telegram.AppHash
	Telegram struct {
		AppHash string `yaml:"app_hash" json:"app_hash" env:"TG_APP_HASH"`
		AppID   int    `yaml:"app_id" json:"app_id" env:"TG_APP_ID"`
	}

	// CORS settings for cross-origin requests
	CORS struct {
		// AllowedOrigins is a comma-separated list of allowed origins
		AllowedOrigins string `yaml:"allowed_origins" json:"allowed_origins" env:"BE_CORS_ALLOWED_ORIGINS" envDefault:""`
	} `yaml:"cors" json:"cors"`

	// GRPC server settings for distributed workers
	GRPC struct {
		Host string `yaml:"host" json:"host" env:"BE_GRPC_HOST" envDefault:"0.0.0.0"`
		Port int    `yaml:"port" json:"port" env:"BE_GRPC_PORT" envDefault:"9090"`
	} `yaml:"grpc" json:"grpc"`

	// OAuth2 settings for Hydra integration
	OAuth2 struct {
		// HydraPublicURL is the URL to Hydra's public endpoints (token, authorize)
		HydraPublicURL string `yaml:"hydra_public_url" json:"hydra_public_url" env:"BE_HYDRA_PUBLIC_URL" envDefault:"http://hydra:4444"`
		// HydraAdminURL is the URL to Hydra's admin endpoints (introspection, revocation)
		HydraAdminURL string `yaml:"hydra_admin_url" json:"hydra_admin_url" env:"BE_HYDRA_ADMIN_URL" envDefault:"http://hydra:4445"`
		// SPAClientID is the OAuth2 client ID for the SPA frontend
		SPAClientID string `yaml:"spa_client_id" json:"spa_client_id" env:"BE_OAUTH2_SPA_CLIENT_ID" envDefault:"shadowapi-spa"`
		// JWKSCacheDuration is how long to cache JWKS keys in seconds
		JWKSCacheDuration int `yaml:"jwks_cache_duration" json:"jwks_cache_duration" env:"BE_JWKS_CACHE_DURATION" envDefault:"300"`
		// CookieDomain is the domain for auth cookies
		CookieDomain string `yaml:"cookie_domain" json:"cookie_domain" env:"BE_OAUTH2_COOKIE_DOMAIN" envDefault:"localtest.me"`
		// CookieSecure sets the Secure flag on cookies (should be true in production)
		CookieSecure bool `yaml:"cookie_secure" json:"cookie_secure" env:"BE_OAUTH2_COOKIE_SECURE" envDefault:"false"`
	} `yaml:"oauth2" json:"oauth2"`

	configPath string
	ext        string
}

// Provide config object instance for the dependency injector
func Provide(i do.Injector) (*Config, error) {
	cPath := do.MustInvokeNamed[string](i, "defaultConfigPath")
	slog.Info("loading configuration", "path", cPath)
	cfg, err := Load(cPath)
	if err != nil {
		return nil, err
	}

	if data, err := yaml.Marshal(cfg); err == nil {
		slog.Info("config loaded", "config", string(data))
	} else {
		slog.Error("marshal config", "err", err)
	}
	return cfg, nil
}

// Load creates a new Config instance
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	if configPath != "" {
		cfg.configPath = configPath
		cfg.ext = strings.ToLower(filepath.Ext(configPath))
		if cfg.ext == "" {
			return nil, fmt.Errorf("config file extension is empty")
		}

		stat, err := os.Stat(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("config file not found: %s", configPath)
			}
			return nil, fmt.Errorf("stat config file: %w", err)
		}
		if stat.IsDir() {
			return nil, fmt.Errorf("config file path is a directory")
		}

		if err := cfg.Load(); err != nil {
			return nil, err
		}

		slog.Info("loaded config file", "path", configPath)
	}

	// env overrides values from file
	if err := env.Parse(cfg); err != nil {
		slog.Error("failed to parse environment variables", "error", err)
	}
	slog.Info("BE_CONFIG_PATH after env parse", "value", os.Getenv("BE_CONFIG_PATH"))

	return cfg, nil
}

// Load loads the configuration from the config file
func (c *Config) Load() error {
	if c.configPath == "" {
		return fmt.Errorf("config file path is empty")
	}

	c.ext = strings.ToLower(filepath.Ext(c.configPath))

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	switch c.ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, c); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, c); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	default:
		return fmt.Errorf("unknown config file format: %s", c.ext)
	}
	return nil
}

// kept for potential future use; not called during normal load
func (c *Config) Save() error {
	if c.ext == "" {
		return fmt.Errorf("config file extension is empty")
	}
	var (
		data []byte
		err  error
	)
	switch c.ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
	default:
		err = fmt.Errorf("unknown config file format: %s", c.ext)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(c.configPath, data, 0o644)
}
