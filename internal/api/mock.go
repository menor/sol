package api

import "context"

// MockClient is a mock implementation of the API interface for testing.
// Set the return values and errors before calling methods.
type MockClient struct {
	// Project methods
	ListProjectsFunc func(ctx context.Context) ([]ProjectRef, error)
	GetProjectFunc   func(ctx context.Context, projectID string) (*Project, error)

	// Environment methods
	ListEnvironmentsFunc       func(ctx context.Context, projectID string) ([]Environment, error)
	GetEnvironmentFunc         func(ctx context.Context, projectID, environmentID string) (*Environment, error)
	ActivateEnvironmentFunc    func(ctx context.Context, projectID, environmentID string) (*Activity, error)
	DeactivateEnvironmentFunc  func(ctx context.Context, projectID, environmentID string) (*Activity, error)
	DeleteEnvironmentFunc      func(ctx context.Context, projectID, environmentID string) error
	RedeployEnvironmentFunc    func(ctx context.Context, projectID, environmentID string) (*Activity, error)
	BranchEnvironmentFunc      func(ctx context.Context, projectID, parentEnvID string, input *BranchInput) (*Activity, error)

	// Activity methods
	ListActivitiesFunc func(ctx context.Context, projectID string, opts *ListActivitiesOptions) ([]Activity, error)
	GetActivityFunc    func(ctx context.Context, projectID, activityID string) (*Activity, error)
	GetActivityLogFunc func(ctx context.Context, projectID, activityID string) (string, error)

	// Variable methods - Project level
	ListProjectVariablesFunc  func(ctx context.Context, projectID string) ([]Variable, error)
	GetProjectVariableFunc    func(ctx context.Context, projectID, name string) (*Variable, error)
	SetProjectVariableFunc    func(ctx context.Context, projectID string, input *VariableInput) (*Variable, error)
	DeleteProjectVariableFunc func(ctx context.Context, projectID, name string) error

	// Variable methods - Environment level
	ListEnvironmentVariablesFunc  func(ctx context.Context, projectID, envID string) ([]Variable, error)
	GetEnvironmentVariableFunc    func(ctx context.Context, projectID, envID, name string) (*Variable, error)
	SetEnvironmentVariableFunc    func(ctx context.Context, projectID, envID string, input *VariableInput) (*Variable, error)
	DeleteEnvironmentVariableFunc func(ctx context.Context, projectID, envID, name string) error

	// Deployment methods
	GetCurrentDeploymentFunc func(ctx context.Context, projectID, envID string) (*Deployment, error)
	ListServicesFunc         func(ctx context.Context, projectID, envID string) ([]ServiceSummary, error)
	ListAppsFunc             func(ctx context.Context, projectID, envID string) ([]AppSummary, error)
	ListRoutesFunc           func(ctx context.Context, projectID, envID string) ([]RouteSummary, error)
	GetRoutesFunc            func(ctx context.Context, projectID, envID string) ([]RouteURL, error)
	GetRelationshipsFunc     func(ctx context.Context, projectID, envID, appName string) ([]Relationship, error)

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

// ActivateEnvironment implements EnvironmentAPI.
func (m *MockClient) ActivateEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ActivateEnvironment", Args: []any{projectID, environmentID}})
	if m.ActivateEnvironmentFunc != nil {
		return m.ActivateEnvironmentFunc(ctx, projectID, environmentID)
	}
	return nil, nil
}

// DeactivateEnvironment implements EnvironmentAPI.
func (m *MockClient) DeactivateEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	m.Calls = append(m.Calls, MockCall{Method: "DeactivateEnvironment", Args: []any{projectID, environmentID}})
	if m.DeactivateEnvironmentFunc != nil {
		return m.DeactivateEnvironmentFunc(ctx, projectID, environmentID)
	}
	return nil, nil
}

