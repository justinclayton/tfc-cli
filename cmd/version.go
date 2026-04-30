package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:         "version",
	Short:       "Print the tfc version",
	Annotations: map[string]string{"skipClient": "true"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "tfc version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
