# Multi-Platform Agent Setup

This guide explains how to set up agents for both **OpenCode** and **Claude Code** using the go-ent plugin.

---

## Directory Structure

```
plugins/go-ent/
├── agents/
│   ├── meta/                    # Agent metadata (YAML)
│   │   ├── coder.yaml
│   │   ├── planner.yaml
│   │   └── ...
│   ├── prompts/
│   │   ├── agents/              # Agent prompts (markdown)
│   │   │   ├── coder.md
│   │   │   ├── planner.md
│   │   │   └── ...
│   │   └── shared/              # Shared prompt sections
│   │       ├── _tooling.md
│   │       ├── _conventions.md
│   │       ├── _handoffs.md
│   │       └── _openspec.md
│   └── templates/               # Platform-specific templates
│       ├── claude.yaml.tmpl
│       └── opencode.yaml.tmpl
└── commands/
    ├── flows/                   # Command flows
    │   ├── plan.md
    │   ├── task.md
    │   └── bug.md
    ├── domains/                 # Domain knowledge
    │   ├── openspec.md
    │   ├── generic.md
    │   └── README.md
    └── *.md                     # Generated commands

project/
├── .claude/
│   └── agents/                  # Generated Claude Code agents
│       ├── coder.md
│       ├── planner.md
│       └── ...
├── .opencode/
│   └── agent/                   # Generated OpenCode agents
│       ├── coder.md
│       ├── planner.md
│       └── ...
├── .go-ent/
│   └── config.yaml             # Plugin configuration
├── CLAUDE.md                    # Claude Code config
├── opencode.json                # OpenCode config
└── AGENTS.md                    # Shared instructions
```

---

## Quick Start

```bash
# Initialize Claude Code agents
ent init --tool=claude

# Initialize OpenCode agents
ent init --tool=opencode

# Initialize both platforms
ent init --tool=all
```

---

## OpenCode Setup

### 1. Initialize Configuration

```bash
ent init --tool=opencode
```

This creates:
- `.opencode/agent/` directory
- Generated agent files with metadata + prompts
- `opencode.json` configuration

### 2. Agent Metadata Format

Agents are defined in `plugins/go-ent/agents/meta/*.yaml`:

