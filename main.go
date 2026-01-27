package main

import (
	"fmt"
	"os"

	"lab.plat.farm/menor/sol/cmd"
	"lab.plat.farm/menor/sol/internal/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		exitCode := errors.ExitUserError

		if cliErr, ok := err.(*errors.CLIError); ok {
			if jsonBytes, jsonErr := cliErr.JSON(); jsonErr == nil {
				fmt.Fprintln(os.Stderr, string(jsonBytes))
			} else {
				// Fallback if JSON marshaling fails (should never happen)
				fmt.Fprintf(os.Stderr, "error: %s (code: %s)\n", cliErr.Message, cliErr.Code)
			}
			exitCode = cliErr.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}

		os.Exit(exitCode)
	}
}
