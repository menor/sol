package output

import (
	"encoding/json"
	"io"
	"os"
)

// Format represents an output format
type Format string

const (
	FormatJSON Format = "json"
	FormatTOON Format = "toon"
	FormatText Format = "text"
)

// Formatter handles output formatting
type Formatter interface {
	Write(v any) error
	WriteError(err error) error
}

// JSONFormatter outputs JSON
type JSONFormatter struct {
	writer io.Writer
	indent bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(w io.Writer, indent bool) *JSONFormatter {
	return &JSONFormatter{writer: w, indent: indent}
}

func (f *JSONFormatter) Write(v any) error {
	encoder := json.NewEncoder(f.writer)
	if f.indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(v)
}

func (f *JSONFormatter) WriteError(err error) error {
	wrapper := map[string]any{
		"error": map[string]any{
			"message": err.Error(),
		},
	}
	return f.Write(wrapper)
}

// TextFormatter outputs human-readable text
type TextFormatter struct {
	writer io.Writer
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(w io.Writer) *TextFormatter {
	return &TextFormatter{writer: w}
}

func (f *TextFormatter) Write(v any) error {
	// For now, fall back to JSON with indentation
	// TODO: implement proper table formatting
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (f *TextFormatter) WriteError(err error) error {
	_, writeErr := f.writer.Write([]byte("Error: " + err.Error() + "\n"))
	return writeErr
}

// TOONFormatter outputs token-efficient format
type TOONFormatter struct {
	writer io.Writer
}

// NewTOONFormatter creates a new TOON formatter
func NewTOONFormatter(w io.Writer) *TOONFormatter {
	return &TOONFormatter{writer: w}
}

func (f *TOONFormatter) Write(v any) error {
	// TODO: implement TOON encoding
	// For now, fall back to compact JSON
	encoder := json.NewEncoder(f.writer)
	return encoder.Encode(v)
}

func (f *TOONFormatter) WriteError(err error) error {
	_, writeErr := f.writer.Write([]byte("error: " + err.Error() + "\n"))
	return writeErr
}

// New creates a formatter based on format string
func New(format string) Formatter {
	switch Format(format) {
	case FormatTOON:
		return NewTOONFormatter(os.Stdout)
	case FormatText:
		return NewTextFormatter(os.Stdout)
	default:
		return NewJSONFormatter(os.Stdout, true)
	}
}
