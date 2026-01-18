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

**Legacy (v1) Format** - Still supported for backward compatibility:

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

**New skills should use v2 format** - see [Skill Authoring (v2 Format)](#skill-authoring-v2-format) below.

---

## Skill Authoring (v2 Format)

### Overview

The v2 skill format provides structured, validated, and high-quality skill definitions with automatic quality scoring. Skills in v2 format include:

- **Required XML sections**: `<role>`, `<instructions>`, `<constraints>`, `<edge_cases>`, `<examples>`, `<output_format>`
- **Enhanced frontmatter**: `version`, `author`, `tags` fields
- **Validation**: Automatic checking for required sections and content
- **Quality scoring**: 0-100 scale with detailed breakdown
- **MCP tools**: `skill_validate` and `skill_quality` for inspection

### File Location

Place new skills in:
```
plugins/go-ent/skills/{category}/{skill-name}/SKILL.md
```

Example:
```
plugins/go-ent/skills/go/go-code/SKILL.md
plugins/go-ent/skills/core/arch-core/SKILL.md
```

### Skill Structure (v2 Format)

A v2 skill file has this structure:

```markdown
---
name: go-code
description: "Modern Go implementation patterns, error handling, concurrency. Auto-activates for: writing Go code, implementing features, refactoring, error handling, configuration."
version: "2.0.0"
author: "go-ent"
tags: ["go", "code", "implementation"]
---

# Go Code Patterns

<role>
Expert Go developer focused on clean architecture, patterns, and idioms. Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality, maintainability, and performance.
</role>

<instructions>
[Detailed task instructions and patterns]
</instructions>

<constraints>
[What to include, what to exclude, boundaries]
</constraints>

<edge_cases>
[5+ scenarios with handling instructions]
</edge_cases>

<examples>
<example>
<input>Example input</input>
<output>Example output</output>
</example>
</examples>

<output_format>
[Expected output structure and format]
</output_format>
```

### Frontmatter Fields

| Field      | Required | Description                               | Example                          |
|------------|----------|-------------------------------------------|----------------------------------|
| `name`     | Yes      | Skill identifier                          | `go-code`                        |
| `description` | Yes   | What this skill does + auto-activation triggers | `"Modern Go patterns. Auto-activates for: writing code, implementing features"` |
| `version`  | No       | Semantic version                          | `"2.0.0"`                        |
| `author`   | No       | Attribution                               | `"go-ent"`                       |
| `tags`     | No       | Categorization array                      | `["go", "code", "implementation"]` |

**Triggers**: Extracted from `description` text following "Auto-activates for:" or "Activates when:"

### XML Sections

#### `<role>` - Expert Persona Definition

Define the AI's expertise and focus:

```xml
<role>
Expert Go developer focused on clean architecture, patterns, and idioms. Prioritize SOLID, DRY, KISS, YAGNI principles with production-grade quality, maintainability, and performance.
</role>
```

**Purpose**: Sets the persona and expertise level
**Content**: Expert identity, principles to follow, quality expectations

#### `<instructions>` - Task Instructions

Provide detailed, actionable guidance:

```xml
<instructions>

## Bootstrap Pattern

```go
func main() {
    // Code example
}
```

**Why this pattern**:
- Reason 1
- Reason 2

## Error Handling

```go
// Code example
```

**Rules**:
- Rule 1
- Rule 2
</instructions>
```

**Purpose**: Core knowledge and patterns
**Content**: Code examples, explanations, rules, patterns
**Format**: Markdown with code blocks, lists, emphasis

#### `<constraints>` - Boundaries and Requirements

Define what to include and exclude:

```xml
<constraints>
- Include clean, idiomatic Go code following standard conventions
- Include proper error wrapping with context using `%w` verb
- Include context propagation as first parameter throughout layers
- Exclude magic numbers (use named constants instead)
- Exclude global mutable state (pass dependencies explicitly)
- Exclude panic in production code (use error handling instead)
- Bound to clean layered architecture: Transport → UseCase → Domain ← Repository
</constraints>
```

**Purpose**: Set clear boundaries and requirements
**Content**: Include rules, exclude rules, architectural boundaries
**Format**: Bullet list starting with "Include" or "Exclude"

#### `<edge_cases>` - Edge Case Handling

Document 5+ scenarios with handling instructions:

```xml
<edge_cases>
If input is unclear or ambiguous: Ask clarifying questions to understand the specific requirement before proceeding with implementation.

If context is missing for a feature: Request additional information about architecture decisions, existing patterns, or integration points.

If performance concerns arise: Delegate to go-perf skill for profiling, optimization strategies, and benchmarking guidance.

If architecture questions emerge: Delegate to go-arch skill for system design, layer boundaries, and structural decisions.

If testing requirements are needed: Delegate to go-test skill for test coverage, table-driven tests, and mocking strategies.
</edge_cases>
```

**Purpose**: Handle edge cases and delegations
**Content**: 5+ scenarios with "If X: Y" format
**Format**: Each scenario on separate line

#### `<examples>` - Input/Output Pairs

Provide 2-3 concrete examples:

```xml
<examples>
<example>
<input>Refactor main() to use bootstrap pattern with graceful shutdown</input>
<output>
```go
func main() {
    if err := run(context.Background(), os.Getenv, os.Stdout, os.Stderr); err != nil {
        slog.Error("fatal", "error", err)
        os.Exit(1)
    }
}
```
</example>

<example>
<input>Fix error handling in this function - it's not wrapping errors properly</input>
<output>
```go
// Before
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    return nil, err
}

