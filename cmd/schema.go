package cmd

import (
	"strconv"

	"github.com/menor/sol/internal/errors"
)

// CommandSchema describes a command's interface for machine consumption.
type CommandSchema struct {
	Command     string            `json:"command"`
	Description string            `json:"description"`
	Flags       []FlagSchema      `json:"flags,omitempty"`
	Arguments   []ArgumentSchema  `json:"arguments,omitempty"`
	GlobalFlags []string          `json:"global_flags"`
	Output      *OutputSchema     `json:"output,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
	ExitCodes   map[string]string `json:"exit_codes"`
	Errors      *ErrorSchema      `json:"errors,omitempty"`
}

// ErrorSchema describes the error contract shared by every command, so agents
// can discover failure handling the same way they discover flags.
type ErrorSchema struct {
	Description string            `json:"description"`
	Envelope    map[string]string `json:"envelope"`
	Codes       map[string]string `json:"codes"`
}

// FlagSchema describes a command flag.
type FlagSchema struct {
	Name        string `json:"name"`
	Short       string `json:"short,omitempty"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required,omitempty"`
	Default     any    `json:"default,omitempty"`
}

// ArgumentSchema describes a positional argument.
type ArgumentSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required,omitempty"`
}

// OutputSchema describes the command output structure.
type OutputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Items      *OutputSchema             `json:"items,omitempty"`
}

// PropertySchema describes a single output property.
type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// globalFlags lists flags available on all commands.
var globalFlags = []string{"output", "project", "environment", "quiet", "no-cache", "debug", "schema"}

// defaultExitCodes are used by all commands. Keys are built from the
// internal/errors constants so the schema cannot drift from the real scheme.
var defaultExitCodes = map[string]string{
	strconv.Itoa(errors.ExitSuccess):   "Success",
	strconv.Itoa(errors.ExitUserError): "Operational error (recoverable; read code and hint from the error envelope)",
	strconv.Itoa(errors.ExitInternal):  "Internal error (bug in Sol; do not retry blindly)",
	strconv.Itoa(errors.ExitUsage):     "Usage/parse error (fix the invocation)",
}

// defaultErrorSchema advertises the error contract. One shared value: every
// command fails through the same render path with the same envelope.
var defaultErrorSchema = &ErrorSchema{
	Description: "On failure the error envelope renders in the active output format (-o) on stdout as {\"error\": {...}}; stderr carries only logs. Branch on code, never on message text.",
	Envelope: map[string]string{
		"code":      "Stable snake_case identifier from the closed set in codes",
		"message":   "Human-readable description",
		"hint":      "Actionable next step; omitted when there is none",
		"retryable": "true if the identical call may later succeed",
		"details":   "Extra context (e.g. status_code); omitted when empty",
	},
	Codes: map[string]string{
		errors.CodeUnauthenticated:        "Not authenticated, or authentication expired",
		errors.CodeNoProjectSpecified:     "No project resolved from flag or environment",
		errors.CodeNoEnvironmentSpecified: "No environment resolved from flag or environment",
		errors.CodeNotFound:               "Resource does not exist",
		errors.CodeInvalidArgument:        "Invalid input value or malformed invocation",
		errors.CodePermissionDenied:       "Authenticated but not allowed",
		errors.CodeAPIUnavailable:         "Upsun API unreachable, 5xx, or rate-limited (retryable)",
		errors.CodeOperationFailed:        "Remote operation failed, was cancelled, or timed out (timeout is retryable)",
		errors.CodeInternal:               "Bug in Sol itself",
	},
}

