package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestWaitForActivity_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			return &api.Activity{
				ID:     activityID,
				State:  "complete",
				Result: "success",
			}, nil
		},
	}

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	activity, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.State != "complete" {
		t.Errorf("expected state 'complete', got %s", activity.State)
	}
	if activity.Result != "success" {
		t.Errorf("expected result 'success', got %s", activity.Result)
	}
}

func TestWaitForActivity_CompleteWithFailure(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			return &api.Activity{
				ID:     activityID,
				State:  "complete",
				Result: "failure",
			}, nil
		},
	}

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	activity, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err == nil {
		t.Fatal("expected error for failed activity")
	}
	if activity == nil {
		t.Fatal("expected activity to be returned even on failure")
	}
	if activity.Result != "failure" {
		t.Errorf("expected result 'failure', got %s", activity.Result)
	}
}

func TestWaitForActivity_FailureState(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			return &api.Activity{
				ID:    activityID,
				State: "failure",
			}, nil
		},
	}

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	activity, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err == nil {
		t.Fatal("expected error for failure state")
	}
	if activity == nil {
		t.Fatal("expected activity to be returned even on failure")
	}
}

func TestWaitForActivity_Cancelled(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			return &api.Activity{
				ID:    activityID,
				State: "cancelled",
			}, nil
		},
	}

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	activity, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err == nil {
		t.Fatal("expected error for cancelled activity")
	}
	if activity == nil {
		t.Fatal("expected activity to be returned even when cancelled")
	}
}

func TestWaitForActivity_ContextCancellation(t *testing.T) {
	callCount := 0
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			callCount++
			return &api.Activity{
				ID:    activityID,
				State: "pending",
			}, nil
		},
	}

	bgCtx, cancel := context.WithCancel(context.Background())

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: bgCtx,
		CLI:     cli,
	}

	// Cancel the context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestWaitForActivity_APIError(t *testing.T) {
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			return nil, &api.APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			}
		},
	}

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	_, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err == nil {
		t.Fatal("expected error when API fails")
	}
}

func TestWaitForActivity_PollsUntilComplete(t *testing.T) {
	callCount := 0
	mockClient := &api.MockClient{
		GetActivityFunc: func(ctx context.Context, projectID, activityID string) (*api.Activity, error) {
			callCount++
			// Return pending for first 2 calls, then complete
			if callCount < 3 {
				return &api.Activity{
					ID:    activityID,
					State: "in_progress",
				}, nil
			}
			return &api.Activity{
				ID:     activityID,
				State:  "complete",
				Result: "success",
			}, nil
		},
	}

	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cli := &CLI{Output: "json", Quiet: true}
	ctx := &Context{
		Context: bgCtx,
		CLI:     cli,
	}

	activity, err := ctx.WaitForActivity(mockClient, "proj123", "act123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.State != "complete" {
		t.Errorf("expected state 'complete', got %s", activity.State)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 API calls, got %d", callCount)
	}
}

func TestRequireProjectID_Success(t *testing.T) {
	cli := &CLI{Project: "proj123"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	projectID, err := ctx.RequireProjectID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projectID != "proj123" {
		t.Errorf("expected 'proj123', got %s", projectID)
	}
}

func TestRequireProjectID_FromEnv(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_PROJECT" {
				return "env-proj"
			}
			return ""
		},
	}

	projectID, err := ctx.RequireProjectID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projectID != "env-proj" {
		t.Errorf("expected 'env-proj', got %s", projectID)
	}
}

func TestRequireProjectID_Missing(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	_, err := ctx.RequireProjectID()
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestResolveEnvironmentID_Explicit(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	envID, err := ctx.ResolveEnvironmentID("staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envID != "staging" {
		t.Errorf("expected 'staging', got %s", envID)
	}
}

func TestResolveEnvironmentID_FromFlag(t *testing.T) {
	cli := &CLI{Environment: "production"}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
	}

	envID, err := ctx.ResolveEnvironmentID("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envID != "production" {
		t.Errorf("expected 'production', got %s", envID)
	}
}

func TestResolveEnvironmentID_FromEnv(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	envID, err := ctx.ResolveEnvironmentID("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envID != "main" {
		t.Errorf("expected 'main', got %s", envID)
	}
}

func TestResolveEnvironmentID_Missing(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	_, err := ctx.ResolveEnvironmentID("")
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}
