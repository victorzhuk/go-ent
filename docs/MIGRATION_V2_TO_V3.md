# Migration Guide: v2.x → v3.0

**Version**: v3.0.0
**Breaking Change**: MCP plugin naming convention and new tool additions
**Release Date**: TBD

## Overview

Version 3.0 introduces new agent execution, skill management, and runtime introspection tools, plus updates to MCP plugin naming conventions for better compatibility with Claude Code.

## What's New in v3.0

### New Tools (3 additions)

**Skill Management:**
- `skill_info` - Get detailed information about a specific skill

**Runtime Management:**
- `runtime_list` - List available runtime environments and capabilities
- `runtime_status` - Get current runtime environment status and configuration

### Updated MCP Plugin Naming

When accessing tools via Claude Code's MCP system, the tool prefix has changed to follow the standard plugin naming convention:

| Context | Old Prefix | New Prefix |
|---------|------------|------------|
| Internal (go-ent code) | N/A | N/A (unchanged: `registry_list`, etc.) |
| MCP Client (Claude Code) | `mcp__go_ent__*` | `mcp__plugin_go-ent_go-ent__*` |

**Example:**
```
Old: mcp__go_ent__registry_list
New: mcp__plugin_go-ent_go-ent__registry_list
```

## Breaking Changes

### 1. Plugin Command Tool References

All plugin command files (`plugins/go-ent/commands/*.md`) have updated their `allowed-tools` fields to use the new MCP naming convention.

**Impact**: If you have custom commands or integrations referencing MCP tools by their full names, you'll need to update them.

**Files affected:**
- `plugins/go-ent/commands/apply.md`
- `plugins/go-ent/commands/status.md`
- `plugins/go-ent/commands/plan.md`
- `plugins/go-ent/commands/loop.md`
- `plugins/go-ent/commands/loop-cancel.md`
- `plugins/go-ent/commands/registry.md`

### 2. Documentation Updates

Documentation has been updated to reflect the new tool names and search patterns.

**Files updated:**
- `openspec/AGENTS.md` - Added skill and runtime management sections
- Tool discovery patterns now include `skill_info`, `runtime_list`, `runtime_status`

## Migration Steps

### For Claude Code Plugin Users

**No action required.** The plugin will automatically use the new naming when you restart Claude Code after updating go-ent.

### For Custom Command Developers

If you've created custom commands in `plugins/go-ent/commands/`, update the `allowed-tools` field:

**Before:**
```markdown
---
allowed-tools: mcp__go_ent__registry_next, mcp__go_ent__registry_update
---
```

**After:**
```markdown
---
allowed-tools: mcp__plugin_go-ent_go-ent__registry_next, mcp__plugin_go-ent_go-ent__registry_update
---
```

**Search and replace:**
```bash
# Find all custom commands with old naming
rg 'mcp__go_ent__' plugins/go-ent/commands/

# Replace (example with sed)
sed -i 's/mcp__go_ent__/mcp__plugin_go-ent_go-ent__/g' plugins/go-ent/commands/custom-*.md
```

### For Script and Automation Users

**No changes needed** if you're using:
- Command-line `openspec` CLI (no prefix used)
- Internal go-ent APIs (no prefix used)
- Tool discovery system (`tool_find`, `tool_load`)

**Changes needed** if you're:
- Directly calling MCP tools with full prefixes in scripts
- Integrating with Claude Code MCP from external tools
- Building custom MCP clients

Update your references:
```bash
# Replace in automation scripts
sed -i 's/mcp__go_ent__/mcp__plugin_go-ent_go-ent__/g' scripts/*.sh
```

### For Agent Developers

Update agent prompts and skill files that reference MCP tools by full name:

**Before:**
```markdown
Use `mcp__go_ent__registry_next` to get the next task.
```

**After:**
```markdown
Use `mcp__plugin_go-ent_go-ent__registry_next` to get the next task.
```

**Better approach** (recommended):
```markdown
Use `registry_next` tool to get the next task.
Or: Use tool discovery to find registry tools: `tool_find(query="next task")`
```

## New Tool Usage

### Skill Management

Get detailed information about a specific skill:

```bash
# Using MCP
mcp call skill_info --name="go-code"

# Using tool discovery
tool_find(query="skill information", limit=3)
# → Returns: skill_info
```

**Response includes:**
- Skill name and description
- Trigger patterns
- File path
- Full skill content (markdown)

### Runtime Management

List available runtime environments:

```bash
# List all runtimes
mcp call runtime_list

# Returns: claude-code, open-code, cli
# With capabilities for each
```

Get current runtime status:

```bash
# Check current runtime
mcp call runtime_status

# Returns: Current runtime, capabilities, configuration
```

**Use cases:**
- Determine which runtime features are available
- Check configuration settings
- Debug runtime-specific issues
- Adapt agent behavior based on capabilities

