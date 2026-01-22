# Change: Clean Up Dead Code

## Why
Remove orphaned packages, deprecated functions, and incomplete stubs discovered during codebase quality review to reduce maintenance burden and improve code clarity before v0.1.0 release.

## What Changes
- Delete 4 orphaned directories: `internal/rules/`, `internal/tool/`, `internal/embedded/`, `internal/spec/cmd/`
- Remove deprecated functions from registry and skill systems
- Clean up incomplete TODO stubs
- **Total reduction**: ~500 lines of dead code

## Impact
- Affected code: internal package cleanup (no external API changes)
- No breaking changes
- Simplifies codebase for v0.1.0 release
- Build, test, and lint must pass after cleanup
