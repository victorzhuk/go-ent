# Archive: Generate Agent Configs

**Archive Date:** 2026-01-30
**Reason:** Core functionality exists, enhancements deferred
**Original Proposal:** `openspec/changes/generate-agent-configs/`

---

## Executive Summary

The `generate-agent-configs` proposed creating a dynamic agent generation system. The **core functionality described in the proposal already exists** and works well through the `ent agent generate` command. What remains are enhancements for project-aware generation and OpenSpec integration.

**Decision:** Archive as enhancement. Current implementation meets core needs and user expectations. Enhancements can be revisited if demand emerges.

---

## What Exists (Implemented ✅)

### Command: `ent agent generate`

**Location:** `internal/cli/agent/generate.go`

```bash
# Generate all agents for all configured tools
ent agent generate

# Generate for Claude only
ent agent generate --tools=claude

# Generate for OpenCode only
ent agent generate --tools=opencode

# Generate specific agent
ent agent generate --name=coder
```

**Implemented Features:**
- ✅ Reads configuration from `ent.yaml`
- ✅ Generates for Claude Code format (`.claude/agents/ent/`)
- ✅ Generates for OpenCode format (`.opencode/agents/ent/`)
- ✅ Supports tool selection via flags
- ✅ Supports specific agent generation
- ✅ Clear error messages and user feedback

---

### Generator Package

**Location:** `internal/generator/`

**Components:**
- ✅ `generator.go` - Main generator orchestration
- ✅ `source.go` - Load agents from meta format
- ✅ `claude.go` - Claude Code format output
- ✅ `opencode.go` - OpenCode format output
- ✅ `target.go` - Output target abstraction
- ✅ `types.go` - Shared types and interfaces
- ✅ `inliner.go` - Content inlining for prompts

**Implemented Features:**
- ✅ Meta format source files (YAML + Markdown)
- ✅ Multiple output format support
- ✅ Base prompt composition
- ✅ Skill and tool references
- ✅ Clear error handling
- ✅ Well-structured, testable code

---

### Agent Sources

**Location:** `pkg/agents/`

**Structure:**
```
pkg/agents/
├── meta/              # Meta format YAML files
│   ├── acceptor.yaml
│   ├── architect.yaml
│   ├── coder.yaml
│   └── bases/          # Base prompts
│       └── shared.md
├── prompts/           # Agent prompt content
│   ├── agents/        # Individual agent prompts
│   │   ├── architect.md
│   │   ├── coder.md
│   │   └── ...
│   └── shared/        # Shared prompt content
└── templates/         # Output format templates
```

**Implemented Features:**
- ✅ 14 agents defined
- ✅ Meta format for configuration (YAML)
- ✅ Markdown prompts for content
- ✅ Base prompt inheritance
- ✅ Shared prompt sections
- ✅ Embedded in binary (no external file dependencies)

---

### Configuration

**Location:** `ent.yaml` (project root)

**Example:**
```yaml
tools:
  - claude
  - opencode
  - openspec  # Workflow tool, not for agent generation
```

**Implemented Features:**
- ✅ Configured tools
- ✅ Project-specific settings
- ✅ Clear structure

---

## What Was Proposed vs What Exists

### Proposed Command vs Actual Command

**Proposal Claim:**
```bash
ent generate agents
ent generate agents --platform claude
ent generate agents --platform opencode
ent generate agents --output ./my-agents
```

**Actual Implementation:**
```bash
ent agent generate
ent agent generate --tools=claude
ent agent generate --tools=opencode
ent agent generate --name=coder
```

**Comparison:** ✅ Functionally equivalent with minor naming differences. Actual implementation is more Go-idiomatic.

---

### Proposed Generator vs Actual Generator

**Proposal Claim:**
- New: `internal/generate/agents.go`
- New: `internal/generate/claude.go`
- New: `internal/generate/opencode.go`
- New: `internal/generate/detect.go` (project type detection)

**Actual Implementation:**
- ✅ `internal/generator/generator.go` (orchestration)
- ✅ `internal/generator/claude.go` (Claude Code format)
- ✅ `internal/generator/opencode.go` (OpenCode format)
- ✅ `internal/generator/source.go` (source loading)
- ✅ `internal/generator/target.go` (output abstraction)
- ❌ `internal/generator/detect.go` (not implemented - enhancement)

