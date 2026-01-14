# Multi-Platform Agent Setup

This guide explains how to set up agents for both **OpenCode** and **Claude Code**.

---

## Directory Structure

```
project/
├── .opencode/
│   └── agent/
│       ├── driver.md       # Orchestrator (primary)
│       ├── coder.md        # Implementation
│       ├── debugger.md     # Bug investigation
│       ├── planner.md      # Task planning
│       ├── reviewer.md     # Code review
│       ├── researcher.md   # Deep analysis
│       ├── architect.md    # System design
│       ├── decomposer.md   # Task breakdown
│       ├── acceptor.md     # Acceptance testing
│       └── tester.md       # Test execution
├── .claude/
│   └── agents/
│       ├── coder.md        # Same agents,
│       ├── debugger.md     # different format
│       ├── planner.md
│       ├── reviewer.md
│       └── researcher.md
├── opencode.json           # OpenCode config
├── CLAUDE.md               # Claude Code config
└── AGENTS.md               # Shared instructions
```

---

## OpenCode Setup

### 1. Create Agent Directory

```bash
mkdir -p .opencode/agent
```

### 2. Copy OpenCode-format agents

Place `.md` files with this frontmatter style:

```yaml
---
name: coder
description: "Go developer. Implements features."
tools:
  read: true
  write: true
  edit: true
  bash: true
  grep: true
  glob: true
  list: true
  todoread: true
  todowrite: true
  skill: true
model: main
skills:
  - go-code
  - go-db
---
```

### 3. Configure opencode.json

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5-20250929",
  "agent": {
    "driver": {
      "description": "Orchestrator - coordinates tasks",
      "mode": "primary",
      "prompt": "{file:.opencode/agent/driver.md}",
      "tools": {
        "read": true,
        "grep": true,
        "glob": true,
        "list": true,
        "todoread": true,
        "todowrite": true,
        "skill": true,
        "task": true,
        "webfetch": true,
        "websearch": true
      },
      "permission": {
        "edit": "deny",
        "bash": "deny"
      }
    }
  },
  "permission": {
    "read": { "*": "allow", "*.env": "deny" },
    "external_directory": "deny",
    "doom_loop": "deny"
  },
  "instructions": ["AGENTS.md"]
}
```

---

## Claude Code Setup

### 1. Create Agent Directory

```bash
mkdir -p .claude/agents
```

### 2. Copy Claude Code-format agents

Place `.md` files with this frontmatter style:

```yaml
---
name: coder
description: Go developer. Implements features following Clean Architecture.
tools: Read, Write, Edit, Bash, Glob, Grep, LS, TodoRead, TodoWrite
model: sonnet
skills: go-code, go-db
---
```

### 3. Create CLAUDE.md

```markdown
# Project Instructions

Read and follow conventions in AGENTS.md.

## Agent Usage

This project has custom agents in `.claude/agents/`:
- @coder - Implementation tasks
- @debugger - Bug investigation
- @planner - Task planning
- @reviewer - Code review
- @researcher - Deep analysis

Use appropriate agents for specialized tasks.
```

### 4. Verify Setup

```bash
# In Claude Code
/agents
```

---

## Format Quick Reference

### Tool Names

| OpenCode | Claude Code |
|----------|-------------|
| `read: true` | `Read` |
| `write: true` | `Write` |
| `edit: true` | `Edit` |
| `bash: true` | `Bash` |
| `grep: true` | `Grep` |
| `glob: true` | `Glob` |
| `list: true` | `LS` |
| `todoread: true` | `TodoRead` |
| `todowrite: true` | `TodoWrite` |
| `webfetch: true` | `WebFetch` |
| `websearch: true` | `WebSearch` |

### Model Names

| OpenCode | Claude Code |
|----------|-------------|
| `fast` | `haiku` |
| `main` | `sonnet` |
| `heavy` | `opus` |

### Permissions

| OpenCode | Claude Code |
|----------|-------------|
| `permission: { edit: "deny" }` | `disallowedTools: Write, Edit` |
| `permission: { bash: "ask" }` | `permissionMode: default` |

---

## Shared AGENTS.md

Both platforms can reference a shared `AGENTS.md` for project conventions:

```markdown
# Project Conventions

## Code Style
- Use short, natural names: cfg, repo, srv, ctx
- Errors: lowercase, wrapped with %w
- ZERO comments explaining WHAT

## Architecture
- Clean Architecture with domain at center
- Interfaces defined at consumer side
- One responsibility per component

## Tools
- Use `rg` instead of `grep` (10x faster)
- Use `fd` instead of `find` (5x faster)
- Always track progress with TODO tools
```

---

## Conversion Script

To convert OpenCode agents to Claude Code format:

```bash
#!/bin/bash
# convert-agents.sh

SRC=".opencode/agent"
DST=".claude/agents"

mkdir -p "$DST"

for file in "$SRC"/*.md; do
    name=$(basename "$file")
    echo "Converting $name..."
    
    # Extract and convert (simplified - use proper YAML parser for production)
    python3 << EOF
import yaml
import re

with open("$file", 'r') as f:
    content = f.read()

# Split frontmatter and body
parts = content.split('---', 2)
if len(parts) >= 3:
    frontmatter = yaml.safe_load(parts[1])
    body = parts[2]
    
    # Convert tools
    tools_map = {
        'read': 'Read', 'write': 'Write', 'edit': 'Edit',
        'bash': 'Bash', 'grep': 'Grep', 'glob': 'Glob',
        'list': 'LS', 'todoread': 'TodoRead', 'todowrite': 'TodoWrite',
        'webfetch': 'WebFetch', 'websearch': 'WebSearch', 'skill': None
    }
    
    tools = frontmatter.get('tools', {})
    if isinstance(tools, dict):
        tool_list = [tools_map[k] for k, v in tools.items() if v and tools_map.get(k)]
        tool_list = [t for t in tool_list if t]  # Remove None
    else:
        tool_list = []
    
    # Convert model
    model_map = {'fast': 'haiku', 'main': 'sonnet', 'heavy': 'opus'}
    model = model_map.get(frontmatter.get('model', 'main'), 'sonnet')
    
    # Convert skills
    skills = frontmatter.get('skills', [])
    if isinstance(skills, list):
        skills_str = ', '.join(skills)
    else:
        skills_str = skills
    
    # Build new frontmatter
    new_fm = {
        'name': frontmatter.get('name', ''),
        'description': frontmatter.get('description', '').strip('"'),
        'tools': ', '.join(tool_list),
        'model': model
    }
    if skills_str:
        new_fm['skills'] = skills_str
    
    # Write output
    with open("$DST/$name", 'w') as out:
        out.write('---\n')
        for k, v in new_fm.items():
            out.write(f'{k}: {v}\n')
        out.write('---\n')
        out.write(body)

print(f"Converted: $name")
EOF
done

echo "Done! Agents written to $DST"
```

---

## Validation

### OpenCode

```bash
# Start OpenCode and check agents
opencode
/agents  # Should list your agents
```

### Claude Code

```bash
# Start Claude Code and check agents
claude
/agents  # Should show custom agents
```

---

## Best Practices

1. **Keep prompts identical** - Only frontmatter differs
2. **Use rg/fd in prompts** - Works on both platforms
3. **Reference shared AGENTS.md** - Common conventions
4. **Test on both platforms** - Ensure compatibility
5. **Version control both formats** - Track changes
