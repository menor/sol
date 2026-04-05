package cmd

import (
	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
)

// EnvironmentListCmd lists all environments in a project.
type EnvironmentListCmd struct {
	Full       bool   `help:"Include all fields (type, machine_name, timestamps, etc.)" short:"f"`
	Status     string `help:"Filter by status (active, inactive, dirty)"`
	NoInactive bool   `help:"Exclude inactive environments" name:"no-inactive"`
	Type       string `help:"Filter by type (production, staging, development)"`
}

// Run executes the environment:list command.
func (c *EnvironmentListCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	environments, err := client.ListEnvironments(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	// Apply filters
	environments = c.filterEnvironments(environments)

	// Default: return lean summary (ID, Name, Status, Parent only)
	// --full: return all fields
	if c.Full {
		return ctx.Output(environments)
	}

	// Convert to lean summaries
	summaries := make([]api.EnvironmentSummary, len(environments))
	for i, e := range environments {
		summaries[i] = e.ToSummary()
	}
	return ctx.Output(summaries)
}

// filterEnvironments applies the configured filters to the environment list.
func (c *EnvironmentListCmd) filterEnvironments(envs []api.Environment) []api.Environment {
	// If no filters, return all
	if c.Status == "" && !c.NoInactive && c.Type == "" {
		return envs
	}

	filtered := make([]api.Environment, 0, len(envs))
	for _, e := range envs {
		// Skip inactive if --no-inactive is set
		if c.NoInactive && e.Status == "inactive" {
			continue
		}

		// Filter by status
		if c.Status != "" && e.Status != c.Status {
			continue
		}

		// Filter by type
		if c.Type != "" && e.Type != c.Type {
			continue
		}

		filtered = append(filtered, e)
	}
	return filtered
}

// EnvironmentInfoCmd shows environment details.
type EnvironmentInfoCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
}

// Run executes the environment:info command.
func (c *EnvironmentInfoCmd) Run(ctx *Context) error {
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

	env, err := client.GetEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(env)
}

// EnvironmentBranchCmd creates a new environment by branching from an existing one.
type EnvironmentBranchCmd struct {
	Name   string `arg:"" required:"" help:"Name for the new branch"`
	Parent string `help:"Parent environment to branch from" default:"main"`
	Title  string `help:"Title for the new environment"`
	Wait   bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:branch command.
func (c *EnvironmentBranchCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	input := &api.BranchInput{
		Name:  c.Name,
		Title: c.Title,
	}

	activity, err := client.BranchEnvironment(ctx, projectID, c.Parent, input)
	if err != nil {
		return handleAPIError(err, "environment", c.Parent)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentActivateCmd activates an inactive environment.
type EnvironmentActivateCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:activate command.
func (c *EnvironmentActivateCmd) Run(ctx *Context) error {
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

	activity, err := client.ActivateEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentDeactivateCmd deactivates an active environment.
type EnvironmentDeactivateCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:deactivate command.
func (c *EnvironmentDeactivateCmd) Run(ctx *Context) error {
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

	activity, err := client.DeactivateEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentDeleteCmd deletes an environment.
// The environment must be deactivated before deletion.
type EnvironmentDeleteCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
}

// Run executes the environment:delete command.
func (c *EnvironmentDeleteCmd) Run(ctx *Context) error {
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

	if err := client.DeleteEnvironment(ctx, projectID, envID); err != nil {
		return handleAPIError(err, "environment", envID)
	}

	return ctx.Output(map[string]string{
		"status":      "deleted",
		"environment": envID,
		"project":     projectID,
	})
}

// EnvironmentMergeCmd merges the current environment into its parent.
type EnvironmentMergeCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:merge command.
func (c *EnvironmentMergeCmd) Run(ctx *Context) error {
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

	activity, err := client.MergeEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}

// EnvironmentSyncCmd synchronizes data and/or code from the parent environment.
type EnvironmentSyncCmd struct {
	EnvironmentID string `arg:"" optional:"" help:"Environment ID (uses --environment or PLATFORM_BRANCH if not specified)"`
	Data          bool   `help:"Synchronize data from parent" short:"d"`
	Code          bool   `help:"Synchronize code from parent" short:"c"`
	Resources     bool   `help:"Synchronize resources from parent" short:"r"`
	Wait          bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the environment:sync command.
func (c *EnvironmentSyncCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	envID, err := ctx.ResolveEnvironmentID(c.EnvironmentID)
	if err != nil {
		return err
	}

	// At least one of data, code, or resources must be specified
	if !c.Data && !c.Code && !c.Resources {
		return errors.NewValidationError("at least one of --data, --code, or --resources must be specified")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	input := &api.SyncInput{
		SynchronizeData:      c.Data,
		SynchronizeCode:      c.Code,
		SynchronizeResources: c.Resources,
	}

	activity, err := client.SyncEnvironment(ctx, projectID, envID, input)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	if c.Wait && activity != nil {
		activity, err = ctx.WaitForActivity(client, projectID, activity.ID)
		if err != nil {
			return err
		}
	}

	return ctx.Output(activity)
}
