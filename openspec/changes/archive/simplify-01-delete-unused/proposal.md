# Proposal: Delete Unused Packages (Phase 1)

## Metadata
- **Change ID:** `simplify-01-delete-unused`
- **Status:** Proposed
- **Type:** Refactoring
- **Priority:** High
- **Affects Specs:** None
- **Part of:** Go-Ent Simplification Series (1/5)

## Problem

The codebase has grown to 27 packages (~15,000 LOC) with significant over-engineering and unused code:

- **Unused multi-worker aggregation** (`aggregator/`) - 1100+ LOC not used
- **Unused pattern storage** (`memory/`) - 300+ LOC not used in MCP mode
- **Unused direct API** (`provider/`) - 800+ LOC, Claude Code handles API
- **Unused routing rules** (`router/`) - 600+ LOC not needed
- **Unused initialization adapters** (`toolinit/`) - 1500+ LOC not needed
- **Empty directory** (`embedded/`)
- **Misplaced integration tests** (`integration/`) - should be in package tests
- **Over-engineered execution** (`execution/`) - 3000+ LOC, will be replaced by ACP
- **Over-engineered workers** (`worker/`) - 1500+ LOC, will be replaced by ACP
- **Complex agent system** (`agent/`) - 1200+ LOC, execution via ACP
- **AST tools** (`ast/`) - 800+ LOC, handled by external tools
- **Code generation** (`generation/`) - 600+ LOC, handled by external tools
- **Metrics system** (`metrics/`) - 400+ LOC, rarely used
- **OpenCode runner** (`opencode/`) - will be replaced by ACP client

**Total deletion: ~12,800 LOC across 15 packages**

## Proposed Solution

Delete 15 unused packages entirely:

### Core Deletions
1. `internal/aggregator/` - multi-worker aggregation
2. `internal/memory/` - pattern storage
3. `internal/provider/` - direct Anthropic API
4. `internal/router/` - routing rules
5. `internal/toolinit/` - initialization adapters
6. `internal/embedded/` - empty directory

### Integration Tests
7. `internal/integration/` - move tests to respective packages

### Over-Engineered Systems (to be replaced by ACP)
8. `internal/execution/` - execution engine
9. `internal/worker/` - worker system
10. `internal/agent/` - agent system
11. `internal/opencode/` - OpenCode runner

### External Tool Responsibilities
12. `internal/ast/` - AST parsing
13. `internal/generation/` - code generation

### Optional Systems
14. `internal/plugin/` - simplified into config
15. `internal/metrics/` - rarely used

### Update Imports
- Remove all imports to deleted packages in `internal/mcp/server/server.go`
- Remove tool registrations in `internal/mcp/tools/register.go`
- Update initialization in `cmd/go-ent/main.go`

## Impact

- **Breaking Changes:** None (internal refactoring)
- **API Changes:** MCP tools will be removed in Phase 4
- **Migration Required:** No
- **Testing Required:** Yes
  - Verify build succeeds
  - Verify remaining tests pass
  - Verify MCP server starts

## Risks

- **Medium Risk:** Large deletion could break imports
- **Mitigation:** Sequential task execution with import updates
- **Rollback:** Git revert is straightforward

## Dependencies

- **Next Proposal:** `simplify-02-merge-packages` (depends on this)

## Success Criteria

- [ ] All 15 packages deleted
- [ ] `make build` succeeds
- [ ] `make test` passes for remaining packages
- [ ] `make lint` shows no errors
- [ ] MCP server starts without errors
- [ ] No dead imports remain
