package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNew_ForceJSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, true, false)
	if _, ok := f.(*jsonFormatter); !ok {
		t.Errorf("expected jsonFormatter when forceJSON=true, got %T", f)
	}
}

func TestNew_NonTTYWriter(t *testing.T) {
	// A bytes.Buffer is not an *os.File, so it should produce plainFormatter.
	var buf bytes.Buffer
	f := New(&buf, false, false)
	if _, ok := f.(*plainFormatter); !ok {
		t.Errorf("expected plainFormatter for non-File writer, got %T", f)
	}
}

func TestNew_NoColor(t *testing.T) {
	// Even with a real file, noColor should produce plainFormatter (non-TTY pipe).
	// We use /dev/null which is a file but not a terminal.
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Skip("cannot open /dev/null")
	}
	defer devnull.Close()
	f := New(devnull, false, false)
	// /dev/null is not a TTY so we get plain
	if _, ok := f.(*plainFormatter); !ok {
		t.Errorf("expected plainFormatter for non-TTY file, got %T", f)
	}
}

// --- Plain formatter ---

func TestPlain_Table_NoHeaders(t *testing.T) {
	var buf bytes.Buffer
	f := &plainFormatter{w: &buf}
	f.Table([]string{"NAME", "ID"}, [][]string{
		{"alpha", "ws-123"},
		{"beta", "ws-456"},
	})
	out := buf.String()
	// Plain mode should not include headers
	if strings.Contains(out, "NAME") {
		t.Error("plain Table() should not output headers")
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if lines[0] != "alpha\tws-123" {
		t.Errorf("unexpected first line: %q", lines[0])
	}
}

func TestPlain_Detail(t *testing.T) {
	var buf bytes.Buffer
	f := &plainFormatter{w: &buf}
	f.Detail([]Field{
		{Label: "Name", Value: "my-ws"},
		{Label: "ID", Value: "ws-abc"},
	})
	out := buf.String()
	if !strings.Contains(out, "Name\tmy-ws") {
		t.Errorf("expected tab-separated detail, got %q", out)
	}
}

func TestPlain_Success(t *testing.T) {
	var buf bytes.Buffer
	f := &plainFormatter{w: &buf}
	f.Success("done")
	if got := strings.TrimSpace(buf.String()); got != "done" {
		t.Errorf("expected 'done', got %q", got)
	}
}

func TestPlain_Error(t *testing.T) {
	var buf bytes.Buffer
	f := &plainFormatter{w: &buf}
	f.Error("oops")
	if got := strings.TrimSpace(buf.String()); got != "Error: oops" {
		t.Errorf("expected 'Error: oops', got %q", got)
	}
}

func TestPlain_Writer(t *testing.T) {
	var buf bytes.Buffer
	f := &plainFormatter{w: &buf}
	if f.Writer() != &buf {
		t.Error("Writer() should return the underlying writer")
	}
}

// --- JSON formatter ---

func TestJSON_Table_ArrayOfObjects(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	f.Table([]string{"NAME", "ID"}, [][]string{
		{"alpha", "ws-123"},
		{"beta", "ws-456"},
	})

	var result []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(result))
	}
	if result[0]["NAME"] != "alpha" || result[0]["ID"] != "ws-123" {
		t.Errorf("unexpected first object: %v", result[0])
	}
	if result[1]["NAME"] != "beta" {
		t.Errorf("unexpected second object: %v", result[1])
	}
}

func TestJSON_Table_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	f.Table([]string{"A"}, nil)

	// null is valid JSON for nil slice
	out := strings.TrimSpace(buf.String())
	if out != "null" {
		t.Errorf("expected null for nil rows, got %q", out)
	}
}

func TestJSON_Detail(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	f.Detail([]Field{
		{Label: "Name", Value: "ws-1"},
		{Label: "ID", Value: "ws-abc"},
	})

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["Name"] != "ws-1" || result["ID"] != "ws-abc" {
		t.Errorf("unexpected detail: %v", result)
	}
}

func TestJSON_Success(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	f.Success("all good")

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "success" || result["message"] != "all good" {
		t.Errorf("unexpected success output: %v", result)
	}
}

func TestJSON_Error(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	f.Error("bad thing")

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "error" || result["message"] != "bad thing" {
		t.Errorf("unexpected error output: %v", result)
	}
}

func TestJSON_RawJSON(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	f.JSON(map[string]int{"count": 42})

	var result map[string]int
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["count"] != 42 {
		t.Errorf("expected count=42, got %d", result["count"])
	}
}

func TestJSON_Writer(t *testing.T) {
	var buf bytes.Buffer
	f := &jsonFormatter{w: &buf}
	if f.Writer() != &buf {
		t.Error("Writer() should return the underlying writer")
	}
}

// --- Table formatter ---

func TestTableFormatter_Table(t *testing.T) {
	var buf bytes.Buffer
	f := &tableFormatter{w: &buf}
	f.Table([]string{"NAME", "ID"}, [][]string{
		{"alpha", "ws-123"},
	})
	out := buf.String()
	// The table formatter should include the header text and the row data
	if !strings.Contains(out, "NAME") {
		t.Error("table output should contain header NAME")
	}
	if !strings.Contains(out, "alpha") {
		t.Error("table output should contain row value 'alpha'")
	}
}

func TestTableFormatter_Detail(t *testing.T) {
	var buf bytes.Buffer
	f := &tableFormatter{w: &buf}
	f.Detail([]Field{
		{Label: "Name", Value: "test"},
		{Label: "ID", Value: "ws-xyz"},
	})
	out := buf.String()
	if !strings.Contains(out, "Name:") || !strings.Contains(out, "test") {
		t.Errorf("detail output missing expected content: %q", out)
	}
	if !strings.Contains(out, "ID:") || !strings.Contains(out, "ws-xyz") {
		t.Errorf("detail output missing ID field: %q", out)
	}
}

func TestTableFormatter_Success(t *testing.T) {
	var buf bytes.Buffer
	f := &tableFormatter{w: &buf}
	f.Success("hooray")
	if !strings.Contains(buf.String(), "hooray") {
		t.Errorf("success output missing message: %q", buf.String())
	}
}

func TestTableFormatter_Error(t *testing.T) {
	var buf bytes.Buffer
	f := &tableFormatter{w: &buf}
	f.Error("nope")
	out := buf.String()
	if !strings.Contains(out, "Error:") || !strings.Contains(out, "nope") {
		t.Errorf("error output missing expected content: %q", out)
	}
}

func TestTableFormatter_JSON_Delegates(t *testing.T) {
	// Table formatter's JSON method should produce valid JSON
	var buf bytes.Buffer
	f := &tableFormatter{w: &buf}
	f.JSON(map[string]string{"key": "val"})

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("table JSON() did not produce valid JSON: %v", err)
	}
	if result["key"] != "val" {
		t.Errorf("unexpected JSON output: %v", result)
	}
}

func TestTableFormatter_Writer(t *testing.T) {
	var buf bytes.Buffer
	f := &tableFormatter{w: &buf}
	if f.Writer() != &buf {
		t.Error("Writer() should return the underlying writer")
	}
}
