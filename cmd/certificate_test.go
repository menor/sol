package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestCertificateListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListCertificatesFunc: func(ctx context.Context, projectID string) ([]api.Certificate, error) {
			return []api.Certificate{
				{ID: "cert1", Domains: []string{"example.com"}, IsProvisioned: true},
				{ID: "cert2", Domains: []string{"other.com"}, IsProvisioned: false},
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

	cmd := &CertificateListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListCertificates" {
		t.Errorf("expected ListCertificates call, got %s", mockClient.Calls[0].Method)
	}
}

func TestCertificateListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListCertificatesFunc: func(ctx context.Context, projectID string) ([]api.Certificate, error) {
			return []api.Certificate{
				{
					ID:            "cert1",
					Domains:       []string{"example.com"},
					IsProvisioned: true,
					ExpiresAt:     "2025-12-31T00:00:00Z",
					Chain:         []string{"intermediate-cert"},
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

	cmd := &CertificateListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertificateListCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &CertificateListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
