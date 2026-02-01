package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestRunVariableList_ProjectLevel(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		ListProjectVariablesFunc: func(ctx context.Context, projectID string) ([]api.Variable, error) {
			return []api.Variable{
				{
					ID:             "var1",
					Name:           "API_KEY",
					IsSensitive:    true,
					IsEnabled:      true,
					VisibleBuild:   true,
					VisibleRuntime: true,
					CreatedAt:      time.Now(),
				},
				{
					ID:             "var2",
					Name:           "DEBUG",
					Value:          "false",
					IsEnabled:      true,
					VisibleBuild:   false,
					VisibleRuntime: true,
					CreatedAt:      time.Now(),
				},
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"variable:list", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListProjectVariables was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListProjectVariables" {
		t.Errorf("expected ListProjectVariables call, got %s", mockClient.Calls[0].Method)
	}
}

func TestRunVariableList_EnvironmentLevel(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		ListEnvironmentVariablesFunc: func(ctx context.Context, projectID, envID string) ([]api.Variable, error) {
			return []api.Variable{
				{
					ID:          "var1",
					Name:        "ENV_VAR",
					Value:       "value",
					Environment: envID,
					CreatedAt:   time.Now(),
				},
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command with --environment flag
	rootCmd.SetArgs([]string{"variable:list", "--environment", "main", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListEnvironmentVariables was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListEnvironmentVariables" {
		t.Errorf("expected ListEnvironmentVariables call, got %s", mockClient.Calls[0].Method)
	}
}

func TestRunVariableGet_Success(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		GetProjectVariableFunc: func(ctx context.Context, projectID, name string) (*api.Variable, error) {
			return &api.Variable{
				ID:          "var1",
				Name:        name,
				Value:       "test-value",
				IsEnabled:   true,
				CreatedAt:   time.Now(),
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command (explicitly set --level to override any persisting --environment flag)
	rootCmd.SetArgs([]string{"variable:get", "MY_VAR", "--level", "project", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetProjectVariable was called with correct name
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetProjectVariable" {
		t.Errorf("expected GetProjectVariable call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[1] != "MY_VAR" {
		t.Errorf("expected variable name 'MY_VAR', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestRunVariableSet_Success(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		SetProjectVariableFunc: func(ctx context.Context, projectID string, input *api.VariableInput) (*api.Variable, error) {
			return &api.Variable{
				ID:        "var1",
				Name:      input.Name,
				Value:     input.Value,
				IsEnabled: true,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command (explicitly set --level to override any persisting --environment flag)
	rootCmd.SetArgs([]string{"variable:set", "MY_VAR", "my-value", "--level", "project", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify SetProjectVariable was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "SetProjectVariable" {
		t.Errorf("expected SetProjectVariable call, got %s", mockClient.Calls[0].Method)
	}

	// Verify the input
	input := mockClient.Calls[0].Args[1].(*api.VariableInput)
	if input.Name != "MY_VAR" {
		t.Errorf("expected variable name 'MY_VAR', got %v", input.Name)
	}
	if input.Value != "my-value" {
		t.Errorf("expected variable value 'my-value', got %v", input.Value)
	}
}

func TestRunVariableDelete_Success(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		DeleteProjectVariableFunc: func(ctx context.Context, projectID, name string) error {
			return nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command (explicitly set --level to override any persisting --environment flag)
	rootCmd.SetArgs([]string{"variable:delete", "MY_VAR", "--level", "project", "--output", "json"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify DeleteProjectVariable was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "DeleteProjectVariable" {
		t.Errorf("expected DeleteProjectVariable call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[1] != "MY_VAR" {
		t.Errorf("expected variable name 'MY_VAR', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestRunVariableGet_NotFound(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client that returns 404
	mockClient := &api.MockClient{
		GetProjectVariableFunc: func(ctx context.Context, projectID, name string) (*api.Variable, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Variable not found",
			}
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command (explicitly set --level to override any persisting --environment flag)
	rootCmd.SetArgs([]string{"variable:get", "NONEXISTENT", "--level", "project", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent variable")
	}
}

func TestRunVariableList_NoProjectSpecified(t *testing.T) {
	// Save and restore the original getEnv
	originalGetEnv := getEnv
	defer func() { getEnv = originalGetEnv }()

	// Reset global flag state (Cobra flags persist between tests)
	varLevel = ""

	// Mock environment to return nothing
	getEnv = func(key string) string {
		return ""
	}

	// Execute command without project
	rootCmd.SetArgs([]string{"variable:list", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
