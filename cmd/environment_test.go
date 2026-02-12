package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestEnvironmentListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListEnvironmentsFunc: func(ctx context.Context, projectID string) ([]api.Environment, error) {
			return []api.Environment{
				{ID: "main", Name: "main", Type: "production", Status: "active"},
				{ID: "staging", Name: "staging", Type: "staging", Status: "active"},
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentListCmd{}
	err := cmd.Run(ctx)
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

func TestEnvironmentListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListEnvironmentsFunc: func(ctx context.Context, projectID string) ([]api.Environment, error) {
			return []api.Environment{
				{ID: "main", Name: "main", Type: "production", Status: "active", MachineName: "main-abc123"},
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	// Test with --full flag
	cmd := &EnvironmentListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListEnvironments was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
}

func TestEnvironmentInfoCmd_Success(t *testing.T) {
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

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentInfoCmd{EnvironmentID: "main"}
	err := cmd.Run(ctx)
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

func TestEnvironmentInfoCmd_NotFound(t *testing.T) {
	mockClient := &api.MockClient{
		GetEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Environment, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Environment not found",
			}
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentInfoCmd{EnvironmentID: "nonexistent"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent environment")
	}
}

func TestEnvironmentListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{Output: "json"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project in environment
		},
	}

	cmd := &EnvironmentListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestEnvironmentID_FromEnvironment(t *testing.T) {
	cli := &CLI{Output: "json"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	got := ctx.EnvironmentID()
	if got != "main" {
		t.Errorf("EnvironmentID() = %q, want %q", got, "main")
	}
}

func TestEnvironmentListCmd_StatusFilter(t *testing.T) {
	mockClient := &api.MockClient{
		ListEnvironmentsFunc: func(ctx context.Context, projectID string) ([]api.Environment, error) {
			return []api.Environment{
				{ID: "main", Name: "main", Type: "production", Status: "active"},
				{ID: "staging", Name: "staging", Type: "staging", Status: "active"},
				{ID: "old-feature", Name: "old-feature", Type: "development", Status: "inactive"},
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	// Filter by active status
	cmd := &EnvironmentListCmd{Status: "active"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnvironmentListCmd_NoInactiveFilter(t *testing.T) {
	mockClient := &api.MockClient{
		ListEnvironmentsFunc: func(ctx context.Context, projectID string) ([]api.Environment, error) {
			return []api.Environment{
				{ID: "main", Name: "main", Type: "production", Status: "active"},
				{ID: "old-feature", Name: "old-feature", Type: "development", Status: "inactive"},
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	// Exclude inactive environments
	cmd := &EnvironmentListCmd{NoInactive: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnvironmentListCmd_TypeFilter(t *testing.T) {
	mockClient := &api.MockClient{
		ListEnvironmentsFunc: func(ctx context.Context, projectID string) ([]api.Environment, error) {
			return []api.Environment{
				{ID: "main", Name: "main", Type: "production", Status: "active"},
				{ID: "staging", Name: "staging", Type: "staging", Status: "active"},
				{ID: "feature", Name: "feature", Type: "development", Status: "active"},
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	// Filter by production type
	cmd := &EnvironmentListCmd{Type: "production"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
