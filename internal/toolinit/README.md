# Tool Init Package

Generates tool-specific configurations from embedded go-ent plugin resources.

## Adapters

### Claude Code Adapter

Generates `.claude/` configuration for Claude Code with PLURAL directory structure.

**Directory Structure:**
```
.claude/
├── commands/           # Slash commands (PLURAL)
│   └── plan.md        # Planning workflow only
├── agents/            # Subagents (PLURAL)
│   ├── planner-smoke.md  # Haiku - quick triage
│   ├── architect.md      # Opus - system design
│   ├── planner.md        # Sonnet - detailed planning
│   └── decomposer.md     # Sonnet - task breakdown
└── skills/            # Agent skills (PLURAL)
    ├── core/          # Cross-cutting skills
    │   ├── arch-core/
    │   ├── api-design/
    │   ├── security-core/
    │   ├── review-core/
    │   └── debug-core/
    └── go/            # Go-specific skills
        ├── go-arch/
        ├── go-api/
        ├── go-code/
        ├── go-db/
        ├── go-ops/
        ├── go-perf/
        ├── go-review/
        ├── go-sec/
        └── go-test/
```

**Filtering:**
- **Commands**: Only `plan.md` (planning workflow for Claude Code driver)
- **Agents**: Only 4 planning agents (planner-smoke, architect, planner, decomposer)
- **Skills**: All skills (shared resource, preserves category hierarchy)

**Model Mapping:**
- `opus` → `claude-opus-4-5-20250514`
- `sonnet` → `claude-sonnet-4-5-20250929`
- `haiku` → `claude-haiku-4-5-20250429`

**Frontmatter Format:**

*Agents:*
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

*Commands:*
```yaml
---
name: plan
description: Create complete OpenSpec change proposal
---
```

*Skills:*
```yaml
---
name: go-arch
description: Go architecture patterns and best practices
version: 1.0.0
---
```

### OpenCode Adapter

Generates `.opencode/` configuration for OpenCode with SINGULAR directory structure.

**Directory Structure:**
```
.opencode/
├── command/           # Slash commands (SINGULAR)
│   ├── task.md       # Task execution workflow
│   └── bug.md        # Bug fixing workflow
├── agent/            # Agents (SINGULAR)
│   ├── task-smoke.md     # GLM 4.7 - simple tasks
│   ├── task-heavy.md     # Kimi K2 - complex tasks
│   ├── coder.md          # GLM 4.7 - implementation
│   ├── reviewer.md       # GLM 4.7 - code review
│   ├── tester.md         # GLM 4.7 - testing
│   ├── acceptor.md       # GLM 4.7 - acceptance validation
│   ├── reproducer.md     # GLM 4.7 - bug reproduction
│   ├── researcher.md     # GLM 4.7 - investigation
│   ├── debugger-smoke.md # GLM 4.7 - simple debugging
│   └── debugger-heavy.md # Kimi K2 - complex debugging
└── skill/            # Agent skills (SINGULAR, flattened)
    ├── core-arch-core/SKILL.md
    ├── core-api-design/SKILL.md
    ├── go-arch/SKILL.md
    ├── go-code/SKILL.md
    └── ...
```

**Filtering:**
- **Commands**: Only `task.md` and `bug.md` (execution workflows for OpenCode worker)
- **Agents**: Only 10 execution agents (task-smoke, task-heavy, coder, reviewer, tester, acceptor, reproducer, researcher, debugger-smoke, debugger-heavy)
- **Skills**: All skills (flattened with category prefix: `core-arch-core`, `go-code`)

**Model Mapping:**
- `glm-4-flash` → `zhipu/glm-4-flash`
- `kimi-k2` → `moonshot/kimi-k2`
- `opus` → `anthropic/claude-opus-4-5-20250514`
- `sonnet` → `anthropic/claude-sonnet-4-5-20250929`
- `haiku` → `anthropic/claude-haiku-4-5-20250429`

**Frontmatter Format:**

*Agents:*
```yaml
---
description: Execute simple tasks efficiently. Fast implementation for straightforward work.
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

*Commands:*
```yaml
---
description: Execute OpenSpec tasks with TDD and validation
---
```

*Skills:*
```yaml
---
name: go-code
description: Go implementation patterns
---
```

**Key Differences from Claude Code:**
- **Directory naming**: SINGULAR (command/, agent/, skill/) vs PLURAL (commands/, agents/, skills/)
- **Skill structure**: Flattened with category prefix vs hierarchical categories
- **Agent mode**: All agents are "subagent" mode
- **Temperature**: Fixed 0.0 for deterministic code generation
- **Tools**: Explicit enable/disable per tool
- **Permissions**: Skill-based permission system

## Usage

```go
import "github.com/victorzhuk/go-ent/internal/toolinit"

// Create Claude Code adapter
claudeAdapter := toolinit.NewClaudeAdapter()

// Create OpenCode adapter
opencodeAdapter := toolinit.NewOpenCodeAdapter()

// Generate configuration
cfg := &toolinit.GenerateConfig{
    Path:      "/path/to/project",
    PluginFS:  goent.PluginFS,
    Force:     false,
    DryRun:    false,
}

// Generate Claude Code config
err := claudeAdapter.Generate(context.Background(), cfg)

// Generate OpenCode config
err = opencodeAdapter.Generate(context.Background(), cfg)
```

## Testing

Run tests:
```bash
go test ./internal/toolinit/ -v
```

Test coverage:

**Claude Adapter:**
- ✅ Name and TargetDir methods
- ✅ Agent transformation with model mapping (Opus, Sonnet, Haiku)
- ✅ Command transformation
- ✅ Skill transformation
- ✅ Model name mapping

**OpenCode Adapter:**
- ✅ Name and TargetDir methods
- ✅ Agent transformation with model mapping (GLM 4.7, Kimi K2)
- ✅ Command transformation
- ✅ Skill transformation
- ✅ Model name mapping (zhipu, moonshot providers)
- ✅ Agent permissions and tool configurations
- ✅ SINGULAR directory naming

## Architecture

```
┌─────────────────────┐
│   Embedded FS       │
│ (goent.PluginFS)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Adapter           │
│  - Name()           │
│  - TargetDir()      │
│  - Generate()       │
│  - Transform*()     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   File Operations   │
│  - Filter resources │
│  - Transform format │
│  - Write to disk    │
└─────────────────────┘
```

## Design Decisions

1. **Filtering by Tool**: Each adapter filters resources based on tool's role
   - Claude Code: Planning agents + `/plan` command
   - OpenCode: Execution agents + `/task`, `/bug` commands

2. **Category Preservation**: Skills maintain category structure for both tools
   - Claude Code reads from `.claude/skills/{category}/{skill}/SKILL.md`
   - OpenCode can also read from `.claude/skills/` for compatibility

3. **Model Mapping**: Internal short names map to full API model IDs
   - Simplifies plugin resource management
   - Easy to update model versions centrally

4. **Dry Run Support**: Preview changes before applying
   - Lists files that would be created
   - No modifications to filesystem

5. **Force Mode**: Allow overwriting existing configurations
   - Useful for updates
   - Safety check by default
