package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Environment represents a Upsun environment (branch).
type Environment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	MachineName string    `json:"machine_name"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Parent      string    `json:"parent,omitempty"`
	Project     string    `json:"project"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Links       HALLinks  `json:"_links"`

	// HTTP access settings
	HTTPAccess *HTTPAccess `json:"http_access,omitempty"`
}

// EnvironmentSummary is a minimal environment representation for list operations.
// Contains essential fields for identifying environments and understanding hierarchy.
type EnvironmentSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Parent string `json:"parent,omitempty"`
}

// ToSummary converts an Environment to a lean EnvironmentSummary.
func (e *Environment) ToSummary() EnvironmentSummary {
	return EnvironmentSummary{
		ID:     e.ID,
		Name:   e.Name,
		Status: e.Status,
		Parent: e.Parent,
	}
}

// HTTPAccess contains HTTP access settings for an environment.
type HTTPAccess struct {
	IsEnabled   bool              `json:"is_enabled"`
	Addresses   []AccessAddress   `json:"addresses,omitempty"`
	BasicAuth   map[string]string `json:"basic_auth,omitempty"`
}

// AccessAddress represents an IP allowlist entry.
type AccessAddress struct {
	Address string `json:"address"`
}

// ListEnvironments returns all environments for a project.
func (c *Client) ListEnvironments(ctx context.Context, projectID string) ([]Environment, error) {
	var environments []Environment
	path := fmt.Sprintf("/projects/%s/environments", url.PathEscape(projectID))
	if err := c.Get(ctx, path, &environments); err != nil {
		return nil, err
	}
	return environments, nil
}

// GetEnvironment returns a single environment by ID.
func (c *Client) GetEnvironment(ctx context.Context, projectID, environmentID string) (*Environment, error) {
	var env Environment
	path := fmt.Sprintf("/projects/%s/environments/%s", url.PathEscape(projectID), url.PathEscape(environmentID))
	if err := c.Get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// embeddedActivitiesResponse wraps the API response that contains activities.
// POST endpoints like activate, deactivate, redeploy return this structure.
type embeddedActivitiesResponse struct {
	Status   string `json:"status"`
	Code     int    `json:"code"`
	Embedded struct {
		Activities []Activity `json:"activities"`
	} `json:"_embedded"`
}

// BranchInput is the request body for creating a branch.
type BranchInput struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}

// BranchEnvironment creates a new environment by branching from an existing one.
// Returns the activity triggered by the branch operation.
func (c *Client) BranchEnvironment(ctx context.Context, projectID, parentEnvID string, input *BranchInput) (*Activity, error) {
	var resp embeddedActivitiesResponse
	path := fmt.Sprintf("/projects/%s/environments/%s/branch",
		url.PathEscape(projectID),
		url.PathEscape(parentEnvID))
	if err := c.Post(ctx, path, input, &resp); err != nil {
		return nil, err
	}
	if len(resp.Embedded.Activities) == 0 {
		return nil, &APIError{Message: "no activity returned from API"}
	}
	return &resp.Embedded.Activities[0], nil
}

// ActivateEnvironment activates an inactive environment.
// Returns the activity triggered by the activation.
func (c *Client) ActivateEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	var resp embeddedActivitiesResponse
	path := fmt.Sprintf("/projects/%s/environments/%s/activate",
		url.PathEscape(projectID),
		url.PathEscape(environmentID))
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Embedded.Activities) == 0 {
		return nil, &APIError{Message: "no activity returned from API"}
	}
	return &resp.Embedded.Activities[0], nil
}

// DeactivateEnvironment deactivates an active environment.
// Returns the activity triggered by the deactivation.
func (c *Client) DeactivateEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	var resp embeddedActivitiesResponse
	path := fmt.Sprintf("/projects/%s/environments/%s/deactivate",
		url.PathEscape(projectID),
		url.PathEscape(environmentID))
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Embedded.Activities) == 0 {
		return nil, &APIError{Message: "no activity returned from API"}
	}
	return &resp.Embedded.Activities[0], nil
}

// DeleteEnvironment deletes an environment.
// The environment must be inactive (deactivated) before deletion.
func (c *Client) DeleteEnvironment(ctx context.Context, projectID, environmentID string) error {
	path := fmt.Sprintf("/projects/%s/environments/%s",
		url.PathEscape(projectID),
		url.PathEscape(environmentID))
	return c.Delete(ctx, path)
}

// RedeployEnvironment triggers a redeployment of an environment.
// This reuses the existing build and only runs the post_deploy hook.
// Returns the activity triggered by the redeployment.
func (c *Client) RedeployEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	var resp embeddedActivitiesResponse
	path := fmt.Sprintf("/projects/%s/environments/%s/redeploy",
		url.PathEscape(projectID),
		url.PathEscape(environmentID))
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Embedded.Activities) == 0 {
		return nil, &APIError{Message: "no activity returned from API"}
	}
	return &resp.Embedded.Activities[0], nil
}

// MergeEnvironment merges the current environment into its parent.
// The environment must have a parent to merge into.
// Returns the activity triggered by the merge operation.
func (c *Client) MergeEnvironment(ctx context.Context, projectID, environmentID string) (*Activity, error) {
	var resp embeddedActivitiesResponse
	path := fmt.Sprintf("/projects/%s/environments/%s/merge",
		url.PathEscape(projectID),
		url.PathEscape(environmentID))
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Embedded.Activities) == 0 {
		return nil, &APIError{Message: "no activity returned from API"}
	}
	return &resp.Embedded.Activities[0], nil
}

// SyncInput is the request body for synchronizing an environment.
type SyncInput struct {
	SynchronizeData      bool `json:"synchronize_data"`
	SynchronizeCode      bool `json:"synchronize_code"`
	SynchronizeResources bool `json:"synchronize_resources,omitempty"`
}

// SyncEnvironment synchronizes data and/or code from the parent environment.
// At least one of data, code, or resources must be true.
// Returns the activity triggered by the sync operation.
func (c *Client) SyncEnvironment(ctx context.Context, projectID, environmentID string, input *SyncInput) (*Activity, error) {
	var resp embeddedActivitiesResponse
	path := fmt.Sprintf("/projects/%s/environments/%s/synchronize",
		url.PathEscape(projectID),
		url.PathEscape(environmentID))
	if err := c.Post(ctx, path, input, &resp); err != nil {
		return nil, err
	}
	if len(resp.Embedded.Activities) == 0 {
		return nil, &APIError{Message: "no activity returned from API"}
	}
	return &resp.Embedded.Activities[0], nil
}
