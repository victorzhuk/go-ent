# Streamline Registry Architecture v2

## Overview

Hybrid sync architecture with markdown as source of truth and BoltDB as runtime cache. File watcher (fsnotify) provides near real-time sync.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    ARCHITECTURE: Streamlined Registry v2                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  LAYER 1: SOURCE OF TRUTH (Markdown)                                    │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  openspec/                                                      │   │
│  │  ├── changes/{id}/proposal.md  - Change metadata                │   │
│  │  ├── changes/{id}/tasks.md     - Tasks with checkboxes          │   │
│  │  └── specs/{cap}/spec.md       - Capability specs               │   │
│  │                                                                 │   │
│  │  Dependency syntax: <!-- depends: 1.2 -->                       │   │
│  │  Status: [ ] pending, [x] completed                             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              ▲                                          │
│                              │ Human/Agent edits                        │
│                              │                                          │
│  LAYER 2: SYNC ENGINE         │                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  File Watcher (fsnotify) + Debounce (100ms)                     │   │
│  │                                                                 │   │
│  │  Watch: openspec/**/*.md                                       │   │
│  │                                                                 │   │
│  │  On event:                                                      │   │
│  │  1. Debounce rapid changes (editor save = multiple events)     │   │
│  │  2. Parse changed file → extract tasks/spec                    │   │
│  │  3. Update BoltDB in transaction                               │   │
│  │  4. Recompute dependency graph for affected change             │   │
│  │                                                                 │   │
│  │  On startup: Full rebuild (parse all → BoltDB)                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  LAYER 3: BOLTDB CACHE                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  openspec/.state.db                                             │   │
│  │                                                                 │   │
│  │  Buckets:                                                       │   │
│  │  ├── changes/      - Change metadata (key: change_id)          │   │
│  │  ├── tasks/        - Tasks (key: change_id/task_num)           │   │
│  │  ├── deps/         - Dependency graph (forward + reverse)      │   │
│  │  ├── blockers/     - Pre-computed blocked tasks                │   │
│  │  ├── runtime/      - In-progress state (ephemeral)             │   │
│  │  └── meta/         - Sync state (mtimes, version)              │   │
│  │                                                                 │   │
│  │  Design: O(1) lookups, no parsing at query time                │   │
│  │  Recovery: If corrupt → delete, rebuild from markdown          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              ▲                                          │
│                              │ Read-only queries                        │
│                              │                                          │
│  LAYER 4: MCP API (Read-Only)                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Query Tools:                                                   │   │
│  │  • registry_list_changes    - List all changes                  │   │
│  │  • registry_get_change      - Get change + tasks                │   │
│  │  • registry_list_tasks      - Filtered, sorted tasks            │   │
│  │  • registry_next_task       - Next unblocked task               │   │
│  │  • registry_deps            - Dependency graph                  │   │
│  │  • registry_search          - Full-text search                  │   │
│  │  • registry_status          - Aggregated stats                  │   │
│  │                                                                 │   │
│  │  Action Tools (write markdown, trigger sync):                   │   │
│  │  • registry_mark_done       - Check task in tasks.md           │   │
│  │  • registry_start_task      - Set in-progress (runtime only)   │   │
│  │  • registry_sync            - Force full rebuild               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Key Decisions

### 1. File Watcher (fsnotify)
- **Library**: github.com/fsnotify/fsnotify v1.7.0
- **Debounce**: 100ms to batch rapid events
- **Watch paths**: openspec/**/*.md
- **Skip patterns**: *.tmp, *~, .git

### 2. In-Progress State (Runtime Only)
- Stored in BoltDB `runtime/` bucket
- **Not persisted** across restarts
- If Claude restarts, agent re-checks current task
- Simpler, fewer edge cases

### 3. No Migration
- No support for old registry.yaml format
- Start fresh with new BoltDB
- Full rebuild on first startup

### 4. Sync Direction
- **One-way**: Markdown → BoltDB
- Markdown is always source of truth
- BoltDB is disposable cache

## Data Model

### BoltDB Buckets

