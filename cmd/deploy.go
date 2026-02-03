package cmd

import (
	"github.com/menor/sol/internal/errors"
)

// RedeployCmd triggers a redeployment of an environment.
// This reuses the existing build and only runs the post_deploy hook.
type RedeployCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the redeploy command.
func (c *RedeployCmd) Run(ctx *Context) error {
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

	activity, err := client.RedeployEnvironment(ctx, projectID, envID)
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
