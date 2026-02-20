# OpenCode Extension Reference

Reference for creating agents, skills, and commands in OpenCode. Verified against opencode.ai/docs.

---

## 1. Tools

| Tool | Purpose | Permission Key |
|------|---------|----------------|
| `bash` | Execute shell commands | `bash` |
| `edit` | Modify files via string replacement | `edit` |
| `write` | Create/overwrite files | `edit` (same key) |
| `read` | Read file contents | `read` |
| `grep` | Regex search (uses ripgrep) | `grep` |
| `glob` | Find files by pattern | `glob` |
| `list` | List directories | `list` |
| `lsp` | LSP operations (experimental) | `lsp` |
| `patch` | Apply diffs | `edit` (same key) |
| `skill` | Load SKILL.md content | `skill` |
| `todowrite` | Manage todo lists | `todowrite` |
| `todoread` | Read todo lists | `todoread` |
| `webfetch` | Fetch web content | `webfetch` |
| `question` | Ask user questions | `question` |

**Notes:**
- `write` and `patch` are controlled by `edit` permission
- `todowrite` and `todoread` are disabled for subagents by default
- No `websearch` tool — use `webfetch` with URLs

---

## 2. Agents

### Built-in Agents

| Agent | Type | Description |
|-------|------|-------------|
| **Build** | primary | Default. All tools enabled. |
| **Plan** | primary | Restricted. edit/bash set to "ask". |
| **General** | subagent | Full tools except todo. Multi-step tasks. |
| **Explore** | subagent | Read-only. Cannot modify files. |

### Configuration

**JSON** (`opencode.json`):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "agent": {
    "my-agent": {
      "description": "What it does and when to use it",
      "mode": "subagent",
      "model": "anthropic/claude-sonnet-4-20250514",
      "temperature": 0.3,
      "maxSteps": 50,
      "tools": {
        "edit": false,
        "write": false
      },
      "permission": {
        "bash": "ask"
      }
    }
  }
}
```

**Markdown** (`.opencode/agents/my-agent.md`):

```markdown
---
description: What it does and when to use it
mode: subagent
model: anthropic/claude-sonnet-4-20250514
tools:
  edit: false
  write: false
---

System prompt content here.
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `description` | **Yes** | What agent does, when to use |
| `mode` | No | `primary`, `subagent`, `all` (default) |
| `model` | No | Format: `provider/model-id` |
| `temperature` | No | 0.0-1.0 |
| `maxSteps` | No | Max iterations |
| `disable` | No | Disable agent |
| `prompt` | No | Path to prompt file: `{file:./path}` |
| `tools` | No | Enable/disable tools |
| `permission` | No | Tool permissions |
| `hidden` | No | Hide from @ menu (subagents only) |

---

## 3. Skills

### Locations

| Location | Scope |
|----------|-------|
| `.opencode/skills/<name>/SKILL.md` | Project |
| `~/.config/opencode/skills/<name>/SKILL.md` | Global |
| `.claude/skills/<name>/SKILL.md` | Claude-compatible |

### SKILL.md Format

```markdown
---
name: my-skill
description: What this skill provides and when to use it
license: MIT
compatibility: opencode
metadata:
  domain: go
---

Skill content here.
```

### Frontmatter Fields

| Field | Required | Constraints |
|-------|----------|-------------|
| `name` | **Yes** | 1-64 chars, `^[a-z0-9]+(-[a-z0-9]+)*$`, must match directory name |
| `description` | **Yes** | 1-1024 chars |
| `license` | No | String |
| `compatibility` | No | String |
| `metadata` | No | String-to-string map |

---

## 4. Commands

### Locations

| Location | Scope |
|----------|-------|
| `.opencode/commands/<name>.md` | Project |
| `~/.config/opencode/commands/<name>.md` | Global |

### Format

```markdown
---
description: What this command does
---

Prompt content here. Use $ARGUMENTS for the full argument string.
```

Filename becomes command name: `test.md` → `/test`

### Argument and Injection Features

| Feature | Syntax | Example |
|---------|--------|---------|
| Arguments | `$ARGUMENTS` | `/component Button` |
| Bash injection | `!command` | `!go test -cover ./...` |
| File inclusion | `@filename` | `@README.md` |

---

## Quick Reference

### Tools
```
bash | edit | write | read | grep | glob | list
patch | skill | todowrite | todoread | webfetch | question | lsp
```

### Built-in Agents
```
Build (primary) | Plan (primary) | General (subagent) | Explore (subagent)
```

### Permissions
```
allow | ask | deny
```

### Agent Modes
```
primary | subagent | all (default)
```
