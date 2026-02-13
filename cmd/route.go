package cmd

import (
	"github.com/menor/sol/internal/errors"
)

// RouteListCmd lists routes for an environment.
type RouteListCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Full          bool   `help:"Include all fields" short:"f"`
}

// Run executes the route:list command.
func (c *RouteListCmd) Run(ctx *Context) error {
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
		// Return full deployment routes with all details
		deployment, err := client.GetCurrentDeployment(ctx, projectID, envID)
		if err != nil {
			return handleAPIError(err, "environment", envID)
		}
		return ctx.Output(deployment.Routes)
	}

	// Default: lean summaries
	routes, err := client.ListRoutes(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(routes)
}
