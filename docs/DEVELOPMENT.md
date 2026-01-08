# go-ent Development Guide

## Self-Hosted Development

go-ent uses its own plugin system for development (dogfooding). This means you can use go-ent's agents, skills, and workflows to develop go-ent itself.

## MCP Server Configuration (Dual Setup)

go-ent uses a **dual-configuration** approach:

### Production Configuration (Plugin)
**File**: `plugins/go-ent/.mcp.json` (committed)
```json
{
  "go-ent": {
    "command": "./scripts/run-mcp.sh",
    "args": [],
    "env": {}
  }
}
```
- ✅ Smart launcher script with auto-detection
- ✅ Works for both local dev and marketplace installs
- ✅ Production-ready, no hardcoded paths
- ✅ Committed to git for distribution

### Development Override (Project)
**File**: `.mcp.json` (project root, gitignored)
```json
{
  "go-ent": {
    "command": "./dist/go-ent",
    "args": [],
    "env": {
      "LOG_LEVEL": "info",
      "LOG_FORMAT": "text"
    }
  }
}
```
- ✅ Direct binary path for instant startup
- ✅ Takes priority over plugin config in this project
- ✅ Created automatically (gitignored)
- ✅ Avoids hardcoded path in plugin

**Why Both?** Claude Code's config priority is: Project `.mcp.json` → Plugin `.mcp.json`. For dogfooding, the project override uses the local binary directly. For marketplace users, the plugin's smart launcher handles their environment.

## Initial Setup

### 1. Build the MCP Server

```bash
make build
```

This creates `./dist/go-ent`, the MCP server binary that Claude Code will connect to.

### 2. Verify Configuration

Check that `.claude/settings.local.json` contains:

```json
{
  "extraKnownMarketplaces": {
    "go-ent-local": {
      "source": {
        "source": "directory",
        "path": "./plugins/go-ent"
      }
    }
  },
  "enabledPlugins": {
    "go-ent@go-ent-local": true
  },
  "permissions": {
    "allow": [
      "Bash(make build-mcp:*)",
      "mcp__go_ent__*",
      ...
    ]
  }
}
```

### 3. Restart Claude Code

Restart Claude Code to load the plugin and connect to the MCP server.

### 4. Verify Installation

After restart, verify:

**Agents Available:**
- `/go-ent:lead` - Orchestration and delegation
- `/go-ent:architect` - System design and architecture
- `/go-ent:planner` - Task breakdown and planning
- `/go-ent:dev` - Implementation and coding
- `/go-ent:tester` - Testing and TDD cycles
- `/go-ent:debug` - Bug investigation
- `/go-ent:reviewer` - Code review with confidence filtering

**Commands Available:**
- `/go-ent:plan` - Full planning workflow (clarify → research → decompose)
- `/go-ent:apply` - Execute next task from registry
- `/go-ent:status` - View workflow state
- `/go-ent:registry` - Manage task registry
- `/go-ent:archive` - Archive completed changes
- And 11 more commands...

**Skills Auto-Activate:**
- `go-code` - Go 1.25+ implementation patterns
- `go-arch` - Clean Architecture, DDD principles
- `go-api` - OpenAPI/gRPC patterns
- `go-db` - PostgreSQL, Redis integration
- `go-test` - Testing patterns, testcontainers
- `go-perf` - Performance profiling
- `go-sec` - Security, OWASP, auth
- `go-ops` - Docker, Kubernetes, CI/CD
- `go-review` - Code review patterns

## Development Workflow

### Making Changes to go-ent

#### 1. Use the Workflow to Plan Changes

```
/go-ent:plan Add new MCP tool for spec diffing
```

This creates a proposal in `openspec/changes/` with:
- `proposal.md` - Overview, rationale, dependencies, success criteria
- `tasks.md` - Implementation checklist
- Optionally `design.md` - Technical decisions

The planning workflow runs through:
1. **Clarify** - Ask 5 clarification questions
2. **Research** - Investigate unknowns and precedents
3. **Decompose** - Break into dependency-aware tasks
4. **Analyze** - Cross-document consistency check
5. **Checklist** - Generate acceptance criteria

