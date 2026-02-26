# Sol

Agent-optimized CLI for Upsun.

## What is Sol?

Sol is a minimal CLI optimized for AI code agents. It provides:

- **Lean output by default** - List commands return essential fields only (up to 99% smaller)
- **TOON format by default** - Token-efficient encoding (~50% smaller than JSON)
- **No interactive prompts** - Flags and stdin only
- **Predictable exit codes** - 0 success, 1 user error, 2 API error, 3 internal
- **Machine-readable errors** - Error codes and structured details
- **JSON output** - Use `-o json` when humans need to read it

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/menor/sol/releases/latest).

Available for macOS, Linux, and Windows (amd64 and arm64).

```bash
# Extract and install (macOS/Linux)
tar -xzf sol_*.tar.gz
sudo mv sol /usr/local/bin/
```

### Build from Source

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
# List projects (lean output: id, title, region)
sol project:list

# List projects with all fields
sol project:list --full

# Project details
sol project:info PROJECT_ID

# List environments (lean output: id, name, status, parent)
sol environment:list --project PROJECT_ID

# List environments with all fields
sol environment:list --project PROJECT_ID --full

# Environment details
sol environment:info main --project PROJECT_ID

# SSH into environment
sol ssh --project PROJECT_ID --environment main
```

### Activities

```bash
# List recent activities (lean output: id, type, state, created_at)
sol activity:list --project PROJECT_ID

# List activities with all fields
sol activity:list --project PROJECT_ID --full

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

### Environment Lifecycle

```bash
# Create a branch environment
sol environment:branch feature-x --project PROJECT_ID --parent main

# Create and wait for completion
sol environment:branch feature-x --project PROJECT_ID --wait

# Activate an inactive environment
sol environment:activate staging --project PROJECT_ID

# Deactivate an environment
sol environment:deactivate staging --project PROJECT_ID

# Delete an environment (must be deactivated first)
sol environment:delete old-feature --project PROJECT_ID
```

### Deployments

```bash
# Push code to Upsun (triggers deployment)
sol push --project PROJECT_ID

# Push to a specific branch
sol push --project PROJECT_ID --target staging

# Force push
sol push --project PROJECT_ID --force

# Redeploy an environment (runs post_deploy hook only)
sol redeploy --project PROJECT_ID --environment main

# Redeploy and wait for completion
sol redeploy --project PROJECT_ID --environment main --wait
```

### Output Formats

Sol optimizes output for agent context windows:

**Lean output (default for list commands):**
- `project:list` returns id, title, region (4KB vs 22KB full)
- `environment:list` returns id, name, status, parent (86B vs 28KB full)
- `activity:list` returns id, type, state, created_at (409B vs 4KB full)
- Use `--full` flag when you need all fields

**Format options:**
```bash
# TOON (default) - token-efficient for LLMs (~50% smaller than JSON)
sol project:list

# JSON - use when humans need to read the output
sol project:list --output json

# Full output with all fields
sol project:list --full
```

### Command Schema

Get machine-readable documentation for any command:

```bash
# Get schema for a specific command
sol project:list --schema

# Get schema in TOON format
sol variable:set --schema --output toon

# List all available commands
sol --schema
```

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `toon` | Output format: toon, json |
| `--project` | `-p` | | Project ID |
| `--environment` | `-e` | | Environment name |
| `--quiet` | `-q` | `false` | Suppress non-essential output |
| `--no-cache` | | `false` | Bypass cache for this request |
| `--debug` | | `false` | Show API request/response details |
| `--schema` | | `false` | Output command schema instead of running |

## Command-Specific Flags

| Flag | Short | Commands | Description |
|------|-------|----------|-------------|
| `--full` | `-f` | `project:list`, `environment:list`, `activity:list` | Include all fields in output |
| `--wait` | `-w` | `environment:branch`, `environment:activate`, `environment:deactivate`, `redeploy` | Wait for activity to complete |

## Configuration

Config file: `~/.sol/config.yaml`

```yaml
default_project: abc123
default_environment: main

output:
  format: toon
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

MIT License. See [LICENSE](LICENSE) for details.
