# Add Self-Reference Configuration

## Overview

Configure go-ent to use its own plugin and MCP server for self-development (dogfooding).

## Rationale

### Problem

- go-ent developers manually create proposals and manage changes without using go-ent workflows
- Available agents, skills, and commands are not utilized for go-ent development itself
- Missing dogfooding means UX issues are not discovered early
- No clear development setup guide for contributors
- Manual OpenSpec workflow is error-prone and tedious

### Solution

- Configure `.claude/settings.local.json` with local plugin marketplace
- Document bootstrap procedure and development workflow
- Create Makefile targets for rebuild/reload operations
- Add troubleshooting guide with layered fallback strategy
- Enable self-hosted development workflow

### Benefits

- **Early UX Discovery:** Discover usability issues early by using our own system
- **Design Validation:** Validate agent/skill/command design through real-world use
- **Better Onboarding:** Improve developer onboarding with clear, tested setup procedures
- **Demonstrate Capabilities:** Show go-ent capabilities to contributors through dogfooding
- **Continuous Improvement:** Create feedback loop for iterative enhancement
- **Reduced Friction:** Automated workflows reduce manual work and errors

## Key Components

### 1. Claude Code Configuration

**File:** `.claude/settings.local.json`

Additions:
- `extraKnownMarketplaces` with local plugin directory (`./plugins/go-ent`)
- `enabledPlugins` to enable `go-ent@go-ent-local`
- Permission allowlist for `make build-mcp` command
- Permission allowlist for all `mcp__go_ent__*` tools

### 2. Build Infrastructure

**File:** `Makefile`

New target:
- `build-mcp` - Alias to `build` with clear messaging about MCP server rebuild
- Reminds user to restart Claude Code after rebuild
- Provides clear separation between hot-reload (plugin) and rebuild (MCP server)

### 3. Documentation

**File:** `docs/DEVELOPMENT.md` (NEW)

Comprehensive development guide including:
- Initial setup procedure (build, config, restart, verify)
- Development workflow (plan → implement → archive)
- Hot-reload vs rebuild guidance
- Development patterns (adding agents, skills, commands, MCP tools)
- Bootstrap problem explanation with layered fallback strategy
- Troubleshooting section for common issues
- Advanced workflows (autonomous loop, registry management)
- Delegation matrix and best practices

**File:** `CLAUDE.md` (UPDATED)

Quick reference section after OpenSpec block:
- Self-hosted development overview
- Key workflow commands table
- Available agents table with models
- Quick start procedure
- Link to full DEVELOPMENT.md

### 4. Layered Fallback Architecture

Ensures recovery if components break:

- **Layer 0 (Manual):** Direct file editing + OpenSpec CLI (always works)
- **Layer 1 (MCP Only):** Use MCP tools directly, bypass plugin (plugin broken)
- **Layer 2 (Plugin Only):** Use agents/skills for guidance, manual OpenSpec (MCP broken)
- **Layer 3 (Full):** Complete workflow with all components (intended state)

This prevents the bootstrap problem from blocking development.

### 5. OpenSpec Proposal Tracking

**This proposal itself** is tracked via OpenSpec, demonstrating dogfooding from the start.

Files:
- `openspec/changes/add-self-reference-config/proposal.md` (this file)
- `openspec/changes/add-self-reference-config/tasks.md` (implementation checklist)

## Dependencies

### Required

None - this is a pure enhancement to development workflow.

### Blocking

None - no other changes are blocked by this.

## Success Criteria

Configuration & Infrastructure:
- [ ] `.claude/settings.local.json` configured with local plugin marketplace
- [ ] `make build-mcp` target added to Makefile
- [ ] MCP server builds successfully to `./dist/go-ent`

Documentation:
- [ ] `docs/DEVELOPMENT.md` created with comprehensive guide
- [ ] Root `CLAUDE.md` updated with self-hosting section
- [ ] Troubleshooting section covers common issues

Functionality:
- [ ] Plugin loads from local directory after Claude Code restart
- [ ] All 7 agents available (`/go-ent:*`)
- [ ] All 9 skills auto-activate on relevant code tasks
- [ ] All 16 commands available (`/go-ent:plan`, `/go-ent:apply`, etc.)
- [ ] MCP tools callable (`mcp__go_ent__*`)

Workflow Validation:
- [ ] Can create new proposal using `/go-ent:plan`
- [ ] Can execute tasks using `/go-ent:apply`
- [ ] Can view status using `/go-ent:status`
- [ ] Can manage registry using `/go-ent:registry`
- [ ] Can archive changes using `/go-ent:archive`

Hot-Reload & Rebuild:
- [ ] Hot-reload works for plugin changes (edit agent, no restart needed)
- [ ] Rebuild works for MCP server changes (`make build-mcp` + restart)

