# Sol - Claude Code Instructions

## Project Overview

Sol is an agent-first CLI for Upsun. It prioritizes structured output (JSON) for AI agents while remaining usable by humans.

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

## Code Patterns

### Error Handling
All errors should use `internal/errors.CLIError` with:
- A code (e.g., `AUTH_FAILED`, `API_ERROR`)
- A human-readable message
- Optional details map
- Optional hint for recovery

Exit codes are mapped from error codes via `CLIError.ExitCode()`.

### Output
Default output is JSON. The `--output` flag supports:
- `json` (default) - structured JSON
- `toon` - token-optimized notation (not yet implemented)
- `text` - human-readable text

### Config Pattern (IMPORTANT)

All commands should use `cli.FromCommand(cmd)` to extract configuration:

```go
func runMyCommand(cmd *cobra.Command, args []string) error {
    cfg, err := cli.FromCommand(cmd)
    if err != nil {
        return err
    }

    // Use cfg.Formatter() for output - respects --output flag
    return cfg.Formatter().Write(result)
}
```

**Never use `output.New("json")` directly** - this ignores the user's `--output` flag.

### Service Layer Pattern (IMPORTANT)

Business logic should live in services, not CLI command handlers. CLI commands should only:
1. Parse flags and arguments
2. Create the service with dependencies
3. Call the service method
4. Format and output the result

```go
// Good: CLI delegates to service
func runMyCommand(cmd *cobra.Command, args []string) error {
    cfg, _ := cli.FromCommand(cmd)
    svc := mypackage.DefaultService()  // Production dependencies

    result, err := svc.DoSomething(ctx, options)
    if err != nil {
        return errors.NewAuthError(err.Error())
    }

    return cfg.Formatter().Write(result)
}

// Bad: Business logic in CLI handler
func runMyCommand(cmd *cobra.Command, args []string) error {
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
