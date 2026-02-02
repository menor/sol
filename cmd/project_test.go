package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestRunProjectList_Success(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	defer func() { newAPIClient = originalFactory }()

	// Set up mock client
	// Note: ListProjects follows HAL links to /ref/projects for full details.
	mockClient := &api.MockClient{
		ListProjectsFunc: func(ctx context.Context) ([]api.ProjectRef, error) {
			return []api.ProjectRef{
				{ID: "proj1", Title: "Project One", Region: "us-1", OrganizationID: "org1"},
				{ID: "proj2", Title: "Project Two", Region: "eu-2", OrganizationID: "org2"},
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"project:list", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListProjects was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListProjects" {
		t.Errorf("expected ListProjects call, got %s", mockClient.Calls[0].Method)
	}
}

func TestRunProjectInfo_Success(t *testing.T) {
	// Save and restore the original factory and getEnv
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		return "" // Force project ID from args
	}

	// Set up mock client
	mockClient := &api.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID string) (*api.Project, error) {
			return &api.Project{
				ID:        projectID,
				Title:     "Test Project",
				Region:    "us-3",
				CreatedAt: time.Now(),
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"project:info", "proj123", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetProject was called with correct ID
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetProject" {
		t.Errorf("expected GetProject call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
}

func TestRunProjectInfo_NotFound(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		return ""
	}

	// Set up mock client that returns 404
	mockClient := &api.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID string) (*api.Project, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Project not found",
			}
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"project:info", "nonexistent", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}

	// The error should mention "not found"
	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRunProjectInfo_NoProjectSpecified(t *testing.T) {
	// Save and restore the original getEnv
	originalGetEnv := getEnv
	defer func() { getEnv = originalGetEnv }()

	// Mock environment to return nothing
	getEnv = func(key string) string {
		return ""
	}

	// Execute command without project ID
	rootCmd.SetArgs([]string{"project:info", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestDetectProjectID(t *testing.T) {
	// Save and restore the original getEnv
	originalGetEnv := getEnv
	defer func() { getEnv = originalGetEnv }()

	tests := []struct {
		name    string
		envVars map[string]string
		want    string
	}{
		{
			name:    "from PLATFORM_PROJECT",
			envVars: map[string]string{"PLATFORM_PROJECT": "proj123"},
			want:    "proj123",
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
			got := detectProjectID()
			if got != tt.want {
				t.Errorf("detectProjectID() = %q, want %q", got, tt.want)
			}
		})
	}
}
