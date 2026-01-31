package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected server host 0.0.0.0, got %s", cfg.Server.Host)
	}

	if cfg.Database.Path != "./gateway.db" {
		t.Errorf("Expected database path ./gateway.db, got %s", cfg.Database.Path)
	}

	if cfg.HealthCheck.Interval != 30 {
		t.Errorf("Expected health check interval 30, got %d", cfg.HealthCheck.Interval)
	}

	if cfg.Session.IdleTimeout != 30 {
		t.Errorf("Expected session idle timeout 30, got %d", cfg.Session.IdleTimeout)
	}

	if !cfg.Metrics.Enabled {
		t.Error("Expected metrics to be enabled")
	}

	if cfg.Metrics.Port != 9090 {
		t.Errorf("Expected metrics port 9090, got %d", cfg.Metrics.Port)
	}
}

func TestService(t *testing.T) {
	svc, err := NewService("/tmp/test_config.yaml")
	if err != nil {
		t.Fatalf("Failed to create config service: %v", err)
	}

	cfg := svc.Get()
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}
}

func TestValidate(t *testing.T) {
	svc, err := NewService("/tmp/test_config2.yaml")
	if err != nil {
		t.Fatalf("Failed to create config service: %v", err)
	}

	// Valid config should pass
	if err := svc.Validate(); err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}
}
