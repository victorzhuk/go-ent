# Archive: consolidate-mcp-tools

**Archived**: 2026-01-30
**Reason**: Out of Scope

## Original Intent
Reduce 30+ MCP tools to 8 core tools by merging overlapping functionality, removing unused tools, and simplifying API surface.

## Why Archived
This proposal is out of scope and based on outdated tool counts:

**Claimed State (from proposal):**
- 30+ MCP tools across 6 categories
- Registry tools (6), Spec tools (8), Workflow tools (5), Skill tools (4), State tools (3), Various others

**Actual State:**
- 13 registered tools in `internal/mcp/tools/register.go`
- 17 tool files in `internal/mcp/tools/` directory
- Major consolidation already completed in earlier phases

**Current Tool Set (13 registered):**
- OpenSpec tools (8): init, list, show, crud, registry, workflow, validate, archive
- Skill tools (4): list, info, validate, quality  
- State tools (1): sync

The tool count has already been reduced significantly from the claimed 30+ to 13. Further consolidation to 8 tools is:
- Not a current priority
- Risk of removing useful functionality
- Would require significant refactoring of agent prompts
- Can be revisited later if needed

## Actual State
The MCP tool suite is already consolidated to a reasonable size (13 tools). The proposal's claim of "30+ tools" appears to be based on an outdated state or incorrect assumptions. The tool set is now focused and manageable without requiring further drastic reduction.

## Files Preserved
- proposal.md
- This archive.md

## Notes for Future
If tool consolidation becomes a priority again:
1. Verify actual current tool count before proposing changes
2. Assess which tools are genuinely unused vs. used but infrequently
3. Consider merging similar tools rather than wholesale deletion
4. Update agent prompts if tools are removed
5. Ensure backward compatibility during transition

The current 13-tool set appears to be a reasonable balance between functionality and complexity.
