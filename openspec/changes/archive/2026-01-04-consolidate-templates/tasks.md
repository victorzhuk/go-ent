# Tasks: Consolidate Templates and Fix Dotfile Embedding

## Dependencies
- **BLOCKED BY**: `refactor-move-core-packages` (creates `/internal/` directory)
- T1.1 → T1.2 → T1.3
- T2.1 (parallel with T1.x)
- T1.3, T2.1 → T3.1

## Phase 1: Template Consolidation

### T1.1: Create target directory
- **Story**: proposal.md#Template Consolidation
- **Files**: Directory creation
- **Depends**: None (assumes /internal/ exists from refactor-move-core-packages)
- **Parallel**: No
- [x] 1.1.1 Create `/internal/templates/` directory
- [x] 1.1.2 Verify directory created successfully

### T1.2: Move templates to /internal/templates/
- **Story**: proposal.md#Template Consolidation
- **Files**: All 15 `.tmpl` files from `/templates/`
- **Depends**: T1.1
- **Parallel**: No
- [x] 1.2.1 Move `.gitignore.tmpl` to `/internal/templates/`
- [x] 1.2.2 Move `.golangci.yml.tmpl` to `/internal/templates/`
- [x] 1.2.3 Move `CLAUDE.md.tmpl` to `/internal/templates/`
- [x] 1.2.4 Move `Makefile.tmpl` to `/internal/templates/`
- [x] 1.2.5 Move `go.mod.tmpl` to `/internal/templates/`
- [x] 1.2.6 Move `build/Dockerfile.tmpl` (preserve subdirectory)
- [x] 1.2.7 Move `cmd/server/main.go.tmpl` (preserve subdirectory)
- [x] 1.2.8 Move `deploy/docker-compose.yml.tmpl` (preserve subdirectory)
- [x] 1.2.9 Move `internal/app/app.go.tmpl` (preserve subdirectory)
- [x] 1.2.10 Move `internal/config/config.go.tmpl` (preserve subdirectory)
- [x] 1.2.11 Move `mcp/` directory recursively to `/internal/templates/mcp/`
- [x] 1.2.12 Verify all 15 files moved successfully
- [x] 1.2.13 Verify directory structure preserved

### T1.3: Create and fix embed.go
- **Story**: proposal.md#Fixed Embed Directive
- **Files**: `/internal/templates/embed.go`
- **Depends**: T1.2
- **Parallel**: No
- [x] 1.3.1 Move `/cmd/goent/templates/embed.go` to `/internal/templates/embed.go`
- [x] 1.3.2 Replace old `//go:embed **/*.tmpl` pattern
- [x] 1.3.3 Add explicit `//go:embed *.tmpl` for root-level templates
- [x] 1.3.4 Add explicit `//go:embed .gitignore.tmpl .golangci.yml.tmpl` for dotfiles
- [x] 1.3.5 Add `//go:embed build/*.tmpl`
- [x] 1.3.6 Add `//go:embed cmd/*.tmpl cmd/**/*.tmpl`
- [x] 1.3.7 Add `//go:embed deploy/*.tmpl`
- [x] 1.3.8 Add `//go:embed internal/*.tmpl internal/**/*.tmpl`
- [x] 1.3.9 Add `//go:embed mcp/*.tmpl mcp/**/*.tmpl`
- [x] 1.3.10 Verify package name is `templates`

## Phase 2: Code Updates

### T2.1: Update import paths
- **Story**: proposal.md#Import Path Updates
- **Files**: tools package files
- **Depends**: None
- **Parallel**: Yes (with T1.x - can start immediately)
- [x] 2.1.1 Update import in `/cmd/goent/internal/tools/generate.go`
- [x] 2.1.2 Update import in `/cmd/goent/internal/tools/generate_from_spec.go`
- [x] 2.1.3 Verify both files import `github.com/victorzhuk/go-ent/internal/templates`

### T2.2: Update Makefile
- **Story**: proposal.md#Build Process Simplification
- **Files**: `/Makefile`
- **Depends**: None
- **Parallel**: Yes (with T1.x - can start immediately)
- [x] 2.2.1 Remove `prepare-templates` target (lines 10-15)
- [x] 2.2.2 Update `build` target to not depend on `prepare-templates`
- [x] 2.2.3 Update `clean` target to not remove `cmd/goent/templates`
- [x] 2.2.4 Verify Makefile syntax is valid

## Phase 3: Cleanup and Verification

### T3.1: Remove old directories
- **Story**: proposal.md#Files Affected
- **Files**: Old template directories
- **Depends**: T1.3, T2.1, T2.2
- **Parallel**: No
- [x] 3.1.1 Remove `/templates/` directory (now empty)
- [x] 3.1.2 Remove `/cmd/goent/templates/` directory (no longer needed)
- [x] 3.1.3 Verify directories no longer exist
- [x] 3.1.4 Verify `.gitignore` still ignores `cmd/goent/templates/` (harmless)

### T3.2: Build and embed verification
- **Story**: proposal.md#Success Criteria
- **Files**: Build system
- **Depends**: T3.1
- **Parallel**: No
- [x] 3.2.1 Run `go mod tidy`
- [x] 3.2.2 Run `make build` (should work without prepare-templates)
- [x] 3.2.3 Verify binary created at `dist/goent`
- [x] 3.2.4 Check embed worked: `go list -f '{{.EmbedFiles}}' ./internal/templates`
- [x] 3.2.5 Verify dotfiles listed in embed output (`.gitignore.tmpl`, `.golangci.yml.tmpl`)
- [x] 3.2.6 Run `make test` - verify all tests pass
- [x] 3.2.7 Run `make lint` - verify no warnings

### T3.3: Functional testing
- **Story**: proposal.md#Success Criteria
- **Files**: Generated project test
- **Depends**: T3.2
- **Parallel**: No
- [x] 3.3.1 Test: Run `goent_generate` to create a test project
- [x] 3.3.2 Verify generated project includes `.gitignore` file
- [x] 3.3.3 Verify generated project includes `.golangci.yml` file
- [x] 3.3.4 Verify all other expected files generated correctly
- [x] 3.3.5 Clean up test project

## Phase 4: Documentation

### T4.1: Update documentation (if needed)
- **Story**: proposal.md#Success Criteria
- **Files**: README.md, build docs
- **Depends**: T3.3
- **Parallel**: No
- [x] 4.1.1 Check if README mentions template structure
- [x] 4.1.2 Update any references to `/templates/` → `/internal/templates/`
- [x] 4.1.3 Update any build instructions if needed
- [x] 4.1.4 Note: Simplified build (no prepare-templates needed)

## Summary

**Total tasks**: 47 checklist items across 8 major tasks
**Estimated complexity**: Medium (template moves + embed fixes)
**Critical path**: T1.1 → T1.2 → T1.3 → T3.1 → T3.2 → T3.3
**Parallelizable**: T2.1, T2.2 can start immediately
**Dependency**: Requires `refactor-move-core-packages` completed first
