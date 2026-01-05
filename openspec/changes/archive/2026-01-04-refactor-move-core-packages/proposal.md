# Change: Refactor Move Core Packages

## Why

The current project structure violates Go project layout best practices by placing significant reusable library code under `/cmd/go-ent/internal/`, making it inaccessible to potential future binaries, SDKs, or alternative frontends.

**Current Problem**:
- ~5000 lines of pure domain logic buried in `/cmd/go-ent/internal/`
- Three packages (`spec`, `template`, `generation`) have zero MCP dependencies
- No technical barrier exists to sharing these packages
- Structure prevents building an SDK or library around OpenSpec

**Industry Standard** (golang-standards/project-layout):
1. `/cmd/<app>/` should contain minimal application wiring and main packages only
2. `/internal/` at project root contains shared internal packages
3. Reusable libraries live separately from CLI-specific code

**Current Violations**:
| Package | LOC | Dependencies | Issue |
|---------|-----|--------------|-------|
| `cmd/go-ent/internal/spec` | ~3200 | stdlib, uuid, yaml | Core domain logic under cmd |
| `cmd/go-ent/internal/template` | ~280 | stdlib only | Generic utility under cmd |
| `cmd/go-ent/internal/generation` | ~1500 | stdlib, yaml | Pure logic under cmd |

These packages would be valuable for:
- Alternative frontends (REST API, gRPC, web UI)
- Direct Go SDK for OpenSpec workflows
- Testing tools and utilities
- Future CLI binaries

## What Changes

### 1. Package Relocation

Move 3 packages from `/cmd/go-ent/internal/` to `/internal/`:

```
Before:
/cmd/go-ent/internal/
    ├── spec/           → Move to /internal/spec/
    ├── template/       → Move to /internal/template/
    ├── generation/     → Move to /internal/generation/
    ├── tools/          ✓ Stay (MCP-specific)
    └── server/         ✓ Stay (MCP-specific)

After:
/internal/
    ├── spec/           (moved)
    ├── template/       (moved)
    └── generation/     (moved)
/cmd/go-ent/internal/
    ├── tools/          (stays - MCP handlers)
    └── server/         (stays - MCP factory)
```

### 2. Import Path Updates

Update all imports in `/cmd/go-ent/internal/tools/` (14 files):

```diff
-import "github.com/victorzhuk/go-ent/cmd/go-ent/internal/spec"
+import "github.com/victorzhuk/go-ent/internal/spec"

-import "github.com/victorzhuk/go-ent/cmd/go-ent/internal/template"
+import "github.com/victorzhuk/go-ent/internal/template"

-import "github.com/victorzhuk/go-ent/cmd/go-ent/internal/generation"
+import "github.com/victorzhuk/go-ent/internal/generation"
```

### 3. Files Affected

**Moved (26 files)**:
- `cmd/go-ent/internal/spec/*.go` (14 files + tests)
- `cmd/go-ent/internal/template/*.go` (2 files + tests + testdata)
- `cmd/go-ent/internal/generation/*.go` (10 files + tests)

**Import updates (14 files)**:
- `cmd/go-ent/internal/tools/archive.go`
- `cmd/go-ent/internal/tools/crud.go`
- `cmd/go-ent/internal/tools/generate.go`
- `cmd/go-ent/internal/tools/generate_component.go`
- `cmd/go-ent/internal/tools/generate_from_spec.go`
- `cmd/go-ent/internal/tools/init.go`
- `cmd/go-ent/internal/tools/list.go`
- `cmd/go-ent/internal/tools/loop.go`
- `cmd/go-ent/internal/tools/registry.go`
- `cmd/go-ent/internal/tools/show.go`
- `cmd/go-ent/internal/tools/validate.go`
- `cmd/go-ent/internal/tools/workflow.go`
- `cmd/go-ent/internal/tools/archetypes.go`
- `cmd/go-ent/internal/server/server.go`

## Impact

- **Affected specs**: None (internal refactoring)
- **New files**: None (moves only)
- **Affected code**: 40 files total (26 moved + 14 import updates)
- **Breaking changes**: None (internal paths only)
- **Dependencies**: None
- **Build system**: No changes needed

## Success Criteria

1. All packages moved to `/internal/` at project root
2. All imports updated successfully
3. `make test` passes with no failures
4. `make lint` passes with no warnings
5. `make build` produces working binary
6. Project structure follows Go best practices

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Missed import updates | Medium | Go compiler will catch missing imports |
| Test file path issues | Low | Tests use same package, relative paths unchanged |
| Build cache issues | Low | `go clean` before rebuild if needed |
| Merge conflicts | Low | No active branches modifying these packages |
