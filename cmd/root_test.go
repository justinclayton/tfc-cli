package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNeedsClient_RootCommand(t *testing.T) {
	// The root command itself should not need a client.
	if needsClient(rootCmd) {
		t.Error("needsClient() should return false for the root command")
	}
}

func TestNeedsClient_SkipClientAnnotation(t *testing.T) {
	parent := &cobra.Command{
		Use:         "test",
		Annotations: map[string]string{"skipClient": "true"},
	}
	child := &cobra.Command{Use: "sub"}
	parent.AddCommand(child)
	rootCmd.AddCommand(parent)
	defer rootCmd.RemoveCommand(parent)

	if needsClient(child) {
		t.Error("needsClient() should return false when ancestor has skipClient annotation")
	}
	if needsClient(parent) {
		t.Error("needsClient() should return false for command with skipClient annotation")
	}
}

func TestNeedsClient_RegularChild(t *testing.T) {
	child := &cobra.Command{Use: "regular"}
	rootCmd.AddCommand(child)
	defer rootCmd.RemoveCommand(child)

	if !needsClient(child) {
		t.Error("needsClient() should return true for a regular child command without skipClient")
	}
}

func TestNeedsClient_ConfigCommand(t *testing.T) {
	// The actual configCmd has the skipClient annotation.
	if needsClient(configCmd) {
		t.Error("needsClient() should return false for config command")
	}
	// Config subcommands inherit the annotation via parent walk.
	if needsClient(configSetCmd) {
		t.Error("needsClient() should return false for config set subcommand")
	}
}

func TestRootCmd_HelpDoesNotError(t *testing.T) {
	// --help should succeed without any credentials or config.
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("--help should not error, got: %v", err)
	}
}

func TestRootCmd_Version(t *testing.T) {
	SetVersion("test-v1.2.3")
	rootCmd.Version = version
	rootCmd.SetArgs([]string{"--version"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("--version should not error, got: %v", err)
	}
	if version != "test-v1.2.3" {
		t.Errorf("SetVersion did not set version, got %q", version)
	}
}
