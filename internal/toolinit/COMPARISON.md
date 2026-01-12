# Claude Code vs OpenCode Adapter Comparison

Side-by-side comparison of the two tool adapters implementing FLOW.md dual-executor architecture.

## Directory Structure Comparison

| Aspect | Claude Code | OpenCode |
|--------|-------------|----------|
| **Root directory** | `.claude/` | `.opencode/` |
| **Commands dir** | `commands/` (PLURAL) | `command/` (SINGULAR) |
| **Agents dir** | `agents/` (PLURAL) | `agent/` (SINGULAR) |
| **Skills dir** | `skills/` (PLURAL) | `skill/` (SINGULAR) |
| **Skill structure** | Hierarchical (`skills/core/arch-core/`) | Flattened (`skill/core-arch-core/`) |

## Role & Responsibility

| Aspect | Claude Code (Driver) | OpenCode (Worker) |
|--------|---------------------|-------------------|
| **Primary role** | Planning & Design | Implementation & Execution |
| **Workflow phase** | Discovery, Research, Architecture | Coding, Testing, Debugging |
| **Decision level** | Strategic (what & why) | Tactical (how) |
| **Output** | Change proposals, specs, tasks | Working code, tests, fixes |

## Resource Filtering

### Commands

| Tool | Commands Included | Purpose |
|------|-------------------|---------|
| **Claude Code** | `plan.md` only (1) | Complete planning workflow |
| **OpenCode** | `task.md`, `bug.md` (2) | Execution and debugging workflows |

### Agents

**Claude Code (4 planning agents):**
- `planner-smoke.md` (Haiku) - Quick triage
- `architect.md` (Opus) - System design
- `planner.md` (Sonnet) - Detailed planning
- `decomposer.md` (Sonnet) - Task breakdown

**OpenCode (10 execution agents):**
- `task-smoke.md` (GLM 4.7) - Simple tasks
- `task-heavy.md` (Kimi K2) - Complex tasks
- `coder.md` (GLM 4.7) - Implementation
- `reviewer.md` (GLM 4.7) - Code review
- `tester.md` (GLM 4.7) - Testing
- `acceptor.md` (GLM 4.7) - Acceptance validation
- `reproducer.md` (GLM 4.7) - Bug reproduction
- `researcher.md` (GLM 4.7) - Investigation
- `debugger-smoke.md` (GLM 4.7) - Simple debugging
- `debugger-heavy.md` (Kimi K2) - Complex debugging

### Skills

| Tool | Skills | Structure |
|------|--------|-----------|
| **Claude Code** | All 14 (core + go) | Hierarchical categories preserved |
| **OpenCode** | All 14 (core + go) | Flattened with category prefix |

**Both tools share the same skill library** but organize it differently.

## Model Strategy

### Claude Code Models

| Model Short Name | Full API Model ID | Use Case |
|------------------|-------------------|----------|
| `opus` | `claude-opus-4-5-20250514` | System architecture, complex design |
| `sonnet` | `claude-sonnet-4-5-20250929` | Planning, task decomposition |
| `haiku` | `claude-haiku-4-5-20250429` | Quick triage, feasibility checks |

**Cost strategy:** Opus for critical decisions, Sonnet for heavy lifting, Haiku for speed

### OpenCode Models

| Model Short Name | Full Provider/Model ID | Use Case |
|------------------|------------------------|----------|
| `glm-4-flash` | `zhipu/glm-4-flash` | Primary worker (high limits) |
| `kimi-k2` | `moonshot/kimi-k2` | Heavy worker (complex tasks) |

**Cost strategy:** GLM 4.7 as default (cheap, high limits), Kimi K2 for escalations

## Frontmatter Format Comparison

### Agent Frontmatter

**Claude Code:**
```yaml
---
name: architect
description: System architect. Designs components, layers, data flow.
model: claude-opus-4-5-20250514
color: "#4169E1"
skills:
  - go-arch
  - go-api
tools:
  - read
  - grep
---
```

**OpenCode:**
```yaml
---
description: Execute simple tasks efficiently
mode: subagent
model: zhipu/glm-4-flash
temperature: 0.0
tools:
  read: true
  write: false
  edit: false
permission:
  skill:
    go-code: allow
    go-test: allow
---
```

**Key differences:**
- Claude: `name` field, `color` for UI, tools as list
- OpenCode: `mode` field, `temperature`, tools as map (explicit enable/disable), `permission` section

### Command Frontmatter

**Claude Code:**
```yaml
---
name: plan
description: Create complete OpenSpec change proposal
---
```

