package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// DomainListCmd lists custom domains for a project.
type DomainListCmd struct {
	Full bool `help:"Include all fields" short:"f"`
}

// Run executes the domain:list command.
func (c *DomainListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	domains, err := client.ListDomains(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	if c.Full {
		return ctx.Output(domains)
	}

	// Return lean summaries
	summaries := make([]api.DomainSummary, len(domains))
	for i, domain := range domains {
		summaries[i] = domain.ToSummary()
	}

	return ctx.Output(summaries)
}
