package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Project represents a Upsun project.
type Project struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Region        string            `json:"region"`
	Organization  string            `json:"organization"`
	Vendor        string            `json:"vendor"`
	Repository    ProjectRepository `json:"repository"`
	DefaultBranch string            `json:"default_branch"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Links         HALLinks          `json:"_links"`

	Subscription *ProjectSubscription `json:"subscription,omitempty"`
}

// ProjectRepository contains git repository information.
type ProjectRepository struct {
	URL string `json:"url"`
}

// ProjectSubscription contains subscription/billing information.
type ProjectSubscription struct {
	LicenseURI    string `json:"license_uri,omitempty"`
	Plan          string `json:"plan,omitempty"`
	Environments  int    `json:"environments,omitempty"`
	Storage       int    `json:"storage,omitempty"`
	IncludedUsers int    `json:"included_users,omitempty"`
	UserLicenses  int    `json:"user_licenses,omitempty"`
	ManagementURI string `json:"subscription_management_uri,omitempty"`
	Restricted    bool   `json:"restricted,omitempty"`
	Suspended     bool   `json:"suspended,omitempty"`
}

// ProjectRef is a lightweight project reference returned in list operations.
type ProjectRef struct {
	ID             string    `json:"id"`
	Region         string    `json:"region"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	OrganizationID string    `json:"organization_id"`
	SubscriptionID string    `json:"subscription_id"`
	Vendor         string    `json:"vendor"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ListProjects returns all projects accessible to the authenticated user.
// This requires first getting the user ID, then fetching their project access.
func (c *Client) ListProjects(ctx context.Context) ([]ProjectRef, error) {
	// First get current user to get their ID
	var user struct {
		ID string `json:"id"`
	}
	if err := c.Get(ctx, "/users/me", &user); err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	// Get user's project access
	accessPath := fmt.Sprintf("/users/%s/extended-access?filter[resource_type]=project", url.PathEscape(user.ID))

	// The extended-access endpoint returns a different structure
	var accessResp struct {
		Items []struct {
			ResourceID     string `json:"resource_id"`
			ResourceType   string `json:"resource_type"`
			OrganizationID string `json:"organization_id"`
		} `json:"items"`
	}
	if err := c.Get(ctx, accessPath, &accessResp); err != nil {
		return nil, fmt.Errorf("get project access: %w", err)
	}

	// Convert to ProjectRef - we only have IDs from extended-access
	// The full project details would require additional API calls to /ref/projects
	var projects []ProjectRef
	for _, item := range accessResp.Items {
		if item.ResourceType == "project" {
			projects = append(projects, ProjectRef{
				ID:             item.ResourceID,
				OrganizationID: item.OrganizationID,
			})
		}
	}

	return projects, nil
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var project Project
	// Use /projects endpoint without /v1 prefix
	path := fmt.Sprintf("/projects/%s", url.PathEscape(projectID))
	if err := c.Get(ctx, path, &project); err != nil {
		return nil, err
	}
	return &project, nil
}
