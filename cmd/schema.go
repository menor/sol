package cmd

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
	Type       string                  `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Items      *OutputSchema           `json:"items,omitempty"`
}

// PropertySchema describes a single output property.
type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// globalFlags lists flags available on all commands.
var globalFlags = []string{"output", "project", "environment", "quiet", "no-cache", "debug", "schema"}

// defaultExitCodes are used by all commands.
var defaultExitCodes = map[string]string{
	"0": "Success",
	"1": "User error (bad input, auth failed)",
	"2": "API error (server error, network issue)",
	"3": "Internal error (bug in CLI)",
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
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
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
		},
		Examples:  []string{"sol environment:list --project abc123"},
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
			{Name: "type", Type: "string", Description: "Filter by activity type"},
		},
		GlobalFlags: globalFlags,
		Output: &OutputSchema{
			Type: "array",
			Items: &OutputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"id":          {Type: "string", Description: "Activity ID"},
					"type":        {Type: "string", Description: "Activity type"},
					"state":       {Type: "string", Description: "Activity state"},
					"result":      {Type: "string", Description: "Activity result"},
					"description": {Type: "string", Description: "Activity description"},
					"created_at":  {Type: "string", Description: "Creation timestamp"},
				},
			},
		},
		Examples:  []string{"sol activity:list --project abc123", "sol activity:list --project abc123 --state complete --limit 5"},
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
}

// GetCommandSchema returns the schema for a command, or nil if not found.
func GetCommandSchema(command string) *CommandSchema {
	schema, ok := commandSchemas[command]
	if !ok {
		return nil
	}
	return &schema
}

// ListCommandSchemas returns all available command schemas.
func ListCommandSchemas() map[string]CommandSchema {
	return commandSchemas
}
