# Migration Guide: Tool Name Changes

**Version**: v0.3.0
**Breaking Change**: Tool names simplified from `go_ent_*` prefix to shorter names
**Reason**: Progressive disclosure system reduces context overhead (70-90% token reduction)

## Overview

All 30 MCP tools have been renamed to remove the `go_ent_` prefix. This change enables:
- Cleaner tool names that match domain concepts
- Better integration with the new tool discovery system
- Reduced cognitive load when reading tool lists
- Consistency with industry patterns (Docker, Kubernetes, etc.)

## Complete Mapping

### Spec Operations (9 tools)

| Old Name | New Name | Purpose |
|----------|----------|---------|
| `go_ent_spec_init` | `spec_init` | Initialize openspec folder |
| `go_ent_spec_create` | `spec_create` | Create new spec, change, or task |
| `go_ent_spec_update` | `spec_update` | Update existing spec, change, or task |
| `go_ent_spec_delete` | `spec_delete` | Delete spec, change, or task |
| `go_ent_spec_list` | `spec_list` | List specs, changes, or tasks |
| `go_ent_spec_show` | `spec_show` | Show detailed content |
| `go_ent_spec_validate` | `spec_validate` | Validate OpenSpec files |
| `go_ent_spec_archive` | `spec_archive` | Archive completed change |

### Registry Operations (6 tools)

| Old Name | New Name | Purpose |
|----------|----------|---------|
| `go_ent_registry_init` | `registry_init` | Initialize empty registry |
| `go_ent_registry_list` | `registry_list` | List tasks with filters |
| `go_ent_registry_next` | `registry_next` | Get next recommended task |
| `go_ent_registry_update` | `registry_update` | Update task status/priority |
| `go_ent_registry_sync` | `registry_sync` | Sync registry from tasks.md |
| `go_ent_registry_deps` | `registry_deps` | Manage task dependencies |

### Workflow Operations (3 tools)

| Old Name | New Name | Purpose |
|----------|----------|---------|
| `go_ent_workflow_start` | `workflow_start` | Start guided workflow |
| `go_ent_workflow_approve` | `workflow_approve` | Approve wait point |
| `go_ent_workflow_status` | `workflow_status` | Check workflow status |

### Loop Operations (4 tools)

| Old Name | New Name | Purpose |
|----------|----------|---------|
| `go_ent_loop_start` | `loop_start` | Start autonomous loop |
| `go_ent_loop_cancel` | `loop_cancel` | Cancel running loop |
| `go_ent_loop_get` | `loop_get` | Get current loop state |
| `go_ent_loop_set` | `loop_set` | Update loop state |

### Generation Operations (4 tools)

| Old Name | New Name | Purpose |
|----------|----------|---------|
| `go_ent_generate` | `generate` | Generate new Go project |
| `go_ent_generate_component` | `generate_component` | Generate component scaffold |
| `go_ent_generate_from_spec` | `generate_from_spec` | Generate project from spec |
| `go_ent_list_archetypes` | `list_archetypes` | List available archetypes |

### Agent Operations (1 tool)

| Old Name | New Name | Purpose |
|----------|----------|---------|
| `go_ent_agent_execute` | `agent_execute` | Execute task with agent selection |

### Discovery Operations (4 NEW tools)

| Tool Name | Purpose |
|-----------|---------|
| `tool_find` | Search tools by semantic query |
| `tool_describe` | Get detailed tool metadata |
| `tool_load` | Activate tools dynamically |
| `tool_active` | List currently loaded tools |

## Migration Steps

### For MCP Client Users

**No configuration changes required.** Tool discovery happens transparently:

1. **Old workflow** (still works, but loads all tools):
   ```python
   # Direct tool call with new name
   result = await client.call_tool("spec_validate", {
       "path": ".",
       "type": "all"
   })
   ```

2. **New workflow** (progressive disclosure):
   ```python
   # Discover relevant tools first
   tools = await client.call_tool("tool_find", {
       "query": "validate spec",
       "limit": 3
   })

   # Load specific tools
   await client.call_tool("tool_load", {
       "names": ["spec_validate"]
   })

   # Use the tool
   result = await client.call_tool("spec_validate", {
       "path": ".",
       "type": "all"
   })
   ```

### For Scripts and Automation

**Update tool names in all scripts:**

```bash
# Before (v0.2.x)
mcp call go_ent_spec_validate --path=. --type=all

# After (v0.3.0+)
mcp call spec_validate --path=. --type=all
```

**Search and replace pattern:**
```bash
# Find all references to old tool names
rg 'go_ent_(spec|registry|workflow|loop|generate|agent)_\w+'

# Replace in files (example with sed)
sed -i 's/go_ent_spec_/spec_/g' scripts/*.sh
sed -i 's/go_ent_registry_/registry_/g' scripts/*.sh
sed -i 's/go_ent_workflow_/workflow_/g' scripts/*.sh
sed -i 's/go_ent_loop_/loop_/g' scripts/*.sh
sed -i 's/go_ent_generate/generate/g' scripts/*.sh
sed -i 's/go_ent_agent_execute/agent_execute/g' scripts/*.sh
sed -i 's/go_ent_list_archetypes/list_archetypes/g' scripts/*.sh
```

