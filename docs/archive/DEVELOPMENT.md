# go-ent Development Guide

## Self-Hosted Development

go-ent uses its own plugin system for development (dogfooding). This means you can use go-ent's agents, skills, and workflows to develop go-ent itself.

## MCP Server Configuration (Dual Setup)

go-ent uses a **dual-configuration** approach:

### Production Configuration (Plugin)
**File**: `.mcp.json` (committed)
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
        "path": "."
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
- `/ent:architect` - System design and architecture (opus)
- `/ent:planner` - Task breakdown and planning (sonnet)
- `/ent:planner-fast` - Quick task assessment (haiku)
- `/ent:planner-heavy` - Deep architectural planning (opus)
- `/ent:coder` - Implementation and coding (sonnet)
- `/ent:tester` - Testing and TDD cycles (sonnet)
- `/ent:debugger` - Bug investigation (sonnet)
- `/ent:debugger-fast` - Quick debugging (haiku)
- `/ent:debugger-heavy` - Complex debugging (opus)
- `/ent:reviewer` - Code review (opus)
- `/ent:researcher` - Codebase research (sonnet)
- `/ent:reproducer` - Bug reproduction (sonnet)
- `/ent:acceptor` - Acceptance criteria validation (sonnet)
- `/ent:decomposer` - Task breakdown (sonnet)
- `/ent:task-fast` - Quick task routing (haiku)
- `/ent:task-heavy` - Complex task analysis (opus)

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

Edit Markdown files in pkg/ directories:
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

**v3 Split Format** - Metadata (YAML) + Prompts (Markdown) for dual-platform support

1. Create metadata file `pkg/agents/meta/newagent.yaml`:

```yaml
name: newagent
description: "Brief description of when to use this agent"
model: main                        # main/fast/heavy (internal names)
color: '#32CD32'
role: execution
complexity: standard
skills:
  - go-code
toolPresets:
  - editing                        # Use presets instead of explicit tools
disallowedToolPresets:
  - serena-editing                 # Deny Serena editing tools
dependencies:
  - tester
  - reviewer
prompts:
  shared:                          # Shared prompt sections
    - _tooling
    - _conventions
    - _handoffs
  main: agents/newagent            # Agent-specific prompt
```

2. Create prompt file `pkg/agents/prompts/agents/newagent.md`:

```markdown
You are a specialized agent for...

## Responsibilities

- Responsibility 1
- Responsibility 2

## Workflow

1. Step 1
2. Step 2

## Constraints

- Constraint 1
- Constraint 2

## Examples

### Example 1
User: "Do something"
You: [approach]
```

2. Test by invoking `/ent:newagent`

3. If agent works well, document in `docs/AGENTS_AND_SKILLS.md`

**See Also**: [AGENTS_AND_SKILLS.md](./AGENTS_AND_SKILLS.md) for complete v3 agent format

### Adding a New Skill

**v3 Format** - Markdown sections with YAML frontmatter (recommended)

1. Create `pkg/skills/{category}/skillname/SKILL.md`:

```markdown
---
name: skillname
description: "Skill description"
version: "1.0.0"
triggers:
  keywords:
    - trigger1
    - trigger2
    - trigger3
  file_pattern: "*.go"
  weight: 0.8
---

## Role

Expert persona definition.

## Instructions

### Pattern 1

Code examples and explanations.

## Constraints

- Include patterns
- Exclude anti-patterns

## Edge Cases

If X: Y format scenarios.

## Examples

### Example 1

**Input**: User request
**Output**: Expected response

## Output Format

Guidelines for output.
```

2. Validate: `make skill-validate strict=true`

3. Test by working on code that should trigger it

**See Also**: [SKILL-AUTHORING.md](./SKILL-AUTHORING.md) for complete v3 skill format

**Backward Compatibility**: v2 skills (XML tags) still work but v3 is recommended for new skills.

---

## Adding New Components

### Adding a New Skill

For complete skill authoring guide, see [SKILL-AUTHORING.md](./SKILL-AUTHORING.md).

**Quick Start**:
1. Create skill file at `pkg/skills/{category}/{skill-id}/SKILL.md`
2. Use v2 format with required XML sections
3. Validate: `ent skill validate pkg/skills/...`
4. Test with Claude Code

**Skill Categories**:
- `core/` - Cross-language skills (api-design, arch-core, etc.)
- `go/` - Go-specific skills (go-code, go-api, etc.)
- `plugins/` - Plugin development skills

