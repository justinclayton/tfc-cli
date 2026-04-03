package client

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

// New creates a go-tfe client for the given hostname and token.
func New(hostname, token string) (*tfe.Client, error) {
	cfg := &tfe.Config{
		Address: fmt.Sprintf("https://%s", hostname),
		Token:   token,
	}
	return tfe.NewClient(cfg)
}
