package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestVariableListCmd_ProjectLevel(t *testing.T) {
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

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return ""
		},
	}

	cmd := &VariableListCmd{} // No level = defaults to project
	err := cmd.Run(ctx)
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

func TestVariableListCmd_EnvironmentLevel(t *testing.T) {
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

	cli := &CLI{Environment: "main"} // Environment flag set
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	cmd := &VariableListCmd{} // Level auto-detected from environment flag
	err := cmd.Run(ctx)
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

func TestVariableGetCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetProjectVariableFunc: func(ctx context.Context, projectID, name string) (*api.Variable, error) {
			return &api.Variable{
				ID:        "var1",
				Name:      name,
				Value:     "test-value",
				IsEnabled: true,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return ""
		},
	}

	cmd := &VariableGetCmd{Name: "MY_VAR", Level: "project"}
	err := cmd.Run(ctx)
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

func TestVariableSetCmd_Success(t *testing.T) {
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

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return ""
		},
	}

	cmd := &VariableSetCmd{
		Name:           "MY_VAR",
		Value:          "my-value",
		Level:          "project",
		VisibleBuild:   true,
		VisibleRuntime: true,
	}
	err := cmd.Run(ctx)
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

func TestVariableDeleteCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		DeleteProjectVariableFunc: func(ctx context.Context, projectID, name string) error {
			return nil
		},
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return ""
		},
	}

	cmd := &VariableDeleteCmd{Name: "MY_VAR", Level: "project"}
	err := cmd.Run(ctx)
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

func TestVariableGetCmd_NotFound(t *testing.T) {
	mockClient := &api.MockClient{
		GetProjectVariableFunc: func(ctx context.Context, projectID, name string) (*api.Variable, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Variable not found",
			}
		},
	}

	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return ""
		},
	}

	cmd := &VariableGetCmd{Name: "NONEXISTENT", Level: "project"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent variable")
	}
}

func TestVariableListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project in environment
		},
	}

	cmd := &VariableListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
