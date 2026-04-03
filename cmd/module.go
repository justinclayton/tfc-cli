package cmd

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/justinclayton/tfc-cli/internal/output"
)

var moduleCmd = &cobra.Command{
	Use:     "module",
	Aliases: []string{"mod"},
	Short:   "Browse and provision private registry modules",
}

// ── list ─────────────────────────────────────────────────────────────────

var modListCmd = &cobra.Command{
	Use:   "list",
	Short: "List private registry modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		opts := &tfe.RegistryModuleListOptions{
			ListOptions: tfe.ListOptions{PageSize: 100},
		}

		headers := []string{"NAME", "PROVIDER", "STATUS", "NO-CODE", "UPDATED"}
		var rows [][]string

		for {
			list, err := app.Client.RegistryModules.List(ctx, app.Org, opts)
			if err != nil {
				return fmt.Errorf("listing modules: %w", err)
			}

			for _, m := range list.Items {
				rows = append(rows, []string{
					m.Name,
					m.Provider,
					string(m.Status),
					fmt.Sprintf("%t", m.NoCode),
					m.UpdatedAt,
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

var modShowProvider string

var modShowCmd = &cobra.Command{
	Use:   "show <module-name>",
	Short: "Show module details and versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		modID := tfe.RegistryModuleID{
			Organization: app.Org,
			Name:         args[0],
			Provider:     modShowProvider,
			Namespace:    app.Org,
			RegistryName: tfe.PrivateRegistry,
		}

		m, err := app.Client.RegistryModules.Read(ctx, modID)
		if err != nil {
			return fmt.Errorf("reading module: %w", err)
		}

		fields := []output.Field{
			{Label: "Name", Value: m.Name},
			{Label: "ID", Value: m.ID},
			{Label: "Provider", Value: m.Provider},
			{Label: "Status", Value: string(m.Status)},
			{Label: "No-Code", Value: fmt.Sprintf("%t", m.NoCode)},
			{Label: "Updated", Value: m.UpdatedAt},
		}

		if len(m.VersionStatuses) > 0 {
			var versions []string
			for _, vs := range m.VersionStatuses {
				versions = append(versions, fmt.Sprintf("%s (%s)", vs.Version, vs.Status))
			}
			fields = append(fields, output.Field{Label: "Versions", Value: strings.Join(versions, ", ")})
		}

		app.Out.Detail(fields)
		return nil
	},
}

// ── provision ────────────────────────────────────────────────────────────

var (
	provisionName     string
	provisionVars     []string
	provisionProvider string
)

var modProvisionCmd = &cobra.Command{
	Use:   "provision <module-name>",
	Short: "Provision a workspace from a no-code module",
	Long: `Creates a new workspace from a no-code enabled private registry module.

In interactive mode, you'll be prompted for any module variables.
In non-interactive mode (or piped), provide all variables via --var flags.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		moduleName := args[0]

		// Read the module to find its no-code module ID
		modID := tfe.RegistryModuleID{
			Organization: app.Org,
			Name:         moduleName,
			Provider:     provisionProvider,
			Namespace:    app.Org,
			RegistryName: tfe.PrivateRegistry,
		}

		m, err := app.Client.RegistryModules.Read(ctx, modID)
		if err != nil {
			return fmt.Errorf("reading module: %w", err)
		}

		if !m.NoCode {
			return fmt.Errorf("module %q does not have no-code provisioning enabled", moduleName)
		}

		// Find the no-code module ID from the relation
		if len(m.RegistryNoCodeModule) == 0 {
			return fmt.Errorf("module %q has no-code flag set but no associated no-code module found", moduleName)
		}
		noCodeModuleID := m.RegistryNoCodeModule[0].ID

		// Read variable options
		ncm, err := app.Client.RegistryNoCodeModules.Read(ctx, noCodeModuleID, &tfe.RegistryNoCodeModuleReadOptions{
			Include: []tfe.RegistryNoCodeModuleIncludeOpt{"variable-options"},
		})
		if err != nil {
			return fmt.Errorf("reading no-code module config: %w", err)
		}

		// Parse --var flags
		varMap := make(map[string]string)
		for _, v := range provisionVars {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --var format %q (expected key=value)", v)
			}
			varMap[parts[0]] = parts[1]
		}

		// Prompt for missing variables if interactive
		if isInteractive() && ncm.VariableOptions != nil {
			reader := bufio.NewReader(os.Stdin)
			for _, vo := range ncm.VariableOptions {
				if _, ok := varMap[vo.VariableName]; ok {
					continue
				}

				prompt := fmt.Sprintf("  %s", vo.VariableName)
				if len(vo.Options) > 0 {
					prompt += fmt.Sprintf(" [%s]", strings.Join(vo.Options, ", "))
				}
				prompt += ": "

				fmt.Fprint(os.Stderr, prompt)
				line, _ := reader.ReadString('\n')
				line = strings.TrimSpace(line)
				if line != "" {
					varMap[vo.VariableName] = line
				}
			}
		}

		// Generate workspace name
		wsName := provisionName
		if wsName == "" {
			wsName = fmt.Sprintf("%s-%s", moduleName, randomSuffix())
		}

		// Build variables
		var tfVars []*tfe.Variable
		for k, v := range varMap {
			tfVars = append(tfVars, &tfe.Variable{
				Key:      k,
				Value:    v,
				Category: tfe.CategoryTerraform,
			})
		}

		createOpts := &tfe.RegistryNoCodeModuleCreateWorkspaceOptions{
			Name:      wsName,
			Variables: tfVars,
		}

		if app.Project != "" {
			// Resolve project ID from name
			projID, err := resolveProjectID(ctx, app.Project)
			if err != nil {
				return err
			}
			createOpts.Project = &tfe.Project{ID: projID}
		}

		ws, err := app.Client.RegistryNoCodeModules.CreateWorkspace(ctx, noCodeModuleID, createOpts)
		if err != nil {
			return fmt.Errorf("provisioning workspace: %w", err)
		}

		app.Out.Detail([]output.Field{
			{Label: "Workspace", Value: ws.Name},
			{Label: "ID", Value: ws.ID},
		})
		app.Out.Success(fmt.Sprintf("Workspace %q provisioned from module %q", ws.Name, moduleName))
		return nil
	},
}

func init() {
	modShowCmd.Flags().StringVar(&modShowProvider, "provider", "", "module provider (required)")
	modShowCmd.MarkFlagRequired("provider")

	modProvisionCmd.Flags().StringVar(&provisionName, "name", "", "workspace name (auto-generated if omitted)")
	modProvisionCmd.Flags().StringArrayVar(&provisionVars, "var", nil, "variable in key=value format (repeatable)")
	modProvisionCmd.Flags().StringVar(&provisionProvider, "provider", "", "module provider (required)")
	modProvisionCmd.MarkFlagRequired("provider")

	moduleCmd.AddCommand(modListCmd, modShowCmd, modProvisionCmd)
	rootCmd.AddCommand(moduleCmd)
}

func resolveProjectID(ctx context.Context, name string) (string, error) {
	opts := &tfe.ProjectListOptions{
		ListOptions: tfe.ListOptions{PageSize: 100},
		Name:        name,
	}
	for {
		list, err := app.Client.Projects.List(ctx, app.Org, opts)
		if err != nil {
			return "", fmt.Errorf("listing projects: %w", err)
		}
		for _, p := range list.Items {
			if p.Name == name {
				return p.ID, nil
			}
		}
		if list.NextPage == 0 {
			break
		}
		opts.PageNumber = list.NextPage
	}
	return "", fmt.Errorf("project %q not found in org %s", name, app.Org)
}

func randomSuffix() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
