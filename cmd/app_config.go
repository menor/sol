//go:build !js

package cmd

import (
	"os"

	"github.com/menor/sol/internal/errors"
	"github.com/menor/sol/internal/upsunconfig"
)

// AppConfigValidateCmd validates .upsun/config.yaml files.
type AppConfigValidateCmd struct {
	Path string `arg:"" optional:"" help:"Path to config file or directory (defaults to current directory)"`
}

// Run executes the app:config-validate command.
func (c *AppConfigValidateCmd) Run(ctx *Context) error {
	// Determine the path to validate
	path := c.Path
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return errors.NewInternalError("failed to get current directory").
				WithDetail("cause", err.Error())
		}
	}

	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewNotFoundError("path", path)
		}
		return errors.NewInternalError("failed to access path").
			WithDetail("path", path).
			WithDetail("cause", err.Error())
	}

	var configPath string
	if info.IsDir() {
		// Discover config file in directory
		configPath, err = upsunconfig.DiscoverConfigFile(path)
		if err != nil {
			return errors.NewNotFoundError("config", ".upsun/config.yaml").
				WithHint("Create a .upsun/config.yaml file or specify the path directly")
		}
	} else {
		// Use the provided file path
		configPath = path
	}

	// Validate the config file
	result, err := upsunconfig.ValidateFile(configPath)
	if err != nil {
		return errors.NewInternalError("failed to read config file").
			WithDetail("path", configPath).
			WithDetail("cause", err.Error())
	}

	return ctx.Output(result)
}
