# Sol

Agent-first CLI for Upsun.

## What is Sol?

Sol is a minimal CLI that optimizes for code agents first, humans second. It provides:

- **Structured JSON output by default** - Machine-parseable responses
- **No interactive prompts** - Flags and stdin only
- **Predictable exit codes** - 0 success, 1 user error, 2 API error, 3 internal
- **Machine-readable errors** - Error codes and structured details
- **TOON output** - Token-efficient format for LLM agents

## Installation

```bash
go install github.com/menor/sol@latest
```

## Usage

### Authentication

```bash
# Log in (opens browser for OAuth)
sol auth:login

# Check authentication status
sol auth:info

# Log out (removes stored credentials)
sol auth:logout
```

For CI/automated environments, use an API token instead of interactive login:

```bash
export UPSUN_TOKEN=your-api-token
sol auth:info  # Shows authentication via environment variable
```

### Projects & Environments

```bash
# List projects
sol project:list

# Project details
sol project:info PROJECT_ID

# List environments
sol environment:list --project PROJECT_ID

# Environment details
sol environment:info main --project PROJECT_ID

# SSH into environment
sol ssh --project PROJECT_ID --environment main
```

### Activities

```bash
# List recent activities
sol activity:list --project PROJECT_ID

# Filter by state/type
sol activity:list --project PROJECT_ID --state complete --limit 5

# View activity log
sol activity:log ACTIVITY_ID --project PROJECT_ID
```

### Variables

```bash
# List project variables
sol variable:list --project PROJECT_ID

# List environment variables
sol variable:list --project PROJECT_ID --environment main

# Set a variable
sol variable:set MY_VAR "value" --project PROJECT_ID

# Set sensitive variable (value hidden)
sol variable:set SECRET "value" --project PROJECT_ID --sensitive

# Delete a variable
sol variable:delete MY_VAR --project PROJECT_ID
```

### Output Formats

```bash
# JSON (default) - machine-parseable
sol project:list --output json

# TOON - token-efficient for LLMs (coming soon)
sol project:list --output toon

# Text - human-readable (coming soon)
sol project:list --output text
```

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `json` | Output format: json, toon, text |
| `--project` | `-p` | | Project ID |
| `--environment` | `-e` | | Environment name |
| `--quiet` | `-q` | `false` | Suppress non-essential output |
| `--no-cache` | | `false` | Bypass cache for this request |
| `--debug` | | `false` | Show API request/response details |

## Configuration

Config file: `~/.sol/config.yaml`

```yaml
default_project: abc123
default_environment: main

output:
  format: json
  color: auto

cache:
  enabled: true
  ttl_seconds: 600
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `UPSUN_TOKEN` | API token (overrides keychain) |
| `UPSUN_PROJECT` | Default project ID |
| `UPSUN_ENVIRONMENT` | Default environment |

## Error Format

All errors return structured JSON:

```json
{
  "error": {
    "code": "AUTH_EXPIRED",
    "message": "Authentication expired and refresh failed",
    "details": {
      "hint": "Run 'sol auth:login' to re-authenticate"
    }
  }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad input, auth failed) |
| 2 | API error (server error, network issue) |
| 3 | Internal error (bug in CLI) |

## Development

```bash
# Build
go build -o sol .

# Run
./sol --help

# Test
go test ./...
```

## Why "Sol"?

Sol means "sun" in Catalan, Spanish, and Latin. It connects to Upsun and represents light/clarity - what this CLI aims to provide for agents interacting with the platform.

## License

Proprietary - Platform.sh
