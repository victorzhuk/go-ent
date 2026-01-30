# Generate Agent Configs

## Summary

Create a command to generate optimized agent configurations for Claude Code and OpenCode based on project specifications. Move from static agent definitions to dynamic, project-specific generation.

## Problem

Current agent setup has these issues:

1. **Static definitions**: Agents are hardcoded in `internal/cli/.claude/` and `.opencode/`
2. **One-size-fits-all**: Same agents for all projects regardless of tech stack
3. **Dual maintenance**: Separate files for Claude Code and OpenCode
4. **Embedded in binary**: Changes require rebuild
5. **Not leveraging native formats**: Custom format instead of native agent definitions

## Solution

### New Command: `ent generate agents`

Generates platform-specific agent configurations from project specs:

```bash
# Generate for current project
ent generate agents

# Generate for specific platform
ent generate agents --platform claude
ent generate agents --platform opencode
ent generate agents --platform both

# Output to specific directory
ent generate agents --output ./my-agents
```

### Generated Structure

**Claude Code** (`.claude/agents/`):
```yaml
# .claude/agents/go-architect.md
---
name: go-architect
description: System architect for Go projects. Designs components, layers, data flow.
model: sonnet
skills:
  - go-arch
  - api-design
tools:
  - Read
  - Glob
  - Grep
---

Expert Go system architect. Focus on Clean Architecture, DDD, and idiomatic Go patterns.

## Responsibilities

- Design system architecture
- Define component boundaries
- Create data flow diagrams
- Make technology decisions
```

**OpenCode** (`.opencode/agents/`):
```yaml
# .opencode/agents/go-architect.md
---
name: go-architect
description: System architect for Go projects
type: agent
model: main
skills:
  - go-arch
  - api-design
---

Expert Go system architect...
```

### Project-Aware Generation

Agents are generated based on:

1. **Project type** (detected from files):
   - Go projects → Go-specific agents
   - Web projects → Frontend agents
   - Mixed → Full stack agents

2. **OpenSpec specs** (`openspec/specs/`):
   - Domain specs → Domain expert agents
   - API specs → API designer agents
   - Infrastructure specs → DevOps agents

3. **Configuration** (`.go-ent/config.yaml`):
   ```yaml
   agents:
     include:
       - go-architect
       - go-coder
       - go-tester
     exclude:
       - python-coder  # Not needed for Go project
     custom:
       - name: my-domain-expert
         description: Expert in our domain
         skills:
           - domain-knowledge
   ```

## Affected Systems

- New: `internal/cli/generate.go` - Generate command
- New: `internal/generate/agents.go` - Agent generation logic
- New: `internal/generate/claude.go` - Claude Code format
- New: `internal/generate/opencode.go` - OpenCode format
- New: `internal/generate/detect.go` - Project type detection
- `internal/cli/root.go` - Add generate command
- `internal/config/config.go` - Add agent config section
- Remove: `internal/cli/.claude/agents/` - Static files
- Remove: `internal/cli/.opencode/agents/` - Static files

## Breaking Changes

- [x] Remove static agent files from binary
- [x] Add generate command
- [x] Change agent distribution model

## Migration Path

1. Create generation logic
2. Generate agents for go-ent itself
3. Test with both Claude Code and OpenCode
4. Remove static agent files
5. Update documentation

## Alternatives Considered

1. **Keep static files**: Rejected - not project-specific
2. **Runtime agent registry**: Rejected - too complex for MVP
3. **Generate per-project (chosen)**: Flexible, project-specific, maintainable

## Success Criteria

- [ ] `ent generate agents` command works
- [ ] Claude Code format generated correctly
- [ ] OpenCode format generated correctly
- [ ] Project type detection works
- [ ] Custom agent configuration supported
- [ ] Generated agents tested with both platforms
- [ ] Static agent files removed
- [ ] Documentation updated

## Effort Estimate

**~14 hours** across 11 tasks
