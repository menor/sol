//go:build !js

package main

import (
	"fmt"
	"os"

	"github.com/menor/sol/cmd"
	"github.com/menor/sol/internal/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		exitCode := errors.ExitUserError

		if cliErr, ok := err.(*errors.CLIError); ok {
			if jsonBytes, jsonErr := cliErr.JSON(); jsonErr == nil {
				fmt.Fprintln(os.Stderr, string(jsonBytes))
			} else {
				fmt.Fprintf(os.Stderr, "error: %s (code: %s)\n", cliErr.Message, cliErr.Code)
			}
			exitCode = cliErr.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}

		os.Exit(exitCode)
	}
}
