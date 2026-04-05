package cmd

import (
	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
)

// CertificateListCmd lists SSL certificates for a project.
type CertificateListCmd struct {
	Full bool `help:"Include all fields" short:"f"`
}

// Run executes the certificate:list command.
func (c *CertificateListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	certificates, err := client.ListCertificates(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	if c.Full {
		return ctx.Output(certificates)
	}

	// Return lean summaries
	summaries := make([]api.CertificateSummary, len(certificates))
	for i, cert := range certificates {
		summaries[i] = cert.ToSummary()
	}

	return ctx.Output(summaries)
}
