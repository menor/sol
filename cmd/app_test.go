package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestAppListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListAppsFunc: func(ctx context.Context, projectID, envID string) ([]api.AppSummary, error) {
			return []api.AppSummary{
				{Name: "frontend", Type: "nodejs:18", Size: "AUTO", Disk: 512},
				{Name: "worker", Type: "python:3.11", Size: "AUTO", Worker: true},
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
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	cmd := &AppListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListApps was called with correct IDs
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListApps" {
		t.Errorf("expected ListApps call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
	if mockClient.Calls[0].Args[1] != "main" {
		t.Errorf("expected environment ID 'main', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestAppListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		GetCurrentDeploymentFunc: func(ctx context.Context, projectID, envID string) (*api.Deployment, error) {
			return &api.Deployment{
				ID: "deploy123",
				Webapps: map[string]api.Webapp{
					"frontend": {Type: "nodejs:18", Size: "AUTO", Disk: 512},
				},
				Workers: map[string]api.Worker{
					"worker": {Type: "python:3.11", Size: "AUTO"},
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
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	// Test with --full flag
	cmd := &AppListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetCurrentDeployment was called (not ListApps)
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetCurrentDeployment" {
		t.Errorf("expected GetCurrentDeployment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestAppListCmd_WithEnvironmentArg(t *testing.T) {
	mockClient := &api.MockClient{
		ListAppsFunc: func(ctx context.Context, projectID, envID string) ([]api.AppSummary, error) {
			return []api.AppSummary{
				{Name: "app", Type: "php:8.2"},
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

	// Environment specified as positional arg
	cmd := &AppListCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify called with correct environment ID
	if mockClient.Calls[0].Args[1] != "staging" {
		t.Errorf("expected environment ID 'staging', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestAppListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project or environment
		},
	}

	cmd := &AppListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestAppListCmd_NoEnvironmentSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return "" // No environment
		},
	}

	cmd := &AppListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}

func TestAppListCmd_APIError(t *testing.T) {
	mockClient := &api.MockClient{
		ListAppsFunc: func(ctx context.Context, projectID, envID string) ([]api.AppSummary, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Environment not found",
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
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	cmd := &AppListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}
