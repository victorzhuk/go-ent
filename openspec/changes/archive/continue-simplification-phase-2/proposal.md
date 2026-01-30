# Proposal: Continue Simplification Phase 2

## Status: pending

## Why

**Current Problem:**

After Phase 1 (simplify-01-delete-unused) successfully removed unused packages, the codebase still has small packages that should be merged into related packages for better organization:

1. **model/** (4 files, ~400 LOC) - Configuration model types
   - Contains category, config, defaults, resolver
   - Logically belongs in `config/` package
   - Creates unnecessary package boundary

2. **openspec/** (2 files, ~200 LOC) - TaskTracker
   - Contains OpenSpec task tracking
   - Logically belongs in `spec/` package
   - Overlapping concerns with spec package

These small packages add navigation complexity without providing clear separation of concerns. The dependency chain shows `simplify-02-merge-packages` depends on `simplify-01-delete-unused` (complete) and blocks `simplify-03-acp-client`.

**Quantified Impact:**
- 2 packages to merge
- 6 files to move
- ~600 LOC affected
- Import updates required in 8-12 files

## What Changes

### Before

```
internal/
├── config/           # Configuration loading
│   ├── loader.go
│   └── config.go
├── model/            # Configuration models (TO MERGE)
│   ├── category.go
│   ├── config.go
│   ├── defaults.go
│   └── resolver.go
├── spec/             # Spec management
│   ├── store.go
│   └── validator.go
└── openspec/         # Task tracking (TO MERGE)
    └── tracker.go
```

Package count: 4 (config, model, spec, openspec)

### After

```
internal/
├── config/           # Configuration loading + models
│   ├── loader.go
│   ├── config.go
│   ├── category.go      # MOVED from model/
│   ├── model_config.go  # MOVED from model/config.go (renamed)
│   ├── defaults.go      # MOVED from model/
│   └── resolver.go      # MOVED from model/
└── spec/             # Spec management + tracking
    ├── store.go
    ├── validator.go
    └── tracker.go       # MOVED from openspec/
```

Package count: 2 (config, spec)

### Key Components

| Source | Destination | Change |
|--------|-------------|--------|
| `internal/model/category.go` | `internal/config/category.go` | Move + update package |
| `internal/model/config.go` | `internal/config/model_config.go` | Move + rename (avoid conflict) |
| `internal/model/defaults.go` | `internal/config/defaults.go` | Move + update package |
| `internal/model/resolver.go` | `internal/config/resolver.go` | Move + update package |
| `internal/openspec/tracker.go` | `internal/spec/tracker.go` | Move + update package |
| `internal/model/` | *deleted* | Remove empty directory |
| `internal/openspec/` | *deleted* | Remove empty directory |

**Import Updates Required:**
- Files importing `go-ent/internal/model` → update to `go-ent/internal/config`
- Files importing `go-ent/internal/openspec` → update to `go-ent/internal/spec`

## Impact

**Breaking Changes:** None (internal refactoring only)

**API Changes:** None (all internal packages)

**Benefits:**
- Reduced package count (4 → 2)
- Clearer package responsibilities
- Easier navigation
- Reduced import complexity
- Consistent with simplification goals

**Dependencies:**
- **Previous:** `simplify-01-delete-unused` (COMPLETE)
- **Next:** `simplify-03-acp-client` (blocked by this change)

## Success Criteria

- [ ] model/ package deleted
- [ ] openspec/ package deleted
- [ ] All 4 model files moved to config/
- [ ] tracker.go moved to spec/
- [ ] All imports updated (0 references to old paths)
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] `make lint` shows no new errors
- [ ] Package count reduced by 2

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Import conflicts during move | Low | Sequential file moves, update imports immediately |
| Naming conflict (config.go exists) | Low | Rename model/config.go to model_config.go |
| Missing import updates | Low | Use grep to verify no old imports remain |
| Build breakage | Low | Build after each file move |
| Test failures | Low | Run tests after each package merge |

## Alternatives Considered

1. **Keep packages separate** - Rejected: Small packages add complexity without benefit
2. **Merge into single package** - Rejected: config/ and spec/ have distinct concerns
3. **Create new combined package** - Rejected: Would require more import updates

## Related Documentation

- `openspec/changes/simplify-01-delete-unused/` - Previous simplification phase
- `openspec/changes/simplify-03-acp-client/` - Next simplification phase (depends on this)
- `docs/REFACTORING_GUIDE.md` - Refactoring patterns
