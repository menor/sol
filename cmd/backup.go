package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// BackupListCmd lists backups for an environment.
type BackupListCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Full          bool   `help:"Include all fields" short:"f"`
}

// Run executes the backup:list command.
func (c *BackupListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	backups, err := client.ListBackups(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Full {
		return ctx.Output(backups)
	}

	// Return lean summaries
	summaries := make([]api.BackupSummary, len(backups))
	for i, b := range backups {
		summaries[i] = b.ToSummary()
	}
	return ctx.Output(summaries)
}

// BackupGetCmd gets details of a specific backup.
type BackupGetCmd struct {
	BackupID      string `arg:"" required:"" help:"Backup ID"`
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
}

// Run executes the backup:get command.
func (c *BackupGetCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	backup, err := client.GetBackup(ctx, projectID, envID, c.BackupID)
	if err != nil {
		return handleAPIError(err, "backup", c.BackupID)
	}

	return ctx.Output(backup)
}

// BackupCreateCmd creates a new backup.
type BackupCreateCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Live          bool   `help:"Create live backup (no service pause, may have inconsistencies)" short:"l"`
	Wait          bool   `help:"Wait for backup to complete" short:"w"`
}

// Run executes the backup:create command.
func (c *BackupCreateCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// safe = !live (default is safe backup)
	safe := !c.Live

	activity, err := client.CreateBackup(ctx, projectID, envID, safe)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// BackupRestoreCmd restores a backup.
type BackupRestoreCmd struct {
	BackupID        string `arg:"" required:"" help:"Backup ID to restore"`
	EnvironmentID   string `arg:"" optional:"" help:"Source environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Target          string `help:"Target environment name (defaults to source environment)" short:"t"`
	BranchFrom      string `help:"Parent branch for new environment (required if target is new)" short:"b"`
	RestoreCode     bool   `help:"Restore code from backup (default: true)" default:"true"`
	Wait            bool   `help:"Wait for restore to complete" short:"w"`
}

// Run executes the backup:restore command.
func (c *BackupRestoreCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Default target to source environment
	target := c.Target
	if target == "" {
		target = envID
	}

	// Default branch_from to target (restore in place)
	branchFrom := c.BranchFrom
	if branchFrom == "" {
		branchFrom = target
	}

	input := api.RestoreBackupInput{
		EnvironmentName: target,
		BranchFrom:      branchFrom,
		RestoreCode:     c.RestoreCode,
	}

	activity, err := client.RestoreBackup(ctx, projectID, envID, c.BackupID, input)
	if err != nil {
		return handleAPIError(err, "backup", c.BackupID)
	}

	if c.Wait {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// BackupDeleteCmd deletes a backup.
type BackupDeleteCmd struct {
	BackupID      string `arg:"" required:"" help:"Backup ID to delete"`
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
}

// Run executes the backup:delete command.
func (c *BackupDeleteCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	if err := client.DeleteBackup(ctx, projectID, envID, c.BackupID); err != nil {
		return handleAPIError(err, "backup", c.BackupID)
	}

	return ctx.Output(map[string]any{
		"deleted":   true,
		"backup_id": c.BackupID,
	})
}
