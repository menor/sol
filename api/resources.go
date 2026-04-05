// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
)

// ResourceAllocation represents the resource allocation for services in an environment.
type ResourceAllocation struct {
	Webapps  map[string]ServiceResources `json:"webapps,omitempty"`
	Services map[string]ServiceResources `json:"services,omitempty"`
	Workers  map[string]ServiceResources `json:"workers,omitempty"`
}

// ServiceResources represents resource settings for a single service.
type ServiceResources struct {
	Size          string            `json:"size,omitempty"`
	Disk          int               `json:"disk,omitempty"`
	InstanceCount int               `json:"instance_count,omitempty"`
	Resources     *ResourceSettings `json:"resources,omitempty"`
}

// ResourceSettings contains detailed resource configuration.
type ResourceSettings struct {
	BaseMemory  int     `json:"base_memory,omitempty"`
	MemoryRatio float64 `json:"memory_ratio,omitempty"`
	ProfileSize string  `json:"profile_size,omitempty"`
}

// ResourceSummary is a lean representation of resource allocation.
type ResourceSummary struct {
	Apps     []ServiceResourceSummary `json:"apps,omitempty"`
	Services []ServiceResourceSummary `json:"services,omitempty"`
	Workers  []ServiceResourceSummary `json:"workers,omitempty"`
}

// ServiceResourceSummary is a lean representation of a service's resources.
type ServiceResourceSummary struct {
	Name          string `json:"name"`
	Size          string `json:"size,omitempty"`
	Disk          int    `json:"disk,omitempty"`
	InstanceCount int    `json:"instance_count,omitempty"`
}

// ToSummary converts ResourceAllocation to ResourceSummary with deterministic ordering.
func (r ResourceAllocation) ToSummary() ResourceSummary {
	summary := ResourceSummary{}

	for name, svc := range r.Webapps {
		summary.Apps = append(summary.Apps, ServiceResourceSummary{
			Name:          name,
			Size:          svc.Size,
			Disk:          svc.Disk,
			InstanceCount: svc.InstanceCount,
		})
	}

	for name, svc := range r.Services {
		summary.Services = append(summary.Services, ServiceResourceSummary{
			Name:          name,
			Size:          svc.Size,
			Disk:          svc.Disk,
			InstanceCount: svc.InstanceCount,
		})
	}

	for name, svc := range r.Workers {
		summary.Workers = append(summary.Workers, ServiceResourceSummary{
			Name:          name,
			Size:          svc.Size,
			Disk:          svc.Disk,
			InstanceCount: svc.InstanceCount,
		})
	}

	// Sort for deterministic output
	sort.Slice(summary.Apps, func(i, j int) bool {
		return summary.Apps[i].Name < summary.Apps[j].Name
	})
	sort.Slice(summary.Services, func(i, j int) bool {
		return summary.Services[i].Name < summary.Services[j].Name
	})
	sort.Slice(summary.Workers, func(i, j int) bool {
		return summary.Workers[i].Name < summary.Workers[j].Name
	})

	return summary
}

// GetResources returns the current resource allocation for an environment.
func (c *Client) GetResources(ctx context.Context, projectID, envID string) (*ResourceAllocation, error) {
	path := fmt.Sprintf("/projects/%s/environments/%s/deployments/next?verbose=true",
		url.PathEscape(projectID), url.PathEscape(envID))

	var result ResourceAllocation
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SetResourcesInput defines the input for updating resource allocation.
type SetResourcesInput struct {
	Webapps  map[string]ServiceResourcesInput `json:"webapps,omitempty"`
	Services map[string]ServiceResourcesInput `json:"services,omitempty"`
	Workers  map[string]ServiceResourcesInput `json:"workers,omitempty"`
}

// ServiceResourcesInput defines resource settings to update for a service.
type ServiceResourcesInput struct {
	Size          string `json:"size,omitempty"`
	Disk          int    `json:"disk,omitempty"`
	InstanceCount int    `json:"instance_count,omitempty"`
}

// SetResources updates the resource allocation for an environment.
// Returns an activity for the deployment that applies the changes.
func (c *Client) SetResources(ctx context.Context, projectID, envID string, input SetResourcesInput) (*Activity, error) {
	path := fmt.Sprintf("/projects/%s/environments/%s/deployments/next",
		url.PathEscape(projectID), url.PathEscape(envID))

	var activity Activity
	if err := c.Patch(ctx, path, input, &activity); err != nil {
		return nil, err
	}

	return &activity, nil
}