For complete details on skill structure, validation, quality scoring, and best practices, see [SKILL-AUTHORING.md](./SKILL-AUTHORING.md).

---

### Adding a New Agent

For complete agent development guide, see [AGENTS_AND_SKILLS.md](./AGENTS_AND_SKILLS.md).

**Quick Start (v3 Split Format)**:
1. Create metadata: `pkg/agents/meta/<agent>.yaml`
2. Create prompt: `pkg/agents/prompts/agents/<agent>.md`
3. Use platform-agnostic model names (`main`/`fast`/`heavy`)
4. Configure `toolPresets` instead of explicit tools
5. List shared prompts in `prompts.shared` array
6. Rebuild: `make build`
7. Test: `go-ent init --tool claude && invoke /ent:<agent>`

**v3 Split Format Features**:
- Metadata (YAML) separated from prompts (Markdown)
- Platform-agnostic tool presets (`editing`, `readonly`, `planning`)
- Shared prompt sections reused across agents
- Dual-platform generation (Claude Code + OpenCode)
- Template-based rendering for each platform

### Adding a New Command

1. Create `pkg/commands/ent:newcmd.md`:
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
2. Check `.claude-plugin/plugin.json` exists
3. Restart Claude Code
4. Check Claude Code logs for plugin loading errors

