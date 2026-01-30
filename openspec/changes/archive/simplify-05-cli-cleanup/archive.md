# Archive: simplify-05-cli-cleanup

**Archived**: 2026-01-30
**Reason**: Invalid

## Original Intent
Clean up CLI layer by removing references to deleted packages, integrating ACP client, and removing obsolete commands (cli/run.go, cli/task.go, cli/agent.go, cli/model.go, cli/migrate.go). Part of "Go-Ent Simplification Series (5/5)".

## Why Archived
This proposal is invalid because it references files that do not exist in the current codebase:

**Referenced Files (DO NOT EXIST):**
- `cli/run.go` - does not exist
- `cli/task.go` - does not exist
- `cli/agent.go` - does not exist
- `cli/migrate.go` - does not exist

**Existing CLI Files:**
- `cli/cli_test.go`
- `cli/config.go`
- `cli/config_test.go`
- `cli/doc.go`
- `cli/errors_test.go`
- `cli/init.go`
- `cli/init_test.go`
- `cli/model.go` (exists but not as described)
- `cli/root.go`
- `cli/skill.go`
- `cli/spec.go`

The proposal is based on an outdated CLI structure that no longer exists.

## Actual State
The current CLI structure is already simplified and does not contain the files the proposal references. The CLI has been updated to match the current architecture without the execution/worker/agent systems that this proposal claims still exist.

## Files Preserved
- proposal.md
- This archive.md

## Notes for Future
This proposal was based on an outdated CLI structure. Always verify actual file structure before creating cleanup proposals. The CLI appears to already be cleaned up and aligned with current architecture.
