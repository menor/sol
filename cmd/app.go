package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// AppListCmd lists applications for an environment.
type AppListCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Full          bool   `help:"Include all fields" short:"f"`
}

// Run executes the app:list command.
func (c *AppListCmd) Run(ctx *Context) error {
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

	if c.Full {
		// Return full deployment with all webapp/worker details
		deployment, err := client.GetCurrentDeployment(ctx, projectID, envID)
		if err != nil {
			return handleAPIError(err, "environment", envID)
		}
		// Combine webapps and workers for full output
		// Note: map output order is non-deterministic; use lean output for deterministic results
		result := struct {
			Webapps map[string]api.Webapp `json:"webapps,omitempty"`
			Workers map[string]api.Worker `json:"workers,omitempty"`
		}{
			Webapps: deployment.Webapps,
			Workers: deployment.Workers,
		}
		return ctx.Output(result)
	}

	// Default: lean summaries
	apps, err := client.ListApps(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(apps)
}
