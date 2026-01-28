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
