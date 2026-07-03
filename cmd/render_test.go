package cmd

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/menor/sol/internal/errors"
)

// decodeEnvelope parses {"error": {...}} JSON output.
func decodeEnvelope(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, data)
	}
	envelope, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatalf("output missing 'error' wrapper: %s", data)
	}
	return envelope
}

func TestRenderToNilError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := renderTo("json", nil, &stdout, &stderr)

	if code != errors.ExitSuccess {
		t.Errorf("exit = %d, want %d", code, errors.ExitSuccess)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Errorf("nil error should write nothing, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestRenderToJSONEnvelope(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := errors.NewNoProjectError()

	code := renderTo("json", err, &stdout, &stderr)

	if code != errors.ExitUserError {
		t.Errorf("exit = %d, want %d", code, errors.ExitUserError)
	}
	// Structured mode: envelope owns stdout, nothing on stderr (ADR 0002).
	if stderr.Len() != 0 {
		t.Errorf("structured mode must not write to stderr, got %q", stderr.String())
	}

	envelope := decodeEnvelope(t, stdout.Bytes())
	if envelope["code"] != errors.CodeNoProjectSpecified {
		t.Errorf("code = %v, want %s", envelope["code"], errors.CodeNoProjectSpecified)
	}
	if envelope["message"] == "" {
		t.Error("message must be present")
	}
	if hint, _ := envelope["hint"].(string); hint == "" {
		t.Error("hint must be non-empty for no_project_specified")
	}
	if _, exists := envelope["retryable"]; !exists {
		t.Error("retryable must always be present")
	}
}

func TestRenderToTOONEnvelope(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := errors.NewNoProjectError()

	code := renderTo("toon", err, &stdout, &stderr)

	if code != errors.ExitUserError {
		t.Errorf("exit = %d, want %d", code, errors.ExitUserError)
	}
	if stderr.Len() != 0 {
		t.Errorf("structured mode must not write to stderr, got %q", stderr.String())
	}

	output := stdout.String()
	// TOON envelope carries the same snake_case fields as JSON.
	for _, want := range []string{"error:", "code: no_project_specified", "retryable: false"} {
		if !strings.Contains(output, want) {
			t.Errorf("TOON envelope missing %q, got:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Code:") || strings.Contains(output, "Message:") {
		t.Errorf("TOON envelope must not use Go field names, got:\n%s", output)
	}
}

func TestRenderToInternalErrorExit70(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := errors.NewInternalError("something broke")

	code := renderTo("json", err, &stdout, &stderr)

	if code != errors.ExitInternal {
		t.Errorf("exit = %d, want %d", code, errors.ExitInternal)
	}
	envelope := decodeEnvelope(t, stdout.Bytes())
	if envelope["code"] != errors.CodeInternal {
		t.Errorf("code = %v, want %s", envelope["code"], errors.CodeInternal)
	}
}

func TestRenderToNonCLIErrorBecomesInternal(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := stderrors.New("unexpected plain error")

	code := renderTo("json", err, &stdout, &stderr)

	if code != errors.ExitInternal {
		t.Errorf("exit = %d, want %d", code, errors.ExitInternal)
	}
	envelope := decodeEnvelope(t, stdout.Bytes())
	if envelope["code"] != errors.CodeInternal {
		t.Errorf("code = %v, want %s", envelope["code"], errors.CodeInternal)
	}
}

// Kong wraps command errors (errors.Join / fmt.Errorf %w); the CLIError and
// its code must survive the wrapping.
func TestRenderToRecoversWrappedCLIError(t *testing.T) {
	cliErr := errors.NewNoProjectError()
	wrapped := []struct {
		name string
		err  error
	}{
		{"errors.Join", stderrors.Join(cliErr, stderrors.New("kong context"))},
		{"fmt.Errorf %w", fmt.Errorf("run failed: %w", cliErr)},
	}

	for _, tt := range wrapped {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := renderTo("json", tt.err, &stdout, &stderr)

			if code != errors.ExitUserError {
				t.Errorf("exit = %d, want %d", code, errors.ExitUserError)
			}
			envelope := decodeEnvelope(t, stdout.Bytes())
			if envelope["code"] != errors.CodeNoProjectSpecified {
				t.Errorf("code = %v, want %s", envelope["code"], errors.CodeNoProjectSpecified)
			}
		})
	}
}

func TestRenderToHumanModeWritesStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := errors.NewNoProjectError()

	code := renderTo("text", err, &stdout, &stderr)

	if code != errors.ExitUserError {
		t.Errorf("exit = %d, want %d", code, errors.ExitUserError)
	}
	if stdout.Len() != 0 {
		t.Errorf("human mode must not write to stdout, got %q", stdout.String())
	}
	output := stderr.String()
	if !strings.Contains(output, "error: no project specified") {
		t.Errorf("stderr missing error line, got %q", output)
	}
	if !strings.Contains(output, "hint: ") {
		t.Errorf("stderr missing hint line, got %q", output)
	}
}

func TestFormatFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"separate -o", []string{"sol", "project:info", "-o", "json"}, "json"},
		{"separate --output", []string{"sol", "--output", "json", "project:info"}, "json"},
		{"equals --output", []string{"sol", "project:info", "--output=json"}, "json"},
		{"equals -o", []string{"sol", "project:info", "-o=toon"}, "toon"},
		{"attached -ojson", []string{"sol", "project:info", "-ojson"}, "json"},
		{"attached -otoon", []string{"sol", "project:info", "-otoon"}, "toon"},
		{"absent defaults to toon", []string{"sol", "project:info"}, "toon"},
		{"dangling -o defaults to toon", []string{"sol", "-o"}, "toon"},
		{"invalid value defaults to toon", []string{"sol", "-o", "yaml"}, "toon"},
		{"invalid equals value defaults to toon", []string{"sol", "--output=yaml"}, "toon"},
		{"invalid then valid picks valid", []string{"sol", "-o", "yaml", "-o", "json"}, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFromArgs(tt.args); got != tt.want {
				t.Errorf("formatFromArgs(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

// The schema path keeps its historical json default; the same scanner must
// honor a caller-supplied default.
func TestFormatFromArgsOrDefault(t *testing.T) {
	if got := formatFromArgsOrDefault([]string{"--schema"}, "json"); got != "json" {
		t.Errorf("absent format = %q, want json default", got)
	}
	if got := formatFromArgsOrDefault([]string{"--schema", "-o", "toon"}, "json"); got != "toon" {
		t.Errorf("explicit toon = %q, want toon", got)
	}
}
