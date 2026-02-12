package api

import (
	"context"
	"fmt"
	"net/url"
	"sort"
)

// Deployment represents the current deployment state of an environment.
// Retrieved from /projects/{id}/environments/{env}/deployments/current
type Deployment struct {
	ID       string              `json:"id"`
	Services map[string]Service  `json:"services"`
	Webapps  map[string]Webapp   `json:"webapps"`
	Workers  map[string]Worker   `json:"workers,omitempty"`
	Routes   map[string]Route    `json:"routes"`
	Links    HALLinks            `json:"_links"`
}

// Service represents a backing service (database, cache, search, etc.)
type Service struct {
	Type          string         `json:"type"`
	Size          string         `json:"size,omitempty"`
	Disk          int            `json:"disk,omitempty"`
	Configuration map[string]any `json:"configuration,omitempty"`
}

// ServiceSummary is a lean representation for list output.
type ServiceSummary struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Disk int    `json:"disk,omitempty"`
}

// Webapp represents an application container.
type Webapp struct {
	Type          string              `json:"type"`
	Size          string              `json:"size,omitempty"`
	Disk          int                 `json:"disk,omitempty"`
	Mounts        map[string]Mount    `json:"mounts,omitempty"`
	Relationships map[string][]string `json:"relationships,omitempty"`
}

// Worker represents a background worker process.
type Worker struct {
	Type   string              `json:"type"`
	Size   string              `json:"size,omitempty"`
	Mounts map[string]Mount    `json:"mounts,omitempty"`
}

// Mount represents a filesystem mount.
type Mount struct {
	Source     string `json:"source"`
	SourcePath string `json:"source_path,omitempty"`
}

// Route represents a URL route configuration.
type Route struct {
	Primary     bool                `json:"primary"`
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	OriginalURL string              `json:"original_url"`
	Upstream    string              `json:"upstream,omitempty"`
	To          string              `json:"to,omitempty"`
	Redirects   *RouteRedirects     `json:"redirects,omitempty"`
	TLS         *RouteTLS           `json:"tls,omitempty"`
}

// RouteRedirects contains redirect configuration.
type RouteRedirects struct {
	Expires string            `json:"expires,omitempty"`
	Paths   map[string]string `json:"paths,omitempty"`
}

// RouteTLS contains TLS configuration.
type RouteTLS struct {
	StrictTransportSecurity *StrictTransportSecurity `json:"strict_transport_security,omitempty"`
}

// StrictTransportSecurity contains HSTS settings.
type StrictTransportSecurity struct {
	Enabled           bool `json:"enabled"`
	IncludeSubdomains bool `json:"include_subdomains"`
	Preload           bool `json:"preload"`
}

// RouteURL is a lean representation for environment:url output.
type RouteURL struct {
	URL     string `json:"url"`
	Primary bool   `json:"primary"`
	Type    string `json:"type"`
}

// Relationship represents a connection between an app and a service.
type Relationship struct {
	App      string `json:"app"`
	Name     string `json:"name"`
	Service  string `json:"service"`
	Endpoint string `json:"endpoint,omitempty"`
}

// GetCurrentDeployment returns the current deployment for an environment.
// This provides services, webapps, workers, and routes.
func (c *Client) GetCurrentDeployment(ctx context.Context, projectID, envID string) (*Deployment, error) {
	var deployment Deployment
	path := fmt.Sprintf("/projects/%s/environments/%s/deployments/current",
		url.PathEscape(projectID),
		url.PathEscape(envID))
	if err := c.Get(ctx, path, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

// ListServices returns services from the current deployment.
func (c *Client) ListServices(ctx context.Context, projectID, envID string) ([]ServiceSummary, error) {
	deployment, err := c.GetCurrentDeployment(ctx, projectID, envID)
	if err != nil {
		return nil, err
	}

	services := make([]ServiceSummary, 0, len(deployment.Services))
	for name, svc := range deployment.Services {
		services = append(services, ServiceSummary{
			Name: name,
			Type: svc.Type,
			Disk: svc.Disk,
		})
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services, nil
}

// GetRoutes returns routes from the current deployment.
func (c *Client) GetRoutes(ctx context.Context, projectID, envID string) ([]RouteURL, error) {
	deployment, err := c.GetCurrentDeployment(ctx, projectID, envID)
	if err != nil {
		return nil, err
	}

	routes := make([]RouteURL, 0, len(deployment.Routes))
	for urlStr, route := range deployment.Routes {
		routes = append(routes, RouteURL{
			URL:     urlStr,
			Primary: route.Primary,
			Type:    route.Type,
		})
	}
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].URL < routes[j].URL
	})
	return routes, nil
}

// GetRelationships returns relationships for apps in the current deployment.
// If appName is specified, only relationships for that app are returned.
func (c *Client) GetRelationships(ctx context.Context, projectID, envID, appName string) ([]Relationship, error) {
	deployment, err := c.GetCurrentDeployment(ctx, projectID, envID)
	if err != nil {
		return nil, err
	}

	var relationships []Relationship

	// Collect relationships from webapps
	for app, webapp := range deployment.Webapps {
		if appName != "" && app != appName {
			continue
		}
		for name, endpoints := range webapp.Relationships {
			for _, endpoint := range endpoints {
				relationships = append(relationships, Relationship{
					App:      app,
					Name:     name,
					Service:  endpoint,
					Endpoint: endpoint,
				})
			}
		}
	}

	sort.Slice(relationships, func(i, j int) bool {
		if relationships[i].App != relationships[j].App {
			return relationships[i].App < relationships[j].App
		}
		return relationships[i].Name < relationships[j].Name
	})
	return relationships, nil
}
