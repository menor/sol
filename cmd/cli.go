package cmd

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"runtime/debug"
	"sort"

	"github.com/alecthomas/kong"
	"github.com/menor/sol/internal/errors"
	"github.com/menor/sol/internal/output"
)

// CLI is the main command-line interface struct.
// Global flags are defined here and available to all commands.
type CLI struct {
	// Global flags
	Output      string `help:"Output format" default:"toon" enum:"toon,json" short:"o"`
	Project     string `help:"Project ID" short:"p" env:"UPSUN_PROJECT"`
	Environment string `help:"Environment name" short:"e" env:"UPSUN_ENVIRONMENT"`
	Quiet       bool   `help:"Suppress non-essential output" short:"q"`
	NoCache     bool   `help:"Bypass cache" name:"no-cache"`
	Debug       bool   `help:"Show API request/response details"`
	Schema      bool   `help:"Output command schema instead of running"`

	// Commands
	Version                  VersionCmd                  `cmd:"" name:"version" help:"Print version information"`
	AuthLogin                AuthLoginCmd                `cmd:"" name:"auth:login" help:"Log in to Upsun"`
	AuthLogout               AuthLogoutCmd               `cmd:"" name:"auth:logout" help:"Log out of Upsun"`
	AuthInfo                 AuthInfoCmd                 `cmd:"" name:"auth:info" help:"Show authentication status"`
	ProjectList              ProjectListCmd              `cmd:"" name:"project:list" aliases:"projects" help:"List all projects"`
	ProjectInfo              ProjectInfoCmd              `cmd:"" name:"project:info" aliases:"project" help:"Show project details"`
	EnvironmentList          EnvironmentListCmd          `cmd:"" name:"environment:list" aliases:"environments,env:list" help:"List environments"`
	EnvironmentInfo          EnvironmentInfoCmd          `cmd:"" name:"environment:info" aliases:"env:info" help:"Show environment details"`
	EnvironmentBranch        EnvironmentBranchCmd        `cmd:"" name:"environment:branch" aliases:"env:branch" help:"Create a new branch environment"`
	EnvironmentActivate      EnvironmentActivateCmd      `cmd:"" name:"environment:activate" aliases:"env:activate" help:"Activate an environment"`
	EnvironmentDeactivate    EnvironmentDeactivateCmd    `cmd:"" name:"environment:deactivate" aliases:"env:deactivate" help:"Deactivate an environment"`
	EnvironmentDelete        EnvironmentDeleteCmd        `cmd:"" name:"environment:delete" aliases:"env:delete" help:"Delete an environment"`
	Redeploy                 RedeployCmd                 `cmd:"" name:"redeploy" help:"Redeploy an environment"`
	Push                     PushCmd                     `cmd:"" name:"push" help:"Push code to trigger deployment"`
	ActivityList             ActivityListCmd             `cmd:"" name:"activity:list" aliases:"activities,act:list" help:"List activities"`
	ActivityLog              ActivityLogCmd              `cmd:"" name:"activity:log" aliases:"act:log" help:"Show activity log"`
	VariableList             VariableListCmd             `cmd:"" name:"variable:list" aliases:"variables,var:list" help:"List variables"`
	VariableGet              VariableGetCmd              `cmd:"" name:"variable:get" aliases:"var:get" help:"Get a variable"`
	VariableSet              VariableSetCmd              `cmd:"" name:"variable:set" aliases:"var:set" help:"Set a variable"`
	VariableDelete           VariableDeleteCmd           `cmd:"" name:"variable:delete" aliases:"var:delete" help:"Delete a variable"`
	SSH                      SSHCmd                      `cmd:"" name:"ssh" help:"SSH into an environment"`
	ServiceList              ServiceListCmd              `cmd:"" name:"service:list" aliases:"services" help:"List services in an environment"`
	AppList                  AppListCmd                  `cmd:"" name:"app:list" aliases:"apps" help:"List applications in an environment"`
	AppConfigValidate        AppConfigValidateCmd        `cmd:"" name:"app:config-validate" aliases:"validate,lint" help:"Validate .upsun/config.yaml"`
	RouteList                RouteListCmd                `cmd:"" name:"route:list" aliases:"routes" help:"List routes for an environment"`
	EnvironmentURL           EnvironmentURLCmd           `cmd:"" name:"environment:url" aliases:"env:url" help:"Show URLs for an environment"`
	EnvironmentRelationships EnvironmentRelationshipsCmd `cmd:"" name:"environment:relationships" aliases:"env:relationships" help:"Show app-service relationships"`
	BackupList               BackupListCmd               `cmd:"" name:"backup:list" aliases:"backups" help:"List backups for an environment"`
	BackupGet                BackupGetCmd                `cmd:"" name:"backup:get" aliases:"backup" help:"Get backup details"`
	BackupCreate             BackupCreateCmd             `cmd:"" name:"backup:create" help:"Create a new backup"`
	BackupRestore            BackupRestoreCmd            `cmd:"" name:"backup:restore" help:"Restore a backup"`
	BackupDelete             BackupDeleteCmd             `cmd:"" name:"backup:delete" help:"Delete a backup"`
	OrganizationList         OrganizationListCmd         `cmd:"" name:"organization:list" aliases:"organizations,org:list" help:"List organizations"`
	OrganizationInfo         OrganizationInfoCmd         `cmd:"" name:"organization:info" aliases:"org:info" help:"Show organization details"`
	UserList                 UserListCmd                 `cmd:"" name:"user:list" aliases:"users" help:"List users with access to a project"`
	ResourcesGet             ResourcesGetCmd             `cmd:"" name:"resources:get" aliases:"resources" help:"Show resource allocation for an environment"`
	ResourcesSet             ResourcesSetCmd             `cmd:"" name:"resources:set" help:"Update resource allocation for a service"`
	IntegrationList          IntegrationListCmd          `cmd:"" name:"integration:list" aliases:"integrations,int:list" help:"List integrations for a project"`
	IntegrationGet           IntegrationGetCmd           `cmd:"" name:"integration:get" aliases:"int:get" help:"Show integration details"`
	EnvironmentMerge         EnvironmentMergeCmd         `cmd:"" name:"environment:merge" aliases:"env:merge" help:"Merge an environment into its parent"`
	EnvironmentSync          EnvironmentSyncCmd          `cmd:"" name:"environment:sync" aliases:"env:sync" help:"Sync data/code from parent environment"`
	DomainList               DomainListCmd               `cmd:"" name:"domain:list" aliases:"domains" help:"List custom domains for a project"`
	CertificateList          CertificateListCmd          `cmd:"" name:"certificate:list" aliases:"certificates,certs" help:"List SSL certificates for a project"`
	SSHKeyList               SSHKeyListCmd               `cmd:"" name:"ssh-key:list" aliases:"ssh-keys" help:"List SSH keys for the current user"`
}

