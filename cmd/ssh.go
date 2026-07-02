package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/menor/sol/internal/errors"
)

// validAppName matches valid Upsun app names: alphanumeric, underscore, hyphen.
// Must start with alphanumeric character.
var validAppName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// SSHCmd opens an SSH connection to an environment.
type SSHCmd struct {
	App     string   `help:"Application name (for multi-app projects)" short:"A"`
	Command []string `arg:"" optional:"" passthrough:"" help:"Command to run on remote (after --)"`
}

// Run executes the ssh command.
func (c *SSHCmd) Run(ctx *Context) error {
	projectID := ctx.ProjectID()
	if projectID == "" {
		return errors.NewNoProjectError()
	}

	envID := ctx.EnvironmentID()
	if envID == "" {
		return errors.NewNoEnvironmentError()
	}

	// Validate app name to prevent SSH argument injection
	if c.App != "" && !validAppName.MatchString(c.App) {
		return errors.NewValidationError("invalid app name").
			WithDetail("app", c.App).
			WithHint("App names must start with a letter or digit and contain only letters, digits, underscores, or hyphens")
	}

	client, err := ctx.APIClient()
	if err != nil {
		return errors.NewAuthError("failed to create API client").WithDetail("cause", err.Error())
	}

	// Get environment to find SSH URL
	env, err := client.GetEnvironment(ctx, projectID, envID)
	if err != nil {
		return handleAPIError(err, "environment", envID)
	}

	// Get SSH URL from _links
	sshURL, ok := env.Links.GetHREF("ssh")
	if !ok {
		return errors.NewValidationError("environment has no SSH access").
			WithHint("The environment may be inactive or SSH access may be disabled")
	}

	// Modify SSH URL for specific app if requested
	if c.App != "" {
		sshURL = modifySSHURLForApp(sshURL, c.App)
	}

	// Parse and execute SSH
	sshArgs := parseSSHURL(sshURL)
	if len(c.Command) > 0 {
		sshArgs = append(sshArgs, c.Command...)
	}

	ctx.Log("Connecting to %s...", envID)

	return execSSH(ctx, sshArgs)
}

// parseSSHURL parses an SSH URL into arguments for the ssh command.
// SSH URLs are in format: ssh://user@host:port or user@host
func parseSSHURL(sshURL string) []string {
	// Remove ssh:// prefix if present
	sshURL = strings.TrimPrefix(sshURL, "ssh://")

	var args []string

	// Check for port in URL (user@host:port)
	if idx := strings.LastIndex(sshURL, ":"); idx > strings.LastIndex(sshURL, "@") {
		host := sshURL[:idx]
		port := sshURL[idx+1:]
		args = append(args, "-p", port, host)
	} else {
		args = append(args, sshURL)
	}

	return args
}

// modifySSHURLForApp modifies the SSH URL to target a specific app.
// For multi-app projects, the app name is appended: user--app@host
func modifySSHURLForApp(sshURL, appName string) string {
	// Remove ssh:// prefix for manipulation
	sshURL = strings.TrimPrefix(sshURL, "ssh://")

	// Split on @ separator
	user, host, found := strings.Cut(sshURL, "@")
	if !found {
		return "ssh://" + sshURL
	}

	// Add app suffix to user
	return fmt.Sprintf("ssh://%s--%s@%s", user, appName, host)
}

// execSSH executes the ssh command with the given arguments.
// The process runs as a child and inherits stdin/stdout/stderr.
func execSSH(ctx context.Context, args []string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return errors.NewInternalError("ssh command not found").
			WithHint("Ensure OpenSSH is installed and in your PATH")
	}

	cmd := exec.CommandContext(ctx, sshPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