```yaml
name: coder
description: Go developer. Implements features, writes code.
model: main
color: "#32CD32"
skills:
  - go-code
  - go-db
tools:
  - read
  - write
  - edit
  - bash
  - glob
  - grep
  - mcp__plugin_serena_serena
dependencies:
  - tester
  - reviewer
  - debugger
tags:
  - "role:execution"
  - "complexity:standard"
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

### 1. Initialize Configuration

```bash
ent init --tool=claude
```

This creates:
- `.claude/agents/` directory
- Generated agent files with metadata + prompts
- `CLAUDE.md` configuration

### 2. Agent File Format

Generated agents in `.claude/agents/*.md`:

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
  mcp__plugin_serena_serena: true
model: main
color: "#32CD32"
tags:
  - "role:execution"
  - "complexity:standard"
skills:
  - go-code
  - go-db
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

## Dependency Management

Agents can declare dependencies on other agents. The go-ent plugin provides several flags to manage these dependencies.

### Dependency Graph

View dependency relationships:

```bash
# Show all dependencies
ent agents deps

# Show dependencies for a specific agent
ent agents deps coder

# Show dependency tree
ent agents deps --tree
```

### Selective Agent Initialization

```bash
# Initialize specific agents only
ent init --tool=claude --agents coder,tester

# Initialize specific agents and auto-resolve transitive dependencies
ent init --tool=claude --agents coder --include-deps

# Initialize specific agents without dependency validation
ent init --tool=claude --agents coder --no-deps
```

### Dependency Flags

| Flag | Description |
|------|-------------|
| `--agents` | Comma-separated list of agent names to include |
| `--include-deps` | Auto-resolve transitive dependencies for selected agents |
| `--no-deps` | Skip dependency validation (use with caution) |

**Note**: `--include-deps` and `--no-deps` are mutually exclusive.

### Example Workflow

```bash
# You want to add only the coder agent to your project
ent init --tool=claude --agents coder

# Error: agent 'coder' has dependencies: tester, reviewer, debugger
# Either include dependencies explicitly or use --include-deps

# Option 1: Include all dependencies automatically
ent init --tool=claude --agents coder --include-deps

# Option 2: Include dependencies explicitly
ent init --tool=claude --agents coder,tester,reviewer,debugger

# Option 3: Skip validation (not recommended)
ent init --tool=claude --agents coder --no-deps
```

---

## Migration from Old Format

The go-ent plugin now uses a split format (metadata + prompts) instead of single-file agents with frontmatter.

### Old Format

Single `.md` files with YAML frontmatter and markdown body:

```markdown
---
name: coder
description: Go developer. Implements features.
model: main
skills:
  - go-code
  - go-db
---

You are a senior Go backend developer...
```

### New Format

Split into three parts:

1. **Metadata** (`meta/*.yaml`):
```yaml
name: coder
description: Go developer. Implements features.
model: main
skills:
  - go-code
  - go-db
tools:
  - read
  - write
dependencies:
  - tester
  - reviewer
```

2. **Prompt** (`prompts/agents/*.md`):
```markdown
You are a senior Go backend developer...
```

3. **Shared Sections** (`prompts/shared/*.md`):
```markdown
# Tooling Reference
Common patterns for Serena, Git, Go...
```

### Manual Migration Steps

1. **Extract metadata** from frontmatter into `meta/agentname.yaml`:
   - Move `name`, `description`, `model`, `skills`
   - Add `tools` list
   - Add `dependencies` list (analyze `@ent:*` references)

2. **Extract body** into `prompts/agents/agentname.md`:
   - Remove YAML frontmatter
   - Replace duplicated sections with `{{include "shared/section"}}`

3. **Create shared sections** in `prompts/shared/*.md`:
   - Extract common tooling patterns
   - Extract shared conventions
   - Extract handoff patterns

### Automated Migration

The go-ent plugin provides a migration command:

```bash
# Check migration status
ent migrate --check

# Perform migration (creates backup)
ent migrate --execute

# Preview changes without executing
ent migrate --dry-run
```

The migration command:
- Scans `agents/*.md` files
- Extracts frontmatter → `meta/*.yaml`
- Extracts body → `prompts/agents/*.md`
- Infers dependencies from `@ent:*` references
- Creates backups before modifying files

---

## Format Quick Reference

### Tool Names

| OpenCode | Claude Code |
|----------|-------------|
| `read` | `read: true` |
| `write` | `write: true` |
| `edit` | `edit: true` |
| `bash` | `bash: true` |
| `grep` | `grep: true` |
| `glob` | `glob: true` |
| `list` | `list: true` |
| `todoread` | `todoread: true` |
| `todowrite` | `todowrite: true` |
| `webfetch` | `webfetch: true` |
| `websearch` | `websearch: true` |

### Model Names

| OpenCode | Claude Code |
|----------|-------------|
| `fast` | `fast` |
| `main` | `main` |
| `heavy` | `heavy` |

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

## Advanced Usage

### Model Overrides

Override the default model for specific agent patterns:

```bash
# Use opus for all heavy agents
ent init --tool=claude --model heavy=opus

# Use opus for planning agents tagged with heavy
ent init --tool=claude --model "planning:heavy=opus"

# Multiple overrides
ent init --tool=claude --model heavy=opus --model fast=haiku
```

### Update Existing Configuration

Update existing configuration without full reinitialization:

```bash
# Update all components
ent init --tool=claude --update

# Update only agents
ent init --tool=claude --update --update-filter=agents

# Update agents and commands
ent init --tool=claude --update --update-filter=agents,commands
```

### Preview Changes

Preview what would be generated without writing files:

```bash
ent init --tool=claude --dry-run
```

### Force Overwrite

Force overwrite existing configuration:

```bash
ent init --tool=claude --force
```

---

## Commands

The plugin also provides commands for common workflows:

### Plan Workflow

```bash
# Generate plan command
ent init --tool=claude --update --update-filter=commands

# Use the command
/go-ent:plan Add user authentication
```

### Task Execution

```bash
# Execute specific task
/go-ent:apply
```

### Debug Workflow

```bash
# Generate debug command
ent init --tool=claude --update --update-filter=commands

# Use the command
/go-ent:bug Fix authentication failing on login
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

### Dependency Validation

```bash
# Validate all dependencies
ent agents deps

# Check for cycles or missing dependencies
ent agents deps --validate
```

---

## Best Practices

1. **Use dependency management** - Let the plugin resolve dependencies automatically with `--include-deps`
2. **Keep prompts modular** - Use shared sections to avoid duplication
3. **Reference shared AGENTS.md** - Common conventions reduce maintenance
4. **Test on both platforms** - Ensure compatibility if using both Claude Code and OpenCode
5. **Version control configuration** - Track `.go-ent/config.yaml`, `CLAUDE.md`, `opencode.json`
6. **Use model overrides wisely** - Balance speed vs. quality based on task complexity
7. **Preview before applying** - Use `--dry-run` to check what will be generated
8. **Update incrementally** - Use `--update` with filters to update specific components

---

## Troubleshooting

### Missing Dependencies Error

```
Error: agent 'coder' has dependencies: tester, reviewer, debugger
```

**Solution**: Use `--include-deps` to auto-resolve, or explicitly list all dependencies with `--agents`.

### Tool Not Found

```
Error: tool not found: some_tool
```

**Solution**: Check the agent metadata file. The `tools` list must match the platform's available tools.

### Template Rendering Error

```
Error: template execution failed: template: agent:1:10: executing...
```

**Solution**: Verify the template syntax in `plugins/go-ent/agents/templates/*.tmpl` and ensure all required fields are present.

### Cycle Detected

```
Error: dependency cycle detected: coder -> tester -> coder
```

**Solution**: Remove circular dependencies from agent metadata files. The dependency graph must be a DAG (Directed Acyclic Graph).
