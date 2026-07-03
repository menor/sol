package cmd

import (
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/menor/sol/internal/errors"
	"github.com/menor/sol/internal/output"
)

// errorEnvelope is the machine-readable wrapper for a rendered error:
// {"error": {code, message, hint?, retryable, details?}}. See ADR 0002.
type errorEnvelope struct {
	Error *errors.CLIError `json:"error"`
}

// render is the single chokepoint that turns any error from Execute into a
// process exit code: it resolves the output format, routes the stream, and
// maps the exit code. main is thereby reduced to os.Exit(cmd.Execute()).
func render(cli *CLI, err error) int {
	return renderTo(resolveFormat(cli), err, os.Stdout, os.Stderr)
}

// renderTo is the testable core of render, with injectable streams.
func renderTo(format string, err error, stdout, stderr io.Writer) int {
	if err == nil {
		return errors.ExitSuccess
	}

	// Kong wraps command errors in errors.Join, so recover the CLIError with
	// errors.As, never a type assertion. Anything that is not a CLIError is an
	// unexpected failure: surface it as internal (exit 70), never as a bare
	// operational error.
	var cliErr *errors.CLIError
	if !stderrors.As(err, &cliErr) {
		cliErr = errors.NewInternalError(err.Error())
	}

	switch format {
	case "json", "toon":
		// Structured mode (ADR 0002): the envelope owns stdout, nothing on
		// stderr. One stream, one format, fully parseable. If stdout is gone
		// (e.g. a closed pipe), fall back to stderr rather than exiting with
		// the envelope silently lost.
		if werr := output.NewWithWriter(format, stdout).Write(errorEnvelope{Error: cliErr}); werr != nil {
			fmt.Fprintf(stderr, "error: %s\n", cliErr.Message)
		}
	default:
		// Human mode: message + optional hint on stderr, as before.
		fmt.Fprintf(stderr, "error: %s\n", cliErr.Message)
		if cliErr.Hint != "" {
			fmt.Fprintf(stderr, "hint: %s\n", cliErr.Hint)
		}
	}

	return cliErr.ExitCode()
}

// renderParseError handles failures from Kong's argument parsing. No parsed
// CLI exists on this path, so the format comes from scanning os.Args. This is
// the only place exit 80 is assigned (ADR 0001): the same invalid_argument
// code coming from a command handler still exits 1, because there it means a
// bad value at runtime, not a malformed invocation.
func renderParseError(err error) int {
	return renderParseErrorIn(formatFromArgs(os.Args), err)
}

// renderParseErrorIn is renderParseError with an explicit format, for callers
// whose path has a different default (--schema historically defaults to json).
func renderParseErrorIn(format string, err error) int {
	// An already-classified CLIError (e.g. from the --schema path) keeps its
	// own message and hint; only bare Kong errors get the generic wrap.
	var cliErr *errors.CLIError
	if !stderrors.As(err, &cliErr) {
		cliErr = errors.NewValidationError(err.Error()).
			WithHint("Run 'sol --help' to see available commands and flags")
	}
	renderTo(format, cliErr, os.Stdout, os.Stderr)
	return errors.ExitUsage
}

// resolveFormat returns the output format. A parsed CLI is authoritative: its
// Output flag defaults to toon, so it is never empty after a successful parse.
// When the CLI is unavailable — a parse failure or a panic before parsing —
// fall back to scanning os.Args.
func resolveFormat(cli *CLI) string {
	if cli != nil && cli.Output != "" {
		return cli.Output
	}
	return formatFromArgs(os.Args)
}

// formatFromArgs scans raw args for -o/--output, defaulting to the CLI's toon
// default when absent. Used only when the parsed CLI is unreliable.
func formatFromArgs(args []string) string {
	return formatFromArgsOrDefault(args, "toon")
}

// formatFromArgsOrDefault scans raw args for every -o/--output spelling Kong
// accepts (separate value, =value, and the attached short form -ojson). An
// absent or invalid value yields def — this scan renders errors, so it must
// never propagate a format the formatter doesn't support.
func formatFromArgsOrDefault(args []string, def string) string {
	for i, arg := range args {
		var format string
		switch {
		case arg == "--output" || arg == "-o":
			if i+1 < len(args) {
				format = args[i+1]
			}
		case strings.HasPrefix(arg, "--output="):
			format = strings.TrimPrefix(arg, "--output=")
		case strings.HasPrefix(arg, "-o="):
			format = strings.TrimPrefix(arg, "-o=")
		case strings.HasPrefix(arg, "-o"):
			format = strings.TrimPrefix(arg, "-o")
		}
		if format == "json" || format == "toon" {
			return format
		}
	}
	return def
}
