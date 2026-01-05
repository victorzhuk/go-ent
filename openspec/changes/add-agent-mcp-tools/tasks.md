# Tasks: Add Agent MCP Tools

## 1. Rename Existing Tools (Breaking Change)
- [ ] Rename all files from go_ent_* pattern to new naming
- [ ] Update tool names in internal/mcp/tools/*.go
- [ ] Update tool registration in internal/mcp/server/server.go
- [ ] Update plugin.json with new tool names

## 2. Create New Agent Tools
- [ ] Create internal/mcp/tools/agent_execute.go
- [ ] Create internal/mcp/tools/agent_status.go
- [ ] Create internal/mcp/tools/agent_list.go
- [ ] Create internal/mcp/tools/agent_delegate.go
- [ ] Create internal/mcp/tools/skill_list.go
- [ ] Create internal/mcp/tools/skill_info.go
- [ ] Create internal/mcp/tools/runtime_list.go
- [ ] Create internal/mcp/tools/runtime_status.go

## 3. Update Documentation
- [ ] Update AGENTS.md with new tool names
- [ ] Update plugin commands with new names
- [ ] Add migration guide for v2 → v3

## 4. Testing
- [ ] Test all renamed tools work
- [ ] Test new agent tools
- [ ] Integration test with Claude Code

## 5. Version Bump
- [ ] Update version to 3.0.0
- [ ] Update CHANGELOG
