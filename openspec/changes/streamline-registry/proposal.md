# Streamline Registry

## Summary

Implement hybrid sync architecture with markdown as source of truth and BoltDB as runtime cache. File watcher (fsnotify) provides near real-time sync for fast MCP access.

## Problem

The current registry system:
- Re-parses markdown files on every MCP query (slow)
- No caching layer for fast lookups
- No real-time sync when files change
- Limited MCP tools for registry access

## Solution

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  LAYER 1: SOURCE OF TRUTH (Markdown)                       │
│  openspec/changes/{id}/tasks.md                            │
│  openspec/changes/{id}/proposal.md                         │
└──────────────────────────┬──────────────────────────────────┘
                           │ Human/Agent edits
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  LAYER 2: SYNC ENGINE (fsnotify + Debounce)                │
│  Watch: openspec/**/*.md                                   │
│  Debounce: 100ms                                           │
└──────────────────────────┬──────────────────────────────────┘
                           │ Parse & Index
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  LAYER 3: BOLTDB CACHE (openspec/.state.db)                │
│  Buckets: changes, tasks, deps, blockers, runtime, meta    │
│  Design: O(1) lookups, no parsing at query time            │
└──────────────────────────┬──────────────────────────────────┘
                           │ Read-only queries
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  LAYER 4: MCP API (Read-Only)                              │
│  Query: registry_list_changes, registry_get_change, etc.   │
│  Action: registry_mark_done, registry_start_task, etc.     │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| File watcher | fsnotify | Real-time sync, industry standard |
| Debounce | 100ms | Batch rapid editor events |
| In-progress state | Runtime only | Transient, simpler recovery |
| Migration | None | Start fresh, no legacy burden |
| Sync direction | One-way (md → db) | Clear source of truth |

### Data Flow

```
Human edits tasks.md ──► File watcher detects ──► Debounce (100ms)
                                                        │
                                                        ▼
                              Parse markdown ──► Update BoltDB
                                                        │
                              MCP query ◄───────────────┘
                              (fast O(1) lookup)
```

## Affected Systems

- `internal/spec/boltdb.go` - New: BoltDB store with buckets
- `internal/spec/watcher.go` - New: File watcher with debounce
- `internal/spec/state.go` - Update: Remove old state tracking
- `internal/mcp/tools/registry_*.go` - New: MCP registry tools
- `internal/mcp/server/server.go` - Update: Initialize BoltDB + watcher
- `.gitignore` - Add `openspec/.state.db`
- `go.mod` - Add `github.com/fsnotify/fsnotify v1.7.0`

## BoltDB Schema

```go
// changes/ bucket - Change metadata
// tasks/ bucket - Task definitions
// deps/ bucket - Dependency graph (forward + reverse)
// blockers/ bucket - Pre-computed blocked tasks
// runtime/ bucket - In-progress state (ephemeral)
// meta/ bucket - Sync state (mtimes, version)
```

## MCP Tools

### Query Tools (Read BoltDB)
- `registry_list_changes` - List all changes
- `registry_get_change` - Get change details + tasks
- `registry_list_tasks` - Filtered, sorted tasks
- `registry_next_task` - Next unblocked task
- `registry_deps` - Dependency graph
- `registry_search` - Full-text search
- `registry_status` - Aggregated stats

### Action Tools (Write Markdown)
- `registry_mark_done` - Check task in tasks.md
- `registry_start_task` - Set in-progress (runtime only)
- `registry_sync` - Force full rebuild

## Error Recovery

| Error | Strategy |
|-------|----------|
| Parse error | Log, skip file, retry on next change |
| BoltDB corrupt | Delete, rebuild from markdown |
| Watcher fail | Fallback to explicit sync only |
| Sync conflict | Last-write-wins (mtime authority) |

## Success Criteria

- [ ] BoltDB store implemented with O(1) lookups
- [ ] File watcher with debounce (100ms)
- [ ] Incremental sync on file change
- [ ] Full rebuild on startup
- [ ] MCP query tools (7 tools)
- [ ] MCP action tools (3 tools)
- [ ] Tests passing
- [ ] Documentation updated

## Effort Estimate

**~12 hours** across 4 phases:
- Phase 1: BoltDB layer (4h)
- Phase 2: File watcher (3h)
- Phase 3: MCP tools (3h)
- Phase 4: Integration (2h)
