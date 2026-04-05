package cmd

import (
	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
)

// IntegrationListCmd lists integrations for a project.
type IntegrationListCmd struct {
	Type string `help:"Filter by integration type" short:"t"`
	Full bool   `help:"Include all fields" short:"f"`
}

// Run executes the integration:list command.
func (c *IntegrationListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	opts := api.ListIntegrationsOptions{
		Type: c.Type,
	}

	integrations, err := client.ListIntegrations(ctx, projectID, opts)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	if c.Full {
		return ctx.Output(integrations)
	}

	// Return lean summaries
	summaries := make([]api.IntegrationSummary, len(integrations))
	for i, integration := range integrations {
		summaries[i] = integration.ToSummary()
	}

	return ctx.Output(summaries)
}

// IntegrationGetCmd gets details for a specific integration.
type IntegrationGetCmd struct {
	IntegrationID string `arg:"" help:"Integration ID"`
}

// Run executes the integration:get command.
func (c *IntegrationGetCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	integration, err := client.GetIntegration(ctx, projectID, c.IntegrationID)
	if err != nil {
		return handleAPIError(err, "integration", c.IntegrationID)
	}

	return ctx.Output(integration)
}
