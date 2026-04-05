package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/api"
)

func TestProjectListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListProjectsFunc: func(ctx context.Context) ([]api.ProjectRef, error) {
			return []api.ProjectRef{
				{ID: "proj1", Title: "Project One", Region: "us-1", OrganizationID: "org1"},
				{ID: "proj2", Title: "Project Two", Region: "eu-2", OrganizationID: "org2"},
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
	}

	cmd := &ProjectListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListProjects was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListProjects" {
		t.Errorf("expected ListProjects call, got %s", mockClient.Calls[0].Method)
	}
}

func TestProjectListCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		ListProjectsFunc: func(ctx context.Context) ([]api.ProjectRef, error) {
			return []api.ProjectRef{
				{ID: "proj1", Title: "Project One", Region: "us-1", OrganizationID: "org1", Status: "active"},
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
	}

	// Test with --full flag
	cmd := &ProjectListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListProjects was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
}

func TestProjectInfoCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID string) (*api.Project, error) {
			return &api.Project{
				ID:        projectID,
				Title:     "Test Project",
				Region:    "us-3",
				CreatedAt: time.Now(),
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
	}

	// Test with explicit project ID argument
	cmd := &ProjectInfoCmd{ProjectID: "proj123"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetProject was called with correct ID
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetProject" {
		t.Errorf("expected GetProject call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[0] != "proj123" {
		t.Errorf("expected project ID 'proj123', got %v", mockClient.Calls[0].Args[0])
	}
}

func TestProjectInfoCmd_NotFound(t *testing.T) {
	mockClient := &api.MockClient{
		GetProjectFunc: func(ctx context.Context, projectID string) (*api.Project, error) {
			return nil, &api.APIError{
				StatusCode: 404,
				Message:    "Project not found",
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
	}

	cmd := &ProjectInfoCmd{ProjectID: "nonexistent"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
}

func TestProjectInfoCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project in environment
		},
	}

	cmd := &ProjectInfoCmd{} // No project ID
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestProjectID_FromEnvironment(t *testing.T) {
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

	got := ctx.ProjectID()
	if got != "proj123" {
		t.Errorf("ProjectID() = %q, want %q", got, "proj123")
	}
}

func TestProjectID_FromFlag(t *testing.T) {
	cli := &CLI{Project: "flag-project"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "env-project" // Should be overridden by flag
		},
	}

	got := ctx.ProjectID()
	if got != "flag-project" {
		t.Errorf("ProjectID() = %q, want %q", got, "flag-project")
	}
}
