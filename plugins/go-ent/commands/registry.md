---
description: Manage OpenSpec task registry
allowed-tools: mcp__plugin_go-ent_go-ent__registry_list, mcp__plugin_go-ent_go-ent__registry_next, mcp__plugin_go-ent_go-ent__registry_update, mcp__plugin_go-ent_go-ent__registry_deps, mcp__plugin_go-ent_go-ent__registry_sync, mcp__plugin_go-ent_go-ent__registry_init
---

# Task Registry Management

Input: `$ARGUMENTS` (subcommand: list|next|update|deps|sync|init)

## Available Subcommands

### `list [filters]`
Show all tasks with optional filters.

Examples:
- `list` - Show all tasks
- `list --status=pending` - Show only pending tasks
- `list --change=add-auth` - Show tasks for specific change
- `list --unblocked` - Show only unblocked tasks
- `list --priority=critical` - Show critical priority tasks

Use tool: `registry_list`

### `next [count]`
Get recommended next task(s) based on priority and dependencies.

Examples:
- `next` - Get single next task
- `next 3` - Get top 3 recommended tasks

Use tool: `registry_next`

### `update <task-id> <field=value>`
Update task status, priority, or assignee.

Examples:
- `update add-auth/1.1 status=completed`
- `update add-auth/2.1 status=in_progress assignee=claude`
- `update add-auth/3.1 priority=critical`

Use tool: `registry_update`

### `deps <task-id> <operation> [dep-id]`
Manage task dependencies (add/remove/show).

Examples:
- `deps add-auth/2.1 show` - Show dependency graph
- `deps add-auth/2.1 add add-auth/1.1` - Add dependency
- `deps add-auth/2.1 remove add-auth/1.1` - Remove dependency

Use tool: `registry_deps`

### `sync [--dry-run]`
Sync registry from tasks.md files in all changes.

Examples:
- `sync` - Rebuild registry from source
- `sync --dry-run` - Preview sync changes

Use tool: `registry_sync`

### `init`
Initialize empty registry.yaml file.

Use tool: `registry_init`

## Implementation

Parse `$ARGUMENTS` to determine subcommand and call appropriate MCP tool with path=".".

## Expected Output

Formatted JSON response from the tools, with clear indication of:
- Task IDs in format `change-id/task-num`
- Current status, priority, dependencies
- Blocking relationships
- Next recommended tasks with reasoning
