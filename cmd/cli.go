package cmd

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
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
