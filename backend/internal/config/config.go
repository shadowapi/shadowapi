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
type Config struct {
	FrontendAssetsDir string `json:"frontend_assets_dir" yaml:"frontend_assets_dir" env:"SA_FRONTEND_ASSETS_DIR" envDefault:"./dist"`

	// Log settings
	Log struct {
		// Level is the log level. Valid values are "debug", "info", "warn", "error".
		Level string `json:"level" yaml:"level" env:"SA_LOG_LEVEL" envDefault:"debug"`
	} `json:"log" yaml:"log"`

	// Server configuration for local UI and API requests
	Server struct {
		Host string `yaml:"host" json:"host" env:"SA_HOST" envDefault:"localhost"`
		Port int    `yaml:"port" json:"port" env:"SA_PORT" envDefault:"8080"`
	} `yaml:"server" json:"server"`

	// DB is a database configuration
	DB struct {
		URI string `yaml:"uri,omitempty" json:"uri,omitempty" env:"SA_DB_URI" envDefault:""`
	} `yaml:"db" json:"db"`

	// API settings
	API struct {
		// SpecsDir is the directory where the API specs are stored to serve them
		SpecsDir string `json:"specs_dir" yaml:"specs_dir" env:"SA_API_SPECS_DIR" envDefault:"../spec/"`
	} `yaml:"api" json:"api"`

	// JWT is a struct that holds all the JWT related methods
	JWT struct {
		// PrivateKey is the secret key used to sign the JWT tokens
		PrivateKey string `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	} `yaml:"jwt" json:"jwt"`

	// InitAdmin configures the first administrator account
	InitAdmin struct {
		Email    string `yaml:"email" json:"email"`
		Password string `yaml:"password" json:"password"`
	} `yaml:"init_admin" json:"init_admin"`

	// Auth is a struct that holds all the authentication settings
	Auth struct {

		// TODO @reactima remove this
		// IgnoreHttpsError disables logging OAuth2 HTTPS errors. Useful for development
		IgnoreHttpsError bool `yaml:"ignore_https_error" json:"ignore_https_error" env:"SA_AUTH_IGNORE_HTTPS_ERROR" envDefault:"false"`

		// BearerToken is used to validate incoming requests that carry an Authorization header.
		BearerToken string `yaml:"bearer_token" json:"bearer_token" env:"SA_AUTH_BEARER_TOKEN" envDefault:"mysecretapikey"`

		// Zitadel configuration for OAuth2 authentication
		Zitadel struct {
			InstanceURL string `json:"instance_url" yaml:"instance_url" env:"SA_ZITADEL_INSTANCE_URL"`

			// ---- machine-to-machine credentials ----
			// Service user → Basic-Auth (client-credentials / introspect)
			ServiceClientID     string `json:"service_client_id" yaml:"service_client_id" env:"SA_ZITADEL_SERVICE_CLIENT_ID"`
			ServiceClientSecret string `json:"service_client_secret" yaml:"service_client_secret" env:"SA_ZITADEL_SERVICE_CLIENT_SECRET"`

			// Optional JWT-Profile flow (no shared secret)
			APIKeyFile string `json:"api_key_file" yaml:"api_key_file" env:"SA_ZITADEL_API_KEY_FILE"`

			// ---- resource-server settings ----
			// Audience API expects in incoming access-tokens
			// Client Id of API application, can be found under Project > Client ID top, right conner
			// Dont forget to add Web App(not API only) Project Redirect Settings
			// http://localhost/auth/callback
			// http://localhost/logout/callback
			Audience string `json:"audience" yaml:"audience" env:"SA_ZITADEL_AUDIENCE"`

			// ---- browser flow ----
			RedirectURI      string   `json:"redirect_uri" yaml:"redirect_uri" env:"SA_ZITADEL_REDIRECT_URI"`
			InterceptedPaths []string `json:"intercepted_paths" yaml:"intercepted_paths" env:"SA_ZITADEL_INTERCEPTED_PATHS" envSeparator:","`
		} `json:"zitadel" yaml:"zitadel"`
	} `yaml:"auth" json:"auth"`

	// Worker settings
	Worker struct {
		// MaxCount is the maximum number of workers that can be started
		MaxCount int `yaml:"max_count" json:"max_count" env:"SA_WORKER_MAX_COUNT" envDefault:"100"`
	} `yaml:"worker" json:"worker"`

	// Queue settings for the NATS queue
	Queue struct {
		URL      string `yaml:"url" json:"url" env:"SA_QUEUE_URL" envDefault:"nats://sa-nats:4222"`
		Prefix   string `yaml:"prefix" json:"prefix" env:"SA_QUEUE_PREFIX" envDefault:"shadowapi"`
		Username string `yaml:"username" json:"username" env:"SA_QUEUE_USERNAME"`
		Password string `yaml:"password" json:"password" env:"SA_QUEUE_PASSWORD"`
	} `yaml:"queue" json:"queue"`

	// Add cfg.Telegram.AppID, cfg.Telegram.AppHash
	Telegram struct {
		AppHash string `yaml:"app_hash" json:"app_hash" env:"TG_APP_HASH" envDefault:""`
		AppID   int    `yaml:"app_id" json:"app_id" env:"TG_APP_ID" envDefault:""`
	}

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
	cfg := &Config{
		configPath: configPath,
		ext:        strings.ToLower(filepath.Ext(configPath)),
	}

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

	// env overrides values from file
	if err := env.Parse(cfg); err != nil {
		slog.Error("failed to parse environment variables", "error", err)
	}
	slog.Info("loaded config file", "path", configPath)
	slog.Info("SA_CONFIG_PATH after env parse", "value", os.Getenv("SA_CONFIG_PATH"))
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
