package output

import (
	"fmt"
	"io"
	"strings"
)

type plainFormatter struct {
	w io.Writer
}

func (f *plainFormatter) Table(headers []string, rows [][]string) {
	// No headers in plain mode — designed for cut/awk/grep
	for _, row := range rows {
		fmt.Fprintln(f.w, strings.Join(row, "\t"))
	}
}

func (f *plainFormatter) Detail(fields []Field) {
	for _, field := range fields {
		fmt.Fprintf(f.w, "%s\t%s\n", field.Label, field.Value)
	}
}

func (f *plainFormatter) Success(msg string) {
	fmt.Fprintln(f.w, msg)
}

func (f *plainFormatter) Error(msg string) {
	fmt.Fprintln(f.w, "Error: "+msg)
}

func (f *plainFormatter) JSON(v any) {
	jf := &jsonFormatter{w: f.w}
	jf.JSON(v)
}

func (f *plainFormatter) Writer() io.Writer {
	return f.w
}
