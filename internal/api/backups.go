package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"
)

// Backup represents a backup of an environment.
type Backup struct {
	ID            string     `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Status        string     `json:"status,omitempty"`
	CommitID      string     `json:"commit_id,omitempty"`
	Environment   string     `json:"environment"`
	Safe          bool       `json:"safe"`
	Restorable    bool       `json:"restorable"`
	Automated     bool       `json:"automated"`
	Index         int        `json:"index,omitempty"`
	SizeOfVolumes *int64     `json:"size_of_volumes,omitempty"`
}

// BackupSummary is a lean representation for listing.
type BackupSummary struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Safe      bool      `json:"safe"`
	Automated bool      `json:"automated"`
	CommitID  string    `json:"commit_id,omitempty"`
}

// ToSummary converts a Backup to BackupSummary.
func (b *Backup) ToSummary() BackupSummary {
	return BackupSummary{
		ID:        b.ID,
		CreatedAt: b.CreatedAt,
		Safe:      b.Safe,
		Automated: b.Automated,
		CommitID:  b.CommitID,
	}
}

// CreateBackupInput is the request body for creating a backup.
type CreateBackupInput struct {
	Safe bool `json:"safe"`
}

// RestoreBackupInput is the request body for restoring a backup.
type RestoreBackupInput struct {
	EnvironmentName string `json:"environment_name"`
	BranchFrom      string `json:"branch_from"`
	RestoreCode     bool   `json:"restore_code"`
}

// ListBackups returns all backups for an environment.
func (c *Client) ListBackups(ctx context.Context, projectID, envID string) ([]Backup, error) {
	path := fmt.Sprintf("/projects/%s/environments/%s/backups",
		url.PathEscape(projectID), url.PathEscape(envID))

	var backups []Backup
	if err := c.Get(ctx, path, &backups); err != nil {
		return nil, err
	}

	// Sort by created_at descending (newest first) for deterministic output
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// GetBackup returns a single backup by ID.
func (c *Client) GetBackup(ctx context.Context, projectID, envID, backupID string) (*Backup, error) {
	path := fmt.Sprintf("/projects/%s/environments/%s/backups/%s",
		url.PathEscape(projectID), url.PathEscape(envID), url.PathEscape(backupID))

	var backup Backup
	if err := c.Get(ctx, path, &backup); err != nil {
		return nil, err
	}

	return &backup, nil
}

// CreateBackup creates a new backup for an environment.
// If safe is true, services are paused for a consistent backup.
// Returns the activity tracking the backup operation.
func (c *Client) CreateBackup(ctx context.Context, projectID, envID string, safe bool) (*Activity, error) {
	path := fmt.Sprintf("/projects/%s/environments/%s/backup",
		url.PathEscape(projectID), url.PathEscape(envID))

	input := CreateBackupInput{Safe: safe}

	var activity Activity
	if err := c.Post(ctx, path, input, &activity); err != nil {
		return nil, err
	}

	return &activity, nil
}

// RestoreBackup restores a backup to a target environment.
// Returns the activity tracking the restore operation.
func (c *Client) RestoreBackup(ctx context.Context, projectID, envID, backupID string, input RestoreBackupInput) (*Activity, error) {
	path := fmt.Sprintf("/projects/%s/environments/%s/backups/%s/restore",
		url.PathEscape(projectID), url.PathEscape(envID), url.PathEscape(backupID))

	var activity Activity
	if err := c.Post(ctx, path, input, &activity); err != nil {
		return nil, err
	}

	return &activity, nil
}

// DeleteBackup deletes a backup.
func (c *Client) DeleteBackup(ctx context.Context, projectID, envID, backupID string) error {
	path := fmt.Sprintf("/projects/%s/environments/%s/backups/%s",
		url.PathEscape(projectID), url.PathEscape(envID), url.PathEscape(backupID))

	return c.Delete(ctx, path)
}
