# OpenCode Engineering Guide — Documentation-Verified

This guide contains ONLY information verified against official OpenCode documentation at opencode.ai/docs.

---

## 1. Tools (from opencode.ai/docs/tools)

### Built-in Tools

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

### Key Notes from Documentation

- `write` and `patch` are controlled by `edit` permission
- `todowrite` and `todoread` are **disabled for subagents by default**
- Tools use ripgrep internally, respecting .gitignore

### Tools That DO NOT Exist

These are commonly assumed but NOT in OpenCode:
- ~~`websearch`~~ — Use `webfetch` with URLs
- ~~`task`~~ — Subagents are invoked via Task tool automatically
- ~~`str_replace`~~ — It's called `edit`

---

## 2. Agents (from opencode.ai/docs/agents)

### Agent Types

| Type | Description | Invocation |
|------|-------------|------------|
| `primary` | Main interaction agents | Tab key cycling |
| `subagent` | Specialized task agents | @ mention or automatic |
| `all` | Both modes (default) | Either method |

### Built-in Agents

| Agent | Type | Description |
|-------|------|-------------|
| **Build** | primary | Default. All tools enabled. |
| **Plan** | primary | Restricted. edit/bash set to "ask". |
| **General** | subagent | Full tools except todo. Multi-step tasks. |
| **Explore** | subagent | Read-only. Cannot modify files. |

### Configuration Methods

**Method 1: JSON** (`opencode.json`)

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

**Method 2: Markdown** (`.opencode/agents/my-agent.md`)

```markdown
---
description: What it does and when to use it
mode: subagent
model: anthropic/claude-sonnet-4-20250514
temperature: 0.3
tools:
  edit: false
  write: false
permission:
  bash: ask
---

System prompt content here (natural language).
```

### Frontmatter Fields (Documented)

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

### Task Permissions

Control which subagents an agent can invoke:

```json
{
  "agent": {
    "orchestrator": {
      "permission": {
        "task": {
          "*": "deny",
          "helper-*": "allow",
          "reviewer": "ask"
        }
      }
    }
  }
}
```

Last matching rule wins.

---

## 3. Skills (from opencode.ai/docs/skills)

### Skill Locations

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

Skill content here (natural language instructions).
```

### Frontmatter Fields (Documented)

| Field | Required | Constraints |
|-------|----------|-------------|
| `name` | **Yes** | 1-64 chars, lowercase, hyphen-separated |
| `description` | **Yes** | 1-1024 chars |
| `license` | No | String |
| `compatibility` | No | String |
| `metadata` | No | String-to-string map |

### Name Validation

```regex
^[a-z0-9]+(-[a-z0-9]+)*$
```

Name MUST match the directory name.

### Skill Permissions

```json
{
  "permission": {
    "skill": {
      "*": "allow",
      "internal-*": "deny"
    }
  }
}
```

---

## 4. Commands (from opencode.ai/docs/commands)

### Command Locations

| Location | Scope |
|----------|-------|
| `.opencode/commands/<name>.md` | Project |
| `~/.config/opencode/commands/<name>.md` | Global |

### Command Format

```markdown
---
description: What this command does
---

Prompt content here.
```

Filename becomes command name: `test.md` → `/test`

### Arguments

Use `$ARGUMENTS` for the full argument string:

```markdown
---
description: Create a component
---

Create a new React component named $ARGUMENTS with TypeScript.
```

Usage: `/component Button`

### Bash Injection

Use `!command` to include bash output:

```markdown
---
description: Analyze test coverage
---

!go test -cover ./...

