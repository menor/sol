package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/menor/sol/api"
	"github.com/menor/sol/internal/errors"
	"github.com/menor/sol/internal/output"
)

// Context holds the execution context for commands.
// It provides access to global flags and shared functionality.
type Context struct {
	context.Context
	CLI *CLI

	// For testing: allows injecting a mock API client
	apiClientFactory func(ctx context.Context) (api.API, error)

	// For testing: override environment variable lookups
	getEnvFunc func(key string) string
}

// ProjectID returns the project ID from flag or environment.
func (c *Context) ProjectID() string {
	if c.CLI.Project != "" {
		return c.CLI.Project
	}
	return c.getEnv("PLATFORM_PROJECT")
}

// EnvironmentID returns the environment ID from flag or environment.
func (c *Context) EnvironmentID() string {
	if c.CLI.Environment != "" {
		return c.CLI.Environment
	}
	return c.getEnv("PLATFORM_BRANCH")
}

// getEnv returns environment variable, using override if set (for testing).
func (c *Context) getEnv(key string) string {
	if c.getEnvFunc != nil {
		return c.getEnvFunc(key)
	}
	return os.Getenv(key)
}

// APIClient creates an API client, using the factory if set (for testing).
func (c *Context) APIClient() (api.API, error) {
	if c.apiClientFactory != nil {
		return c.apiClientFactory(c)
	}
	if c.CLI.Debug {
		return api.New(c, api.WithLogFunc(c.debugLog))
	}
	return api.New(c)
}

// Output writes the value using the configured output format.
func (c *Context) Output(v any) error {
	return output.New(c.CLI.Output).Write(v)
}

// Log writes a message to stderr if not in quiet mode.
func (c *Context) Log(format string, args ...any) {
	if !c.CLI.Quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// debugLog prints debug messages to stderr.
func (c *Context) debugLog(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
}

// RequireProjectID returns the project ID or an error if not specified.
func (c *Context) RequireProjectID() (string, error) {
	projectID := c.ProjectID()
	if projectID == "" {
		return "", errors.NewValidationError("no project specified").
			WithHint("Use --project or run from within a project directory")
	}
	return projectID, nil
}

// ResolveEnvironmentID returns the environment ID from the explicit argument,
// falling back to flag/environment. Returns an error if not specified.
func (c *Context) ResolveEnvironmentID(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	envID := c.EnvironmentID()
	if envID == "" {
		return "", errors.NewValidationError("no environment specified").
			WithHint("Provide an environment ID or use --environment flag")
	}
	return envID, nil
}

// WaitForActivity polls the activity status until it completes.
// It returns the final activity state.
func (c *Context) WaitForActivity(client api.API, projectID, activityID string) (*api.Activity, error) {
	const (
		pollInterval = 2 * time.Second
		maxWait      = 30 * time.Minute
	)

	deadline := time.Now().Add(maxWait)

	for {
		activity, err := client.GetActivity(c, projectID, activityID)
		if err != nil {
			return nil, err
		}

		// Check if activity is complete
		switch activity.State {
		case "complete":
			// Check result for success vs failure
			if activity.Result == "failure" {
				return activity, errors.NewInternalError("activity failed").
					WithDetail("activity_id", activityID).
					WithDetail("result", activity.Result).
					WithHint("Check activity:log for details")
			}
			return activity, nil
		case "cancelled":
			return activity, errors.NewValidationError("activity was cancelled").
				WithDetail("activity_id", activityID)
		case "failure":
			return activity, errors.NewInternalError("activity failed").
				WithDetail("activity_id", activityID).
				WithHint("Check activity:log for details")
		}

		// Check timeout
		if time.Now().After(deadline) {
			return activity, errors.NewValidationError("timeout waiting for activity").
				WithDetail("activity_id", activityID).
				WithDetail("state", activity.State)
		}

		c.Log("Activity %s: %s...", activityID, activity.State)

		// Wait before next poll
		select {
		case <-c.Done():
			return nil, c.Err()
		case <-time.After(pollInterval):
		}
	}
}