#### 2. Implement Using Agents

Execute tasks from the registry:

```
/go-ent:apply
```

This:
- Fetches next recommended task from registry
- Checks dependencies are satisfied
- Uses appropriate agent based on task type
- Updates registry when complete

Or invoke agents directly:
- `/go-ent:dev` - Implementation assistance
- `/go-ent:tester` - Write tests
- `/go-ent:reviewer` - Code review
- `/go-ent:architect` - Design guidance

#### 3. Archive When Deployed

```
/go-ent:archive add-spec-diff-tool
```

This:
- Validates all tasks completed
- Moves proposal to `openspec/changes/archive/YYYY-MM-DD-{id}/`
- Updates `openspec/specs/` with delta changes
- Clears workflow state

### Hot-Reloading Changes

**Plugin Changes (Agents/Skills/Commands):**

Edit Markdown files in `plugins/go-ent/`:
- `agents/go-ent:*.md` - Agent definitions
- `skills/go-*/SKILL.md` - Skill knowledge bases
- `commands/go-ent:*.md` - Command definitions

Claude Code auto-reloads (no restart needed).

**MCP Server Changes (Go Code):**

Edit code in `cmd/go-ent/` or `internal/`:

```bash
make build-mcp
```

Then restart Claude Code to reload the MCP connection.

## Development Patterns

### Adding a New Agent

1. Create `plugins/go-ent/agents/go-ent:newagent.md`:

```markdown
---
name: go-ent:newagent
description: "Brief description of what this agent does"
tools: Read, Write, Grep, Glob
model: sonnet
color: blue
skills: go-code
---

# Agent Instructions

You are a specialized agent for...

## Responsibilities
- Responsibility 1
- Responsibility 2

## Workflow
1. Step 1
2. Step 2

## Examples
...
```

2. Test by invoking `/go-ent:newagent`

3. If agent works well, consider updating the delegation matrix in `go-ent:lead.md`

### Adding a New Skill

1. Create `plugins/go-ent/skills/go-skillname/SKILL.md`:

```markdown
---
name: go-skillname
description: "Skill description and auto-activation triggers..."
---

# Skill Name

## When This Activates

This skill auto-activates when:
- Trigger condition 1
- Trigger condition 2

## Knowledge Base

### Pattern 1
...

### Pattern 2
...
```

2. Test by working on code that should trigger it

3. Verify skill content appears in agent context when active

### Adding a New Command

1. Create `plugins/go-ent/commands/go-ent:newcmd.md`:

```markdown
---
name: go-ent:newcmd
description: "Command description..."
---

# Command: /go-ent:newcmd

## Purpose

What this command does...

## Usage

```
/go-ent:newcmd [arguments]
```

## Implementation

This command:
1. Step 1
2. Step 2
3. Invokes MCP tool `mcp__go_ent__tool_name`

## Examples

### Example 1
...
```

2. If the command needs MCP server functionality, add the tool

3. Test by invoking `/go-ent:newcmd`

### Adding a New MCP Tool

1. Create `internal/mcp/tools/newtool.go`:

```go
package tools

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
)

func registerNewToolTool(s *mcp.Server) {
	tool := mcp.NewTool("new_tool",
		mcp.WithDescription("Tool description"),
		mcp.WithString("param1",
			mcp.Required(),
			mcp.Description("Parameter description"),
		),
	)

	s.AddTool(tool, newToolHandler)
}

func newToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Param1 string `json:"param1"`
	}

	if err := request.UnmarshalArguments(&params); err != nil {
		return mcp.NewToolResultError(err), nil
	}

	// Implementation
	result := map[string]any{
		"success": true,
	}

	return mcp.NewToolResultText("Success", result), nil
}
```

2. Register in `internal/mcp/tools/register.go`:

```go
func Register(s *mcp.Server) {
	// ... existing tools ...
	registerNewToolTool(s)
}
```

3. Rebuild and restart:

```bash
make build-mcp
# Restart Claude Code
```

4. Test via MCP tool call: `mcp__go_ent__new_tool`

## Avoiding Circular Dependencies

### The Bootstrap Problem

