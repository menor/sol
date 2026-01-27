package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestJSONFormatterWrite(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf, false)

	data := map[string]string{"key": "value"}
	if err := f.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify valid JSON
	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Write() produced invalid JSON: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("result[key] = %q, want %q", result["key"], "value")
	}
}

func TestJSONFormatterWriteIndented(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf, true)

	data := map[string]string{"key": "value"}
	if err := f.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "  ") {
		t.Error("Write() with indent=true should produce indented output")
	}
}

func TestJSONFormatterWriteCompact(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf, false)

	data := map[string]string{"key": "value"}
	if err := f.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// Compact JSON should not have leading spaces (indentation)
	if strings.Contains(output, "  \"key\"") {
		t.Error("Write() with indent=false should produce compact output")
	}
}

func TestJSONFormatterWriteError(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(&buf, false)

	testErr := errors.New("something went wrong")
	if err := f.WriteError(testErr); err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}

	// Verify structure
	var result map[string]map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("WriteError() produced invalid JSON: %v", err)
	}

	if result["error"]["message"] != "something went wrong" {
		t.Errorf("error.message = %q, want %q", result["error"]["message"], "something went wrong")
	}
}

func TestTextFormatterWriteError(t *testing.T) {
	var buf bytes.Buffer
	f := NewTextFormatter(&buf)

	testErr := errors.New("something went wrong")
	if err := f.WriteError(testErr); err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}

	output := buf.String()
	if output != "Error: something went wrong\n" {
		t.Errorf("WriteError() = %q, want %q", output, "Error: something went wrong\n")
	}
}

func TestTOONFormatterWriteError(t *testing.T) {
	var buf bytes.Buffer
	f := NewTOONFormatter(&buf)

	testErr := errors.New("something went wrong")
	if err := f.WriteError(testErr); err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}

	output := buf.String()
	if output != "error: something went wrong\n" {
		t.Errorf("WriteError() = %q, want %q", output, "error: something went wrong\n")
	}
}

func TestNewFormatterJSON(t *testing.T) {
	f := New("json")
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("New(\"json\") returned %T, want *JSONFormatter", f)
	}
}

func TestNewFormatterText(t *testing.T) {
	f := New("text")
	if _, ok := f.(*TextFormatter); !ok {
		t.Errorf("New(\"text\") returned %T, want *TextFormatter", f)
	}
}

func TestNewFormatterTOON(t *testing.T) {
	f := New("toon")
	if _, ok := f.(*TOONFormatter); !ok {
		t.Errorf("New(\"toon\") returned %T, want *TOONFormatter", f)
	}
}

func TestNewFormatterDefault(t *testing.T) {
	// Unknown format should default to JSON
	f := New("unknown")
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("New(\"unknown\") returned %T, want *JSONFormatter (default)", f)
	}
}

func TestFormatConstants(t *testing.T) {
	// Verify format constants have expected values
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
	if FormatTOON != "toon" {
		t.Errorf("FormatTOON = %q, want %q", FormatTOON, "toon")
	}
	if FormatText != "text" {
		t.Errorf("FormatText = %q, want %q", FormatText, "text")
	}
}
