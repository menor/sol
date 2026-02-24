package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// UserListCmd lists users with access to a project.
type UserListCmd struct {
	Full bool `help:"Include all fields" short:"f"`
}

// Run executes the user:list command.
func (c *UserListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	users, err := client.ListProjectUsers(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	if c.Full {
		return ctx.Output(users)
	}

	// Return lean summaries
	summaries := make([]api.ProjectUserAccessSummary, len(users))
	for i, u := range users {
		summaries[i] = u.ToSummary()
	}
	return ctx.Output(summaries)
}
