/*
Copyright © 2023 Shadowapi <support@shadowapi.com>
*/
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

	// Auth is a struct that holds all the authentication settings
	Auth struct {

		// BearerToken is used to validate incoming requests that carry an Authorization header.
		BearerToken string `yaml:"bearer_token" json:"bearer_token" env:"SA_AUTH_BEARER_TOKEN" envDefault:"mysecretapikey"`

		// Zitadel configuration for OAuth2 authentication
		Zitadel struct {
			InstanceURL      string   `yaml:"instance_url" json:"instance_url" env:"SA_AUTH_ZITADEL_INSTANCE_URL"`
			ClientID         string   `yaml:"client_id" json:"client_id" env:"SA_AUTH_ZITADEL_CLIENT_ID"`
			ClientSecret     string   `yaml:"client_secret" json:"client_secret" env:"SA_AUTH_ZITADEL_CLIENT_SECRET"`
			RedirectURI      string   `yaml:"redirect_uri" json:"redirect_uri" env:"SA_AUTH_ZITADEL_REDIRECT_URI"`
			InterceptedPaths []string `yaml:"intercepted_paths" json:"intercepted_paths" env:"SA_AUTH_ZITADEL_INTERCEPTED_PATHS" envSeparator:","`
		} `yaml:"zitadel" json:"zitadel"`
	} `yaml:"auth" json:"auth"`

	// Worker settings
	Worker struct {
		// MaxCount is the maximum number of workers that can be started
		MaxCount int `yaml:"max_count" json:"max_count" env:"SA_WORKER_MAX_COUNT" envDefault:"100"`
	} `yaml:"worker" json:"worker"`

	// Queue settings for the NATS queue
	Queue struct {
		URL    string `yaml:"url" json:"url" env:"SA_QUEUE_URL" envDefault:"nats://sa-nats:4222"`
		Prefix string `yaml:"prefix" json:"prefix" env:"SA_QUEUE_PREFIX" envDefault:"shadowapi"`
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
	return Load(cPath)
}

// Load creates a new Config instance
func Load(configPath string) (*Config, error) {
	config := &Config{
		configPath: configPath,
		ext:        strings.ToLower(filepath.Ext(configPath)),
	}
	defer func() {
		if err := env.Parse(config); err != nil {
			slog.Error("failed to parse environment variables", "error", err)
		}
	}()

	stat, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := config.Save(); err != nil {
				return nil, err
			}
		}
	} else if stat.IsDir() {
		return nil, fmt.Errorf("config file path is a directory")
	} else {
		if err := config.Load(); err != nil {
			return nil, err
		}
		slog.Info("loaded config file", "path", configPath)
	}

	return config, nil
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
		if err := yaml.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	default:
		return fmt.Errorf("unknown config file format: %s", c.ext)
	}
	return nil
}

// Save saves the configuration to the config file
func (c *Config) Save() (err error) {
	if c.ext == "" {
		return fmt.Errorf("config file extension is empty")
	}

	var data []byte
	switch c.ext {
	case ".yaml", ".yml":
		if data, err = yaml.Marshal(c); err != nil {
			return fmt.Errorf("failed to marshal config file: %w", err)
		}
	case ".json":
		if data, err = json.MarshalIndent(c, "", "  "); err != nil {
			return fmt.Errorf("failed to marshal config file: %w", err)
		}
	default:
		return fmt.Errorf("unknown config file format: %s", c.ext)
	}

	return os.WriteFile(c.configPath, data, 0o644)
}
