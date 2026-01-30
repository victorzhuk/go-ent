# Consolidate MCP Tools

## Summary

Reduce the 30+ MCP tools to a focused set of 8-10 core tools that map directly to OpenSpec workflow essentials. Merge overlapping functionality, remove unused tools, and simplify the API surface.

## Problem

The current MCP tool suite has grown to 30+ tools with significant overlap:

1. **Registry tools**: `registry_list`, `registry_next`, `registry_update`, `registry_deps`, `registry_sync`, `registry_init` (6 tools)
2. **Spec tools**: `spec_init`, `spec_create`, `spec_list`, `spec_show`, `spec_update`, `spec_delete`, `spec_validate`, `spec_archive` (8 tools)
3. **Workflow tools**: `workflow_state`, `workflow_start`, `workflow_complete`, `workflow_pause`, `workflow_resume` (5 tools)
4. **Skill tools**: `skill_list`, `skill_info`, `skill_validate`, `skill_quality` (4 tools)
5. **State tools**: `state_sync`, `state_show`, `state_update` (3 tools)
6. **Various others**: `generate_code`, `loop_detect`, `validate_change`, etc.

This creates:
- Maintenance burden across many handlers
- Confusion for agents about which tool to use
- Inconsistent APIs between similar tools
- Testing overhead

## Solution

### Core Tool Set (8 tools)

| Tool | Purpose | Merged From |
|------|---------|-------------|
| `spec_init` | Initialize OpenSpec structure | `spec_init` |
| `spec_list` | List specs/changes/tasks | `spec_list`, `registry_list` |
| `spec_show` | Show spec/change/task details | `spec_show`, `state_show` |
| `spec_create` | Create spec/change/task | `spec_create` |
| `spec_update` | Update spec/change/task status | `spec_update`, `registry_update`, `state_update` |
| `spec_validate` | Validate spec/change structure | `spec_validate`, `validate_change` |
| `spec_archive` | Archive completed change | `spec_archive` |
| `spec_sync` | Sync registry from tasks.md | `registry_sync`, `state_sync` |

### Removed Tools

| Tool | Reason |
|------|--------|
| `registry_next` | Merge into `spec_list` with filter |
| `registry_deps` | Move to `spec_update` dependency field |
| `registry_init` | Merge into `spec_init` |
| `workflow_*` | Simplify - use `spec_update` for status changes |
| `skill_list` | Remove - not core to OpenSpec workflow |
| `skill_info` | Remove - not core to OpenSpec workflow |
| `skill_validate` | Remove - validate on load instead |
| `skill_quality` | Remove - quality scoring removed |
| `generate_code` | Remove - out of scope for MVP |
| `loop_detect` | Remove - not used in practice |

## Affected Systems

- `internal/mcp/tools/registry.go` - Merge into spec tools
- `internal/mcp/tools/crud.go` - Simplify spec CRUD
- `internal/mcp/tools/workflow.go` - Remove, merge into spec_update
- `internal/mcp/tools/state.go` - Merge into spec tools
- `internal/mcp/tools/skill_*.go` - Remove skill tools
- `internal/mcp/tools/generate.go` - Remove
- `internal/mcp/tools/loop.go` - Remove
- `internal/mcp/tools/validate.go` - Merge into spec_validate
- `internal/mcp/server/server.go` - Update tool registration
- `docs/MCP_API.md` - Update documentation

## Breaking Changes

- [x] Remove 20+ MCP tools
- [x] Merge registry tools into spec tools
- [x] Simplify tool APIs
- [x] Remove skill management tools

## Migration Path

1. Create new consolidated tools
2. Mark old tools as deprecated (keep for one release)
3. Update agents to use new tools
4. Remove deprecated tools

## Alternatives Considered

1. **Keep all tools**: Rejected - maintenance burden too high
2. **Merge into fewer than 8**: Rejected - would lose essential functionality
3. **8 core tools (chosen)**: Balances simplicity with functionality

## Success Criteria

- [ ] 8 core tools implemented
- [ ] 20+ old tools removed
- [ ] Agent prompts updated
- [ ] Tests updated and passing
- [ ] Documentation updated
- [ ] Backward compatibility handled

## Effort Estimate

**~12 hours** across 10 tasks
