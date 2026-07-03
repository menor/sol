package cmd

import (
	"context"
	stderrors "errors"
	"net"
	"net/url"

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

// handleAPIError converts API errors to CLI errors. The api package wraps
// errors with %w ("get current user: ..."), so unwrap with errors.As, never a
// type assertion. Transport-level failures (DNS, refused connection, timeout)
// are operational and retryable — api_unavailable, not a Sol bug. Only errors
// with no better classification fall through to internal.
func handleAPIError(err error, resourceType, resourceID string) error {
	var apiErr *api.APIError
	if stderrors.As(err, &apiErr) {
		if apiErr.StatusCode == 404 {
			return errors.NewNotFoundError(resourceType, resourceID)
		}
		return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
	}

	// http.Client wraps every request failure in *url.Error; also catch bare
	// net.Error and context deadlines from custom transports.
	var urlErr *url.Error
	var netErr net.Error
	if stderrors.As(err, &urlErr) || stderrors.As(err, &netErr) ||
		stderrors.Is(err, context.DeadlineExceeded) {
		return errors.NewAPIUnreachableError(err.Error())
	}

	// Cancellation (e.g. Ctrl-C between requests) is the caller's doing, not
	// a Sol bug. Retryable: the identical call succeeds if not cancelled.
	if stderrors.Is(err, context.Canceled) {
		return errors.NewOperationFailedError("operation cancelled").
			WithRetryable(true)
	}

	return errors.NewInternalError(err.Error())
}
