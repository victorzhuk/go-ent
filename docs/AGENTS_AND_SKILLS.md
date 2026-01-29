# Agents and Skills System (v3)

Comprehensive guide to go-ent's v3 agent and skill architecture with dual-platform support for Claude Code and OpenCode.

---

## Table of Contents

- [Overview](#overview)
- [Agent Architecture](#agent-architecture)
- [Available Agents](#available-agents)
- [Skill System](#skill-system)
- [Available Skills](#available-skills)
- [Reference Skills](#reference-skills)
- [Usage Patterns](#usage-patterns)

---

## Overview

go-ent uses a **multi-agent system** where specialized agents collaborate to accomplish development tasks. Each agent has:

- **Domain expertise** (skills preloaded at startup)
- **Tool access** (allowed/disallowed tools)
- **Delegation chains** (dependencies)
- **Model assignment** (main/fast/heavy internally, sonnet/opus/haiku for Claude Code)

### v3 Design Principles

1. **Dual-Platform Support** - Works with both Claude Code and OpenCode
2. **Split Format** - Metadata (YAML) separated from prompts (Markdown)
3. **Template Generation** - Platform-specific output from unified source
4. **Reference Skills** - Agents preload shared skills via `skills:` field
5. **Markdown Sections** - Skills use `## Role`, `## Instructions` format
6. **Explicit Configuration** - All tool access and skills declared upfront

---

## Agent Architecture

### Split Format (v3)

Agents use a **split format** with metadata separated from prompts, allowing dual-platform support:

```
plugins/go-ent/
└── agents/
    ├── meta/                      # Agent metadata (YAML)
    │   ├── architect.yaml         # System design
    │   ├── coder.yaml             # Implementation
    │   ├── planner.yaml           # Task planning
    │   ├── debugger.yaml          # Bug investigation
    │   ├── tester.yaml            # Test coverage
    │   ├── reviewer.yaml          # Code review
    │   └── ...                    # 16 total agents
    ├── prompts/
    │   ├── shared/                # Shared prompt sections
    │   │   ├── _tooling.md        # Tool usage guidance
    │   │   ├── _conventions.md    # Go coding standards
    │   │   ├── _handoffs.md       # Agent handoff patterns
    │   │   ├── _judgment.md       # Decision framework
    │   │   ├── _principals.md     # Value hierarchy
    │   │   └── _openspec.md       # OpenSpec workflow
    │   └── agents/                # Agent-specific prompts
    │       ├── coder.md           # Implementation prompt
    │       ├── planner.md         # Planning prompt
    │       └── ...
    ├── presets/
    │   └── tools.yaml             # Tool preset definitions
    └── templates/
        ├── claude.yaml.tmpl       # Claude Code generation
        └── opencode.yaml.tmpl     # OpenCode generation
```

### Generated Output

Run `go-ent init --tool <platform>` to generate platform-specific files:

**Claude Code** (`.claude/agents/ent/*.md`):
```markdown
---
name: coder
description: Go developer. Implements features, writes code.
model: sonnet                      # Claude Code model name
skills:
  - go-code
  - go-db
disallowedTools:
  - mcp__plugin_serena_serena__replace_symbol_body
color: "#32CD32"
role: execution
complexity: standard
dependencies:
  - tester
  - reviewer
---

[Combined shared prompts + agent-specific prompt]
```

**OpenCode** (`.opencode/agents/ent/*.md`):
```markdown
---
name: coder
description: Go developer. Implements features, writes code.
model: main                        # OpenCode model name
tools:
    read: true                     # Lowercase object format
    write: true
    edit: true
    bash: true
skills:
  - go-code
  - go-db
tags:
  - role:execution                 # Tag format for role
color: "#32CD32"
dependencies:
  - tester
  - reviewer
---

[Combined shared prompts + agent-specific prompt]
```

### Metadata Structure

**File**: `agents/meta/<name>.yaml`

```yaml
name: coder
description: Go developer. Implements features, writes code.
model: main                        # Internal: main/fast/heavy (platform-agnostic)
color: '#32CD32'
skills:
  - go-code
  - go-db
toolPresets:
  - editing                        # Expands to tools based on platform
disallowedToolPresets:
  - serena-editing                 # Expands to Serena MCP tools
role: execution
complexity: standard
dependencies:
  - tester
  - reviewer
  - debugger
prompts:
  shared:                          # Shared prompt sections (preloaded skills)
    - _tooling
    - _conventions
    - _handoffs
    - _judgment
    - _principals
    - _openspec
  main: agents/coder               # Agent-specific prompt path
```

### Metadata Fields

**Required:**
- `name` - Agent identifier (lowercase, hyphens, no `ent:` prefix)
- `description` - When to delegate to this agent
- `model` - Internal model: `main` (Sonnet), `fast` (Haiku), `heavy` (Opus)
- `prompts.main` - Path to agent-specific prompt (e.g., `agents/coder`)

**Optional:**
- `color` - Agent color in hex (e.g., `'#32CD32'`)
- `role` - Agent role: `orchestration`, `planning`, `execution`, `validation`, `research`
- `complexity` - Complexity level: `fast`, `standard`, `heavy`
- `skills` - Domain skills to preload (array)
- `toolPresets` - Tool groupings (e.g., `editing`, `readonly`, `planning`)
- `disallowedToolPresets` - Denied tool groupings (e.g., `serena-editing`)
- `tools` - Explicit tool list (array, overrides presets)
- `disallowedTools` - Explicit denied tools (array)
- `dependencies` - Other agents this depends on (array, no `ent:` prefix)
- `prompts.shared` - Shared prompt sections (array, e.g., `_tooling`, `_conventions`)
- `permissionMode` - Permission handling mode
- `hooks` - Lifecycle hooks (PreToolUse, PostToolUse, Stop)

**Optional (go-ent extensions):**
- `color` - Hex color for UI
- `role` - Role category (planning, execution, validation, etc.)
- `complexity` - Complexity level (light, standard, heavy)
- `dependencies` - Agents this can delegate to (array)

---

## Available Agents

### Planning Agents (3)

**planner.md** - Standard task planning (sonnet)
```yaml
description: Task planner. Breaks features into actionable tasks.
model: sonnet
skills: [go-arch, go-code, ent-tools-planning, ent-conventions]
```

**planner-fast.md** - Quick planning (haiku)
```yaml
description: Quick task assessment and planning.
model: haiku
skills: [go-arch, ent-tools-readonly]
```

**planner-heavy.md** - Deep architectural planning (opus)
```yaml
description: Complex architectural planning with deep analysis.
model: opus
skills: [go-arch, arch-core, api-design, ent-tools-planning]
```

### Execution Agents (3)

**coder.md** - Go implementation (sonnet)
```yaml
description: Go developer. Implements features, writes code.
model: sonnet
skills: [go-code, go-db, ent-tools-editing, ent-conventions]
dependencies: [tester, reviewer, debugger]
```

**tester.md** - Test coverage and TDD (sonnet)
```yaml
description: Test engineer. Writes tests, TDD cycles.
model: sonnet
skills: [go-test, go-code, ent-tools-editing]
dependencies: [debugger]
```

**reproducer.md** - Bug reproduction (sonnet)
```yaml
description: Create minimal bug reproductions. Write failing tests first.
model: sonnet
skills: [go-test, debug-core, ent-tools-editing]
```

### Debugging Agents (3)

**debugger.md** - Standard debugging (sonnet)
```yaml
description: Standard debugging. Systematic issue investigation.
model: sonnet
skills: [go-code, debug-core, ent-tools-editing]
dependencies: [tester, reviewer]
```

**debugger-fast.md** - Quick fixes (haiku)
```yaml
description: Quick debugging for simple issues.
model: haiku
skills: [debug-core, ent-tools-editing]
```

**debugger-heavy.md** - Complex issues (opus)
```yaml
description: Complex debugging. Concurrency, performance, multi-component.
model: opus
skills: [go-code, go-perf, debug-core, ent-tools-editing]
```

### Review & Research (3)

**reviewer.md** - Code review (opus)
```yaml
description: Code reviewer. Reviews for bugs, quality, adherence.
model: opus
skills: [go-review, review-core, go-code, ent-tools-readonly]
```

**researcher.md** - Codebase research (sonnet)
```yaml
description: Research agent. Deep code analysis and investigation.
model: sonnet
skills: [go-arch, ent-tools-serena-analysis]
```

**architect.md** - System design (opus)
```yaml
description: System architect. Designs components, layers, data flow.
model: opus
skills: [go-arch, arch-core, api-design, ent-tools-planning]
dependencies: [planner, coder]
```

### Task Management (3)

**task-fast.md** - Quick task assessment (haiku)
```yaml
description: Quick task assessment and routing.
model: haiku
skills: [arch-core, ent-tools-readonly]
```

**task-heavy.md** - Complex task analysis (opus)
```yaml
description: Complex task analysis with deep reasoning.
model: opus
skills: [arch-core, go-arch, ent-tools-planning]
```

**decomposer.md** - Task breakdown (sonnet)
```yaml
description: Task breakdown and dependency analysis.
model: sonnet
skills: [go-arch, ent-tools-planning]
```

### Validation (1)

**acceptor.md** - Acceptance criteria (sonnet)
```yaml
description: Validate acceptance criteria and requirements.
model: sonnet
skills: [go-test, ent-tools-readonly]
```

---

## Skill System

### Skill Format (v3)

Skills use **Markdown sections** instead of XML tags:

**File**: `skills/<category>/<name>/SKILL.md`

```markdown
---
name: go-error
description: "Error handling patterns"
version: "1.0.0"
triggers:
  keywords:
    - error handling
  file_pattern: "*.go"
  weight: 0.8
---

## Role

Expert Go error handling engineer.

## Instructions

### Error Wrapping
Always wrap errors with context using %w.

## Constraints

- Include proper error wrapping
- Exclude unwrapped errors

## Examples

\```go
if err != nil {
    return fmt.Errorf("query: %w", err)
}
\```
```

### Skill Categories

**Core Skills (5):**
- `api-design` - REST/GraphQL API patterns
- `arch-core` - Architecture principles
- `debug-core` - Debugging approaches
- `review-core` - Code review frameworks
- `security-core` - Security best practices

**Go Skills (12):**
- `go-api` - Go API implementation
- `go-arch` - Go architecture
- `go-code` - Go coding patterns
- `go-config` - Configuration management
- `go-db` - Database patterns
- `go-error` - Error handling
- `go-migration` - Database migrations
- `go-ops` - Operational patterns
- `go-perf` - Performance optimization
- `go-review` - Go code review
- `go-sec` - Go security
- `go-test` - Testing and TDD

---

## Reference Skills

### What Are Reference Skills?

**Reference skills** are skills that agents preload via the `skills:` field instead of embedding content inline. They provide:

1. **Single source of truth** - Update once, affects all agents
2. **Composability** - Mix and match as needed
3. **Reusability** - Shared across multiple agents
4. **Maintainability** - Clear separation of concerns

### Tool Skills (4)

**ent-tools-readonly** - Read-only access
```yaml
skills:
  - ent-tools-readonly  # Grants: Read, Glob, Grep
```

**ent-tools-editing** - Full editing access
```yaml
skills:
  - ent-tools-editing  # Grants: Read, Write, Edit, Bash, Glob, Grep
```

**ent-tools-serena-analysis** - Semantic analysis
```yaml
skills:
  - ent-tools-serena-analysis  # Grants: Serena read-only tools
```

**ent-tools-planning** - Planning toolset
```yaml
skills:
  - ent-tools-planning  # Combined: read-only + task management + semantic
```

### Shared Knowledge Skills (6)

**ent-tooling** - Tool usage guidance
- Native tools (Read, Write, Edit)
- Serena semantic analysis
- Git commands
- Go commands
- Modern search (rg, fd)

**ent-conventions** - Go code style
- Naming conventions
- Error handling
- Comments policy
- Clean Architecture layers
- File organization

**ent-handoffs** - Agent delegation
- When to delegate
- Irreversible action checkpoints
- Handoff vs. escalation
- Agent responsibility matrix

**ent-judgment** - Constitutional AI
- Senior developer judgment
- When to ask vs. decide
- Non-negotiable boundaries
- Decision frameworks

**ent-principals** - Principal hierarchy
- Conflict resolution (Convention > Intent > Practice > Safety > Simplicity)
- When to ask vs. decide
- Escalation criteria

**ent-openspec** - OpenSpec workflow
- File structure (proposal.md, tasks.md, designs/)
- Workflow steps
- Task completion tracking
- Design documentation

---

## Usage Patterns

### Direct Invocation

```
/ent:coder "Implement user authentication"
/ent:planner "Break down payment system"
/ent:debugger "Fix race condition in handler"
```

### Automatic Delegation

Claude automatically delegates based on agent descriptions:

```
User: "I need to implement a new API endpoint for user profiles"

Claude: [Analyzes request] → Delegates to @ent:coder
```

### Skill Activation

Skills activate automatically based on:
1. **Preloaded skills** - Agent's `skills:` field
2. **Trigger keywords** - Skill's `triggers.keywords`
3. **File patterns** - Skill's `triggers.file_pattern`
4. **Manual invocation** - User requests specific skill

### Agent Chains

**Feature Implementation:**
```
architect → planner → coder → tester → reviewer
```

**Bug Fix:**
```
debugger → tester → reviewer
```

**Complex Architecture:**
```
architect → reviewer (heavy) → planner → decomposer → coder
```

---

## See Also

- [SKILL-AUTHORING.md](./SKILL-AUTHORING.md) - Write v3 skills
- [CLAUDE_CODE_COMPATIBILITY.md](./CLAUDE_CODE_COMPATIBILITY.md) - Alignment guide
- [MIGRATION_V3.md](./MIGRATION_V3.md) - Migrate from v2
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
