//go:build !js

package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/alecthomas/kong"
	"github.com/menor/sol/internal/output"
)

// Execute parses command-line arguments and runs the appropriate command.
// This is the main entry point for the CLI.
func Execute() error {
	// Handle --schema flag early, before Kong validates arguments
	if hasSchemaFlag(os.Args) {
		return handleSchemaRequest(os.Args)
	}

	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("sol"),
		kong.Description("Agent-optimized CLI for Upsun"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	if err != nil {
		return err
	}

	kongCtx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	// Create execution context
	ctx := &Context{
		Context: context.Background(),
		CLI:     &cli,
	}

	return kongCtx.Run(ctx)
}

// hasSchemaFlag checks if --schema is present in args.
func hasSchemaFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--schema" {
			return true
		}
	}
	return false
}

// handleSchemaRequest extracts the command name and outputs its schema.
func handleSchemaRequest(args []string) error {
	// Skip program name (args[0])
	if len(args) > 0 {
		args = args[1:]
	}

	// Find the command name (first arg that doesn't start with -)
	var command string
	var outputFormat string = "json"

	for i, arg := range args {
		if arg == "--output" || arg == "-o" {
			if i+1 < len(args) {
				outputFormat = args[i+1]
			}
		} else if len(arg) > 9 && arg[:9] == "--output=" {
			outputFormat = arg[9:]
		} else if len(arg) > 3 && arg[:3] == "-o=" {
			outputFormat = arg[3:]
		} else if !isFlag(arg) && command == "" {
			command = arg
		}
	}

	if command == "" {
		// No command specified, list all commands
		return listAllSchemas(outputFormat)
	}

	schema := GetCommandSchema(command)
	if schema == nil {
		// Unknown command - return error with available commands
		var available []string
		for name := range commandSchemas {
			available = append(available, name)
		}
		sort.Strings(available)
		return fmt.Errorf("unknown command %q. Available commands: %v", command, available)
	}

	formatter := output.New(outputFormat)
	return formatter.Write(schema)
}

// isFlag returns true if the arg looks like a flag.
func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

// listAllSchemas outputs a summary of all available commands.
func listAllSchemas(outputFormat string) error {
	type commandSummary struct {
		Command     string `json:"command"`
		Description string `json:"description"`
	}

	var summaries []commandSummary
	for name, schema := range commandSchemas {
		summaries = append(summaries, commandSummary{
			Command:     name,
			Description: schema.Description,
		})
	}

	// Sort for deterministic output
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Command < summaries[j].Command
	})

	formatter := output.New(outputFormat)
	return formatter.Write(summaries)
}
