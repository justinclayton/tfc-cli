package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type credentialsFile struct {
	Credentials map[string]credential `json:"credentials"`
}

type credential struct {
	Token string `json:"token"`
}

// LoadToken reads the Terraform CLI credentials file and returns the token
// for the given hostname. It checks TF_CLI_CONFIG_FILE first, then falls
// back to ~/.terraform.d/credentials.tfrc.json.
func LoadToken(hostname string) (string, error) {
	path, err := credentialsPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("credentials file not found at %s\nRun 'terraform login' to authenticate", path)
		}
		return "", fmt.Errorf("reading credentials file: %w", err)
	}

	var creds credentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parsing credentials file: %w", err)
	}

	cred, ok := creds.Credentials[hostname]
	if !ok {
		return "", fmt.Errorf("no credentials found for %s\nRun 'terraform login %s' to authenticate", hostname, hostname)
	}

	if cred.Token == "" {
		return "", fmt.Errorf("empty token for %s in credentials file", hostname)
	}

	return cred.Token, nil
}

func credentialsPath() (string, error) {
	if p := os.Getenv("TF_CLI_CONFIG_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".terraform.d", "credentials.tfrc.json"), nil
}
