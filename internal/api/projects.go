package api

import (
	"context"
	"fmt"
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
func (c *Client) ListProjects(ctx context.Context) ([]ProjectRef, error) {
	// The /me/projects endpoint returns project references
	return Collect[ProjectRef](ctx, c, "/v1/me/projects")
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var project Project
	path := fmt.Sprintf("/v1/projects/%s", projectID)
	if err := c.Get(ctx, path, &project); err != nil {
		return nil, err
	}
	return &project, nil
}
