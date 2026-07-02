package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// ProjectListCmd lists all projects.
type ProjectListCmd struct {
	Full bool `help:"Include all fields (status, org, subscription, timestamps)" short:"f"`
}

// Run executes the project:list command.
func (c *ProjectListCmd) Run(ctx *Context) error {
	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		return handleAPIError(err, "projects", "")
	}

	// Default: return lean summary (ID, Title, Region only)
	// --full: return all fields
	if c.Full {
		return ctx.Output(projects)
	}

	// Convert to lean summaries
	summaries := make([]api.ProjectSummary, len(projects))
	for i, p := range projects {
		summaries[i] = p.ToSummary()
	}
	return ctx.Output(summaries)
}

// ProjectInfoCmd shows project details.
type ProjectInfoCmd struct {
	ProjectID string `arg:"" optional:"" help:"Project ID (uses --project or PLATFORM_PROJECT if not specified)"`
}

// Run executes the project:info command.
func (c *ProjectInfoCmd) Run(ctx *Context) error {
	projectID := c.ProjectID
	if projectID == "" {
		projectID = ctx.ProjectID()
		if projectID == "" {
			return errors.NewNoProjectError()
		}
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	return ctx.Output(project)
}

// handleAPIError converts API errors to CLI errors.
func handleAPIError(err error, resourceType, resourceID string) error {
	if apiErr, ok := err.(*api.APIError); ok {
		if apiErr.StatusCode == 404 {
			return errors.NewNotFoundError(resourceType, resourceID)
		}
		return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
	}
	return errors.NewInternalError(err.Error())
}
