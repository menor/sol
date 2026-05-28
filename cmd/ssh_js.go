//go:build js

package cmd

import "github.com/menor/sol/internal/errors"

// SSHCmd is unavailable in the browser runtime — SSH requires local exec.
type SSHCmd struct {
	App     string   `help:"Application name (for multi-app projects)" short:"A"`
	Command []string `arg:"" optional:"" passthrough:"" help:"Command to run on remote (after --)"`
}

func (c *SSHCmd) Run(ctx *Context) error {
	return errors.NewUnsupportedError("ssh not available in browser runtime").
		WithHint("Use the native sol CLI to open an SSH session")
}
