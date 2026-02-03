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

func TestTOONFormatterWrite(t *testing.T) {
	var buf bytes.Buffer
	f := NewTOONFormatter(&buf)

	// Test with a simple struct
	type Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	data := Project{ID: "abc123", Name: "my-app"}
	if err := f.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// TOON output should be more compact than JSON
	// and should not have JSON braces
	if strings.HasPrefix(output, "{") {
		t.Errorf("TOON output should not start with JSON brace, got: %s", output)
	}
	// Should contain the field values
	if !strings.Contains(output, "abc123") || !strings.Contains(output, "my-app") {
		t.Errorf("TOON output should contain field values, got: %s", output)
	}
}

func TestTOONFormatterWriteArray(t *testing.T) {
	var buf bytes.Buffer
	f := NewTOONFormatter(&buf)

	// Test with array of structs - the key TOON use case
	type Project struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Region string `json:"region"`
	}
	data := []Project{
		{ID: "abc123", Name: "my-app", Region: "us-east"},
		{ID: "def456", Name: "api-service", Region: "eu-west"},
	}
	if err := f.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// TOON arrays use format: [count]{fields}:
	// Example: [2]{ID,Name,Region}:
	if !strings.Contains(output, "[2]") {
		t.Errorf("TOON output should contain array count [2], got: %s", output)
	}
	// Should contain field schema
	if !strings.Contains(output, "ID") || !strings.Contains(output, "Name") || !strings.Contains(output, "Region") {
		t.Errorf("TOON output should contain field names, got: %s", output)
	}
	// Should contain all values
	if !strings.Contains(output, "abc123") || !strings.Contains(output, "def456") {
		t.Errorf("TOON output should contain all IDs, got: %s", output)
	}
	// Should NOT contain JSON syntax
	if strings.Contains(output, `"id"`) || strings.Contains(output, `"name"`) {
		t.Errorf("TOON output should not contain JSON quoted keys, got: %s", output)
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
