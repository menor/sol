package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/internal/api"
)

func TestEnvironmentMergeCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		MergeEnvironmentFunc: func(ctx context.Context, projectID, environmentID string) (*api.Activity, error) {
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.merge",
				State: "pending",
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
				return "feature-x"
			}
			return ""
		},
	}

	cmd := &EnvironmentMergeCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "MergeEnvironment" {
		t.Errorf("expected MergeEnvironment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestEnvironmentMergeCmd_WithEnvironmentArg(t *testing.T) {
	mockClient := &api.MockClient{
		MergeEnvironmentFunc: func(ctx context.Context, projectID, environmentID string) (*api.Activity, error) {
			if environmentID != "staging" {
				t.Errorf("expected environment 'staging', got '%s'", environmentID)
			}
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.merge",
				State: "pending",
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

	cmd := &EnvironmentMergeCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnvironmentSyncCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		SyncEnvironmentFunc: func(ctx context.Context, projectID, environmentID string, input *api.SyncInput) (*api.Activity, error) {
			if !input.SynchronizeData {
				t.Error("expected SynchronizeData to be true")
			}
			if input.SynchronizeCode {
				t.Error("expected SynchronizeCode to be false")
			}
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.synchronize",
				State: "pending",
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
				return "staging"
			}
			return ""
		},
	}

	cmd := &EnvironmentSyncCmd{Data: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "SyncEnvironment" {
		t.Errorf("expected SyncEnvironment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestEnvironmentSyncCmd_DataAndCode(t *testing.T) {
	mockClient := &api.MockClient{
		SyncEnvironmentFunc: func(ctx context.Context, projectID, environmentID string, input *api.SyncInput) (*api.Activity, error) {
			if !input.SynchronizeData {
				t.Error("expected SynchronizeData to be true")
			}
			if !input.SynchronizeCode {
				t.Error("expected SynchronizeCode to be true")
			}
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.synchronize",
				State: "pending",
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
				return "staging"
			}
			return ""
		},
	}

	cmd := &EnvironmentSyncCmd{Data: true, Code: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnvironmentSyncCmd_NoOptions(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			if key == "PLATFORM_BRANCH" {
				return "staging"
			}
			return ""
		},
	}

	// No data, code, or resources specified
	cmd := &EnvironmentSyncCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no sync options specified")
	}
}

func TestEnvironmentMergeCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &EnvironmentMergeCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestEnvironmentSyncCmd_NoEnvironment(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "proj123"
			}
			return ""
		},
	}

	cmd := &EnvironmentSyncCmd{Data: true}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}
