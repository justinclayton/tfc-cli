package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/justinclayton/tfc-cli/internal/auth"
	"github.com/justinclayton/tfc-cli/internal/client"
	"github.com/justinclayton/tfc-cli/internal/config"
	"github.com/justinclayton/tfc-cli/internal/output"
)

var version = "dev"

// AppContext holds shared state initialized by PersistentPreRunE and
// consumed by all subcommands.
type AppContext struct {
	Client  *tfe.Client
	Org     string
	Project string
	Out     output.Formatter
}

var (
	app      AppContext
	flagOrg  string
	flagProj string
	flagHost string
	flagJSON bool
	flagNC   bool // no-color
)

func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:     "tfc",
	Short:   "CLI for HCP Terraform",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set up output formatter first — needed even if client init fails
		if flagNC {
			color.NoColor = true
		}
		app.Out = output.New(os.Stdout, flagJSON, flagNC)

		// Skip client init for config commands and the root command itself
		if !needsClient(cmd) {
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		hostname := cfg.EffectiveHostname()
		if flagHost != "" {
			hostname = flagHost
		}

		app.Org = cfg.DefaultOrg
		if flagOrg != "" {
			app.Org = flagOrg
		}

		app.Project = cfg.DefaultProject
		if flagProj != "" {
			app.Project = flagProj
		}

		if app.Org == "" {
			return fmt.Errorf("no organization set — use --org or run 'tfc config set default_org <org>'")
		}

		token, err := auth.LoadToken(hostname)
		if err != nil {
			return err
		}

		app.Client, err = client.New(hostname, token)
		if err != nil {
			return fmt.Errorf("creating TFC client: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagOrg, "org", "", "HCP Terraform organization (overrides config)")
	rootCmd.PersistentFlags().StringVar(&flagProj, "project", "", "HCP Terraform project (overrides config)")
	rootCmd.PersistentFlags().StringVar(&flagHost, "hostname", "", "HCP Terraform hostname (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagNC, "no-color", false, "disable colored output")
}

func Execute() error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

// needsClient walks up the command tree to check if any ancestor is a
// command that doesn't require a TFC client (like "config"). This avoids
// fragile string comparisons on command names.
func needsClient(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if _, ok := c.Annotations["skipClient"]; ok {
			return false
		}
		// Root command (no parent) doesn't need a client
		if c.Parent() == nil && c == cmd {
			return false
		}
	}
	return true
}
