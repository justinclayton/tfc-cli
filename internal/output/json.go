package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonFormatter struct {
	w io.Writer
}

func (f *jsonFormatter) Table(headers []string, rows [][]string) {
	// Convert to array of objects keyed by header
	var result []map[string]string
	for _, row := range rows {
		obj := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				obj[h] = row[i]
			}
		}
		result = append(result, obj)
	}
	f.JSON(result)
}

func (f *jsonFormatter) Detail(fields []Field) {
	obj := make(map[string]string, len(fields))
	for _, field := range fields {
		obj[field.Label] = field.Value
	}
	f.JSON(obj)
}

func (f *jsonFormatter) Success(msg string) {
	f.JSON(map[string]string{"status": "success", "message": msg})
}

func (f *jsonFormatter) Error(msg string) {
	f.JSON(map[string]string{"status": "error", "message": msg})
}

func (f *jsonFormatter) JSON(v any) {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(f.w, "{\"error\": \"failed to encode JSON: %s\"}\n", err)
	}
}

func (f *jsonFormatter) Writer() io.Writer {
	return f.w
}
