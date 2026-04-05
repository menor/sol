package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestServiceListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListServicesFunc: func(ctx context.Context, projectID, envID string) ([]api.ServiceSummary, error) {
			return []api.ServiceSummary{
				{Name: "mysql", Type: "mysql:10.4", Disk: 2048},
				{Name: "redis", Type: "redis:6.0"},
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

	cmd := &ServiceListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListServices was called with correct IDs
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListServices" {
		t.Errorf("expected ListServices call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
	if mockClient.Calls[0].Args[1] != "main" {
		t.Errorf("expected environment ID 'main', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestServiceListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		GetCurrentDeploymentFunc: func(ctx context.Context, projectID, envID string) (*api.Deployment, error) {
			return &api.Deployment{
				ID: "deploy123",
				Services: map[string]api.Service{
					"mysql": {Type: "mysql:10.4", Size: "AUTO", Disk: 2048},
					"redis": {Type: "redis:6.0", Size: "AUTO"},
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
	cmd := &ServiceListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetCurrentDeployment was called (not ListServices)
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetCurrentDeployment" {
		t.Errorf("expected GetCurrentDeployment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestServiceListCmd_WithEnvironmentArg(t *testing.T) {
	mockClient := &api.MockClient{
		ListServicesFunc: func(ctx context.Context, projectID, envID string) ([]api.ServiceSummary, error) {
			return []api.ServiceSummary{
				{Name: "mysql", Type: "mysql:10.4", Disk: 2048},
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
	cmd := &ServiceListCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify called with correct environment ID
	if mockClient.Calls[0].Args[1] != "staging" {
		t.Errorf("expected environment ID 'staging', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestServiceListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project or environment
		},
	}

	cmd := &ServiceListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestServiceListCmd_NoEnvironmentSpecified(t *testing.T) {
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

	cmd := &ServiceListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}

func TestServiceListCmd_APIError(t *testing.T) {
	mockClient := &api.MockClient{
		ListServicesFunc: func(ctx context.Context, projectID, envID string) ([]api.ServiceSummary, error) {
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

	cmd := &ServiceListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}
