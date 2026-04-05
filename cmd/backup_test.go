package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/menor/sol/api"
)

func TestBackupListCmd_Success(t *testing.T) {
	now := time.Now()
	mockClient := &api.MockClient{
		ListBackupsFunc: func(ctx context.Context, projectID, envID string) ([]api.Backup, error) {
			return []api.Backup{
				{ID: "backup1", CreatedAt: now, Safe: true, Automated: false, CommitID: "abc123"},
				{ID: "backup2", CreatedAt: now.Add(-time.Hour), Safe: false, Automated: true, CommitID: "def456"},
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

	cmd := &BackupListCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
	}
	if mockClient.Calls[0].Method != "ListBackups" {
		t.Errorf("expected ListBackups call, got %s", mockClient.Calls[0].Method)
	}
}

func TestBackupListCmd_Full(t *testing.T) {
	now := time.Now()
	mockClient := &api.MockClient{
		ListBackupsFunc: func(ctx context.Context, projectID, envID string) ([]api.Backup, error) {
			return []api.Backup{
				{ID: "backup1", CreatedAt: now, Safe: true, Restorable: true},
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

	cmd := &BackupListCmd{Full: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBackupGetCmd_Success(t *testing.T) {
	now := time.Now()
	mockClient := &api.MockClient{
		GetBackupFunc: func(ctx context.Context, projectID, envID, backupID string) (*api.Backup, error) {
			return &api.Backup{
				ID:         backupID,
				CreatedAt:  now,
				Safe:       true,
				Restorable: true,
				CommitID:   "abc123",
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

	cmd := &BackupGetCmd{BackupID: "backup123"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.Calls[0].Args[2] != "backup123" {
		t.Errorf("expected backup ID 'backup123', got %v", mockClient.Calls[0].Args[2])
	}
}

func TestBackupCreateCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		CreateBackupFunc: func(ctx context.Context, projectID, envID string, safe bool) (*api.Activity, error) {
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.backup",
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

	cmd := &BackupCreateCmd{}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default is safe backup (safe=true)
	if mockClient.Calls[0].Args[2] != true {
		t.Errorf("expected safe=true, got %v", mockClient.Calls[0].Args[2])
	}
}

func TestBackupCreateCmd_Live(t *testing.T) {
	mockClient := &api.MockClient{
		CreateBackupFunc: func(ctx context.Context, projectID, envID string, safe bool) (*api.Activity, error) {
			return &api.Activity{ID: "activity123"}, nil
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

	cmd := &BackupCreateCmd{Live: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// --live means safe=false
	if mockClient.Calls[0].Args[2] != false {
		t.Errorf("expected safe=false for --live, got %v", mockClient.Calls[0].Args[2])
	}
}

func TestBackupRestoreCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		RestoreBackupFunc: func(ctx context.Context, projectID, envID, backupID string, input api.RestoreBackupInput) (*api.Activity, error) {
			return &api.Activity{
				ID:    "activity123",
				Type:  "environment.restore",
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

	cmd := &BackupRestoreCmd{BackupID: "backup123", RestoreCode: true}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.Calls[0].Method != "RestoreBackup" {
		t.Errorf("expected RestoreBackup call, got %s", mockClient.Calls[0].Method)
	}
}

func TestBackupDeleteCmd_Success(t *testing.T) {
	mockClient := &api.MockClient{
		DeleteBackupFunc: func(ctx context.Context, projectID, envID, backupID string) error {
			return nil
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

	cmd := &BackupDeleteCmd{BackupID: "backup123"}
	err := cmd.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.Calls[0].Args[2] != "backup123" {
		t.Errorf("expected backup ID 'backup123', got %v", mockClient.Calls[0].Args[2])
	}
}

func TestBackupListCmd_NoProject(t *testing.T) {
	cli := &CLI{}
	ctx := &Context{
		Context: context.Background(),
		CLI:     cli,
		getEnvFunc: func(key string) string {
			return ""
		},
	}

	cmd := &BackupListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestBackupListCmd_NoEnvironment(t *testing.T) {
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

	cmd := &BackupListCmd{}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when no environment specified")
	}
}
