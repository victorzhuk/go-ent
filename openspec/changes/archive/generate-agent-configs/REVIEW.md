# Generate Agent Configs - Review

**Date:** 2026-01-30
**Status:** Partially Implemented, Enhancements Deferred
**Proposal ID:** generate-agent-configs

---

## Executive Summary

The `generate-agent-configs` proposal describes a dynamic agent generation system. The core functionality described in the proposal **already exists** and works well. What's missing are enhancements for project-aware generation and OpenSpec integration.

**Verdict:** Archive proposal as enhancement. Current implementation meets core needs.

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

**Features:**
- ✅ Reads configuration from `ent.yaml`
- ✅ Generates for Claude Code format (`.claude/agents/ent/`)
- ✅ Generates for OpenCode format (`.opencode/agents/ent/`)
- ✅ Supports tool selection via flags
- ✅ Supports specific agent generation
- ✅ Clear error messages

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

**Features:**
- ✅ Meta format source files (YAML + Markdown)
- ✅ Multiple output format support
- ✅ Base prompt composition
- ✅ Skill and tool references
- ✅ Clear error handling

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
│   ├── ...
│   └── bases/          # Base prompts
│       └── shared.md
├── prompts/           # Agent prompt content
│   ├── agents/        # Individual agent prompts
│   │   ├── architect.md
│   │   ├── coder.md
│   │   ├── ...
│   └── shared/        # Shared prompt content
└── templates/         # Output format templates
```

**Features:**
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

**Features:**
- ✅ Configured tools
- ✅ Project-specific settings
- ✅ Clear structure

---

## What's Missing (Enhancements ❌)

### 1. Project Type Detection

**Proposal Claim:**
> "Project type detection (Go vs Web vs Mixed)"

**Current State:**
- ❌ No automatic project type detection
- ❌ All agents generated regardless of project type

**Proposed Enhancement:**
- Detect project type from files (Go, TypeScript, etc.)
- Generate only relevant agents based on project type

**Effort Estimate:** 4-6 hours

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

**Effort Estimate:** 6-8 hours

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

**Effort Estimate:** 4-6 hours

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

**Effort Estimate:** 2-4 hours

---

## Comparison Summary

| Feature | Exists | Missing | Effort |
|---------|--------|---------|--------|
| Command: `ent agent generate` | ✅ | - | - |
| Generator package | ✅ | - | - |
| Claude Code format | ✅ | - | - |
| OpenCode format | ✅ | - | - |
| Tool selection | ✅ | - | - |
| Project type detection | - | ❌ | 4-6h |
| OpenSpec integration | - | ❌ | 6-8h |
| Agent config section | - | ❌ | 4-6h |
| Project-specific selection | - | ❌ | 2-4h |

**Total Missing Effort:** ~16-24 hours

---

## Recommendation

### Archive Proposal as Enhancement

**Rationale:**

1. **Core Functionality Exists**
   - The main problem (dynamic agent generation) is solved
   - `ent agent generate` works well
   - Both Claude Code and OpenCode formats supported

2. **Current Workflow Works**
   - `ent generate` → generates skills and configs
   - `ent agent generate` → generates agents
   - Clear separation of concerns
   - No blocking issues reported

3. **Enhancements Are Nice-to-Have**
   - Project type detection is valuable but not critical
   - OpenSpec integration is interesting but undefined scope
   - Agent configuration adds complexity for limited benefit
   - Current all-agents approach is simple and reliable

4. **Better Use of Effort**
   - Focus on Phase 1-3 proposals (~36h)
   - Those have clearer business value
   - Cleanup and core improvements first
   - Enhancements can wait

5. **No Customer Demand**
   - No feature requests for project-aware generation
   - Current system meets user needs
   - Don't over-engineer without requirements

---

## Alternative Path

If project-aware generation becomes a requirement, implement as new proposal:

**New Proposal Structure:**
```
# enhance-agent-generation

## Summary
Add project-aware generation to existing agent system.

## Changes
1. Detect project type automatically
2. Filter agents based on project type
3. Add `.go-ent/config.yaml` support

## Effort
~16 hours
```

**Key Difference:**
- Acknowledges existing system
- Focuses on enhancements, not replacement
- Clearer scope and deliverables

---

## Questions & Answers

**Q: Why not complete the proposal?**
A: The core functionality exists. Enhancements add ~16-24 hours for moderate value. Better to focus on higher-impact work (Phase 1-3 proposals).

**Q: Will this be reconsidered later?**
A: Yes. If users request project-aware generation or OpenSpec-driven agents becomes important, create a focused enhancement proposal.

**Q: What if project type detection is easy to add?**
A: If it's truly simple (<4h total), could be done as a quick win. But given the scope (detection, filtering, config, testing), it's likely more work.

**Q: Does the current system have any problems?**
A: No known blocking issues. The system works as intended. All agents are generated, which is fine for most projects.

---

## Action Items

1. **Archive this proposal** as "enhancement - deferred"
2. **Create archive.md** documenting the current state
3. **Focus on Phase 1-3 proposals** (cleanup, simplify, streamline)
4. **Monitor user feedback** for demand for project-aware generation
5. **Revisit in Q2 2026** if demand emerges

---

## References

**Existing Implementation:**
- `internal/cli/agent/generate.go` - Command implementation
- `internal/generator/` - Generator package
- `pkg/agents/` - Agent sources (embedded)
- `ent.yaml` - Project configuration

**Original Proposal:**
- `openspec/changes/generate-agent-configs/proposal.md`

**Roadmap:**
- `openspec/ROADMAP.md` - Deferred section

---

**Review Date:** 2026-01-30
**Reviewer:** System Architecture Review
**Status:** Archive as Enhancement
