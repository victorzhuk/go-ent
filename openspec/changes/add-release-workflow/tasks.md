# Tasks: Add Release Workflow

## 1. Create GitHub Actions workflow
- [ ] 1.1 Create `.github/workflows/release.yml`
- [ ] 1.2 Configure GoReleaser action with tag trigger (`v*` pattern)
- [ ] 1.3 Set proper permissions for GitHub token (contents: write)
- [ ] 1.4 Add release notes generation from CHANGELOG.md

## 2. Add Makefile targets
- [ ] 2.1 Add `release-dry-run` target (runs `goreleaser release --snapshot --clean`)
- [ ] 2.2 Add `snapshot` target (alias for release-dry-run)
- [ ] 2.3 Add `release-check` target (runs `goreleaser check`)

## 3. Fix CHANGELOG.md
- [ ] 3.1 Update placeholder dates (2025-XX-XX → actual dates)
- [ ] 3.2 Review version numbering (verify 0.2.0 → 3.0.0 jump is intentional)
- [ ] 3.3 Ensure format is compatible with GoReleaser release notes

## 4. Fix Go version alignment
- [ ] 4.1 Update `.github/workflows/validate.yml` to Go 1.24
- [ ] 4.2 Verify `go.mod` specifies `go 1.24`
- [ ] 4.3 Ensure `.goreleaser.yml` uses Go 1.24

## 5. Verify release workflow
- [ ] 5.1 Run `make release-dry-run`
- [ ] 5.2 Run `goreleaser check`
- [ ] 5.3 Verify workflow file syntax with `actionlint` (if available)