// DeleteEnvironment implements EnvironmentAPI.
func (m *MockClient) DeleteEnvironment(ctx context.Context, projectID, environmentID string) error {
	m.Calls = append(m.Calls, MockCall{Method: "DeleteEnvironment", Args: []any{projectID, environmentID}})
	if m.DeleteEnvironmentFunc != nil {
		return m.DeleteEnvironmentFunc(ctx, projectID, environmentID)
	}
	return nil
}

// RedeployEnvironment implements EnvironmentAPI.
func (m *MockClient) RedeployEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	m.Calls = append(m.Calls, MockCall{Method: "RedeployEnvironment", Args: []any{projectID, environmentID}})
	if m.RedeployEnvironmentFunc != nil {
		return m.RedeployEnvironmentFunc(ctx, projectID, environmentID)
	}
	return nil, nil
}

// BranchEnvironment implements EnvironmentAPI.
func (m *MockClient) BranchEnvironment(ctx context.Context, projectID, parentEnvID string, input *BranchInput) (*Activity, error) {
	m.Calls = append(m.Calls, MockCall{Method: "BranchEnvironment", Args: []any{projectID, parentEnvID, input}})
	if m.BranchEnvironmentFunc != nil {
		return m.BranchEnvironmentFunc(ctx, projectID, parentEnvID, input)
	}
	return nil, nil
}

// ListActivities implements ActivityAPI.
func (m *MockClient) ListActivities(ctx context.Context, projectID string, opts *ListActivitiesOptions) ([]Activity, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListActivities", Args: []any{projectID, opts}})
	if m.ListActivitiesFunc != nil {
		return m.ListActivitiesFunc(ctx, projectID, opts)
	}
	return nil, nil
}

// GetActivity implements ActivityAPI.
func (m *MockClient) GetActivity(ctx context.Context, projectID, activityID string) (*Activity, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetActivity", Args: []any{projectID, activityID}})
	if m.GetActivityFunc != nil {
		return m.GetActivityFunc(ctx, projectID, activityID)
	}
	return nil, nil
}

// GetActivityLog implements ActivityAPI.
func (m *MockClient) GetActivityLog(ctx context.Context, projectID, activityID string) (string, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetActivityLog", Args: []any{projectID, activityID}})
	if m.GetActivityLogFunc != nil {
		return m.GetActivityLogFunc(ctx, projectID, activityID)
	}
	return "", nil
}

// ListProjectVariables implements VariableAPI.
func (m *MockClient) ListProjectVariables(ctx context.Context, projectID string) ([]Variable, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListProjectVariables", Args: []any{projectID}})
	if m.ListProjectVariablesFunc != nil {
		return m.ListProjectVariablesFunc(ctx, projectID)
	}
	return nil, nil
}

// GetProjectVariable implements VariableAPI.
func (m *MockClient) GetProjectVariable(ctx context.Context, projectID, name string) (*Variable, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetProjectVariable", Args: []any{projectID, name}})
	if m.GetProjectVariableFunc != nil {
		return m.GetProjectVariableFunc(ctx, projectID, name)
	}
	return nil, nil
}

// SetProjectVariable implements VariableAPI.
func (m *MockClient) SetProjectVariable(ctx context.Context, projectID string, input *VariableInput) (*Variable, error) {
	m.Calls = append(m.Calls, MockCall{Method: "SetProjectVariable", Args: []any{projectID, input}})
	if m.SetProjectVariableFunc != nil {
		return m.SetProjectVariableFunc(ctx, projectID, input)
	}
	return nil, nil
}

// DeleteProjectVariable implements VariableAPI.
func (m *MockClient) DeleteProjectVariable(ctx context.Context, projectID, name string) error {
	m.Calls = append(m.Calls, MockCall{Method: "DeleteProjectVariable", Args: []any{projectID, name}})
	if m.DeleteProjectVariableFunc != nil {
		return m.DeleteProjectVariableFunc(ctx, projectID, name)
	}
	return nil
}

