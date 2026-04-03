package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/justinclayton/tfc-cli/internal/output"
)

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"ws"},
	Short:   "Manage workspaces",
}

// ── list ─────────────────────────────────────────────────────────────────

var (
	wsListSearch string
	wsListTags   string
	wsListStatus string
	wsListSort   string
)

var wsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		opts := &tfe.WorkspaceListOptions{
			ListOptions: tfe.ListOptions{PageSize: 100},
			Include:     []tfe.WSIncludeOpt{tfe.WSCurrentRun, tfe.WSProject},
		}

		if wsListSearch != "" {
			opts.Search = wsListSearch
		}
		if wsListTags != "" {
			opts.Tags = wsListTags
		}
		if wsListStatus != "" {
			opts.CurrentRunStatus = wsListStatus
		}
		if wsListSort != "" {
			opts.Sort = wsListSort
		}

		// --project flag from global scope filters by project
		if app.Project != "" {
			projID, err := resolveProjectID(ctx, app.Project)
			if err != nil {
				return err
			}
			opts.ProjectID = projID
		}

		var headers = []string{"NAME", "ID", "PROJECT", "STATUS", "TERRAFORM VERSION", "UPDATED"}
		var rows [][]string

		for {
			list, err := app.Client.Workspaces.List(ctx, app.Org, opts)
			if err != nil {
				return fmt.Errorf("listing workspaces: %w", err)
			}

			for _, ws := range list.Items {
				status := ""
				if ws.CurrentRun != nil {
					status = string(ws.CurrentRun.Status)
				}
				project := ""
				if ws.Project != nil {
					project = ws.Project.Name
				}
				updated := ws.UpdatedAt.Format("2006-01-02 15:04")
				rows = append(rows, []string{
					ws.Name,
					ws.ID,
					project,
					status,
					ws.TerraformVersion,
					updated,
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

var wsShowCmd = &cobra.Command{
	Use:   "show <workspace>",
	Short: "Show workspace details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ws, err := app.Client.Workspaces.ReadWithOptions(ctx, app.Org, args[0], &tfe.WorkspaceReadOptions{
			Include: []tfe.WSIncludeOpt{tfe.WSCurrentRun},
		})
		if err != nil {
			return fmt.Errorf("reading workspace: %w", err)
		}

		currentRun := "(none)"
		if ws.CurrentRun != nil {
			currentRun = fmt.Sprintf("%s (%s)", ws.CurrentRun.ID, ws.CurrentRun.Status)
		}

		updated := ws.UpdatedAt.Format("2006-01-02 15:04:05")

		vcsRepo := "(none)"
		if ws.VCSRepo != nil {
			vcsRepo = ws.VCSRepo.Identifier
		}

		app.Out.Detail([]output.Field{
			{Label: "Name", Value: ws.Name},
			{Label: "ID", Value: ws.ID},
			{Label: "Terraform Version", Value: ws.TerraformVersion},
			{Label: "Execution Mode", Value: string(ws.ExecutionMode)},
			{Label: "Auto Apply", Value: fmt.Sprintf("%t", ws.AutoApply)},
			{Label: "Working Directory", Value: ws.WorkingDirectory},
			{Label: "VCS Repo", Value: vcsRepo},
			{Label: "Current Run", Value: currentRun},
			{Label: "Resource Count", Value: fmt.Sprintf("%d", ws.ResourceCount)},
			{Label: "Updated", Value: updated},
		})
		return nil
	},
}

// ── run ──────────────────────────────────────────────────────────────────

var wsRunMsg string

var wsRunCmd = &cobra.Command{
	Use:   "run <workspace>",
	Short: "Trigger a run on a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ws, err := app.Client.Workspaces.Read(ctx, app.Org, args[0])
		if err != nil {
			return fmt.Errorf("reading workspace: %w", err)
		}

		opts := tfe.RunCreateOptions{
			Workspace: ws,
			Message:   tfe.String(wsRunMsg),
		}

		run, err := app.Client.Runs.Create(ctx, opts)
		if err != nil {
			return fmt.Errorf("creating run: %w", err)
		}

		app.Out.Success(fmt.Sprintf("Run %s created on workspace %s (status: %s)", run.ID, args[0], run.Status))
		return nil
	},
}

