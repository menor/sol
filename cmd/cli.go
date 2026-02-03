package cmd

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	"github.com/menor/sol/internal/output"
)

// CLI is the main command-line interface struct.
// Global flags are defined here and available to all commands.
type CLI struct {
	// Global flags
	Output      string `help:"Output format" default:"json" enum:"json,toon" short:"o"`
	Project     string `help:"Project ID" short:"p" env:"UPSUN_PROJECT"`
	Environment string `help:"Environment name" short:"e" env:"UPSUN_ENVIRONMENT"`
	Quiet       bool   `help:"Suppress non-essential output" short:"q"`
	NoCache     bool   `help:"Bypass cache" name:"no-cache"`
	Debug       bool   `help:"Show API request/response details"`
	Schema      bool   `help:"Output command schema instead of running"`

	// Commands
	Version         VersionCmd         `cmd:"" name:"version" help:"Print version information"`
	AuthLogin       AuthLoginCmd       `cmd:"" name:"auth:login" help:"Log in to Upsun"`
	AuthLogout      AuthLogoutCmd      `cmd:"" name:"auth:logout" help:"Log out of Upsun"`
	AuthInfo        AuthInfoCmd        `cmd:"" name:"auth:info" help:"Show authentication status"`
	ProjectList     ProjectListCmd     `cmd:"" name:"project:list" aliases:"projects" help:"List all projects"`
	ProjectInfo     ProjectInfoCmd     `cmd:"" name:"project:info" aliases:"project" help:"Show project details"`
	EnvironmentList       EnvironmentListCmd       `cmd:"" name:"environment:list" aliases:"environments,env:list" help:"List environments"`
	EnvironmentInfo       EnvironmentInfoCmd       `cmd:"" name:"environment:info" aliases:"env:info" help:"Show environment details"`
	EnvironmentBranch     EnvironmentBranchCmd     `cmd:"" name:"environment:branch" aliases:"env:branch" help:"Create a new branch environment"`
	EnvironmentActivate   EnvironmentActivateCmd   `cmd:"" name:"environment:activate" aliases:"env:activate" help:"Activate an environment"`
	EnvironmentDeactivate EnvironmentDeactivateCmd `cmd:"" name:"environment:deactivate" aliases:"env:deactivate" help:"Deactivate an environment"`
	EnvironmentDelete     EnvironmentDeleteCmd     `cmd:"" name:"environment:delete" aliases:"env:delete" help:"Delete an environment"`
	Redeploy              RedeployCmd              `cmd:"" name:"redeploy" help:"Redeploy an environment"`
	Push                  PushCmd                  `cmd:"" name:"push" help:"Push code to trigger deployment"`
	ActivityList    ActivityListCmd    `cmd:"" name:"activity:list" aliases:"activities,act:list" help:"List activities"`
	ActivityLog     ActivityLogCmd     `cmd:"" name:"activity:log" aliases:"act:log" help:"Show activity log"`
	VariableList    VariableListCmd    `cmd:"" name:"variable:list" aliases:"variables,var:list" help:"List variables"`
	VariableGet     VariableGetCmd     `cmd:"" name:"variable:get" aliases:"var:get" help:"Get a variable"`
	VariableSet     VariableSetCmd     `cmd:"" name:"variable:set" aliases:"var:set" help:"Set a variable"`
	VariableDelete  VariableDeleteCmd  `cmd:"" name:"variable:delete" aliases:"var:delete" help:"Delete a variable"`
	SSH             SSHCmd             `cmd:"" name:"ssh" help:"SSH into an environment"`
}

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
		schema = &CommandSchema{
			Command:     command,
			Description: "No schema available for this command",
			GlobalFlags: globalFlags,
			ExitCodes:   defaultExitCodes,
		}
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

	formatter := output.New(outputFormat)
	return formatter.Write(summaries)
}
