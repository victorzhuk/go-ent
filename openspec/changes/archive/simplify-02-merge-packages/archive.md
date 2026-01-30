# Archive: simplify-02-merge-packages

**Archived**: 2026-01-30
**Reason**: Invalid

## Original Intent
Merge small packages `internal/model/` and `internal/openspec/` into `internal/config/` and `internal/spec/` respectively to reduce navigation complexity.

## Why Archived
This proposal is invalid because it references packages that do not exist in the codebase:
- `internal/model/` - does not exist
- `internal/openspec/` - does not exist

The proposal appears to be based on outdated assumptions about the codebase structure.

## Actual State
The current `internal/` directory contains:
- `config/` - Configuration management (already exists)
- `spec/` - Spec store and validation (already exists)
- No `model/` or `openspec/` packages exist

The intended functionality appears to already be correctly organized in the `config/` and `spec/` packages.

## Files Preserved
- proposal.md
- tasks.md (if exists)
- This archive.md

## Notes for Future
This proposal was likely created during an earlier refactoring phase that was already completed or the packages were deleted in a different way. Always verify codebase structure before creating proposals based on assumptions.
