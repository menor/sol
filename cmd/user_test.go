package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestUserListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListProjectUsersFunc: func(ctx context.Context, projectID string) ([]api.ProjectUserAccess, error) {
			return []api.ProjectUserAccess{
				{UserID: "user1", Email: "user1@example.com", Role: "admin"},
				{UserID: "user2", Email: "user2@example.com", Role: "viewer"},
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

	cmd := &UserListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListProjectUsers" {
		t.Errorf("expected ListProjectUsers call, got %s", mockClient.Calls[0].Method)
	}
}

func TestUserListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListProjectUsersFunc: func(ctx context.Context, projectID string) ([]api.ProjectUserAccess, error) {
			return []api.ProjectUserAccess{
				{UserID: "user1", Email: "user1@example.com", Role: "admin", Permissions: []string{"admin", "viewer"}},
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

	cmd := &UserListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserListCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &UserListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
