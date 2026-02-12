package cmd

import (
	"fmt"
	"time"

	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/errors"
)

// ActivityListCmd lists project activities.
type ActivityListCmd struct {
	Limit       int    `help:"Maximum number of activities to return" default:"10"`
	Type        string `help:"Filter by activity type"`
	State       string `help:"Filter by state (pending, in_progress, complete)"`
	Result      string `help:"Filter by result (success, failure)"`
	ExcludeType string `help:"Exclude activities of this type" name:"exclude-type"`
	Start       string `help:"Only activities after this date (ISO 8601 format)"`
	Incomplete  bool   `help:"Only show incomplete activities (pending or in_progress)"`
	All         bool   `help:"Show all activities (ignore limit)" short:"a"`
	Full        bool   `help:"Include all fields (result, description, timestamps, etc.)" short:"f"`
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

	// Determine limit
	limit := c.Limit
	if c.All {
		limit = 1000 // High limit to get all activities
	}

	// Handle --incomplete flag by setting state filter
	state := c.State
	if c.Incomplete && state == "" {
		// We'll filter client-side for both pending and in_progress
		state = ""
	}

	opts := &api.ListActivitiesOptions{
		Limit:       limit,
		Type:        c.Type,
		State:       state,
		Result:      c.Result,
		Environment: ctx.CLI.Environment,
	}

	activities, err := client.ListActivities(ctx, projectID, opts)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	// Apply client-side filters
	activities = c.filterActivities(activities)

	// Default: return lean summary (ID, Type, State, CreatedAt only)
	// --full: return all fields
	if c.Full {
		return ctx.Output(activities)
	}

	// Convert to lean summaries
	summaries := make([]api.ActivitySummary, len(activities))
	for i, a := range activities {
		summaries[i] = a.ToSummary()
	}
	return ctx.Output(summaries)
}

// filterActivities applies client-side filters that the API doesn't support.
func (c *ActivityListCmd) filterActivities(activities []api.Activity) []api.Activity {
	// If no client-side filters, return all
	if c.ExcludeType == "" && c.Start == "" && !c.Incomplete {
		return activities
	}

	// Parse start date if provided
	var startTime time.Time
	if c.Start != "" {
		var err error
		// Try various date formats
		for _, format := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02",
		} {
			startTime, err = time.Parse(format, c.Start)
			if err == nil {
				break
			}
		}
	}

	filtered := make([]api.Activity, 0, len(activities))
	for _, a := range activities {
		// Exclude by type
		if c.ExcludeType != "" && a.Type == c.ExcludeType {
			continue
		}

		// Filter by start date
		if !startTime.IsZero() && a.CreatedAt.Before(startTime) {
			continue
		}

		// Filter incomplete activities
		if c.Incomplete && a.State != "pending" && a.State != "in_progress" {
			continue
		}

		filtered = append(filtered, a)
	}
	return filtered
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