// After
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    if err != nil {
        return fmt.Errorf("query user %s: %w", id, err)
    }
}
```
</example>
</examples>
```

**Purpose**: Demonstrate skill application
**Content**: 2-3 examples with `<input>` and `<output>` tags
**Format**: Realistic user requests and responses

#### `<output_format>` - Expected Output Structure

Define expected output format:

```xml
<output_format>
Provide production-ready Go code following established patterns:

1. **Code Structure**: Clean, idiomatic Go with proper package organization
2. **Naming**: Short, natural variable names (cfg, repo, ctx, req, resp)
3. **Error Handling**: Wrapped errors with lowercase context using `%w`
4. **Context**: Always first parameter, propagated through all layers
5. **Interfaces**: Minimal interfaces at consumer side, return structs

Focus on practical implementation with minimal abstractions unless complexity demands it.
</output_format>
```

**Purpose**: Guide output structure and format
**Content**: Format requirements, structure expectations, emphasis
**Format**: Clear, actionable guidelines

### Backward Compatibility

**v1 format** (no XML tags) still works:
- Detected by absence of `<role>` and `<instructions>` tags
- Loaded as legacy format
- No validation or quality scoring

**v2 format**:
- Detected by presence of `<role>` or `<instructions>` tags
- Fully validated and scored
- Enhanced metadata (version, author, tags)

### Quality Scoring

Quality scores range from 0-100 and are computed automatically:

| Category        | Points | Criteria                                                                 |
|-----------------|--------|--------------------------------------------------------------------------|
| **Frontmatter** | 20     | name (5), description (5), version (5), tags (5)                         |
| **Structure**   | 30     | `<role>` (10), `<instructions>` (10), `<examples>` (10)                  |
| **Content**     | 30     | 2+ examples (15), `<edge_cases>` (15)                                    |
| **Triggers**    | 20     | 3+ triggers (20), 1-2 triggers (6.67 each)                               |

**Thresholds**:
- ≥ 90: Excellent quality (template skills)
- ≥ 80: Good quality (acceptable for production)
- < 80: Needs improvement (add sections, examples, triggers)

### Validation

Validate skills with strict mode:

```bash
make skill-validate
```

This checks:
- Required frontmatter fields
- Well-formed XML tags
- Required XML sections
- Proper example format
- Edge cases present
- Semantic version format

### Quality Report

Generate quality report:

```bash
make skill-quality
```

Output example:
```
Skill Quality Report
===================

go-code: Score 95/100 ✓
  Frontmatter: 20/20
  Structure: 30/30
  Content: 30/30
  Triggers: 15/20

go-arch: Score 88/100 ✓
  Frontmatter: 20/20
  Structure: 30/30
  Content: 25/30 (edge_cases missing 1 case)
  Triggers: 13/20

my-new-skill: Score 65/100 ✗
  Frontmatter: 15/20 (version missing)
  Structure: 20/30 (examples missing)
  Content: 15/30 (edge_cases missing)
  Triggers: 15/20

Summary: 2/3 skills meet quality threshold (≥80)
```