// Execute parses command-line arguments and runs the appropriate command,
// returning the process exit code. Every error path funnels through render(),
// so main is reduced to os.Exit(cmd.Execute()).
func Execute() (exitCode int) {
	// An unexpected panic is an internal error (exit 70), not a crash with a
	// Go stack trace on stdout. The stack survives in details — without it a
	// "please report this" error is undiagnosable.
	defer func() {
		if r := recover(); r != nil {
			exitCode = render(nil, errors.NewInternalError(fmt.Sprintf("panic: %v", r)).
				WithDetail("stack", string(debug.Stack())))
		}
	}()

	// Handle --schema flag early, before Kong validates arguments.
	if hasSchemaFlag(os.Args) {
		if err := handleSchemaRequest(os.Args); err != nil {
			// A bad command name is a malformed invocation — same treatment
			// as a Kong parse failure (exit 80). Anything else (e.g. a write
			// failure) stays on the normal render path.
			var cliErr *errors.CLIError
			if stderrors.As(err, &cliErr) && cliErr.Code == errors.CodeInvalidArgument {
				return renderParseErrorIn(formatFromArgsOrDefault(os.Args, "json"), cliErr)
			}
			return render(nil, err)
		}
		return errors.ExitSuccess
	}

	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("sol"),
		kong.Description("Agent-optimized CLI for Upsun"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	if err != nil {
		return render(nil, err)
	}

	kongCtx, err := parser.Parse(os.Args[1:])
	if err != nil {
		return renderParseError(err)
	}

	// Create execution context
	ctx := &Context{
		Context: context.Background(),
		CLI:     &cli,
	}

	return render(&cli, kongCtx.Run(ctx))
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

	// The schema path historically defaults to json, not the global toon.
	outputFormat := formatFromArgsOrDefault(args, "json")

	// Find the command name: the first arg that is neither a flag nor the
	// value of -o/--output.
	var command string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--output" || arg == "-o" {
			i++ // skip the flag's value
			continue
		}
		if isFlag(arg) {
			continue
		}
		if command == "" {
			command = arg
		}
	}

	if command == "" {
		// No command specified, list all commands
		return listAllSchemas(outputFormat)
	}

	schema := GetCommandSchema(command)
	if schema == nil {
		return errors.NewValidationError(fmt.Sprintf("unknown command %q", command)).
			WithHint("Run 'sol --schema' to list all commands")
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