To develop go-ent, you want to use go-ent workflows. But if go-ent is broken, you can't use it.

### Solution: Fallback Layers

The system has 4 layers of fallback:

#### Layer 0 (Always Works) - Manual Editing

Even if everything is broken:
- Edit `openspec/changes/*/proposal.md` by hand
- Use `openspec` CLI directly:
  ```bash
  openspec list
  openspec show add-feature
  openspec validate add-feature --strict
  openspec archive add-feature
  ```

This layer ALWAYS works (just files + CLI).

#### Layer 1 (Plugin Broken) - MCP Server Only

If agents/skills/commands are broken:
- Use MCP tools directly via Serena or other MCP clients
- Bypass agents/skills
- Manual OpenSpec operations

#### Layer 2 (MCP Server Broken) - Plugin Only

If the MCP server crashes or won't start:
- Use agents/skills for guidance
- Manual OpenSpec operations
- Can still get code review, design advice

#### Layer 3 (Everything Works) - Full Workflow

The intended state:
- `/go-ent:plan`, `/go-ent:apply` work
- Agents, skills, MCP tools all functional
- Full self-hosting achieved

### Recovery Process

If you break something:

1. **Identify which layer is broken**
   - Plugin not loading? → Use Layer 1 (MCP only)
   - MCP server crashing? → Use Layer 2 (Plugin only)
   - Both broken? → Use Layer 0 (Manual)

2. **Fall back to lower layer**
   - Use that layer to fix the issue
   - Test the fix

3. **Rebuild up**
   - Restart Claude Code
   - Verify each layer works
   - Resume full workflow

## Troubleshooting

### Plugin Not Loading

