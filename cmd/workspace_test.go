package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatRunLog_JSONDiagnostics(t *testing.T) {
	// Simulate Terraform JSON log output with a diagnostic entry.
	diag := tfLogEntry{
		Level:   "error",
		Message: "something broke",
		Type:    "diagnostic",
		Diagnostic: &tfDiagnostic{
			Severity: "error",
			Summary:  "Invalid reference",
			Detail:   "The reference is not valid.",
			Range: &tfRange{
				Filename: "main.tf",
				Start:    tfPos{Line: 10, Column: 5},
				End:      tfPos{Line: 10, Column: 20},
			},
			Snippet: &tfSnippet{
				Context:           "resource \"aws_instance\" \"web\"",
				Code:              "  name = var.missing",
				StartLine:         10,
				HighlightStartOff: 9,
				HighlightEndOff:   20,
			},
		},
	}

	line, err := json.Marshal(diag)
	if err != nil {
		t.Fatal(err)
	}

	r := strings.NewReader(string(line) + "\n")
	out := formatRunLog(r)

	if !strings.Contains(out, "Invalid reference") {
		t.Errorf("expected diagnostic summary in output, got:\n%s", out)
	}
	if !strings.Contains(out, "main.tf") {
		t.Errorf("expected filename in output, got:\n%s", out)
	}
	if !strings.Contains(out, "The reference is not valid.") {
		t.Errorf("expected diagnostic detail in output, got:\n%s", out)
	}
}

func TestFormatRunLog_FallbackRawTail(t *testing.T) {
	// When there are no JSON diagnostic lines, formatRunLog should return
	// raw lines as a fallback.
	rawLog := "line 1\nline 2\nline 3\nOperation failed\n"
	r := strings.NewReader(rawLog)
	out := formatRunLog(r)

	// The trailer "Operation failed" should be captured but the raw fallback
	// path returns all lines since there are fewer than 20.
	if !strings.Contains(out, "line 1") {
		t.Errorf("expected raw fallback to contain 'line 1', got:\n%s", out)
	}
	if !strings.Contains(out, "line 3") {
		t.Errorf("expected raw fallback to contain 'line 3', got:\n%s", out)
	}
}

func TestFormatRunLog_FallbackTailsTwentyLines(t *testing.T) {
	// Build more than 20 non-JSON lines to trigger the tail-20 behavior.
	var lines []string
	for i := 0; i < 30; i++ {
		lines = append(lines, "plain log line")
	}
	r := strings.NewReader(strings.Join(lines, "\n"))
	out := formatRunLog(r)

	outLines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(outLines) != 20 {
		t.Errorf("expected 20 tail lines, got %d", len(outLines))
	}
}

func TestFormatRunLog_MixedJSONAndPlain(t *testing.T) {
	// A diagnostic line followed by a plain trailer.
	diag := tfLogEntry{
		Level:   "error",
		Message: "err",
		Type:    "diagnostic",
		Diagnostic: &tfDiagnostic{
			Severity: "error",
			Summary:  "Cycle detected",
		},
	}
	line, _ := json.Marshal(diag)
	input := string(line) + "\nOperation failed: see above\n"

	r := strings.NewReader(input)
	out := formatRunLog(r)

	if !strings.Contains(out, "Cycle detected") {
		t.Errorf("expected diagnostic summary, got:\n%s", out)
	}
	if !strings.Contains(out, "Operation failed") {
		t.Errorf("expected trailer in output, got:\n%s", out)
	}
}

func TestFormatRunLog_WarningDiagnostic(t *testing.T) {
	diag := tfLogEntry{
		Level:   "warn",
		Message: "deprecated",
		Type:    "diagnostic",
		Diagnostic: &tfDiagnostic{
			Severity: "warning",
			Summary:  "Deprecated attribute",
			Detail:   "Use the new attribute instead.",
		},
	}
	line, _ := json.Marshal(diag)

	r := strings.NewReader(string(line) + "\n")
	out := formatRunLog(r)

	if !strings.Contains(out, "Deprecated attribute") {
		t.Errorf("expected warning summary, got:\n%s", out)
	}
}

func TestFormatRunLog_EmptyInput(t *testing.T) {
	r := strings.NewReader("")
	out := formatRunLog(r)
	// Empty input should not panic and should return something (empty fallback).
	if out == "" {
		// An empty string is acceptable for empty input.
		return
	}
}

func TestIsInteractive_DoesNotPanic(t *testing.T) {
	// isInteractive() checks os.Stdin.Fd() via isatty. In test environments
	// it will return false (not a terminal), but it must not panic.
	_ = isInteractive()
}

func TestDefaultVal(t *testing.T) {
	if got := defaultVal("present", "fallback"); got != "present" {
		t.Errorf("expected 'present', got %q", got)
	}
	if got := defaultVal("", "fallback"); got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}
