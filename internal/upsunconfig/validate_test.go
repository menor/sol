package upsunconfig

import (
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	data := []byte(`
applications:
  app:
    type: nodejs:18
    web:
      commands:
        start: npm start
services:
  db:
    type: postgresql:15
routes:
  "https://{default}/":
    type: upstream
    upstream: app:http
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
	if len(result.Apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(result.Apps))
	}
	if result.Apps[0].Name != "app" {
		t.Errorf("expected app name 'app', got '%s'", result.Apps[0].Name)
	}
	if result.Apps[0].Type != "nodejs:18" {
		t.Errorf("expected type 'nodejs:18', got '%s'", result.Apps[0].Type)
	}
	if len(result.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(result.Services))
	}
	if result.Routes != 1 {
		t.Errorf("expected 1 route, got %d", result.Routes)
	}
}

func TestValidate_MissingApplications(t *testing.T) {
	data := []byte(`
services:
  db:
    type: mysql:8.0
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid config for missing applications")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Type != "schema" {
		t.Errorf("expected schema error, got %s", result.Errors[0].Type)
	}
}

func TestValidate_InvalidYAML(t *testing.T) {
	data := []byte(`
applications:
  app:
    type: nodejs:18
    - invalid yaml here
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid config for bad YAML")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Type != "syntax" {
		t.Errorf("expected syntax error, got %s", result.Errors[0].Type)
	}
}

func TestValidate_MissingType(t *testing.T) {
	data := []byte(`
applications:
  app:
    web:
      commands:
        start: npm start
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid config for missing type")
	}
	if len(result.Apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(result.Apps))
	}
	if result.Apps[0].Valid {
		t.Error("expected app to be invalid")
	}
}

func TestValidate_InvalidAppName(t *testing.T) {
	data := []byte(`
applications:
  "my-app!":
    type: nodejs:18
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid config for bad app name")
	}
	if len(result.Apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(result.Apps))
	}
	if result.Apps[0].Valid {
		t.Error("expected app to be invalid")
	}
}

func TestValidate_InvalidTypeFormat(t *testing.T) {
	data := []byte(`
applications:
  app:
    type: nodejs
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid config for bad type format")
	}
	if len(result.Apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(result.Apps))
	}
	if result.Apps[0].Valid {
		t.Error("expected app to be invalid")
	}
}

func TestValidate_UnknownTopLevelKey(t *testing.T) {
	data := []byte(`
applications:
  app:
    type: nodejs:18
unknown_key: value
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid config (unknown keys are warnings)")
	}
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestValidate_UnknownServiceType(t *testing.T) {
	data := []byte(`
applications:
  app:
    type: nodejs:18
services:
  custom:
    type: unknown-service:1.0
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid config (unknown service types are warnings)")
	}
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestValidate_MultipleApps(t *testing.T) {
	data := []byte(`
applications:
  frontend:
    type: nodejs:18
  backend:
    type: python:3.11
  worker:
    type: python:3.11
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid config, got errors: %v", result.Errors)
	}
	if len(result.Apps) != 3 {
		t.Errorf("expected 3 apps, got %d", len(result.Apps))
	}
	// Verify sorted order
	if result.Apps[0].Name != "backend" {
		t.Errorf("expected first app 'backend', got '%s'", result.Apps[0].Name)
	}
	if result.Apps[1].Name != "frontend" {
		t.Errorf("expected second app 'frontend', got '%s'", result.Apps[1].Name)
	}
	if result.Apps[2].Name != "worker" {
		t.Errorf("expected third app 'worker', got '%s'", result.Apps[2].Name)
	}
}

func TestValidate_EmptyApplications(t *testing.T) {
	data := []byte(`
applications: {}
`)
	result, err := Validate(data, ".upsun/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid config for empty applications")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}
