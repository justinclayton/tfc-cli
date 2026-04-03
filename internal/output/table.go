package output

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type tableFormatter struct {
	w io.Writer
}

func (f *tableFormatter) Table(headers []string, rows [][]string) {
	headerFmt := color.New(color.Bold, color.FgCyan).SprintfFunc()

	ifaces := make([]any, len(headers))
	for i, h := range headers {
		ifaces[i] = h
	}

	tbl := table.New(ifaces...)
	tbl.WithWriter(f.w)
	tbl.WithHeaderFormatter(headerFmt)
	tbl.WithPadding(2)

	for _, row := range rows {
		rowIfaces := make([]any, len(row))
		for i, v := range row {
			rowIfaces[i] = v
		}
		tbl.AddRow(rowIfaces...)
	}
	tbl.Print()
}

func (f *tableFormatter) Detail(fields []Field) {
	bold := color.New(color.Bold)
	maxLen := 0
	for _, field := range fields {
		if len(field.Label) > maxLen {
			maxLen = len(field.Label)
		}
	}
	for _, field := range fields {
		bold.Fprintf(f.w, "%-*s  ", maxLen, field.Label+":")
		fmt.Fprintln(f.w, field.Value)
	}
}

func (f *tableFormatter) Success(msg string) {
	color.New(color.FgGreen).Fprintln(f.w, msg)
}

func (f *tableFormatter) Error(msg string) {
	color.New(color.FgRed).Fprintln(f.w, "Error: "+msg)
}

func (f *tableFormatter) JSON(v any) {
	jf := &jsonFormatter{w: f.w}
	jf.JSON(v)
}

func (f *tableFormatter) Writer() io.Writer {
	return f.w
}
