package cmd

import (
	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
)

// VariableListCmd lists variables.
type VariableListCmd struct {
	Level string `help:"Variable level: project or environment (auto-detected if --environment is set)"`
}

// Run executes the variable:list command.
func (c *VariableListCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Determine level: explicit flag > environment flag presence > default to project
	level := c.Level
	if level == "" {
		if ctx.CLI.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	var variables []api.Variable
	if level == "environment" {
		envID := ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified for environment-level variables").
				WithHint("Use --environment or --level=project for project variables")
		}
		variables, err = client.ListEnvironmentVariables(ctx, projectID, envID)
	} else {
		variables, err = client.ListProjectVariables(ctx, projectID)
	}

	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	return ctx.Output(variables)
}

// VariableGetCmd gets a variable.
type VariableGetCmd struct {
	Name  string `arg:"" required:"" help:"Variable name"`
	Level string `help:"Variable level: project or environment"`
}

// Run executes the variable:get command.
func (c *VariableGetCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	level := c.Level
	if level == "" {
		if ctx.CLI.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	var variable *api.Variable
	if level == "environment" {
		envID := ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified for environment-level variable").
				WithHint("Use --environment or --level=project for project variables")
		}
		variable, err = client.GetEnvironmentVariable(ctx, projectID, envID, c.Name)
	} else {
		variable, err = client.GetProjectVariable(ctx, projectID, c.Name)
	}

	if err != nil {
		return handleAPIError(err, "variable", c.Name)
	}

	return ctx.Output(variable)
}

// VariableSetCmd sets a variable.
type VariableSetCmd struct {
	Name           string `arg:"" required:"" help:"Variable name"`
	Value          string `arg:"" required:"" help:"Variable value"`
	Level          string `help:"Variable level: project or environment"`
	Sensitive      bool   `help:"Mark variable as sensitive (value will be hidden)"`
	VisibleBuild   bool   `help:"Make variable available during build" default:"true" name:"visible-build"`
	VisibleRuntime bool   `help:"Make variable available at runtime" default:"true" name:"visible-runtime"`
}

// Run executes the variable:set command.
func (c *VariableSetCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	level := c.Level
	if level == "" {
		if ctx.CLI.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	input := &api.VariableInput{
		Name:           c.Name,
		Value:          c.Value,
		IsSensitive:    c.Sensitive,
		VisibleBuild:   c.VisibleBuild,
		VisibleRuntime: c.VisibleRuntime,
	}

	var variable *api.Variable
	if level == "environment" {
		envID := ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified for environment-level variable").
				WithHint("Use --environment or --level=project for project variables")
		}
		variable, err = client.SetEnvironmentVariable(ctx, projectID, envID, input)
	} else {
		variable, err = client.SetProjectVariable(ctx, projectID, input)
	}

	if err != nil {
		return handleAPIError(err, "variable", c.Name)
	}

	return ctx.Output(variable)
}

// VariableDeleteCmd deletes a variable.
type VariableDeleteCmd struct {
	Name  string `arg:"" required:"" help:"Variable name"`
	Level string `help:"Variable level: project or environment"`
}

// Run executes the variable:delete command.
func (c *VariableDeleteCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	level := c.Level
	if level == "" {
		if ctx.CLI.Environment != "" {
			level = "environment"
		} else {
			level = "project"
		}
	}

	if level == "environment" {
		envID := ctx.EnvironmentID()
		if envID == "" {
			return errors.NewValidationError("no environment specified for environment-level variable").
				WithHint("Use --environment or --level=project for project variables")
		}
		err = client.DeleteEnvironmentVariable(ctx, projectID, envID, c.Name)
	} else {
		err = client.DeleteProjectVariable(ctx, projectID, c.Name)
	}

	if err != nil {
		return handleAPIError(err, "variable", c.Name)
	}

	result := struct {
		Deleted string `json:"deleted"`
		Level   string `json:"level"`
	}{
		Deleted: c.Name,
		Level:   level,
	}
	return ctx.Output(result)
}
