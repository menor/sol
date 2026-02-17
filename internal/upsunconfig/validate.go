// Package upsunconfig provides validation for .upsun/config.yaml files.
package upsunconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the .upsun/config.yaml structure.
type Config struct {
	Applications map[string]Application `yaml:"applications"`
	Services     map[string]Service     `yaml:"services,omitempty"`
	Routes       map[string]Route       `yaml:"routes,omitempty"`
}

// Application represents an application definition.
type Application struct {
	Type          string                       `yaml:"type"`
	Source        *ApplicationSource           `yaml:"source,omitempty"`
	Dependencies  map[string]map[string]string `yaml:"dependencies,omitempty"`
	Variables     map[string]any               `yaml:"variables,omitempty"`
	Runtime       *Runtime                     `yaml:"runtime,omitempty"`
	Relationships map[string]string            `yaml:"relationships,omitempty"`
	Mounts        map[string]Mount             `yaml:"mounts,omitempty"`
	Build         *Build                       `yaml:"build,omitempty"`
	Hooks         *Hooks                       `yaml:"hooks,omitempty"`
	Web           *Web                         `yaml:"web,omitempty"`
	Crons         map[string]Cron              `yaml:"crons,omitempty"`
	Workers       map[string]Worker            `yaml:"workers,omitempty"`
}

// ApplicationSource defines the source root for an application.
type ApplicationSource struct {
	Root       string         `yaml:"root,omitempty"`
	Operations map[string]any `yaml:"operations,omitempty"`
}

// Runtime defines runtime configuration.
type Runtime struct {
	Extensions []string `yaml:"extensions,omitempty"`
}

// Mount defines a writable mount.
type Mount struct {
	Source     string `yaml:"source"`
	SourcePath string `yaml:"source_path"`
}

// Build defines build configuration.
type Build struct {
	Flavor string `yaml:"flavor,omitempty"`
}

// Hooks defines lifecycle hooks.
type Hooks struct {
	Build      string `yaml:"build,omitempty"`
	Deploy     string `yaml:"deploy,omitempty"`
	PostDeploy string `yaml:"post_deploy,omitempty"`
}

// Web defines web configuration.
type Web struct {
	Commands  *WebCommands        `yaml:"commands,omitempty"`
	Locations map[string]Location `yaml:"locations,omitempty"`
}

// WebCommands defines start commands.
type WebCommands struct {
	Start string `yaml:"start,omitempty"`
}

// Location defines a web location.
type Location struct {
	Root     string         `yaml:"root,omitempty"`
	Expires  string         `yaml:"expires,omitempty"`
	Passthru string         `yaml:"passthru,omitempty"`
	Allow    any            `yaml:"allow,omitempty"`
	Scripts  any            `yaml:"scripts,omitempty"`
	Headers  map[string]any `yaml:"headers,omitempty"`
	Rules    map[string]any `yaml:"rules,omitempty"`
}

// Cron defines a cron job.
type Cron struct {
	Spec string `yaml:"spec"`
	Cmd  string `yaml:"cmd"`
}

// Worker defines a worker process.
type Worker struct {
	Commands *WebCommands `yaml:"commands,omitempty"`
}

// Service represents a service definition.
type Service struct {
	Type string `yaml:"type"`
}

// Route represents a route definition.
type Route struct {
	Type     string `yaml:"type"`
	Upstream string `yaml:"upstream,omitempty"`
	To       string `yaml:"to,omitempty"`
}

// ValidationResult holds the result of config validation.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	FilePath string            `json:"file_path"`
	Apps     []AppValidation   `json:"applications,omitempty"`
	Services []string          `json:"services,omitempty"`
	Routes   int               `json:"routes,omitempty"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// AppValidation holds validation info for an application.
type AppValidation struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Type    string `json:"type"` // "syntax", "schema", "field"
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

// validName matches valid application and service names.
var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// validTypeFormat matches valid type format like "nodejs:18" or "php:8.2".
var validTypeFormat = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*:[a-zA-Z0-9.]+$`)

// knownServiceTypes lists known service types.
var knownServiceTypes = map[string]bool{
	"mysql": true, "mariadb": true, "oracle-mysql": true,
	"postgresql": true,
	"redis": true, "redis-persistent": true,
	"memcached":       true,
	"elasticsearch":   true,
	"opensearch":      true,
	"solr":            true,
	"rabbitmq":        true,
	"kafka":           true,
	"mongodb":         true,
	"influxdb":        true,
	"varnish":         true,
	"network-storage": true,
	"chrome-headless": true,
	"vault-kms":       true,
}