**Symptoms:**
- Agents not available (`/go-ent:architect` doesn't autocomplete)
- Commands not recognized
- Skills don't activate

**Checks:**
1. Verify `.claude/settings.local.json` has `extraKnownMarketplaces` config
2. Check `plugins/go-ent/.claude-plugin/plugin.json` exists
3. Restart Claude Code
4. Check Claude Code logs for plugin loading errors

**Fix:**
```bash
# Verify plugin.json is valid
cat plugins/go-ent/.claude-plugin/plugin.json

# Check marketplace registration
cat .claude-plugin/marketplace.json

# Restart Claude Code
```

### MCP Server Not Connecting

**Symptoms:**
- MCP tools not available
- `mcp__go_ent__*` calls fail
- Claude Code status bar shows MCP connection error

**Checks:**
1. Verify `./dist/go-ent` exists:
   ```bash
   ls -la ./dist/go-ent
   ```

2. Test MCP server manually:
   ```bash
   ./dist/go-ent
   # Should start stdio transport
   ```

3. Check Claude Code logs for MCP connection errors

**Fix:**
```bash
# Rebuild MCP server
make build-mcp

# Test binary
./dist/go-ent --version

# Restart Claude Code
```

### Tools Not Available

**Symptoms:**
- Specific MCP tool calls fail
- Permission denied errors

**Checks:**
1. Check MCP server is running (Claude Code status bar)
2. Verify tool names: `mcp__go_ent__*`
3. Check permissions in `.claude/settings.local.json`

**Fix:**
```bash
# List available tools (if MCP inspector available)
# Or check internal/mcp/tools/register.go

# Add permission if needed
# Edit .claude/settings.local.json:
# "mcp__go_ent__*"

# Restart Claude Code
```

### Build Failures

**Symptoms:**
- `make build` or `make build-mcp` fails
- Go compilation errors

**Checks:**
1. Go version: `go version` (should be 1.24+)
2. Dependencies: `go mod tidy`
3. Build errors in output

**Fix:**
```bash
# Update dependencies
go mod tidy

# Clean and rebuild
make clean
make build

# Check Go version
go version
```

### Hot-Reload Not Working

**Symptoms:**
- Plugin Markdown changes don't reflect
- Need to manually restart Claude Code

**Expected Behavior:**
- Agents/Skills/Commands (Markdown files) → auto-reload
- MCP server (Go code) → requires restart

**Fix:**
- For plugin changes: Wait a few seconds, should auto-reload
- For MCP server changes: `make build-mcp` and restart Claude Code

## Advanced Workflows

### Using the Autonomous Loop

For self-correcting implementation:

```
/go-ent:loop Implement spec diffing tool with error handling --max-iterations=10
```

This:
1. Executes tasks from registry
2. Runs validation after each task
3. Auto-corrects on errors
4. Stops after max iterations or completion

Cancel anytime:
```
/go-ent:loop-cancel
```

### Registry Management

View all tasks:
```
/go-ent:registry list
```

Get next recommended tasks:
```
/go-ent:registry next 5
```

Update task status/priority:
```
/go-ent:registry update T042 status=in_progress
/go-ent:registry update T042 priority=high
```

Sync tasks across proposals:
```
/go-ent:registry sync
```

### Multi-Phase Planning

For complex changes, use the full planning workflow:

```
/go-ent:plan Add distributed tracing support
```

This runs:
1. **Clarify** - Ask questions about scope, requirements
2. **Research** - Investigate tracing libraries, patterns
3. **Decompose** - Break into dependency-aware tasks
4. **Analyze** - Check consistency with existing specs
5. **Checklist** - Generate acceptance criteria

Or run phases individually:
```
/go-ent:clarify add-tracing
/go-ent:research add-tracing
/go-ent:decompose add-tracing
```

## Delegation Matrix

When `/go-ent:lead` delegates work, it uses this matrix:

| Task Type | Agent Flow |
|-----------|------------|
| New Feature | architect → planner → dev → tester → reviewer |
| Bug Fix | debug → tester → reviewer |
| Refactor | planner → dev → tester → reviewer |
| Simple Change | dev → tester |
| Architecture Decision | architect (consult only) |
| Performance Issue | debug (identify) → dev (fix) → tester (verify) |
| Security Issue | reviewer → dev → tester |

You can invoke agents directly if you know which one you need.

## Best Practices

### 1. Always Start with Planning

Don't jump straight to code. Use `/go-ent:plan` to:
- Clarify requirements
- Research patterns
- Decompose into tasks
- Get approval before implementation

### 2. Use Registry for Task Tracking

The registry tracks:
- Task ID, description, phase
- Dependencies (which tasks block others)
- Status (pending, in_progress, completed, blocked)
- Priority (low, medium, high, critical)
- Parallelization markers

This prevents:
- Forgotten tasks
- Dependency conflicts
- Out-of-order implementation

### 3. Archive Changes Properly

When archiving:
- Ensure all tasks completed
- Validate specs updated correctly
- Test deployed functionality
- Document any migration steps

The archive creates a historical record.

### 4. Dogfood Early and Often

Use go-ent to develop go-ent. When you hit friction:
- Note the UX issue
- Create a proposal to fix it
- Improve the experience

This feedback loop is the entire point of dogfooding.

### 5. Maintain Layer Separation

Keep clear boundaries:
- **Plugin** (agents/skills/commands) - Markdown definitions
- **MCP Server** (Go code) - Tools and infrastructure
- **OpenSpec** (proposals/specs) - Change management

Don't mix concerns across layers.

## Contributing

### For External Contributors

1. Fork the repository
2. Follow this development guide
3. Use go-ent workflows to create your changes
4. Submit PR with:
   - Proposal in `openspec/changes/`
   - Implementation matching tasks
   - Tests passing

### For Maintainers

1. Use self-hosted workflow for all changes
2. Track UX friction points
3. Iterate on agent/skill/command design
4. Keep documentation up-to-date

## Resources

- **OpenSpec Workflow:** `openspec/AGENTS.md` (835 lines of comprehensive instructions)
- **Plugin Structure:** `plugins/go-ent/README.md`
- **Skill Definitions:** `plugins/go-ent/skills/*/SKILL.md`
- **Agent Definitions:** `plugins/go-ent/agents/*.md`
- **Command Definitions:** `plugins/go-ent/commands/*.md`

## Version History

### v0.3.0 (Current)
- Initial self-hosted development setup
- 7 agents, 9 skills, 16 commands
- Full OpenSpec workflow integration
- Layered fallback architecture
