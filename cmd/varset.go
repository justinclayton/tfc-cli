package cmd

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

var varsetCmd = &cobra.Command{
	Use:     "varset",
	Aliases: []string{"vs"},
	Short:   "Manage variable sets",
}

// ── list ─────────────────────────────────────────────────────────────────

var vsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List variable sets",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		opts := &tfe.VariableSetListOptions{
			ListOptions: tfe.ListOptions{PageSize: 100},
		}

		headers := []string{"NAME", "ID", "GLOBAL", "DESCRIPTION"}
		var rows [][]string

		for {
			list, err := app.Client.VariableSets.List(ctx, app.Org, opts)
			if err != nil {
				return fmt.Errorf("listing variable sets: %w", err)
			}

			for _, vs := range list.Items {
				desc := vs.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				rows = append(rows, []string{
					vs.Name,
					vs.ID,
					fmt.Sprintf("%t", vs.Global),
					desc,
				})
			}

			if list.NextPage == 0 {
				break
			}
			opts.PageNumber = list.NextPage
		}

		app.Out.Table(headers, rows)
		return nil
	},
}

// ── show ─────────────────────────────────────────────────────────────────

var vsShowCmd = &cobra.Command{
	Use:   "show <varset-name-or-id>",
	Short: "Show variables in a variable set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		vsID, err := resolveVarsetID(ctx, args[0])
		if err != nil {
			return err
		}

		opts := &tfe.VariableSetVariableListOptions{
			ListOptions: tfe.ListOptions{PageSize: 100},
		}

		headers := []string{"KEY", "VALUE", "CATEGORY", "SENSITIVE", "HCL"}
		var rows [][]string

		for {
			list, err := app.Client.VariableSetVariables.List(ctx, vsID, opts)
			if err != nil {
				return fmt.Errorf("listing variables: %w", err)
			}

			for _, v := range list.Items {
				value := v.Value
				if v.Sensitive {
					value = "(sensitive)"
				}
				rows = append(rows, []string{
					v.Key,
					value,
					string(v.Category),
					fmt.Sprintf("%t", v.Sensitive),
					fmt.Sprintf("%t", v.HCL),
				})
			}

			if list.NextPage == 0 {
				break
			}
			opts.PageNumber = list.NextPage
		}

		app.Out.Table(headers, rows)
		return nil
	},
}

func init() {
	varsetCmd.AddCommand(vsListCmd, vsShowCmd)
	rootCmd.AddCommand(varsetCmd)
}

// resolveVarsetID takes a name or ID and returns the variable set ID.
// If the input looks like a TFC ID (starts with "varset-"), it's used directly.
// Otherwise, it searches by name.
func resolveVarsetID(ctx context.Context, nameOrID string) (string, error) {
	if len(nameOrID) > 7 && nameOrID[:7] == "varset-" {
		return nameOrID, nil
	}

	opts := &tfe.VariableSetListOptions{
		ListOptions: tfe.ListOptions{PageSize: 100},
	}
	for {
		list, err := app.Client.VariableSets.List(ctx, app.Org, opts)
		if err != nil {
			return "", fmt.Errorf("listing variable sets: %w", err)
		}
		for _, vs := range list.Items {
			if vs.Name == nameOrID {
				return vs.ID, nil
			}
		}
		if list.NextPage == 0 {
			break
		}
		opts.PageNumber = list.NextPage
	}
	return "", fmt.Errorf("variable set %q not found in org %s", nameOrID, app.Org)
}

// resolveVarID finds a variable in a varset by key and returns its ID.
func resolveVarID(ctx context.Context, vsID, key string) (string, error) {
	opts := &tfe.VariableSetVariableListOptions{
		ListOptions: tfe.ListOptions{PageSize: 100},
	}
	for {
		list, err := app.Client.VariableSetVariables.List(ctx, vsID, opts)
		if err != nil {
			return "", fmt.Errorf("listing variables: %w", err)
		}
		for _, v := range list.Items {
			if v.Key == key {
				return v.ID, nil
			}
		}
		if list.NextPage == 0 {
			break
		}
		opts.PageNumber = list.NextPage
	}
	return "", fmt.Errorf("variable %q not found in varset %s", key, vsID)
}

