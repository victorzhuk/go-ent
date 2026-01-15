# Compatibility Guide

## Breaking Changes (v0.2.0)

This section documents breaking changes introduced in the agent/command refactoring. **Review carefully before upgrading.**

### Critical Breaking Changes

1. **`--tools` flag now required**
   - **Before**: Auto-detected from metadata
   - **After**: Must specify `--tools claude` or `--tools opencode`
   - **Impact**: All `go-ent init` commands will fail without this flag

2. **Single-file agent format no longer supported**
   - **Before**: `agents/coder.md` with frontmatter + body
   - **After**: Split format with `meta/coder.yaml` + `prompts/agents/coder.md`
   - **Impact**: Existing single-file agents must be migrated using migration script

3. **Agent handoffs require explicit dependencies**
   - **Before**: Handoffs via `@ent:*` worked without validation
   - **After**: Dependencies must be declared in `meta/*.yaml` and validated
   - **Impact**: Handoffs to undeclared agents will fail validation

4. **Agent metadata moved from frontmatter to YAML files**
   - **Before**: All metadata in `---` frontmatter
   - **After**: Metadata in separate `meta/*.yaml` files
   - **Impact**: Custom agents must use new file structure

### Migration Required

If you have existing custom agents or commands, run:

```bash
# Check migration status
go-ent migrate --check

# Execute migration (creates backup automatically)
go-ent migrate --execute
```

The migration script:
- Scans `agents/*.md` for legacy single-file agents
- Extracts frontmatter → `meta/*.yaml`
- Extracts body → `prompts/agents/*.md`
- Infers dependencies from `@ent:*` references
- Creates backups in `.ent-backup/`

### New Features

#### Dependency Validation

```bash
# Validate dependencies explicitly
go-ent init my-service --tools claude --agents coder,debugger

# Include transitive dependencies automatically
go-ent init my-service --tools claude --agents coder --include-deps

# Skip dependency validation
go-ent init my-service --tools claude --agents coder --no-deps
```

#### Dependency Visualization

```bash
# List all dependencies
go-ent agents deps

# Show specific agent dependencies
go-ent agents deps coder

# Tree visualization
go-ent agents deps --tree
```

---

## New Agent Format

### Old Format (Single File)

```
agents/
├── coder.md
├── reviewer.md
└── debugger.md
```

Each file contained:
- YAML frontmatter with metadata
- Markdown body with instructions

### New Format (Split Structure)

```
agents/
├── meta/
│   ├── coder.yaml
│   ├── reviewer.yaml
│   └── debugger.yaml
├── prompts/
│   ├── shared/
│   │   ├── _tooling.md
│   │   ├── _conventions.md
│   │   └── _handoffs.md
│   └── agents/
│       ├── coder.md
│       ├── reviewer.md
│       └── debugger.md
└── templates/
    ├── claude.yaml.tmpl
    └── opencode.yaml.tmpl
```

#### Structure Explained

**`meta/*.yaml`** - Agent metadata
```yaml
name: coder
description: "Go developer. Implements features, writes code."
model: sonnet
color: "#4CAF50"
skills:
  - go-code
  - go-db
tools:
  - read
  - write
  - edit
tags:
  - role:execution
dependencies:
  - debugger
  - tester
prompts:
  shared:
    - _tooling
    - _conventions
    - _handoffs
  main: agents/coder
```

**`prompts/shared/*.md`** - Reusable prompt sections
```markdown
<!-- _tooling.md -->
Use Serena tools for code operations:
- `serena_get_symbols_overview` - Understand file structure
- `serena_find_symbol` - Find specific symbols
- `serena_replace_symbol_body` - Modify code safely
```

**`prompts/agents/*.md`** - Agent-specific instructions
```markdown
You are a senior Go developer implementing features.

{{include "_tooling"}}
{{include "_conventions"}}

## Implementation Workflow

1. Read task from tasks.md
2. Use Serena to explore codebase
3. Implement following patterns
4. Run tests
5. Mark task complete
```

**`templates/*.yaml.tmpl`** - Platform-specific generation
```yaml
# Claude Code format
---
name: {{.Name}}
description: {{.Description}}
tools: {{tools .Tools}}
model: {{.Model}}
skills: {{list .Skills}}
---
{{compose .Name}}
```

#### Benefits of Split Format