### MCP Tools

#### skill_validate

Validate a single skill or all skills:

```
mcp__go_ent__skill_validate
  skill_id: string (optional) - Skill name to validate, or validate all
  strict: boolean (optional) - Enable strict validation (default: false)
```

Example via Claude Code:
```
Use skill_validate with skill_id="go-code", strict=true
```

#### skill_quality

Get quality report with threshold:

```
mcp__go_ent__skill_quality
  skill_id: string (optional) - Skill name to check, or report all
  threshold: number (optional) - Minimum score (default: 80)
```

Example via Claude Code:
```
Use skill_quality with skill_id="go-arch", threshold=90
```

### Migration Checklist

Migrating from v1 to v2 format:

1. **Use go-code as template**
   ```bash
   cp plugins/go-ent/skills/go/go-code/SKILL.md plugins/go-ent/skills/your-category/your-skill/SKILL.md
   ```

2. **Update frontmatter**
   ```yaml
   ---
   name: your-skill
   description: "Your skill description. Auto-activates for: trigger1, trigger2, trigger3"
   version: "2.0.0"
   author: "your-name"
   tags: ["category", "keyword"]
   ---
   ```

3. **Add required XML sections**
   - `<role>` - Expert persona
   - `<instructions>` - Core patterns and guidance
   - `<constraints>` - Include/exclude rules
   - `<edge_cases>` - 5+ handling scenarios
   - `<examples>` - 2-3 input/output pairs
   - `<output_format>` - Expected output structure

4. **Preserve existing content**
   - Keep all existing patterns, examples, and knowledge
   - Format as markdown within appropriate sections
   - Ensure code blocks have language tags

5. **Validate with strict mode**
   ```bash
   make skill-validate
   ```

6. **Check quality score**
   ```bash
   make skill-quality
   ```
   Target ≥ 80 (≥ 90 for template/skills)

7. **Test with real work**
   - Use Claude Code to trigger the skill
   - Verify skill content appears in context
   - Check output quality and relevance

### Example: Creating a New Skill

1. **Create directory and file**
   ```bash
   mkdir -p plugins/go-ent/skills/go/web-async
   cp plugins/go-ent/skills/go/go-code/SKILL.md plugins/go-ent/skills/go/web-async/SKILL.md
   ```

2. **Edit frontmatter**
   ```yaml
   ---
   name: web-async
   description: "Web asynchronous programming patterns with goroutines, channels, errgroups. Auto-activates for: async web handlers, background jobs, concurrent API calls"
   version: "2.0.0"
   author: "go-ent"
   tags: ["go", "web", "async", "concurrency"]
   ---
   ```

3. **Update sections**
   - Edit `<role>` for async web expertise
   - Update `<instructions>` with async patterns
   - Adjust `<constraints>` for async-specific rules
   - Add web async `<edge_cases>`
   - Provide async web `<examples>`
   - Update `<output_format>` expectations

4. **Validate and test**
   ```bash
   make skill-validate
   make skill-quality
   ```

5. **Test in Claude Code**
   - Trigger skill with async web task
   - Verify context includes skill content
   - Check output quality

### Best Practices

- **Use existing skills as templates**: go-code for Go skills, arch-core for architecture
- **Provide multiple triggers**: 3+ for better activation
- **Include 2-3 examples**: Realistic input/output pairs
- **Document 5+ edge cases**: Common scenarios and delegations
- **Target quality ≥ 80**: Higher (≥90) for template skills
- **Validate before commit**: Always run `make skill-validate`
- **Preserve existing content**: When migrating, keep all valuable knowledge
- **Use clear descriptions**: Include auto-activation triggers in description
- **Keep role concise**: 1-2 sentences defining expertise
- **Make examples realistic**: Real-world user requests and responses

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
- **Metrics System:** `docs/METRICS.md` (opt-out, data collection details)

## Version History

### v0.3.0 (Current)
- Initial self-hosted development setup
- 7 agents, 9 skills, 16 commands
- Full OpenSpec workflow integration
- Layered fallback architecture
