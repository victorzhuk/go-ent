# Go-Ent Generator

Go-Ent has been transformed from an MCP server into a **generator tool** that compiles unified agent/skill definitions into optimized output for multiple code tools (Claude Code, OpenCode).

## Architecture

```
Source (Unified Format)          ent generate           Per-Tool Output
┌─────────────────────────┐                        ┌─────────────────────────┐
│ src/                    │                        │ .claude/                │
│   agents/               │     ┌─────────┐        │   agents/*.md           │
│     coder.yaml          │────▶│   ent   │───────▶│   (Claude format)       │
│     coder.prompt.md     │     │generate │        └─────────────────────────┘
│   prompts/              │     └─────────┘        ┌─────────────────────────┐
│     shared/*.md         │          │             │ .opencode/              │
│   skills/               │          └────────────▶│   agents/*.md           │
│     go-code/SKILL.md    │                        │   (OpenCode format)     │
└─────────────────────────┘                        └─────────────────────────┘
```

## Quick Start

### 1. Initialize Project

```bash
make init
# or
./bin/ent init --tools=claude,opencode
```

This creates:
- `.claude/agents/` directory
- `.opencode/agents/` directory
- `ent.yaml` configuration with default model mappings

### 2. Customize Source

Edit files in `src/` to define your agents:

**Agent metadata (`src/agents/coder.yaml`):**
```yaml
name: coder
description: Go developer. Implements features, writes code.
model:
  claude: sonnet
  opencode: anthropic/claude-sonnet-4-20250514
skills:
  - go-code
  - go-test
tools:
  claude:
    allowed: [Read, Write, Edit, Bash, Grep, Glob]
    disallowed: [mcp__serena__*]
  opencode:
    read: true
    write: true
    edit: true
    bash: true
prompts:
  shared: [conventions, judgment, principals]
  main: coder
color: "#32CD32"
```

**Agent prompt (`src/agents/coder.prompt.md`):**
```markdown
## Role

Expert Go developer focused on clean implementation...

## Workflow

1. Understand requirements
2. Check existing patterns
3. Implement with minimal changes
4. Write tests
```

### 3. Generate Output

```bash
make generate
# or
./bin/ent generate
```

Output:
```
Generated coder → .claude/agents/coder.md
Generated coder → .opencode/agents/coder.md
Generated planner → .claude/agents/planner.md
Generated planner → .opencode/agents/planner.md

Generation complete!
```

### 4. Validate

```bash
make validate
# or
./bin/ent validate
```

Output:
```
claude (.claude/agents/)
==================================================
✓ coder.md
✓ planner.md
...

opencode (.opencode/agents/)
==================================================
✓ coder.md
✓ planner.md
...

✓ All files valid!
```

## Configuration

Edit `ent.yaml` to customize:

```yaml
tools:
  - claude
  - opencode

models:
  fast:
    claude: haiku
    opencode: anthropic/claude-3-5-haiku-20241022
  main:
    claude: sonnet
    opencode: anthropic/claude-sonnet-4-20250514
  heavy:
    claude: opus
    opencode: anthropic/claude-opus-4-20250514

openspec:
  schema: spec-driven
```

## Directory Structure

```
go-ent/
├── src/                        # SOURCE OF TRUTH
│   ├── agents/                 # Agent definitions
│   │   ├── coder.yaml          # Metadata (model, tools, skills)
│   │   ├── coder.prompt.md     # Agent-specific prompt
│   │   ├── planner.yaml
│   │   └── planner.prompt.md
│   ├── prompts/                # Shared prompt fragments
│   │   ├── conventions.md
│   │   ├── judgment.md
│   │   ├── principals.md
│   │   ├── openspec.md
│   │   └── tooling.md
│   └── skills/                 # Skill definitions
│       └── go-code/SKILL.md
├── .claude/agents/             # Generated Claude Code output
├── .opencode/agents/           # Generated OpenCode output
├── cmd/ent/                    # Generator CLI
├── internal/
│   ├── generator/              # Core generation logic
│   │   ├── generator.go
│   │   ├── source.go
│   │   ├── claude.go
│   │   └── opencode.go
│   └── genconfig/              # Configuration
└── ent.yaml                    # Project configuration
```

## Format Differences

| Aspect | Claude Code | OpenCode |
|--------|-------------|----------|
| Agent file | `.md` with YAML frontmatter | `.md` with YAML frontmatter |
| Model | `sonnet`, `opus`, `haiku` | `anthropic/claude-sonnet-4-*` |
| Tools | `tools: [...]` or `disallowedTools: [...]` | `tools: {read: true, write: false}` |
| Output dir | `.claude/agents/` | `.opencode/agents/` |

## Commands

```bash
# Initialize project
ent init --tools=claude,opencode [--force]

# Generate all agents for all configured tools
ent generate

# Generate for specific tools only
ent generate --tools=claude

# Validate generated output (TODO)
ent validate
```

## Makefile Targets

```bash
make build      # Build ent binary
make init       # Initialize with default tools
make generate   # Generate agent output
make validate   # Validate output (TODO)
```

## How It Works

1. **Source Parser** (`internal/generator/source.go`) - Loads agent YAML + prompt markdown
2. **Inliner** (`internal/generator/inliner.go`) - Combines main + shared prompts
3. **Target Interface** (`internal/generator/target.go`) - Abstracts tool-specific generation
4. **Claude Target** (`internal/generator/claude.go`) - Transforms to Claude Code format
5. **OpenCode Target** (`internal/generator/opencode.go`) - Transforms to OpenCode format

## Next Steps

1. ~~Create source directory structure~~ ✓
2. ~~Build generator core~~ ✓
3. ~~Implement Claude target~~ ✓
4. ~~Implement OpenCode target~~ ✓
5. ~~Create CLI commands~~ ✓
6. ~~Create tool spec schemas for validation~~ ✓
7. ~~Migrate all 16 agents~~ ✓
8. Archive old MCP server code
9. Update documentation

## Benefits

- **Single source of truth**: Edit once in `src/`, generate for all tools
- **Build-time only**: No runtime MCP server needed
- **Fast**: Go binary compiles agents in milliseconds
- **Extensible**: Easy to add new targets (Cursor, Zed, etc.)
- **Type-safe**: YAML validation at load time
- **Customizable**: Per-project model mappings in `ent.yaml`
