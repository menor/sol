package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Variable represents a Upsun project or environment variable.
type Variable struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Value          string    `json:"value,omitempty"`
	IsSensitive    bool      `json:"is_sensitive"`
	IsEnabled      bool      `json:"is_enabled"`
	IsInheritable  bool      `json:"is_inheritable"`
	VisibleBuild   bool      `json:"visible_build"`
	VisibleRuntime bool      `json:"visible_runtime"`
	InheritedFrom  string    `json:"inherited_from,omitempty"`
	Project        string    `json:"project,omitempty"`
	Environment    string    `json:"environment,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Links          HALLinks  `json:"_links"`
}

// VariableInput is the input for creating or updating a variable.
// Note: is_enabled and is_inheritable are not accepted by the API for creation.
type VariableInput struct {
	Name           string `json:"name"`
	Value          string `json:"value"`
	IsSensitive    bool   `json:"is_sensitive"`
	VisibleBuild   bool   `json:"visible_build"`
	VisibleRuntime bool   `json:"visible_runtime"`
}

// ListProjectVariables returns all variables for a project.
func (c *Client) ListProjectVariables(ctx context.Context, projectID string) ([]Variable, error) {
	var variables []Variable
	path := fmt.Sprintf("/projects/%s/variables", url.PathEscape(projectID))
	if err := c.Get(ctx, path, &variables); err != nil {
		return nil, err
	}
	return variables, nil
}

// GetProjectVariable returns a single project variable by name.
func (c *Client) GetProjectVariable(ctx context.Context, projectID, name string) (*Variable, error) {
	var variable Variable
	path := fmt.Sprintf("/projects/%s/variables/%s",
		url.PathEscape(projectID),
		url.PathEscape(name))
	if err := c.Get(ctx, path, &variable); err != nil {
		return nil, err
	}
	return &variable, nil
}

// SetProjectVariable creates or updates a project variable.
func (c *Client) SetProjectVariable(ctx context.Context, projectID string, input *VariableInput) (*Variable, error) {
	var variable Variable
	path := fmt.Sprintf("/projects/%s/variables", url.PathEscape(projectID))

	// Try to create first
	err := c.Post(ctx, path, input, &variable)
	if err != nil {
		// If it already exists, update instead
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 409 {
			path := fmt.Sprintf("/projects/%s/variables/%s",
				url.PathEscape(projectID),
				url.PathEscape(input.Name))
			if err := c.Patch(ctx, path, input, &variable); err != nil {
				return nil, err
			}
			return &variable, nil
		}
		return nil, err
	}
	return &variable, nil
}

// DeleteProjectVariable deletes a project variable.
func (c *Client) DeleteProjectVariable(ctx context.Context, projectID, name string) error {
	path := fmt.Sprintf("/projects/%s/variables/%s",
		url.PathEscape(projectID),
		url.PathEscape(name))
	return c.Delete(ctx, path)
}

// ListEnvironmentVariables returns all variables for an environment.
func (c *Client) ListEnvironmentVariables(ctx context.Context, projectID, envID string) ([]Variable, error) {
	var variables []Variable
	path := fmt.Sprintf("/projects/%s/environments/%s/variables",
		url.PathEscape(projectID),
		url.PathEscape(envID))
	if err := c.Get(ctx, path, &variables); err != nil {
		return nil, err
	}
	return variables, nil
}

// GetEnvironmentVariable returns a single environment variable by name.
func (c *Client) GetEnvironmentVariable(ctx context.Context, projectID, envID, name string) (*Variable, error) {
	var variable Variable
	path := fmt.Sprintf("/projects/%s/environments/%s/variables/%s",
		url.PathEscape(projectID),
		url.PathEscape(envID),
		url.PathEscape(name))
	if err := c.Get(ctx, path, &variable); err != nil {
		return nil, err
	}
	return &variable, nil
}

// SetEnvironmentVariable creates or updates an environment variable.
func (c *Client) SetEnvironmentVariable(ctx context.Context, projectID, envID string, input *VariableInput) (*Variable, error) {
	var variable Variable
	path := fmt.Sprintf("/projects/%s/environments/%s/variables",
		url.PathEscape(projectID),
		url.PathEscape(envID))

	// Try to create first
	err := c.Post(ctx, path, input, &variable)
	if err != nil {
		// If it already exists, update instead
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 409 {
			path := fmt.Sprintf("/projects/%s/environments/%s/variables/%s",
				url.PathEscape(projectID),
				url.PathEscape(envID),
				url.PathEscape(input.Name))
			if err := c.Patch(ctx, path, input, &variable); err != nil {
				return nil, err
			}
			return &variable, nil
		}
		return nil, err
	}
	return &variable, nil
}

// DeleteEnvironmentVariable deletes an environment variable.
func (c *Client) DeleteEnvironmentVariable(ctx context.Context, projectID, envID, name string) error {
	path := fmt.Sprintf("/projects/%s/environments/%s/variables/%s",
		url.PathEscape(projectID),
		url.PathEscape(envID),
		url.PathEscape(name))
	return c.Delete(ctx, path)
}