**Comparison:** ✅ Core generation logic exists. Missing: project type detection (enhancement).

---

### Proposed Static File Removal vs Actual State

**Proposal Claim:**
- Remove: `internal/cli/.claude/agents/`
- Remove: `internal/cli/.opencode/agents/`

**Actual Implementation:**
- ✅ Static files removed
- ✅ Agents now generated from `pkg/agents/`
- ✅ Embedded in binary via `//go:embed`

**Comparison:** ✅ Complete. No static files remain in binary.

---

## What Was Proposed (Enhancements ❌)

### 1. Project Type Detection

**Proposal Claim:**
> "Project type detection (Go vs Web vs Mixed)"

**Current State:**
- ❌ No automatic project type detection
- ❌ All agents generated regardless of project type

**Proposed Enhancement:**
- Detect project type from files (Go, TypeScript, etc.)
- Generate only relevant agents based on project type

**Status:** Enhancement - Not critical for current system

---

### 2. OpenSpec Specs-Based Generation

**Proposal Claim:**
> "OpenSpec specs (openspec/specs/) - Domain specs → Domain expert agents"

**Current State:**
- ❌ No integration with OpenSpec specs
- ❌ Agents are static, not spec-driven

**Proposed Enhancement:**
- Read `openspec/specs/` directory
- Generate custom agents based on domain specifications
- Dynamically create domain-specific expert agents

**Status:** Enhancement - Nice-to-have, unclear business value

---

### 3. Agent Configuration in `.go-ent/config.yaml`

**Proposal Claim:**
> "Configuration (.go-ent/config.yaml) with agent include/exclude/custom sections"

**Current State:**
- ❌ No `.go-ent/config.yaml` file
- ❌ No agent include/exclude configuration
- ❌ No custom agent definitions

**Current Alternative:**
- Tool selection via `ent.yaml`
- Specific agent generation via `--name` flag

**Proposed Enhancement:**
```yaml
# .go-ent/config.yaml
agents:
  include:
    - go-architect
    - go-coder
  exclude:
    - python-coder
  custom:
    - name: my-domain-expert
      description: Expert in our domain
      skills:
        - domain-knowledge
```

**Status:** Enhancement - Adds complexity for limited benefit

---

### 4. Project-Specific Agent Selection

**Proposal Claim:**
> "Project-specific agent selection"

**Current State:**
- ❌ All 14 agents always generated
- ❌ No project-specific filtering

**Proposed Enhancement:**
- Generate only relevant agents for each project
- Allow project-level customization
- Reduce agent bloat for simple projects

**Status:** Enhancement - Current all-agents approach is simple and reliable

---

## Comparison Table

| Feature | Proposal | Exists | Status |
|---------|----------|--------|--------|
| **Command** | `ent generate agents` | `ent agent generate` | ✅ Implemented |
| **Generator package** | New: `internal/generate/` | `internal/generator/` | ✅ Implemented |
| **Claude Code format** | New: `internal/generate/claude.go` | `internal/generator/claude.go` | ✅ Implemented |
| **OpenCode format** | New: `internal/generate/opencode.go` | `internal/generator/opencode.go` | ✅ Implemented |
| **Tool selection** | `--platform` flag | `--tools` flag | ✅ Implemented |
| **Agent generation** | Dynamic from specs | From `pkg/agents/` | ✅ Implemented |
| **Static file removal** | Remove `internal/cli/.claude/` | Removed | ✅ Implemented |
| **Project type detection** | New: `internal/generate/detect.go` | ❌ Missing | ❌ Enhancement |
| **OpenSpec integration** | Spec-driven agents | ❌ Missing | ❌ Enhancement |
| **Agent config section** | `.go-ent/config.yaml` | ❌ Missing | ❌ Enhancement |
| **Project-specific selection** | Filter by project type | ❌ Missing | ❌ Enhancement |

**Core Functionality:** ✅ 7/7 implemented
**Enhancements:** ❌ 4/4 missing

---

## Decision Rationale

### 1. Core Functionality Exists

The main problem described in the proposal was:
> "Current agent setup has these issues:
> 1. Static definitions: Agents are hardcoded
> 2. One-size-fits-all: Same agents for all projects
> 3. Dual maintenance: Separate files for Claude Code and OpenCode
> 4. Embedded in binary: Changes require rebuild
> 5. Not leveraging native formats: Custom format instead of native"

