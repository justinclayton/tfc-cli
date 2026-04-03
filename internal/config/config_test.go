package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Load on empty dir returns zero-value config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.DefaultOrg != "" {
		t.Fatalf("expected empty default_org, got %q", cfg.DefaultOrg)
	}

	// Save and reload
	cfg.DefaultOrg = "myorg"
	cfg.DefaultProject = "myproj"
	cfg.Hostname = "tfe.example.com"
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error: %v", err)
	}
	if cfg2.DefaultOrg != "myorg" {
		t.Errorf("expected default_org=myorg, got %q", cfg2.DefaultOrg)
	}
	if cfg2.DefaultProject != "myproj" {
		t.Errorf("expected default_project=myproj, got %q", cfg2.DefaultProject)
	}
	if cfg2.Hostname != "tfe.example.com" {
		t.Errorf("expected hostname=tfe.example.com, got %q", cfg2.Hostname)
	}
}

func TestEffectiveHostname(t *testing.T) {
	cfg := &Config{}
	if got := cfg.EffectiveHostname(); got != "app.terraform.io" {
		t.Errorf("expected default hostname, got %q", got)
	}
	cfg.Hostname = "custom.tfe.io"
	if got := cfg.EffectiveHostname(); got != "custom.tfe.io" {
		t.Errorf("expected custom hostname, got %q", got)
	}
}

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := Set("default_org", "testorg"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	val, err := Get("default_org")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if val != "testorg" {
		t.Errorf("expected testorg, got %q", val)
	}

	// Unknown key
	if err := Set("bogus", "x"); err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestConfigFileCreated(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{DefaultOrg: "org"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	path := filepath.Join(dir, "tfc", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("config file not created at %s", path)
	}
}
