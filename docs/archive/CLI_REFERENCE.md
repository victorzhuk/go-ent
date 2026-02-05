# CLI Reference

Complete reference for the `go-ent` command-line interface.

## Installation

### From Source

```bash
go install github.com/zhuk/go-ent@latest
```

### Using Make

```bash
make build
sudo make install
```

## Global Flags

All commands support these global flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `.go-ent/config.yaml` or `~/.config/go-ent/config.yaml` |
| `--verbose` | Enable verbose output | `false` |
| `--json` | Output in JSON format | `false` |
| `--help`, `-h` | Show help for command | - |

## Commands

### `ent version`

Display version information.

```bash
ent version
```

**Output:**
```
go-ent version v0.3.0
Built: 2026-01-28T12:00:00Z
Go: go1.21.5
```

---

### `ent init`

Initialize go-ent in the current project.

```bash
ent init [flags]
```

**Flags:**
- `--runtime <name>` - Target runtime (claude-code, opencode, cli)
- `--agent-config <path>` - Path to agent config template
- `--openspec` - Initialize OpenSpec structure
- `--force` - Overwrite existing configuration

**What it creates:**
- `.go-ent/config.yaml` - Project configuration
- `openspec/specs/` - Specification directory (if --openspec)
- `openspec/changes/` - Changes directory (if --openspec)
- `openspec/config.yaml` - OpenSpec configuration (if --openspec)

**Example:**
```bash
# Initialize for Claude Code
ent init --runtime claude-code --openspec

# Initialize with custom agent config
ent init --agent-config ./my-agents.yaml
```

---

### `ent validate`

Validate project configuration and structure.

```bash
ent validate [--strict]
```

**Validates:**
- Configuration file syntax
- Agent definitions
- Skill definitions
- OpenSpec structure (if present)
- Tool configurations

**Flags:**
- `--strict` - Enable strict validation (fail on warnings)

---

### `ent config`

Manage configuration settings.

#### `ent config get <key>`

Get configuration value.

```bash
ent config get models.main
# Output: claude-sonnet-4.5
```

#### `ent config set <key> <value>`

Set configuration value.

```bash
ent config set models.main claude-opus-4.5
```

#### `ent config list`

List all configuration values.

```bash
ent config list
```

**Output:**
```yaml
runtime: claude-code
models:
  fast: claude-haiku-4
  main: claude-sonnet-4.5
  heavy: claude-opus-4.5
budget:
  daily: 100.00
  monthly: 500.00
```

---

### `ent skill`

Manage skills.

#### `ent skill list`

List all available skills.

```bash
ent skill list [--category <category>]
```

**Flags:**
- `--category` - Filter by category (go, core, frontend, backend)

**Output:**
```
SKILL         CATEGORY  QUALITY  TRIGGERS
go-api        go        95       API, endpoint, handler
go-arch       go        90       architecture, DDD
go-code       go        92       implement, write
go-test       go        88       test, TDD
```

#### `ent skill info <skill-id>`

Show detailed skill information.

```bash
ent skill info go-api
```

**Output:**
```yaml
id: go-api
name: Go API Development
category: go
quality_score: 95
description: Expert in building production-ready REST and GraphQL APIs in Go
triggers:
  - API
  - endpoint
  - handler
  - REST
  - GraphQL
capabilities:
  - HTTP handler implementation
  - Middleware development
  - Request validation
  - ...
```

#### `ent skill validate <path>`

Validate skill definition.

```bash
ent skill validate plugins/go-ent/skills/go/go-api/SKILL.md
```

**Validates:**
- YAML frontmatter syntax
- Required fields (id, name, category, description)
- Quality score range (0-100)
- XML section structure

---

### `ent spec`

Manage OpenSpec specifications and changes.

#### `ent spec init`

Initialize OpenSpec structure.

```bash
ent spec init
```

Creates:
- `openspec/specs/`
- `openspec/changes/`
- `openspec/changes/archive/`
- `openspec/config.yaml`

#### `ent spec list [type]`

List specifications or changes.

```bash
# List all specs
ent spec list specs

# List active changes
ent spec list changes

# List archived changes
ent spec list archive

# List tasks in a change
ent spec list tasks <change-id>
```

#### `ent spec show <type> <id>`

Show specification or change details.

```bash
# Show spec
ent spec show spec api/handlers

# Show change
ent spec show change add-auth-middleware

# Show task
ent spec show task add-auth-middleware/1
```

#### `ent spec create <type> <id>`

Create new specification or change.

```bash
# Create spec
ent spec create spec api/handlers

# Create change
ent spec create change add-auth-middleware
```

#### `ent spec update <type> <id> [--field <key=value>]`

Update specification or change.

```bash
# Update change status
ent spec update change add-auth-middleware --field status=in_progress

# Update task status
ent spec update task add-auth-middleware/1 --field status=completed
```

#### `ent spec delete <type> <id>`

Delete specification or change.

```bash
ent spec delete change add-auth-middleware
```

#### `ent spec validate [path]`

Validate spec or change format.

```bash
# Validate all
ent spec validate

# Validate specific change
ent spec validate openspec/changes/add-auth-middleware/
```

#### `ent spec archive <change-id>`

Archive completed change.

```bash
ent spec archive add-auth-middleware
```

**What it does:**
1. Validates change is complete
2. Merges delta specs into main specs
3. Moves change to `changes/archive/`
4. Updates change index

---

### `ent model`

Manage model configuration.

#### `ent model list`

List available models and current configuration.

```bash
ent model list
```

**Output:**
```
CATEGORY  MODEL                  PROVIDER
fast      claude-haiku-4         anthropic
main      claude-sonnet-4.5      anthropic
heavy     claude-opus-4.5        anthropic

Available models:
- claude-haiku-4
- claude-sonnet-4.5
- claude-opus-4.5
- gpt-4-turbo (openai)
- gpt-4o (openai)
```

#### `ent model set <category> <model>`

Set model for a category.

```bash
# Set main model
ent model set main claude-opus-4.5

# Set fast model
ent model set fast claude-haiku-4
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Validation error |
| 4 | Not found |
| 5 | Permission denied |

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GO_ENT_CONFIG` | Path to config file | `.go-ent/config.yaml` |
| `GO_ENT_RUNTIME` | Runtime mode | Detected from context |
| `ANTHROPIC_API_KEY` | Anthropic API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |

---

## See Also

- [CLI Examples](./CLI_EXAMPLES.md) - Common usage patterns and workflows
- [Configuration Reference](./CONFIGURATION.md) - Detailed configuration options
- [MCP API Reference](./MCP_API.md) - MCP tool API documentation
- [Commands Reference](./COMMANDS_REFERENCE.md) - Slash command reference
