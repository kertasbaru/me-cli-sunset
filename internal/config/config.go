// Package config provides application configuration loaded from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds all configuration values for the application.
type Config struct {
	// Server settings
	ServerPort string

	// API endpoints
	BaseAPIURL  string
	BaseCIAMURL string

	// Authentication
	BasicAuth string

	// Device fingerprint
	AXFPKey string

	// HTTP client
	UserAgent string

	// API keys and secrets
	APIKey            string
	EncryptedFieldKey string
	XDataKey          string
	AXAPISigKey       string
	XAPIBaseSecret    string
	CircleMSISDNKey   string

	// Database
	DatabasePath string
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:        getEnvOrDefault("SERVER_PORT", "8080"),
		BaseAPIURL:        os.Getenv("BASE_API_URL"),
		BaseCIAMURL:       os.Getenv("BASE_CIAM_URL"),
		BasicAuth:         os.Getenv("BASIC_AUTH"),
		AXFPKey:           os.Getenv("AX_FP_KEY"),
		UserAgent:         os.Getenv("UA"),
		APIKey:            os.Getenv("API_KEY"),
		EncryptedFieldKey: os.Getenv("ENCRYPTED_FIELD_KEY"),
		XDataKey:          os.Getenv("XDATA_KEY"),
		AXAPISigKey:       os.Getenv("AX_API_SIG_KEY"),
		XAPIBaseSecret:    os.Getenv("X_API_BASE_SECRET"),
		CircleMSISDNKey:   os.Getenv("CIRCLE_MSISDN_KEY"),
		DatabasePath:      getEnvOrDefault("DATABASE_PATH", "me_cli.db"),
	}

	if cfg.BaseAPIURL == "" {
		return nil, fmt.Errorf("BASE_API_URL environment variable is required")
	}
	if cfg.BaseCIAMURL == "" {
		return nil, fmt.Errorf("BASE_CIAM_URL environment variable is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
