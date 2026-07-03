package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Activity represents a Upsun activity (deployment, backup, etc.).
type Activity struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	State        string     `json:"state"`
	Result       string     `json:"result"`
	Description  string     `json:"description"`
	Project      string     `json:"project"`
	Environments []string   `json:"environments"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Links        HALLinks   `json:"_links"`
}

// ActivitySummary is a minimal activity representation for list operations.
// Contains essential fields for identifying activities and their status.
type ActivitySummary struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

// ToSummary converts an Activity to a lean ActivitySummary.
func (a *Activity) ToSummary() ActivitySummary {
	return ActivitySummary{
		ID:        a.ID,
		Type:      a.Type,
		State:     a.State,
		CreatedAt: a.CreatedAt,
	}
}

// ActivityLog represents a log entry from an activity.
type ActivityLog struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ListActivitiesOptions configures activity listing.
type ListActivitiesOptions struct {
	Environment string // Filter by environment
	Type        string // Filter by activity type
	State       string // Filter by state (pending, in_progress, complete)
	Result      string // Filter by result (success, failure)
	Limit       int    // Maximum number of results
}

// ListActivities returns activities for a project.
func (c *Client) ListActivities(ctx context.Context, projectID string, opts *ListActivitiesOptions) ([]Activity, error) {
	path := fmt.Sprintf("/projects/%s/activities", url.PathEscape(projectID))

	// Build query string
	query := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			query.Set("count", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Type != "" {
			query.Set("type", opts.Type)
		}
		if opts.State != "" {
			query.Set("state", opts.State)
		}
		if opts.Result != "" {
			query.Set("result", opts.Result)
		}
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var activities []Activity
	if err := c.Get(ctx, path, &activities); err != nil {
		return nil, err
	}

	// Filter by environment if specified (API may not support this filter)
	if opts != nil && opts.Environment != "" {
		var filtered []Activity
		for _, a := range activities {
			for _, env := range a.Environments {
				if env == opts.Environment {
					filtered = append(filtered, a)
					break
				}
			}
		}
		return filtered, nil
	}

	return activities, nil
}

// GetActivity returns a single activity by ID.
func (c *Client) GetActivity(ctx context.Context, projectID, activityID string) (*Activity, error) {
	var activity Activity
	path := fmt.Sprintf("/projects/%s/activities/%s",
		url.PathEscape(projectID),
		url.PathEscape(activityID))
	if err := c.Get(ctx, path, &activity); err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetActivityLog returns the log output for an activity.
func (c *Client) GetActivityLog(ctx context.Context, projectID, activityID string) (string, error) {
	path := fmt.Sprintf("/projects/%s/activities/%s/log",
		url.PathEscape(projectID),
		url.PathEscape(activityID))

	// The log endpoint returns plain text, not JSON
	return c.GetText(ctx, path)
}
