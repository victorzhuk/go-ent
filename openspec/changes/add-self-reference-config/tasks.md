# Implementation Tasks

## 1. Configuration Files

### 1.1 Update Claude Code Settings
- [x] Update `.claude/settings.local.json` with `extraKnownMarketplaces`
- [x] Add `go-ent-local` marketplace pointing to `./plugins/go-ent`
- [x] Add `enabledPlugins` entry for `goent@go-ent-local`
- [x] Add `Bash(make build-mcp:*)` to permissions allowlist
- [x] Add `mcp__go_ent__*` to permissions allowlist

### 1.2 Verify Plugin Configuration
- [x] Check `plugins/go-ent/.claude-plugin/plugin.json` exists and is valid JSON
- [x] Verify plugin name is `goent`
- [x] Verify version is `1.0.0`
- [x] Check marketplace registration (not needed for directory source)

## 2. Build Infrastructure

### 2.1 Add Makefile Target
- [x] Add `build-mcp` target to Makefile
- [x] Make it depend on `build` target
- [x] Add message about MCP server location
- [x] Add reminder to restart Claude Code

### 2.2 Test Build Process
- [x] Run `make build-mcp` successfully
- [x] Verify `./dist/go-ent` binary created
- [x] Test `./dist/go-ent version` shows correct version
- [x] Verify binary size is reasonable (12MB < 50MB)

### 2.3 Verify MCP Server Functionality
- [x] Test MCP server starts via stdio transport (binary runs: `./dist/go-ent version`)
- [x] Fix MCP command path in plugin.json (`../../dist/go-ent` from plugin dir)
- [ ] Check graceful shutdown works (30s timeout) - **Requires Claude Code restart to test**
- [ ] Verify logging configuration (JSON/text modes) - **Requires Claude Code restart to test**

## 3. Documentation

### 3.1 Create Development Guide
- [x] Create `docs/DEVELOPMENT.md`
- [x] Add "Self-Hosted Development" section
- [x] Document initial setup procedure
- [x] Add development workflow section
- [x] Include hot-reload vs rebuild guidance
- [x] Document development patterns (agents, skills, commands, MCP tools)
- [x] Add bootstrap problem explanation
- [x] Include layered fallback strategy
- [x] Add troubleshooting section
- [x] Document advanced workflows (loop, registry)
- [x] Add delegation matrix
- [x] Include best practices

### 3.2 Update Root CLAUDE.md
- [x] Add self-hosted development section after OpenSpec block
- [x] Add key workflow commands table
- [x] Add available agents table with models
- [x] Add quick start procedure
- [x] Add link to `docs/DEVELOPMENT.md`

### 3.3 Documentation Quality Check
- [x] Verify all code examples are correct
- [x] Check all file paths are accurate
- [x] Ensure command syntax is valid
- [x] Test that links work correctly
- [x] Proofread for clarity and completeness

## 4. Validation

### 4.1 Plugin Loading
- [ ] Restart Claude Code
- [ ] Verify plugin loads from local directory
- [ ] Check Claude Code status bar for plugin confirmation
- [ ] Verify no errors in Claude Code logs

### 4.2 Agents Availability
- [ ] Verify `/go-ent:lead` available
- [ ] Verify `/go-ent:architect` available
- [ ] Verify `/go-ent:planner` available
- [ ] Verify `/go-ent:dev` available
- [ ] Verify `/go-ent:tester` available
- [ ] Verify `/go-ent:debug` available
- [ ] Verify `/go-ent:reviewer` available

### 4.3 Commands Availability
- [ ] Test `/go-ent:plan` autocomplete and help
- [ ] Test `/go-ent:apply` autocomplete and help
- [ ] Test `/go-ent:status` autocomplete and help
- [ ] Test `/go-ent:registry` autocomplete and help
- [ ] Test `/go-ent:archive` autocomplete and help
- [ ] Verify all 16 commands present in autocomplete

### 4.4 Skills Auto-Activation
- [ ] Open Go code file, verify `go-code` skill activates
- [ ] Work on architecture, verify `go-arch` skill activates
- [ ] Edit OpenAPI spec, verify `go-api` skill activates
- [ ] Work on database code, verify `go-db` skill activates
- [ ] Write tests, verify `go-test` skill activates

### 4.5 MCP Server Connection
- [ ] Verify MCP server connects successfully
- [ ] Check Claude Code status bar shows MCP connection
- [ ] Test MCP tool call: `mcp__go_ent__spec_list`
- [ ] Verify tools return expected results
- [ ] Check no errors in Claude Code MCP logs

## 5. Functional Testing

### 5.1 Planning Workflow
- [ ] Create test proposal: `/go-ent:plan Test self-hosted workflow`
- [ ] Verify proposal created in `openspec/changes/`
- [ ] Check `proposal.md` has correct structure
- [ ] Verify `tasks.md` generated with checklist

### 5.2 Execution Workflow
- [ ] Run `/go-ent:apply` to execute first task
- [ ] Verify task marked as in_progress
- [ ] Complete task, verify marked as completed
- [ ] Check registry updated correctly

### 5.3 Registry Management
- [ ] Run `/go-ent:registry list` to view all tasks
- [ ] Run `/go-ent:registry next 3` to get recommendations
- [ ] Update task status: `/go-ent:registry update T001 status=in_progress`
- [ ] Verify changes reflected in registry

