package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/menor/sol/internal/api"
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
