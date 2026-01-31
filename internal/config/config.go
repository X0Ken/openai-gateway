package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	_ "gopkg.in/yaml.v3"
)

// Config holds all configuration for the gateway
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Session     SessionConfig     `yaml:"session"`
	Metrics     MetricsConfig     `yaml:"metrics"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Interval int `yaml:"interval"`
	Timeout  int `yaml:"timeout"`
}

// SessionConfig holds session management configuration
type SessionConfig struct {
	IdleTimeout int `yaml:"idle_timeout"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Database: DatabaseConfig{
			Path: "./gateway.db",
		},
		HealthCheck: HealthCheckConfig{
			Interval: 30,
			Timeout:  5,
		},
		Session: SessionConfig{
			IdleTimeout: 30,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    9090,
		},
	}
}

// Service manages configuration with hot reload
type Service struct {
	mu      sync.RWMutex
	config  *Config
	path    string
	watcher *fsnotify.Watcher
}

// NewService creates a new configuration service
func NewService(path string) (*Service, error) {
	svc := &Service{
		path: path,
	}

	// Load initial config
	if err := svc.Load(); err != nil {
		return nil, err
	}

	return svc, nil
}

// Load reads configuration from file
func (s *Service) Load() error {
	// For now, return default config
	// In production, this would read from YAML file
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = DefaultConfig()
	return nil
}

// Get returns the current configuration
func (s *Service) Get() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Validate checks if the configuration is valid
func (s *Service) Validate() error {
	cfg := s.Get()

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if cfg.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	return nil
}
