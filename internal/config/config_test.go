package config

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for k, v := range pairs {
		t.Setenv(k, v)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Ensure required DS_HOST is set; everything else uses defaults.
	setEnv(t, map[string]string{"DS_HOST": "synology.local"})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerPort != 8080 {
		t.Errorf("expected ServerPort 8080, got %d", cfg.ServerPort)
	}
	// DS_PORT default is 5000 (HTTP) rather than 5001 (HTTPS) for my local setup
	if cfg.DSPort != 5000 {
		t.Errorf("expected DSPort 5000, got %d", cfg.DSPort)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %s", cfg.LogLevel)
	}
	if cfg.APIBasePath != "/api/v1" {
		t.Errorf("expected APIBasePath '/api/v1', got %s", cfg.APIBasePath)
	}
	// Verify DSName defaults to empty string when not set
	if cfg.DSName != "" {
		t.Errorf("expected DSName to be empty by default, got %s", cfg.DSName)
	}
	// Verify DSHost is correctly stored from env
	if cfg.DSHost != "synology.local" {
		t.Errorf("expected DSHost 'synology.local', got %s", cfg.DSHost)
	}
	// Verify DSUser defaults to empty string when not set
	if cfg.DSUser != "" {
		t.Errorf("expected DSUser to be empty by default, got %s", cfg.DSUser)
	}
	// Verify DSPassword defaults to empty string when not set
	if cfg.DSPassword != "" {
		t.Errorf("expected DSPassword to be empty by default, got %s", cfg.DSPassword)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setEnv(t, map[string]string{
		"SERVER_PORT":   "9090",
		"DS_HOST":       "nas.home",
		"DS_PORT":       "5002",
		"DS_USER":       "admin",
		"DS_PASSWORD":   "secret",
		"DS_NAME":       "mynas",
		"LOG_LEVEL":     "debug",
		"API_BASE_PATH": "/api/v2",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerPort != 9090 {
		t.Errorf("expected ServerPort 9090, got %d", cfg.ServerPort)
	}
	if cfg.DSHost != "nas.home" {
		t.Errorf("expected DSHost 'nas.home', got %s", cfg.DSHost)
	}
	if cfg.DSUser != "admin" {
		t.Errorf("expected DSUser 'admin', got %s", cfg.DSUser)
	}
	// Also verify DSName is picked up correctly
	if cfg.DSName != "mynas" {
		t.Errorf("expected DSName 'mynas', got %s", cfg.DSName)
	}
	// Verify DSPort custom value is loaded
	if cfg.DSPort != 5002 {
		t.Errorf("expected DSPort 5002, got %d", cfg.DSPort)
	}
	// Verify DSPassword custom value is loaded
	if cfg.DSPassword != "secret" {
		t.Errorf("expected DSPassword 'secret', got %s", cfg.DSPassword)
	}
}

func TestLoad_MissingDSHost(t *testing.T) {
	os.Unsetenv("DS_HOST")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DS_HOST is missing, got nil")
	}
}

func TestLoad_InvalidServerPort(t *testing.T) {
	setEnv(t, map[string]string{
		"DS_HOST":     "nas.home",
		"SERVER_PORT": "not-a-number",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid SERVER_PORT, got nil")
	}
}

func TestLoad_InvalidDSPort(t *testing.T) {
	// Ensure an invalid DS_PORT value also returns an error
	setEnv(t, map[string]string{
		"DS_HOST": "nas.home",
		"DS_PORT": "not-a-number",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid DS_PORT, got nil")
	}
}
