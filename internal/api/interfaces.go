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

// API is the full interface for all API operations.
// It composes ProjectAPI and EnvironmentAPI interfaces.
type API interface {
	ProjectAPI
	EnvironmentAPI
}

// Verify Client implements API at compile time.
var _ API = (*Client)(nil)
