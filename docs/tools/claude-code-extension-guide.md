# Claude Code Extension Development Guide

Reference for creating skills, subagents, and commands in Claude Code.

---

## Extension Types

| Aspect | Skills | Subagents | Commands |
|--------|--------|-----------|----------|
| Trigger | Claude decides | Claude delegates or user asks | User types `/name` |
| Context | Main conversation | Separate context window | Main conversation |
| File format | Directory with SKILL.md | Single .md file | Single .md file |
| Can modify files | Yes (if tools allow) | Yes (if tools allow) | Yes (if tools allow) |
| Frontmatter field | `allowed-tools` | `tools` | `allowed-tools` |

### File Locations

**Project-level** (shared via git):
```
.claude/
├── skills/
│   └── skill-name/
│       └── SKILL.md
├── agents/
│   └── agent-name.md
└── commands/
    └── command-name.md
```

**User-level** (personal, all projects):
```
~/.claude/
├── skills/
├── agents/
└── commands/
```

**Priority**: Project-level > User-level (when names conflict)

---

## Skills — Frontmatter Fields

| Field | Required | Max Length | Description |
|-------|----------|------------|-------------|
| `name` | Yes | 64 chars | Lowercase letters, numbers, hyphens only |
| `description` | Yes | 1024 chars | What it does AND when to use it |
| `allowed-tools` | No | — | Comma-separated tool list (omit to inherit all) |

---

## Subagents — Frontmatter Fields

| Field | Required | Values | Description |
|-------|----------|--------|-------------|
| `name` | Yes | lowercase-hyphens | Unique identifier |
| `description` | Yes | text | When Claude should delegate to this agent |
| `tools` | No | comma-separated | Omit to inherit all tools from main thread |
| `model` | No | `sonnet`, `opus`, `haiku`, `inherit` | Defaults to configured subagent model |
| `permissionMode` | No | `default`, `acceptEdits`, `bypassPermissions`, `plan`, `ignore` | How to handle permissions |
| `skills` | No | comma-separated | Skills to auto-load when agent starts |

---

## Commands — Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `description` | Yes | Shown in /help output |
| `argument-hint` | No | Shows expected arguments |
| `allowed-tools` | No | Tool restrictions |
| `model` | No | Model override |

### Variable Substitution

| Syntax | Meaning |
|--------|---------|
| `$ARGUMENTS` | All arguments as string |
| `$1`, `$2`, `$3` | Positional arguments |
| `${1:-default}` | Default value if arg missing |
| `` !`command` `` | Execute bash, insert output (preprocessing) |
| `@filepath` | Reference file content |

---

## Tool Reference

| Tool | Purpose | Type |
|------|---------|------|
| `Read` | Read file contents (text, images, PDFs) | Read-only |
| `Write` | Create or overwrite files | Write |
| `Edit` | Exact string replacements in files | Write |
| `MultiEdit` | Batch edits to single file | Write |
| `Glob` | File pattern matching (`**/*.ts`) | Read-only |
| `Grep` | Content search with regex (ripgrep) | Read-only |
| `Bash` | Execute shell commands | Varies |
| `WebFetch` | Fetch web page content | Read-only |
| `WebSearch` | Search the web | Read-only |
| `TodoRead` | Read todo list state | Read-only |
| `TodoWrite` | Create/update todo lists | Write |
| `Task` | Launch subagent | Special |

### Tool Permission Syntax

```yaml
allowed-tools: Read, Grep, Glob
allowed-tools: Bash(git:*), Bash(npm:*)
```

### Recommended Tool Sets by Agent Type

| Agent Type | Recommended Tools |
|------------|-------------------|
| Read-only reviewers | Read, Grep, Glob |
| Research agents | Read, Grep, Glob, WebFetch, WebSearch |
| Code writers | Read, Write, Edit, Bash, Glob, Grep |
| Documentation | Read, Write, Edit, Glob, Grep, WebFetch |
| Orchestrators | Read, Grep, Glob, TodoRead, TodoWrite |

---

## OpenSpec Integration

OpenSpec is a spec-driven development framework for AI coding assistants. It provides structure before code through proposals, specs, designs, and tasks.

### Installation

```bash
npm install -g @fission-ai/openspec@latest
```

Requires Node.js 20.19.0+

### Initialize in Project

```bash
openspec init
```

Select your AI tool (Claude Code, Cursor, etc.) during setup.

### Slash Commands

