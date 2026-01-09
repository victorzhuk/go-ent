# Tasks: Add Agent MCP Tools

## 1. Rename Existing Tools (Breaking Change)
- [x] Rename all files from go_ent_* pattern to new naming
- [x] Update tool names in internal/mcp/tools/*.go
- [x] Update tool registration in internal/mcp/server/server.go
- [x] Update plugin.json with new tool names

## 2. Create New Agent Tools
- [x] Create internal/mcp/tools/agent_execute.go
- [x] Create internal/mcp/tools/agent_status.go
- [x] Create internal/mcp/tools/agent_list.go
- [x] Create internal/mcp/tools/agent_delegate.go
- [x] Create internal/mcp/tools/skill_list.go
- [x] Create internal/mcp/tools/skill_info.go
- [x] Create internal/mcp/tools/runtime_list.go
- [x] Create internal/mcp/tools/runtime_status.go

## 3. Update Documentation
- [x] Update AGENTS.md with new tool names
- [x] Update plugin commands with new names
- [x] Add migration guide for v2 → v3

## 4. Testing
- [x] Test all renamed tools work
- [x] Test new agent tools
- [x] Integration test with Claude Code

## 5. Version Bump
- [x] Update version to 3.0.0
- [x] Update CHANGELOG
