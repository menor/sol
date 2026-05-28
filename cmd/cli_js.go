//go:build js

package cmd

import "github.com/menor/sol/internal/errors"

// Execute is a placeholder for the browser build — the wasm module
// uses cmd/sol-wasm/main.go as its entrypoint, not argv parsing.
func Execute() error {
	return errors.NewUnsupportedError("CLI argument parsing not available in browser runtime")
}