1. **Shared prompts** - Common sections (`_tooling`, `_conventions`) defined once
2. **Metadata validation** - JSON schema ensures correctness
3. **Template engine** - `{{include}}` and `{{model}}` functions for composition
4. **Platform flexibility** - Different templates for Claude Code vs OpenCode
5. **Dependency tracking** - Explicit dependencies in meta files

---

## New Command Format

### Old Format (Embedded Domain Knowledge)

Commands contained all workflow logic and domain-specific rules in a single file:

```
commands/
├── plan.md      # Contains agent chains, OpenSpec rules, phases
└── task.md      # Contains task execution, project conventions
```

### New Format (Flows + Domains)

```
commands/
├── flows/
│   ├── plan.md      # Agent chains and phases
│   ├── task.md      # Task execution workflow
│   └── bug.md       # Bug fixing workflow
└── domains/
    ├── openspec.md  # OpenSpec-specific rules
    └── generic.md   # Generic project rules
```

#### Structure Explained

**`flows/*.md`** - Agent workflows
```markdown
# Plan Flow

## Phase 0: Clarification
1. @ent:planner-fast asks clarifying questions
2. Wait for user answers

## Phase 1: Research
1. @ent:researcher investigates unknowns
2. {{include "domains/openspec"}}  <!-- Load OpenSpec rules -->
3. Present findings for approval
```

**`domains/*.md`** - Domain-specific rules
```markdown
<!-- domains/openspec.md -->
When working with OpenSpec changes:
- Always read `openspec/changes/{id}/tasks.md`
- Validate against `openspec/schemas/`
- Use `go_ent_spec_*` MCP tools for spec management
```

#### Template Functions

**`include(name)`** - Include domain content
```markdown
{{include "domains/openspec"}}
{{include "domains/generic"}}
{{include "prompts/shared/_conventions"}}
```

**`if_tool(tool, then, else)`** - Conditional based on platform
```markdown
{{if_tool "claude" "Use MCP tools" "Use Serena tools"}}
```

**`model(category, tool)`** - Resolve model aliases
```markdown
{{model "fast", "claude"}}    → haiku
{{model "main", "opencode"}}  → main
```

#### Benefits of Split Format

1. **Domain separation** - Project-specific rules in `domains/`
2. **Reusability** - Multiple flows can use same domain knowledge
3. **Template evaluation** - `{{include}}` resolved at runtime
4. **Platform independence** - Flows work across Claude Code/OpenCode

---

## Agent Format Compatibility: OpenCode vs Claude Code

The split format applies to both platforms. Only the template generation differs.

### Format Comparison

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

### Strategy 1: Using the Built-in Template Engine (Recommended)

The go-ent plugin includes a template engine that generates platform-specific formats from the split structure.

**No manual conversion needed** - specify `--tools claude` or `--tools opencode` and the plugin generates the correct format.

```bash
# Generate Claude Code agents
go-ent init my-service --tools claude

# Generate OpenCode agents
go-ent init my-service --tools opencode
```

### Strategy 2: Dual Files (Manual)

If you need custom formatting beyond templates:

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

### Strategy 3: Symlinks with Preprocessing

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

### Strategy 4: Claude Code with AGENTS.md Reference

Claude Code can reference AGENTS.md via CLAUDE.md:

```markdown
# CLAUDE.md
Read and follow instructions in AGENTS.md for project conventions.
```

This allows sharing high-level instructions but not agent definitions.

---

## Recommended Format Using Split Structure

### Shared Source (New Format)

```
agents/
├── meta/
│   └── coder.yaml
├── prompts/
│   ├── shared/_tooling.md
│   └── agents/coder.md
└── templates/
    ├── claude.yaml.tmpl
    └── opencode.yaml.tmpl
```

**`meta/coder.yaml`**
```yaml
name: coder
description: "Go developer. Implements features, writes code."
model: sonnet
color: "#4CAF50"
skills:
  - go-code
  - go-db
tools:
  - read
  - write
  - edit
  - bash
tags:
  - role:execution
dependencies:
  - debugger
  - tester
prompts:
  shared:
    - _tooling
    - _conventions
  main: agents/coder
```

