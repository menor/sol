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

// API is the full interface for all API operations.
// It composes ProjectAPI, EnvironmentAPI, ActivityAPI, and VariableAPI interfaces.
type API interface {
	ProjectAPI
	EnvironmentAPI
	ActivityAPI
	VariableAPI
}

// Verify Client implements API at compile time.
var _ API = (*Client)(nil)
