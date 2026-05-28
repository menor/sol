//go:build js

package cmd

import "github.com/menor/sol/internal/errors"

// PushCmd is unavailable in the browser runtime — git push requires local exec.
type PushCmd struct {
	Target string `help:"Target branch (defaults to current branch)" short:"t"`
	Force  bool   `help:"Force push" short:"f"`
}

func (c *PushCmd) Run(ctx *Context) error {
	return errors.NewUnsupportedError("push not available in browser runtime").
		WithHint("Use the native sol CLI to push code")
}
