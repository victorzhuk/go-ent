# Project Context: go-ent

## Purpose

**go-ent** is an enterprise Go development toolkit that implements the OpenSpec workflow methodology. It provides:

1. **MCP Server**: Model Context Protocol server exposing 30+ tools for AI assistants
2. **Claude Code Plugin**: Full integration with Claude Code CLI (agents, commands, skills)
3. **OpenCode Plugin**: Compatible with OpenCode AI IDE
4. **OpenSpec Workflow**: Spec-driven development with change proposals, delta specs, and artifact management

The project dogfoods itself - all development uses the /ent:* commands and agents defined in the plugin.

**Key Goals**:
- Enable AI-assisted enterprise Go development with best practices built-in
- Provide spec-first workflow for change management and documentation
- Offer reusable skills/agents for Go architecture, testing, API design, etc.
- Support both CLI (Claude Code) and IDE (OpenCode) environments

## Tech Stack

### Core Technologies
- **Go 1.24+**: All implementation code
- **MCP Protocol**: stdio transport for tool/resource/prompt serving
- **BoltDB**: Embedded KV store for state management (agent memory, task registry)
- **YAML**: Agent/skill/command configuration format
- **Markdown**: Documentation and prompt engineering

### Key Go Libraries
- `gopkg.in/yaml.v3`: YAML parsing for configs
- `go.etcd.io/bbolt`: BoltDB wrapper for state
- `github.com/spf13/cobra`: CLI command framework
- `embed` package: Embedded filesystem for plugin assets

### Development Tools
- `golangci-lint`: Linting (see .golangci.yml)
- `make`: Build automation (see Makefile)
- `rg` (ripgrep): Fast code search
- `fd`: Fast file finding

## Project Conventions

### Code Style

**Philosophy**: SOLID, DRY, KISS, YAGNI with Clean Architecture + DDD

**Critical Rules**:
1. **Zero comments except WHY** - Comments explaining WHAT = bad naming. Fix the name instead.
2. **Natural variable names** - Use `cfg`, `repo`, `srv`, not `applicationConfiguration`
3. **Private by default** - Only expose what's necessary
4. **Errors are lowercase** - `fmt.Errorf("query user: %w", err)` not `"Failed to query"`
5. **Happy path left** - Early returns for error cases

**Naming Conventions**:
- Variables: `cfg`, `repo`, `srv`, `pool`, `ctx`, `req`, `resp`, `err`
- Constructors: `New()` public, `new*()` private
- Structs: Private unless domain entity (`type app struct` vs `type User struct`)
- Receivers: Short (`s *service`, `a *Agent`)

**File Organization**:
```
package/
├── contract.go      # public interfaces/types
├── app.go          # main application struct
├── lifecycle.go    # Start/Stop methods
├── uc.go           # use cases (if applicable)
└── <domain>.go     # domain-specific logic
```

### Architecture Patterns

**Clean Architecture Layers**:
```
cmd/           # Entry points (CLI, MCP server)
internal/
  ├── domain/   # Pure business logic (ZERO external deps)
  ├── usecase/  # Orchestration + business rules
  ├── repo/     # Data access (interfaces in usecase/)
  └── adapter/  # External integrations
```

**Dependency Flow**: `cmd → adapter → usecase → domain` (inward only)

**Key Patterns**:
- **Dependency Injection**: Constructor-based, no DI frameworks
- **Repository Pattern**: Abstract data access behind interfaces
- **Domain-Driven Design**: Bounded contexts, ubiquitous language
- **Functional Options**: For complex constructors

**MCP Tools**:
- Must be **idempotent** where possible
- Must validate inputs rigorously
- Must return structured errors
- Should support `--dry-run` for state changes

### Testing Strategy

**Approach**: Table-driven tests with testify assertions

**Coverage Targets**:
- Core packages (`internal/domain`, `internal/config`): >80%
- Use cases: >70%
- Adapters: Integration tests with real dependencies

**Test Structure**:
```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name string
        // fields
    }{
        {"valid input", ...},
        {"empty input", ...},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // test with tt fields
        })
    }
}
```

**Integration Tests**:
- Use `testcontainers` for databases
- Use real implementations over mocks when simple
- Test context is `t.Context()`

**Avoid**:
- Mocking simple things (use real `net.Listen`)
- `reflect.DeepEqual` (test specific fields)
- Testing unexported functions directly

### Git Workflow

**Branching**: `master` is main branch (not `main`)

**Commit Messages**:
- Format: `type(scope): description` (lowercase)
- Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
- Examples:
  - `feat(cli): add init command for plugin setup`
  - `fix(mcp): handle nil pointer in tool validation`
  - `refactor(agent): simplify prompt template logic`

**OpenSpec Workflow**:
1. Create change: `/ent:plan "description"`
2. Iterate on artifacts: proposal → design → tasks
3. Execute tasks: `/ent:apply`
4. Archive when deployed: `/ent:archive change-id`

**Changes live in**: `openspec/changes/<change-id>/`

**Artifacts**:
- `proposal.md`: Why, what, impact, success criteria
- `design.md`: Technical approach (optional for complex changes)
- `tasks.md`: Implementation checklist
- `specs/<domain>/spec.md`: Delta specs for domain changes

## Domain Context

### OpenSpec Workflow

**Core Concepts**:
- **Change**: A logical unit of work (feature, fix, refactor) with unique ID
- **Proposal**: High-level "why" and "what" for the change
- **Design**: Technical "how" with diagrams, data flows, interfaces
- **Tasks**: Checklist of implementation steps with status tracking
- **Delta Spec**: Domain-specific specification changes (merged to main specs on archive)

