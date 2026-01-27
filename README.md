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
go install lab.plat.farm/menor/sol@latest
```

## Usage

```bash
# Authenticate
sol auth:login

# List projects (JSON output)
sol project:list

# List projects (token-efficient for LLMs)
sol project:list --output toon

# List projects (human-readable)
sol project:list --output text

# SSH into environment
sol ssh --project abc123 --environment main
```

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `json` | Output format: json, toon, text |
| `--project` | `-p` | | Project ID |
| `--environment` | `-e` | | Environment name |
| `--quiet` | `-q` | `false` | Suppress non-essential output |
| `--no-cache` | | `false` | Bypass cache for this request |

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
