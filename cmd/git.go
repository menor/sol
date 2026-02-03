package cmd

import (
	"os"
	"os/exec"

	"github.com/menor/sol/internal/errors"
)

// PushCmd pushes code to trigger a deployment.
type PushCmd struct {
	Target string `help:"Target branch (defaults to current branch)" short:"t"`
	Force  bool   `help:"Force push" short:"f"`
	Wait   bool   `help:"Wait for the activity to complete" short:"w"`
}

// Run executes the push command.
func (c *PushCmd) Run(ctx *Context) error {
	projectID, err := ctx.RequireProjectID()
	if err != nil {
		return err
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Get project to find repository URL
	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return handleAPIError(err, "project", projectID)
	}

	if project.Repository.URL == "" {
		return errors.NewValidationError("project has no repository URL").
			WithHint("Ensure the project has Git integration enabled")
	}

	// Build git push command
	args := []string{"push", project.Repository.URL}
	if c.Target != "" {
		args = append(args, "HEAD:refs/heads/"+c.Target)
	}
	if c.Force {
		args = append(args, "--force")
	}

	if err := execGit(ctx, args...); err != nil {
		return errors.NewInternalError("git push failed").
			WithDetail("cause", err.Error()).
			WithHint("Ensure you have push access to the repository")
	}

	return ctx.Output(map[string]string{
		"status":  "pushed",
		"project": projectID,
		"target":  c.Target,
	})
}

// execGit runs a git command.
func execGit(ctx *Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
