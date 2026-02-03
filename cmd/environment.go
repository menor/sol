package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// EnvironmentListCmd lists all environments in a project.
type EnvironmentListCmd struct{}

// Run executes the environment:list command.
func (c *EnvironmentListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	environments, err := client.ListEnvironments(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	return ctx.Output(environments)
}

// EnvironmentInfoCmd shows environment details.
type EnvironmentInfoCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
}

// Run executes the environment:info command.
func (c *EnvironmentInfoCmd) Run(ctx *Context) error {
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

	env, err := client.GetEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(env)
}

// EnvironmentBranchCmd creates a new environment by branching from an existing one.
type EnvironmentBranchCmd struct {
	Name   string `arg:"" required:"" help:"Name for the new branch"`
	Parent string `help:"Parent environment to branch from" default:"main"`
	Title  string `help:"Title for the new environment"`
	Wait   bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:branch command.
func (c *EnvironmentBranchCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
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

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
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

	activity, err := client.ActivateEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
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

	activity, err := client.DeactivateEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
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
}

// Run executes the environment:delete command.
func (c *EnvironmentDeleteCmd) Run(ctx *Context) error {
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

	if err := client.DeleteEnvironment(ctx, projectID, envID); err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(map[string]string{
		"status":      "deleted",
		"environment": envID,
		"project":     projectID,
	})
}
