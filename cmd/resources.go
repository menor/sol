package cmd

import (
	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// ResourcesGetCmd shows current resource allocation for an environment.
type ResourcesGetCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Full          bool   `help:"Include all fields" short:"f"`
}

// Run executes the resources:get command.
func (c *ResourcesGetCmd) Run(ctx *Context) error {
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

	resources, err := client.GetResources(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Full {
		return ctx.Output(resources)
	}

	// Return lean summary (sorting handled in ToSummary)
	return ctx.Output(resources.ToSummary())
}

// ResourcesSetCmd updates resource allocation for an environment.
type ResourcesSetCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Service       string `help:"Service name to update" short:"s" required:""`
	Size          string `help:"Resource size profile (e.g., S, M, L, XL)" long:"size"`
	Disk          int    `help:"Disk size in MB" long:"disk"`
	Instances     int    `help:"Number of instances" long:"instances"`
	Wait          bool   `help:"Wait for deployment to complete" short:"w"`
}

// Run executes the resources:set command.
func (c *ResourcesSetCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	// At least one resource setting must be specified
	if c.Size == "" && c.Disk == 0 && c.Instances == 0 {
		return errors.NewValidationError("at least one of --size, --disk, or --instances must be specified")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Build the input - we put the service in all categories and let the API figure out which one it is
	serviceInput := api.ServiceResourcesInput{
		Size:          c.Size,
		Disk:          c.Disk,
		InstanceCount: c.Instances,
	}

	// First get current resources to determine if it's an app, service, or worker
	current, err := client.GetResources(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	input := api.SetResourcesInput{}
	found := false

	// Check if it's a webapp
	if _, ok := current.Webapps[c.Service]; ok {
		input.Webapps = map[string]api.ServiceResourcesInput{c.Service: serviceInput}
		found = true
	}
	// Check if it's a service
	if _, ok := current.Services[c.Service]; ok {
		input.Services = map[string]api.ServiceResourcesInput{c.Service: serviceInput}
		found = true
	}
	// Check if it's a worker
	if _, ok := current.Workers[c.Service]; ok {
		input.Workers = map[string]api.ServiceResourcesInput{c.Service: serviceInput}
		found = true
	}

	if !found {
		return errors.NewNotFoundError("service", c.Service).
			WithHint("Use 'sol resources:get' to see available services")
	}

	activity, err := client.SetResources(ctx, projectID, envID, input)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}
