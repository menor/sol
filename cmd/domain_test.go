package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestDomainListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListDomainsFunc: func(ctx context.Context, projectID string) ([]api.Domain, error) {
			return []api.Domain{
				{ID: "dom1", Name: "example.com", IsDefault: true},
				{ID: "dom2", Name: "www.example.com", IsDefault: false},
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

	cmd := &DomainListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListDomains" {
		t.Errorf("expected ListDomains call, got %s", mockClient.Calls[0].Method)
	}
}

func TestDomainListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListDomainsFunc: func(ctx context.Context, projectID string) ([]api.Domain, error) {
			return []api.Domain{
				{ID: "dom1", Name: "example.com", IsDefault: true, CreatedAt: "2024-01-01T00:00:00Z"},
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

	cmd := &DomainListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDomainListCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &DomainListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
