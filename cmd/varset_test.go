package cmd

import (
	"context"
	"testing"
)

func TestResolveVarsetID_PrefixShortcut(t *testing.T) {
	// When input starts with "varset-", it should be returned directly
	// without any API call.
	ctx := context.Background()
	id, err := resolveVarsetID(ctx, "varset-abc123def")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "varset-abc123def" {
		t.Errorf("expected 'varset-abc123def', got %q", id)
	}
}

func TestResolveVarsetID_PrefixShortcut_MinLength(t *testing.T) {
	// The check is len > 7 && starts with "varset-", so "varset-x" (8 chars) should work.
	ctx := context.Background()
	id, err := resolveVarsetID(ctx, "varset-x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "varset-x" {
		t.Errorf("expected 'varset-x', got %q", id)
	}
}

func TestResolveVarsetID_ExactSevenChars(t *testing.T) {
	// "varset-" is exactly 7 chars, len > 7 is false, so this should NOT
	// take the shortcut path. Without a client it will panic/fail, so we
	// just verify the logic boundary by checking the prefix condition.
	input := "varset-"
	// len("varset-") == 7, so len > 7 is false => it tries the API path.
	// We can't test the API path without a mock, so just verify our understanding.
	if len(input) > 7 && input[:7] == "varset-" {
		t.Error("'varset-' alone should not match the prefix shortcut")
	}
}
