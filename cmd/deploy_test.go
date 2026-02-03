package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestEnvironmentActivateCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ActivateEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Activity, error) {
			return &api.Activity{
				ID:          "act123",
				Type:        "environment.activate",
				State:       "pending",
				Description: "Activating environment",
				Project:     projectID,
				Environments: []string{envID},
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentActivateCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ActivateEnvironment was called with correct IDs
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ActivateEnvironment" {
		t.Errorf("expected ActivateEnvironment call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
	if mockClient.Calls[0].Args[1] != "staging" {
		t.Errorf("expected environment ID 'staging', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestEnvironmentActivateCmd_NoProject(t *testing.T) {
	cli := &CLI{Output: "json"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project in environment
		},
	}

	cmd := &EnvironmentActivateCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestEnvironmentActivateCmd_NoEnvironment(t *testing.T) {
	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentActivateCmd{} // No environment ID
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}

func TestEnvironmentDeactivateCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		DeactivateEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Activity, error) {
			return &api.Activity{
				ID:          "act456",
				Type:        "environment.deactivate",
				State:       "pending",
				Description: "Deactivating environment",
				Project:     projectID,
				Environments: []string{envID},
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentDeactivateCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify DeactivateEnvironment was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "DeactivateEnvironment" {
		t.Errorf("expected DeactivateEnvironment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestEnvironmentDeleteCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		DeleteEnvironmentFunc: func(ctx context.Context, projectID, envID string) error {
			return nil
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentDeleteCmd{EnvironmentID: "old-feature", Yes: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify DeleteEnvironment was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "DeleteEnvironment" {
		t.Errorf("expected DeleteEnvironment call, got %s", mockClient.Calls[0].Method)
	}
}

func TestEnvironmentDeleteCmd_APIError(t *testing.T) {
	mockClient := &api.MockClient{
		DeleteEnvironmentFunc: func(ctx context.Context, projectID, envID string) error {
			return &api.APIError{
				StatusCode: 400,
				Message:    "Environment must be deactivated before deletion",
			}
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &EnvironmentDeleteCmd{EnvironmentID: "active-env", Yes: true}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when environment is still active")
	}
}

func TestRedeployCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		RedeployEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Activity, error) {
			return &api.Activity{
				ID:          "act789",
				Type:        "environment.redeploy",
				State:       "pending",
				Description: "Redeploying environment",
				Project:     projectID,
				Environments: []string{envID},
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	// Test using environment from PLATFORM_BRANCH
	cmd := &RedeployCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify RedeployEnvironment was called with main (from env var)
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "RedeployEnvironment" {
		t.Errorf("expected RedeployEnvironment call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[1] != "main" {
		t.Errorf("expected environment ID 'main', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestRedeployCmd_WithExplicitEnvironment(t *testing.T) {
	mockClient := &api.MockClient{
		RedeployEnvironmentFunc: func(ctx context.Context, projectID, envID string) (*api.Activity, error) {
			return &api.Activity{
				ID:    "act789",
				Type:  "environment.redeploy",
				State: "pending",
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &RedeployCmd{EnvironmentID: "staging"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify staging was used
	if mockClient.Calls[0].Args[1] != "staging" {
		t.Errorf("expected environment ID 'staging', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestPushCmd_NoProject(t *testing.T) {
	cli := &CLI{Output: "json"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project
		},
	}

	cmd := &PushCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestPushCmd_NoRepositoryURL(t *testing.T) {
	mockClient := &api.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID string) (*api.Project, error) {
			return &api.Project{
				ID:    projectID,
				Title: "Test Project",
				// Repository.URL is empty
			}, nil
		},
	}

	cli := &CLI{Output: "json"}
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

	cmd := &PushCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when repository URL is empty")
	}
}

func TestPushCmd_GetProjectError(t *testing.T) {
	mockClient := &api.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID string) (*api.Project, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Project not found",
			}
		},
	}

	cli := &CLI{Output: "json"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		apiClientFactory: func(ctx context.Context) (api.API, error) {
			return mockClient, nil
		},
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "nonexistent"
			}
			return ""
		},
	}

	cmd := &PushCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when project not found")
	}
}
