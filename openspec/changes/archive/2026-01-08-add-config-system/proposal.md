# Proposal: Add Configuration System

## Overview

Implement `.go-ent/config.yaml` configuration system for project-level and user-level settings. Enables customization of agent preferences, runtime selection, budget limits, and model mappings.

## Rationale

### Problem Statement

The current system lacks configuration for:
- **Agent preferences**: Which models to use for which roles
- **Runtime selection**: Preferred runtime (OpenCode vs Claude Code)
- **Budget controls**: Daily, monthly, per-task spending limits
- **Model mappings**: Friendly names → actual model IDs
- **Skill enablement**: Which skills are active

### Current State

`internal/generation/config.go` exists for template generation but doesn't cover:
- Agent configuration
- Runtime preferences
- Budget tracking
- Multi-level configuration (user vs project)

### Target State

New `.go-ent/config.yaml` with hierarchical configuration:
```yaml
version: "1.0"

agents:
  default: senior
  roles:
    architect: {model: opus, skills: [go-arch, go-api]}
    senior: {model: sonnet, skills: [go-code, go-db]}

runtime:
  preferred: opencode
  fallback: [claude-code, cli]

budget:
  daily: 10.0
  monthly: 200.0
  per_task: 1.0

models:
  opus: claude-opus-4-5-20251101
  sonnet: claude-sonnet-4-5-20251101
```

## Design Decisions

### 1. YAML Format

**Decision**: Use YAML for configuration

**Rationale**:
- Already using YAML (registry.yaml, workflow.yaml)
- Human-readable and editable
- Supports comments for documentation
- Go YAML library (`gopkg.in/yaml.v3`) is stable

### 2. Hierarchical Loading

**Decision**: Support project and user-level config

**Priority**: Project-level for v3.0, user-level for v3.1

**Loading order** (future):
1. `~/.go-ent/config.yaml` (user defaults)
2. `.go-ent/config.yaml` (project overrides)
3. Environment variables (runtime overrides)

**v3.0 scope**: Project-level only

### 3. Config Location

**Decision**: `.go-ent/config.yaml` in project root

**Rationale**:
- Aligns with `.go-ent/` convention
- Gitignored by default (contains budget preferences)
- Easy to find and edit

### 4. Environment Variable Override

**Decision**: Support env var overrides (e.g., `GOENT_BUDGET_DAILY`)

**Rationale**:
- CI/CD integration
- Temporary overrides without editing file
- Follows 12-factor app pattern

## Configuration Sections

### Agents Section
```yaml
agents:
  default: senior                    # Default role if not specified
  roles:
    architect:
      model: opus                    # Model mapping key
      skills: [go-arch, go-api]      # Enabled skills
      budget_limit: 5.0              # Per-execution limit (USD)
    senior:
      model: sonnet
      skills: [go-code, go-db, go-test]
      budget_limit: 2.0
  delegation:
    auto: true                       # Auto-delegate based on complexity
    approval_required: [architect, ops]  # Roles requiring approval
```

### Runtime Section
```yaml
runtime:
  preferred: opencode                # Primary runtime
  fallback: [claude-code, cli]       # Fallback order
  options:
    claude_code_path: ./dist/go-ent   # Custom paths if needed
```

### Budget Section
```yaml
budget:
  daily: 10.0                        # Daily spending limit (USD)
  monthly: 200.0                     # Monthly limit (USD)
  per_task: 1.0                      # Per-task limit (USD)
  tracking: true                     # Enable budget tracking
```

### Models Section
```yaml
models:
  opus: claude-opus-4-5-20251101
  sonnet: claude-sonnet-4-5-20251101
  haiku: claude-haiku-3-5-20241022
```

### Skills Section
```yaml
skills:
  enabled: [go-code, go-arch, go-db, go-test]
  custom_dir: .go-ent/skills          # Custom skill directory
```

## Integration Points

### With internal/spec/store.go
- Add `ConfigPath() string` method
- Add `LoadConfig() (*config.Config, error)` method
- Add `SaveConfig(cfg *config.Config) error` method

### With internal/generation/config.go
- Import new config system
- Extend with agent configuration
- Maintain backward compatibility

## Breaking Changes

None - this is a pure addition. Existing code continues to work without config file.

## Dependencies

- **Requires**: P0 (restructure), P1 (domain-types for AgentRole, Runtime enums)
- **Blocks**: P3 (agent-system), P4 (execution-engine)

## Success Criteria

- [x] Config loads from `.go-ent/config.yaml`
- [x] Defaults work when config file missing
- [x] Environment variables override config
- [x] Validation catches invalid configuration
- [x] Integration with existing spec store works

**Status:** COMPLETE → ARCHIVED
**Archived:** 2026-01-08
