package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/internal/api"
)

func TestIntegrationListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListIntegrationsFunc: func(ctx context.Context, projectID string, opts api.ListIntegrationsOptions) ([]api.Integration, error) {
			return []api.Integration{
				{ID: "int1", Type: "github"},
				{ID: "int2", Type: "webhook"},
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

	cmd := &IntegrationListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListIntegrations" {
		t.Errorf("expected ListIntegrations call, got %s", mockClient.Calls[0].Method)
	}
}

func TestIntegrationListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListIntegrationsFunc: func(ctx context.Context, projectID string, opts api.ListIntegrationsOptions) ([]api.Integration, error) {
			return []api.Integration{
				{ID: "int1", Type: "github", Repository: "org/repo"},
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

	cmd := &IntegrationListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegrationListCmd_TypeFilter(t *testing.T) {
	mockClient := &api.MockClient{
		ListIntegrationsFunc: func(ctx context.Context, projectID string, opts api.ListIntegrationsOptions) ([]api.Integration, error) {
			if opts.Type != "github" {
				t.Errorf("expected type filter 'github', got '%s'", opts.Type)
			}
			return []api.Integration{
				{ID: "int1", Type: "github"},
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

	cmd := &IntegrationListCmd{Type: "github"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegrationGetCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetIntegrationFunc: func(ctx context.Context, projectID, integrationID string) (*api.Integration, error) {
			return &api.Integration{
				ID:         integrationID,
				Type:       "github",
				Repository: "org/repo",
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

	cmd := &IntegrationGetCmd{IntegrationID: "int123"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetIntegration" {
		t.Errorf("expected GetIntegration call, got %s", mockClient.Calls[0].Method)
	}
}

func TestIntegrationListCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &IntegrationListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