### For Documentation

**Update references in:**
- Command documentation (`plugins/go-ent/commands/*.md`)
- Agent instructions (`plugins/go-ent/agents/*.md`)
- OpenSpec agents guide (`openspec/AGENTS.md`)
- README files
- Tutorial content

**Automated check:**
```bash
# Find documentation with old tool names
rg 'go_ent_(spec|registry|workflow|loop|generate|agent)_\w+' docs/ plugins/ openspec/
```

### For Custom Agents

**Update tool references in agent prompts:**

```markdown
<!-- Before -->
Use `go_ent_spec_validate` to validate OpenSpec files.

<!-- After -->
Use `spec_validate` to validate OpenSpec files.
Or better yet:
Use `tool_find` to discover validation tools, then load as needed.
```

## Benefits of New Names

### 1. Cleaner Namespace
```
Before: go_ent_spec_validate
After:  spec_validate
Saving: 8 characters per reference
```

### 2. Domain-Aligned
- `spec_*` - Clearly spec operations
- `registry_*` - Task registry operations
- `workflow_*` - Workflow orchestration
- `loop_*` - Autonomous loops
- `generate*` - Code generation

### 3. Progressive Disclosure
```
Startup: Only meta tools (~147 tokens)
On-demand: Load 1-5 relevant tools (~200-350 tokens)
Old way: Load all 30 tools (~2,385 tokens)

Reduction: 85-94% for typical workflows
```

### 4. Semantic Discovery
```python
# Instead of knowing exact tool names
tool_find(query="validate spec")
# Returns: [spec_validate, spec_show, spec_list]

# Agents discover tools by intent, not by memorizing names
```

## Backward Compatibility

**No backward compatibility layer.** All references must be updated.

**Rationale:**
- Clean break preferred over maintaining legacy aliases
- Plugin is pre-1.0 (v0.x), breaking changes acceptable
- Tool discovery system reduces migration pain
- Small user base makes coordinated migration feasible

## Troubleshooting

### Error: "Unknown tool: go_ent_spec_validate"

**Cause**: Using old tool name after upgrade

**Fix**: Update to new name
```bash
# Old
mcp call go_ent_spec_validate

# New
mcp call spec_validate
```

### Error: "Tool not found in search results"

**Cause**: Query may be too specific or using old terminology

**Fix**: Broaden query
```python
# Too specific
tool_find(query="go_ent_spec_validate")  # Won't match

# Better
tool_find(query="validate spec")  # Matches spec_validate
tool_find(query="validate")       # Also works
```

### Gradual Migration

If updating all references is difficult:

1. **Identify critical paths** - Update production scripts first
2. **Search tooling** - Use grep/ripgrep to find all references
3. **Batch updates** - Group by category (spec, registry, workflow, etc.)
4. **Test incrementally** - Verify each category after update
5. **Document assumptions** - Note any scripts assuming old names

## Testing Migration

**Verify tool availability:**
```bash
# List all tools
mcp list-tools | grep -E '^(spec|registry|workflow|loop|generate|agent|tool)_'

# Expected: 30 tools with new names
# spec_* (9), registry_* (6), workflow_* (3), loop_* (4),
# generate* (4), agent_execute (1), tool_* (4)
```

**Test discovery:**
```bash
# Search for spec tools
mcp call tool_find --query="spec management" --limit=5

# Should return: spec_init, spec_create, spec_update, spec_list, spec_show
```

**Smoke test key operations:**
```bash
# Initialize spec (if new project)
mcp call spec_init --path=./test-project

# List specs
mcp call spec_list --type=spec

# Validate
mcp call spec_validate --path=. --type=all
```

## Version History

| Version | Tool Prefix | Count | Notes |
|---------|-------------|-------|-------|
| v0.1.x | `go_ent_*` | 26 | Initial release |
| v0.2.x | `go_ent_*` | 27 | Added agent_execute |
| v0.3.0 | (none) | 30 | Removed prefix, added discovery (4 tools) |

## Related Documentation

- [Tool Discovery Guide](../openspec/AGENTS.md#tool-discovery) - Using the discovery system
- [Development Guide](./DEVELOPMENT.md) - Building and testing
- [OpenSpec Agents](../openspec/AGENTS.md) - Agent instructions updated

## Support

**Questions or issues?**
- Check [GitHub Issues](https://github.com/victorzhuk/go-ent/issues)
- Review [Tool Discovery Guide](../openspec/AGENTS.md#tool-discovery)
- Test with `tool_find` to discover tools by intent

**Reporting migration problems:**
1. Include old tool name and expected new name
2. Provide error message if any
3. Note MCP client version and go-ent version
4. Share relevant script/code snippet (if applicable)
