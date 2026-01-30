package api

import "context"

// MockClient is a mock implementation of the API interface for testing.
// Set the return values and errors before calling methods.
type MockClient struct {
	// Project methods
	ListProjectsFunc func(ctx context.Context) ([]ProjectRef, error)
	GetProjectFunc   func(ctx context.Context, projectID string) (*Project, error)

	// Environment methods
	ListEnvironmentsFunc func(ctx context.Context, projectID string) ([]Environment, error)
	GetEnvironmentFunc   func(ctx context.Context, projectID, environmentID string) (*Environment, error)

	// Track calls for assertions
	Calls []MockCall
}

// MockCall records a method call with its arguments.
type MockCall struct {
	Method string
	Args   []any
}

// ListProjects implements ProjectAPI.
func (m *MockClient) ListProjects(ctx context.Context) ([]ProjectRef, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListProjects", Args: nil})
	if m.ListProjectsFunc != nil {
		return m.ListProjectsFunc(ctx)
	}
	return nil, nil
}

// GetProject implements ProjectAPI.
func (m *MockClient) GetProject(ctx context.Context, projectID string) (*Project, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetProject", Args: []any{projectID}})
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, projectID)
	}
	return nil, nil
}

// ListEnvironments implements EnvironmentAPI.
func (m *MockClient) ListEnvironments(ctx context.Context, projectID string) ([]Environment, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListEnvironments", Args: []any{projectID}})
	if m.ListEnvironmentsFunc != nil {
		return m.ListEnvironmentsFunc(ctx, projectID)
	}
	return nil, nil
}

// GetEnvironment implements EnvironmentAPI.
func (m *MockClient) GetEnvironment(ctx context.Context, projectID, environmentID string) (*Environment, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetEnvironment", Args: []any{projectID, environmentID}})
	if m.GetEnvironmentFunc != nil {
		return m.GetEnvironmentFunc(ctx, projectID, environmentID)
	}
	return nil, nil
}

// Verify MockClient implements API at compile time.
var _ API = (*MockClient)(nil)
