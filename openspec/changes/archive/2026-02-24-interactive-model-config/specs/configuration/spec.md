# Model Configuration Specification

## Purpose

Define how go-ent loads and manages model configuration for different AI tool runtimes (Claude, OpenCode). This spec replaces the legacy fragmented configuration system with per-runtime configuration files.

## Requirements

### Requirement: Runtime-Specific Configuration Loading

**Level**: MUST

The system MUST load configuration from runtime-specific paths:
- Claude: `.claude/ent.yaml` in project root
- OpenCode: `.opencode/ent.yaml` in project root

#### Scenario: Load Claude configuration

**Given** a project with `.claude/ent.yaml` exists
**When** the system loads configuration for Claude runtime
**Then** it reads configuration from `.claude/ent.yaml`
**And** it returns a ClaudeConfig with specified model aliases

#### Scenario: Load OpenCode configuration

**Given** a project with `.opencode/ent.yaml` exists
**When** the system loads configuration for OpenCode runtime
**Then** it reads configuration from `.opencode/ent.yaml`
**And** it returns an OpenCodeConfig with specified model IDs

#### Scenario: Missing config file uses defaults

**Given** a project without runtime-specific config file
**When** the system loads configuration for any runtime
**Then** it returns a config with default model values
**And** no error is returned

### Requirement: Claude Short Alias Support

**Level**: MUST

The system MUST support short model aliases for Claude that auto-resolve to full model IDs:
- `sonnet` → `claude-sonnet-4-20250514`
- `opus` → `claude-opus-4-20250514`
- `haiku` → `claude-haiku-3-5-20241022`

#### Scenario: Use short alias in config

**Given** a `.claude/ent.yaml` with `sonnet` as model value
**When** the system resolves the model
**Then** it returns the full model ID `claude-sonnet-4-20250514`

#### Scenario: Unknown alias passes through

**Given** a `.claude/ent.yaml` with a full model ID as value
**When** the system resolves the model
**Then** it returns the value unchanged

### Requirement: OpenCode Model Discovery

**Level**: MUST

The system MUST discover available OpenCode models by running `opencode models` CLI and caching results.

#### Scenario: Discover models with cache miss

**Given** no cached models exist at `~/.cache/go-ent/opencode-models.json`
**When** the system discovers OpenCode models
**Then** it runs `opencode models` command
**And** it caches the result to `~/.cache/go-ent/opencode-models.json`
**And** it returns the discovered models

#### Scenario: Use cached models within expiry

**Given** cached models exist at `~/.cache/go-ent/opencode-models.json`
**And** the cache is less than 24 hours old
**When** the system discovers OpenCode models
**Then** it returns cached models without running CLI

#### Scenario: Refresh expired cache

**Given** cached models exist at `~/.cache/go-ent/opencode-models.json`
**And** the cache is more than 24 hours old
**When** the system discovers OpenCode models
**Then** it runs `opencode models` command
**And** it updates the cache with new results

#### Scenario: CLI unavailable uses defaults

**Given** `opencode` CLI is not in PATH
**When** the system discovers OpenCode models
**Then** it logs a warning
**And** it returns default model values

### Requirement: Default Model Values

**Level**: MUST

The system MUST provide sensible default model values when no configuration exists.

#### Scenario: Default Claude models

**Given** no `.claude/ent.yaml` exists
**When** the system loads Claude configuration
**Then** Sonnet is the default model
**And** all aliases (sonnet, opus, haiku) have known mappings

#### Scenario: Default OpenCode models

**Given** no `.opencode/ent.yaml` exists
**And** `opencode` CLI is unavailable
**When** the system loads OpenCode configuration
**Then** it returns hardcoded default model IDs

### Requirement: Legacy Configuration Removal

**Level**: MUST

The system MUST NOT support legacy configuration locations:
- Root `ent.yaml`
- `.go-ent/models.yaml`
- Global XDG config directory

#### Scenario: Ignore root ent.yaml

**Given** a root `ent.yaml` file exists
**When** the system loads configuration
**Then** the file is ignored
**And** runtime-specific config is used instead

#### Scenario: No .go-ent directory created

**Given** the system runs any command
**When** configuration is needed
**Then** no `.go-ent/` directory is created
