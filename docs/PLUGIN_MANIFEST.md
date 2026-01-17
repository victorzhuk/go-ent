# Plugin Manifest Reference

This document provides complete reference for the `plugin.yaml` manifest format used by go-ent plugins.

## Overview

The `plugin.yaml` file defines plugin metadata, version, and contained resources (skills, agents, rules). It's the primary configuration file for all go-ent plugins.

## Location

The manifest must be named `plugin.yaml` and placed at the root of the plugin directory:

```
my-plugin/
└── plugin.yaml    # Required
```

## Manifest Schema

### Root Structure

```yaml
name: string                    # Required
version: string                 # Required
description: string             # Required
author: string                 # Required
skills: []SkillRef             # Optional
agents: []AgentRef             # Optional
rules: []RuleRef              # Optional
min_version: string            # Optional
```

### Required Fields

#### name

**Type**: `string`
**Required**: Yes
**Constraints**:
- Maximum 100 characters
- Must be unique across all plugins
- Should use kebab-case (e.g., `data-validation`, `go-code`)

**Example**:
```yaml
name: my-plugin
```

#### version

**Type**: `string`
**Required**: Yes
**Constraints**:
- Must follow semantic versioning (MAJOR.MINOR.PATCH)
- Format: `^\d+\.\d+\.\d+$`

**Example**:
```yaml
version: 1.2.3
```

**Versioning Guidelines**:
- **MAJOR** (X.0.0): Breaking changes, incompatible API changes
- **MINOR** (1.X.0): New features, backward compatible
- **PATCH** (1.0.X): Bug fixes, backward compatible

#### description

**Type**: `string`
**Required**: Yes
**Constraints**:
- Maximum 500 characters
- Should clearly describe plugin's purpose

**Example**:
```yaml
description: Comprehensive data validation patterns and tools
```

#### author

**Type**: `string`
**Required**: Yes
**Constraints**:
- Must not be empty
- Recommended format: `Name <email>` or `Name`

**Example**:
```yaml
author: John Doe <john@example.com>
```

### Optional Fields

#### skills

**Type**: `[]SkillRef`
**Required**: No
**Description**: List of skills provided by this plugin

**SkillRef Structure**:
```yaml
skills:
  - name: string    # Required: Skill identifier
    path: string    # Required: Relative path to SKILL.md
```

**Constraints**:
- `name` must not be empty
- `path` must not be empty
- `path` is relative to plugin root
- Skill names must be unique within plugin

**Example**:
```yaml
skills:
  - name: validation
    path: skills/validation/SKILL.md

  - name: schema
    path: skills/schema/SKILL.md
```

#### agents

**Type**: `[]AgentRef`
**Required**: No
**Description**: List of agents provided by this plugin

**AgentRef Structure**:
```yaml
agents:
  - name: string    # Required: Agent identifier
    path: string    # Required: Relative path to agent.md
```

**Constraints**:
- `name` must not be empty
- `path` must not be empty
- `path` is relative to plugin root
- Agent names must be unique within plugin

**Example**:
```yaml
agents:
  - name: validator
    path: agents/validator.md

  - name: schema-builder
    path: agents/schema-builder.md
```

#### rules

**Type**: `[]RuleRef`
**Required**: No
**Description**: List of rules provided by this plugin

**RuleRef Structure**:
```yaml
rules:
  - name: string    # Required: Rule identifier
    path: string    # Required: Relative path to rule.yaml
```

**Constraints**:
- `name` must not be empty
- `path` must not be empty
- `path` is relative to plugin root
- Rule names must be unique within plugin

**Example**:
```yaml
rules:
  - name: enforce-validation
    path: rules/enforce-validation.yaml

  - name: reject-todos
    path: rules/reject-todos.yaml
```

#### min_version

**Type**: `string`
**Required**: No
**Description**: Minimum version of go-ent required for this plugin

**Constraints**:
- Must follow semantic versioning
- If set, plugin won't load on older go-ent versions