**Generated Claude Code Output** (via `--tools claude`)
```yaml
---
name: coder
description: Go developer. Implements features, writes code.
tools: Read, Write, Edit, Bash, Glob, Grep, LS, TodoRead, TodoWrite
model: sonnet
skills: go-code, go-db
---
```

**Generated OpenCode Output** (via `--tools opencode`)
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

---

## Common Prompt Body (Works for Both)

The system prompt body (after frontmatter) is **identical** for both platforms. Only the YAML frontmatter differs.

This means you can:
1. Write the prompt body once in `prompts/agents/*.md`
2. Use different templates (`templates/*.yaml.tmpl`)
3. Template engine generates platform-specific frontmatter

---

## Migration Script Usage

The `go-ent migrate` command handles migration from old single-file to new split format.

### Check Migration Status

```bash
go-ent migrate --check
```

Output example:
```
╔════════════════════════════════════════╗
║  Agent Migration Status              ║
╚════════════════════════════════════════╝

Agent                 Status               Details
--------------------------------------------------------------------------------
coder                 needs_migration      → Needs migration
reviewer              migrated             ✓ Already in split format
debugger              partially_migrated   ⚠ Partial: missing meta

Summary:
  Total:    3
  Migrated: 1
  Partial:  1
  Needs:    1

Run 'go-ent migrate --execute' to migrate agents
```

### Execute Migration

```bash
go-ent migrate --execute
```

What it does:
1. Scans `agents/*.md` for legacy single-file agents
2. Extracts YAML frontmatter → `meta/*.yaml`
3. Extracts markdown body → `prompts/agents/*.md`
4. Infers dependencies from `@ent:*` references
5. Creates backup in `.ent-backup/{timestamp}/`

Backup example:
```
.ent-backup/20260115-103045/
├── coder.md
├── reviewer.md
└── debugger.md
```

### What Gets Migrated

**From single-file (`agents/coder.md`):**
```markdown
---
name: coder
description: Go developer
model: sonnet
skills:
  - go-code
tools:
  - read
  - write
---

You are a senior Go developer.

{{include "_tooling"}}
{{include "_conventions"}}
```

**To split format:**

**`meta/coder.yaml`:**
```yaml
name: coder
description: Go developer
model: sonnet
skills:
  - go-code
tools:
  - read
  - write
dependencies: []
prompts:
  shared: []
  main: agents/coder
```

**`prompts/agents/coder.md`:**
```markdown
You are a senior Go developer.

{{include "_tooling"}}
{{include "_conventions"}}
```

### Manual Migration

If you prefer manual control:

```bash
# 1. Create directory structure
mkdir -p agents/meta
mkdir -p agents/prompts/shared
mkdir -p agents/prompts/agents
mkdir -p agents/templates

# 2. Extract frontmatter to meta/
# Edit agents/coder.md, copy YAML frontmatter to meta/coder.yaml

# 3. Extract body to prompts/
# Edit agents/coder.md, copy markdown body to prompts/agents/coder.md

# 4. Delete original file
rm agents/coder.md
```

---

## Validation Checklist

When creating custom agents in split format:

### Metadata (`meta/*.yaml`)
- [ ] `name` is required and unique
- [ ] `model` uses valid alias (haiku/sonnet/opus)
- [ ] `tools` is array of tool names
- [ ] `skills` is array (for Claude Code) or array (for both)
- [ ] `dependencies` list is valid (referenced agents exist)
- [ ] `prompts.shared` and `prompts.main` are valid paths

### Prompt Files (`prompts/agents/*.md`)
- [ ] File exists at path specified in `prompts.main`
- [ ] Uses `{{include}}` for shared sections
- [ ] No YAML frontmatter (only markdown content)

### Shared Prompts (`prompts/shared/*.md`)
- [ ] Files prefixed with `_` (e.g., `_tooling.md`)
- [ ] Pure markdown content (no frontmatter)
- [ ] Referenced correctly in agent metadata

### Templates (`templates/*.yaml.tmpl`)
- [ ] Uses Go `text/template` syntax
- [ ] Has access to template functions: `include`, `if_tool`, `model`, `list`, `tools`
- [ ] Generates valid YAML frontmatter

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

### New Split Format Benefits (Both Platforms)
- Shared prompt sections (`_tooling`, `_conventions`)
- Dependency validation and resolution
- Template engine for platform generation
- Explicit dependency declarations
- JSON schema validation for metadata
