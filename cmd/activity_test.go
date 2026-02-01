package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestActivityListCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		ListActivitiesFunc: func(ctx context.Context, projectID string, opts *api.ListActivitiesOptions) ([]api.Activity, error) {
			return []api.Activity{
				{
					ID:          "act1",
					Type:        "environment.push",
					State:       "complete",
					Result:      "success",
					Description: "Deployed changes",
					Project:     projectID,
					CreatedAt:   time.Now(),
				},
				{
					ID:          "act2",
					Type:        "environment.backup",
					State:       "in_progress",
					Description: "Creating backup",
					Project:     projectID,
					CreatedAt:   time.Now(),
				},
			}, nil
		},
	}

	// Create fresh CLI and context - no globals to reset!
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

	// Run the command directly
	cmd := &ActivityListCmd{Limit: 10}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListActivities was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListActivities" {
		t.Errorf("expected ListActivities call, got %s", mockClient.Calls[0].Method)
	}
}

func TestActivityLogCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityLogFunc: func(ctx context.Context, projectID, activityID string) (string, error) {
			return "Building application...\nDeploying...\nDone.", nil
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

	cmd := &ActivityLogCmd{ActivityID: "act123"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify GetActivityLog was called with correct ID
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetActivityLog" {
		t.Errorf("expected GetActivityLog call, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[0].Args[1] != "act123" {
		t.Errorf("expected activity ID 'act123', got %v", mockClient.Calls[0].Args[1])
	}
}

func TestActivityLogCmd_NotFound(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityLogFunc: func(ctx context.Context, projectID, activityID string) (string, error) {
			return "", &api.APIError{
				StatusCode: 404,
				Message:    "Activity not found",
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

	cmd := &ActivityLogCmd{ActivityID: "nonexistent"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent activity")
	}
}

func TestActivityListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{Output: "json"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return "" // No project in environment
		},
	}

	cmd := &ActivityListCmd{Limit: 10}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
