# Cleanup Deprecated Features

## Summary

Remove unused code, archived changes, and deprecated features that accumulated during development. Clean up the codebase to align with the simplified architecture vision.

## Problem

The codebase has accumulated technical debt:

1. **Archived changes in git**: 11 archived changes still in `openspec/changes/archive/`
2. **Unused packages**: Code for features that were never completed or are no longer needed
3. **Deprecated APIs**: Old function signatures kept for backward compatibility
4. **Dead code**: Functions, types, and constants that are never used
5. **Test artifacts**: Temporary test files and coverage reports committed

This creates:
- Slower builds
- Confusion for new contributors
- Maintenance overhead
- Larger binary size

## Solution

### 1. Clean up Archived Changes

Move archived changes out of active git tracking:

```bash
# Current (in git)
openspec/changes/archive/
├── 2026-01-18-add-context-memory/
├── 2026-01-18-add-dynamic-mcp-discovery/
└── ... (11 total)

# New (gitignored, kept locally)
openspec/changes/archive/  # <- added to .gitignore
```

Keep archive locally but don't track in git - they're historical record.

### 2. Remove Unused Packages

| Package | Status | Action |
|---------|--------|--------|
| `internal/execution/` | Unused | Remove |
| `internal/marketplace/` | Not used | Remove |
| `internal/template/testdata/` | Test only | Clean up |
| `internal/mcp/tools/exports/` | Temporary | Remove |
| `internal/mcp/tools/testdata/fallback/` | Test only | Review |

### 3. Remove Deprecated Code

- Old skill format parsing (v1, v2)
- Deprecated MCP tool handlers
- Unused domain types
- Old validation rules

### 4. Clean up Test Artifacts

```bash
# Remove from git, add to .gitignore
coverage.out
reg_coverage.out
coverage_skill.out
coverage_template.out
combined_coverage.out
internal/mcp/tools/*.json
internal/mcp/tools/*.csv
internal/mcp/tools/*.prom
```

### 5. Update .gitignore

```gitignore
# Test artifacts
*.out
coverage*.out

# OpenSpec archive (kept locally, not in git)
openspec/changes/archive/

# Temporary files
.tmp/
.crush/

# IDE
.idea/
```

## Affected Systems

- `openspec/changes/archive/` - Move to .gitignore
- `internal/execution/` - Remove
- `internal/marketplace/` - Remove
- `internal/mcp/tools/exports/` - Remove
- `internal/skill/` - Remove deprecated parsers
- `internal/domain/` - Remove unused types
- Root directory - Remove coverage files
- `.gitignore` - Update

## Breaking Changes

- [x] Remove unused packages
- [x] Remove deprecated APIs
- [x] Archive changes not in git
- [x] Clean up test artifacts

## Migration Path

1. Backup archived changes locally
2. Add `openspec/changes/archive/` to .gitignore
3. Remove unused packages
4. Remove deprecated code
5. Remove test artifacts from git
6. Update .gitignore
7. Verify build and tests still pass

## Alternatives Considered

1. **Keep everything**: Rejected - technical debt accumulates
2. **Archive in separate repo**: Rejected - overkill for MVP
3. **Local archive + gitignore (chosen)**: Simple, keeps history accessible

## Success Criteria

- [ ] Archived changes moved to .gitignore
- [ ] Unused packages removed
- [ ] Deprecated code removed
- [ ] Test artifacts removed from git
- [ ] .gitignore updated
- [ ] Build passes
- [ ] Tests pass
- [ ] Binary size reduced
- [ ] Clean `git status`

## Effort Estimate

**~8 hours** across 7 tasks