// commandSchemas holds schema definitions for each command.
var commandSchemas = map[string]CommandSchema{
	"version": {
		Command:     "version",
		Description: "Print version information",
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"version": {Type: "string", Description: "Version string"},
				"commit":  {Type: "string", Description: "Git commit hash"},
				"date":    {Type: "string", Description: "Build date"},
			},
		},
		Examples:  []string{"sol version"},
		ExitCodes: defaultExitCodes,
	},
	"auth:login": {
		Command:     "auth:login",
		Description: "Log in to Upsun via OAuth browser flow",
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"status":  {Type: "string", Description: "Login status"},
				"message": {Type: "string", Description: "Status message"},
			},
		},
		Examples:  []string{"sol auth:login"},
		ExitCodes: defaultExitCodes,
	},
	"auth:logout": {
		Command:     "auth:logout",
		Description: "Log out of Upsun (removes stored credentials)",
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"status":  {Type: "string", Description: "Logout status"},
				"message": {Type: "string", Description: "Status message"},
			},
		},
		Examples:  []string{"sol auth:logout"},
		ExitCodes: defaultExitCodes,
	},
	"auth:info": {
		Command:     "auth:info",
		Description: "Show current authentication status",
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"authenticated": {Type: "boolean", Description: "Whether user is authenticated"},
				"method":        {Type: "string", Description: "Authentication method (oauth, env)"},
				"username":      {Type: "string", Description: "Authenticated user name"},
				"email":         {Type: "string", Description: "Authenticated user email"},
				"expired":       {Type: "boolean", Description: "Whether token is expired"},
			},
		},
		Examples:  []string{"sol auth:info"},
		ExitCodes: defaultExitCodes,
	},
	"project:list": {
		Command:     "project:list",
		Description: "List all projects accessible to the authenticated user",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (status, org, subscription, timestamps)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":     {Type: "string", Description: "Project ID"},
					"title":  {Type: "string", Description: "Project title"},
					"region": {Type: "string", Description: "Region hostname"},
				},
			},
		},
		Examples:  []string{"sol project:list", "sol project:list --full"},
		ExitCodes: defaultExitCodes,
	},
	"project:info": {
		Command:     "project:info",
		Description: "Show details for a specific project",
		Arguments: []ArgumentSchema{
			{Name: "project_id", Type: "string", Description: "Project ID (optional if --project set)", Required: false},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":              {Type: "string", Description: "Project ID"},
				"title":           {Type: "string", Description: "Project title"},
				"region":          {Type: "string", Description: "Region hostname"},
				"status":          {Type: "string", Description: "Project status"},
				"organization_id": {Type: "string", Description: "Organization ID"},
				"default_branch":  {Type: "string", Description: "Default branch name"},
			},
		},
		Examples:  []string{"sol project:info abc123", "sol project:info --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"environment:list": {
		Command:     "environment:list",
		Description: "List environments for a project",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (type, machine_name, timestamps, etc.)"},
			{Name: "status", Type: "string", Description: "Filter by status (active, inactive, dirty)"},
			{Name: "no-inactive", Type: "bool", Description: "Exclude inactive environments"},
			{Name: "type", Type: "string", Description: "Filter by type (production, staging, development)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":     {Type: "string", Description: "Environment ID"},
					"name":   {Type: "string", Description: "Environment name"},
					"status": {Type: "string", Description: "Environment status"},
					"parent": {Type: "string", Description: "Parent environment ID"},
				},
			},
		},
		Examples:  []string{"sol environment:list --project abc123", "sol environment:list -p abc123 --status active", "sol environment:list -p abc123 --no-inactive"},
		ExitCodes: defaultExitCodes,
	},
	"environment:info": {
		Command:     "environment:info",
		Description: "Show details for a specific environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":     {Type: "string", Description: "Environment ID"},
				"name":   {Type: "string", Description: "Environment name"},
				"title":  {Type: "string", Description: "Environment title"},
				"type":   {Type: "string", Description: "Environment type"},
				"status": {Type: "string", Description: "Environment status"},
				"parent": {Type: "string", Description: "Parent environment ID"},
			},
		},
		Examples:  []string{"sol environment:info main --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"environment:branch": {
		Command:     "environment:branch",
		Description: "Create a new branch environment from an existing one",
		Arguments: []ArgumentSchema{
			{Name: "name", Type: "string", Description: "Name for the new environment", Required: true},
		},
		Flags: []FlagSchema{
			{Name: "parent", Type: "string", Description: "Parent environment to branch from", Default: "main"},
			{Name: "title", Type: "string", Description: "Title for the new environment"},
			{Name: "wait", Short: "w", Type: "boolean", Description: "Wait for activity to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"activity_id": {Type: "string", Description: "ID of the branch activity"},
				"state":       {Type: "string", Description: "Activity state"},
			},
		},
		Examples:  []string{"sol environment:branch feature-x --project abc123", "sol environment:branch feature-x --parent staging --wait"},
		ExitCodes: defaultExitCodes,
	},
	"environment:activate": {
		Command:     "environment:activate",
		Description: "Activate an inactive environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "wait", Short: "w", Type: "boolean", Description: "Wait for activity to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"activity_id": {Type: "string", Description: "ID of the activation activity"},
				"state":       {Type: "string", Description: "Activity state"},
			},
		},
		Examples:  []string{"sol environment:activate staging --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"environment:deactivate": {
		Command:     "environment:deactivate",
		Description: "Deactivate an active environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "wait", Short: "w", Type: "boolean", Description: "Wait for activity to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"activity_id": {Type: "string", Description: "ID of the deactivation activity"},
				"state":       {Type: "string", Description: "Activity state"},
			},
		},
		Examples:  []string{"sol environment:deactivate staging --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"environment:delete": {
		Command:     "environment:delete",
		Description: "Delete an environment (must be deactivated first)",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"status":  {Type: "string", Description: "Deletion status"},
				"message": {Type: "string", Description: "Status message"},
			},
		},
		Examples:  []string{"sol environment:delete old-feature --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"activity:list": {
		Command:     "activity:list",
		Description: "List activities for a project or environment",
		Flags: []FlagSchema{
			{Name: "limit", Type: "integer", Description: "Maximum number of activities to return", Default: 10},
			{Name: "state", Type: "string", Description: "Filter by state (pending, in_progress, complete)"},
			{Name: "result", Type: "string", Description: "Filter by result (success, failure)"},
			{Name: "type", Type: "string", Description: "Filter by activity type"},
			{Name: "exclude-type", Type: "string", Description: "Exclude activities of this type"},
			{Name: "start", Type: "string", Description: "Only activities after this date (ISO 8601 format)"},
			{Name: "incomplete", Type: "bool", Description: "Only show incomplete activities (pending or in_progress)"},
			{Name: "all", Short: "a", Type: "bool", Description: "Show all activities (ignore limit)"},
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (result, description, timestamps, etc.)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":         {Type: "string", Description: "Activity ID"},
					"type":       {Type: "string", Description: "Activity type"},
					"state":      {Type: "string", Description: "Activity state"},
					"created_at": {Type: "string", Description: "Creation timestamp"},
				},
			},
		},
		Examples:  []string{"sol activity:list --project abc123", "sol activity:list -p abc123 --result failure --limit 5", "sol activity:list -p abc123 --incomplete", "sol activity:list -p abc123 --all"},
		ExitCodes: defaultExitCodes,
	},
	"activity:log": {
		Command:     "activity:log",
		Description: "Show the log output for an activity",
		Arguments: []ArgumentSchema{
			{Name: "activity_id", Type: "string", Description: "Activity ID", Required: true},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"log": {Type: "string", Description: "Activity log output"},
			},
		},
		Examples:  []string{"sol activity:log abc123 --project proj123"},
		ExitCodes: defaultExitCodes,
	},
	"variable:list": {
		Command:     "variable:list",
		Description: "List variables for a project or environment",
		Flags: []FlagSchema{
			{Name: "level", Short: "l", Type: "string", Description: "Variable level: project or environment"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name":         {Type: "string", Description: "Variable name"},
					"value":        {Type: "string", Description: "Variable value (hidden if sensitive)"},
					"is_sensitive": {Type: "boolean", Description: "Whether value is hidden"},
					"is_json":      {Type: "boolean", Description: "Whether value is JSON"},
				},
			},
		},
		Examples:  []string{"sol variable:list --project abc123", "sol variable:list --project abc123 --environment main"},
		ExitCodes: defaultExitCodes,
	},
	"variable:get": {
		Command:     "variable:get",
		Description: "Get a specific variable",
		Arguments: []ArgumentSchema{
			{Name: "name", Type: "string", Description: "Variable name", Required: true},
		},
		Flags: []FlagSchema{
			{Name: "level", Short: "l", Type: "string", Description: "Variable level: project or environment"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"name":         {Type: "string", Description: "Variable name"},
				"value":        {Type: "string", Description: "Variable value (hidden if sensitive)"},
				"is_sensitive": {Type: "boolean", Description: "Whether value is hidden"},
				"is_json":      {Type: "boolean", Description: "Whether value is JSON"},
			},
		},
		Examples:  []string{"sol variable:get MY_VAR --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"variable:set": {
		Command:     "variable:set",
		Description: "Set a variable value",
		Arguments: []ArgumentSchema{
			{Name: "name", Type: "string", Description: "Variable name", Required: true},
			{Name: "value", Type: "string", Description: "Variable value", Required: true},
		},
		Flags: []FlagSchema{
			{Name: "level", Short: "l", Type: "string", Description: "Variable level: project or environment"},
			{Name: "sensitive", Short: "s", Type: "boolean", Description: "Mark as sensitive (hidden in output)"},
			{Name: "json", Short: "j", Type: "boolean", Description: "Value is JSON"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"name":   {Type: "string", Description: "Variable name"},
				"status": {Type: "string", Description: "Operation status"},
			},
		},
		Examples:  []string{"sol variable:set MY_VAR value --project abc123", "sol variable:set SECRET value --project abc123 --sensitive"},
		ExitCodes: defaultExitCodes,
	},
	"variable:delete": {
		Command:     "variable:delete",
		Description: "Delete a variable",
		Arguments: []ArgumentSchema{
			{Name: "name", Type: "string", Description: "Variable name", Required: true},
		},
		Flags: []FlagSchema{
			{Name: "level", Short: "l", Type: "string", Description: "Variable level: project or environment"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"status":  {Type: "string", Description: "Deletion status"},
				"message": {Type: "string", Description: "Status message"},
			},
		},
		Examples:  []string{"sol variable:delete MY_VAR --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"ssh": {
		Command:     "ssh",
		Description: "SSH into an environment",
		Flags: []FlagSchema{
			{Name: "app", Short: "A", Type: "string", Description: "App name for multi-app projects"},
		},
		GlobalFlags: globalFlags,
		Examples:    []string{"sol ssh --project abc123 --environment main", "sol ssh --project abc123 --app frontend"},
		ExitCodes:   defaultExitCodes,
	},
	"redeploy": {
		Command:     "redeploy",
		Description: "Redeploy an environment (runs post_deploy hook only)",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "wait", Short: "w", Type: "boolean", Description: "Wait for activity to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"activity_id": {Type: "string", Description: "ID of the redeploy activity"},
				"state":       {Type: "string", Description: "Activity state"},
			},
		},
		Examples:  []string{"sol redeploy --project abc123 --environment main", "sol redeploy --project abc123 --environment main --wait"},
		ExitCodes: defaultExitCodes,
	},
	"push": {
		Command:     "push",
		Description: "Push code to Upsun (triggers deployment)",
		Flags: []FlagSchema{
			{Name: "target", Short: "t", Type: "string", Description: "Target branch (defaults to current branch)"},
			{Name: "force", Short: "f", Type: "boolean", Description: "Force push"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"status":  {Type: "string", Description: "Push status"},
				"message": {Type: "string", Description: "Git output"},
			},
		},
		Examples:  []string{"sol push --project abc123", "sol push --project abc123 --target staging"},
		ExitCodes: defaultExitCodes,
	},
	"service:list": {
		Command:     "service:list",
		Description: "List services (databases, caches, etc.) in an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (size, configuration, etc.)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name": {Type: "string", Description: "Service name"},
					"type": {Type: "string", Description: "Service type (mysql, redis, etc.)"},
					"disk": {Type: "integer", Description: "Disk size in MB"},
				},
			},
		},
		Examples:  []string{"sol service:list --project abc123 --environment main", "sol service:list -p abc123 -e main --full"},
		ExitCodes: defaultExitCodes,
	},
	"environment:url": {
		Command:     "environment:url",
		Description: "Show URLs for an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "primary", Short: "1", Type: "bool", Description: "Show only the primary URL"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"url":     {Type: "string", Description: "Route URL"},
					"primary": {Type: "boolean", Description: "Whether this is the primary URL"},
					"type":    {Type: "string", Description: "Route type (upstream, redirect)"},
				},
			},
		},
		Examples:  []string{"sol environment:url --project abc123 --environment main", "sol environment:url -p abc123 -e main --primary"},
		ExitCodes: defaultExitCodes,
	},
	"environment:relationships": {
		Command:     "environment:relationships",
		Description: "Show relationships between apps and services",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "app", Short: "A", Type: "string", Description: "Filter by app name"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"app":      {Type: "string", Description: "App name"},
					"name":     {Type: "string", Description: "Relationship name"},
					"service":  {Type: "string", Description: "Service name"},
					"endpoint": {Type: "string", Description: "Service endpoint"},
				},
			},
		},
		Examples:  []string{"sol environment:relationships --project abc123 --environment main", "sol environment:relationships -p abc123 -e main --app frontend"},
		ExitCodes: defaultExitCodes,
	},
	"app:list": {
		Command:     "app:list",
		Description: "List applications (webapps and workers) in an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (mounts, relationships, configuration)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name":   {Type: "string", Description: "App name"},
					"type":   {Type: "string", Description: "App type (nodejs, php, python, etc.)"},
					"size":   {Type: "string", Description: "Container size"},
					"disk":   {Type: "integer", Description: "Disk size in MB"},
					"worker": {Type: "boolean", Description: "Whether this is a worker process"},
				},
			},
		},
		Examples:  []string{"sol app:list --project abc123 --environment main", "sol app:list -p abc123 -e main --full"},
		ExitCodes: defaultExitCodes,
	},
	"app:config-validate": {
		Command:     "app:config-validate",
		Description: "Validate .upsun/config.yaml configuration file",
		Arguments: []ArgumentSchema{
			{Name: "path", Type: "string", Description: "Path to config file or directory (defaults to current directory)", Required: false},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"valid":        {Type: "boolean", Description: "Whether the config is valid"},
				"file_path":    {Type: "string", Description: "Path to the validated file"},
				"applications": {Type: "array", Description: "List of validated applications"},
				"services":     {Type: "array", Description: "List of service names"},
				"routes":       {Type: "integer", Description: "Number of routes defined"},
				"errors":       {Type: "array", Description: "Validation errors (if any)"},
				"warnings":     {Type: "array", Description: "Validation warnings (if any)"},
			},
		},
		Examples:  []string{"sol app:config-validate", "sol app:config-validate .upsun/config.yaml", "sol app:config-validate /path/to/project"},
		ExitCodes: defaultExitCodes,
	},
	"route:list": {
		Command:     "route:list",
		Description: "List routes for an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (TLS settings, redirects, cache)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"url":      {Type: "string", Description: "Route URL"},
					"primary":  {Type: "boolean", Description: "Whether this is the primary route"},
					"type":     {Type: "string", Description: "Route type (upstream, redirect)"},
					"upstream": {Type: "string", Description: "Upstream app:endpoint for upstream routes"},
					"to":       {Type: "string", Description: "Redirect target for redirect routes"},
				},
			},
		},
		Examples:  []string{"sol route:list --project abc123 --environment main", "sol route:list -p abc123 -e main --full"},
		ExitCodes: defaultExitCodes,
	},
	"backup:list": {
		Command:     "backup:list",
		Description: "List backups for an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (expiry, size, status)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":         {Type: "string", Description: "Backup ID"},
					"created_at": {Type: "string", Description: "Creation timestamp"},
					"safe":       {Type: "boolean", Description: "Whether backup is consistent (services paused)"},
					"automated":  {Type: "boolean", Description: "Whether backup was automated"},
					"commit_id":  {Type: "string", Description: "Git commit ID"},
				},
			},
		},
		Examples:  []string{"sol backup:list --project abc123 --environment main", "sol backup:list -p abc123 -e main --full"},
		ExitCodes: defaultExitCodes,
	},
	"backup:get": {
		Command:     "backup:get",
		Description: "Get details of a specific backup",
		Arguments: []ArgumentSchema{
			{Name: "backup_id", Type: "string", Description: "Backup ID", Required: true},
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":              {Type: "string", Description: "Backup ID"},
				"created_at":      {Type: "string", Description: "Creation timestamp"},
				"expires_at":      {Type: "string", Description: "Expiration timestamp"},
				"safe":            {Type: "boolean", Description: "Whether backup is consistent"},
				"automated":       {Type: "boolean", Description: "Whether backup was automated"},
				"restorable":      {Type: "boolean", Description: "Whether backup can be restored"},
				"commit_id":       {Type: "string", Description: "Git commit ID"},
				"environment":     {Type: "string", Description: "Source environment"},
				"size_of_volumes": {Type: "integer", Description: "Backup size in bytes"},
			},
		},
		Examples:  []string{"sol backup:get backup123 --project abc123 --environment main"},
		ExitCodes: defaultExitCodes,
	},
	"backup:create": {
		Command:     "backup:create",
		Description: "Create a new backup of an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "live", Short: "l", Type: "bool", Description: "Create live backup (no pause, may have inconsistencies)"},
			{Name: "wait", Short: "w", Type: "bool", Description: "Wait for backup to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":         {Type: "string", Description: "Activity ID"},
				"type":       {Type: "string", Description: "Activity type"},
				"state":      {Type: "string", Description: "Activity state"},
				"created_at": {Type: "string", Description: "Creation timestamp"},
			},
		},
		Examples:  []string{"sol backup:create --project abc123 --environment main", "sol backup:create -p abc123 -e main --wait", "sol backup:create --live"},
		ExitCodes: defaultExitCodes,
	},
	"backup:restore": {
		Command:     "backup:restore",
		Description: "Restore a backup to an environment",
		Arguments: []ArgumentSchema{
			{Name: "backup_id", Type: "string", Description: "Backup ID to restore", Required: true},
			{Name: "environment_id", Type: "string", Description: "Source environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "target", Short: "t", Type: "string", Description: "Target environment name (defaults to source)"},
			{Name: "branch-from", Short: "b", Type: "string", Description: "Parent branch for new environment"},
			{Name: "restore-code", Type: "bool", Description: "Restore code from backup (default: true)", Default: true},
			{Name: "wait", Short: "w", Type: "bool", Description: "Wait for restore to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":         {Type: "string", Description: "Activity ID"},
				"type":       {Type: "string", Description: "Activity type"},
				"state":      {Type: "string", Description: "Activity state"},
				"created_at": {Type: "string", Description: "Creation timestamp"},
			},
		},
		Examples:  []string{"sol backup:restore backup123 --project abc123 --environment main", "sol backup:restore backup123 --target staging --wait"},
		ExitCodes: defaultExitCodes,
	},
	"backup:delete": {
		Command:     "backup:delete",
		Description: "Delete a backup",
		Arguments: []ArgumentSchema{
			{Name: "backup_id", Type: "string", Description: "Backup ID to delete", Required: true},
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"deleted":   {Type: "boolean", Description: "Whether deletion succeeded"},
				"backup_id": {Type: "string", Description: "Deleted backup ID"},
			},
		},
		Examples:  []string{"sol backup:delete backup123 --project abc123 --environment main"},
		ExitCodes: defaultExitCodes,
	},
	"organization:list": {
		Command:     "organization:list",
		Description: "List organizations for the current user",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (owner, country, capabilities)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":    {Type: "string", Description: "Organization ID"},
					"name":  {Type: "string", Description: "Organization name (slug)"},
					"label": {Type: "string", Description: "Organization display label"},
				},
			},
		},
		Examples:  []string{"sol organization:list", "sol organization:list --full"},
		ExitCodes: defaultExitCodes,
	},
	"organization:info": {
		Command:     "organization:info",
		Description: "Show details for an organization",
		Arguments: []ArgumentSchema{
			{Name: "org_id", Type: "string", Description: "Organization ID", Required: true},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":           {Type: "string", Description: "Organization ID"},
				"name":         {Type: "string", Description: "Organization name (slug)"},
				"label":        {Type: "string", Description: "Organization display label"},
				"owner":        {Type: "string", Description: "Owner user ID"},
				"country":      {Type: "string", Description: "Country code"},
				"capabilities": {Type: "array", Description: "Organization capabilities"},
			},
		},
		Examples:  []string{"sol organization:info org123"},
		ExitCodes: defaultExitCodes,
	},
	"user:list": {
		Command:     "user:list",
		Description: "List users with access to a project",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (permissions)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"user_id": {Type: "string", Description: "User ID"},
					"email":   {Type: "string", Description: "User email"},
					"role":    {Type: "string", Description: "User role"},
				},
			},
		},
		Examples:  []string{"sol user:list --project abc123", "sol user:list -p abc123 --full"},
		ExitCodes: defaultExitCodes,
	},
	"resources:get": {
		Command:     "resources:get",
		Description: "Show resource allocation for an environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (memory ratios, profile details)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"apps":     {Type: "array", Description: "Application resource allocations"},
				"services": {Type: "array", Description: "Service resource allocations"},
				"workers":  {Type: "array", Description: "Worker resource allocations"},
			},
		},
		Examples:  []string{"sol resources:get --project abc123 --environment main", "sol resources:get -p abc123 -e main --full"},
		ExitCodes: defaultExitCodes,
	},
	"resources:set": {
		Command:     "resources:set",
		Description: "Update resource allocation for a service",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "service", Short: "s", Type: "string", Description: "Service name to update", Required: true},
			{Name: "size", Type: "string", Description: "Resource size profile (e.g., S, M, L, XL)"},
			{Name: "disk", Type: "integer", Description: "Disk size in MB"},
			{Name: "instances", Type: "integer", Description: "Number of instances"},
			{Name: "wait", Short: "w", Type: "bool", Description: "Wait for deployment to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":         {Type: "string", Description: "Activity ID"},
				"type":       {Type: "string", Description: "Activity type"},
				"state":      {Type: "string", Description: "Activity state"},
				"created_at": {Type: "string", Description: "Creation timestamp"},
			},
		},
		Examples:  []string{"sol resources:set --service myapp --size L --project abc123 --environment main", "sol resources:set -s database --disk 2048 -p abc123 -e main --wait"},
		ExitCodes: defaultExitCodes,
	},
	"integration:list": {
		Command:     "integration:list",
		Description: "List integrations for a project",
		Flags: []FlagSchema{
			{Name: "type", Short: "t", Type: "string", Description: "Filter by integration type (github, gitlab, webhook, etc.)"},
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (configuration details)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":   {Type: "string", Description: "Integration ID"},
					"type": {Type: "string", Description: "Integration type"},
				},
			},
		},
		Examples:  []string{"sol integration:list --project abc123", "sol integration:list -p abc123 --type github", "sol integration:list -p abc123 --full"},
		ExitCodes: defaultExitCodes,
	},
	"integration:get": {
		Command:     "integration:get",
		Description: "Show details for a specific integration",
		Arguments: []ArgumentSchema{
			{Name: "integration_id", Type: "string", Description: "Integration ID", Required: true},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":         {Type: "string", Description: "Integration ID"},
				"type":       {Type: "string", Description: "Integration type"},
				"created_at": {Type: "string", Description: "Creation timestamp"},
				"repository": {Type: "string", Description: "Repository URL (for VCS integrations)"},
				"url":        {Type: "string", Description: "Webhook URL (for webhook integrations)"},
			},
		},
		Examples:  []string{"sol integration:get int123 --project abc123"},
		ExitCodes: defaultExitCodes,
	},
	"environment:merge": {
		Command:     "environment:merge",
		Description: "Merge an environment into its parent",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "wait", Short: "w", Type: "bool", Description: "Wait for activity to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":         {Type: "string", Description: "Activity ID"},
				"type":       {Type: "string", Description: "Activity type"},
				"state":      {Type: "string", Description: "Activity state"},
				"created_at": {Type: "string", Description: "Creation timestamp"},
			},
		},
		Examples:  []string{"sol environment:merge feature-x --project abc123", "sol environment:merge -e feature-x -p abc123 --wait"},
		ExitCodes: defaultExitCodes,
	},
	"environment:sync": {
		Command:     "environment:sync",
		Description: "Synchronize data and/or code from the parent environment",
		Arguments: []ArgumentSchema{
			{Name: "environment_id", Type: "string", Description: "Environment ID (optional if --environment set)", Required: false},
		},
		Flags: []FlagSchema{
			{Name: "data", Short: "d", Type: "bool", Description: "Synchronize data from parent"},
			{Name: "code", Short: "c", Type: "bool", Description: "Synchronize code from parent"},
			{Name: "resources", Short: "r", Type: "bool", Description: "Synchronize resources from parent"},
			{Name: "wait", Short: "w", Type: "bool", Description: "Wait for activity to complete"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"id":         {Type: "string", Description: "Activity ID"},
				"type":       {Type: "string", Description: "Activity type"},
				"state":      {Type: "string", Description: "Activity state"},
				"created_at": {Type: "string", Description: "Creation timestamp"},
			},
		},
		Examples:  []string{"sol environment:sync staging --data --project abc123", "sol environment:sync -e staging -p abc123 --data --code --wait"},
		ExitCodes: defaultExitCodes,
	},
	"domain:list": {
		Command:     "domain:list",
		Description: "List custom domains for a project",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (SSL details, timestamps)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"name":       {Type: "string", Description: "Domain name"},
					"is_default": {Type: "boolean", Description: "Whether this is the default domain"},
				},
			},
		},
		Examples:  []string{"sol domain:list --project abc123", "sol domain:list -p abc123 --full"},
		ExitCodes: defaultExitCodes,
	},
	"certificate:list": {
		Command:     "certificate:list",
		Description: "List SSL certificates for a project",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (issuer, chain, timestamps)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":             {Type: "string", Description: "Certificate ID"},
					"domains":        {Type: "array", Description: "Domains covered by this certificate"},
					"expires_at":     {Type: "string", Description: "Expiration timestamp"},
					"is_provisioned": {Type: "boolean", Description: "Whether certificate is provisioned"},
				},
			},
		},
		Examples:  []string{"sol certificate:list --project abc123", "sol certificate:list -p abc123 --full"},
		ExitCodes: defaultExitCodes,
	},
	"ssh-key:list": {
		Command:     "ssh-key:list",
		Description: "List SSH keys for the current user",
		Flags: []FlagSchema{
			{Name: "full", Short: "f", Type: "bool", Description: "Include all fields (public key value, timestamps)"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"key_id":      {Type: "string", Description: "SSH key ID"},
					"title":       {Type: "string", Description: "Key title/name"},
					"fingerprint": {Type: "string", Description: "Key fingerprint"},
				},
			},
		},
		Examples:  []string{"sol ssh-key:list", "sol ssh-key:list --full"},
		ExitCodes: defaultExitCodes,
	},
}

// GetCommandSchema returns the schema for a command, or nil if not found.
func GetCommandSchema(command string) *CommandSchema {
	schema, ok := commandSchemas[command]
	if !ok {
		return nil
	}
	// The error contract is identical for every command, so it is attached
	// here instead of being repeated in each commandSchemas entry.
	schema.Errors = defaultErrorSchema
	return &schema
}

// ListCommandSchemas returns all available command schemas.
func ListCommandSchemas() map[string]CommandSchema {
	return commandSchemas
}
