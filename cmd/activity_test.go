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

func TestActivityListCmd_Full(t *testing.T) {
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

	// Test with --full flag
	cmd := &ActivityListCmd{Limit: 10, Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ListActivities was called
	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
}

func TestActivityLogCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityLogFunc: func(ctx context.Context, projectID, activityID string) (string, error) {
			return "Building application...\nDeploying...\nDone.", nil
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

	cmd := &ActivityLogCmd{ActivityID: "nonexistent"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent activity")
	}
}

func TestActivityListCmd_NoProjectSpecified(t *testing.T) {
	cli := &CLI{}
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

func TestActivityListCmd_ExcludeTypeFilter(t *testing.T) {
	mockClient := &api.MockClient{
		ListActivitiesFunc: func(ctx context.Context, projectID string, opts *api.ListActivitiesOptions) ([]api.Activity, error) {
			return []api.Activity{
				{ID: "act1", Type: "environment.push", State: "complete", CreatedAt: time.Now()},
				{ID: "act2", Type: "environment.backup", State: "complete", CreatedAt: time.Now()},
				{ID: "act3", Type: "environment.push", State: "complete", CreatedAt: time.Now()},
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

	// Exclude backup activities
	cmd := &ActivityListCmd{Limit: 10, ExcludeType: "environment.backup"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivityListCmd_IncompleteFilter(t *testing.T) {
	mockClient := &api.MockClient{
		ListActivitiesFunc: func(ctx context.Context, projectID string, opts *api.ListActivitiesOptions) ([]api.Activity, error) {
			return []api.Activity{
				{ID: "act1", Type: "environment.push", State: "complete", CreatedAt: time.Now()},
				{ID: "act2", Type: "environment.push", State: "in_progress", CreatedAt: time.Now()},
				{ID: "act3", Type: "environment.backup", State: "pending", CreatedAt: time.Now()},
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

	// Only incomplete activities
	cmd := &ActivityListCmd{Limit: 10, Incomplete: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivityListCmd_AllFlag(t *testing.T) {
	mockClient := &api.MockClient{
		ListActivitiesFunc: func(ctx context.Context, projectID string, opts *api.ListActivitiesOptions) ([]api.Activity, error) {
			// Verify high limit is passed when --all is used
			if opts.Limit != 1000 {
				t.Errorf("expected limit 1000 with --all, got %d", opts.Limit)
			}
			return []api.Activity{
				{ID: "act1", Type: "environment.push", State: "complete", CreatedAt: time.Now()},
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

	cmd := &ActivityListCmd{Limit: 10, All: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivityListCmd_StartDateFilter(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	mockClient := &api.MockClient{
		ListActivitiesFunc: func(ctx context.Context, projectID string, opts *api.ListActivitiesOptions) ([]api.Activity, error) {
			return []api.Activity{
				{ID: "act1", Type: "environment.push", State: "complete", CreatedAt: now},
				{ID: "act2", Type: "environment.push", State: "complete", CreatedAt: yesterday},
				{ID: "act3", Type: "environment.push", State: "complete", CreatedAt: twoDaysAgo},
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

	// Filter activities after yesterday (should exclude twoDaysAgo)
	cmd := &ActivityListCmd{Limit: 10, Start: yesterday.Add(-time.Hour).Format(time.RFC3339)}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
