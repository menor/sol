package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/internal/api"
)

func TestRunActivityList_Success(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
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
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"activity:list", "--output", "json"})
	err := rootCmd.Execute()
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

func TestRunActivityLog_Success(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client
	mockClient := &api.MockClient{
		GetActivityLogFunc: func(ctx context.Context, projectID, activityID string) (string, error) {
			return "Building application...\nDeploying...\nDone.", nil
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"activity:log", "act123", "--output", "json"})
	err := rootCmd.Execute()
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

func TestRunActivityLog_NotFound(t *testing.T) {
	// Save and restore the original factory
	originalFactory := newAPIClient
	originalGetEnv := getEnv
	defer func() {
		newAPIClient = originalFactory
		getEnv = originalGetEnv
	}()

	// Mock environment
	getEnv = func(key string) string {
		if key == "PLATFORM_PROJECT" {
			return "proj123"
		}
		return ""
	}

	// Set up mock client that returns 404
	mockClient := &api.MockClient{
		GetActivityLogFunc: func(ctx context.Context, projectID, activityID string) (string, error) {
			return "", &api.APIError{
				StatusCode: 404,
				Message:    "Activity not found",
			}
		},
	}
	newAPIClient = func(ctx context.Context) (api.API, error) {
		return mockClient, nil
	}

	// Execute command
	rootCmd.SetArgs([]string{"activity:log", "nonexistent", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent activity")
	}
}

func TestRunActivityList_NoProjectSpecified(t *testing.T) {
	// Save and restore the original getEnv
	originalGetEnv := getEnv
	defer func() { getEnv = originalGetEnv }()

	// Mock environment to return nothing
	getEnv = func(key string) string {
		return ""
	}

	// Execute command without project
	rootCmd.SetArgs([]string{"activity:list", "--output", "json"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}
