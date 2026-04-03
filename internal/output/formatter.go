package output

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// Field represents a key-value pair for detail views.
type Field struct {
	Label string
	Value string
}

// Formatter handles all output rendering. Commands call these methods
// without knowing whether output is a pretty table, TSV, or JSON.
type Formatter interface {
	// Table renders tabular data with headers.
	Table(headers []string, rows [][]string)
	// Detail renders key-value pairs for "show" commands.
	Detail(fields []Field)
	// Success prints a success message.
	Success(msg string)
	// Error prints an error message.
	Error(msg string)
	// JSON renders arbitrary data as JSON.
	JSON(v any)
	// Writer returns the underlying writer.
	Writer() io.Writer
}

// New creates a Formatter based on the environment. When --json is set,
// JSON output is used. When stdout is a TTY, pretty tables with color.
// Otherwise, tab-separated plain text for piping.
func New(w io.Writer, forceJSON bool, noColor bool) Formatter {
	if forceJSON {
		return &jsonFormatter{w: w}
	}

	if f, ok := w.(*os.File); ok && isatty.IsTerminal(f.Fd()) && !noColor {
		return &tableFormatter{w: w}
	}

	return &plainFormatter{w: w}
}
