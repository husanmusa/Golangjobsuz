package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from environment variables or config files.
type Config struct {
	AppName     string
	Environment string
	Telegram    TelegramConfig
	Database    DatabaseConfig
	Metrics     MetricsConfig
	HTTP        HTTPConfig
}

// TelegramConfig contains credentials and webhook settings.
type TelegramConfig struct {
	Token      string
	WebhookURL string
}

// DatabaseConfig contains database connection details.
type DatabaseConfig struct {
	URL string
}

// MetricsConfig exposes Prometheus settings.
type MetricsConfig struct {
	Address string
}

// HTTPConfig configures outgoing HTTP client behavior.
type HTTPConfig struct {
	TimeoutSeconds int
	MaxRetries     int
}

// Load reads configuration from the provided file path if set and always overlays environment variables.
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("GJ")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("appname", "golangjobs-bot")
	v.SetDefault("environment", "development")
	v.SetDefault("telegram.token", "")
	v.SetDefault("telegram.webhookurl", "")
	v.SetDefault("database.url", "")
	v.SetDefault("metrics.address", ":9090")
	v.SetDefault("http.timeoutseconds", 30)
	v.SetDefault("http.maxretries", 2)

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
