package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	toon "github.com/toon-format/toon-go"
)

// Format represents an output format
type Format string

const (
	FormatJSON Format = "json"
	FormatTOON Format = "toon"
	FormatText Format = "text"
)

// Formatter handles output formatting. Errors render through Write as an
// envelope built in cmd.render — there is no separate error method, so the
// envelope shape cannot fork between call sites.
type Formatter interface {
	Write(v any) error
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

// TOONFormatter outputs token-efficient format
type TOONFormatter struct {
	writer io.Writer
}

// NewTOONFormatter creates a new TOON formatter
func NewTOONFormatter(w io.Writer) *TOONFormatter {
	return &TOONFormatter{writer: w}
}

func (f *TOONFormatter) Write(v any) error {
	// toon-go reads `toon` struct tags, which sol's types don't carry.
	// Round-trip through encoding/json so the json tags govern TOON field
	// names and omitempty — one contract, two encodings. This also converts
	// nil pointers (e.g. *time.Time) before toon-go sees them, which
	// otherwise panic inside toon.Marshal.
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var generic any
	if err := json.Unmarshal(jsonBytes, &generic); err != nil {
		return err
	}

	data, err := f.marshalWithRecover(generic)
	if err != nil {
		// Fall back to compact JSON if TOON encoding fails
		encoder := json.NewEncoder(f.writer)
		return encoder.Encode(v)
	}
	_, err = f.writer.Write(data)
	if err != nil {
		return err
	}
	// Add newline if not present
	if len(data) > 0 && data[len(data)-1] != '\n' {
		_, err = f.writer.Write([]byte("\n"))
	}
	return err
}

// marshalWithRecover calls toon.Marshal with panic recovery. The JSON
// round-trip in Write removes the known panic triggers (nil pointer types),
// but toon-go is pre-1.0 so keep the safety net.
func (f *TOONFormatter) marshalWithRecover(v any) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &toonPanicError{recovered: r}
		}
	}()
	return toon.Marshal(v)
}

// toonPanicError wraps a recovered panic as an error.
type toonPanicError struct {
	recovered any
}

func (e *toonPanicError) Error() string {
	return fmt.Sprintf("toon encoding panic: %v", e.recovered)
}

// New creates a formatter based on format string, writing to stdout.
func New(format string) Formatter {
	return NewWithWriter(format, os.Stdout)
}

// NewWithWriter creates a formatter based on format string, writing to w.
// The error render path uses this to inject a writer (stdout in production,
// a buffer in tests).
func NewWithWriter(format string, w io.Writer) Formatter {
	switch Format(format) {
	case FormatTOON:
		return NewTOONFormatter(w)
	case FormatText:
		return NewTextFormatter(w)
	default:
		return NewJSONFormatter(w, true)
	}
}
