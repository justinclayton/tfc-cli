package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.tfrc.json")

	content := `{
  "credentials": {
    "app.terraform.io": {
      "token": "test-token-123"
    },
    "tfe.example.com": {
      "token": "other-token-456"
    }
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TF_CLI_CONFIG_FILE", path)

	token, err := LoadToken("app.terraform.io")
	if err != nil {
		t.Fatalf("LoadToken() error: %v", err)
	}
	if token != "test-token-123" {
		t.Errorf("expected test-token-123, got %q", token)
	}

	token, err = LoadToken("tfe.example.com")
	if err != nil {
		t.Fatalf("LoadToken() error: %v", err)
	}
	if token != "other-token-456" {
		t.Errorf("expected other-token-456, got %q", token)
	}

	_, err = LoadToken("missing.example.com")
	if err == nil {
		t.Error("expected error for missing hostname")
	}
}

func TestLoadTokenMissingFile(t *testing.T) {
	t.Setenv("TF_CLI_CONFIG_FILE", "/nonexistent/path/credentials.tfrc.json")

	_, err := LoadToken("app.terraform.io")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