### 5.4 Status and Monitoring
- [ ] Run `/go-ent:status` to view workflow state
- [ ] Verify shows current phase, progress, blockers
- [ ] Check displays active proposals correctly

### 5.5 Archiving
- [ ] Complete all tasks in test proposal
- [ ] Run `/go-ent:archive test-self-hosted-workflow`
- [ ] Verify proposal moved to `openspec/changes/archive/`
- [ ] Check specs updated correctly
- [ ] Verify registry cleaned up

## 6. Hot-Reload Testing

### 6.1 Plugin Changes (No Restart Required)
- [ ] Edit `plugins/go-ent/agents/go-ent:dev.md`
- [ ] Add temporary comment or change description
- [ ] Invoke `/go-ent:dev` without restarting Claude Code
- [ ] Verify changes reflected immediately
- [ ] Revert test changes

### 6.2 Skill Changes (No Restart Required)
- [ ] Edit `plugins/go-ent/skills/go-code/SKILL.md`
- [ ] Add temporary content
- [ ] Work on Go code to trigger skill
- [ ] Verify new content appears in context
- [ ] Revert test changes

### 6.3 Command Changes (No Restart Required)
- [ ] Edit `plugins/go-ent/commands/go-ent:plan.md`
- [ ] Modify description
- [ ] Check autocomplete shows new description
- [ ] Revert test changes

## 7. Rebuild Testing

### 7.1 MCP Server Code Changes (Restart Required)
- [ ] Edit `cmd/go-ent/internal/tools/list.go`
- [ ] Add temporary log message
- [ ] Run `make build-mcp`
- [ ] Verify build succeeds
- [ ] Restart Claude Code
- [ ] Test tool, verify changes reflected
- [ ] Revert test changes and rebuild

### 7.2 Rebuild Performance
- [ ] Measure `make build-mcp` time (should be < 10s)
- [ ] Verify binary size reasonable
- [ ] Check no errors or warnings during build

## 8. Dogfooding

### 8.1 Use This Proposal as Test Case
- [x] Created this proposal using OpenSpec format
- [ ] Track completion of all tasks
- [ ] Use `/go-ent:apply` for remaining implementation
- [ ] Document any UX friction discovered

### 8.2 Create Next Change Using Workflow
- [ ] Plan next go-ent change using `/go-ent:plan`
- [ ] Execute tasks using `/go-ent:apply`
- [ ] Use agents for implementation guidance
- [ ] Archive when complete using `/go-ent:archive`

### 8.3 Document UX Insights
- [ ] Note any confusing agent behaviors
- [ ] Record missing features or improvements
- [ ] Document workarounds used
- [ ] Create proposals for discovered issues

## 9. Error Handling and Recovery

### 9.1 Test Fallback Layers
- [ ] Test Layer 0: Manual OpenSpec CLI operations
- [ ] Test Layer 1: MCP tools without plugin
- [ ] Test Layer 2: Plugin without MCP server
- [ ] Document recovery procedures for each

### 9.2 Common Errors
- [ ] Test plugin not loading scenario
- [ ] Test MCP server not connecting scenario
- [ ] Test permissions denied scenario
- [ ] Verify troubleshooting guide covers all cases

## 10. Final Validation

### 10.1 Clean Slate Test
- [ ] Remove plugin from Claude Code
- [ ] Clear all caches
- [ ] Follow DEVELOPMENT.md setup procedure
- [ ] Verify successful setup in < 5 minutes

### 10.2 Cross-Platform Verification
- [ ] Test on Linux (primary platform)
- [ ] Test on macOS (if available)
- [ ] Test on Windows (if available)
- [ ] Document platform-specific issues

### 10.3 Performance Metrics
- [ ] Setup time: < 5 minutes
- [ ] Hot-reload time: < 3 seconds
- [ ] Rebuild time: < 10 seconds
- [ ] Command availability: 100%
- [ ] Agent success rate: > 90%

### 10.4 Documentation Completeness
- [ ] All sections in DEVELOPMENT.md complete
- [ ] All examples tested and working
- [ ] All troubleshooting scenarios covered
- [ ] Quick start in CLAUDE.md accurate

## 11. Cleanup and Finalization

### 11.1 Remove Test Data
- [ ] Remove test proposal created in 5.1
- [ ] Clean up any test files created
- [ ] Verify no temporary changes remain

### 11.2 Final Code Review
- [ ] Review all file changes
- [ ] Check for typos and errors
- [ ] Verify code style consistency
- [ ] Ensure no debug code remains

### 11.3 Git Commit
- [ ] Stage all changes
- [ ] Review diff carefully
- [ ] Create commit with descriptive message
- [ ] Include co-authored attribution

## Success Summary

When all tasks are complete:
- Configuration enables local plugin and MCP server
- Documentation provides clear setup and usage guide
- All 7 agents, 9 skills, and 16 commands functional
- Hot-reload works for plugin changes
- Rebuild works for MCP server changes
- Dogfooding validates the entire workflow
- Recovery procedures tested and documented

## Notes

- Tasks marked [x] are completed during initial implementation
- Tasks marked [ ] require validation or testing
- Dogfooding section demonstrates self-hosting from the start
- Focus on UX insights to drive future improvements