**Example**:
```yaml
min_version: 3.0.0
```

## Complete Example

```yaml
name: data-validation-plugin
version: 1.2.3
description: Comprehensive data validation patterns and tools for enterprise applications
author: Data Team <data@example.com>

skills:
  - name: validation
    path: skills/validation/SKILL.md

  - name: schema
    path: skills/schema/SKILL.md

agents:
  - name: validator
    path: agents/validator.md

  - name: schema-builder
    path: agents/schema-builder.md

rules:
  - name: enforce-validation
    path: rules/enforce-validation.yaml

  - name: reject-todos
    path: rules/reject-todos.yaml

min_version: 3.0.0
```

## Minimal Example

A minimal plugin with just required fields:

```yaml
name: simple-plugin
version: 1.0.0
description: A simple plugin example
author: Author Name
```

## Validation Rules

### Name Validation

- Must be non-empty
- Maximum 100 characters
- Should use kebab-case
- Must be unique across installed plugins
- No spaces or special characters except `-`

### Version Validation

- Must match semantic versioning: `^\d+\.\d+\.\d+$`
- Examples: `1.0.0`, `2.3.1`, `0.0.1`
- Invalid: `1`, `1.0`, `1.0.0-beta`, `v1.0.0`

### Description Validation

- Must be non-empty
- Maximum 500 characters
- Should describe plugin purpose clearly

### Author Validation

- Must be non-empty
- Recommended format: `Name <email>` or `Name`

### SkillRef Validation

For each skill in `skills` array:

- `name`: Non-empty, unique within plugin
- `path`: Non-empty, relative to plugin root
- File must exist at path
- Must point to a valid `SKILL.md` file

### AgentRef Validation

For each agent in `agents` array:

- `name`: Non-empty, unique within plugin
- `path`: Non-empty, relative to plugin root
- File must exist at path
- Must point to a valid agent markdown file

### RuleRef Validation

For each rule in `rules` array:

- `name`: Non-empty, unique within plugin
- `path`: Non-empty, relative to plugin root
- File must exist at path
- Must point to a valid YAML file

### MinVersion Validation

- If set, must match semantic versioning: `^\d+\.\d+\.\d+$`
- If set, plugin won't load on older go-ent versions

## Path Resolution

All paths in references (`skills[].path`, `agents[].path`, `rules[].path`) are relative to the plugin root directory.

**Example structure**:
```
my-plugin/
├── plugin.yaml
├── skills/
│   └── validation/
│       └── SKILL.md
├── agents/
│   └── validator.md
└── rules/
    └── enforce-validation.yaml
```

**Manifest**:
```yaml
name: my-plugin
version: 1.0.0
description: Example plugin
author: Author

skills:
  - name: validation
    path: skills/validation/SKILL.md    # Resolves to my-plugin/skills/validation/SKILL.md

agents:
  - name: validator
    path: agents/validator.md            # Resolves to my-plugin/agents/validator.md

rules:
  - name: enforce-validation
    path: rules/enforce-validation.yaml  # Resolves to my-plugin/rules/enforce-validation.yaml
```

## File Type Requirements

### Skill Files (SKILL.md)

Must be a markdown file with YAML frontmatter:

```markdown
---
name: skill-name
description: "Skill description. Auto-activates for: trigger1, trigger2."
---

# Skill Content

Detailed documentation...
```

**Requirements**:
- File extension: `.md`
- YAML frontmatter required
- `name` field in frontmatter must match `skills[].name` in manifest
- `description` field required for auto-activation

### Agent Files (agent.md)

Must be a markdown file with YAML frontmatter:

```markdown
---
name: agent-name
description: "Agent description"
model: claude-3-sonnet
color: "#3B82F6"
skills:
  - skill-name
tools:
  file-read: true
  file-write: true
---

# Agent Instructions

Agent documentation...
```

**Requirements**:
- File extension: `.md`
- YAML frontmatter required
- `name` field in frontmatter must match `agents[].name` in manifest
- `model` field required (`claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`)

