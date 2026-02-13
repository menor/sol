# Sol - Claude Code Instructions

## Project Overview

Sol is an agent-first CLI for Upsun. It prioritizes structured output (JSON) for AI agents while remaining usable by humans.

## Design Philosophy

### Agent-First, Composable Commands

Sol follows Unix philosophy: each command does one thing well. Agents compose multiple commands as needed.

**Do:**
- Keep commands focused and fast
- Return the minimal data needed for the command's purpose
- Let agents orchestrate multiple calls based on user intent

**Don't:**
- Fetch extra data "just in case" it might be useful
- Combine multiple concerns into one command
- Add flags to include tangentially related data

**Example:** User asks "list projects in Observability org"

Bad approach (monolithic):
```
sol project:list --with-org-details  # Fetches org data for every project
```

Good approach (composable):
```
sol organization:list    # Agent finds org ID for "Observability"
sol project:list         # Agent filters by organization_id
```

The composable approach is faster (fewer API calls when org details aren't needed) and more flexible (agent decides what data to fetch based on actual user intent).

### Output Formats

Sol supports two output formats optimized for different use cases:

- **TOON** (default): Token-Oriented Object Notation, ~50% fewer tokens than JSON
- **JSON**: Universal, machine-parseable format for human readability

Use `--output json` when humans need to read the output.

### Command Schema

The `--schema` flag outputs machine-readable documentation for any command:

```bash
sol project:list --schema          # Schema for project:list
sol variable:set --schema          # Works even for commands with required args
sol --schema                       # List all available commands
```

This enables agents to discover command capabilities programmatically.

## Upsun Documentation

When you need Upsun documentation, search these sites:
- https://docs.upsun.com/ - Official documentation
- https://devcenter.upsun.com/ - Developer center with guides and tutorials

### Key Environment Variables

| Variable | Available | Description |
|----------|-----------|-------------|
| `PLATFORM_APPLICATION` | Build + Runtime | Base64-encoded JSON app config |
| `PLATFORM_SOURCE_DIR` | Source operations only | Root directory of code repo during source operation |
| `PLATFORM_CACHE_DIR` | Build only | Directory for caching between builds |
| `PLATFORM_PROJECT` | Build + Runtime | Project ID |
| `PLATFORM_BRANCH` | Runtime only | Git branch name |
| `PLATFORM_ENVIRONMENT` | Runtime only | Environment name |

### Source Operations
- Source operations run in a container where home directory isn't writable
- Use `PLATFORM_SOURCE_DIR` to detect source operation context
- Search "source operations" on docs.upsun.com for details

### API
- Base URL: `https://api.upsun.com`
- Format: HAL+JSON with `_links` for pagination
- Auth: Bearer token in Authorization header

### Upsun CLI Commands Reference

Sol aims to implement key commands from the Upsun CLI. Use this as reference for command naming and functionality:

**Activity:** `activity:list`, `activity:get`, `activity:log`, `activity:cancel`
**App:** `app:list`, `app:config-get`, `app:config-validate`
**Auth:** `auth:browser-login`, `auth:api-token-login`, `auth:info`, `auth:logout`
**Backup:** `backup:list`, `backup:create`, `backup:restore`, `backup:delete`, `backup:get`
**Environment:** `environment:list`, `environment:info`, `environment:activate`, `environment:branch`, `environment:delete`, `environment:deploy`, `environment:merge`, `environment:push`, `environment:ssh`, `environment:url`, `environment:logs`, `environment:relationships`
**Integration:** `integration:list`, `integration:add`, `integration:get`, `integration:update`, `integration:delete`
**Organization:** `organization:list`, `organization:info`, `organization:create`, `organization:user:list`, `organization:user:add`
**Project:** `project:list`, `project:info`, `project:create`, `project:delete`, `project:get`
**Resources:** `resources:get`, `resources:set`, `resources:size:list`
**Route:** `route:list`, `route:get`
**Service:** `service:list`
**SSH:** `ssh` (alias for `environment:ssh`)
**Tunnel:** `tunnel:open`, `tunnel:list`, `tunnel:close`, `tunnel:info`
**User:** `user:list`, `user:add`, `user:get`, `user:update`, `user:delete`
**Variable:** `variable:list`, `variable:get`, `variable:create`, `variable:update`, `variable:delete`

Design principle: Each command does one thing. Agents compose multiple commands as needed (e.g., `organization:list` + `project:list` to get projects with org names).

## Code Patterns

### Error Handling
All errors should use `internal/errors.CLIError` with:
- A code (e.g., `AUTH_FAILED`, `API_ERROR`)
- A human-readable message
- Optional details map
- Optional hint for recovery

Exit codes are mapped from error codes via `CLIError.ExitCode()`.

### Output
Default output is TOON. The `--output` flag supports:
- `toon` (default) - Token-Oriented Object Notation, ~50% fewer tokens than JSON
- `json` - structured JSON for human readability

### Kong Command Pattern (IMPORTANT)

Sol uses Kong for CLI parsing. Commands are structs with a `Run` method:

```go
// Command struct with flags as fields
type MyCmd struct {
    Name  string `arg:"" help:"The name to use"`
    Force bool   `help:"Force operation" short:"f"`
}

// Run method receives the Context
func (c *MyCmd) Run(ctx *Context) error {
    // Access global flags via ctx.CLI
    projectID := ctx.ProjectID()  // Helper that checks flag then env

    // Create API client
    client, err := ctx.APIClient()
    if err != nil {
        return err
    }

    // Do work...

    // Output using configured format
    return ctx.Output(result)
}
```

**Never use `output.New("json")` directly** - use `ctx.Output()` which respects the `--output` flag.

### Service Layer Pattern (IMPORTANT)

Business logic should live in services, not CLI command handlers. CLI commands should only:
1. Get configuration from Context
2. Create the service with dependencies
3. Call the service method
4. Format and output the result

```go
// Good: CLI delegates to service
func (c *MyCmd) Run(ctx *Context) error {
    svc := mypackage.DefaultService()  // Production dependencies

    result, err := svc.DoSomething(ctx, options)
    if err != nil {
        return errors.NewAuthError(err.Error())
    }

    return ctx.Output(result)
}

// Bad: Business logic in CLI handler
func (c *MyCmd) Run(ctx *Context) error {
    // Don't do 100 lines of orchestration here
}
```

### Dependency Injection via Interfaces

Define interfaces for external dependencies to enable testing:

```go
// Define interface
type TokenStore interface {
    Save(token *StoredToken) error
    Load() (*StoredToken, error)
    Delete() error
    Exists() bool
}

// Production implementation
type KeyringStore struct{}
func (s *KeyringStore) Save(token *StoredToken) error { ... }

// Test implementation
type MemoryStore struct{ token *StoredToken }
func (s *MemoryStore) Save(token *StoredToken) error { ... }

// Service uses interface
type Service struct {
    store TokenStore  // Injected dependency
}

// Factory for production use
func DefaultService() *Service {
    return NewService(&KeyringStore{}, &SystemBrowser{})
}
```

### Progress Callbacks for UI Output

Don't hardcode `fmt.Fprintln(os.Stderr, ...)` in business logic. Use callbacks:

```go
// Define callback type
type ProgressFunc func(message string)

// Accept in options
type LoginOptions struct {
    OnProgress ProgressFunc
}

// Use in service
func (s *Service) Login(ctx context.Context, opts LoginOptions) error {
    progress := opts.OnProgress
    if progress == nil {
        progress = func(string) {}  // No-op default
    }
    progress("Starting...")
}

// CLI provides the implementation
progress := func(msg string) {
    if !cfg.Quiet {
        fmt.Fprintln(os.Stderr, msg)
    }
}
```

### Error Wrapping

Always use `%w` for error wrapping to preserve the error chain:

```go
// Good
return fmt.Errorf("load token: %w", err)

// Bad - loses error chain
return fmt.Errorf("load token: %v", err)
```

### Sentinel Errors (IMPORTANT)

Never compare error strings. Use sentinel errors and `errors.Is()`:

```go
// Define sentinel in package
var ErrAlreadyLoggedIn = errors.New("already logged in")

// Return sentinel from service
if alreadyLoggedIn {
    return nil, ErrAlreadyLoggedIn
}

// Check with errors.Is() in caller
if errors.Is(err, auth.ErrAlreadyLoggedIn) {
    // Handle specifically
}
```

**Never do this** - breaks if message changes:
```go
if err.Error() == "already logged in" {  // FRAGILE
```

### Context in Structs (Avoid)

Don't store `context.Context` in structs - it's an anti-pattern. Pass context to methods.

**Exception**: When implementing interfaces that don't accept context (like `oauth2.TokenSource`), document the limitation clearly.

### Package Structure

When a package grows to handle multiple concerns, consider splitting:

```
internal/
  auth/
    auth.go           # Public API, Service struct
    interfaces.go     # TokenStore, BrowserOpener interfaces
    keyring.go        # KeyringStore implementation
    store_memory.go   # MemoryStore for testing
    oauth.go          # OAuth protocol (PKCE, URLs, exchange)
    browser.go        # SystemBrowser implementation
    token.go          # Token sources, refresh logic
```

Keep related concerns together until the package becomes unwieldy.

### Security Patterns (IMPORTANT)

**Validate user input that goes into shell commands:**
```go
// Prevent SSH argument injection via --app flag
var validAppName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

if appName != "" && !validAppName.MatchString(appName) {
    return errors.NewValidationError("invalid app name")
}
```

**Escape user input in URL paths:**
```go
// Prevent path traversal via crafted IDs
path := fmt.Sprintf("/v1/projects/%s", url.PathEscape(projectID))
```

### Testing Patterns

**Context injection for testing:**

Tests create fresh CLI and Context structs with mock dependencies injected:

```go
func TestMyCmd_Success(t *testing.T) {
    mockClient := &api.MockClient{
        GetProjectFunc: func(ctx context.Context, id string) (*api.Project, error) {
            return &api.Project{ID: id, Title: "Test"}, nil
        },
    }

    // Create fresh CLI and Context - no globals to manage
    cli := &CLI{}  // Uses default toon output
    ctx := &Context{
        Context: context.Background(),
        CLI:     cli,
        apiClientFactory: func(ctx context.Context) (api.API, error) {
            return mockClient, nil
        },
        getEnvFunc: func(key string) string {
            if key == "PLATFORM_PROJECT" {
                return "proj123"
            }
            return ""
        },
    }

    cmd := &MyCmd{Name: "test"}
    err := cmd.Run(ctx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify mock was called correctly
    if len(mockClient.Calls) != 1 {
        t.Errorf("expected 1 call, got %d", len(mockClient.Calls))
    }
}
```

**MockClient pattern with call tracking:**
```go
type MockClient struct {
    ListProjectsFunc func(ctx context.Context) ([]ProjectRef, error)
    Calls           []MockCall  // Track calls for assertions
}

func (m *MockClient) ListProjects(ctx context.Context) ([]ProjectRef, error) {
    m.Calls = append(m.Calls, MockCall{Method: "ListProjects"})
    if m.ListProjectsFunc != nil {
        return m.ListProjectsFunc(ctx)
    }
    return nil, nil
}
```

### API Patterns

**Use `api.API` interface for testability:**
```go
// Commands depend on interface, not concrete *Client
var newAPIClient = func(ctx context.Context) (api.API, error) {
    return api.New(ctx)
}

// api.API composes ProjectAPI and EnvironmentAPI
type API interface {
    ProjectAPI      // ListProjects, GetProject
    EnvironmentAPI  // ListEnvironments, GetEnvironment
}
```

**HAL links can be object or array** - handle both:
```go
func (l HALLinks) GetHREF(name string) (string, bool) {
    // Try as single object {"href": "..."}
    var link HALLink
    if err := json.Unmarshal(raw, &link); err == nil && link.HREF != "" {
        return link.HREF, true
    }
    // Try as array [{"href": "..."}]
    var links []HALLink
    if err := json.Unmarshal(raw, &links); err == nil && len(links) > 0 {
        return links[0].HREF, true
    }
    return "", false
}
```

**Respect Retry-After header for rate limits:**
```go
if resp.StatusCode == http.StatusTooManyRequests {
    if retryAfter := parseRetryAfter(resp.Header.Get("Retry-After")); retryAfter > 0 {
        time.Sleep(retryAfter)
    }
}
```

### Go Idioms

**Deterministic output from maps (IMPORTANT):**

Go map iteration order is non-deterministic. When outputting data from maps, always convert to a sorted slice:

```go
// Bad - non-deterministic output order
return ctx.Output(deployment.Routes)  // map[string]Route

// Good - deterministic output order
type routeWithURL struct {
    URL string `json:"url"`
    api.Route
}
routes := make([]routeWithURL, 0, len(deployment.Routes))
for url, route := range deployment.Routes {
    routes = append(routes, routeWithURL{URL: url, Route: route})
}
sort.Slice(routes, func(i, j int) bool {
    return routes[i].URL < routes[j].URL
})
return ctx.Output(routes)
```

This applies to all command output, especially `--full` flags that return raw API data. Agents and tests expect consistent output ordering.

**Use local rand.Rand instead of global:**
```go
type Transport struct {
    rng     *rand.Rand
    rngOnce sync.Once
}

func (t *Transport) rand() *rand.Rand {
    t.rngOnce.Do(func() {
        t.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
    })
    return t.rng
}
```

**Use exec.CommandContext for cancellation:**
```go
// Good - respects context cancellation
cmd := exec.CommandContext(ctx, sshPath, args...)

// Bad - ignores context
cmd := exec.Command(sshPath, args...)
```

**Never ignore errors from io operations:**
```go
// Good
bodyBytes, err := io.ReadAll(req.Body)
if err != nil {
    return nil, fmt.Errorf("read body: %w", err)
}

// Bad - silently ignores errors
bodyBytes, _ := io.ReadAll(req.Body)
```

## Development

### Building
```bash
go build ./...
```

### Running
```bash
./sol --help
./sol version
```
