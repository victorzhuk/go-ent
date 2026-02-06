# Agent Routing Inefficiency Analysis & Design

**Date:** 2026-02-05
**Status:** Approved for Implementation
**Author:** AI-Assisted Analysis

## Executive Summary

The go-ent generated agents for Claude Code and opencode suffer from inefficient subagent routing. The parent AI almost always defaults to the `coder` agent regardless of task type, wasting cost (haiku-tier agents never used for simple tasks), quality (opus-tier agents skipped for complex work), and specialization (domain-specific agents like `tester`, `debugger`, `decomposer` rarely invoked).

**Root Cause:** Agent descriptions are the only routing signal, and they lack specificity. The `coder` description ("Implements features, writes code") matches every coding task. Additionally, valuable routing metadata (`complexityHints`, `modelMapping`) exists in agent YAML files but is never emitted to generated frontmatter, making it invisible to the parent AI.

**Solution:** Improve descriptions with routing guidance, add a new `whenToUse` field for explicit routing instructions, surface complexity hints in generated frontmatter, enhance the workflow prompt with a routing decision tree, and make handoff templates use real agent descriptions instead of hardcoded text.

---

## Part 1: Root Cause Analysis

### The Problem

When go-ent generates agents for Claude Code or opencode, the parent AI almost always defaults to the `coder` agent regardless of task type. This wastes:

- **Cost** — haiku-tier agents (`planner-fast`, `debugger-fast`) are never used for simple tasks
- **Quality** — opus-tier agents (`architect`, `reviewer`, `researcher`) are skipped for tasks that need deeper reasoning
- **Specialization** — tester, debugger, decomposer prompts are optimized for their domain but never invoked

### Root Cause 1: Description is the ONLY Routing Signal

Both Claude Code and opencode select subagents based on the `description` field in frontmatter. The coder's description is fatally broad:

```yaml
description: "Go developer. Implements features, writes code."
```

Every coding task is "implementing features" or "writing code." The parent AI sees this as the default catch-all.

**Evidence from templates:**
- `claude.yaml.tmpl` emits only: `name`, `description`, `model`, `skills`, `disallowedTools`, `color`, `role`, `complexity`, `dependencies`
- `opencode.yaml.tmpl` emits: `name`, `mode`, `description`, `model`, `tools`, `skills`, `tags`, `color`, `dependencies`
- Neither template emits `complexityHints`, `modelMapping`, or any "when to use" guidance

### Root Cause 2: No Negative Constraints in Descriptions

Current descriptions say WHAT agents do but never say WHEN to use them vs alternatives:

| Agent | Description | Missing Context |
|-------|-------------|-----------------|
| coder | "Implements features, writes code" | When NOT to use: planning, debugging, reviewing |
| debugger | "Systematic issue investigation" | Could be interpreted as what a coder does |
| tester | "Writes tests, TDD cycles" | Sounds like a sub-task of coding |
| researcher | "Research root causes, investigate code" | Overlaps heavily with debugger |
| decomposer | "Decomposes features into granular tasks" | Nearly identical to planner |
| acceptor | "Validate acceptance criteria" | Overlaps with reviewer |

### Root Cause 3: Complexity Routing Metadata is Invisible

The planner and debugger agents have excellent complexity routing data:

```yaml
# planner.yaml
complexityHints:
  simple: "Quick feasibility check and triage..."
  standard: "Multi-file changes, moderate scope..."
  complex: "Complex architectural planning..."
modelMapping:
  simple: haiku
  standard: sonnet
  complex: opus
```

But `complexityHints` and `modelMapping` are **not emitted** in `claude.yaml.tmpl` or `opencode.yaml.tmpl`. The parent AI never sees this guidance. These fields exist only in the meta YAML — they're dead metadata.

### Root Cause 4: The Shared Workflow Prompt is Too Generic

`_workflow.md` contains the only inter-agent routing guidance, and it's just 4 bullet points:

```markdown
- Planning → Execution: After creating task breakdown
- Execution → Testing: After implementing features
- Testing → Review: After running test suites
- Review → Coordination: After validation complete
```

No decision tree. No "if task is X, use agent Y." No routing table.

### Root Cause 5: Handoff Template is Hardcoded and Shallow

`_handoff.md.tmpl` generates static, uninformative handoff descriptions:

```go
{{ if eq . "tester" }}For test coverage
{{ else if eq . "debugger" }}If stuck on issue
{{ else if eq . "coder" }}For implementation
{{ else }}For collaboration{{ end }}
```

"For collaboration" is meaningless. "If stuck on issue" is too vague for debugging. This gives the parent AI no real guidance.

### Impact on Model Efficiency (opencode)

