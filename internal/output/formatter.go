package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"gopkg.in/yaml.v3"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	Format(data interface{}) error
	FormatList(items interface{}, columns []Column) error
	FormatError(statusCode int, body []byte) error
}

// Column defines a table column.
type Column struct {
	Header string
	Field  string
	Width  int
}

// DetectFormat returns the output format based on flag and TTY detection.
func DetectFormat(outputFlag string) string {
	if outputFlag != "" {
		return outputFlag
	}
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return "table"
	}
	return "json"
}

// NewFormatter creates a Formatter for the given format.
func NewFormatter(format string, w io.Writer) Formatter {
	switch format {
	case "json":
		return &jsonFormatter{w: w}
	case "ndjson":
		return &ndjsonFormatter{w: w}
	case "yaml":
		return &yamlFormatter{w: w}
	default:
		return &tableFormatter{w: w}
	}
}

// --- JSON Formatter ---

type jsonFormatter struct{ w io.Writer }

func (f *jsonFormatter) Format(data interface{}) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *jsonFormatter) FormatList(items interface{}, _ []Column) error {
	return f.Format(items)
}

func (f *jsonFormatter) FormatError(statusCode int, body []byte) error {
	// Pass through raw JSON error
	_, err := f.w.Write(body)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f.w)
	return err
}

// --- NDJSON Formatter ---

type ndjsonFormatter struct{ w io.Writer }

func (f *ndjsonFormatter) Format(data interface{}) error {
	return json.NewEncoder(f.w).Encode(data)
}

func (f *ndjsonFormatter) FormatList(items interface{}, _ []Column) error {
	// items should be a slice; encode each item as one line
	return json.NewEncoder(f.w).Encode(items)
}

func (f *ndjsonFormatter) FormatError(statusCode int, body []byte) error {
	_, err := f.w.Write(body)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f.w)
	return err
}

// WriteNDJSONItem writes a single item as one NDJSON line. Useful for streaming.
func WriteNDJSONItem(w io.Writer, item interface{}) error {
	return json.NewEncoder(w).Encode(item)
}

// --- YAML Formatter ---

type yamlFormatter struct{ w io.Writer }

func (f *yamlFormatter) Format(data interface{}) error {
	return yaml.NewEncoder(f.w).Encode(data)
}

func (f *yamlFormatter) FormatList(items interface{}, _ []Column) error {
	return f.Format(items)
}

func (f *yamlFormatter) FormatError(statusCode int, body []byte) error {
	_, err := fmt.Fprintf(f.w, "error: %s\n", string(body))
	return err
}
