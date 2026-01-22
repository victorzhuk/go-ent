# Proposal: Add Plugin System

## Overview

Create plugin manager and marketplace system for third-party skills, custom agents, and enterprise rules engine. Required for v3.0 per PRD.

## Rationale

### Problem
No way to extend go-ent with custom skills, agents, or rules without modifying core code.

### Solution
- **Plugin manager**: Install, load, validate plugins from marketplace or local
- **Plugin manifest**: YAML-based plugin definition (skills, agents, rules)
- **Marketplace client**: Search, download, publish plugins
- **Rules engine**: Execute enterprise coding standards as rules

## Key Components

1. `internal/plugin/manager.go` - Plugin lifecycle (install, load, enable, disable)
2. `internal/plugin/manifest.go` - Plugin manifest parsing and validation
3. `internal/marketplace/client.go` - Marketplace API integration
4. `internal/rules/engine.go` - Rules engine for enterprise standards

## Plugin Manifest Format

```yaml
name: my-skill-pack
version: 1.0.0
description: Custom skills for my organization
author: My Org

skills:
  - name: my-custom-skill
    path: skills/my-custom-skill/SKILL.md

agents:
  - name: my-custom-agent
    path: agents/my-custom-agent.md

rules:
  - name: require-review-for-security
    path: rules/security-review.yaml
```

## Dependencies

- Requires: P1 (domain-types for Skill interface), P2 (config), P3 (agent/skill registries)
- Can develop in parallel with P5-P6

## Success Criteria

- [ ] Plugin manager installs from marketplace
- [ ] Custom skills load from plugin
- [ ] Rules engine validates code
- [ ] MCP tools: plugin_list, plugin_install, plugin_search work
