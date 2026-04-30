package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	SetVersion("test-v9.9.9")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("version command should not error, got: %v", err)
	}

	got := buf.String()
	want := "tfc version test-v9.9.9"
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q, got %q", want, got)
	}
}
