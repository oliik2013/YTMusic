package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// AuthConfig holds optional pre-seeded authentication credentials.
type AuthConfig struct {
	Cookies string `yaml:"cookies"`
}

// Config is the top-level configuration loaded from ~/.ytmusic/config.yaml.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Auth   AuthConfig   `yaml:"auth"`
}

// DefaultConfig returns a config with sane defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
	}
}

// ConfigPath returns the default config file path (~/.ytmusic/config.yaml).
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(home, ".ytmusic", "config.yaml"), nil
}

// LoadConfig reads config from ~/.ytmusic/config.yaml.
// If the file does not exist, it creates one with defaults and returns it.
func LoadConfig() (*Config, error) {
	cfgPath, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			// Create the config directory and a default config file
			if mkErr := os.MkdirAll(filepath.Dir(cfgPath), 0755); mkErr != nil {
				return cfg, nil // silently use defaults
			}
			if writeErr := SaveConfig(cfg, cfgPath); writeErr != nil {
				return cfg, nil
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return cfg, nil
}

// SaveConfig writes the config to the given path.
func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
