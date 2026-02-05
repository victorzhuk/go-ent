# Configuration Reference

Complete reference for go-ent configuration options.

## Configuration File Locations

go-ent searches for configuration in the following order:

1. **Project-level** (highest priority): `.go-ent/config.yaml`
2. **Global**: `~/.config/go-ent/config.yaml`
3. **Environment variable**: `$GO_ENT_CONFIG`

### Precedence Rules

- Project-level config overrides global config
- Environment variables override config file values
- CLI flags override all other sources

---

## Configuration File Format

Configuration uses YAML format with the following structure:

```yaml
# Runtime configuration
runtime: claude-code  # claude-code, opencode, cli

# Model configuration
models:
  fast: claude-haiku-4
  main: claude-sonnet-4.5
  heavy: claude-opus-4.5

# Budget limits
budget:
  daily: 100.00
  monthly: 500.00
  per_task: 10.00

# Agent configuration
agents:
  roles:
    architect:
      model: heavy
      skills: [go-arch, api-design]
    coder:
      model: main
      skills: [go-code, go-api, go-test]
    tester:
      model: fast
      skills: [go-test]

  # Agent-specific overrides
  overrides:
    ent:architect:
      temperature: 0.3
      max_tokens: 8000

# Skill configuration
skills:
  enabled: true
  directories:
    - plugins/go-ent/skills

  # Progressive loading
  progressive_load: true

  # Quality threshold (0-100)
  min_quality: 70

# OpenSpec configuration
openspec:
  enabled: true
  specs_dir: openspec/specs
  changes_dir: openspec/changes
  archive_dir: openspec/changes/archive

# MCP Server configuration
mcp:
  server:
    name: go-ent
    version: 0.3.0

  tools:
    enabled:
      - go_ent_spec_*
      - go_ent_registry_*
      - go_ent_workflow_*

  resources:
    enabled: true

# Logging
logging:
  level: info  # debug, info, warn, error
  format: json  # json, text
  output: stderr  # stdout, stderr, file

# Feature flags
features:
  execution_engine_v2: true
  skill_quality_scoring: true
  progressive_skill_load: true
```

---

## Configuration Options

### `runtime`

**Type:** `string`
**Default:** Auto-detected
**Values:** `claude-code`, `opencode`, `cli`

Specifies the runtime environment. Auto-detected based on:
- Claude Code: Presence of Claude Code MCP context
- OpenCode: Presence of `.opencode/` directory
- CLI: Fallback mode

```yaml
runtime: claude-code
```

---

### `models`

**Type:** `object`

Model configuration for different task categories.

#### `models.fast`

**Type:** `string`
**Default:** `claude-haiku-4`

Fast model for simple tasks (quick analysis, triage, simple tests).

```yaml
models:
  fast: claude-haiku-4
```

#### `models.main`

**Type:** `string`
**Default:** `claude-sonnet-4.5`

Main model for standard development tasks (coding, planning, debugging).

```yaml
models:
  main: claude-sonnet-4.5
```

#### `models.heavy`

**Type:** `string`
**Default:** `claude-opus-4.5`

Heavy model for complex tasks (architecture, deep analysis, critical reviews).

```yaml
models:
  heavy: claude-opus-4.5
```

**Available Models:**
- Anthropic: `claude-haiku-4`, `claude-sonnet-4.5`, `claude-opus-4.5`
- OpenAI: `gpt-4-turbo`, `gpt-4o`, `gpt-4o-mini`

---

### `budget`

**Type:** `object`

Budget limits to control API costs.

#### `budget.daily`

**Type:** `float`
**Default:** `100.00`

Maximum daily spend in USD.

```yaml
budget:
  daily: 50.00
```

#### `budget.monthly`

**Type:** `float`
**Default:** `500.00`

Maximum monthly spend in USD.

```yaml
budget:
  monthly: 500.00
```

#### `budget.per_task`

**Type:** `float`
**Default:** `10.00`

Maximum spend per individual task.

```yaml
budget:
  per_task: 5.00
```

---

### `agents`

**Type:** `object`

Agent system configuration.

#### `agents.roles`

**Type:** `map[string]object`

Agent role definitions with model and skill assignments.

```yaml
agents:
  roles:
    architect:
      model: heavy  # fast, main, heavy
      skills:
        - go-arch
        - api-design
        - security-core
      temperature: 0.3
      max_tokens: 8000

    coder:
      model: main
      skills:
        - go-code
        - go-api
        - go-db
      temperature: 0.7
      max_tokens: 4000
```

**Role Fields:**
- `model`: Model category (fast, main, heavy)
- `skills`: List of skill IDs to enable
- `temperature`: Model temperature (0.0-1.0)
- `max_tokens`: Maximum tokens per response

#### `agents.overrides`

**Type:** `map[string]object`

Agent-specific overrides by full agent ID (plugin:agent).

```yaml
agents:
  overrides:
    ent:architect:
      model: claude-opus-4.5  # Override model directly
      temperature: 0.2
```

---

### `skills`

**Type:** `object`

Skill system configuration.

#### `skills.enabled`

**Type:** `bool`
**Default:** `true`

Enable/disable skill system globally.

```yaml
skills:
  enabled: true
```

#### `skills.directories`

**Type:** `[]string`
**Default:** `["plugins/go-ent/skills"]`

Directories to search for skills.

```yaml
skills:
  directories:
    - plugins/go-ent/skills
    - custom/skills
```

#### `skills.progressive_load`

**Type:** `bool`
**Default:** `true`

Enable progressive skill loading (metadata → core → extended).

```yaml
skills:
  progressive_load: true
```

#### `skills.min_quality`

**Type:** `int`
**Default:** `70`
**Range:** `0-100`

