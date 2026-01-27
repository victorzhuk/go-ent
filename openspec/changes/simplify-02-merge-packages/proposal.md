# Proposal: Merge Small Packages (Phase 2)

## Metadata
- **Change ID:** `simplify-02-merge-packages`
- **Status:** Proposed
- **Type:** Refactoring
- **Priority:** High
- **Affects Specs:** None
- **Part of:** Go-Ent Simplification Series (2/5)
- **Depends On:** `simplify-01-delete-unused`

## Problem

After Phase 1, we still have small packages that should be merged into related packages:

1. **model/** (4 files, ~400 LOC) - Configuration model types
   - Contains category, config, defaults, resolver
   - Logically belongs in `config/` package

2. **openspec/** (2 files, ~200 LOC) - TaskTracker
   - Contains OpenSpec task tracking
   - Logically belongs in `spec/` package

These small packages add unnecessary navigation complexity without providing clear separation of concerns.

## Proposed Solution

### Merge model/ into config/

Move 4 files from `internal/model/` to `internal/config/`:
- `category.go` - Model category types
- `config.go` - Model configuration
- `defaults.go` - Default configurations
- `resolver.go` - Model resolution logic

Update all imports from `go-ent/internal/model` to `go-ent/internal/config`.

### Merge openspec/ into spec/

Move 2 files from `internal/openspec/` to `internal/spec/`:
- `tracker.go` - Task tracking for OpenSpec
- Any test files

Update all imports from `go-ent/internal/openspec` to `go-ent/internal/spec`.

## Impact

- **Breaking Changes:** None (internal refactoring)
- **API Changes:** Import paths change internally
- **Migration Required:** No
- **Testing Required:** Yes
  - Verify build succeeds
  - Verify tests pass
  - Verify no import errors

## Risks

- **Low Risk:** Simple file moves with import updates
- **Mitigation:** Update imports systematically
- **Rollback:** Git revert is straightforward

## Dependencies

- **Previous Proposal:** `simplify-01-delete-unused` (must complete first)
- **Next Proposal:** `simplify-03-acp-client` (depends on this)

## Success Criteria

- [ ] model/ package deleted
- [ ] openspec/ package deleted
- [ ] All files moved successfully
- [ ] All imports updated
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] `make lint` shows no errors
