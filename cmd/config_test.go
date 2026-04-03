package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/justinclayton/tfc-cli/internal/config"
	"github.com/justinclayton/tfc-cli/internal/output"
)

func TestConfigSetGetRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Set a value via the config package (same path the command uses)
	if err := config.Set("default_org", "roundtrip-org"); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	val, err := config.Get("default_org")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "roundtrip-org" {
		t.Errorf("expected roundtrip-org, got %q", val)
	}
}

func TestConfigGetAll(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := config.Set("default_org", "org1"); err != nil {
		t.Fatal(err)
	}
	if err := config.Set("default_project", "proj1"); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultOrg != "org1" {
		t.Errorf("expected org1, got %q", cfg.DefaultOrg)
	}
	if cfg.DefaultProject != "proj1" {
		t.Errorf("expected proj1, got %q", cfg.DefaultProject)
	}
	// Hostname should return the default when not set
	if cfg.EffectiveHostname() != "app.terraform.io" {
		t.Errorf("expected default hostname, got %q", cfg.EffectiveHostname())
	}
}

func TestConfigSetInvalidKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	err := config.Set("nonexistent_key", "value")
	if err == nil {
		t.Fatal("expected error for invalid config key")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("expected 'unknown config key' error, got: %v", err)
	}
}

func TestConfigCommands_SkipClient(t *testing.T) {
	// Config commands should have the skipClient annotation so they work without credentials.
	if needsClient(configCmd) {
		t.Error("config command should not need client")
	}
	if needsClient(configSetCmd) {
		t.Error("config set should not need client")
	}
	if needsClient(configGetCmd) {
		t.Error("config get should not need client")
	}
	if needsClient(configInitCmd) {
		t.Error("config init should not need client")
	}
}

func TestConfigSetCmd_Execute(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Set up the formatter so PersistentPreRunE doesn't fail
	var buf bytes.Buffer
	app.Out = output.New(&buf, false, true)

	rootCmd.SetArgs([]string{"config", "set", "default_org", "cli-org"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	val, err := config.Get("default_org")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "cli-org" {
		t.Errorf("expected cli-org, got %q", val)
	}
}

func TestConfigGetCmd_Execute(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := config.Set("default_org", "get-org"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	app.Out = output.New(&buf, false, true)

	rootCmd.SetArgs([]string{"config", "get", "default_org"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config get failed: %v", err)
	}
}

func TestConfigSetCmd_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	var buf bytes.Buffer
	app.Out = output.New(&buf, false, true)

	rootCmd.SetArgs([]string{"config", "set", "bogus_key", "val"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid key via CLI")
	}
}