Minimum quality score to enable a skill.

```yaml
skills:
  min_quality: 80
```

---

### `openspec`

**Type:** `object`

OpenSpec workflow configuration.

#### `openspec.enabled`

**Type:** `bool`
**Default:** `true`

Enable OpenSpec workflow features.

```yaml
openspec:
  enabled: true
```

#### `openspec.specs_dir`

**Type:** `string`
**Default:** `openspec/specs`

Directory for specification source of truth.

```yaml
openspec:
  specs_dir: openspec/specs
```

#### `openspec.changes_dir`

**Type:** `string`
**Default:** `openspec/changes`

Directory for active changes.

```yaml
openspec:
  changes_dir: openspec/changes
```

#### `openspec.archive_dir`

**Type:** `string`
**Default:** `openspec/changes/archive`

Directory for archived changes.

```yaml
openspec:
  archive_dir: openspec/changes/archive
```

---

### `mcp`

**Type:** `object`

MCP server configuration.

#### `mcp.server`

**Type:** `object`

Server metadata.

```yaml
mcp:
  server:
    name: go-ent
    version: 0.3.0
    description: Go-focused development toolkit
```

#### `mcp.tools.enabled`

**Type:** `[]string`

Enabled tool patterns (supports wildcards).

```yaml
mcp:
  tools:
    enabled:
      - go_ent_spec_*      # All spec tools
      - go_ent_registry_*  # All registry tools
      - go_ent_workflow_*  # All workflow tools
```

#### `mcp.resources.enabled`

**Type:** `bool`
**Default:** `true`

Enable MCP resources for agent/skill/command discovery.

```yaml
mcp:
  resources:
    enabled: true
```

---

### `logging`

**Type:** `object`

Logging configuration.

#### `logging.level`

**Type:** `string`
**Default:** `info`
**Values:** `debug`, `info`, `warn`, `error`

Log level.

```yaml
logging:
  level: debug
```

#### `logging.format`

**Type:** `string`
**Default:** `json`
**Values:** `json`, `text`

Log output format.

```yaml
logging:
  format: text
```

#### `logging.output`

**Type:** `string`
**Default:** `stderr`
**Values:** `stdout`, `stderr`, `file:<path>`

Log output destination.

```yaml
logging:
  output: file:/var/log/go-ent.log
```

---

### `features`

**Type:** `object`

Feature flags for experimental features.

```yaml
features:
  execution_engine_v2: true
  skill_quality_scoring: true
  progressive_skill_load: true
  parallel_task_execution: false
```

---

## Environment Variables

Environment variables override config file values.

| Variable | Config Path | Example |
|----------|-------------|---------|
| `GO_ENT_CONFIG` | - | `/custom/config.yaml` |
| `GO_ENT_RUNTIME` | `runtime` | `claude-code` |
| `GO_ENT_MODEL_FAST` | `models.fast` | `claude-haiku-4` |
| `GO_ENT_MODEL_MAIN` | `models.main` | `claude-sonnet-4.5` |
| `GO_ENT_MODEL_HEAVY` | `models.heavy` | `claude-opus-4.5` |
| `GO_ENT_BUDGET_DAILY` | `budget.daily` | `50.00` |
| `GO_ENT_BUDGET_MONTHLY` | `budget.monthly` | `500.00` |
| `GO_ENT_LOG_LEVEL` | `logging.level` | `debug` |
| `ANTHROPIC_API_KEY` | - | `sk-ant-...` |
| `OPENAI_API_KEY` | - | `sk-...` |

---

## Examples

### Minimal Configuration

```yaml
runtime: claude-code
models:
  main: claude-sonnet-4.5
```

### Development Configuration

```yaml
runtime: cli
models:
  fast: claude-haiku-4
  main: claude-sonnet-4.5
  heavy: claude-opus-4.5
budget:
  daily: 50.00
  monthly: 500.00
logging:
  level: debug
  format: text
features:
  execution_engine_v2: true
```

### Production Configuration

```yaml
runtime: opencode
models:
  fast: claude-haiku-4
  main: claude-sonnet-4.5
  heavy: claude-opus-4.5
budget:
  daily: 100.00
  monthly: 1000.00
  per_task: 10.00
agents:
  roles:
    architect:
      model: heavy
      skills: [go-arch, api-design, security-core]
    coder:
      model: main
      skills: [go-code, go-api, go-db, go-test]
    reviewer:
      model: heavy
      skills: [review-core, security-core]
logging:
  level: info
  format: json
  output: file:/var/log/go-ent.log
features:
  execution_engine_v2: true
  skill_quality_scoring: true
```

### Team Configuration

```yaml
runtime: claude-code
models:
  fast: claude-haiku-4
  main: claude-sonnet-4.5
  heavy: claude-opus-4.5

# Per-developer budget limits
budget:
  daily: 30.00
  monthly: 300.00
  per_task: 5.00

# Standard agent configuration
agents:
  roles:
    architect:
      model: heavy
      skills: [go-arch, api-design]
    coder:
      model: main
      skills: [go-code, go-api, go-db]
    tester:
      model: fast
      skills: [go-test]

# Quality threshold for skills
skills:
  min_quality: 85
  progressive_load: true

logging:
  level: info
  format: json
```

---

## Configuration Validation

Validate your configuration:

```bash
ent validate
```

Check specific configuration value:

```bash
ent config get models.main
```

List all configuration:

```bash
ent config list
```

---

## See Also

- [CLI Reference](./CLI_REFERENCE.md) - Command-line interface
- [MCP API Reference](./MCP_API.md) - MCP tool configuration
- [Agent System](./AGENTS_AND_SKILLS.md) - Agent and skill configuration
