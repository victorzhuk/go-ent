# Proposal: Add Agent MCP Tools

## Overview

Expose agent functionality via new MCP tools and rename existing tools from `go_ent_*` to `spec_*`, `agent_*`, `skill_*` prefixes for cleaner API.

## Rationale

### Problem
1. No MCP tools for agent execution (agent_execute, agent_status, etc.)
2. Tool names are verbose (`go_ent_spec_init` instead of `spec_init`)

### Solution
- **New tools**: `agent_execute`, `agent_status`, `agent_list`, `agent_delegate`, `skill_list`, `skill_info`, `runtime_list`, `runtime_status`
- **Rename tools**: `go_ent_spec_*` → `spec_*`, `go_ent_registry_*` → `registry_*`, etc.
- **Breaking change**: v3.0 version bump required

## Tool Renaming Map

| Old Name | New Name |
|----------|----------|
| `go_ent_spec_init` | `spec_init` |
| `go_ent_spec_list` | `spec_list` |
| `go_ent_registry_list` | `registry_list` |
| `go_ent_workflow_start` | `workflow_start` |
| `go_ent_loop_start` | `loop_start` |
| `go_ent_generate` | `project_generate` |

## New Tools

- `agent_execute` - Execute task with agent selection
- `agent_status` - Get execution status
- `agent_list` - List available agents
- `skill_list` - List available skills
- `runtime_list` - List runtimes

## Dependencies

- Requires: P0-P4 (execution engine)
- Breaking change: Requires v3.0 version bump

## Success Criteria

- [x] All tools renamed
- [x] New agent tools work
- [x] MCP server registers all tools
- [x] Claude Code plugin updated

## Status

**ARCHIVED** - All tasks completed and deployed.
**Archived:** 2026-01-10