| Command | Purpose |
|---------|---------|
| `/opsx:explore` | Think through ideas before committing to a change |
| `/opsx:new id` | Start a new change, create folder structure |
| `/opsx:continue` | Create next artifact in dependency chain |
| `/opsx:ff id` | Fast-forward: create all planning artifacts at once |
| `/opsx:apply id` | Implement tasks from tasks.md |
| `/opsx:verify id` | Validate implementation matches artifacts |
| `/opsx:sync` | Merge delta specs into main specs |
| `/opsx:archive id` | Archive completed change |
| `/opsx:bulk-archive` | Archive multiple completed changes |
| `/opsx:onboard` | Guided tutorial through complete workflow |

### CLI Commands (Read-Only Queries)

```bash
openspec list
openspec status --change change-id
openspec status --change change-id --json
openspec show change-id
openspec validate change-id
openspec instructions artifact --change change-id
```

### Artifact Flow

```
/opsx:new → /opsx:ff → /opsx:apply → /opsx:verify → /opsx:archive
    │           │           │             │              │
    ▼           ▼           ▼             ▼              ▼
 folder    proposal.md   implement    validate      merge specs
 created   specs/        tasks        matches       to library
           design.md
           tasks.md
```

### Directory Structure

```
openspec/
├── project.md
├── AGENTS.md
├── specs/
│   └── capability/
│       └── spec.md
├── changes/
│   └── change-id/
│       ├── proposal.md
│       ├── specs/
│       ├── design.md
│       └── tasks.md
└── archive/
    └── YYYY-MM-DD-change-id/
```

### When to Use OpenSpec

| Complexity | Recommendation |
|------------|----------------|
| 1-3 (simple) | Skip OpenSpec |
| 4-5 (medium) | Optional |
| 6-7 (complex) | Recommended |
| 8-10 (architecture) | Required |

### Integration with Subagents

Orchestrator pattern with OpenSpec:

1. **ent-driver** (orchestrator) assesses task, scores complexity
2. **ent-planner** researches and creates plan
3. If complexity 6+: recommend OpenSpec
4. **ent-coder** uses `/opsx:new`, `/opsx:ff` to create artifacts
5. **ent-coder** uses `/opsx:apply` to implement
6. **ent-driver** verifies with read-only inspection
7. **ent-coder** uses `/opsx:archive` to complete

---

## Anti-Patterns

### Skills

| Don't | Do |
|-------|-----|
| Vague description: "Helps with data" | Specific: "Analyze Excel spreadsheets, create pivot tables. Use when working with .xlsx files" |
| One skill for everything | Focused skills with single purpose |
| Forget triggers | Include "Use when..." in description |

### Subagents

| Don't | Do |
|-------|-----|
| Grant all tools by default | Specify only needed tools |
| Skip the system prompt | Write detailed role and constraints |
| Expect chaining | Subagents can't spawn subagents |

### Commands

| Don't | Do |
|-------|-----|
| Complex multi-step logic | Keep commands focused |
| Missing argument hints | Include `argument-hint` for clarity |
| Forget verification | Always include verification steps |

### General

| Don't | Do |
|-------|-----|
| `@agent` syntax | Natural language: "Use the X subagent" |
| Imagine tools exist | Verify tool names in docs |
| Skip complexity assessment | Score before planning |
| Implement in orchestrator | Delegate all modifications |

---

## Quick Reference

### Frontmatter Cheatsheet

**Skill:**
```yaml
---
name: skill-name
description: What + when (max 1024 chars)
allowed-tools: Read, Grep, Glob
---
```

**Subagent:**
```yaml
---
name: agent-name
description: What + when. Use PROACTIVELY for auto-delegation.
tools: Read, Grep, Glob, Bash
model: sonnet
permissionMode: default
skills: skill1, skill2
---
```

**Command:**
```yaml
---
description: Purpose for /help
argument-hint: [args]
allowed-tools: Read, Edit, Bash(git:*)
---
```

### Variable Substitution

| Syntax | Meaning |
|--------|---------|
| `$ARGUMENTS` | All args as string |
| `$1`, `$2` | Positional args |
| `${1:-default}` | Default value |
| `` !`cmd` `` | Execute bash (preprocessing) |
| `@path` | Reference file |

### OpenSpec Quick Commands

```bash
/opsx:new my-feature
/opsx:ff my-feature
/opsx:apply my-feature
/opsx:archive my-feature
openspec status --change my-feature
```