opencode model mapping from `ent.yaml`:
- `fast` → `zai-coding-plan/glm-4.5-air` (cheap, fast)
- `main` → `zai-coding-plan/glm-4.7` (balanced)
- `heavy` → `kimi-for-coding/k2p5` (expensive, powerful)

When everything routes to `coder` (main tier), you burn `glm-4.7` tokens on tasks that `glm-4.5-air` could handle (simple triage) and miss `k2p5` quality for tasks that need it (architecture decisions).

---

## Part 2: Design — Fix Agent Routing

### Approach: Improve Descriptions + Add Routing Context (No Code Logic Changes)

The fix is entirely in the **data layer** (descriptions, templates, prompts). No Go code changes needed for the core fix. The parent AI is the router — we just need to give it better routing signals.

### Change 1: Rewrite Agent Descriptions with Routing Guidance

Each description should follow the pattern: `[Role]. [What it does]. Use when [trigger]. Not for [exclusion].`

**Proposed new descriptions:**

| Agent | New Description |
|-------|----------------|
| `coder` | `Go developer. Implements features, writes code. Use ONLY for writing new code or modifying existing code after a plan exists. Not for planning, debugging, testing, or reviewing.` |
| `architect` | `System architect. Designs components, layers, data flow. Use for architecture decisions, API design, component boundaries, and system design before implementation.` |
| `planner` | `Task planner. Breaks features into actionable tasks. Supports complexity routing. Use for any task that needs decomposition before coding begins.` |
| `planner-fast` | `Quick task assessment and routing. Fast feasibility check and triage. Use for simple single-file changes with clear requirements.` |
| `planner-heavy` | `Deep architectural planning. Complex analysis for large-scale changes and architecture decisions. Use for multi-component changes affecting 5+ files.` |
| `debugger` | `Standard debugging. Systematic issue investigation and resolution. Supports complexity routing. Use when something is broken, failing, or behaving unexpectedly.` |
| `debugger-fast` | `Quick debugging for simple issues. Fast troubleshooting of obvious bugs. Use for clear error messages, typos, or single-line fixes.` |
| `debugger-heavy` | `Complex debugging. Concurrency issues, performance problems, multi-component failures. Use for race conditions, deadlocks, or failures spanning multiple packages.` |
| `tester` | `Test engineer. Writes tests, TDD cycles, creates reproductions. Use when the task is specifically about writing or fixing tests, not when implementing features that need tests (coder handles that).` |
| `reviewer` | `Code reviewer. Reviews code for bugs, security, quality, adherence to conventions. Use after implementation is complete to validate before merging. Not for fixing code.` |
| `researcher` | `Research root causes, investigate code, analyze bugs. Deep code analysis. Use for read-only investigation when you need to understand code before acting. Not for making changes.` |
| `decomposer` | `Task breakdown and dependency analysis. Decomposes features into granular tasks with ordering. Use when a feature needs to be split into ordered subtasks with dependencies.` |
| `acceptor` | `Validate acceptance criteria and requirements. Ensures implementation meets specifications. Use after implementation to verify requirements are met, distinct from code review.` |

### Change 2: Add `whenToUse` Field to Agent Meta Schema

