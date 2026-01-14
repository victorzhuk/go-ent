# Agent Format Compatibility: OpenCode vs Claude Code

## Format Comparison

| Feature | OpenCode | Claude Code |
|---------|----------|-------------|
| **File location** | `.opencode/agent/` | `.claude/agents/` |
| **Tools format** | Object: `tools: { read: true }` | String: `tools: Read, Grep` |
| **Tool names** | lowercase: `read`, `write`, `edit` | PascalCase: `Read`, `Write`, `Edit` |
| **Model reference** | Tier: `model: main/fast/heavy` | Alias: `model: sonnet/opus/haiku/inherit` |
| **Mode** | `mode: primary/subagent/all` | Not used (all are subagents) |
| **Permissions** | `permission: { bash: { ... } }` | `permissionMode: default/acceptEdits/bypassPermissions` |
| **Skills** | Array: `skills: [go-code, go-db]` | String: `skills: go-code, go-db` |
| **MCP tools** | `mcp__plugin_name: true` | Inherited from main thread |
| **Denylist** | Not directly supported | `disallowedTools: Write, Edit` |
| **Tags** | `tags: [role:execution]` | Not supported |
| **Color** | `color: "#32CD32"` | `color: green` (via /agents UI) |

---

## Tool Name Mapping

| OpenCode | Claude Code |
|----------|-------------|
| `read` | `Read` |
| `write` | `Write` |
| `edit` | `Edit` |
| `bash` | `Bash` |
| `grep` | `Grep` |
| `glob` | `Glob` |
| `list` | `LS` |
| `webfetch` | `WebFetch` |
| `websearch` | `WebSearch` |
| `todoread` | `TodoRead` |
| `todowrite` | `TodoWrite` |
| `skill` | (auto-loaded via `skills:` field) |
| `task` | `Task` |
| `patch` | `MultiEdit` |
| `multiedit` | `MultiEdit` |

---

## Model Mapping

| OpenCode | Claude Code | Actual Model |
|----------|-------------|--------------|
| `fast` | `haiku` | claude-haiku |
| `main` | `sonnet` | claude-sonnet |
| `heavy` | `opus` | claude-opus |
| - | `inherit` | Same as parent |

---

## Compatibility Strategies

### Strategy 1: Dual Files (Recommended)

Maintain separate files for each platform:

```
project/
├── .opencode/
│   └── agent/
│       ├── coder.md      # OpenCode format
│       └── reviewer.md
├── .claude/
│   └── agents/
│       ├── coder.md      # Claude Code format
│       └── reviewer.md
```

**Pros:** Native support, full features
**Cons:** Duplication, maintenance overhead

### Strategy 2: Symlinks with Preprocessing

Use a build script to generate platform-specific files:

```bash
# generate-agents.sh
for agent in agents/*.template.md; do
  name=$(basename "$agent" .template.md)
  
  # Generate OpenCode version
  sed -e 's/Read/read/g' -e 's/Write/write/g' \
      -e 's/model: sonnet/model: main/g' \
      "$agent" > ".opencode/agent/${name}.md"
  
  # Generate Claude Code version
  sed -e 's/read: true/Read/g' \
      "$agent" > ".claude/agents/${name}.md"
done
```

### Strategy 3: Claude Code with AGENTS.md Reference

Claude Code can reference AGENTS.md via CLAUDE.md:

```markdown
# CLAUDE.md
Read and follow instructions in AGENTS.md for project conventions.
```

This allows sharing high-level instructions but not agent definitions.

---

## Recommended Dual-Format Structure

### OpenCode Format (`.opencode/agent/coder.md`)

```yaml
---
name: coder
description: "Go developer. Implements features, writes code."
tools:
  read: true
  write: true
  edit: true
  bash: true
  glob: true
  grep: true
  list: true
  todoread: true
  todowrite: true
  skill: true
model: main
tags:
  - "role:execution"
skills:
  - go-code
  - go-db
---
```

### Claude Code Format (`.claude/agents/coder.md`)

```yaml
---
name: coder
description: Go developer. Implements features, writes code.
tools: Read, Write, Edit, Bash, Glob, Grep, LS, TodoRead, TodoWrite
model: sonnet
skills: go-code, go-db
---
```

---

## Common Prompt Body (Works for Both)

The system prompt body (after frontmatter) is **identical** for both platforms. Only the YAML frontmatter differs.

This means you can:
1. Write the prompt body once
2. Use different frontmatter headers
3. Concatenate them during build

---

## Migration Script

```bash
#!/bin/bash
# migrate-opencode-to-claude.sh

convert_tools() {
  local tools="$1"
  # Parse YAML object and convert to comma-separated PascalCase
  echo "$tools" | grep -oP '\w+(?=:\s*true)' | \
    sed 's/read/Read/; s/write/Write/; s/edit/Edit/; s/bash/Bash/; 
         s/grep/Grep/; s/glob/Glob/; s/list/LS/; 
         s/todoread/TodoRead/; s/todowrite/TodoWrite/' | \
    tr '\n' ', ' | sed 's/,$//'
}

convert_model() {
  case "$1" in
    "fast") echo "haiku" ;;
    "main") echo "sonnet" ;;
    "heavy") echo "opus" ;;
    *) echo "sonnet" ;;
  esac
}

# Usage: ./migrate-opencode-to-claude.sh .opencode/agent/coder.md
```

---

## Validation Checklist

When creating dual-format agents:

### OpenCode
- [ ] `tools` is YAML object with boolean values
- [ ] Tool names are lowercase
- [ ] `model` uses tier names (fast/main/heavy)
- [ ] `skills` is YAML array
- [ ] `mode` specified if needed
- [ ] MCP tools explicitly listed

### Claude Code
- [ ] `tools` is comma-separated string
- [ ] Tool names are PascalCase
- [ ] `model` uses aliases (haiku/sonnet/opus/inherit)
- [ ] `skills` is comma-separated string
- [ ] `permissionMode` set if needed
- [ ] MCP tools inherited automatically

---

## Feature Parity Notes

### OpenCode-Only Features
- `mode: primary` (Tab switching)
- `tags` for categorization
- Fine-grained `permission` rules
- `color` in hex format
- `temperature` setting

### Claude Code-Only Features
- `permissionMode: bypassPermissions`
- `disallowedTools` denylist
- `inherit` model option
- Automatic MCP tool inheritance
- `/agents` management UI
- Resumable agents with `agentId`

---

## Best Practice: Template Approach

Create a template with placeholders:

```markdown
---
name: {{NAME}}
description: {{DESCRIPTION}}
{{#OPENCODE}}
tools:
  read: true
  write: true
  edit: true
  bash: true
  grep: true
  glob: true
model: {{MODEL_OPENCODE}}
skills:
  - {{SKILLS}}
{{/OPENCODE}}
{{#CLAUDE}}
tools: Read, Write, Edit, Bash, Grep, Glob
model: {{MODEL_CLAUDE}}
skills: {{SKILLS_COMMA}}
{{/CLAUDE}}
---

{{PROMPT_BODY}}
```

Use a templating engine (mustache, envsubst, etc.) to generate both formats.
