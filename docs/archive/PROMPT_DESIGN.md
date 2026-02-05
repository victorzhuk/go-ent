# Prompt Design Guide

> **Updated:** February 2026 - Template-Based Architecture

Design patterns and best practices for go-ent's template-based prompt system.

---

## Table of Contents

- [Template Architecture](#template-architecture)
- [Constitutional AI Principles](#constitutional-ai-principles)
- [Section Templates](#section-templates)
- [Agent-Specific Prompts](#agent-specific-prompts)
- [Shared Prompts](#shared-prompts)
- [Template Parameters](#template-parameters)
- [Best Practices](#best-practices)
- [Testing](#testing)

---

## Template Architecture

### Overview

go-ent uses a **slot-based template composition** system:

```
Final Agent = Base Template + Section Templates + Agent Content + Shared Prompts
```

This architecture provides:
- **Structural consistency** across all agents
- **Role-specific content** through parameterization
- **Conditional rendering** based on agent configuration
- **Single source of truth** for shared guidance

### Components

1. **Section Templates** (`pkg/agents/templates/sections/*.tmpl`)
   - Reusable, parameterized sections
   - Role-specific variants
   - Conditional logic

2. **Agent Prompts** (`pkg/agents/prompts/agents/*.md`)
   - Agent-specific content only
   - Responsibilities, patterns, examples
   - ~50-80 lines (down from ~200)

3. **Shared Prompts** (`pkg/prompts/*.md`)
   - Embedded in generated agents
   - Mapped to skills via `sharedPromptToSkill`
   - Foundation, workflow, tooling

4. **Base Template** (`pkg/agents/templates/base-agent.md.tmpl`)
   - Assembles all components
   - Manages embedding order

---

## Constitutional AI Principles

### Philosophy

Our prompt design follows Constitutional AI principles adapted for autonomous development agents:

- **Judgment over rigidity** - Agents think like senior developers who understand context
- **Context awareness** - Rules are starting points, not absolute laws
- **Principal hierarchy** - Clear conflict resolution framework
- **Responsibility with checkpoints** - Irreversible actions require verification

### Thoughtful Senior Developer Standard

**The Test**: Would a senior professional with 10+ years experience make this same decision in this exact context? If yes, proceed. If no, reconsider.

**Behavioral Guidelines:**
- Prefer clarity over cleverness
- Choose progress over perfection
- Document unusual decisions
- Ask when genuinely uncertain
- Own decisions with clear reasoning

### Principal Hierarchy

When values conflict, apply in order:

1. **Project conventions** - Established patterns in THIS codebase
2. **User intent** - What the human actually wants/needs
3. **Best practices** - Industry standards and proven patterns
4. **Safety** - Security, data integrity, production stability
5. **Simplicity** - KISS, YAGNI, avoid over-engineering

### Irreversible Action Checkpoints

Before actions that cannot be easily undone:
- Code deletions: Verify with user or provide clear warning
- Force operations: Explain risks and get confirmation
- Breaking changes: Explicitly flag and verify
- Major refactors: Ensure backup/revert path exists

---

## Section Templates

### _tooling.md.tmpl

**Purpose:** Tool usage guidance with conditional restrictions

**Parameters Used:**
- `.HasDisallowedTools` - Controls CRITICAL section visibility
- `.DisallowedTools` - List of forbidden tools

**Design Pattern:**
```go-template
## Optimal Tooling

[Universal tool guidance - always shown]

{{- if .HasDisallowedTools }}
## CRITICAL: Tool Restrictions

**NEVER use:**
{{- range .DisallowedTools }}
- ❌ `{{ . }}`
{{- end }}
{{- end }}
```

**Best Practices:**
- Keep tool alternatives list stable
- Only show restrictions when applicable
- Use clear, unmistakable warnings for forbidden tools

### _workflow.md.tmpl

**Purpose:** Context gathering and role-specific workflow

**Parameters Used:**
- `.Role` - Determines workflow variant (execution, planning, validation, research)

**Design Pattern:**
```go-template
### 1. Context Gathering

[Universal steps - always shown]

{{- if eq .Role "execution" }}
[Execution-specific workflow]
{{- else if eq .Role "planning" }}
[Planning-specific workflow]
{{- else if eq .Role "validation" }}
[Validation-specific workflow]
{{- end }}
```

**Best Practices:**
- Start with common steps applicable to all roles
- Branch only where workflows diverge significantly
- Keep role-specific content focused and actionable

### _principles.md.tmpl

**Purpose:** Constitutional AI judgment framework, role-specific

**Parameters Used:**
- `.RoleTitle` - Human-readable role name
- `.Role` - Determines examples and guidance

**Design Pattern:**
```go-template
## Constitutional AI Principles

### Judgment for {{ .RoleTitle }}

Exercise judgment as a thoughtful senior {{ .RoleTitle }} agent.

{{- if eq .Role "execution" }}
**Implementation Judgment Examples:**
[Execution-specific scenarios]
{{- else if eq .Role "planning" }}
**Planning Judgment Examples:**
[Planning-specific scenarios]
{{- end }}

[Universal principal hierarchy]

{{- if eq .Role "execution" }}
**Coding Conflict Examples:**
[Execution-specific conflicts]
{{- else if eq .Role "planning" }}
**Planning Conflict Examples:**
[Planning-specific conflicts]
{{- end }}
```

**Best Practices:**
- Keep core principles consistent across roles
- Parameterize examples, not principles
- Make role-specific guidance concrete and actionable

### _handoff.md.tmpl

**Purpose:** Agent delegation patterns

**Parameters Used:**
- `.Dependencies` - Array of agents to delegate to

**Design Pattern:**
```go-template
{{- if .Dependencies }}
## Handoff

{{- range .Dependencies }}
- @ent/{{ . }} - [purpose]
{{- end }}
{{- end }}
```

**Best Practices:**
- Only render when dependencies exist
- Provide clear delegation purposes
- Keep handoff instructions concise

---

## Agent-Specific Prompts

### Structure

Agent prompts contain ONLY agent-unique content:

```markdown
You are a [role description].

## Responsibilities
- Specific responsibility 1
- Specific responsibility 2

## Patterns

### Pattern 1
[Code example]

### Pattern 2
[Code example]
```

### What to Include

**DO include:**
- Role-specific introduction
- Concrete responsibilities
- Code patterns and examples
- Agent-unique workflows

**DON'T include:**
- Tool usage guidance (in _tooling.tmpl)
- Constitutional AI principles (in _principles.tmpl)
- OpenSpec workflow (in _workflow.tmpl)
- Delegation patterns (in _handoff.tmpl)

### Size Guidelines

- **Target:** 50-80 lines
- **Maximum:** 100 lines
- If exceeding 100 lines, consider:
  - Moving patterns to a skill
  - Extracting common content to section template
  - Simplifying or removing redundant guidance

### Quality Checklist

- [ ] Starts with clear role introduction
- [ ] Lists specific, actionable responsibilities
- [ ] Provides concrete code examples
- [ ] No duplication with section templates
- [ ] No constitutional AI or workflow guidance
- [ ] Focused on "what makes this agent unique"

---

## Shared Prompts

### Purpose

Shared prompts are embedded into generated agents and mapped to skills for runtime loading.

### Files

1. **foundation.md** - Core decision framework
   - Constitutional AI principles (condensed)
   - Principal hierarchy
   - Go conventions summary
   - ~90 lines

2. **workflow.md** - OpenSpec workflow essentials
   - Quick command reference
   - Common workflow patterns
   - ~60 lines

3. **tooling.md** - Tool usage guidance
   - Modern CLI alternatives
   - Native vs. MCP tools
   - ~90 lines

### Design Principles

**Concise but complete:**
- Shared prompts are embedded, so brevity matters
- Include essential guidance only
- Reference skills for full details

**Self-contained:**
- Each shared prompt must work standalone
- Don't assume other shared prompts are read

**Mapping to skills:**
```go
var sharedPromptToSkill = map[string]string{
    "_foundation": "ent-foundation",
    "_workflow":   "ent-workflow",
    "_tooling":    "ent-tooling",
}
```

Agent metadata references `_foundation`, generated agent loads `ent-foundation` skill.

---

## Template Parameters

### AgentTemplateData Structure

```go
type AgentTemplateData struct {
    // Identity
    Name         string
    Description  string

    // Role information
    Role         string   // execution, planning, validation, research
    RoleTitle    string   // Implementation, Planning, Validation, Research
    Complexity   string   // fast, standard, heavy

    // Relationships
    Dependencies []string // Agents to delegate to
    Skills       []string // Skills loaded by agent

    // Tools
    AllowedTools      []string
    DisallowedTools   []string
    HasDisallowedTools bool

    // Content
    AgentContent   string   // Agent-specific prompt
    SharedPrompts  []string // Embedded shared prompts
}
```

### Parameter Usage Patterns

**Role-based branching:**
```go-template
{{- if eq .Role "execution" }}
[Execution content]
{{- else if eq .Role "planning" }}
[Planning content]
{{- end }}
```

**Conditional sections:**
```go-template
{{- if .HasDisallowedTools }}
[Tool restrictions]
{{- end }}

{{- if .Dependencies }}
[Handoff guidance]
{{- end }}
```

**Array iteration:**
```go-template
{{- range .DisallowedTools }}
- ❌ `{{ . }}`
{{- end }}

{{- range .Dependencies }}
- @ent/{{ . }}
{{- end }}
```

**String manipulation:**
```go-template
{{ .RoleTitle | upper }}
{{ .Name | title }}
```

---

## Best Practices

### Template Design

1. **Progressive Disclosure**
   - Start with common content
   - Branch for role-specific variants
   - Use conditionals for optional sections

2. **Single Responsibility**
   - Each template has one clear purpose
   - Don't mix concerns (tooling + workflow in one template)

3. **Fail-Safe Defaults**
   - Templates should work with minimal parameters
   - Use conditional checks before accessing arrays
   - Provide fallbacks for optional fields

4. **Testing-Friendly**
   - Design templates to be easily testable
   - Make parameters explicit
   - Avoid complex nested conditionals

### Content Design

1. **Clarity over Cleverness**
   - Prefer explicit over implicit
   - Use natural language, not jargon
   - Make intent obvious

2. **Actionable Guidance**
   - Every instruction should be concrete
   - Provide examples, not just descriptions
   - Focus on "how" not just "what"

3. **Role-Appropriate Content**
   - Execution agents: concrete code patterns
   - Planning agents: architecture guidance
   - Validation agents: review criteria
   - Research agents: investigation techniques

4. **Consistency**
   - Use same terminology across templates
   - Follow same structure patterns
   - Maintain consistent tone

### Maintenance

1. **Update Coordination**
   - Skills and shared prompts must stay in sync
   - Test template changes with all agents
   - Update tests when templates change

2. **Version Control**
   - Regenerate agents after template changes
   - Commit generated output to verify diffs
   - Document breaking changes

3. **Documentation**
   - Keep this guide updated with template changes
   - Document new parameters added
   - Explain design decisions in comments

---

## Testing

### Unit Tests

Test template loading and rendering:

```go
func TestRenderSectionTemplate(t *testing.T) {
    tpl, err := loadSectionTemplate("_tooling")
    require.NoError(t, err)

    data := &AgentTemplateData{
        Role: "execution",
        DisallowedTools: []string{"tool1", "tool2"},
        HasDisallowedTools: true,
    }

    output, err := renderSectionTemplate(tpl, data)
    require.NoError(t, err)
    assert.Contains(t, output, "CRITICAL: Tool Restrictions")
    assert.Contains(t, output, "tool1")
}
```

### Integration Tests

Test complete agent generation:

```bash
# Backup current
cp -r .claude/agents/ent/ /tmp/agents-backup/

# Regenerate
make build
./bin/ent init --tools=claude --force

# Verify no unexpected changes
diff -r .claude/agents/ent/ /tmp/agents-backup/
```

### Validation Tests

Verify template quality:

```bash
# Check all templates load
go test ./internal/cli -v -run "TestLoadSection"

# Check role variants
go test ./internal/cli -v -run "TestRenderWorkflow"

# Check parameterization
go test ./internal/cli -v -run "TestRenderPrinciples"

# Check assembly
go test ./internal/cli -v -run "TestAssemble"
```

### Manual Testing

After template changes:

1. Regenerate agents
2. Review generated output
3. Test with actual agent usage
4. Verify no behavior regressions
5. Check skill loading works

---

## Migration from Old System

### Before (Pre-Refactoring)

Agent prompts were ~200 lines with significant duplication:
- Each agent embedded constitutional AI guidance
- Tool usage guidance repeated across agents
- Workflow patterns duplicated
- No parameterization by role

### After (Template System)

Agent prompts are ~60 lines, focused content:
- Constitutional AI in section template (role-parameterized)
- Tool usage in section template (conditionally rendered)
- Workflow in section template (role-specific)
- Agent prompts contain only unique content

### Benefits Achieved

- **65% reduction** in maintenance burden
- **100% elimination** of duplication
- **Role-appropriate** guidance automatically
- **Consistent quality** across all agents
- **Single source of truth** for shared content

---

## See Also

- [AGENTS_AND_SKILLS.md](AGENTS_AND_SKILLS.md) - Agent system overview
- [DEVELOPMENT.md](DEVELOPMENT.md) - Template development guide
- [SKILL-AUTHORING.md](SKILL-AUTHORING.md) - Skill creation guide
- [AGENT_INHERITANCE.md](AGENT_INHERITANCE.md) - Agent inheritance patterns
