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
	rootCmd.AddCommand(environmentListCmd)
	rootCmd.AddCommand(environmentInfoCmd)
}

var environmentListCmd = &cobra.Command{
	Use:   "environment:list",
	Short: "List all environments in a project",
	Long: `List all environments (branches) in a project.

Returns a JSON array of environments with their IDs, names, types, and status.`,
	Aliases: []string{"environments", "env:list"},
	RunE:    runEnvironmentList,
}

func runEnvironmentList(cmd *cobra.Command, args []string) error {
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

	client, err := api.New(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	environments, err := client.ListEnvironments(ctx, projectID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("project", projectID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("list environments: %v", err))
	}

	return cfg.Formatter().Write(environments)
}

var environmentInfoCmd = &cobra.Command{
	Use:   "environment:info [environment-id]",
	Short: "Show environment details",
	Long: `Show detailed information about a specific environment.

If no environment ID is provided, uses the current environment from the
local git branch or PLATFORM_BRANCH environment variable.`,
	Aliases: []string{"env:info"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runEnvironmentInfo,
}

func runEnvironmentInfo(cmd *cobra.Command, args []string) error {
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

	// Get environment ID from args or detect from environment
	var envID string
	if len(args) > 0 {
		envID = args[0]
	} else {
		envID = detectEnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified").
				WithHint("Provide an environment ID or run from within an environment")
		}
	}

	client, err := api.New(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	env, err := client.GetEnvironment(ctx, projectID, envID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("environment", envID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("get environment: %v", err))
	}

	return cfg.Formatter().Write(env)
}

// detectEnvironmentID attempts to determine the current environment from context.
// It checks:
// 1. PLATFORM_BRANCH environment variable
// 2. Current git branch (TODO)
func detectEnvironmentID() string {
	if branch := os.Getenv("PLATFORM_BRANCH"); branch != "" {
		return branch
	}
	return ""
}
