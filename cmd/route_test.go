package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/internal/api"
)

func TestRouteListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListRoutesFunc: func(ctx context.Context, projectID, envID string) ([]api.RouteSummary, error) {
			return []api.RouteSummary{
				{URL: "https://main-abc123.example.com/", Primary: true, Type: "upstream", Upstream: "app:http"},
				{URL: "https://www.main-abc123.example.com/", Primary: false, Type: "redirect", To: "https://main-abc123.example.com/"},
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

	cmd := &RouteListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListRoutes was called with correct IDs
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListRoutes" {
		t.Errorf("expected ListRoutes call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
	if mockClient.Calls[0].Args[1] != "main" {
		t.Errorf("expected environment ID 'main', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestRouteListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		GetCurrentDeploymentFunc: func(ctx context.Context, projectID, envID string) (*api.Deployment, error) {
			return &api.Deployment{
				ID: "deploy123",
				Routes: map[string]api.Route{
					"https://main-abc123.example.com/": {
						Primary:     true,
						Type:        "upstream",
						OriginalURL: "https://{default}/",
						Upstream:    "app:http",
						TLS: &api.RouteTLS{
							StrictTransportSecurity: &api.StrictTransportSecurity{
								Enabled:           true,
								IncludeSubdomains: true,
							},
						},
					},
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
	cmd := &RouteListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetCurrentDeployment was called (not ListRoutes)
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetCurrentDeployment" {
		t.Errorf("expected GetCurrentDeployment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestRouteListCmd_WithEnvironmentArg(t *testing.T) {
	mockClient := &api.MockClient{
		ListRoutesFunc: func(ctx context.Context, projectID, envID string) ([]api.RouteSummary, error) {
			return []api.RouteSummary{
				{URL: "https://staging-abc123.example.com/", Primary: true, Type: "upstream"},
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
	cmd := &RouteListCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify called with correct environment ID
	if mockClient.Calls[0].Args[1] != "staging" {
		t.Errorf("expected environment ID 'staging', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestRouteListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project or environment
		},
	}

	cmd := &RouteListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestRouteListCmd_NoEnvironmentSpecified(t *testing.T) {
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

	cmd := &RouteListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}

func TestRouteListCmd_APIError(t *testing.T) {
	mockClient := &api.MockClient{
		ListRoutesFunc: func(ctx context.Context, projectID, envID string) ([]api.RouteSummary, error) {
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

	cmd := &RouteListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}
