# Proposal: Repository Restructure

## Overview

Reorganize the repository structure by moving `cmd/go-ent/internal/*` packages into the top-level `internal/` directory. This establishes a clean, scalable foundation for the multi-agent system migration.

## Rationale

### Current Structure Problems

1. **Deep nesting**: `cmd/go-ent/internal/tools/*.go` creates unnecessarily deep import paths
2. **Inconsistency**: Core logic in both `internal/` and `cmd/go-ent/internal/`
3. **Tight coupling**: Tool handlers directly in cmd/ instead of being reusable packages
4. **Future conflicts**: New `internal/mcp/` would conflict with existing structure

### Current Layout

```
cmd/go-ent/
├── main.go
└── internal/
    ├── tools/          # ~15 MCP tool handlers
    ├── server/         # MCP server factory
    └── version/        # Version info

internal/
├── spec/              # Spec management
├── template/          # Template engine
├── templates/         # Embedded templates
└── generation/        # AI-assisted generation
```

### Target Layout

```
cmd/go-ent/
└── main.go            # Entry point only

internal/
├── mcp/               # MCP-specific code (moved from cmd/go-ent/internal)
│   ├── server/
│   └── tools/
├── version/           # Version info (moved)
├── spec/              # Existing
├── template/          # Existing
├── templates/         # Existing
└── generation/        # Existing
```

## Benefits

1. **Cleaner imports**: `internal/mcp/tools` vs `cmd/go-ent/internal/tools`
2. **Better organization**: Clear separation of concerns
3. **Easier testing**: Test packages can import from `internal/` directly
4. **Foundation for v3.0**: Prepares structure for agent/execution/plugin packages
5. **Standard Go layout**: Aligns with community best practices

## Breaking Changes

### Import Path Changes

All imports referencing `cmd/go-ent/internal/*` will change:

```go
// Before
import "github.com/victorzhuk/go-ent/cmd/go-ent/internal/tools"
import "github.com/victorzhuk/go-ent/cmd/go-ent/internal/server"

// After
import "github.com/victorzhuk/go-ent/internal/mcp/tools"
import "github.com/victorzhuk/go-ent/internal/mcp/server"
```

### Impact

- **External users**: None - these are internal packages
- **Internal codebase**: All files that import moved packages need updates
- **Build system**: No changes needed
- **Tests**: Import paths need updates

## Migration Strategy

### Phase 1: Move Files

```bash
# Move tool handlers
mv cmd/go-ent/internal/tools internal/mcp/tools

# Move server
mv cmd/go-ent/internal/server internal/mcp/server

# Move version
mv cmd/go-ent/internal/version internal/version
```

### Phase 2: Update Imports

Update imports in:
- `cmd/go-ent/main.go`
- `internal/mcp/server/server.go`
- `internal/mcp/tools/register.go`
- All tool handler files
- Test files

### Phase 3: Verify

```bash
# Ensure build works
go build ./cmd/go-ent

# Run tests
go test ./...

# Verify MCP server still works
./dist/go-ent serve
```

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Missed import updates | Build failure | Comprehensive grep for old paths |
| Test breakage | CI failure | Run full test suite before commit |
| MCP server broken | Runtime failure | Manual testing with Claude Code |

## Dependencies

- **Blocks**: All subsequent proposals (P1-P7)
- **Blocked by**: None

## Success Criteria

- [ ] All files moved to new locations
- [ ] All imports updated
- [ ] `go build ./cmd/go-ent` succeeds
- [ ] `go test ./...` passes
- [ ] MCP server starts and tools work
- [ ] No references to old paths in codebase