Add a dedicated `whenToUse` field to the YAML schema. This separates **UI-facing description** (short, for display) from **routing guidance** (detailed, for the parent AI's decision-making).

**Files to modify:**
- `pkg/agents/meta/*.yaml` — add `whenToUse` text to each agent
- `internal/agent/registry.go` — add `WhenToUse string` to `AgentMeta` struct
- `internal/generator/types.go` — add `WhenToUse string` to `AgentMetaSource`
- `internal/generator/source.go` — propagate field from meta to source
- `pkg/agents/templates/claude.yaml.tmpl` — emit as extended description or section
- `pkg/agents/templates/opencode.yaml.tmpl` — emit in appropriate format

The `whenToUse` field provides explicit routing instructions the parent AI sees:

```yaml
# Example: coder.yaml
whenToUse: |
  Use when there is a clear plan and the task is to write or modify Go code.
  Do NOT use for: debugging (use debugger), planning (use planner),
  writing tests only (use tester), reviewing code (use reviewer),
  or investigating code (use researcher).
```

### Change 3: Surface complexityHints in Generated Frontmatter

Update `claude.yaml.tmpl` and `opencode.yaml.tmpl` to emit complexity hints when available.

**File:** `pkg/agents/templates/claude.yaml.tmpl`

Add after `complexity`:
```yaml
{{- if .ComplexityHints}}
complexityHints:{{range $level, $hint := .ComplexityHints}}
    {{$level}}: "{{$hint}}"{{end}}{{end}}
```

**File:** `internal/generator/types.go`

Add `ComplexityHints map[string]string` to `AgentMetaSource` (already exists in agent registry's `AgentMeta` struct, just needs to flow through to templates).

**File:** `internal/generator/source.go`

Propagate `ComplexityHints` from meta YAML into the template data struct via `ConvertMetaToSource`.

### Change 4: Enhance the Workflow Prompt with a Routing Decision Tree

**File:** `pkg/agents/prompts/shared/_workflow.md`

Add an explicit routing table after the existing delegation section:

```markdown
## Agent Selection Guide

**Before delegating, match the task type to the right agent:**

| Task Type | Agent | Model Tier |
|-----------|-------|------------|
| Write new code / modify code | @ent/coder | main |
| Design architecture / API boundaries | @ent/architect | heavy |
| Break feature into tasks | @ent/planner | main (auto-routes by complexity) |
| Quick triage of simple task | @ent/planner-fast | fast |
| Deep architectural planning | @ent/planner-heavy | heavy |
| Fix a bug / investigate failure | @ent/debugger | main (auto-routes by complexity) |
| Fix obvious single-line bug | @ent/debugger-fast | fast |
| Debug concurrency / perf issue | @ent/debugger-heavy | heavy |
| Write or fix tests specifically | @ent/tester | main |
| Review completed code | @ent/reviewer | heavy |
| Read-only code investigation | @ent/researcher | heavy |
| Break large task into subtasks | @ent/decomposer | main |
| Verify requirements are met | @ent/acceptor | main |

**Key rules:**
- Do NOT use @ent/coder for debugging — use @ent/debugger
- Do NOT use @ent/coder for planning — use @ent/planner
- Do NOT use @ent/coder for test-only tasks — use @ent/tester
- Use fast-tier agents for simple tasks to save cost and latency
- Use heavy-tier agents for architecture and review to get better reasoning
```

### Change 5: Improve Handoff Template with Real Descriptions

**File:** `pkg/agents/templates/sections/_handoff.md.tmpl`

The current template hardcodes descriptions. Instead, it should access the agent registry to pull real descriptions:

```go
{{- if .Dependencies }}
## Handoff

{{- range .DependencyDetails }}
- @ent/{{ .Name }} - {{ .Description }}
{{- end }}
{{- end }}
```

This requires:
1. Passing dependency agent metadata (not just names) to the template
2. Loading dependency agent details in the generator before template execution
3. The generator already has access to all agent meta via the registry, so this is a data-passing change, not a logic change

---

## Part 3: Implementation Plan

### Step 1: Write design document ✓
- **File:** `docs/plans/2026-02-05-agent-routing-design.md`
- **Content:** This document
- **Effort:** Small — write markdown

### Step 2: Add `whenToUse` field to Go structs and meta schema
- **Files:**
  - `internal/agent/registry.go` — add `WhenToUse string` to `AgentMeta`
  - `internal/generator/types.go` — add `WhenToUse string` to `AgentMetaSource`
  - `internal/generator/source.go` — propagate `WhenToUse` from meta → source via `LoadAgentMetaSource`
- **Effort:** Small — 3 one-line additions + propagation

### Step 3: Rewrite agent descriptions + add whenToUse (13 YAML files)
- **Files:** `pkg/agents/meta/*.yaml` (all 13 agents)
- **Changes per file:** Improve `description`, add `whenToUse` field
- **Effort:** Medium — text changes across 13 files

### Step 4: Surface complexityHints and whenToUse in templates
- **Files:**
  - `pkg/agents/templates/claude.yaml.tmpl` — emit `whenToUse` and `complexityHints`
  - `pkg/agents/templates/opencode.yaml.tmpl` — emit in opencode format
  - `internal/generator/source.go` — propagate ComplexityHints field to template data in `ConvertMetaToSource`
  - `internal/generator/claude.go` — pass ComplexityHints to template data
  - `internal/generator/opencode.go` — pass ComplexityHints to template data
- **Effort:** Small — template changes + data propagation

### Step 5: Enhance the workflow prompt with routing decision tree
- **File:** `pkg/agents/prompts/shared/_workflow.md`
- **Changes:** Add agent selection guide table, key routing rules
- **Effort:** Small — add markdown section

### Step 6: Improve handoff template with real descriptions
- **Files:**
  - `pkg/agents/templates/sections/_handoff.md.tmpl`
  - Generator code to pass dependency agent metadata to templates
- **Effort:** Medium — requires passing agent registry data to templates

### Step 7: Regenerate all agents and verify
- Run `make build && ent agent generate`
- Verify generated `.claude/agents/ent/*.md` and `.opencode/agents/ent/*.md`

---

## Verification Checklist

1. `make build` — project compiles
2. `make test` — existing tests pass
3. `ent agent generate` — regenerate agents for both platforms
4. Inspect generated `coder.md`, `debugger.md`, `planner.md` to verify:
   - New descriptions appear in frontmatter
   - `whenToUse` field is present (as extended description or comment)
   - `complexityHints` appear for planner/debugger
   - Handoff sections use real descriptions
5. Diff before/after of generated files to confirm all changes propagated

---

## Expected Impact

### Cost Efficiency
- Simple tasks (typos, single-line fixes) → `planner-fast` / `debugger-fast` (haiku tier)
- Standard tasks (multi-file changes) → `coder` / `debugger` (sonnet tier)
- Complex tasks (architecture, review) → `architect` / `reviewer` (opus tier)

### Quality Improvement
- Architecture decisions get deep reasoning from opus-tier `architect`
- Code reviews get rigorous analysis from opus-tier `reviewer`
- Debugging gets systematic investigation from specialized `debugger` prompts

### Specialization Utilization
- Test writing uses TDD-optimized `tester` prompt
- Task breakdown uses decomposition-focused `decomposer` or `planner`
- Read-only investigation uses research-focused `researcher` instead of conflating with implementation

---

## Future Enhancements (Not in This Change)

1. **Dynamic Routing Logic** — Add Go code to auto-select agents based on task keywords (would require changing how parent AI delegates)
2. **Routing Metrics** — Track which agents are actually invoked to measure routing effectiveness
3. **User-Configurable Routing** — Allow users to customize routing rules via config
4. **Multi-Agent Collaboration** — Enable agents to spawn peer agents for parallel work (would require agent-to-agent communication)

These are deferred to future work — this change focuses on fixing the **data** that informs routing decisions, not the routing logic itself.

---

## Appendix: Agent Descriptions Matrix

### Before (Current)

| Agent | Description |
|-------|-------------|
| coder | "Go developer. Implements features, writes code." |
| architect | "System architect. Designs components, layers, data flow." |
| planner | "Task planner. Breaks features into actionable tasks. Supports complexity routing." |
| planner-fast | "Quick task assessment and routing. Fast feasibility check and triage." |
| planner-heavy | "Deep architectural planning. Complex analysis for large-scale changes and architecture decisions." |
| debugger | "Standard debugging. Systematic issue investigation and resolution. Supports complexity routing." |
| debugger-fast | "Quick debugging for simple issues. Fast troubleshooting of obvious bugs." |
| debugger-heavy | "Complex debugging. Concurrency issues, performance problems, multi-component failures." |
| tester | "Test engineer. Writes tests, TDD cycles, creates reproductions." |
| reviewer | "Code reviewer. Reviews code for bugs, security, quality, adherence to conventions, and validates acceptance criteria." |
| researcher | "Research root causes, investigate code, analyze bugs. Deep code analysis." |
| decomposer | "Task breakdown and dependency analysis. Decomposes features into granular tasks." |
| acceptor | "Validate acceptance criteria and requirements. Ensures implementation meets specifications." |

### After (Improved)

| Agent | Description |
|-------|-------------|
| coder | "Go developer. Implements features, writes code. Use ONLY for writing new code or modifying existing code after a plan exists. Not for planning, debugging, testing, or reviewing." |
| architect | "System architect. Designs components, layers, data flow. Use for architecture decisions, API design, component boundaries, and system design before implementation." |
| planner | "Task planner. Breaks features into actionable tasks. Supports complexity routing. Use for any task that needs decomposition before coding begins." |
| planner-fast | "Quick task assessment and routing. Fast feasibility check and triage. Use for simple single-file changes with clear requirements." |
| planner-heavy | "Deep architectural planning. Complex analysis for large-scale changes and architecture decisions. Use for multi-component changes affecting 5+ files." |
| debugger | "Standard debugging. Systematic issue investigation and resolution. Supports complexity routing. Use when something is broken, failing, or behaving unexpectedly." |
| debugger-fast | "Quick debugging for simple issues. Fast troubleshooting of obvious bugs. Use for clear error messages, typos, or single-line fixes." |
| debugger-heavy | "Complex debugging. Concurrency issues, performance problems, multi-component failures. Use for race conditions, deadlocks, or failures spanning multiple packages." |
| tester | "Test engineer. Writes tests, TDD cycles, creates reproductions. Use when the task is specifically about writing or fixing tests, not when implementing features that need tests (coder handles that)." |
| reviewer | "Code reviewer. Reviews code for bugs, security, quality, adherence to conventions. Use after implementation is complete to validate before merging. Not for fixing code." |
| researcher | "Research root causes, investigate code, analyze bugs. Deep code analysis. Use for read-only investigation when you need to understand code before acting. Not for making changes." |
| decomposer | "Task breakdown and dependency analysis. Decomposes features into granular tasks with ordering. Use when a feature needs to be split into ordered subtasks with dependencies." |
| acceptor | "Validate acceptance criteria and requirements. Ensures implementation meets specifications. Use after implementation to verify requirements are met, distinct from code review." |
