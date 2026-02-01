package cmd

import (
	"github.com/menor/sol/internal/errors"
)

// EnvironmentListCmd lists all environments in a project.
type EnvironmentListCmd struct{}

// Run executes the environment:list command.
func (c *EnvironmentListCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
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

	env, err := client.GetEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(env)
}