// DiscoverConfigFile finds .upsun/config.yaml in the given directory.
func DiscoverConfigFile(dir string) (string, error) {
	path := filepath.Join(dir, ".upsun", "config.yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found: .upsun/config.yaml")
		}
		return "", fmt.Errorf("check config file: %w", err)
	}
	return path, nil
}

// ValidateFile validates a .upsun/config.yaml file.
func ValidateFile(path string) (*ValidationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return Validate(data, path)
}

// Validate validates configuration data.
func Validate(data []byte, filePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		FilePath: filePath,
	}

	// Step 1: YAML syntax validation
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "syntax",
			Message: fmt.Sprintf("invalid YAML: %v", err),
		})
		return result, nil
	}

	// Step 2: Check required 'applications' key
	if _, ok := raw["applications"]; !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "schema",
			Message: "missing required 'applications' key",
			Path:    "applications",
		})
		return result, nil
	}

	// Step 3: Check for unknown top-level keys
	knownKeys := map[string]bool{
		"applications": true,
		"services":     true,
		"routes":       true,
	}
	for key := range raw {
		if !knownKeys[key] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown top-level key: %s", key))
		}
	}

	// Step 4: Parse the full config
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "schema",
			Message: fmt.Sprintf("failed to parse config: %v", err),
		})
		return result, nil
	}

	// Step 5: Validate applications
	validateApplications(config.Applications, result)

	// Step 6: Validate services
	validateServices(config.Services, result)

	// Step 7: Count routes
	result.Routes = len(config.Routes)

	return result, nil
}

// validateApplications validates application definitions.
func validateApplications(apps map[string]Application, result *ValidationResult) {
	if len(apps) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "schema",
			Message: "no applications defined",
			Path:    "applications",
		})
		return
	}

	// Collect app names for deterministic output
	names := make([]string, 0, len(apps))
	for name := range apps {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		app := apps[name]
		appValid := AppValidation{
			Name:  name,
			Type:  app.Type,
			Valid: true,
		}

		// Validate app name
		if !validName.MatchString(name) {
			appValid.Valid = false
			appValid.Reason = "invalid name format"
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "field",
				Message: fmt.Sprintf("invalid application name '%s': must start with alphanumeric, contain only alphanumeric, underscore, or hyphen", name),
				Path:    fmt.Sprintf("applications.%s", name),
			})
		}

		// Validate type is present
		if app.Type == "" {
			appValid.Valid = false
			appValid.Reason = "missing type"
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "field",
				Message: fmt.Sprintf("application '%s' missing required 'type' field", name),
				Path:    fmt.Sprintf("applications.%s.type", name),
			})
		} else if !validTypeFormat.MatchString(app.Type) {
			appValid.Valid = false
			appValid.Reason = "invalid type format"
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "field",
				Message: fmt.Sprintf("application '%s' has invalid type format '%s': expected runtime:version (e.g., nodejs:18)", name, app.Type),
				Path:    fmt.Sprintf("applications.%s.type", name),
			})
		}

		// Validate mounts
		for mountPath, mount := range app.Mounts {
			if mount.Source == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("mount '%s' in app '%s' missing 'source'", mountPath, name))
			} else if mount.Source != "instance" && mount.Source != "storage" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("mount '%s' in app '%s' has unknown source '%s': expected 'instance' or 'storage'", mountPath, name, mount.Source))
			}
		}

		result.Apps = append(result.Apps, appValid)
	}
}

// validateServices validates service definitions.
func validateServices(services map[string]Service, result *ValidationResult) {
	if len(services) == 0 {
		return
	}

	// Collect service names for deterministic output
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		svc := services[name]
		result.Services = append(result.Services, name)

		// Validate service name
		if !validName.MatchString(name) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "field",
				Message: fmt.Sprintf("invalid service name '%s': must start with alphanumeric, contain only alphanumeric, underscore, or hyphen", name),
				Path:    fmt.Sprintf("services.%s", name),
			})
		}

		// Validate type is present
		if svc.Type == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "field",
				Message: fmt.Sprintf("service '%s' missing required 'type' field", name),
				Path:    fmt.Sprintf("services.%s.type", name),
			})
		} else {
			// Check service type is known
			parts := strings.Split(svc.Type, ":")
			if len(parts) != 2 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("service '%s' has unusual type format '%s': expected type:version", name, svc.Type))
			} else if !knownServiceTypes[parts[0]] {
				result.Warnings = append(result.Warnings, fmt.Sprintf("service '%s' has unknown type '%s'", name, parts[0]))
			}
		}
	}
}
