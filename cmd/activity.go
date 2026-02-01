package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/cli"
	"github.com/menor/sol/internal/errors"
)

var (
	activityLimit int
	activityType  string
	activityState string
)

func init() {
	rootCmd.AddCommand(activityListCmd)
	rootCmd.AddCommand(activityLogCmd)

	// Flags for activity:list
	activityListCmd.Flags().IntVar(&activityLimit, "limit", 10, "Maximum number of activities to return")
	activityListCmd.Flags().StringVar(&activityType, "type", "", "Filter by activity type")
	activityListCmd.Flags().StringVar(&activityState, "state", "", "Filter by state (pending, in_progress, complete)")
}

var activityListCmd = &cobra.Command{
	Use:   "activity:list",
	Short: "List project activities",
	Long: `List activities for a project.

Activities include deployments, backups, variable changes, and other operations.
By default returns the 10 most recent activities.`,
	Aliases: []string{"activities", "act:list"},
	RunE:    runActivityList,
}

func runActivityList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	projectID := cfg.ProjectID
	if projectID == "" {
		projectID = detectProjectID()
		if projectID == "" {
			return errors.NewValidationError("no project specified").
				WithHint("Use --project or run from within a project directory")
		}
	}

	client, err := newAPIClient(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	opts := &api.ListActivitiesOptions{
		Limit:       activityLimit,
		Type:        activityType,
		State:       activityState,
		Environment: cfg.Environment, // Filter by environment if specified via --environment flag
	}

	activities, err := client.ListActivities(ctx, projectID, opts)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("project", projectID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("list activities: %v", err))
	}

	return cfg.Formatter().Write(activities)
}

var activityLogCmd = &cobra.Command{
	Use:   "activity:log <activity-id>",
	Short: "Show activity log output",
	Long: `Show the log output for a specific activity.

The activity ID can be found using the activity:list command.`,
	Aliases: []string{"act:log"},
	Args:    cobra.ExactArgs(1),
	RunE:    runActivityLog,
}

func runActivityLog(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	projectID := cfg.ProjectID
	if projectID == "" {
		projectID = detectProjectID()
		if projectID == "" {
			return errors.NewValidationError("no project specified").
				WithHint("Use --project or run from within a project directory")
		}
	}

	activityID := args[0]

	client, err := newAPIClient(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	log, err := client.GetActivityLog(ctx, projectID, activityID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("activity", activityID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("get activity log: %v", err))
	}

	// For text output, print the log directly
	// For JSON output, wrap it in a struct
	if cfg.Output == "json" {
		result := struct {
			ActivityID string `json:"activity_id"`
			Log        string `json:"log"`
		}{
			ActivityID: activityID,
			Log:        log,
		}
		return cfg.Formatter().Write(result)
	}

	fmt.Println(log)
	return nil
}
