# Tasks: Refactor Move Core Packages

## Dependencies
- None (independent refactoring)
- T1.1 → T1.2 → T1.3
- T2.1 (parallel with T1.x)
- T1.3, T2.1 → T3.1

## Phase 1: Create Target Structure

### T1.1: Create /internal directory structure
- **Story**: proposal.md#Package Relocation
- **Files**: Directory creation only
- **Depends**: None
- **Parallel**: No
- [x] 1.1.1 Create `/internal/` directory at project root
- [x] 1.1.2 Verify directory is created successfully
- [x] 1.1.3 Add entry to `.gitignore` if needed (unlikely)

### T1.2: Move spec package
- **Story**: proposal.md#Package Relocation
- **Files**: All files in `cmd/go-ent/internal/spec/`
- **Depends**: T1.1
- **Parallel**: No
- [x] 1.2.1 Move `cmd/go-ent/internal/spec/*.go` to `/internal/spec/`
- [x] 1.2.2 Move all test files (*_test.go) to `/internal/spec/`
- [x] 1.2.3 Verify 14 files moved successfully
- [x] 1.2.4 Verify package declaration remains `package spec`

### T1.3: Move template and generation packages
- **Story**: proposal.md#Package Relocation
- **Files**: template/ and generation/ packages
- **Depends**: T1.2
- **Parallel**: No
- [x] 1.3.1 Move `cmd/go-ent/internal/template/` to `/internal/template/`
- [x] 1.3.2 Move `cmd/go-ent/internal/template/testdata/` recursively
- [x] 1.3.3 Move `cmd/go-ent/internal/generation/` to `/internal/generation/`
- [x] 1.3.4 Verify all 12 files (template + generation) moved
- [x] 1.3.5 Verify package declarations unchanged

## Phase 2: Update Import Paths

### T2.1: Update tools package imports
- **Story**: proposal.md#Import Path Updates
- **Files**: All files in `cmd/go-ent/internal/tools/`
- **Depends**: None
- **Parallel**: Yes (with T1.x - can start immediately)
- [x] 2.1.1 Update imports in `archive.go` (spec)
- [x] 2.1.2 Update imports in `crud.go` (spec)
- [x] 2.1.3 Update imports in `init.go` (spec)
- [x] 2.1.4 Update imports in `list.go` (spec)
- [x] 2.1.5 Update imports in `loop.go` (spec)
- [x] 2.1.6 Update imports in `registry.go` (spec)
- [x] 2.1.7 Update imports in `show.go` (spec)
- [x] 2.1.8 Update imports in `validate.go` (spec)
- [x] 2.1.9 Update imports in `workflow.go` (spec)
- [x] 2.1.10 Update imports in `generate.go` (template)
- [x] 2.1.11 Update imports in `generate_component.go` (generation)
- [x] 2.1.12 Update imports in `generate_from_spec.go` (generation, template)
- [x] 2.1.13 Update imports in `archetypes.go` (generation)
- [x] 2.1.14 Update imports in `server/server.go` (tools - transitive)

## Phase 3: Verification and Cleanup

### T3.1: Build and test verification
- **Story**: proposal.md#Success Criteria
- **Files**: Build system
- **Depends**: T1.3, T2.1
- **Parallel**: No
- [x] 3.1.1 Run `go mod tidy` to verify module consistency
- [x] 3.1.2 Run `make build` to verify successful compilation
- [x] 3.1.3 Run `make test` to verify all tests pass
- [x] 3.1.4 Run `make lint` to verify no new warnings
- [x] 3.1.5 Verify binary executes: `./dist/go-ent version`
- [x] 3.1.6 Remove old `/cmd/go-ent/internal/{spec,template,generation}` directories
- [x] 3.1.7 Verify no orphaned files remain in old locations

## Phase 4: Documentation

### T4.1: Update project documentation (if any)
- **Story**: proposal.md#Success Criteria
- **Files**: README.md, ARCHITECTURE_REVIEW.md
- **Depends**: T3.1
- **Parallel**: No
- [x] 4.1.1 Check if README.md mentions package structure
- [x] 4.1.2 Check if ARCHITECTURE_REVIEW.md mentions package locations
- [x] 4.1.3 Update any references to old paths
- [x] 4.1.4 Add note about new structure if beneficial

## Summary

**Total tasks**: 37 checklist items across 7 major tasks
**Estimated complexity**: Low-Medium (mostly mechanical moves)
**Critical path**: T1.1 → T1.2 → T1.3 → T3.1
**Parallelizable**: T2.1 can start immediately (import updates)
