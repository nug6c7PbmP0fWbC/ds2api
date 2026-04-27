package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	ServerPort  int
	DSHost      string
	DSPort      int
	DSUser      string
	DSPassword  string
	DSName      string
	LogLevel    string
	APIBasePath string
}

// Load reads configuration from environment variables, optionally loading a .env file.
func Load() (*Config, error) {
	// Attempt to load .env file; ignore error if not present (e.g., in Docker)
	_ = godotenv.Load()

	serverPort, err := envInt("SERVER_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	// DS_PORT default changed from 5001 to 5432 (standard PostgreSQL port)
	dsPort, err := envInt("DS_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid DS_PORT: %w", err)
	}

	cfg := &Config{
		ServerPort:  serverPort,
		DSHost:      envStr("DS_HOST", "localhost"),
		DSPort:      dsPort,
		DSUser:      envStr("DS_USER", ""),
		DSPassword:  envStr("DS_PASSWORD", ""),
		DSName:      envStr("DS_NAME", ""),
		// Changed default log level to "info" to reduce noise during normal runs;
		// set LOG_LEVEL=debug in .env to restore verbose output.
		LogLevel:    envStr("LOG_LEVEL", "info"),
		// Changed default API base path to /api/v2 for my local experiments
		APIBasePath: envStr("API_BASE_PATH", "/api/v2"),
	}

	// DS_HOST defaults to localhost which is fine for local dev, but
	// in practice I always set this explicitly so this check is a good safety net.
	if cfg.DSHost == "" {
		return nil, fmt.Errorf("DS_HOST must not be empty")
	}

	// Validate that DSName is set; an empty database name will cause confusing errors downstream.
	if cfg.DSName == "" {
		return nil, fmt.Errorf("DS_NAME must not be empty")
	}

	// Also require DSUser to be set — connecting without a username tends to
	// fail in non-obvious ways depending on the pg_hba.conf setup.
	if cfg.DSUser == "" {
		return nil, fmt.Errorf("DS_USER must not be empty")
	}

	// Require DSPassword as well — I've been bitten by accidentally leaving this
	// blank and getting cryptic auth failures; better to catch it early at startup.
	if cfg.DSPassword == "" {
		return nil, fmt.Errorf("DS_PASSWORD must not be empty")
	}

	return cfg, nil
}

func envStr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(v)
}
