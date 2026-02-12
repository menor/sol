package cmd

import (
	"github.com/menor/sol/internal/errors"
)

// ServiceListCmd lists services for an environment.
type ServiceListCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Full          bool   `help:"Include all fields" short:"f"`
}

// Run executes the service:list command.
func (c *ServiceListCmd) Run(ctx *Context) error {
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
		// Return full deployment with all service details
		deployment, err := client.GetCurrentDeployment(ctx, projectID, envID)
		if err != nil {
			return handleAPIError(err, "environment", envID)
		}
		return ctx.Output(deployment.Services)
	}

	// Default: lean summaries
	services, err := client.ListServices(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(services)
}
