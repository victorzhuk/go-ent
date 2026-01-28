# Claude Code Extension Development Guide

A comprehensive reference for creating skills, subagents, and commands in Claude Code, based on official Anthropic documentation and OpenSpec v1.0.

---

## Table of Contents

1. [Overview](#overview)
2. [Extension Types](#extension-types)
3. [Skills (Model-Invoked)](#skills-model-invoked)
4. [Subagents (Delegated Specialists)](#subagents-delegated-specialists)
5. [Commands (User-Invoked)](#commands-user-invoked)
6. [Tool Reference](#tool-reference)
7. [OpenSpec Integration](#openspec-integration)
8. [Best Practices](#best-practices)
9. [Anti-Patterns](#anti-patterns)
10. [Examples](#examples)

---

## Overview

Claude Code provides three extension mechanisms:

| Type | Invocation | Purpose | Location |
|------|------------|---------|----------|
| **Skills** | Model decides automatically | Extend Claude's capabilities | `.claude/skills/` |
| **Subagents** | Model delegates or user requests | Specialized isolated tasks | `.claude/agents/` |
| **Commands** | User types `/command` | Explicit workflow triggers | `.claude/commands/` |

### Key Principles

From Anthropic's official guidance:

1. **Start simple, add complexity only when needed** — The most successful implementations use simple, composable patterns
2. **Skills are model-invoked** — Claude autonomously decides when to use them based on task + description
3. **Subagents preserve context** — Each operates in its own context window, preventing pollution of main conversation
4. **Commands merge with skills** — A file at `.claude/commands/review.md` and `.claude/skills/review/SKILL.md` both create `/review`

---

## Extension Types

### Comparison Matrix

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

## Skills (Model-Invoked)

Skills are modular capabilities that Claude autonomously decides to use based on task context and the skill's description.

### Directory Structure

```
skill-name/
├── SKILL.md           # Required: frontmatter + instructions
├── reference.md       # Optional: detailed documentation
├── examples.md        # Optional: usage examples
├── scripts/           # Optional: executable utilities
│   └── helper.py
└── templates/         # Optional: output templates
    └── template.txt
```

### SKILL.md Format

```yaml
---
name: lowercase-with-hyphens
description: What this skill does AND when Claude should use it. Include specific triggers.
allowed-tools: Read, Grep, Glob
---

# Skill Name

## Instructions
Clear step-by-step guidance for Claude.

## Examples
Concrete examples of using this skill.
```

### Frontmatter Fields

| Field | Required | Max Length | Description |
|-------|----------|------------|-------------|
| `name` | Yes | 64 chars | Lowercase letters, numbers, hyphens only |
| `description` | Yes | 1024 chars | What it does AND when to use it |
| `allowed-tools` | No | — | Comma-separated tool list (omit to inherit all) |

### Description Best Practices

The `description` field is **critical** for Claude to discover when to use your skill.

**Too vague:**
```yaml
description: Helps with documents
```

**Specific and actionable:**
```yaml
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.
```

### Progressive Disclosure

Claude reads files on-demand to manage context:

| Level | When Loaded | Content |
|-------|-------------|---------|
| Always | Session start | name + description (~100 words) |
| Triggered | Skill matches task | Full SKILL.md body |
| On-demand | Claude needs details | reference.md, scripts/, etc. |

Reference files from SKILL.md:
```markdown
For advanced usage, see [reference.md](reference.md).

Run the helper script:
```bash
python scripts/helper.py input.txt
```
```

### Tool Restrictions

Use `allowed-tools` to limit what Claude can do when a skill is active:

```yaml
---
name: safe-reader
description: Read files without making changes. Use for read-only analysis.
allowed-tools: Read, Grep, Glob
---
```

When active, Claude can only use specified tools without permission prompts.

---

## Subagents (Delegated Specialists)

Subagents are specialized AI assistants that Claude can delegate tasks to. Each subagent has its own context window, preventing pollution of the main conversation.

### File Format

```yaml
---
name: agent-name
description: What this agent does. Include "Use PROACTIVELY" for auto-delegation.
tools: Read, Grep, Glob, Bash
model: sonnet
permissionMode: default
skills: skill1, skill2
---

System prompt goes here. This can be multiple paragraphs
defining the agent's role, capabilities, and approach.

Include specific instructions, best practices, and constraints.
```

### Frontmatter Fields

| Field | Required | Values | Description |
|-------|----------|--------|-------------|
| `name` | Yes | lowercase-hyphens | Unique identifier |
| `description` | Yes | text | When Claude should delegate to this agent |
| `tools` | No | comma-separated | Omit to inherit all tools from main thread |
| `model` | No | `sonnet`, `opus`, `haiku`, `inherit` | Defaults to configured subagent model |
| `permissionMode` | No | `default`, `acceptEdits`, `bypassPermissions`, `plan`, `ignore` | How to handle permissions |
| `skills` | No | comma-separated | Skills to auto-load when agent starts |

### Model Selection

```yaml
model: sonnet    # Use Sonnet (default for subagents)
model: opus      # Use Opus for complex reasoning
model: haiku     # Use Haiku for fast, simple tasks
model: inherit   # Use same model as main conversation
```

### Permission Modes

| Mode | Behavior |
|------|----------|
| `default` | Normal permission prompts |
| `acceptEdits` | Auto-accept file edits |
| `bypassPermissions` | Skip all permission prompts |
| `plan` | Read-only, no modifications |
| `ignore` | Ignore permission requests |

### Built-in Subagents

Claude Code includes three built-in subagents:

#### Explore (Haiku)
Fast, lightweight, read-only codebase exploration.

- **Model**: Haiku (fast, low-latency)
- **Mode**: Strictly read-only
- **Tools**: Glob, Grep, Read, Bash (read-only: ls, git status, git log, git diff, find, cat, head, tail)

**Thoroughness levels**: quick, medium, very thorough

#### Plan (Sonnet)
Research during plan mode.

- **Model**: Sonnet
- **Mode**: Read-only research
- **Tools**: Read, Glob, Grep, Bash
- **Invocation**: Automatic in plan mode

#### general-purpose (Sonnet)
Complex multi-step tasks requiring both exploration and action.

- **Model**: Sonnet
- **Mode**: Can read and write files
- **Tools**: All tools available
- **Purpose**: Complex research, multi-step operations, code modifications

### Invoking Subagents

**Automatic delegation**: Claude decides based on `description` field.

**Explicit invocation**:
```
Use the code-reviewer subagent to check my recent changes
Have the debugger subagent investigate this error
```

**Proactive delegation**: Include "Use PROACTIVELY" or "MUST BE USED" in description.

### Subagent Constraints

- Subagents cannot spawn other subagents (prevents infinite nesting)
- Each operates in its own context window
- Results are returned to main conversation
- Can be resumed using agent ID

---

## Commands (User-Invoked)

Commands are explicit workflow triggers invoked by typing `/command-name`.

### File Format

```yaml
---
description: Brief purpose shown in /help
argument-hint: [file] [options]
allowed-tools: Read, Grep, Bash(git:*)
model: haiku
---

Command instructions go here.
Use $ARGUMENTS for all args or $1, $2 for positional.
```

### Frontmatter Fields

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

### Example Command

```yaml
---
description: Fix a GitHub issue by number
argument-hint: [issue-number]
allowed-tools: Read, Write, Edit, Bash(git:*), Bash(gh:*)
---

# Fix Issue $1

1. Get issue details: !`gh issue view $1`
2. Read the issue description
3. Understand the requirements
4. Implement the fix
5. Write tests
6. Create a commit with message referencing #$1
```

Usage: `/fix-issue 123`

### Commands vs Skills

Commands and skills can coexist with the same name:
- `.claude/commands/review.md` → Creates `/review`
- `.claude/skills/review/SKILL.md` → Also creates `/review`

Both work identically. Skills add optional features: directory for supporting files, auto-invocation by Claude.

---

## Tool Reference

### Available Tools

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
# Specific tools
allowed-tools: Read, Grep, Glob

# Bash with restrictions
allowed-tools: Bash(git:*), Bash(npm:*)

# All tools (omit field)
# tools: (not specified)
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
# List active changes
openspec list

# Check artifact status
openspec status --change change-id

# Machine-readable status
openspec status --change change-id --json

# View change details
openspec show change-id

# Validate spec formatting
openspec validate change-id

# Get creation guidance
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
├── project.md          # Project context for AI
├── AGENTS.md           # Instructions for AI assistants
├── specs/              # Permanent specifications library
│   └── capability/
│       └── spec.md
├── changes/            # Active changes
│   └── change-id/
│       ├── proposal.md
│       ├── specs/
│       ├── design.md
│       └── tasks.md
└── archive/            # Completed changes
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

## Best Practices

### From Official Documentation

#### Skills

1. **Keep skills focused** — One skill, one capability
2. **Write clear descriptions** — Include what AND when to use
3. **Test with your team** — Verify skills activate when expected
4. **Document versions** — Track changes in SKILL.md

#### Subagents

1. **Start with Claude-generated agents** — Use `/agents` → Generate with Claude, then customize
2. **Design focused subagents** — Single, clear responsibilities
3. **Write detailed prompts** — Include specific instructions and constraints
4. **Limit tool access** — Only grant necessary tools
5. **Version control** — Check project subagents into git

#### General

1. **Use proactive language** — Include "Use PROACTIVELY" in descriptions
2. **Specify tool restrictions** — Don't grant more access than needed
3. **Include verification steps** — Every change should have a verify command
4. **Reference existing patterns** — Point to specific file:line references

### Prompt Engineering (Claude 4.x)

From Anthropic's Claude 4 best practices:

1. **Be explicit with instructions** — Claude 4 follows instructions precisely
2. **Add context for motivation** — Explain WHY rules matter
3. **Tell Claude what to do, not what not to do**
4. **Use XML format indicators** — `<section>content</section>`
5. **Match prompt style to desired output**

### Context Management

- **Long data (>20K tokens)**: Place at TOP of context
- **Instructions and queries**: Place at END of context
- **This ordering improves response quality by up to 30%**

### Extended Thinking

Keywords map to thinking depth:
```
"think" < "think hard" < "think harder" < "ultrathink"
```

Include "ultrathink" in skill content to enable extended thinking.

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

## Examples

### Read-Only Code Reviewer (Subagent)

```yaml
---
name: code-reviewer
description: Reviews code for quality and best practices. Use PROACTIVELY after code changes.
tools: Read, Grep, Glob, Bash
model: sonnet
permissionMode: plan
---

You are a senior code reviewer ensuring high standards.

When invoked:
1. Run git diff to see recent changes
2. Focus on modified files
3. Begin review immediately

Review checklist:
- Code is simple and readable
- Functions and variables are well-named
- No duplicated code
- Proper error handling
- No exposed secrets
- Input validation implemented
- Good test coverage

Provide feedback organized by priority:
- Critical issues (must fix)
- Warnings (should fix)
- Suggestions (consider)

Include specific examples of how to fix issues.
```

### PDF Processing Skill

```
pdf-processing/
├── SKILL.md
├── FORMS.md
├── REFERENCE.md
└── scripts/
    ├── fill_form.py
    └── validate.py
```

**SKILL.md:**
```yaml
---
name: pdf-processing
description: Extract text, fill forms, merge PDFs. Use when working with PDF files, forms, or document extraction. Requires pypdf and pdfplumber packages.
---

# PDF Processing

## Quick start

Extract text:
```python
import pdfplumber
with pdfplumber.open("doc.pdf") as pdf:
    text = pdf.pages[0].extract_text()
```

For form filling, see [FORMS.md](FORMS.md).
For detailed API reference, see [REFERENCE.md](REFERENCE.md).

## Requirements

```bash
pip install pypdf pdfplumber
```
```

### Fix Issue Command

```yaml
---
description: Fix a GitHub issue by number
argument-hint: [issue-number]
allowed-tools: Read, Write, Edit, Bash(git:*), Bash(gh:*)
---

# Fix Issue $1

## Context
- Issue details: !`gh issue view $1 --json title,body,labels`
- Current branch: !`git branch --show-current`

## Instructions

1. Read the issue description above
2. Understand the requirements
3. Search codebase for relevant files
4. Implement the fix following existing patterns
5. Write tests for the change
6. Verify: `make test`
7. Create commit: `git commit -m "Fix #$1: brief description"`
```

### Orchestrator Subagent

```yaml
---
name: ent-driver
description: Read-only orchestrator for task coordination. Use PROACTIVELY for multi-step implementations, OpenSpec workflows, complex feature development. Analyzes requirements, scores complexity, delegates to specialized subagents, tracks state. Never modifies files directly.
tools: Read, Grep, Glob, Bash, TodoWrite, TodoRead
model: sonnet
permissionMode: default
---

# Enterprise Driver — Task Orchestrator

You are a read-only orchestrator. You understand requirements, plan work, delegate implementation to subagents, and verify completion.

## Core Workflow

1. ASSESS — Gather context with read-only tools
2. SCORE — Rate complexity 1-10
3. PLAN — Delegate planning to ent-planner subagent
4. APPROVE — Present plan, wait for explicit approval
5. DELEGATE — Route to implementation subagents
6. VERIFY — Confirm completion via inspection
7. COMPLETE — Archive if OpenSpec used

## Self-Check

Before every action:
- Am I using Edit/Write/MultiEdit? → Delegate to ent-coder
- Am I running Bash that modifies files? → Delegate
- Am I implementing code? → Delegate

## Delegation Pattern

Use the [subagent-name] subagent to [task].

Context:
- relevant details

Files:
- paths

Verification:
- commands

## Never

- Use Edit, Write, MultiEdit directly
- Proceed without explicit approval
- Skip verification after implementation
```

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
# Start change
/opsx:new my-feature

# Create all artifacts
/opsx:ff my-feature

# Implement
/opsx:apply my-feature

# Complete
/opsx:archive my-feature

# Check status
openspec status --change my-feature
```

---

## Sources

- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills)
- [Claude Code Subagents Documentation](https://code.claude.com/docs/en/sub-agents)
- [Claude 4 Best Practices](https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/claude-4-best-practices)
- [OpenSpec GitHub](https://github.com/Fission-AI/OpenSpec)
- [OpenSpec Documentation](https://openspec.dev)
