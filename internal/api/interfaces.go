package api

import "context"

// ProjectAPI defines the project-related API operations.
// Commands depend on this interface rather than the concrete Client,
// enabling testing with mock implementations.
type ProjectAPI interface {
	ListProjects(ctx context.Context) ([]ProjectRef, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
}

// EnvironmentAPI defines the environment-related API operations.
type EnvironmentAPI interface {
	ListEnvironments(ctx context.Context, projectID string) ([]Environment, error)
	GetEnvironment(ctx context.Context, projectID, environmentID string) (*Environment, error)
	ActivateEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error)
	DeactivateEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error)
	DeleteEnvironment(ctx context.Context, projectID, environmentID string) error
	RedeployEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error)
	BranchEnvironment(ctx context.Context, projectID, parentEnvID string, input *BranchInput) (*Activity, error)
}

// ActivityAPI defines the activity-related API operations.
type ActivityAPI interface {
	ListActivities(ctx context.Context, projectID string, opts *ListActivitiesOptions) ([]Activity, error)
	GetActivity(ctx context.Context, projectID, activityID string) (*Activity, error)
	GetActivityLog(ctx context.Context, projectID, activityID string) (string, error)
}

// VariableAPI defines the variable-related API operations.
type VariableAPI interface {
	// Project-level variables
	ListProjectVariables(ctx context.Context, projectID string) ([]Variable, error)
	GetProjectVariable(ctx context.Context, projectID, name string) (*Variable, error)
	SetProjectVariable(ctx context.Context, projectID string, input *VariableInput) (*Variable, error)
	DeleteProjectVariable(ctx context.Context, projectID, name string) error

	// Environment-level variables
	ListEnvironmentVariables(ctx context.Context, projectID, envID string) ([]Variable, error)
	GetEnvironmentVariable(ctx context.Context, projectID, envID, name string) (*Variable, error)
	SetEnvironmentVariable(ctx context.Context, projectID, envID string, input *VariableInput) (*Variable, error)
	DeleteEnvironmentVariable(ctx context.Context, projectID, envID, name string) error
}

// DeploymentAPI defines deployment-related API operations.
// These provide observability into the current deployment state.
type DeploymentAPI interface {
	GetCurrentDeployment(ctx context.Context, projectID, envID string) (*Deployment, error)
	ListServices(ctx context.Context, projectID, envID string) ([]ServiceSummary, error)
	GetRoutes(ctx context.Context, projectID, envID string) ([]RouteURL, error)
	GetRelationships(ctx context.Context, projectID, envID, appName string) ([]Relationship, error)
}

// API is the full interface for all API operations.
// It composes ProjectAPI, EnvironmentAPI, ActivityAPI, VariableAPI, and DeploymentAPI interfaces.
type API interface {
	ProjectAPI
	EnvironmentAPI
	ActivityAPI
	VariableAPI
	DeploymentAPI
}

// Verify Client implements API at compile time.
var _ API = (*Client)(nil)
