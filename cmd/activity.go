package cmd

import (
	"fmt"

	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// ActivityListCmd lists project activities.
type ActivityListCmd struct {
	Limit int    `help:"Maximum number of activities to return" default:"10"`
	Type  string `help:"Filter by activity type"`
	State string `help:"Filter by state (pending, in_progress, complete)"`
}

// Run executes the activity:list command.
func (c *ActivityListCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	opts := &api.ListActivitiesOptions{
		Limit:       c.Limit,
		Type:        c.Type,
		State:       c.State,
		Environment: ctx.CLI.Environment,
	}

	activities, err := client.ListActivities(ctx, projectID, opts)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	return ctx.Output(activities)
}

// ActivityLogCmd shows activity log output.
type ActivityLogCmd struct {
	ActivityID string `arg:"" required:"" help:"Activity ID"`
}

// Run executes the activity:log command.
func (c *ActivityLogCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	log, err := client.GetActivityLog(ctx, projectID, c.ActivityID)
	if err != nil {
		return handleAPIError(err, "activity", c.ActivityID)
	}

	// For JSON output, wrap in a struct
	if ctx.CLI.Output == "json" {
		result := struct {
			ActivityID string `json:"activity_id"`
			Log        string `json:"log"`
		}{
			ActivityID: c.ActivityID,
			Log:        log,
		}
		return ctx.Output(result)
	}

	fmt.Println(log)
	return nil
}
