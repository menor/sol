package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/menor/sol/internal/api"
	"github.com/menor/sol/internal/cli"
	"github.com/menor/sol/internal/errors"
)

var (
	varLevel     string // "project" or "environment"
	varSensitive bool
	varBuild     bool
	varRuntime   bool
	varYes       bool // Skip confirmation
)

func init() {
	rootCmd.AddCommand(variableListCmd)
	rootCmd.AddCommand(variableGetCmd)
	rootCmd.AddCommand(variableSetCmd)
	rootCmd.AddCommand(variableDeleteCmd)

	// Common flags
	variableListCmd.Flags().StringVar(&varLevel, "level", "", "Variable level: project or environment (auto-detected if --environment is set)")
	variableGetCmd.Flags().StringVar(&varLevel, "level", "", "Variable level: project or environment")
	variableSetCmd.Flags().StringVar(&varLevel, "level", "", "Variable level: project or environment")
	variableDeleteCmd.Flags().StringVar(&varLevel, "level", "", "Variable level: project or environment")

	// Flags for variable:set
	variableSetCmd.Flags().BoolVar(&varSensitive, "sensitive", false, "Mark variable as sensitive (value will be hidden)")
	variableSetCmd.Flags().BoolVar(&varBuild, "visible-build", true, "Make variable available during build")
	variableSetCmd.Flags().BoolVar(&varRuntime, "visible-runtime", true, "Make variable available at runtime")
	variableSetCmd.Flags().BoolVar(&varYes, "yes", false, "Skip confirmation prompt")

	// Flags for variable:delete
	variableDeleteCmd.Flags().BoolVar(&varYes, "yes", false, "Skip confirmation prompt")
}

var variableListCmd = &cobra.Command{
	Use:   "variable:list",
	Short: "List variables",
	Long: `List variables for a project or environment.

If --environment is specified, lists environment-level variables.
Otherwise, lists project-level variables.`,
	Aliases: []string{"variables", "var:list"},
	RunE:    runVariableList,
}

func runVariableList(cmd *cobra.Command, args []string) error {
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

	// Determine level: explicit flag > environment flag presence > default to project
	level := varLevel
	if level == "" {
		if cfg.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	var variables []api.Variable
	if level == "environment" {
		envID := cfg.Environment
		if envID == "" {
			envID = detectEnvironmentID()
			if envID == "" {
				return errors.NewValidationError("no environment specified for environment-level variables").
					WithHint("Use --environment or --level=project for project variables")
			}
		}
		variables, err = client.ListEnvironmentVariables(ctx, projectID, envID)
	} else {
		variables, err = client.ListProjectVariables(ctx, projectID)
	}

	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("project", projectID)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("list variables: %v", err))
	}

	return cfg.Formatter().Write(variables)
}

var variableGetCmd = &cobra.Command{
	Use:   "variable:get <name>",
	Short: "Get a variable",
	Long: `Get details of a specific variable.

If --environment is specified, gets an environment-level variable.
Otherwise, gets a project-level variable.`,
	Aliases: []string{"var:get"},
	Args:    cobra.ExactArgs(1),
	RunE:    runVariableGet,
}

func runVariableGet(cmd *cobra.Command, args []string) error {
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

	name := args[0]

	client, err := newAPIClient(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	level := varLevel
	if level == "" {
		if cfg.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	var variable *api.Variable
	if level == "environment" {
		envID := cfg.Environment
		if envID == "" {
			envID = detectEnvironmentID()
			if envID == "" {
				return errors.NewValidationError("no environment specified for environment-level variable").
					WithHint("Use --environment or --level=project for project variables")
			}
		}
		variable, err = client.GetEnvironmentVariable(ctx, projectID, envID, name)
	} else {
		variable, err = client.GetProjectVariable(ctx, projectID, name)
	}

	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("variable", name)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("get variable: %v", err))
	}

	return cfg.Formatter().Write(variable)
}

var variableSetCmd = &cobra.Command{
	Use:   "variable:set <name> <value>",
	Short: "Set a variable",
	Long: `Set a project or environment variable.

If --environment is specified, sets an environment-level variable.
Otherwise, sets a project-level variable.

Use --sensitive to mark the variable as sensitive (value will be hidden in output).
Use --yes to skip the confirmation prompt.`,
	Aliases: []string{"var:set"},
	Args:    cobra.ExactArgs(2),
	RunE:    runVariableSet,
}

func runVariableSet(cmd *cobra.Command, args []string) error {
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

	name := args[0]
	value := args[1]

	client, err := newAPIClient(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	level := varLevel
	if level == "" {
		if cfg.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	input := &api.VariableInput{
		Name:           name,
		Value:          value,
		IsSensitive:    varSensitive,
		VisibleBuild:   varBuild,
		VisibleRuntime: varRuntime,
	}

	var variable *api.Variable
	if level == "environment" {
		envID := cfg.Environment
		if envID == "" {
			envID = detectEnvironmentID()
			if envID == "" {
				return errors.NewValidationError("no environment specified for environment-level variable").
					WithHint("Use --environment or --level=project for project variables")
			}
		}
		variable, err = client.SetEnvironmentVariable(ctx, projectID, envID, input)
	} else {
		variable, err = client.SetProjectVariable(ctx, projectID, input)
	}

	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("set variable: %v", err))
	}

	return cfg.Formatter().Write(variable)
}

var variableDeleteCmd = &cobra.Command{
	Use:   "variable:delete <name>",
	Short: "Delete a variable",
	Long: `Delete a project or environment variable.

If --environment is specified, deletes an environment-level variable.
Otherwise, deletes a project-level variable.

Use --yes to skip the confirmation prompt.`,
	Aliases: []string{"var:delete"},
	Args:    cobra.ExactArgs(1),
	RunE:    runVariableDelete,
}

func runVariableDelete(cmd *cobra.Command, args []string) error {
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

	name := args[0]

	client, err := newAPIClient(ctx)
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	level := varLevel
	if level == "" {
		if cfg.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	if level == "environment" {
		envID := cfg.Environment
		if envID == "" {
			envID = detectEnvironmentID()
			if envID == "" {
				return errors.NewValidationError("no environment specified for environment-level variable").
					WithHint("Use --environment or --level=project for project variables")
			}
		}
		err = client.DeleteEnvironmentVariable(ctx, projectID, envID, name)
	} else {
		err = client.DeleteProjectVariable(ctx, projectID, name)
	}

	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 404 {
				return errors.NewNotFoundError("variable", name)
			}
			return errors.NewAPIError(apiErr.Message, apiErr.StatusCode)
		}
		return errors.NewInternalError(fmt.Sprintf("delete variable: %v", err))
	}

	// Return success message
	result := struct {
		Deleted string `json:"deleted"`
		Level   string `json:"level"`
	}{
		Deleted: name,
		Level:   level,
	}
	return cfg.Formatter().Write(result)
}
