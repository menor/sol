package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAppConfigValidateCmd_ValidFile(t *testing.T) {
	// Create a temp directory with a valid config
	tmpDir := t.TempDir()
	upsunDir := filepath.Join(tmpDir, ".upsun")
	if err := os.MkdirAll(upsunDir, 0755); err != nil {
		t.Fatalf("failed to create .upsun dir: %v", err)
	}

	configPath := filepath.Join(upsunDir, "config.yaml")
	configData := []byte(`
applications:
  app:
    type: nodejs:18
    web:
      commands:
        start: npm start
`)
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	cmd := &AppConfigValidateCmd{Path: tmpDir}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppConfigValidateCmd_DirectFile(t *testing.T) {
	// Create a temp file with valid config
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configData := []byte(`
applications:
  app:
    type: python:3.11
`)
	if _, err := tmpFile.Write(configData); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	tmpFile.Close()

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	cmd := &AppConfigValidateCmd{Path: tmpFile.Name()}
	err = cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppConfigValidateCmd_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	cmd := &AppConfigValidateCmd{Path: tmpDir}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestAppConfigValidateCmd_NonexistentPath(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	cmd := &AppConfigValidateCmd{Path: "/nonexistent/path/that/does/not/exist"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestAppConfigValidateCmd_InvalidConfig(t *testing.T) {
	// Create a temp file with invalid config (missing applications)
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configData := []byte(`
services:
  db:
    type: mysql:8.0
`)
	if _, err := tmpFile.Write(configData); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	tmpFile.Close()

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	// Command should succeed (validation errors are in output, not returned as error)
	cmd := &AppConfigValidateCmd{Path: tmpFile.Name()}
	err = cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