**Change Lifecycle**:
```
active/ → (implementation) → archive/
```

**Registry**: BoltDB-backed task tracker (supports distributed agent work)

### Agent System

**Agent Types** (17 total):
- **Meta Agents**: Core roles (architect, planner, coder, tester, debugger, reviewer)
- **Specialized**: Domain-specific (acceptor, decomposer, researcher, reproducer)
- **Variants**: Fast/heavy versions for latency vs depth tradeoffs

**Agent Configuration** (YAML):
```yaml
name: coder
description: "..."
model: sonnet
tools: [Read, Write, Edit, Bash, ...]
system_prompt: "prompts/agents/coder.md"
```

**Prompt Structure**:
- Base prompts in `agents/prompts/agents/<name>.md`
- Shared fragments in `agents/prompts/shared/_*.md`
- Template rendering for plugin format (Claude vs OpenCode)

### Skill System

**Skill Categories**:
- **Core**: Universal patterns (arch, api-design, security, review, debug)
- **Go**: Language-specific (go-code, go-arch, go-test, go-api, go-db, etc.)

**Skill Frontmatter** (YAML):
```yaml
name: go-code
description: "..."
version: "2.0.0"
author: "go-ent"
license: "MIT"
compatibility:
  claude_code: ">=1.0"
  opencode: ">=0.1"
tags: ["go", "code"]
quality_score: 94
category: "go"
depends_on: ["other-skill"]  # optional
```

**Auto-Activation**: Skills define triggers for automatic loading based on context

### Command System

**Command Types**:
- **Workflows**: `/ent:plan`, `/ent:apply`, `/ent:status`, `/ent:archive`
- **Registry**: `/ent:registry list`, `/ent:registry next`
- **Skill Management**: `/ent:skill-sync` (sync to Claude Code)
- **Skills**: Auto-activate based on task content (see `task-router` skill for routing)

**OpenSpec Aliases** (added in this refactor):
- `/opsx:new` → `/ent:plan`
- `/opsx:apply` → `/ent:apply`
- `/opsx:archive` → `/ent:archive`

## Important Constraints

### Technical Constraints

1. **Domain Layer**: ZERO external dependencies allowed
   - Pure business logic only
   - Interfaces define contracts with outer layers

2. **MCP Tools**: Must be idempotent or clearly document side effects
   - State changes should support dry-run mode
   - Errors must be actionable

3. **Backward Compatibility**: Plugin format must support both Claude Code and OpenCode
   - Use template rendering for platform-specific differences
   - Maintain schema versions for agents/skills/commands

4. **Embedded Assets**: All plugin files embedded in binary via `//go:embed`
   - Changes require rebuild
   - Hot-reload not supported (by design)

5. **BoltDB State**: Single-writer, multiple-reader pattern
   - No distributed locking
   - Agent coordination via task registry only

### Code Quality Constraints

1. **No AI-Style Code**: Write like a senior dev, not a tutorial
2. **No Over-Engineering**: YAGNI - implement when needed, not "just in case"
3. **No Magic Numbers**: Use named constants
4. **No `helper`/`util` packages**: Use domain names instead
5. **No Global Mutable State**: Dependency injection only

### Development Constraints

1. **Dogfooding Required**: All development uses `/ent:*` workflow
2. **OpenSpec Compliance**: Changes must have proposal/tasks in `openspec/changes/`
3. **Simplicity First**: Impact minimal code, avoid complexity

## External Dependencies

### MCP Protocol
- **Spec**: https://modelcontextprotocol.io
- **Transport**: stdio only (no HTTP/SSE)
- **Used for**: Tool/resource/prompt serving to Claude Code

### Claude Code CLI
- **Integration**: Plugin system via `~/.claude/plugins/go-ent/`
- **Required**: Agents, skills, commands in plugin format
- **Hot-reload**: Not supported (requires restart)

### OpenCode IDE
- **Integration**: Same plugin format, different rendering
- **Compatibility**: Template-based with `.opencode.yaml.tmpl`

### Build Dependencies
- Go 1.24+ toolchain
- `make` for build automation
- `golangci-lint` for linting
- No runtime dependencies (BoltDB, YAML parser embedded)

### Optional Tools (for developers)
- `rg` (ripgrep): Fast search - PREFERRED over `grep`
- `fd`: Fast find - PREFERRED over `find`
- `bat`: Syntax-highlighted cat
- `eza`/`exa`: Better ls

## Architecture Decision Records

### ADR-001: Embedded Plugin Assets
**Decision**: Use `//go:embed` for all plugin files (agents, skills, commands)
**Rationale**: Single binary distribution, no file path dependencies
**Trade-off**: Requires rebuild for changes (acceptable for plugin stability)

### ADR-002: BoltDB for State
**Decision**: Use BoltDB for task registry and agent memory
**Rationale**: Embedded, zero-config, ACID transactions
**Trade-off**: Single-writer only (sufficient for current use case)

### ADR-003: OpenSpec Workflow
**Decision**: Adopt OpenSpec artifacts (proposal/design/tasks/specs)
**Rationale**: Spec-first development improves clarity and documentation
**Trade-off**: More upfront planning (beneficial for quality)

### ADR-004: Dual Plugin Format
**Decision**: Support both Claude Code and OpenCode via templates
**Rationale**: Maximize compatibility across AI IDEs
**Trade-off**: Template maintenance overhead (mitigated by generation)

### ADR-005: Tool Naming Convention
**Decision**: Use `rg`/`fd` instead of `grep`/`find` in prompts
**Rationale**: 10-100x faster, better defaults, respects .gitignore
**Trade-off**: Requires tools installed (acceptable for dev environment)
