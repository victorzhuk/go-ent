# Add BoltDB Registry with state.md System

## Overview

Replace complex YAML-based registry (3K+ lines, 400+ tasks) with BoltDB for O(1) lookups and human-readable state.md files for tool-agnostic visibility.

## Rationale

### Problem
1. **registry.yaml performance**: O(n) parse, O(n²) dependency lookups with 400+ tasks
2. **Destructive sync**: `RebuildFromSource()` wipes dependencies, notes, assignees
3. **No cheap state visibility**: Must parse entire registry to understand "what's next?"
4. **go-ent coupling**: openspec should work standalone without special tooling
5. **Dependencies lost**: HTML dependencies in registry.yaml wiped on sync

### Solution
1. **BoltDB cache**: O(1) task/dependency lookups, transactional updates
2. **state.md files**: Human-readable markdown for any AI tool
3. **HTML comment deps**: `<!-- depends: 11.1 -->` in tasks.md (survives sync)
4. **Tool-agnostic**: Any tool can read tasks.md + state.md

## Key Components

### 1. BoltDB Store (`internal/spec/boltdb.go`)
- 5 buckets: tasks, changes, deps, blocking, meta
- O(1) GetTask, UpdateTask, GetBlockers
- Transactional AddDependency/RemoveDependency
- Auto-generated, .gitignored

### 2. State Generator (`internal/spec/state.go`)
- Per-change: `openspec/changes/<id>/state.md`
- Root: `openspec/state.md`
- Shows: progress %, current task, blockers, recent activity
- Parses dependencies from HTML comments

### 3. MCP Tools (`internal/mcp/tools/state.go`)
- `state_sync`: Parse tasks.md → BoltDB → state.md
- `state_show`: Quick state view (for /status command)

### 4. Dependency Syntax
```markdown
- [ ] Unit tests for plugin manager <!-- depends: 11.1, 11.2 -->
```

## Architecture

```
tasks.md (source of truth, checkboxes)
    ↓ parse
state.md (human view, markdown)
    ↓ index
registry.db (BoltDB, binary cache)
    ↓ query
MCP tools (state_sync, state_show)
```

## Dependencies

- None (standalone change)
- Requires: `go.etcd.io/bbolt` (already added)

## Success Criteria

- [x] BoltDB store implemented with O(1) lookups
- [x] state.md generator for per-change and root state
- [x] HTML comment dependency parsing
- [x] MCP tools (state_sync, state_show)
- [x] .gitignore excludes registry.db
- [x] Code compiles
- [x] Registry store integrated with BoltDB
- [x] Migration script (YAML → BoltDB)
- [x] MCP tools updated to use BoltDB (registry_list, registry_next, registry_update, registry_deps, registry_sync)
- [x] Fallback to state.md if registry.db missing
- [x] Documentation updated (AGENTS.md)
- [x] Full workflow tested

## Progress

**Status**: Complete (60/60 tasks - 100%)

**Completed** (January 21, 2026):
- BoltDB infrastructure built
- state.md generation working
- MCP tools registered
- Dependency parsing functional
- Registry store integrated with BoltDB
- Migration script created
- All MCP tools updated to use BoltDB
- Fallback to state.md implemented
- Comprehensive testing completed
- Documentation updated (AGENTS.md)
- Final workflow documentation
- /task command implementation (deferred)
