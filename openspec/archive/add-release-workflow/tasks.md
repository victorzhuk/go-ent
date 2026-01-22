# Tasks: Add Release Workflow

## Status: complete

## 1. Create GitHub Actions workflow
- [x] 1.1 Create `.github/workflows/release.yml`
- [x] 1.2 Configure GoReleaser action with tag trigger (`v*` pattern)
- [x] 1.3 Set proper permissions for GitHub token (contents: write)
- [x] 1.4 Add release notes generation from CHANGELOG.md

## 2. Add Makefile targets
- [x] 2.1 Add `release-dry-run` target (runs `goreleaser release --snapshot --clean`)
- [x] 2.2 Add `snapshot` target (alias for release-dry-run)
- [x] 2.3 Add `release-check` target (runs `goreleaser check`)

## 3. Fix CHANGELOG.md
- [x] 3.1 Update placeholder dates (2025-XX-XX → actual dates)
- [x] 3.2 Review version numbering (verify 0.2.0 → 3.0.0 jump is intentional)
- [x] 3.3 Ensure format is compatible with GoReleaser release notes

## 4. Fix Go version alignment
- [x] 4.1 Update `.github/workflows/validate.yml` to Go 1.24
- [x] 4.2 Verify `go.mod` specifies `go 1.24`
- [x] 4.3 Ensure `.goreleaser.yml` uses Go 1.24

## 5. Verify release workflow
- [x] 5.1 Run `make release-dry-run`
- [x] 5.2 Run `goreleaser check`
- [x] 5.3 Verify workflow file syntax with `actionlint` (if available)
