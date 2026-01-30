package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lab.plat.farm/menor/sol/internal/api"
	"lab.plat.farm/menor/sol/internal/cli"
	"lab.plat.farm/menor/sol/internal/errors"
)

func init() {
	rootCmd.AddCommand(projectListCmd)
	rootCmd.AddCommand(projectInfoCmd)
}

var projectListCmd = &cobra.Command{
	Use:   "project:list",
	Short: "List all projects",
	Long: `List all projects accessible to the authenticated user.

Returns a JSON array of projects with their IDs, titles, regions, and status.`,
	Aliases: []string{"projects"},
	RunE:    runProjectList,
}

func runProjectList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	client, err := api.New(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("list projects: %v", err))
	}

	return cfg.Formatter().Write(projects)
}

var projectInfoCmd = &cobra.Command{
	Use:   "project:info [project-id]",
	Short: "Show project details",
	Long: `Show detailed information about a specific project.

If no project ID is provided, uses the current project from the
local git repository or PLATFORM_PROJECT environment variable.`,
	Aliases: []string{"project"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runProjectInfo,
}

func runProjectInfo(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := cli.FromCommand(cmd)
	if err != nil {
		return err
	}

	// Get project ID from args or detect from environment
	var projectID string
	if len(args) > 0 {
		projectID = args[0]
	} else {
		projectID = detectProjectID()
		if projectID == "" {
			return errors.NewValidationError("no project specified").
				WithHint("Provide a project ID or run from within a project directory")
		}
	}

	client, err := api.New(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("project", projectID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("get project: %v", err))
	}

	return cfg.Formatter().Write(project)
}

// detectProjectID attempts to determine the current project from context.
// It checks:
// 1. PLATFORM_PROJECT environment variable
// 2. Git remote URLs (TODO)
// 3. Local .platform/local/project.yaml (TODO)
func detectProjectID() string {
	// For now, only check environment variable
	return getEnv("PLATFORM_PROJECT")
}

// getEnv wraps os.Getenv for testing.
var getEnv = os.Getenv
