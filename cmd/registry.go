package cmd

// CLI is the main command-line interface struct.
// Global flags are defined here and available to all commands.
// Kept in a build-tag-free file so both CLI and wasm targets can reflect over it.
type CLI struct {
	Output      string `help:"Output format" default:"toon" enum:"toon,json" short:"o"`
	Project     string `help:"Project ID" short:"p" env:"UPSUN_PROJECT"`
	Environment string `help:"Environment name" short:"e" env:"UPSUN_ENVIRONMENT"`
	Quiet       bool   `help:"Suppress non-essential output" short:"q"`
	NoCache     bool   `help:"Bypass cache" name:"no-cache"`
	Debug       bool   `help:"Show API request/response details"`
	Schema      bool   `help:"Output command schema instead of running"`

	Version               VersionCmd               `cmd:"" name:"version" help:"Print version information"`
	AuthLogin             AuthLoginCmd             `cmd:"" name:"auth:login" help:"Log in to Upsun"`
	AuthLogout            AuthLogoutCmd            `cmd:"" name:"auth:logout" help:"Log out of Upsun"`
	AuthInfo              AuthInfoCmd              `cmd:"" name:"auth:info" help:"Show authentication status"`
	ProjectList           ProjectListCmd           `cmd:"" name:"project:list" aliases:"projects" help:"List all projects"`
	ProjectInfo           ProjectInfoCmd           `cmd:"" name:"project:info" aliases:"project" help:"Show project details"`
	EnvironmentList       EnvironmentListCmd       `cmd:"" name:"environment:list" aliases:"environments,env:list" help:"List environments"`
	EnvironmentInfo       EnvironmentInfoCmd       `cmd:"" name:"environment:info" aliases:"env:info" help:"Show environment details"`
	EnvironmentBranch     EnvironmentBranchCmd     `cmd:"" name:"environment:branch" aliases:"env:branch" help:"Create a new branch environment"`
	EnvironmentActivate   EnvironmentActivateCmd   `cmd:"" name:"environment:activate" aliases:"env:activate" help:"Activate an environment"`
	EnvironmentDeactivate EnvironmentDeactivateCmd `cmd:"" name:"environment:deactivate" aliases:"env:deactivate" help:"Deactivate an environment"`
	EnvironmentDelete     EnvironmentDeleteCmd     `cmd:"" name:"environment:delete" aliases:"env:delete" help:"Delete an environment"`
	Redeploy              RedeployCmd              `cmd:"" name:"redeploy" help:"Redeploy an environment"`
	Push                  PushCmd                  `cmd:"" name:"push" help:"Push code to trigger deployment"`
	ActivityList          ActivityListCmd          `cmd:"" name:"activity:list" aliases:"activities,act:list" help:"List activities"`
	ActivityLog           ActivityLogCmd           `cmd:"" name:"activity:log" aliases:"act:log" help:"Show activity log"`
	VariableList          VariableListCmd          `cmd:"" name:"variable:list" aliases:"variables,var:list" help:"List variables"`
	VariableGet           VariableGetCmd           `cmd:"" name:"variable:get" aliases:"var:get" help:"Get a variable"`
	VariableSet           VariableSetCmd           `cmd:"" name:"variable:set" aliases:"var:set" help:"Set a variable"`
	VariableDelete        VariableDeleteCmd        `cmd:"" name:"variable:delete" aliases:"var:delete" help:"Delete a variable"`

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
