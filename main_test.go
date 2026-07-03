package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// solBinary is the path to the sol binary built once in TestMain. The
// subprocess layer is deliberately thin: it guards real exit codes and stream
// routing; envelope shape and format details live in the render() unit tests.
var solBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "sol-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdtemp: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	solBinary = filepath.Join(dir, "sol")
	build := exec.Command("go", "build", "-o", solBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runSol executes the built binary with a project-free environment and
// returns stdout, stderr, and the exit code.
func runSol(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(solBinary, args...)
	// Strip project/environment context so "no project" paths are
	// deterministic regardless of the host machine.
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		switch key {
		case "PLATFORM_PROJECT", "UPSUN_PROJECT", "PLATFORM_BRANCH", "UPSUN_ENVIRONMENT":
		default:
			cmd.Env = append(cmd.Env, e)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run %v: %v", args, err)
		}
		exitCode = exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), exitCode
}

func decodeErrorEnvelope(t *testing.T, stdout string) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	envelope, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatalf("stdout missing 'error' wrapper: %s", stdout)
	}
	return envelope
}

func TestNoProjectJSON(t *testing.T) {
	stdout, stderr, exitCode := runSol(t, "project:info", "-o", "json")

	if exitCode != 1 {
		t.Errorf("exit = %d, want 1", exitCode)
	}
	if stderr != "" {
		t.Errorf("structured mode must keep stderr empty, got %q", stderr)
	}
	envelope := decodeErrorEnvelope(t, stdout)
	if envelope["code"] != "no_project_specified" {
		t.Errorf("code = %v, want no_project_specified", envelope["code"])
	}
	if hint, _ := envelope["hint"].(string); hint == "" {
		t.Error("hint must be non-empty")
	}
}

func TestNoProjectTOON(t *testing.T) {
	stdout, stderr, exitCode := runSol(t, "project:info", "-o", "toon")

	if exitCode != 1 {
		t.Errorf("exit = %d, want 1", exitCode)
	}
	if stderr != "" {
		t.Errorf("structured mode must keep stderr empty, got %q", stderr)
	}
	for _, want := range []string{"error:", "code: no_project_specified", "retryable: false"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("TOON envelope missing %q, got:\n%s", want, stdout)
		}
	}
}

func TestBogusCommandExits80(t *testing.T) {
	stdout, stderr, exitCode := runSol(t, "bogus:command", "-o", "json")

	if exitCode != 80 {
		t.Errorf("exit = %d, want 80", exitCode)
	}
	if stderr != "" {
		t.Errorf("structured mode must keep stderr empty, got %q", stderr)
	}
	envelope := decodeErrorEnvelope(t, stdout)
	if envelope["code"] != "invalid_argument" {
		t.Errorf("code = %v, want invalid_argument", envelope["code"])
	}
}

// A typo'd command with --schema is the same mistake as a typo'd command
// without it: exit 80, invalid_argument. No -o flag: the error must follow
// the schema path's json default, not the global toon default.
func TestSchemaUnknownCommandExits80(t *testing.T) {
	stdout, _, exitCode := runSol(t, "projct:list", "--schema")

	if exitCode != 80 {
		t.Errorf("exit = %d, want 80", exitCode)
	}
	envelope := decodeErrorEnvelope(t, stdout)
	if envelope["code"] != "invalid_argument" {
		t.Errorf("code = %v, want invalid_argument", envelope["code"])
	}
}

// Kong accepts the attached short-flag form; the parse-error render path must
// honor it too, not fall back to TOON.
func TestBogusCommandAttachedFormatFlag(t *testing.T) {
	stdout, _, exitCode := runSol(t, "bogus:command", "-ojson")

	if exitCode != 80 {
		t.Errorf("exit = %d, want 80", exitCode)
	}
	decodeErrorEnvelope(t, stdout) // fails the test if stdout is not JSON
}

func TestVersionUnchanged(t *testing.T) {
	stdout, _, exitCode := runSol(t, "version")

	if exitCode != 0 {
		t.Errorf("exit = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "sol") {
		t.Errorf("version output missing binary name, got %q", stdout)
	}
}

func TestSchemaUnchanged(t *testing.T) {
	stdout, _, exitCode := runSol(t, "project:list", "--schema")

	if exitCode != 0 {
		t.Errorf("exit = %d, want 0", exitCode)
	}
	var schema map[string]any
	if err := json.Unmarshal([]byte(stdout), &schema); err != nil {
		t.Fatalf("--schema output is not valid JSON: %v", err)
	}
	if schema["command"] != "project:list" {
		t.Errorf("schema.command = %v, want project:list", schema["command"])
	}
}
