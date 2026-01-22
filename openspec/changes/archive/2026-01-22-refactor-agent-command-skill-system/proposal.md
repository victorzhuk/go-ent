# Proposal: Refactor Agent, Command, and Skill System

## Why

The current agent/command/skill system has structural limitations that prevent flexible, composable, and maintainable agent definitions:

1. **Monolithic agent files** - Frontmatter and prompts are coupled in single files, making it hard to share common sections
2. **Implicit dependencies** - Agent handoffs are described in text, not validated structurally
3. **Hardcoded domain logic** - Commands embed OpenSpec-specific knowledge, preventing reuse for generic projects
4. **No selective generation** - `ent init` generates all agents; can't filter by role or team
5. **Auto-detection ambiguity** - When both `.claude/` and `.opencode/` exist, unclear which to update

This refactoring enables:
- **Composable prompts** - Shared sections (_tooling, _conventions) reduce duplication
- **Dependency validation** - Explicit graph prevents generating incomplete agent sets
- **Domain-agnostic flows** - Commands reference runtime-loaded domains
- **Selective agent generation** - Teams can install only needed agents (e.g., only planning agents for Claude Code)
- **Explicit tool selection** - Required `--tools` flag eliminates ambiguity

Inspired by:
- [docs/COMPATIBILITY_GUIDE.md](/home/zhuk/Projects/own/go-ent/docs/COMPATIBILITY_GUIDE.md:1-263) - Dual OpenCode/Claude Code formats
- [docs/REFACTORING_GUIDE.md](/home/zhuk/Projects/own/go-ent/docs/REFACTORING_GUIDE.md:1-409) - Tool unification patterns
- [prompts/README.md] (referenced in code) - Template system proposal

## What Changes

### Architecture

Split single-file agents into composable components:

```
plugins/go-ent/
├── agents/
│   ├── meta/*.yaml          # Metadata + dependencies
│   ├── prompts/
│   │   ├── shared/*.md      # Reusable sections (_tooling, _conventions)
│   │   └── agents/*.md      # Agent-specific prompts
│   └── templates/*.tmpl     # Tool-specific frontmatter
├── commands/
│   ├── flows/*.md           # Generic workflows (plan, task, bug)
│   └── domains/*.md         # Domain knowledge (openspec, generic)
└── skills/
    └── prompts/*.md         # Shared skill sections
```

### CLI Changes

Make `--tools` required, add dependency management:

```bash
# Required: explicit tool selection
ent init --tools claude
ent init --tools claude,opencode

# Selective agent generation
ent init --tools claude --agents coder,tester

# Dependency handling
ent init --tools claude --agents coder --include-deps  # Auto-resolve
ent init --tools claude --agents coder --no-deps       # Skip validation
```

### Dependency Graph

Explicit agent dependencies with validation:

```yaml
# meta/coder.yaml
dependencies:
  - tester
  - reviewer
  - debugger
```

By default, `ent init --agents coder` fails unless dependencies included.

### Command Domain Loading

Commands become thin wrappers: flow + domain knowledge:

```markdown
<!-- commands/flows/plan.md -->
{{include "domains/openspec"}}  # Runtime-loaded

## Agent Chain
| @ent:planner-fast | Feasibility | fast |
| @ent:architect    | Design      | heavy |
```

## Impact

### Affected Specs
- **agent-system** - Core refactor (breaking change)
- **cli-build** - New flags (`--tools` required, `--include-deps`, `--no-deps`)
- **mcp-tools** - Agent metadata schema

### Affected Code
- **internal/toolinit/** - Adapter interface extended with composer
- **internal/cli/init.go** - Required `--tools`, dependency validation
- **internal/agent/** (NEW) - Dependency graph, composer, validator
- **plugins/go-ent/agents/** - Restructure to meta + prompts + templates
- **plugins/go-ent/commands/** - Split into flows + domains

### Dependencies
- Go stdlib `text/template` for composition
- `gopkg.in/yaml.v3` for meta parsing (already used)

### Breaking Changes
- `ent init` requires `--tools` flag (no auto-detect)
- Single-file `.md` agents no longer supported (migration script provided)
- Agent handoffs require explicit dependencies in `meta/*.yaml`

## Key Benefits

1. **Reduced Duplication** - Shared prompts (_tooling, _conventions) used by all 18 agents
2. **Explicit Dependencies** - Dependency graph prevents incomplete agent sets, visualized with `ent agents deps --tree`
3. **Flexible Team Setup** - Teams install only needed agents (e.g., planning-only for architects)
4. **Domain Flexibility** - Same flow works for OpenSpec, generic, or custom project domains
5. **Tool Clarity** - Required `--tools` eliminates "which tool?" ambiguity
6. **Better Testing** - Dependency resolution, validation, composition all unit-testable

## Migration Path

1. **Phase 1**: Parallel support - Both old single-file and new split formats work
2. **Phase 2**: Migration script - `ent migrate` converts existing projects
3. **Phase 3**: Deprecation warnings - Old format logs warnings
4. **Phase 4**: Removal - Clean up old format support (target: v4.0.0)

Migration script handles:
- Extract frontmatter → `meta/*.yaml`
- Extract body → `prompts/agents/*.md`
- Analyze handoffs → infer `dependencies` field
- Generate backups before transformation

## Success Criteria

- [ ] `ent init --tools claude` generates agents from new split format
- [ ] `ent init --tools claude --agents coder` fails with missing dependency error
- [ ] `ent init --tools claude --agents coder --include-deps` auto-includes tester, reviewer, debugger
- [ ] Commands use `{{include "domains/openspec"}}` for runtime domain loading
- [ ] `ent agents deps --tree` visualizes dependency graph
- [ ] `ent migrate` converts existing single-file agents
- [ ] All existing integration tests pass with new format
- [ ] Documentation updated (SETUP_GUIDE, COMPATIBILITY_GUIDE, plugin README)

## Open Questions

- [ ] Should skills also split into skill + prompt? (Optional, can defer to future)
- [ ] Should we support command composition at generation time vs runtime include? (Decision: runtime include for flexibility)
- [ ] How to handle circular dependencies? (Decision: fail with clear error, no auto-resolution)
