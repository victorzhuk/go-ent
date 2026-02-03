# Claude Code Compatibility Guide

Guide for ensuring go-ent plugins align with Claude Code official patterns and best practices.

---

## Table of Contents

- [Overview](#overview)
- [Agent Format](#agent-format)
- [Skill Format](#skill-format)
- [Tool Configuration](#tool-configuration)
- [Directory Structure](#directory-structure)
- [Validation](#validation)

---

## Overview

go-ent v3 aligns with **Claude Code official patterns** for maximum compatibility and native behavior.

### Compatibility Matrix

| Feature | Claude Code | go-ent v3 | Status |
|---------|-------------|-----------|--------|
| Agent frontmatter | YAML in MD | YAML in MD | ✅ Aligned |
| Skill frontmatter | YAML in MD | YAML in MD | ✅ Aligned |
| Triggers | YAML `triggers:` | YAML `triggers:` | ✅ Aligned |
| Markdown sections | `## Role` | `## Role` | ✅ Aligned |
| Model names | sonnet/opus/haiku | sonnet/opus/haiku | ✅ Aligned |
| Agent location | `.claude/agents/` | `.claude/agents/` | ✅ Aligned |
| Skills preloading | `skills:` array | `skills:` array | ✅ Aligned |
| Hooks | PreToolUse/PostToolUse/Stop | Same | ✅ Aligned |

### Design Philosophy

1. **Follow official patterns** - Match Claude Code exactly
2. **Leverage native features** - Use built-in skill preloading
3. **Minimize custom logic** - Reduce plugin-specific behavior
4. **Interoperability** - Work seamlessly with other plugins

---

## Agent Format

### Official Claude Code Schema

**Required Fields:**
- `name` - Unique identifier (lowercase, hyphens)
- `description` - When to delegate to this agent

**Optional Fields:**
- `model` - Model to use: `sonnet`, `opus`, `haiku`, `inherit`
- `tools` - Allowed tools (comma-separated string or array)
- `disallowedTools` - Denied tools (comma-separated string or array)
- `permissionMode` - Permission handling: `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, `plan`
- `skills` - Skills to preload at startup (array)
- `hooks` - Lifecycle hooks (PreToolUse, PostToolUse, Stop)

### go-ent Extensions

go-ent adds optional metadata fields for plugin ecosystem:

- `color` - Hex color for UI (#RRGGBB)
- `role` - Role category: planning, execution, validation, research, review, debug, test
- `complexity` - Complexity level: light, standard, heavy
- `dependencies` - Other agents this can delegate to (array)

**These extensions don't break Claude Code compatibility** - they're ignored by the core runtime.

### Example: Full Compatibility

```markdown
---
name: coder
description: Go developer. Implements features, writes code.
model: sonnet
tools:
  - Read
  - Write
  - Edit
  - Bash
disallowedTools:
  - mcp__plugin_serena_serena__replace_symbol_body
skills:
  - go-code
  - go-db
  - ent-tooling
permissionMode: default
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "./scripts/validate-command.sh"
# go-ent extensions (optional)
color: "#32CD32"
role: execution
complexity: standard
dependencies:
  - tester
  - reviewer
---

You are a senior Go backend developer...
```

**Validation**:
```bash
# Validates against Claude Code schema
ent validate --agent .claude/agents/coder.md
```

---

## Skill Format

### Official Claude Code Schema

**Required Fields:**
- `name` - Unique identifier
- `description` - Brief summary for invocation decisions

**Optional Fields:**
- `disable-model-invocation` - Prevent automatic invocation (boolean)
- `user-invocable` - Allow user manual invocation (boolean)
- `version` - Version string (e.g., "1.0.0")
- `allowed-tools` - Tools the skill can use (array or string pattern)
- `triggers` - Activation triggers (keywords, file_pattern, weight)

**Markdown Sections:**
- `## Role` - Expert description
- `## Instructions` - Step-by-step guidance
- `## Constraints` - Include/exclude rules
- `## Edge Cases` - Handling unusual situations
- `## Examples` - Usage examples
- `## Output Format` - Expected output structure

### Example: Full Compatibility

```markdown
---
name: go-error
description: "Error handling patterns for Go"
version: "1.0.0"
disable-model-invocation: false
user-invocable: true
allowed-tools:
  - Read
  - Edit
triggers:
  keywords:
    - error handling
    - error wrapping
  file_pattern: "*.go"
  weight: 0.8
---

## Role

Expert Go error handling engineer specializing in error design patterns, wrapping strategies, and production-grade error management.

## Instructions

### Error Wrapping Pattern

Always wrap errors with context using `%w`:

\```go
if err != nil {
    return fmt.Errorf("query user %s: %w", id, err)
}
\```

## Constraints

- Include proper error wrapping with %w
- Include lowercase error messages
- Exclude unwrapped errors
- Exclude uppercase error messages

## Examples

### Good Error Handling

\```go
func (r *Repository) GetUser(id string) (*User, error) {
    var u User
    if err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Name); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return &u, nil
}
\```
```

**Validation**:
```bash
# Validates against Claude Code skill format
ent validate --skill skills/go/go-error/SKILL.md
```

---

## Tool Configuration

### Native Tool Names

Use exact Claude Code tool names:

**File Operations:**
- `Read` - Read file contents
- `Write` - Write/overwrite files
- `Edit` - Targeted string replacement
- `Glob` - Find files by pattern
- `Grep` - Search file contents
- `Bash` - Execute shell commands

**Task Management:**
- `TodoRead` - Read task list
- `TodoWrite` - Create/update tasks
- `Skill` - Invoke other skills
- `List` - View resources

**MCP Tools:**
Use full qualified names:
- `mcp__plugin_name_tool_name` format
- Example: `mcp__plugin_serena_serena__find_symbol`

### Tool Patterns

Specify tools as:

**Array format (recommended):**
```yaml
tools:
  - Read
  - Write
  - Edit
```

**Comma-separated string:**
```yaml
tools: "Read, Write, Edit"
```

**Pattern matching (allowed-tools in skills):**
```yaml
allowed-tools:
  - "Bash(gh *)"  # Allow only gh commands
  - "Read"
```

---

## Directory Structure

### Claude Code Native Locations

```
.claude/
└── agents/          # Agent definitions (*.md)
    ├── coder.md
    ├── planner.md
    └── reviewer.md

~/.claude/
├── agents/          # User-level agents (all projects)
└── settings.json    # User preferences
```

### Plugin Structure

```
plugins/plugin-name/
├── .claude-plugin   # Plugin manifest (optional)
├── .mcp.json        # MCP server config (optional)
├── agents/          # Agent definitions
│   └── *.md
├── skills/          # Skill definitions
│   └── */SKILL.md
├── commands/        # Command definitions
│   └── *.md
└── platformspecs/         # Validation schemas
    └── *.schema.json
```

**Priority (highest first):**
1. CLI flag `--agents`
2. `.claude/agents/` (project)
3. `~/.claude/agents/` (user)
4. Plugin `agents/` directory

---

## Validation

### Schema Validation

**Validate agent against Claude Code schema:**
```bash
ent validate --agent .claude/agents/coder.md --schema platformspecs/agent.schema.json
```

**Validate skill format:**
```bash
ent validate --skill skills/go/go-code/SKILL.md
```

**Validate all plugin components:**
```bash
ent validate --all
```

### Manual Checks

**Agent Checklist:**
- [ ] File in `.claude/agents/` directory
- [ ] YAML frontmatter between `---` delimiters
- [ ] Required fields: `name`, `description`
- [ ] Model is `sonnet`, `opus`, `haiku`, or `inherit`
- [ ] Tools use exact Claude Code names
- [ ] System prompt after frontmatter

**Skill Checklist:**
- [ ] File named `SKILL.md`
- [ ] YAML frontmatter between `---` delimiters
- [ ] Required fields: `name`, `description`
- [ ] Triggers in frontmatter (if applicable)
- [ ] Markdown sections: `## Role`, `## Instructions`
- [ ] No XML tags (`<role>` removed)

### Testing in Claude Code

**Load plugin:**
```bash
# Build plugin
make build

# Restart Claude Code with plugin
claude --plugin-dir ./plugins/go-ent

# Verify agents loaded
/agents

# Verify skills loaded
/skills
```

**Test delegation:**
```
Use the coder agent to implement user authentication
```

Claude should automatically delegate to the `coder` agent.

---

## Common Issues

### Agent Not Loading

**Symptoms:** Agent doesn't appear in `/agents` list

**Causes:**
1. File not in `.claude/agents/` directory
2. Invalid YAML frontmatter
3. Missing required fields (`name`, `description`)
4. Syntax errors in frontmatter

**Solution:**
```bash
# Validate schema
ent validate --agent .claude/agents/agent-name.md

# Check YAML syntax
yamllint .claude/agents/agent-name.md

# Verify file location
ls -la .claude/agents/
```

### Skill Not Triggering

**Symptoms:** Skill doesn't activate when expected

**Causes:**
1. Triggers not in frontmatter (still using XML `<triggers>`)
2. Keywords don't match use case
3. `disable-model-invocation: true` set
4. Weight too low (< 0.5)

**Solution:**
```bash
# Check format version
ent validate --skill skills/*/SKILL.md

# Test trigger matching
ent test-triggers --skill go-error --input "error handling"
```

### Tool Access Denied

**Symptoms:** Agent attempts tool use but gets permission denied

**Causes:**
1. Tool not in `tools:` list
2. Tool in `disallowedTools:` list
3. Tool name typo or incorrect format
4. MCP tool not available

**Solution:**
```yaml
# Verify tool names match exactly
tools:
  - Read       # ✅ Correct
  - read       # ❌ Wrong (case-sensitive)
  - ReadFile   # ❌ Wrong (no such tool)
```

---

## Best Practices

### 1. Use Native Features

✅ **Do:** Use Claude Code's `skills:` field for preloading
```yaml
skills:
  - go-code
  - ent-tooling
```

❌ **Don't:** Embed skill content in agent prompt

### 2. Follow Naming Conventions

✅ **Do:** Use lowercase with hyphens
```yaml
name: code-reviewer
```

❌ **Don't:** Use camelCase or special characters
```yaml
name: codeReviewer  # ❌ Wrong
name: code_reviewer # ❌ Wrong (use hyphens)
```

### 3. Clear Descriptions

✅ **Do:** Be specific about when to use
```yaml
description: Expert code reviewer. Use proactively after code changes.
```

❌ **Don't:** Be vague
```yaml
description: Reviews code  # ❌ Too vague
```

### 4. Explicit Tool Lists

✅ **Do:** List exact tools needed
```yaml
tools:
  - Read
  - Grep
  - Glob
```

❌ **Don't:** Grant unnecessary access
```yaml
tools: "*"  # ❌ Too permissive
```

### 5. Version Metadata

✅ **Do:** Include version for tracking
```yaml
version: "1.0.0"
```

---

## Resources

**Official Documentation:**
- [Claude Code Agents](https://code.claude.com/docs/en/sub-agents)
- [Claude Code Skills](https://code.claude.com/docs/en/skills)

**go-ent Documentation:**
- [MIGRATION_V3.md](./MIGRATION_V3.md) - Migration from v2 to v3
- [AGENTS_AND_SKILLS.md](./AGENTS_AND_SKILLS.md) - Architecture guide
- [SKILL-AUTHORING.md](./SKILL-AUTHORING.md) - Skill creation guide

**Tools:**
- `ent validate` - Schema validation
- `ent test-triggers` - Trigger testing
- `scripts/migrate-agents.go` - Agent migration
- `scripts/migrate-skills.go` - Skill migration