**Fix:**
```bash
# Verify plugin.json is valid
cat .claude-plugin/plugin.json

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
- **Plugin Structure:** `README.md` (project root)
- **Skill Definitions:** `pkg/skills/*/SKILL.md`
- **Agent Definitions:** `pkg/agents/`
- **Command Definitions:** `pkg/commands/*.md`
- **Metrics System:** `docs/METRICS.md` (opt-out, data collection details)

## Version History

### v0.3.0 (Current)
- Initial self-hosted development setup
- 7 agents, 9 skills, 16 commands
- Full OpenSpec workflow integration
- Layered fallback architecture

---

## Template Development

> **Added:** February 2026 - Template system development guide

### Overview

The template system enables consistent agent generation through parameterized, role-specific sections.

### Directory Structure

```
pkg/agents/templates/
├── base-agent.md.tmpl          # Base template with slots
├── claude.yaml.tmpl             # Claude Code frontmatter
├── opencode.yaml.tmpl           # OpenCode frontmatter
└── sections/                    # Section templates
    ├── _tooling.md.tmpl         # Tool usage
    ├── _workflow.md.tmpl        # Context gathering
    ├── _principles.md.tmpl      # Constitutional AI
    └── _handoff.md.tmpl         # Agent delegation
```

### Creating a New Section Template

**1. Create Template File**

`pkg/agents/templates/sections/_mysection.md.tmpl`:

```go-template
## My Section Title

Universal content shown to all agents.

{{- if eq .Role "execution" }}
Execution-specific content.
{{- else if eq .Role "planning" }}
Planning-specific content.
{{- end }}

{{- if .HasSomeCondition }}
Conditional content.
{{- end }}
```

**2. Add Load Function**

Update `internal/cli/template_sections.go`:

```go
func loadAllSectionTemplates() (map[string]*template.Template, error) {
    sections := make(map[string]*template.Template)
    
    sectionNames := []string{
        "_tooling",
        "_workflow",
        "_principles",
        "_handoff",
        "_mysection",  // Add new section
    }
    
    // ... rest of function
}
```

**3. Add to Assembly Order**

Update `assembleAgentFromSections()`:

```go
sectionOrder := []string{
    "_tooling",
    "_workflow",
    "_mysection",    // Add in desired order
    "_principles",
    "_handoff",
}
```

**4. Write Tests**

`internal/cli/template_sections_test.go`:

```go
func TestRenderMySectionTemplate(t *testing.T) {
    tpl, err := loadSectionTemplate("_mysection")
    require.NoError(t, err)
    
    tests := []struct {
        name string
        data *AgentTemplateData
        want []string
    }{
        {
            name: "execution role",
            data: &AgentTemplateData{
                Role: "execution",
            },
            want: []string{
                "My Section Title",
                "Execution-specific content",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := renderSectionTemplate(tpl, tt.data)
            require.NoError(t, err)
            
            for _, want := range tt.want {
                assert.Contains(t, got, want)
            }
        })
    }
}
```

**5. Test and Verify**

```bash
# Run tests
go test ./internal/cli -v -run "TestRenderMySection"

# Regenerate agents
make build
./bin/ent init --tools=claude --force

# Verify output
cat .claude/agents/ent/coder.md | grep -A10 "My Section Title"
```

### Modifying Existing Templates

**1. Edit Template**

Make changes to `pkg/agents/templates/sections/*.tmpl`

**2. Update Tests**

Adjust tests to match new content/structure

**3. Run Tests**

```bash
go test ./internal/cli -v
```

**4. Regenerate and Verify**

```bash
# Backup
cp -r .claude/agents/ent/ /tmp/agents-backup/

# Regenerate
make build
./bin/ent init --tools=claude --force

# Compare
diff -r .claude/agents/ent/ /tmp/agents-backup/
```

### Template Parameters

Available in all templates via `AgentTemplateData`:

| Parameter | Type | Example | Usage |
|-----------|------|---------|-------|
| Name | string | "coder" | `{{ .Name }}` |
| Role | string | "execution" | `{{- if eq .Role "execution" }}` |
| RoleTitle | string | "Implementation" | `{{ .RoleTitle }}` |
| Dependencies | []string | ["tester"] | `{{- range .Dependencies }}` |
| DisallowedTools | []string | ["tool1"] | `{{- range .DisallowedTools }}` |
| HasDisallowedTools | bool | true | `{{- if .HasDisallowedTools }}` |

### Template Helpers

Available functions:

```go-template
{{ .Name | title }}        # Title case
{{ .Name | upper }}        # Uppercase
{{ .Name | lower }}        # Lowercase
{{ contains .Name "cod" }} # Substring check
{{ hasPrefix .Name "co" }} # Prefix check
{{ hasSuffix .Name "er" }} # Suffix check
{{ join .Skills ", " }}    # Join array
{{ replace .Name "o" "0" 2 }} # Replace all
```

### Best Practices

**1. Role Parameterization**
```go-template
{{- if eq .Role "execution" }}
Concrete code patterns
{{- else if eq .Role "planning" }}
Architecture guidance
{{- else if eq .Role "validation" }}
Review criteria
{{- end }}
```

**2. Conditional Sections**
```go-template
{{- if .HasDisallowedTools }}
## CRITICAL: Tool Restrictions
{{- range .DisallowedTools }}
- ❌ `{{ . }}`
{{- end }}
{{- end }}
```

**3. Safe Array Iteration**
```go-template
{{- if .Dependencies }}
## Handoff
{{- range .Dependencies }}
- @ent/{{ . }}
{{- end }}
{{- end }}
```

**4. Testing Strategy**
- Test each role variant
- Test conditional rendering
- Test with/without optional fields
- Test array iteration edge cases

### Common Patterns

**Progressive Disclosure:**
```go-template
## Section Title

[Always shown]

{{- if condition }}
[Conditionally shown]
{{- end }}
```

**Role-Based Content:**
```go-template
{{- if eq .Role "execution" }}
[Execution content]
{{- else if eq .Role "planning" }}
[Planning content]
{{- else if eq .Role "validation" }}
[Validation content]
{{- else if eq .Role "research" }}
[Research content]
{{- end }}
```

**Array Safety:**
```go-template
{{- if .Items }}
{{- range .Items }}
- {{ . }}
{{- end }}
{{- end }}
```

### Troubleshooting

**Template Parse Errors:**
```
Error: template: section:5: unexpected "}" in operand
```
Fix: Check conditional syntax, ensure `{{- end }}` matches `{{- if }}`

**Missing Parameters:**
```
Error: can't evaluate field Role in type *AgentTemplateData
```
Fix: Add parameter to `AgentTemplateData` struct in `template_helpers.go`

**Test Failures:**
```
Error: output missing expected string: "foo"
```
Fix: Update test expectations to match new template output

### Integration with Agent Generation

**Flow:**
1. Load agent metadata (YAML)
2. Load agent-specific prompt (Markdown)
3. Load section templates
4. Render sections with agent data
5. Assemble: metadata + content + sections
6. Write to `.claude/agents/ent/`

**Code Entry Point:**

`internal/cli/init.go`:
```go
// Load templates
sections, err := loadAllSectionTemplates()

// Create template data
data := NewAgentTemplateData(meta, agentContent, sharedPrompts)

// Assemble
content, err := assembleAgentFromSections(data, sections)
```

### See Also

- [PROMPT_DESIGN.md](PROMPT_DESIGN.md) - Template design patterns
- [AGENTS_AND_SKILLS.md](AGENTS_AND_SKILLS.md) - Agent system overview
- `internal/cli/template_sections.go` - Template implementation
- `internal/cli/template_sections_test.go` - Template tests
