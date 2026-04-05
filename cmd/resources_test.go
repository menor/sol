package cmd

import (
	"context"
	"testing"

	"github.com/menor/sol/api"
)

func TestResourcesGetCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetResourcesFunc: func(ctx context.Context, projectID, envID string) (*api.ResourceAllocation, error) {
			return &api.ResourceAllocation{
				Webapps: map[string]api.ServiceResources{
					"myapp": {Size: "M", Disk: 1024, InstanceCount: 1},
				},
				Services: map[string]api.ServiceResources{
					"database": {Size: "L", Disk: 2048},
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
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	cmd := &ResourcesGetCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetResources" {
		t.Errorf("expected GetResources call, got %s", mockClient.Calls[0].Method)
	}
}

func TestResourcesGetCmd_Full(t *testing.T) {
	mockClient := &api.MockClient{
		GetResourcesFunc: func(ctx context.Context, projectID, envID string) (*api.ResourceAllocation, error) {
			return &api.ResourceAllocation{
				Webapps: map[string]api.ServiceResources{
					"myapp": {
						Size:          "M",
						Disk:          1024,
						InstanceCount: 1,
						Resources:     &api.ResourceSettings{BaseMemory: 512, MemoryRatio: 1.0},
					},
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
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	cmd := &ResourcesGetCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResourcesSetCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		GetResourcesFunc: func(ctx context.Context, projectID, envID string) (*api.ResourceAllocation, error) {
			return &api.ResourceAllocation{
				Webapps: map[string]api.ServiceResources{
					"myapp": {Size: "M"},
				},
			}, nil
		},
		SetResourcesFunc: func(ctx context.Context, projectID, envID string, input api.SetResourcesInput) (*api.Activity, error) {
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.resources",
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
				return "main"
			}
			return ""
		},
	}

	cmd := &ResourcesSetCmd{Service: "myapp", Size: "L"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should call GetResources first to determine service type, then SetResources
	if len(mockClient.Calls) != 2 {
		t.Errorf("expected 2 calls, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "GetResources" {
		t.Errorf("expected GetResources call first, got %s", mockClient.Calls[0].Method)
	}
	if mockClient.Calls[1].Method != "SetResources" {
		t.Errorf("expected SetResources call second, got %s", mockClient.Calls[1].Method)
	}
}

func TestResourcesSetCmd_ServiceNotFound(t *testing.T) {
	mockClient := &api.MockClient{
		GetResourcesFunc: func(ctx context.Context, projectID, envID string) (*api.ResourceAllocation, error) {
			return &api.ResourceAllocation{
				Webapps: map[string]api.ServiceResources{
					"myapp": {Size: "M"},
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
			if key == "PLATFORM_BRANCH" {
				return "main"
			}
			return ""
		},
	}

	cmd := &ResourcesSetCmd{Service: "nonexistent", Size: "L"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when service not found")
	}
}

func TestResourcesSetCmd_NoSettings(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
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

	// No size, disk, or instances specified
	cmd := &ResourcesSetCmd{Service: "myapp"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no resource settings specified")
	}
}

func TestResourcesGetCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &ResourcesGetCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestResourcesGetCmd_NoEnvironment(t *testing.T) {
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

	cmd := &ResourcesGetCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}
