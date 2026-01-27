# Proposal: Reduce MCP Tools to Core Set (Phase 4)

## Metadata
- **Change ID:** `simplify-04-mcp-tools`
- **Status:** Proposed
- **Type:** Refactoring
- **Priority:** High
- **Affects Specs:** None
- **Part of:** Go-Ent Simplification Series (4/5)
- **Depends On:** `simplify-03-acp-client`

## Problem

After deleting complex internal systems in Phase 1, we have 70+ MCP tools with many no longer functional:

**Broken/Unused Tools (55+ tools):**
- **Agent tools (10):** `agent_list`, `agent_delegate`, `agent_execute`, etc. - agent/ package deleted
- **AST tools (6):** `ast_parse`, `ast_query`, etc. - ast/ package deleted
- **Execution tools (7):** `execution_start`, `execution_status`, etc. - execution/ package deleted
- **Generation tools (3):** `generate_agent`, `generate_skill`, etc. - generation/ package deleted
- **Metrics tools (4):** `metrics_list`, `metrics_get`, etc. - metrics/ package deleted
- **Plugin tools (4):** `plugin_list`, `plugin_load`, etc. - plugin/ package deleted
- **Runtime tools (2):** `runtime_list`, `runtime_switch` - only OpenCode now
- **Worker tools (8):** `worker_list`, `worker_spawn`, etc. - worker/ package deleted
- **Meta tools (4):** `tool_discover`, `tool_search` - rarely used
- **Provider tools (2):** `provider_list`, `provider_config` - provider/ package deleted

**Keep Core Tools (15 tools):**
- **OpenSpec tools (~10):** init, list, show, create, update, delete, registry_list, workflow_status, validate, archive
- **Skill tools (~4):** skill_list, skill_info, skill_validate, skill_quality
- **State tools (~1):** state_sync

## Proposed Solution

### Delete 55+ Tool Files

Remove all tool implementation files for deleted packages:
- `internal/mcp/tools/agent_*.go` (10 files)
- `internal/mcp/tools/ast_*.go` (6 files)
- `internal/mcp/tools/execution_*.go` (7 files)
- `internal/mcp/tools/generate_*.go` (3 files)
- `internal/mcp/tools/metrics_*.go` (4 files)
- `internal/mcp/tools/plugin_*.go` (4 files)
- `internal/mcp/tools/runtime_*.go` (2 files)
- `internal/mcp/tools/worker_*.go` (8 files)
- `internal/mcp/tools/meta_*.go` (4 files)
- `internal/mcp/tools/provider_*.go` (2 files)

### Update Tool Registration

Update `internal/mcp/tools/register.go`:
- Remove all registrations for deleted tools
- Keep only OpenSpec, Skill, and State tools
- Simplify registration logic

### Update MCP Server

Update `internal/mcp/server/server.go`:
- Remove dependencies on deleted packages
- Remove initialization of deleted managers
- Simplify Server struct

**Result: 15 tools (from 70+), ~1,500 LOC (from ~4,000 LOC)**

## Impact

- **Breaking Changes:** Yes - many MCP tools removed
- **API Changes:** Clients using deleted tools will fail
- **Migration Required:** No (pre-release, no external users)
- **Testing Required:** Yes
  - Verify remaining tools work
  - Verify server starts
  - Verify tool registration

## Risks

- **Low Risk:** Tools already broken due to Phase 1 deletions
- **Mitigation:** Keep only tools with working implementations
- **Rollback:** Git revert is straightforward

## Dependencies

- **Previous Proposal:** `simplify-03-acp-client` (must complete first)
- **Next Proposal:** `simplify-05-cli-cleanup` (depends on this)

## Success Criteria

- [ ] 55+ tool files deleted
- [ ] Tool registrations updated
- [ ] Server dependencies cleaned up
- [ ] Remaining 15 tools work correctly
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] `make lint` shows no errors
- [ ] MCP server starts and lists only core tools