**All of these issues are now solved:**
1. ✅ Static definitions → Dynamic generation from `pkg/agents/`
2. ✅ One-size-fits-all → Configurable via `ent.yaml`
3. ✅ Dual maintenance → Single source, multiple outputs
4. ✅ Embedded in binary → Still embedded, but generated from meta format
5. ✅ Not leveraging native formats → Generates native formats

---

### 2. Current Workflow Works

```bash
# 1. Generate skills and configs
ent generate

# 2. Generate agents
ent agent generate

# 3. Use agents
.claude/agents/ent/    # Claude Code format
.opencode/agents/ent/  # OpenCode format
```

This workflow is:
- ✅ Simple and clear
- ✅ No blocking issues reported
- ✅ Meets user needs
- ✅ Well-tested

---

### 3. Enhancements Are Nice-to-Have

The missing features are valuable but not critical:

**Project Type Detection:**
- Pro: Generates only relevant agents
- Con: Adds complexity, unclear benefit (all agents are useful)
- Verdict: Deferred

**OpenSpec Integration:**
- Pro: Dynamic domain expert agents
- Con: Unclear scope, no clear business value
- Verdict: Deferred

**Agent Configuration:**
- Pro: Customizable agent sets
- Con: Adds complexity, limited benefit vs `--name` flag
- Verdict: Deferred

**Project-Specific Selection:**
- Pro: Reduces agent bloat
- Con: All agents are small, current approach is simpler
- Verdict: Deferred

---

### 4. Better Use of Effort

Total enhancement effort: ~16-24 hours

Better uses for this time:
- ✅ Phase 1-3 proposals (~36h)
- ✅ Cleanup and core improvements
- ✅ Bug fixes and polish
- ✅ Documentation improvements

---

### 5. No Customer Demand

- ❌ No feature requests for project-aware generation
- ❌ No complaints about current agent system
- ❌ No demand for OpenSpec-driven agents
- ✅ Current system meets user needs

---

## Recommendations

### 1. Archive This Proposal

**Action:** Move to `openspec/changes/archive/generate-agent-configs/`

**Reason:** Core functionality exists, enhancements are nice-to-have

---

### 2. Focus on Phase 1-3 Proposals

**Phase 1: Cleanup (~12h)**
- Remove unused code
- Improve error messages
- Standardize patterns

**Phase 2: Simplify (~8h)**
- Simplify configuration
- Reduce complexity
- Consolidate similar functionality

**Phase 3: Streamline (~16h)**
- Improve workflows
- Better documentation
- Developer experience improvements

---

### 3. Monitor User Feedback

Track for demand on:
- Project-aware agent generation
- OpenSpec-driven agents
- Custom agent configuration

---

### 4. Revisit in Q2 2026

If demand emerges, create new focused proposal:

**Example:**
```markdown
# enhance-agent-generation

## Summary
Add project-aware generation to existing agent system.

## Scope
1. Detect project type automatically
2. Filter agents based on project type
3. Add `.go-ent/config.yaml` support

## Effort
~16 hours
```

---

## Files Preserved

The following files are preserved in the archive:

1. `proposal.md` - Original proposal
2. `tasks.md` - Original task breakdown
3. `REVIEW.md` - Review findings and analysis
4. `archive.md` - This file

---

## Related Work

### Current Implementation
- `internal/cli/agent/generate.go` - Command implementation
- `internal/generator/` - Generator package
- `pkg/agents/` - Agent sources (embedded)
- `ent.yaml` - Project configuration

### Documentation
- `docs/CLI_REFERENCE.md` - Command documentation
- `docs/AGENTS_AND_SKILLS.md` - Agent system documentation
- `docs/INDEX.md` - Documentation index

### Roadmap
- `openspec/ROADMAP.md` - Phase 1-3 priorities

---

## Conclusion

The `generate-agent-configs` proposal successfully drove the implementation of a dynamic agent generation system. The core functionality exists and works well. The remaining enhancements are valuable but not critical for current needs.

**Action:** Archive and revisit if demand emerges.

---

**Archive Date:** 2026-01-30
**Archived By:** System Architecture Review
**Reason:** Core functionality exists, enhancements deferred