Dogfooding:
- [ ] This proposal itself managed via `/go-ent:plan` and `/go-ent:apply`
- [ ] Next go-ent change uses self-hosted workflow
- [ ] UX friction points documented for future improvement

## Implementation Notes

### Configuration Approach

Using `extraKnownMarketplaces` with `directory` source instead of direct plugin paths:
- Follows Claude Code's marketplace-based plugin system
- Allows versioning and updates
- Supports both local development and published marketplace versions

### MCP Server Connection

The MCP server runs as a stdio transport:
- Command: `./dist/go-ent`
- No arguments needed
- Graceful 30s shutdown timeout
- Supports both JSON and text logging

### Plugin vs MCP Server Separation

Clear separation of concerns:
- **Plugin (Markdown):** Agent definitions, skill knowledge bases, command specs
  - Hot-reloads automatically (data files)
- **MCP Server (Go):** Tool implementations, spec management, code generation
  - Requires rebuild + restart (compiled binary)

This separation enables fast iteration on plugin content while maintaining stability in MCP server infrastructure.

## Risk Analysis

### Low Risk

This change is entirely additive:
- No breaking changes to existing functionality
- Existing workflows continue to work (Layer 0 fallback)
- Plugin loading is isolated and won't affect other Claude Code functionality
- Settings are project-local (`.claude/settings.local.json`)

### Potential Issues

#### 1. Circular Dependency Confusion

**Risk:** Developers might get confused when the system they're developing breaks

**Mitigation:**
- Comprehensive fallback layer documentation
- Clear recovery procedures in DEVELOPMENT.md
- Troubleshooting section addresses this explicitly

**Likelihood:** Medium
**Impact:** Low (documentation handles it)

#### 2. Plugin Loading Conflicts

**Risk:** Local plugin might conflict with installed marketplace version

**Mitigation:**
- Use unique marketplace name `go-ent-local`
- Document plugin priority (local overrides marketplace)
- Provide uninstall procedure if needed

**Likelihood:** Low
**Impact:** Low (easily resolved)

#### 3. MCP Server Crashes

**Risk:** MCP server bugs could crash Claude Code's MCP client

**Mitigation:**
- Claude Code handles MCP crashes gracefully
- Fallback to Layer 1 or Layer 2
- Can disable MCP server without disabling plugin

**Likelihood:** Low
**Impact:** Medium (Claude Code remains stable)

#### 4. Path Issues Across Platforms

**Risk:** Relative paths might not work consistently

**Mitigation:**
- Use `./` prefix for relative paths (portable)
- Test on multiple platforms before release
- Document platform-specific issues if found

**Likelihood:** Low
**Impact:** Low (portable path format)

## Post-Implementation Plan

### Phase 1: Validation (Immediate)

1. Verify plugin loads and all components work
2. Test hot-reload workflow
3. Test rebuild workflow
4. Document any issues discovered

### Phase 2: Dogfooding (First Week)

1. Create next go-ent change using `/go-ent:plan`
2. Use `/go-ent:apply` for implementation
3. Track UX friction points
4. Note any missing features or improvements

### Phase 3: Iteration (Ongoing)

1. Improve agent instructions based on usage
2. Enhance skill knowledge bases with patterns discovered
3. Add new commands for common workflows
4. Optimize MCP tools based on performance data

### Phase 4: Documentation Expansion

1. Add video walkthrough of setup and workflow
2. Create FAQ from common issues
3. Document common patterns and anti-patterns
4. Share dogfooding insights with community

## Metrics for Success

### Quantitative

- **Setup Time:** < 5 minutes from clone to first `/go-ent:plan`
- **Hot-Reload Time:** < 3 seconds for plugin changes to reflect
- **Rebuild Time:** < 10 seconds for MCP server rebuild
- **Command Availability:** 100% of commands functional
- **Agent Success Rate:** > 90% of agent invocations succeed

### Qualitative

- Developers prefer using workflow over manual OpenSpec
- UX issues discovered and documented
- Positive feedback from contributors
- Reduced errors in proposal creation
- Faster iteration cycles on go-ent development

## Future Enhancements

After this change is complete:

1. **Plugin Marketplace Publishing:** Publish `go-ent` plugin to official Claude Code marketplace
2. **Auto-Update Support:** Implement version checking and update notifications
3. **Multi-Runtime Support:** Add OpenCode and CLI execution modes (from `add-execution-engine`)
4. **Enhanced MCP Tools:** Add more tools based on dogfooding insights
5. **Template System:** Improve code generation templates based on usage
6. **Agent Improvements:** Enhance agent instructions based on real-world usage patterns

## Conclusion

This change enables go-ent to develop itself using its own workflows, creating a powerful feedback loop for continuous improvement. The layered architecture ensures robustness, while comprehensive documentation makes onboarding straightforward for all contributors.

By dogfooding our own system, we ensure that go-ent remains practical, usable, and valuable for real-world development workflows.
