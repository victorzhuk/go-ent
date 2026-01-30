# Archive: continue-simplification-phase-2

**Archived**: 2026-01-30
**Reason**: Duplicate

## Original Intent
Continue Phase 2 of simplification by merging `internal/model/` and `internal/openspec/` packages into `internal/config/` and `internal/spec/` respectively.

## Why Archived
This proposal is an exact duplicate of `simplify-02-merge-packages`. Both proposals describe:
- Merging `internal/model/` (4 files) into `internal/config/`
- Merging `internal/openspec/` (2 files) into `internal/spec/`
- Same files to move, same import updates, same success criteria

Additionally, like `simplify-02-merge-packages`, this proposal is invalid because it references packages that do not exist:
- `internal/model/` - does not exist
- `internal/openspec/` - does not exist

## Actual State
The current `internal/` directory does not contain `model/` or `openspec/` packages. The proposed functionality appears to already be correctly organized in the `config/` and `spec/` packages.

## Files Preserved
- proposal.md
- This archive.md

## Notes for Future
Always check for duplicate proposals before creating new ones. This appears to have been created as a continuation/follow-up to an earlier proposal that already covered the same work.
