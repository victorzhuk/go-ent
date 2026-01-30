# Archive: simplify-03-acp-client

**Archived**: 2026-01-30
**Reason**: Blocked/Deprecated

## Original Intent
Create an ACP (Agent Communication Protocol) client in `internal/acp/` to communicate with OpenCode server for task execution, monitoring, cancellation, and output streaming. Part of the "Go-Ent Simplification Series (3/5)".

## Why Archived
This proposal is blocked and deprecated for several reasons:

1. **Dependency on Invalid Proposal**: Depends on `simplify-02-merge-packages`, which is invalid (references non-existent packages)

2. **Architecture Mismatch**: The current codebase has `RuntimeOpenCode` as a runtime type but no indication of an actual ACP server or need for an ACP client

3. **Fragmented Series**: Part of a "simplification series" that appears to be based on outdated assumptions about the codebase structure

4. **Unclear Use Case**: No evidence that go-ent needs to execute tasks via OpenCode ACP server in the current architecture

## Actual State
- Runtime types include `RuntimeOpenCode` in `internal/domain/runtime.go`
- No ACP protocol implementation exists
- No OpenCode ACP server integration exists
- Current architecture doesn't appear to require external ACP client communication

## Files Preserved
- proposal.md
- This archive.md

## Notes for Future
If ACP client functionality is needed in the future:
1. Verify actual OpenCode ACP server availability and protocol
2. Reassess architecture alignment before implementation
3. Create a fresh proposal based on current codebase state
4. Ensure dependencies are valid before proposing
