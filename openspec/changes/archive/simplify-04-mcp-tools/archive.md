# Archive: simplify-04-mcp-tools

**Archived**: 2026-01-30
**Reason**: Mostly Done

## Original Intent
Reduce MCP tools from 70+ to 15 core tools by deleting broken/unused tools related to deleted packages (agent, ast, execution, generation, metrics, plugin, runtime, worker, meta, provider). Part of "Go-Ent Simplification Series (4/5)".

## Why Archived
This proposal is mostly completed. The major consolidation work has already been done in previous phases.

## Actual State
Current MCP tools are already consolidated to core set:

**Existing Tool Files (17):**
- `archive.go` - Archive operations
- `crud.go` - CRUD operations
- `generate.go` - Generation operations
- `init.go` - Initialization
- `list.go` - Listing operations
- `loop.go` - Loop operations
- `middleware.go` - Middleware
- `register.go` - Tool registration
- `registry.go` - Registry operations
- `show.go` - Show operations
- `skill_info.go` - Skill info
- `skill_list.go` - Skill list
- `skill_quality.go` - Skill quality
- `skill_validate.go` - Skill validation
- `state.go` - State management
- `validate.go` - Validation
- `workflow.go` - Workflow operations

**Registered Tools (13):**
- OpenSpec tools (8): init, list, show, crud, registry, workflow, validate, archive
- Skill tools (4): list, info, validate, quality
- State tools (1): sync

This is very close to the proposed 15 tools. The consolidation from "70+" to core set has already been completed.

## Files Preserved
- proposal.md
- This archive.md

## Notes for Future
The major tool consolidation work is complete. Current tool set appears to be the minimal necessary core functionality. No further reduction recommended unless specific tools prove unused or unnecessary.
