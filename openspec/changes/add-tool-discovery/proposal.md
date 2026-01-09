# Proposal: Add Tool Discovery System

## Overview

Implement progressive tool disclosure with lazy loading, TF-IDF search, and dynamic tool activation based on industry best practices from Anthropic, Docker, and Martin Fowler.

## Rationale

### Problem

- All 27 MCP tools load at startup, consuming ~3800 tokens
- No mechanism for agents to discover relevant tools
- Static tool registration violates progressive disclosure principle
- Context bloat for simple tasks that only need 2-3 tools

### Solution

- **Tool Registry**: Two-tier architecture (metadata always available, handlers lazy-loaded)
- **TF-IDF Search**: Stdlib-only relevance scoring for tool discovery
- **4 New MCP Tools**: `tool_find`, `tool_describe`, `tool_load`, `tool_active`
- **Dynamic Loading**: Load tools on-demand, reducing initial context by 70-90%

## Key Components

### Implementation Files

1. `internal/mcp/tools/discovery.go` - ToolRegistry with lazy loading
2. `internal/mcp/tools/search.go` - TF-IDF search index
3. `internal/mcp/tools/meta.go` - Discovery MCP tools
4. Updated `register.go` and `server.go` - Integration

### New MCP Tools

| Tool | Description |
|------|-------------|
| `tool_find` | Search tools by query using TF-IDF scoring |
| `tool_describe` | Get detailed metadata for a specific tool |
| `tool_load` | Activate tools dynamically into the active set |
| `tool_active` | List currently active (loaded) tools |

### Architecture

```
ToolRegistry
├── metadata (map[string]*ToolMeta)      # Always loaded
├── registrators (map[string]RegistrationFunc)  # Lazy loaded
├── active (map[string]bool)              # Currently active
└── index (*SearchIndex)                  # TF-IDF search
```

## Dependencies

- Requires: None (P0 - Foundation)
- Blocks: Future tool lazy-loading refactor
- Impacts: All tools (renamed from `go_ent_*` to shorter names)

## Breaking Changes

**BREAKING**: Tool names changed from `go_ent_*` prefix to shorter names:
- `go_ent_spec_init` → `spec_init`
- `go_ent_registry_list` → `registry_list`
- `go_ent_agent_execute` → `agent_execute`
- etc. (27 tools total)

## Success Criteria

- [x] TF-IDF search implementation (stdlib only)
- [x] ToolRegistry with lazy loading
- [x] 4 new MCP tools implemented
- [x] Tool renaming complete
- [x] Documentation updated
- [ ] Search index accuracy >80% for common queries
- [ ] Token reduction: 3800 → <500 for simple tasks

## Impact

### Performance

- **Token Savings**: 70-90% reduction for focused tasks
- **Initial Load**: Minimal (only metadata, ~200 tokens)
- **Search Performance**: O(n) where n = number of terms

### User Experience

- Agents discover tools via semantic search
- Progressive disclosure reduces cognitive load
- Backwards incompatible (tool name changes)

## Migration

**For Users:**
1. Update tool references in scripts/docs from `go_ent_*` to short names
2. No changes to MCP client configuration needed
3. New discovery tools available immediately

**For Developers:**
1. Future tools can opt-in to lazy loading
2. Existing tools work alongside discovery system
3. Tool metadata can be enhanced with keywords/categories
