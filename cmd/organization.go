package cmd

import (
	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
)

// OrganizationListCmd lists all organizations for the current user.
type OrganizationListCmd struct {
	Full bool `help:"Include all fields" short:"f"`
}

// Run executes the organization:list command.
func (c *OrganizationListCmd) Run(ctx *Context) error {
	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	orgs, err := client.ListOrganizations(ctx)
	if err != nil {
		return handleAPIError(err, "organizations", "")
	}

	if c.Full {
		return ctx.Output(orgs)
	}

	// Return lean summaries
	summaries := make([]api.OrganizationSummary, len(orgs))
	for i, o := range orgs {
		summaries[i] = o.ToSummary()
	}
	return ctx.Output(summaries)
}

// OrganizationInfoCmd shows details for a specific organization.
type OrganizationInfoCmd struct {
	OrgID string `arg:"" required:"" help:"Organization ID"`
}

// Run executes the organization:info command.
func (c *OrganizationInfoCmd) Run(ctx *Context) error {
	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	org, err := client.GetOrganization(ctx, c.OrgID)
	if err != nil {
		return handleAPIError(err, "organization", c.OrgID)
	}

	return ctx.Output(org)
}
