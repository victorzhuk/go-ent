# Agents and Skills

> **Note:** This document describes the current 13-agent structure. Agent simplification to 7 agents with complexity routing is planned for the next release.

This document describes the agent and skill system in go-ent after the comprehensive refactoring that consolidated skills and introduced a template-based architecture.

## Table of Contents

- [Overview](#overview)
- [Consolidated Skills](#consolidated-skills)
- [Agents](#agents)
- [Template System](#template-system)
- [Architecture](#architecture)
- [Usage](#usage)

## Overview

The go-ent agent system uses a **hybrid model**: templates for structural consistency, skills for domain knowledge. This architecture was established through a multi-phase refactoring (2024-2026) that:

1. **Phase 1**: Consolidated 11 skills to 7, reduced prompts from 5 to 3 (48% line reduction)
2. **Phase 2**: Built template infrastructure with parameterized sections
3. **Phase 3**: Migrated agents to template-based generation (70% prompt reduction)

### Key Benefits

- **Single source of truth** - Skills define domain knowledge once
- **Parameterized templates** - Role-specific content (execution, planning, validation, research)
- **Conditional rendering** - Tool restrictions and dependencies rendered dynamically
- **Reduced duplication** - ~65% reduction in total maintenance burden
- **Consistent quality** - All agents follow same patterns and best practices

## Consolidated Skills

### Core Skills (3)

#### 1. ent-foundation
**Purpose:** Core decision framework, judgment, principals, and Go conventions
**Preloaded by:** All agents
**Content:**
- Constitutional AI judgment framework
- Principal hierarchy (project conventions > user intent > best practices > safety > simplicity)
- Go code conventions (naming, errors, comments, Clean Architecture)
- When to ask vs. decide guidance
- Non-negotiable boundaries

**Triggers:** `ent-foundation`, `judgment`, `principals`, `conventions`

#### 2. ent-workflow
**Purpose:** OpenSpec workflow and agent delegation patterns
**Preloaded by:** All agents
**Content:**
- OpenSpec CLI workflow (new change, fast-forward, validate, archive)
- Artifact management (proposal.md, tasks.md, specs/)
- Agent delegation patterns (when to delegate, handoffs)
- Safety checkpoints before irreversible operations

**Triggers:** `ent-workflow`, `openspec`, `change proposal`, `spec workflow`, `handoffs`, `delegation`

#### 3. ent-tooling
**Purpose:** Tool usage guidance and modern CLI alternatives
**Preloaded by:** All agents
**Content:**
- Modern tool alternatives (rg, fd, bat, eza)
- Native vs. MCP tool usage
- Performance optimization guidance

**Triggers:** `ent-tooling`, `tools`, `cli`

### Tool Presets

Tool access is configured through `pkg/agents/presets/tools.yaml` with the following presets:

- **readonly**: Read, Glob, Grep - for research and validation agents
- **editing**: Read, Write, Edit, Bash, Glob, Grep - for implementation agents
- **planning**: Readonly tools + Task* tools + Serena analysis - for planning agents
- **serena-analysis**: Serena semantic code analysis tools - available to all agents

This replaces the previous ent-tools-* empty shell skills that were removed in the modernization.

### Archived Skills

The following skills were consolidated in Phase 1 (2026-01):
- `ent-judgment` → merged into `ent-foundation`
- `ent-principals` → merged into `ent-foundation`
- `ent-conventions` → merged into `ent-foundation`
- `ent-handoffs` → merged into `ent-workflow`
- `ent-openspec` → merged into `ent-workflow`

Location: `pkg/skills/ent/_archived/`

## Agents

### Agent Roles

Agents are categorized by role, which determines their behavior and content:

| Role | Purpose | Examples |
|------|---------|----------|
| **execution** | Code implementation | coder, tester, debugger |
| **planning** | Task breakdown, architecture | planner, architect, decomposer |
| **validation** | Code review, quality checks | reviewer, acceptor |
| **research** | Investigation, analysis | researcher |

### Agent List (13 total)

#### Execution Agents (5)

**coder** - Go developer, implements features
- Role: execution
- Model: main (sonnet)
- Dependencies: tester, reviewer, debugger
- Skills: go-code, go-db, ent-foundation, ent-workflow, ent-tooling

**tester** - Test engineer, writes tests
- Role: execution
- Model: main (sonnet)
- Skills: go-test, ent-foundation, ent-workflow, ent-tooling

**debugger** - Standard debugging and investigation
- Role: execution
- Model: main (sonnet)
- Skills: go-code, ent-foundation, ent-workflow, ent-tooling

**debugger-fast** - Quick debugging for simple issues
- Role: execution
- Model: fast (haiku)
- Skills: go-code, ent-foundation, ent-workflow, ent-tooling

**debugger-heavy** - Complex debugging (concurrency, performance)
- Role: execution
- Model: heavy (opus)
- Skills: go-code, go-perf, ent-foundation, ent-workflow, ent-tooling

#### Planning Agents (4)

**planner** - Task breakdown and planning
- Role: planning
- Model: main (sonnet)
- Dependencies: coder, tester
- Skills: go-arch, ent-foundation, ent-workflow, ent-tooling

**planner-fast** - Quick task assessment and routing
- Role: planning
- Model: fast (haiku)
- Skills: ent-foundation, ent-workflow, ent-tooling

**planner-heavy** - Deep architectural planning
- Role: planning
- Model: heavy (opus)
- Skills: go-arch, go-api, ent-foundation, ent-workflow, ent-tooling

**architect** - System design and architecture
- Role: planning
- Model: heavy (opus)
- Dependencies: planner, coder
- Skills: go-arch, go-api, ent-foundation, ent-workflow, ent-tooling

#### Validation Agents (3)

**reviewer** - Code review for bugs, quality, adherence
- Role: validation
- Model: heavy (opus)
- Skills: go-review, ent-foundation, ent-workflow, ent-tooling

**acceptor** - Validate acceptance criteria and requirements
- Role: validation
- Model: main (sonnet)
- Skills: ent-foundation, ent-workflow, ent-tooling

#### Research Agents (1)

**researcher** - Codebase research and deep analysis
- Role: research
- Model: main (sonnet)
- Skills: ent-foundation, ent-workflow, ent-tooling

#### Other Agents (1)

**decomposer** - Task decomposition specialist
- Model: main (sonnet)
- Skills: ent-foundation, ent-workflow, ent-tooling

## Template System

### Architecture

The template system uses a **slot-based composition** model:

```
Agent Definition = Base Template + Section Templates + Agent-Specific Content + Shared Prompts
```

#### Components

1. **Base Template** (`pkg/agents/templates/base-agent.md.tmpl`)
   - Defines overall structure
   - Embeds agent-specific content
   - Embeds shared prompts

2. **Section Templates** (`pkg/agents/templates/sections/`)
   - `_tooling.md.tmpl` - Tool usage guidance (conditional tool restrictions)
   - `_workflow.md.tmpl` - Context gathering workflow (role-specific steps)
   - `_principles.md.tmpl` - Constitutional AI guidance (role-specific examples)
   - `_handoff.md.tmpl` - Agent delegation patterns (uses dependencies)

3. **Agent-Specific Prompts** (`pkg/agents/prompts/agents/`)
   - Role introduction and responsibilities
   - Agent-specific patterns and examples
   - Unique guidance for that agent

4. **Shared Prompts** (`pkg/prompts/`)
   - `foundation.md` - Core decision framework
   - `workflow.md` - OpenSpec workflow essentials
   - `tooling.md` - Tool usage guidance

### Template Parameters

Templates receive an `AgentTemplateData` structure with:

| Parameter | Type | Purpose |
|-----------|------|---------|
| Name | string | Agent name |
| Description | string | Agent description |
| Role | string | execution, planning, validation, research |
| RoleTitle | string | Human-readable role (Implementation, Planning, etc.) |
| Complexity | string | fast, standard, heavy |
| Dependencies | []string | Agents to delegate to |
| Skills | []string | Skills loaded by agent |
| DisallowedTools | []string | Tools agent cannot use |
| HasDisallowedTools | bool | Whether tool restrictions exist |
| AgentContent | string | Agent-specific prompt |
| SharedPrompts | []string | Shared prompt contents |

### Template Helpers

Available in all templates:

- `title` - Title case
- `upper` - Uppercase
- `lower` - Lowercase
- `contains` - Substring check
- `hasPrefix` - Prefix check
- `hasSuffix` - Suffix check
- `join` - Join array
- `replace` - Replace all

### Example Template Usage

```go-template
## Constitutional AI Principles

### Judgment for {{ .RoleTitle }}

Exercise judgment as a thoughtful senior {{ .RoleTitle }} agent.

{{- if eq .Role "execution" }}
**Implementation Judgment Examples:**
- Testing Decisions: Test critical logic, skip trivial getters
{{- else if eq .Role "planning" }}
**Planning Judgment Examples:**
- Task Granularity: Balance detail with usefulness
{{- end }}

{{- if .HasDisallowedTools }}
## CRITICAL: Tool Restrictions

**NEVER use:**
{{- range .DisallowedTools }}
- ❌ `{{ . }}`
{{- end }}
{{- end }}
```

## Architecture

### Directory Structure

```
pkg/
├── agents/
│   ├── meta/                    # Agent metadata (YAML)
│   │   ├── coder.yaml
│   │   ├── planner.yaml
│   │   └── ...
│   ├── prompts/
│   │   └── agents/              # Agent-specific prompts
│   │       ├── coder.md
│   │       ├── planner.md
│   │       └── ...
│   └── templates/               # Template system
│       ├── base-agent.md.tmpl
│       ├── claude.yaml.tmpl     # Claude Code format
│       ├── opencode.yaml.tmpl   # OpenCode format
│       └── sections/
│           ├── _tooling.md.tmpl
│           ├── _workflow.md.tmpl
│           ├── _principles.md.tmpl
│           └── _handoff.md.tmpl
├── prompts/                     # Shared prompts
│   ├── foundation.md
│   ├── workflow.md
│   ├── tooling.md
│   └── _archived/
└── skills/
    └── ent/                     # Skills
        ├── ent-foundation/
        ├── ent-workflow/
        ├── ent-tooling/
        ├── ent-tools-*/
        └── _archived/

.claude/                         # Generated output
├── agents/ent/                  # Generated agents
│   ├── coder.md
│   └── ...
└── skills/ent/ent/              # Copied skills
    ├── ent-foundation/
    └── ...
```

### Generation Flow

1. **Load Metadata** - Read agent YAML files from `pkg/agents/meta/`
2. **Load Tool Presets** - Resolve tool configurations
3. **Load Prompts** - Read agent-specific prompts from `pkg/agents/prompts/agents/`
4. **Load Templates** - Load base and section templates
5. **Assemble Content**:
   - Render agent metadata as YAML frontmatter
   - Assemble agent-specific content with section templates
   - Embed shared prompts (mapped via `sharedPromptToSkill`)
6. **Write Output** - Generate to `.claude/agents/ent/`

### Shared Prompt Mapping

The `sharedPromptToSkill` map in `internal/cli/init.go` converts shared prompt references to skill names:

```go
var sharedPromptToSkill = map[string]string{
    "_foundation": "ent-foundation",
    "_workflow":   "ent-workflow",
    "_tooling":    "ent-tooling",
}
```

This allows agent metadata to reference `_foundation` while the generated agent loads the `ent-foundation` skill.

## Usage

### Creating a New Agent

1. **Create metadata** (`pkg/agents/meta/myagent.yaml`):
```yaml
name: myagent
description: Brief description
model: main
role: execution
skills:
  - go-code
toolPresets:
  - editing
dependencies:
  - tester
prompts:
  shared:
    - _foundation
    - _workflow
    - _tooling
  main: agents/myagent
```

2. **Create prompt** (`pkg/agents/prompts/agents/myagent.md`):
```markdown
You are a [role description].

## Responsibilities
- Specific responsibility 1
- Specific responsibility 2

## Patterns
[Agent-specific patterns and examples]
```

3. **Regenerate**:
```bash
make build
./bin/ent init --tools=claude --force
```

### Modifying Shared Content

**For constitutional AI / judgment / conventions:**
Edit `pkg/skills/ent/ent-foundation/SKILL.md` and `pkg/prompts/foundation.md`

**For OpenSpec workflow / handoffs:**
Edit `pkg/skills/ent/ent-workflow/SKILL.md` and `pkg/prompts/workflow.md`

**For tool usage:**
Edit `pkg/skills/ent/ent-tooling/SKILL.md` and `pkg/prompts/tooling.md`

After editing, regenerate agents:
```bash
make build
./bin/ent init --tools=claude --force
```

### Modifying Templates

**For structural changes affecting all agents:**
1. Edit section templates in `pkg/agents/templates/sections/`
2. Test changes: `go test ./internal/cli -v`
3. Regenerate: `make build && ./bin/ent init --tools=claude --force`

**For role-specific content:**
Use conditionals in templates:
```go-template
{{- if eq .Role "execution" }}
[Execution-specific content]
{{- else if eq .Role "planning" }}
[Planning-specific content]
{{- end }}
```

## Testing

### Unit Tests

Template system tests:
```bash
go test ./internal/cli -v -run "TestLoadSection|TestRenderSection|TestAssemble"
```

### Integration Tests

Regenerate and verify:
```bash
# Backup current output
cp -r .claude/agents/ent/ /tmp/agents-backup/

# Regenerate
make build
./bin/ent init --tools=claude --force

# Verify no unexpected changes
diff -r .claude/agents/ent/ /tmp/agents-backup/
```

### Validation

```bash
# Verify skill loading
grep -r "ent-foundation\|ent-workflow\|ent-tooling" .claude/agents/ent/

# Verify agent count
ls .claude/agents/ent/*.md | wc -l  # Should be 13

# Verify skill count
ls -d .claude/skills/ent/ent/ent-* | grep -v _archived | wc -l  # Should be 7
```

## Migration History

### Phase 1: Skill & Prompt Consolidation (2026-01)
- Consolidated 11 skills → 7 (36% reduction)
- Consolidated 5 prompts → 3 (40% reduction)
- Eliminated 225 lines of duplication (48% reduction)
- Archived deprecated skills and prompts

### Phase 2: Template System Enhancement (2026-02)
- Created 4 section templates with role parameterization
- Built template engine with 12 parameters and 8 helpers
- Added comprehensive test suite (13 tests, 100% coverage)
- Zero breaking changes

### Phase 3: Agent Prompt Migration (2026-02)
- Migrated all 13 agents to template-based generation
- Reduced agent prompts from ~200 lines to ~60 lines (70% reduction)
- Eliminated remaining duplication across agents
- Total maintenance reduction: 65%

## See Also

- [PROMPT_DESIGN.md](PROMPT_DESIGN.md) - Template design patterns and best practices
- [DEVELOPMENT.md](DEVELOPMENT.md) - Template development guide
- [AGENT_INHERITANCE.md](AGENT_INHERITANCE.md) - Agent inheritance system
- [SKILL-AUTHORING.md](SKILL-AUTHORING.md) - Skill authoring guide
