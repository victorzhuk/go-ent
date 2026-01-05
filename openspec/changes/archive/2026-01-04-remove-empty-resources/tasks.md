# Tasks: Remove Empty Resources Directory

## Dependencies
- None (independent cleanup)

## Phase 1: Verification

### T1.1: Verify directory is empty and unused
- **Story**: proposal.md#Directory Removal
- **Files**: None
- **Depends**: None
- **Parallel**: No
- [x] 1.1.1 Confirm `/cmd/go-ent/internal/resources/` contains zero files
- [x] 1.1.2 Search codebase for `import.*resources` - expect zero results
- [x] 1.1.3 Search for string "resources" in go files - verify no package references
- [x] 1.1.4 Check documentation for mentions

## Phase 2: Removal

### T2.1: Remove directory
- **Story**: proposal.md#Directory Removal
- **Files**: `/cmd/go-ent/internal/resources/`
- **Depends**: T1.1
- **Parallel**: No
- [x] 2.1.1 Delete `/cmd/go-ent/internal/resources/` directory
- [x] 2.1.2 Verify directory no longer exists

## Phase 3: Validation

### T3.1: Build and test verification
- **Story**: proposal.md#Success Criteria
- **Files**: Build system
- **Depends**: T2.1
- **Parallel**: No
- [x] 3.1.1 Run `make build` - verify success
- [x] 3.1.2 Run `make test` - verify all tests pass
- [x] 3.1.3 Run `make lint` - verify no warnings
- [x] 3.1.4 Verify binary executes: `./dist/go-ent version`

## Summary

**Total tasks**: 9 checklist items across 3 major tasks
**Estimated complexity**: Trivial
**Critical path**: T1.1 → T2.1 → T3.1
**Parallelizable**: None (linear flow)
