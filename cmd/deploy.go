package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// EnvironmentBranchCmd creates a new environment by branching from an existing one.
type EnvironmentBranchCmd struct {
	Name   string `arg:"" required:"" help:"Name for the new branch"`
	Parent string `help:"Parent environment to branch from" default:"main"`
	Title  string `help:"Title for the new environment"`
	Wait   bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:branch command.
func (c *EnvironmentBranchCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	input := &api.BranchInput{
		Name:  c.Name,
		Title: c.Title,
	}

	activity, err := client.BranchEnvironment(ctx, projectID, c.Parent, input)
	if err != nil {
		return handleAPIError(err, "environment", c.Parent)
	}

	// Wait for completion if requested
	if c.Wait && activity != nil {
		activity, err = waitForActivity(ctx, client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentActivateCmd activates an inactive environment.
type EnvironmentActivateCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:activate command.
func (c *EnvironmentActivateCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	envID := c.EnvironmentID
	if envID == "" {
		envID = ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified").
				WithHint("Provide an environment ID or use --environment flag")
		}
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	activity, err := client.ActivateEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	// Wait for completion if requested
	if c.Wait && activity != nil {
		activity, err = waitForActivity(ctx, client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentDeactivateCmd deactivates an active environment.
type EnvironmentDeactivateCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:deactivate command.
func (c *EnvironmentDeactivateCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	envID := c.EnvironmentID
	if envID == "" {
		envID = ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified").
				WithHint("Provide an environment ID or use --environment flag")
		}
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	activity, err := client.DeactivateEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	// Wait for completion if requested
	if c.Wait && activity != nil {
		activity, err = waitForActivity(ctx, client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentDeleteCmd deletes an environment.
// The environment must be deactivated before deletion.
type EnvironmentDeleteCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Yes           bool   `help:"Skip confirmation" short:"y"`
}

// Run executes the environment:delete command.
func (c *EnvironmentDeleteCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	envID := c.EnvironmentID
	if envID == "" {
		envID = ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified").
				WithHint("Provide an environment ID or use --environment flag")
		}
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	if err := client.DeleteEnvironment(ctx, projectID, envID); err != nil {
		return handleAPIError(err, "environment", envID)
	}

	// Return a simple success response
	return ctx.Output(map[string]string{
		"status":      "deleted",
		"environment": envID,
		"project":     projectID,
	})
}

// RedeployCmd triggers a redeployment of an environment.
// This reuses the existing build and only runs the post_deploy hook.
type RedeployCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the redeploy command.
func (c *RedeployCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	envID := c.EnvironmentID
	if envID == "" {
		envID = ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified").
				WithHint("Provide an environment ID or use --environment flag")
		}
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	activity, err := client.RedeployEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	// Wait for completion if requested
	if c.Wait && activity != nil {
		activity, err = waitForActivity(ctx, client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// PushCmd pushes code to the Upsun Git remote to trigger a deployment.
type PushCmd struct {
	Target string `help:"Target branch to push to (defaults to current branch)" short:"t"`
	Force  bool   `help:"Force push (use with caution)" short:"f"`
}

// Run executes the push command.
func (c *PushCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Get project to find Git URL
	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	if project.Repository.URL == "" {
		return errors.NewValidationError("project has no repository URL").
			WithHint("The project may not be properly configured")
	}

	// Build git push command
	args := []string{"push"}

	if c.Force {
		args = append(args, "--force")
	}

	args = append(args, project.Repository.URL)

	// Determine what to push
	if c.Target != "" {
		// Push current HEAD to target branch
		args = append(args, fmt.Sprintf("HEAD:%s", c.Target))
	} else {
		// Push current branch
		args = append(args, "HEAD")
	}

	ctx.Log("Pushing to %s...", project.Repository.URL)

	// Execute git push
	if err := execGit(ctx, args); err != nil {
		return errors.NewInternalError("git push failed").
			WithDetail("cause", err.Error()).
			WithHint("Check that you have push access and the remote is reachable")
	}

	// Return success response
	return ctx.Output(map[string]any{
		"status":     "pushed",
		"project":    projectID,
		"repository": project.Repository.URL,
		"target":     c.Target,
		"force":      c.Force,
	})
}

// execGit executes a git command with the given arguments.
func execGit(ctx context.Context, args []string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git command not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, gitPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// waitForActivity polls the activity status until it completes.
// It returns the final activity state.
func waitForActivity(ctx *Context, client api.API, projectID, activityID string) (*api.Activity, error) {
	const (
		pollInterval = 2 * time.Second
		maxWait      = 30 * time.Minute
	)

	deadline := time.Now().Add(maxWait)

	for {
		activity, err := client.GetActivity(ctx, projectID, activityID)
		if err != nil {
			return nil, err
		}

		// Check if activity is complete
		switch activity.State {
		case "complete":
			return activity, nil
		case "cancelled":
			return activity, errors.NewValidationError("activity was cancelled").
				WithDetail("activity_id", activityID)
		}

		// Check timeout
		if time.Now().After(deadline) {
			return activity, errors.NewValidationError("timeout waiting for activity").
				WithDetail("activity_id", activityID).
				WithDetail("state", activity.State)
		}

		ctx.Log("Activity %s: %s...", activityID, activity.State)

		// Wait before next poll
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}