// ── runs ─────────────────────────────────────────────────────────────────

var wsRunsLimit int

var wsRunsCmd = &cobra.Command{
	Use:   "runs <workspace>",
	Short: "List recent runs for a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ws, err := app.Client.Workspaces.Read(ctx, app.Org, args[0])
		if err != nil {
			return fmt.Errorf("reading workspace: %w", err)
		}

		pageSize := wsRunsLimit
		if pageSize > 100 {
			pageSize = 100
		}
		opts := &tfe.RunListOptions{
			ListOptions: tfe.ListOptions{PageSize: pageSize},
		}

		list, err := app.Client.Runs.List(ctx, ws.ID, opts)
		if err != nil {
			return fmt.Errorf("listing runs: %w", err)
		}

		headers := []string{"ID", "STATUS", "MESSAGE", "CREATED"}
		var rows [][]string
		for _, r := range list.Items {
			msg := r.Message
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			rows = append(rows, []string{
				r.ID,
				string(r.Status),
				msg,
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		app.Out.Table(headers, rows)
		return nil
	},
}

// ── show-run ─────────────────────────────────────────────────────────────

var wsShowRunCmd = &cobra.Command{
	Use:   "show-run <workspace> [run-id]",
	Short: "Show details of a run (defaults to latest)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		var runID string
		if len(args) >= 2 {
			runID = args[1]
		} else {
			// Fetch latest run for the workspace
			ws, err := app.Client.Workspaces.Read(ctx, app.Org, args[0])
			if err != nil {
				return fmt.Errorf("reading workspace: %w", err)
			}
			list, err := app.Client.Runs.List(ctx, ws.ID, &tfe.RunListOptions{
				ListOptions: tfe.ListOptions{PageSize: 1},
			})
			if err != nil {
				return fmt.Errorf("listing runs: %w", err)
			}
			if len(list.Items) == 0 {
				return fmt.Errorf("no runs found for workspace %s", args[0])
			}
			runID = list.Items[0].ID
		}

		run, err := app.Client.Runs.ReadWithOptions(ctx, runID, &tfe.RunReadOptions{
			Include: []tfe.RunIncludeOpt{tfe.RunPlan, tfe.RunApply},
		})
		if err != nil {
			return fmt.Errorf("reading run: %w", err)
		}

		fields := []output.Field{
			{Label: "ID", Value: run.ID},
			{Label: "Status", Value: string(run.Status)},
			{Label: "Message", Value: run.Message},
			{Label: "Is Destroy", Value: fmt.Sprintf("%t", run.IsDestroy)},
			{Label: "Source", Value: string(run.Source)},
			{Label: "Created", Value: run.CreatedAt.Format("2006-01-02 15:04:05")},
			{Label: "Plan Only", Value: fmt.Sprintf("%t", run.PlanOnly)},
		}

		if run.Plan != nil {
			fields = append(fields, output.Field{Label: "Plan Status", Value: string(run.Plan.Status)})
		}
		if run.Apply != nil {
			fields = append(fields, output.Field{Label: "Apply Status", Value: string(run.Apply.Status)})
		}

		app.Out.Detail(fields)

		// If the run errored, fetch and display the error output
		if run.Status == tfe.RunErrored {
			errLog, phase := fetchErrorLog(ctx, run)
			if errLog != "" {
				fmt.Fprintf(app.Out.Writer(), "\n── %s error output ──\n\n%s\n", phase, errLog)
			}
		}

		return nil
	},
}

// ── destroy ──────────────────────────────────────────────────────────────

var wsDestroyConfirm bool

var wsDestroyCmd = &cobra.Command{
	Use:   "destroy <workspace>",
	Short: "Queue a destroy run for a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wsName := args[0]
		ctx := context.Background()

		if !wsDestroyConfirm {
			if !isInteractive() {
				return fmt.Errorf("refusing to destroy in non-interactive mode without --confirm")
			}

			fmt.Fprintf(os.Stderr, "\n  WARNING: This will queue a DESTROY run for workspace '%s'.\n", wsName)
			fmt.Fprintf(os.Stderr, "  All managed resources will be destroyed.\n\n")
			fmt.Fprintf(os.Stderr, "  Type the workspace name to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			input := readLine(reader)
			if input != wsName {
				return fmt.Errorf("confirmation failed — input did not match workspace name")
			}
		}

		ws, err := app.Client.Workspaces.Read(ctx, app.Org, wsName)
		if err != nil {
			return fmt.Errorf("reading workspace: %w", err)
		}

		run, err := app.Client.Runs.Create(ctx, tfe.RunCreateOptions{
			Workspace: ws,
			IsDestroy: tfe.Bool(true),
			Message:   tfe.String("Destroy initiated via tfc CLI"),
		})
		if err != nil {
			return fmt.Errorf("creating destroy run: %w", err)
		}

		app.Out.Success(fmt.Sprintf("Destroy run %s queued for workspace %s", run.ID, wsName))
		return nil
	},
}

// ── delete ───────────────────────────────────────────────────────────────

var wsDeleteConfirm bool

var wsDeleteCmd = &cobra.Command{
	Use:   "delete <workspace>",
	Short: "Permanently delete a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wsName := args[0]
		ctx := context.Background()

		if !wsDeleteConfirm {
			if !isInteractive() {
				return fmt.Errorf("refusing to delete in non-interactive mode without --confirm")
			}

			fmt.Fprintf(os.Stderr, "\n  DANGER: This will PERMANENTLY DELETE workspace '%s'.\n", wsName)
			fmt.Fprintf(os.Stderr, "  This action CANNOT be undone. State, runs, and all history will be lost.\n\n")
			fmt.Fprintf(os.Stderr, "  Type 'DELETE %s' to confirm: ", wsName)

			reader := bufio.NewReader(os.Stdin)
			input := readLine(reader)
			expected := fmt.Sprintf("DELETE %s", wsName)
			if input != expected {
				return fmt.Errorf("confirmation failed — expected '%s'", expected)
			}
		}

		if err := app.Client.Workspaces.Delete(ctx, app.Org, wsName); err != nil {
			return fmt.Errorf("deleting workspace: %w", err)
		}

		app.Out.Success(fmt.Sprintf("Workspace %s deleted", wsName))
		return nil
	},
}

