package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultHostname = "app.terraform.io"

type Config struct {
	DefaultOrg     string `yaml:"default_org,omitempty"`
	DefaultProject string `yaml:"default_project,omitempty"`
	Hostname       string `yaml:"hostname,omitempty"`
}

// EffectiveHostname returns the configured hostname or the default.
func (c *Config) EffectiveHostname() string {
	if c.Hostname != "" {
		return c.Hostname
	}
	return defaultHostname
}

// Load reads the config file. Returns a zero-value Config if the file
// does not exist (not an error — the user just hasn't configured yet).
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Set updates a single config key and saves.
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "default_org":
		cfg.DefaultOrg = value
	case "default_project":
		cfg.DefaultProject = value
	case "hostname":
		cfg.Hostname = value
	default:
		return fmt.Errorf("unknown config key: %s (valid: default_org, default_project, hostname)", key)
	}

	return Save(cfg)
}

// Get returns the value of a single config key.
func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "default_org":
		return cfg.DefaultOrg, nil
	case "default_project":
		return cfg.DefaultProject, nil
	case "hostname":
		return cfg.EffectiveHostname(), nil
	default:
		return "", fmt.Errorf("unknown config key: %s (valid: default_org, default_project, hostname)", key)
	}
}

func configPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "tfc", "config.yaml"), nil
}
