package cmd

import (
	"context"
	"testing"
	"time"

	"lab.plat.farm/menor/sol/internal/api"
)

func TestRunEnvironmentList_Success(t *testing.T) {
	// Save and restore originals
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		ListEnvironmentsFunc: func(ctx context.Context, projectID string) ([]api.Environment, error) {
			return []api.Environment{
				{ID: "main", Name: "main", Type: "production", Status: "active"},
				{ID: "staging", Name: "staging", Type: "staging", Status: "active"},
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"environment:list", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListEnvironments was called with correct project ID
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListEnvironments" {
		t.Errorf("expected ListEnvironments call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
}

func TestRunEnvironmentInfo_Success(t *testing.T) {
	// Save and restore originals
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		GetEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Environment, error) {
			return &api.Environment{
				ID:        envID,
				Name:      envID,
				Type:      "production",
				Status:    "active",
				Project:   projectID,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"environment:info", "main", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetEnvironment was called with correct IDs
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetEnvironment" {
		t.Errorf("expected GetEnvironment call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
	if mockClient.Calls[0].Args[1] != "main" {
		t.Errorf("expected environment ID 'main', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestRunEnvironmentInfo_NotFound(t *testing.T) {
	// Save and restore originals
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client that returns 404
	mockClient := &api.MockClient{
		GetEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Environment, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Environment not found",
			}
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"environment:info", "nonexistent", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent environment")
	}
}

func TestRunEnvironmentList_NoProjectSpecified(t *testing.T) {
	// Save and restore originals
	originalGetEnv := getEnv
	defer func() { getEnv = originalGetEnv }()

	// Mock environment to return nothing
	getEnv = func(key string) string {
		return ""
	}

	// Execute command without project
	rootCmd.SetArgs([]string{"environment:list", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestDetectEnvironmentID(t *testing.T) {
	// Save and restore originals
	originalGetEnv := getEnv
	defer func() { getEnv = originalGetEnv }()

	tests := []struct {
		name    string
		envVars map[string]string
		want    string
	}{
		{
			name:    "from PLATFORM_BRANCH",
			envVars: map[string]string{"PLATFORM_BRANCH": "main"},
			want:    "main",
		},
		{
			name:    "empty when not set",
			envVars: map[string]string{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getEnv = func(key string) string {
				return tt.envVars[key]
			}
			got := detectEnvironmentID()
			if got != tt.want {
				t.Errorf("detectEnvironmentID() = %q, want %q", got, tt.want)
			}
		})
	}
}