// ListEnvironmentVariables implements VariableAPI.
func (m *MockClient) ListEnvironmentVariables(ctx context.Context, projectID, envID string) ([]Variable, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListEnvironmentVariables", Args: []any{projectID, envID}})
	if m.ListEnvironmentVariablesFunc != nil {
		return m.ListEnvironmentVariablesFunc(ctx, projectID, envID)
	}
	return nil, nil
}

// GetEnvironmentVariable implements VariableAPI.
func (m *MockClient) GetEnvironmentVariable(ctx context.Context, projectID, envID, name string) (*Variable, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetEnvironmentVariable", Args: []any{projectID, envID, name}})
	if m.GetEnvironmentVariableFunc != nil {
		return m.GetEnvironmentVariableFunc(ctx, projectID, envID, name)
	}
	return nil, nil
}

// SetEnvironmentVariable implements VariableAPI.
func (m *MockClient) SetEnvironmentVariable(ctx context.Context, projectID, envID string, input *VariableInput) (*Variable, error) {
	m.Calls = append(m.Calls, MockCall{Method: "SetEnvironmentVariable", Args: []any{projectID, envID, input}})
	if m.SetEnvironmentVariableFunc != nil {
		return m.SetEnvironmentVariableFunc(ctx, projectID, envID, input)
	}
	return nil, nil
}

// DeleteEnvironmentVariable implements VariableAPI.
func (m *MockClient) DeleteEnvironmentVariable(ctx context.Context, projectID, envID, name string) error {
	m.Calls = append(m.Calls, MockCall{Method: "DeleteEnvironmentVariable", Args: []any{projectID, envID, name}})
	if m.DeleteEnvironmentVariableFunc != nil {
		return m.DeleteEnvironmentVariableFunc(ctx, projectID, envID, name)
	}
	return nil
}

// GetCurrentDeployment implements DeploymentAPI.
func (m *MockClient) GetCurrentDeployment(ctx context.Context, projectID, envID string) (*Deployment, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetCurrentDeployment", Args: []any{projectID, envID}})
	if m.GetCurrentDeploymentFunc != nil {
		return m.GetCurrentDeploymentFunc(ctx, projectID, envID)
	}
	return nil, nil
}

// ListServices implements DeploymentAPI.
func (m *MockClient) ListServices(ctx context.Context, projectID, envID string) ([]ServiceSummary, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListServices", Args: []any{projectID, envID}})
	if m.ListServicesFunc != nil {
		return m.ListServicesFunc(ctx, projectID, envID)
	}
	return nil, nil
}

// ListApps implements DeploymentAPI.
func (m *MockClient) ListApps(ctx context.Context, projectID, envID string) ([]AppSummary, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListApps", Args: []any{projectID, envID}})
	if m.ListAppsFunc != nil {
		return m.ListAppsFunc(ctx, projectID, envID)
	}
	return nil, nil
}

// ListRoutes implements DeploymentAPI.
func (m *MockClient) ListRoutes(ctx context.Context, projectID, envID string) ([]RouteSummary, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ListRoutes", Args: []any{projectID, envID}})
	if m.ListRoutesFunc != nil {
		return m.ListRoutesFunc(ctx, projectID, envID)
	}
	return nil, nil
}

// GetRoutes implements DeploymentAPI.
func (m *MockClient) GetRoutes(ctx context.Context, projectID, envID string) ([]RouteURL, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetRoutes", Args: []any{projectID, envID}})
	if m.GetRoutesFunc != nil {
		return m.GetRoutesFunc(ctx, projectID, envID)
	}
	return nil, nil
}

// GetRelationships implements DeploymentAPI.
func (m *MockClient) GetRelationships(ctx context.Context, projectID, envID, appName string) ([]Relationship, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GetRelationships", Args: []any{projectID, envID, appName}})
	if m.GetRelationshipsFunc != nil {
		return m.GetRelationshipsFunc(ctx, projectID, envID, appName)
	}
	return nil, nil
}

// Verify MockClient implements API at compile time.
var _ API = (*MockClient)(nil)
