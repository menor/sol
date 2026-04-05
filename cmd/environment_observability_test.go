package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestEnvironmentURLCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetRoutesFunc: func(ctx context.Context, projectID, envID string) ([]api.RouteURL, error) {
			return []api.RouteURL{
				{URL: "https://main-abc123.example.com/", Primary: true, Type: "upstream"},
				{URL: "https://www.main-abc123.example.com/", Primary: false, Type: "redirect"},
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

	cmd := &EnvironmentURLCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetRoutes was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetRoutes" {
		t.Errorf("expected GetRoutes call, got %s", mockClient.Calls[0].Method)
	}
}

func TestEnvironmentURLCmd_Primary(t *testing.T) {
	mockClient := &api.MockClient{
		GetRoutesFunc: func(ctx context.Context, projectID, envID string) ([]api.RouteURL, error) {
			return []api.RouteURL{
				{URL: "https://www.main-abc123.example.com/", Primary: false, Type: "redirect"},
				{URL: "https://main-abc123.example.com/", Primary: true, Type: "upstream"},
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

	// Test with --primary flag
	cmd := &EnvironmentURLCmd{Primary: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetRoutes was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
}

func TestEnvironmentURLCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &EnvironmentURLCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestEnvironmentRelationshipsCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetRelationshipsFunc: func(ctx context.Context, projectID, envID, appName string) ([]api.Relationship, error) {
			return []api.Relationship{
				{App: "frontend", Name: "database", Service: "mysql", Endpoint: "mysql:3306"},
				{App: "frontend", Name: "cache", Service: "redis", Endpoint: "redis:6379"},
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

	cmd := &EnvironmentRelationshipsCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetRelationships was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetRelationships" {
		t.Errorf("expected GetRelationships call, got %s", mockClient.Calls[0].Method)
	}
}

func TestEnvironmentRelationshipsCmd_WithAppFilter(t *testing.T) {
	mockClient := &api.MockClient{
		GetRelationshipsFunc: func(ctx context.Context, projectID, envID, appName string) ([]api.Relationship, error) {
			if appName != "frontend" {
				t.Errorf("expected app filter 'frontend', got %q", appName)
			}
			return []api.Relationship{
				{App: "frontend", Name: "database", Service: "mysql", Endpoint: "mysql:3306"},
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

	// Test with --app filter
	cmd := &EnvironmentRelationshipsCmd{App: "frontend"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetRelationships was called with app filter
	if mockClient.Calls[0].Args[2] != "frontend" {
		t.Errorf("expected app filter 'frontend', got %v", mockClient.Calls[0].Args[2])
	}
}

func TestEnvironmentRelationshipsCmd_EmptyResult(t *testing.T) {
	mockClient := &api.MockClient{
		GetRelationshipsFunc: func(ctx context.Context, projectID, envID, appName string) ([]api.Relationship, error) {
			return nil, nil // No relationships
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

	cmd := &EnvironmentRelationshipsCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnvironmentRelationshipsCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &EnvironmentRelationshipsCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestEnvironmentRelationshipsCmd_APIError(t *testing.T) {
	mockClient := &api.MockClient{
		GetRelationshipsFunc: func(ctx context.Context, projectID, envID, appName string) ([]api.Relationship, error) {
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

	cmd := &EnvironmentRelationshipsCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}