## Tool Discovery Updates

The tool discovery system now includes the new tools:

**Skill queries:**
```python
tool_find(query="skill information details", limit=3)
# → skill_info

tool_find(query="get skill content", limit=3)
# → skill_info
```

**Runtime queries:**
```python
tool_find(query="list runtime environments", limit=3)
# → runtime_list, runtime_status

tool_find(query="current runtime status", limit=3)
# → runtime_status, runtime_list
```

## Performance Impact

Total tools: 30 (v0.3.0) → 33 (v3.0)

| Scenario | Tools Loaded | Tokens | Reduction |
|----------|-------------|--------|-----------|
| All tools (baseline) | 33 | 2,600 | - |
| Meta tools only | 4 | 147 | 94.3% |
| Simple spec task | 5 | 218 | 91.6% |
| Registry workflow | 7 | 350 | 86.5% |
| Complex workflow | 10 | 500 | 80.8% |

**Impact**: Minimal increase in baseline (215 tokens), but progressive disclosure maintains 85%+ reduction for typical workflows.

## Testing Your Migration

### 1. Verify Tool Availability

```bash
# List all tools with new naming
mcp list-tools | grep -E '^(spec|registry|workflow|loop|generate|agent|skill|runtime|tool)_'

# Expected: 33 tools
# - spec_* (9)
# - registry_* (6)
# - workflow_* (3)
# - loop_* (4)
# - generate* (4)
# - agent_execute (1)
# - skill_info (1)
# - runtime_* (2)
# - tool_* (4)
```

### 2. Test New Tools

```bash
# Test skill info
mcp call skill_info --name="go-code"

# Test runtime list
mcp call runtime_list

# Test runtime status
mcp call runtime_status
```

### 3. Test Tool Discovery

```bash
# Search for new tools
mcp call tool_find --query="skill management" --limit=5
# Should return: skill_info

mcp call tool_find --query="runtime status" --limit=5
# Should return: runtime_status, runtime_list
```

### 4. Test Commands

```bash
# Verify commands work with new tool names
/go-ent:apply
/go-ent:status
/go-ent:registry list
```

## Rollback Instructions

If you need to rollback to v2.x:

1. **Reinstall previous version:**
   ```bash
   cd plugins/go-ent
   git checkout v2.x.x
   make build-mcp
   ```

2. **Restart Claude Code** to reload the plugin

3. **Revert custom command changes** (if any):
   ```bash
   git checkout HEAD -- plugins/go-ent/commands/
   ```

**Note**: No data migration needed. Registry, specs, and changes are forward/backward compatible.

## Troubleshooting

### Error: "Unknown tool: mcp__go_ent__registry_list"

**Cause**: Using old MCP tool prefix after upgrade

**Fix**:
- Restart Claude Code to reload plugin
- If in custom command, update to new prefix
- If in script, update references

### Error: "Tool not found: skill_info"

**Cause**: Using v2.x plugin with v3.0 commands

**Fix**: Update to v3.0:
```bash
cd plugins/go-ent
git pull
make build-mcp
# Restart Claude Code
```

### Custom Commands Not Working

**Cause**: Commands reference old MCP tool names

**Fix**:
```bash
# Check for old references
rg 'mcp__go_ent__' plugins/go-ent/commands/

# Update to new prefix
sed -i 's/mcp__go_ent__/mcp__plugin_go-ent_go-ent__/g' plugins/go-ent/commands/*.md
```

## Version Comparison

| Feature | v2.x | v3.0 |
|---------|------|------|
| Total tools | 30 | 33 |
| MCP prefix | `mcp__go_ent__*` | `mcp__plugin_go-ent_go-ent__*` |
| Skill introspection | ❌ | ✅ `skill_info` |
| Runtime introspection | ❌ | ✅ `runtime_list`, `runtime_status` |
| Tool discovery | ✅ | ✅ (enhanced) |
| Progressive disclosure | ✅ | ✅ (improved) |

## Related Documentation

- [Tool Discovery Guide](../openspec/AGENTS.md#tool-discovery) - Using the discovery system
- [Development Guide](./DEVELOPMENT.md) - Building and testing
- [OpenSpec Agents](../openspec/AGENTS.md) - Updated agent instructions
- [v0.3.0 Migration](./MIGRATION_TOOL_NAMES.md) - Previous tool name changes

## Support

**Questions or issues?**
- Check [GitHub Issues](https://github.com/victorzhuk/go-ent/issues)
- Review [Tool Discovery Guide](../openspec/AGENTS.md#tool-discovery)
- Test with `tool_find` to discover tools by intent

**Reporting migration problems:**
1. Include go-ent version (`go-ent version`)
2. Provide error message if any
3. Note Claude Code version
4. Share relevant command/script snippet
5. Indicate whether custom commands are involved