**OpenCode:**
```yaml
---
description: Execute OpenSpec tasks with TDD and validation
---
```

**Key difference:** Claude includes `name` field, OpenCode derives from filename

### Skill Frontmatter

**Both tools use the same format:**
```yaml
---
name: go-arch
description: Go architecture patterns
version: 1.0.0
---
```

## Implementation Details

### File Counts

| Aspect | Claude Code | OpenCode |
|--------|-------------|----------|
| **Commands** | 1 file | 2 files |
| **Agents** | 4 files | 10 files |
| **Skills** | 14 files (hierarchical) | 14 files (flattened) |
| **Total files** | ~19 files | ~26 files |

### Code Size

| File | Lines of Code | Purpose |
|------|---------------|---------|
| `claude.go` | 427 | Claude adapter implementation |
| `opencode.go` | 439 | OpenCode adapter implementation |
| `claude_test.go` | 147 | Claude adapter tests |
| `opencode_test.go` | 171 | OpenCode adapter tests |

### Test Coverage

| Adapter | Test Functions | Assertions | Status |
|---------|----------------|------------|--------|
| **Claude** | 7 | ~20 | ✅ All passing |
| **OpenCode** | 8 | ~25 | ✅ All passing |

## Workflow Integration

### Planning Phase (Claude Code)

```
User request
     ↓
/plan command
     ↓
@planner-smoke (Haiku) - Quick triage
     ↓
@architect (Opus) - System design
     ↓
@planner (Sonnet) - Detailed planning
     ↓
@decomposer (Sonnet) - Task breakdown
     ↓
Change proposal with tasks
```

### Execution Phase (OpenCode)

```
Change proposal with tasks
     ↓
/task command (auto-select from registry)
     ↓
@task-smoke (GLM) or @task-heavy (Kimi)
     ↓
@coder (GLM) - Implementation
     ↓
@tester (GLM) - Write tests
     ↓
@reviewer (GLM) - Code review
     ↓
@acceptor (GLM) - Validate requirements
     ↓
Task complete, mark in registry
```

### Bug Fixing (OpenCode)

```
Bug report
     ↓
/bug command
     ↓
@reproducer (GLM) - Create failing test
     ↓
@researcher (GLM) - Root cause analysis
     ↓
@debugger-smoke (GLM) or @debugger-heavy (Kimi)
     ↓
@coder (GLM) - Fix implementation
     ↓
@tester (GLM) - Validate fix
     ↓
Bug resolved
```

## Configuration Generation

Both adapters support:
- ✅ **Dry-run mode** - Preview without writing files
- ✅ **Force mode** - Overwrite existing configurations
- ✅ **Selective generation** - Filter agents/commands/skills
- ✅ **Error handling** - Contextual error messages
- ✅ **Directory creation** - Automatic parent directory creation
- ✅ **File permissions** - Proper 0644 for files, 0755 for dirs

## Usage Pattern

```go
// Detect tool based on existing directory
var adapter toolinit.Adapter

if _, err := os.Stat(".claude"); err == nil {
    adapter = toolinit.NewClaudeAdapter()
} else if _, err := os.Stat(".opencode"); err == nil {
    adapter = toolinit.NewOpenCodeAdapter()
} else {
    // Auto-detect or ask user
}

// Generate configuration
cfg := &toolinit.GenerateConfig{
    Path:     ".",
    PluginFS: goent.PluginFS,
    Force:    false,
    DryRun:   false,
}

if err := adapter.Generate(ctx, cfg); err != nil {
    log.Fatal(err)
}
```

## Design Principles

### Common Principles (Both)
- Single source of truth (embedded FS)
- Transform, don't duplicate
- Filter by tool responsibility
- Preserve metadata integrity
- Fail fast with clear errors

### Claude-Specific
- PLURAL naming (matches Claude Code conventions)
- Hierarchical skill organization
- Color coding for UI
- Model name as metadata

### OpenCode-Specific
- SINGULAR naming (matches OpenCode conventions)
- Flattened skill structure
- Permission-based access control
- Temperature for reproducibility
- Tool explicit enable/disable

## Future Enhancements

### Potential Features
- [ ] Cursor adapter (TBD if needed)
- [ ] Custom skill filtering by category
- [ ] Agent subset profiles (minimal, full, custom)
- [ ] Model override per agent
- [ ] Configuration validation
- [ ] Update/sync existing configs
- [ ] Migration between tools

### Compatibility
- OpenCode can read `.claude/skills/` directory
- Shared skill format enables cross-tool compatibility
- Agents remain tool-specific (different responsibilities)