```go
// changes/ bucket
// Key: change_id (string)
// Value: ChangeMetadata (JSON)
type ChangeMetadata struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    TaskCount   int       `json:"task_count"`
    Completed   int       `json:"completed"`
    InProgress  int       `json:"in_progress"`
    Blocked     int       `json:"blocked"`
}

// tasks/ bucket
// Key: change_id/task_num (string)
// Value: Task (JSON)
type Task struct {
    ID          string     `json:"id"`           // change_id/task_num
    ChangeID    string     `json:"change_id"`
    TaskNum     string     `json:"task_num"`
    Content     string     `json:"content"`
    Status      string     `json:"status"`       // pending, completed
    Priority    string     `json:"priority"`
    DependsOn   []string   `json:"depends_on"`   // task_nums
    SourceLine  int        `json:"source_line"`
    SyncedAt    time.Time  `json:"synced_at"`
}

// deps/ bucket
// Key: change_id/task_num (string)
// Value: DependencyInfo (JSON)
type DependencyInfo struct {
    TaskID        string   `json:"task_id"`
    DependsOn     []string `json:"depends_on"`     // task_ids
    DependedBy    []string `json:"depended_by"`    // task_ids (reverse)
    IsBlocked     bool     `json:"is_blocked"`
    BlockingTasks []string `json:"blocking_tasks"`
}

// runtime/ bucket (ephemeral)
// Key: change_id/task_num (string)
// Value: RuntimeState (JSON)
type RuntimeState struct {
    TaskID      string    `json:"task_id"`
    Status      string    `json:"status"`       // in_progress
    Assignee    string    `json:"assignee"`
    Session     string    `json:"session"`
    StartedAt   time.Time `json:"started_at"`
    Notes       string    `json:"notes"`
}

// meta/ bucket
// Key: string
// Value: JSON
type SyncMeta struct {
    Version       string            `json:"version"`
    LastFullSync  time.Time         `json:"last_full_sync"`
    FileMTimes    map[string]int64  `json:"file_mtimes"`    // path -> mtime
}
```

## Implementation Phases

### Phase 1: Core BoltDB Layer (~4h)
Files:
- `internal/spec/boltdb.go` - BoltStore implementation
- `internal/spec/boltdb_test.go` - Unit tests

Key methods:
```go
type BoltStore struct {
    db *bolt.DB
    path string
}

func NewBoltStore(path string) (*BoltStore, error)
func (s *BoltStore) Open() error
func (s *BoltStore) Close() error
func (s *BoltStore) GetTask(changeID, taskNum string) (*Task, error)
func (s *BoltStore) PutTask(task *Task) error
func (s *BoltStore) DeleteTask(changeID, taskNum string) error
func (s *BoltStore) GetChange(changeID string) (*ChangeMetadata, error)
func (s *BoltStore) PutChange(change *ChangeMetadata) error
func (s *BoltStore) GetDeps(changeID, taskNum string) (*DependencyInfo, error)
func (s *BoltStore) PutDeps(deps *DependencyInfo) error
func (s *BoltStore) RebuildFromMarkdown(rootPath string) error
```

### Phase 2: File Watcher (~3h)
Files:
- `internal/spec/watcher.go` - File watcher with debounce

Dependencies:
- `github.com/fsnotify/fsnotify v1.7.0`

Key types:
```go
type Watcher struct {
    fsWatcher   *fsnotify.Watcher
    debounce    time.Duration
    boltStore   *BoltStore
    stateStore  *StateStore
    mu          sync.Mutex
    pending     map[string]bool  // files pending sync
}

func NewWatcher(boltStore *BoltStore, stateStore *StateStore) (*Watcher, error)
func (w *Watcher) Start(ctx context.Context) error
func (w *Watcher) Stop() error
func (w *Watcher) handleEvent(event fsnotify.Event)
func (w *Watcher) syncFile(path string) error
```

### Phase 3: MCP Tools (~3h)
Files:
- `internal/mcp/tools/registry_list.go`
- `internal/mcp/tools/registry_get.go`
- `internal/mcp/tools/registry_next.go`
- `internal/mcp/tools/registry_sync.go`

Update:
- `internal/mcp/tools/register.go`

### Phase 4: Integration (~2h)
Files to update:
- `internal/mcp/server/server.go` - Initialize BoltDB + watcher
- `.gitignore` - Add `openspec/.state.db`
- `go.mod` - Add fsnotify dependency

## Error Recovery

### 1. Parse Error in Markdown
- Log error, skip file
- Retry on next file change
- MCP tools show "parse error" for that change

### 2. BoltDB Corruption
- Detect on open (checksum validation)
- Delete .state.db, trigger full rebuild
- Log warning, continue

### 3. File Watcher Failure
- Fallback to explicit sync only (no auto-sync)
- Log error, alert user
- User can run `registry_sync` manually

### 4. Sync Conflict
- Last-write-wins (file mtime is authority)
- Log warning with details

## Debounce Pattern

```
Event 1 ──┐
          ▼
      ┌─────────┐
      │  Timer  │──┐
      │  100ms  │  │
      └─────────┘  │
          ▲        │
Event 2 ──┘        │ Reset timer
                   │
              (no events for 100ms)
                   │
                   ▼
              ┌─────────┐
              │  Sync   │
              └─────────┘
```

## Consistency Model

- **Eventual consistency**: BoltDB may lag behind markdown by ~100ms
- **Strong durability**: Markdown is persisted
- **Cache semantics**: BoltDB is disposable, can be rebuilt anytime

## Testing Strategy

1. Unit tests for BoltStore operations
2. Integration tests for file watcher
3. End-to-end tests: markdown → watcher → BoltDB → MCP
4. Error injection tests (corruption, parse errors)

## Performance Targets

- Full rebuild: < 1s for 100 changes
- Incremental sync: < 50ms per file
- MCP query: < 10ms
- Memory: < 50MB for 1000 tasks