Based on these results, suggest improvements.
```

### File Inclusion

Use `@filename` to include file contents:

```markdown
Review the changes in @README.md
```

---

## 5. Permissions (from opencode.ai/docs/permissions)

### Permission Values

| Value | Behavior |
|-------|----------|
| `allow` | Execute without asking |
| `ask` | Prompt for approval |
| `deny` | Block entirely |

### Global Permissions

```json
{
  "permission": {
    "edit": "ask",
    "bash": "ask",
    "webfetch": "allow"
  }
}
```

### Granular Bash Permissions

```json
{
  "permission": {
    "bash": {
      "*": "ask",
      "go build*": "allow",
      "go test*": "allow",
      "rm *": "deny"
    }
  }
}
```

**Rule**: Last matching pattern wins.

### Per-Agent Permissions

```json
{
  "agent": {
    "build": {
      "permission": {
        "bash": {
          "*": "ask",
          "git status*": "allow"
        }
      }
    }
  }
}
```

---

## 6. File Locations Summary

| Component | Project | Global |
|-----------|---------|--------|
| Config | `opencode.json` | `~/.config/opencode/opencode.json` |
| Rules | `AGENTS.md` | `~/.config/opencode/AGENTS.md` |
| Agents | `.opencode/agents/*.md` | `~/.config/opencode/agents/*.md` |
| Skills | `.opencode/skills/*/SKILL.md` | `~/.config/opencode/skills/*/SKILL.md` |
| Commands | `.opencode/commands/*.md` | `~/.config/opencode/commands/*.md` |

---

## 7. Agent Prompt Writing

### What Documentation Shows

Agent prompts are **natural language instructions**. The documentation examples show:

```markdown
---
description: Reviews code for quality
mode: subagent
tools:
  edit: false
---

You are in code review mode. Focus on:

- Code quality and best practices
- Potential bugs and edge cases
- Performance implications
- Security considerations

Provide constructive feedback without making direct changes.
```

### What Documentation Does NOT Show

- YAML syntax for tool invocation
- Parameter formats for tools
- Tool response formats

Tools are invoked by the LLM based on the conversation context, not by explicit syntax in the prompt.

---

## 8. Subagent Invocation

### From Documentation

1. **Automatic**: Primary agents can invoke subagents via Task tool based on descriptions
2. **Manual**: Users can @ mention subagents: `@explore find auth files`

### Task Permissions

Control automatic invocation:

```json
{
  "agent": {
    "orchestrator": {
      "permission": {
        "task": {
          "*": "deny",
          "helper-*": "allow"
        }
      }
    }
  }
}
```

### Hidden Subagents

Hide from @ autocomplete but allow programmatic invocation:

```json
{
  "agent": {
    "internal-helper": {
      "mode": "subagent",
      "hidden": true
    }
  }
}
```

---

## 9. Common Patterns

### Read-Only Agent

```markdown
---
description: Analyzes code without making changes
mode: subagent
tools:
  edit: false
  write: false
  patch: false
  bash: false
---

You analyze code and provide insights.
You cannot modify files.
```

### Restricted Primary Agent

```markdown
---
description: Orchestrator with limited bash access
mode: primary
tools:
  edit: false
  write: false
permission:
  bash:
    "*": deny
    "ls *": allow
    "cat *": allow
    "go build*": allow
    "go test*": allow
---

You coordinate work by delegating to subagents.
```

### Full-Access Subagent

```markdown
---
description: Implementation agent with file access
mode: subagent
tools:
  edit: true
  write: true
  bash: true
  todowrite: false
---

You implement code changes as directed.
```

Note: `todowrite: false` because it's disabled for subagents by default per docs.

---

## 10. What NOT to Do

### Don't Invent Tool Syntax

❌ Wrong (invented):
```yaml
read:
  path: "file.go"
  
grep:
  pattern: "func.*"
  path: "internal/"
```

✅ Right (natural language):
```
Read the file at internal/service/user.go.
Search for functions matching "Handler" in the internal directory.
```

### Don't Assume Tools Exist

❌ Wrong:
```
Use websearch to find documentation.
Use the task tool to delegate work.
```

✅ Right:
```
Use webfetch to retrieve the documentation URL.
@ mention the subagent to delegate work.
```

### Don't Invent Frontmatter Fields

Only use fields documented at opencode.ai/docs/agents:
- description, mode, model, temperature, maxSteps
- disable, prompt, tools, permission, hidden

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
