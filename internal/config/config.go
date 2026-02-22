package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Server_host string
	Server_port string

	// Database settings
	Database_url string

	// Cloud provider settings
	AWS_region         string
	GCP_project        string
	AWS_access_key     string
	AWS_secret_key     string
	Discovery_enabled  bool
	Discovery_interval int // Discovery interval in seconds

	// Notification settings
	Slack_webhook_url string
	Slack_app_token   string
	Pre_stop_minutes  int

	// Aggregation settings
	Aggregation_enabled   bool
	Aggregation_threshold int
	Aggregation_delay     int

	// Cost settings
	Currency string
}

// Load reads configuration from environment variables and returns a Config
func Load() (*Config, error) {
	cfg := &Config{}

	// Server settings (required)
	cfg.Server_host = getEnv("SERVER_HOST", "0.0.0.0")
	cfg.Server_port = getEnv("SERVER_PORT", "8080")

	// Database settings (required)
	cfg.Database_url = getEnv("DATABASE_URL", "postgresql://localhost:5432/snoozeql?sslmode=disable")

	// Cloud provider settings
	cfg.AWS_region = getEnv("AWS_REGION", "us-east-1")
	cfg.GCP_project = getEnv("GCP_PROJECT", "")
	cfg.AWS_access_key = getEnv("AWS_ACCESS_KEY_ID", "")
	cfg.AWS_secret_key = getEnv("AWS_SECRET_ACCESS_KEY", "")
	cfg.Discovery_enabled = getEnvBool("DISCOVERY_ENABLED", true)
	cfg.Discovery_interval = getEnvInt("DISCOVERY_INTERVAL_SECONDS", 30)

	// Notification settings
	cfg.Slack_webhook_url = getEnv("SLACK_WEBHOOK_URL", "")
	cfg.Slack_app_token = getEnv("SLACK_APP_TOKEN", "")
	cfg.Pre_stop_minutes = getEnvInt("PRE_STOP_MINUTES", 10)

	// Aggregation settings
	cfg.Aggregation_enabled = getEnvBool("AGGREGATION_ENABLED", true)
	cfg.Aggregation_threshold = getEnvInt("AGGREGATION_THRESHOLD", 2)
	cfg.Aggregation_delay = getEnvInt("AGGREGATION_DELAY_SECONDS", 30)

	// Cost settings
	cfg.Currency = getEnv("CURRENCY", "USD")

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.ParseBool(value); err == nil {
			return v
		}
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	if c.Server_port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	if c.Database_url == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}