func init() {
	wsListCmd.Flags().StringVarP(&wsListSearch, "search", "s", "", "filter by workspace name (substring match)")
	wsListCmd.Flags().StringVarP(&wsListTags, "tags", "t", "", "filter by tags (comma-separated)")
	wsListCmd.Flags().StringVar(&wsListStatus, "status", "", "filter by current run status (e.g. applied, errored, planning)")
	wsListCmd.Flags().StringVar(&wsListSort, "sort", "", "sort field: name, -name, current-run.created-at, -current-run.created-at")

	wsRunsCmd.Flags().IntVarP(&wsRunsLimit, "limit", "n", 20, "maximum number of runs to show")

	wsRunCmd.Flags().StringVarP(&wsRunMsg, "message", "m", "Triggered via tfc CLI", "run message")
	wsDestroyCmd.Flags().BoolVar(&wsDestroyConfirm, "confirm", false, "skip confirmation prompt")
	wsDeleteCmd.Flags().BoolVar(&wsDeleteConfirm, "confirm", false, "skip confirmation prompt")

	workspaceCmd.AddCommand(wsListCmd, wsShowCmd, wsRunCmd, wsRunsCmd, wsShowRunCmd, wsDestroyCmd, wsDeleteCmd)
	rootCmd.AddCommand(workspaceCmd)
}

func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

// tfLogEntry represents a single JSON log line from Terraform's machine-readable output.
type tfLogEntry struct {
	Level      string        `json:"@level"`
	Message    string        `json:"@message"`
	Module     string        `json:"@module"`
	Type       string        `json:"type"`
	Diagnostic *tfDiagnostic `json:"diagnostic,omitempty"`
}

type tfDiagnostic struct {
	Severity string     `json:"severity"`
	Summary  string     `json:"summary"`
	Detail   string     `json:"detail"`
	Address  string     `json:"address,omitempty"`
	Range    *tfRange   `json:"range,omitempty"`
	Snippet  *tfSnippet `json:"snippet,omitempty"`
}

type tfRange struct {
	Filename string    `json:"filename"`
	Start    tfPos     `json:"start"`
	End      tfPos     `json:"end"`
}

type tfPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type tfSnippet struct {
	Context            string `json:"context"`
	Code               string `json:"code"`
	StartLine          int    `json:"start_line"`
	HighlightStartOff  int    `json:"highlight_start_offset"`
	HighlightEndOff    int    `json:"highlight_end_offset"`
}

// fetchErrorLog reads the log from the errored phase and returns
// pretty-formatted diagnostics.
func fetchErrorLog(ctx context.Context, run *tfe.Run) (string, string) {
	if run.Plan != nil && run.Plan.Status == tfe.PlanErrored {
		logs, err := app.Client.Plans.Logs(ctx, run.Plan.ID)
		if err != nil {
			return fmt.Sprintf("(failed to fetch plan logs: %s)", err), "plan"
		}
		return formatRunLog(logs), "plan"
	}

	if run.Apply != nil && run.Apply.Status == tfe.ApplyErrored {
		logs, err := app.Client.Applies.Logs(ctx, run.Apply.ID)
		if err != nil {
			return fmt.Sprintf("(failed to fetch apply logs: %s)", err), "apply"
		}
		return formatRunLog(logs), "apply"
	}

	return "", ""
}

// formatRunLog parses Terraform JSON log output and pretty-prints diagnostics.
// Non-JSON lines (like the "Operation failed" trailer) are included as-is.
func formatRunLog(r io.Reader) string {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Sprintf("(failed to read logs: %s)", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")

	var diags []tfDiagnostic
	var trailer string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry tfLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Non-JSON line (e.g. "Operation failed: ...")
			if strings.HasPrefix(line, "Operation") || strings.HasPrefix(line, "Error") {
				trailer = line
			}
			continue
		}

		if entry.Type == "diagnostic" && entry.Diagnostic != nil {
			diags = append(diags, *entry.Diagnostic)
		}
	}

	if len(diags) == 0 {
		// Fallback: no structured diagnostics found, show last 20 raw lines
		if len(lines) > 20 {
			lines = lines[len(lines)-20:]
		}
		return strings.Join(lines, "\n")
	}

	var buf strings.Builder
	bold := color.New(color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	dim := color.New(color.Faint)

	// Soft red — 24-bit truecolor: rgb(196, 85, 75), a warm terracotta
	softRed := func(s string) string {
		if color.NoColor {
			return s
		}
		return fmt.Sprintf("\033[1;38;2;196;85;75m%s\033[0m", s)
	}

	for i, d := range diags {
		if i > 0 {
			buf.WriteString("\n")
		}

		// Header line: "Error: summary" or "Warning: summary"
		switch d.Severity {
		case "error":
			buf.WriteString(softRed("Error: "))
		case "warning":
			yellow.Fprintf(&buf, "Warning: ")
		}
		bold.Fprintf(&buf, "%s\n", d.Summary)

		// Source location
		if d.Range != nil {
			dim.Fprintf(&buf, "  on %s line %d", d.Range.Filename, d.Range.Start.Line)
			if d.Address != "" {
				dim.Fprintf(&buf, ", in %s", d.Address)
			}
			buf.WriteString("\n")
		} else if d.Address != "" {
			dim.Fprintf(&buf, "  in %s\n", d.Address)
		}

		// Code snippet with highlight
		if d.Snippet != nil && d.Snippet.Code != "" {
			dim.Fprintf(&buf, "  %d: ", d.Snippet.StartLine)
			code := d.Snippet.Code
			start := d.Snippet.HighlightStartOff
			end := d.Snippet.HighlightEndOff
			if start >= 0 && end > start && end <= len(code) {
				buf.WriteString(code[:start])
				bold.Fprint(&buf, code[start:end])
				buf.WriteString(code[end:])
			} else {
				buf.WriteString(code)
			}
			buf.WriteString("\n")
		}

		// Detail text in the same soft red
		if d.Detail != "" {
			buf.WriteString("\n")
			for _, dline := range strings.Split(d.Detail, "\n") {
				fmt.Fprintf(&buf, "  %s\n", softRed(dline))
			}
		}
	}

	if trailer != "" {
		fmt.Fprintf(&buf, "\n%s\n", softRed(trailer))
	}

	return buf.String()
}
