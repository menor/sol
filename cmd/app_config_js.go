//go:build js

package cmd

import "github.com/menor/sol/internal/errors"

// AppConfigValidateCmd is unavailable in the browser runtime —
// config validation requires filesystem access.
type AppConfigValidateCmd struct {
	Path string `arg:"" optional:"" help:"Path to config file or directory (defaults to current directory)"`
}

func (c *AppConfigValidateCmd) Run(ctx *Context) error {
	return errors.NewUnsupportedError("config validation requires filesystem access")
}