### Rule Files (rule.yaml)

Must be a YAML file with rule definition:

```yaml
name: rule-name
description: "Rule description"
priority: high

trigger:
  event: file_modified
  pattern: "*.go"

conditions:
  - type: content_matches
    pattern: "TODO"

actions:
  - type: reject
    message: "TODOs not allowed"
```

**Requirements**:
- File extension: `.yaml` or `.yml`
- `name` field must match `rules[].name` in manifest
- `trigger` and `actions` fields required

## Error Messages

The validation system provides clear error messages:

```
Error: validate manifest: name cannot be empty
```

```
Error: validate manifest: name too long (max 100 characters)
```

```
Error: validate manifest: version cannot be empty
```

```
Error: validate manifest: description too long (max 500 characters)
```

```
Error: validate manifest: skill[0]: name cannot be empty
```

```
Error: validate manifest: agent[1]: path cannot be empty
```

## Common Patterns

### Skill-Only Plugin

```yaml
name: go-perf
version: 1.0.0
description: Performance profiling and optimization patterns
author: Performance Team

skills:
  - name: go-perf
    path: skills/go-perf/SKILL.md
```

### Agent-Only Plugin

```yaml
name: code-reviewer
version: 1.0.0
description: Automated code review agent
author: Quality Team

agents:
  - name: reviewer
    path: agents/reviewer.md
```

### Rule-Only Plugin

```yaml
name: code-quality-rules
version: 1.0.0
description: Automated code quality enforcement rules
author: Quality Team

rules:
  - name: reject-todos
    path: rules/reject-todos.yaml

  - name: enforce-comments
    path: rules/enforce-comments.yaml
```

### Multi-Resource Plugin

```yaml
name: enterprise-toolkit
version: 1.0.0
description: Comprehensive enterprise development toolkit
author: Enterprise Team

skills:
  - name: enterprise-arch
    path: skills/architecture/SKILL.md

  - name: enterprise-security
    path: skills/security/SKILL.md

agents:
  - name: architect
    path: agents/architect.md

  - name: security-auditor
    path: agents/security-auditor.md

rules:
  - name: enforce-security
    path: rules/enforce-security.yaml
```

## Migration Guide

### Upgrading Plugin Versions

When upgrading a plugin, update the `version` field:

```yaml
# Before
version: 1.0.0

# After (minor update, backward compatible)
version: 1.1.0

# After (major update, breaking changes)
version: 2.0.0
```

### Adding New Resources

When adding new skills, agents, or rules:

```yaml
# Before
skills:
  - name: validation
    path: skills/validation/SKILL.md

# After (added new skill)
skills:
  - name: validation
    path: skills/validation/SKILL.md

  - name: schema
    path: skills/schema/SKILL.md  # New skill
```

### Updating min_version

When plugin requires newer go-ent features:

```yaml
# Before
min_version: 2.0.0

# After (requires go-ent 3.0.0+)
min_version: 3.0.0
```

## Testing Manifests

### Manual Validation

Use go-ent to validate manifest:

```bash
go-ent plugin validate plugin.yaml
```

### Automated Validation

Validation is automatically performed during:
- Plugin installation
- Plugin loading at startup
- Plugin submission to marketplace

## Implementation Reference

The manifest parsing and validation is implemented in:

- **Source**: `internal/plugin/manifest.go`
- **Parser**: `ParseManifest(path string) (*Manifest, error)`
- **Validator**: `(*Manifest).Validate() error`
- **Type**: `Manifest struct`

Key methods:
- `GetSkillPath(name string) (string, error)` - Resolve skill path
- `GetAgentPath(name string) (string, error)` - Resolve agent path
- `GetRulePath(name string) (string, error)` - Resolve rule path

## See Also

- **Plugin Development Guide**: `PLUGIN_DEVELOPMENT.md`
- **Marketplace Usage**: `PLUGIN_MARKETPLACE.md`
- **E2E Tests**: `internal/plugin/e2e_test.go` (contains example manifests)
