package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// EnvironmentURLCmd shows URLs for an environment.
type EnvironmentURLCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Primary       bool   `help:"Show only the primary URL" short:"1"`
}

// Run executes the environment:url command.
func (c *EnvironmentURLCmd) Run(ctx *Context) error {
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

	routes, err := client.GetRoutes(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Primary {
		// Return only the primary URL
		for _, r := range routes {
			if r.Primary {
				return ctx.Output(r)
			}
		}
		// No primary found, return first upstream route
		for _, r := range routes {
			if r.Type == "upstream" {
				return ctx.Output(r)
			}
		}
		// Return first route if no primary or upstream
		if len(routes) > 0 {
			return ctx.Output(routes[0])
		}
		return ctx.Output(nil)
	}

	return ctx.Output(routes)
}

// EnvironmentRelationshipsCmd shows relationships between apps and services.
type EnvironmentRelationshipsCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	App           string `help:"Filter by app name" short:"A"`
}

// Run executes the environment:relationships command.
func (c *EnvironmentRelationshipsCmd) Run(ctx *Context) error {
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

	relationships, err := client.GetRelationships(ctx, projectID, envID, c.App)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	// Handle empty results consistently
	if relationships == nil {
		relationships = []api.Relationship{}
	}

	return ctx.Output(relationships)
}
