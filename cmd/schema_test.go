//go:build !js

package cmd

import (
	"strings"
	"testing"
)

func TestGetCommandSchema_Exists(t *testing.T) {
	schema := GetCommandSchema("project:list")
	if schema == nil {
		t.Fatal("expected schema for project:list, got nil")
	}
	if schema.Command != "project:list" {
		t.Errorf("expected command 'project:list', got %s", schema.Command)
	}
	if schema.Description == "" {
		t.Error("expected description, got empty string")
	}
	if len(schema.GlobalFlags) == 0 {
		t.Error("expected global flags, got none")
	}
	if len(schema.ExitCodes) == 0 {
		t.Error("expected exit codes, got none")
	}
}

func TestGetCommandSchema_NotExists(t *testing.T) {
	schema := GetCommandSchema("nonexistent:command")
	if schema != nil {
		t.Errorf("expected nil for nonexistent command, got %+v", schema)
	}
}

func TestListCommandSchemas(t *testing.T) {
	schemas := ListCommandSchemas()
	if len(schemas) == 0 {
		t.Fatal("expected at least one schema")
	}

	// Verify some expected commands exist
	expectedCommands := []string{
		"project:list",
		"environment:list",
		"variable:set",
		"ssh",
		"auth:login",
	}

	for _, cmd := range expectedCommands {
		if _, ok := schemas[cmd]; !ok {
			t.Errorf("expected schema for %s", cmd)
		}
	}
}

func TestCommandSchema_HasRequiredFields(t *testing.T) {
	commands := []string{
		"project:list",
		"variable:set",
		"ssh",
		"environment:branch",
	}

	for _, cmd := range commands {
		schema := GetCommandSchema(cmd)
		if schema == nil {
			t.Errorf("missing schema for %s", cmd)
			continue
		}

		if schema.Command == "" {
			t.Errorf("%s: missing command name", cmd)
		}
		if schema.Description == "" {
			t.Errorf("%s: missing description", cmd)
		}
		if len(schema.GlobalFlags) == 0 {
			t.Errorf("%s: missing global flags", cmd)
		}
		if len(schema.ExitCodes) == 0 {
			t.Errorf("%s: missing exit codes", cmd)
		}
	}
}

func TestHasSchemaFlag(t *testing.T) {
	tests := []struct {
		args     []string
		expected bool
	}{
		{[]string{"sol", "project:list", "--schema"}, true},
		{[]string{"sol", "--schema", "project:list"}, true},
		{[]string{"sol", "project:list"}, false},
		{[]string{"sol", "project:list", "--output", "json"}, false},
		{[]string{"sol", "--schema"}, true},
	}

	for _, tc := range tests {
		result := hasSchemaFlag(tc.args)
		if result != tc.expected {
			t.Errorf("hasSchemaFlag(%v) = %v, want %v", tc.args, result, tc.expected)
		}
	}
}

func TestIsFlag(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"--schema", true},
		{"-o", true},
		{"project:list", false},
		{"abc123", false},
		{"-", true},
		{"", false},
	}

	for _, tc := range tests {
		result := isFlag(tc.arg)
		if result != tc.expected {
			t.Errorf("isFlag(%q) = %v, want %v", tc.arg, result, tc.expected)
		}
	}
}

func TestHandleSchemaRequest_UnknownCommand(t *testing.T) {
	// Test that unknown commands return an error
	err := handleSchemaRequest([]string{"sol", "unknown:cmd", "--schema"})
	if err == nil {
		t.Error("expected error for unknown command")
	}
	// Error should mention the unknown command
	if err != nil && !strings.Contains(err.Error(), "unknown:cmd") {
		t.Errorf("error should mention unknown command, got: %s", err.Error())
	}
}

func TestListAllSchemas_Sorted(t *testing.T) {
	// This is a simple check that listAllSchemas doesn't error
	// The sort order is verified by running sol --schema manually
	err := listAllSchemas("json")
	if err != nil {
		t.Errorf("listAllSchemas failed: %v", err)
	}
}
