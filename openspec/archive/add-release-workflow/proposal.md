# Change: Add Release Workflow

## Status: complete

## Why
Enable automated releases for v0.1.0 and future versions using GitHub Actions and GoReleaser. Current `.goreleaser.yml` exists but lacks automation and has inconsistencies (Go version mismatch, CHANGELOG placeholders).

## What Changes
- Add GitHub Actions release workflow (`.github/workflows/release.yml`)
- Add Makefile targets for release dry-run and snapshot builds
- Fix CHANGELOG.md placeholder dates and version numbering
- Align Go version across all CI workflows (1.24)
- Enable tag-based automated releases with GitHub token permissions

## Impact
- Affected code: CI/CD infrastructure only
- No code changes to application logic
- Enables one-command releases: `git tag v0.1.0 && git push origin v0.1.0`
- Required for v0.1.0 public release
