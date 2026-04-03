package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/justinclayton/tfc-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:         "config",
	Short:       "Manage tfc configuration",
	Annotations: map[string]string{"skipClient": "true"},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive first-time configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("HCP Terraform hostname [%s]: ", defaultVal(cfg.EffectiveHostname(), "app.terraform.io"))
		hostname := readLine(reader)
		if hostname != "" {
			cfg.Hostname = hostname
		}

		fmt.Printf("Default organization [%s]: ", cfg.DefaultOrg)
		org := readLine(reader)
		if org != "" {
			cfg.DefaultOrg = org
		}

		fmt.Printf("Default project [%s]: ", cfg.DefaultProject)
		proj := readLine(reader)
		if proj != "" {
			cfg.DefaultProject = proj
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println("Configuration saved.")
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  "Valid keys: default_org, default_project, hostname",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Set(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", args[0], args[1])
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Show configuration values",
	Long:  "Valid keys: default_org, default_project, hostname. Omit key to show all.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			fmt.Printf("hostname:        %s\n", cfg.EffectiveHostname())
			fmt.Printf("default_org:     %s\n", cfg.DefaultOrg)
			fmt.Printf("default_project: %s\n", cfg.DefaultProject)
			return nil
		}

		val, err := config.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd, configSetCmd, configGetCmd)
	rootCmd.AddCommand(configCmd)
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func defaultVal(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}
