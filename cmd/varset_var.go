package cmd

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

var varsetVarCmd = &cobra.Command{
	Use:   "var",
	Short: "Manage variables within a variable set",
}

// ── create ───────────────────────────────────────────────────────────────

var (
	varCreateKey       string
	varCreateValue     string
	varCreateCategory  string
	varCreateSensitive bool
	varCreateHCL       bool
	varCreateDesc      string
)

var vsVarCreateCmd = &cobra.Command{
	Use:   "create <varset>",
	Short: "Add a variable to a variable set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		vsID, err := resolveVarsetID(ctx, args[0])
		if err != nil {
			return err
		}

		cat := tfe.CategoryTerraform
		if varCreateCategory == "env" {
			cat = tfe.CategoryEnv
		}

		opts := tfe.VariableSetVariableCreateOptions{
			Key:         tfe.String(varCreateKey),
			Value:       tfe.String(varCreateValue),
			Category:    tfe.Category(cat),
			Sensitive:   tfe.Bool(varCreateSensitive),
			HCL:         tfe.Bool(varCreateHCL),
			Description: tfe.String(varCreateDesc),
		}

		v, err := app.Client.VariableSetVariables.Create(ctx, vsID, &opts)
		if err != nil {
			return fmt.Errorf("creating variable: %w", err)
		}

		app.Out.Success(fmt.Sprintf("Variable %q created (ID: %s)", v.Key, v.ID))
		return nil
	},
}

// ── update ───────────────────────────────────────────────────────────────

var (
	varUpdateValue string
	varUpdateDesc  string
)

var vsVarUpdateCmd = &cobra.Command{
	Use:   "update <varset> <key>",
	Short: "Update a variable in a variable set",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		vsID, err := resolveVarsetID(ctx, args[0])
		if err != nil {
			return err
		}

		varID, err := resolveVarID(ctx, vsID, args[1])
		if err != nil {
			return err
		}

		opts := tfe.VariableSetVariableUpdateOptions{}

		if cmd.Flags().Changed("value") {
			opts.Value = tfe.String(varUpdateValue)
		}
		if cmd.Flags().Changed("sensitive") {
			s, _ := cmd.Flags().GetBool("sensitive")
			opts.Sensitive = tfe.Bool(s)
		}
		if cmd.Flags().Changed("hcl") {
			h, _ := cmd.Flags().GetBool("hcl")
			opts.HCL = tfe.Bool(h)
		}
		if cmd.Flags().Changed("description") {
			opts.Description = tfe.String(varUpdateDesc)
		}

		v, err := app.Client.VariableSetVariables.Update(ctx, vsID, varID, &opts)
		if err != nil {
			return fmt.Errorf("updating variable: %w", err)
		}

		app.Out.Success(fmt.Sprintf("Variable %q updated", v.Key))
		return nil
	},
}

// ── delete ───────────────────────────────────────────────────────────────

var vsVarDeleteCmd = &cobra.Command{
	Use:   "delete <varset> <key>",
	Short: "Remove a variable from a variable set",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		vsID, err := resolveVarsetID(ctx, args[0])
		if err != nil {
			return err
		}

		varID, err := resolveVarID(ctx, vsID, args[1])
		if err != nil {
			return err
		}

		if err := app.Client.VariableSetVariables.Delete(ctx, vsID, varID); err != nil {
			return fmt.Errorf("deleting variable: %w", err)
		}

		app.Out.Success(fmt.Sprintf("Variable %q deleted from varset %s", args[1], args[0]))
		return nil
	},
}

func init() {
	vsVarCreateCmd.Flags().StringVar(&varCreateKey, "key", "", "variable key (required)")
	vsVarCreateCmd.Flags().StringVar(&varCreateValue, "value", "", "variable value (required)")
	vsVarCreateCmd.Flags().StringVar(&varCreateCategory, "category", "terraform", "variable category: terraform or env")
	vsVarCreateCmd.Flags().BoolVar(&varCreateSensitive, "sensitive", false, "mark as sensitive")
	vsVarCreateCmd.Flags().BoolVar(&varCreateHCL, "hcl", false, "parse value as HCL")
	vsVarCreateCmd.Flags().StringVar(&varCreateDesc, "description", "", "variable description")
	vsVarCreateCmd.MarkFlagRequired("key")
	vsVarCreateCmd.MarkFlagRequired("value")

	vsVarUpdateCmd.Flags().StringVar(&varUpdateValue, "value", "", "new variable value")
	vsVarUpdateCmd.Flags().Bool("sensitive", false, "mark as sensitive")
	vsVarUpdateCmd.Flags().Bool("hcl", false, "parse value as HCL")
	vsVarUpdateCmd.Flags().StringVar(&varUpdateDesc, "description", "", "variable description")

	varsetVarCmd.AddCommand(vsVarCreateCmd, vsVarUpdateCmd, vsVarDeleteCmd)
	varsetCmd.AddCommand(varsetVarCmd)
}
