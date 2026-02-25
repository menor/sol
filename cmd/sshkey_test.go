package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/internal/api"
)

func TestSSHKeyListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListSSHKeysFunc: func(ctx context.Context) ([]api.SSHKey, error) {
			return []api.SSHKey{
				{KeyID: "key1", Title: "Work laptop", Fingerprint: "SHA256:abc123"},
				{KeyID: "key2", Title: "Home desktop", Fingerprint: "SHA256:def456"},
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
			return ""
		},
	}

	cmd := &SSHKeyListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListSSHKeys" {
		t.Errorf("expected ListSSHKeys call, got %s", mockClient.Calls[0].Method)
	}
}

func TestSSHKeyListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListSSHKeysFunc: func(ctx context.Context) ([]api.SSHKey, error) {
			return []api.SSHKey{
				{
					KeyID:       "key1",
					Title:       "Work laptop",
					Fingerprint: "SHA256:abc123",
					Value:       "ssh-ed25519 AAAAC3...",
					Changed:     "2024-01-01T00:00:00Z",
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
			return ""
		},
	}

	cmd := &SSHKeyListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyListCmd_Empty(t *testing.T) {
	mockClient := &api.MockClient{
		ListSSHKeysFunc: func(ctx context.Context) ([]api.SSHKey, error) {
			return []api.SSHKey{}, nil
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
			return ""
		},
	}

	cmd := &SSHKeyListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
