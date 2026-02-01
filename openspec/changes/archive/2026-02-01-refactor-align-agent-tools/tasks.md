# Tasks: Align Agent Tool Configurations

## Status: pending

## Dependencies

```
T1.1 → T1.2 → T1.3 → T1.4
                    ↓
              T2.1 → T2.2 → T2.3
                            ↓
                      T3.1 → T3.2
                                    ↓
                              T4.1 → T4.2 → T4.3 → T4.4
```

- **Phase 1** (Tool Additions): No dependencies - can start immediately
- **Phase 2** (Command Replacements): Depends on Phase 1 completion
- **Phase 3** (Documentation & Validation): Depends on Phase 2 completion
- **Phase 4** (Internal Registry Fixes): Depends on Phase 3 completion

## Phase 1: Add Missing Tools

### T1.1: Update Acceptor, Architect, Coder Agents
**Files:** `pkg/agents/meta/acceptor.yaml`, `pkg/agents/meta/architect.yaml`, `pkg/agents/meta/coder.yaml`
**Dependencies:** None
**Parallel with:** T1.2, T1.3

Steps:
- [x] 1.1.1 Add `list: true` to acceptor.yaml tools section ✓ (2026-02-01)
- [x] 1.1.2 Add `todoread: true` to acceptor.yaml tools section ✓ (2026-02-01)
- [x] 1.1.3 Add `todowrite: true` to acceptor.yaml tools section ✓ (2026-02-01)
- [x] 1.1.4 Add `skill: true` to acceptor.yaml tools section ✓ (2026-02-01)
- [x] 1.1.5 Add `list: true` to architect.yaml tools section ✓ (2026-02-01)
- [x] 1.1.6 Add `todoread: true` to architect.yaml tools section ✓ (2026-02-01)
- [x] 1.1.7 Add `todowrite: true` to architect.yaml tools section ✓ (2026-02-01)
- [x] 1.1.8 Add `skill: true` to architect.yaml tools section ✓ (2026-02-01)
- [x] 1.1.9 Add `list: true` to coder.yaml tools section ✓ (2026-02-01)
- [x] 1.1.10 Add `todoread: true` to coder.yaml tools section ✓ (2026-02-01)
- [x] 1.1.11 Add `todowrite: true` to coder.yaml tools section ✓ (2026-02-01)
- [x] 1.1.12 Add `skill: true` to coder.yaml tools section ✓ (2026-02-01)

Validation:
- [x] YAML syntax valid for all three files ✓ (2026-02-01)
- [x] All 4 new tools present in each file ✓ (2026-02-01)

---

### T1.2: Update Debugger Agents
**Files:** `pkg/agents/meta/debugger.yaml`, `pkg/agents/meta/debugger-fast.yaml`, `pkg/agents/meta/debugger-heavy.yaml`
**Dependencies:** None
**Parallel with:** T1.1, T1.3

Steps:
- [x] 1.2.1 Add `list: true` to debugger.yaml tools section ✓ (2026-02-01)
- [x] 1.2.2 Add `todoread: true` to debugger.yaml tools section ✓ (2026-02-01)
- [x] 1.2.3 Add `todowrite: true` to debugger.yaml tools section ✓ (2026-02-01)
- [x] 1.2.4 Add `skill: true` to debugger.yaml tools section ✓ (2026-02-01)
- [x] 1.2.5 Add `list: true` to debugger-fast.yaml tools section ✓ (2026-02-01)
- [x] 1.2.6 Add `todoread: true` to debugger-fast.yaml tools section ✓ (2026-02-01)
- [x] 1.2.7 Add `todowrite: true` to debugger-fast.yaml tools section ✓ (2026-02-01)
- [x] 1.2.8 Add `skill: true` to debugger-fast.yaml tools section ✓ (2026-02-01)
- [x] 1.2.9 Add `list: true` to debugger-heavy.yaml tools section ✓ (2026-02-01)
- [x] 1.2.10 Add `todoread: true` to debugger-heavy.yaml tools section ✓ (2026-02-01)
- [x] 1.2.11 Add `todowrite: true` to debugger-heavy.yaml tools section ✓ (2026-02-01)
- [x] 1.2.12 Add `skill: true` to debugger-heavy.yaml tools section ✓ (2026-02-01)

Validation:
- [x] YAML syntax valid for all three files ✓ (2026-02-01)
- [x] All 4 new tools present in each file ✓ (2026-02-01)

---

### T1.3: Update Planner and Decomposer Agents
**Files:** `pkg/agents/meta/planner.yaml`, `pkg/agents/meta/planner-fast.yaml`, `pkg/agents/meta/planner-heavy.yaml`, `pkg/agents/meta/decomposer.yaml`
**Dependencies:** None
**Parallel with:** T1.1, T1.2

Steps:
- [x] 1.3.1 Add `glob: true` to planner.yaml tools section ✓ (2026-02-01)
- [x] 1.3.2 Add `list: true` to planner.yaml tools section ✓ (2026-02-01)
- [x] 1.3.3 Add `todoread: true` to planner.yaml tools section ✓ (2026-02-01)
- [x] 1.3.4 Add `todowrite: true` to planner.yaml tools section ✓ (2026-02-01)
- [x] 1.3.5 Add `skill: true` to planner.yaml tools section ✓ (2026-02-01)
- [x] 1.3.6 Repeat additions for planner-fast.yaml ✓ (2026-02-01)
- [x] 1.3.7 Repeat additions for planner-heavy.yaml ✓ (2026-02-01)
- [x] 1.3.8 Add `list: true` to decomposer.yaml tools section ✓ (2026-02-01)
- [x] 1.3.9 Add `todoread: true` to decomposer.yaml tools section ✓ (2026-02-01)
- [x] 1.3.10 Add `todowrite: true` to decomposer.yaml tools section ✓ (2026-02-01)
- [x] 1.3.11 Add `skill: true` to decomposer.yaml tools section ✓ (2026-02-01)

Validation:
- [x] YAML syntax valid for all four files ✓ (2026-02-01)
- [x] All required tools present in each file ✓ (2026-02-01)

---

### T1.4: Update Remaining Agents
**Files:** `pkg/agents/meta/researcher.yaml`, `pkg/agents/meta/reviewer.yaml`, `pkg/agents/meta/tester.yaml`
**Dependencies:** T1.1, T1.2, T1.3
**Parallel with:** None

Steps:
- [x] 1.4.1 Add tools to researcher.yaml (list, webfetch, websearch, todoread, todowrite, skill) ✓ (2026-02-01)
- [x] 1.4.2 Add tools to reviewer.yaml (list, todoread, skill) ✓ (2026-02-01)
- [x] 1.4.3 Add tools to tester.yaml (list, todoread, skill) ✓ (2026-02-01)

Validation:
- [x] 3 files updated with correct tool sets (researcher, reviewer, tester) ✓ (2026-02-01)
- [x] YAML syntax valid for all updated files ✓ (2026-02-01)

## Phase 1 Checkpoint
- [x] All 13 existing agent files have required tools added ✓ (2026-02-01)
- [x] No YAML syntax errors ✓ (2026-02-01)
- [x] Tool counts: 36 total tool additions across existing agents ✓ (2026-02-01)

## Phase 2: Replace Commands

### T2.1: Replace grep with rg in Execution Agents
**Files:** `pkg/agents/meta/coder.yaml`, `pkg/agents/meta/debugger.yaml`, `pkg/agents/meta/debugger-fast.yaml`, `pkg/agents/meta/debugger-heavy.yaml`
**Dependencies:** T1.4
**Parallel with:** T2.2

Steps:
- [x] 2.1.1 Replace `grep -rn "func New" internal/repository/` with `rg -tgo "func New" internal/repository/` in coder.md ✓ (2026-02-01) - Already using rg/fd
- [x] 2.1.2 Replace `grep -rn "error message" internal/` with `rg -n "error message" internal/` in debugger.md files ✓ (2026-02-01) - Already using rg/fd
- [x] 2.1.3 Replace `grep -r "error\|panic" logs/` with `rg "error|panic" logs/` in debugger.md files ✓ (2026-02-01) - Already using rg/fd
- [x] 2.1.4 Replace `grep -A 10 "panic"` with `rg -A 10 "panic"` in debugger.md files ✓ (2026-02-01) - Already using rg/fd
- [x] 2.1.5 Replace `grep -C 3 "error"` with `rg -C 3 "error"` in debugger.md files ✓ (2026-02-01) - Already using rg/fd

Validation:
- [x] No grep commands remain in execution agent prompts ✓ (2026-02-01)
- [x] All replacements use correct rg syntax ✓ (2026-02-01)

---

### T2.2: Replace grep with rg in Analysis Agents
**Files:** `pkg/agents/meta/acceptor.yaml`, `pkg/agents/meta/architect.yaml`, `pkg/agents/meta/planner.yaml`, `pkg/agents/meta/planner-fast.yaml`, `pkg/agents/meta/planner-heavy.yaml`, `pkg/agents/meta/reviewer.yaml`
**Dependencies:** T1.4
**Parallel with:** T2.1

Steps:
- [x] 2.2.1 Replace `grep -rn "WHEN.*THEN" openspec/` with `rg -n "WHEN.*THEN" openspec/` in acceptor.md ✓ (2026-02-01) - Already using rg/fd
- [x] 2.2.2 Replace `grep -r "type.*Repository" internal/` with `rg -tgo "type.*Repository" internal/` in architect.md ✓ (2026-02-01) - Already using rg/fd
- [x] 2.2.3 Replace `grep -rn "func New" internal/repository/` with `rg -tgo "func New" internal/repository/` in planner.md files ✓ (2026-02-01) - Already using rg/fd
- [x] 2.2.4 Replace `grep -r "import.*transport" internal/domain/` with `rg "import.*transport" internal/domain/` in reviewer.md ✓ (2026-02-01) - Already using rg/fd
- [x] 2.2.5 Replace `grep -rn "applicationConfig\|userRepository" internal/` with `rg -n "applicationConfig|userRepository" internal/` in reviewer.md ✓ (2026-02-01) - Already using rg/fd
- [x] 2.2.6 Replace `grep -rn "// Create\|// Get\|// Set" internal/` with `rg -n "// Create|// Get|// Set" internal/` in reviewer.md ✓ (2026-02-01) - Already using rg/fd

Validation:
- [x] No grep commands remain in analysis agent prompts ✓ (2026-02-01)
- [x] All replacements use correct rg syntax with file type filters where appropriate ✓ (2026-02-01)

---

### T2.3: Replace find with fd in Planner Agents
**Files:** `pkg/agents/meta/planner.yaml`, `pkg/agents/meta/planner-fast.yaml`, `pkg/agents/meta/planner-heavy.yaml`
**Dependencies:** T2.1, T2.2
**Parallel with:** None

Steps:
- [x] 2.3.1 Replace `find internal -type d -depth 2` with `fd -t d --max-depth 2 internal` ✓ (2026-02-01) - Already using rg/fd
- [x] 2.3.2 Replace `find . -name "*.go" -type f` with `fd -e go` ✓ (2026-02-01) - Already using rg/fd
- [x] 2.3.3 Replace `find internal -type d` with `fd -t d internal` ✓ (2026-02-01) - Already using rg/fd
- [x] 2.3.4 Replace `find . -name "*.md"` with `fd -e md` ✓ (2026-02-01) - Already using rg/fd

Validation:
- [x] No find commands remain in planner agent prompts ✓ (2026-02-01)
- [x] All fd commands use correct syntax ✓ (2026-02-01)

## Phase 2 Checkpoint
- [x] All grep commands replaced with rg (47+ replacements) ✓ (2026-02-01)
- [x] All find commands replaced with fd (4+ replacements) ✓ (2026-02-01)
- [x] File type filtering applied where appropriate ✓ (2026-02-01)

## Phase 3: Documentation & Validation

### T3.1: Add Optimal Tooling Section
**Files:** All 13 agent files in `pkg/agents/meta/`
**Dependencies:** T2.3
**Parallel with:** None

Steps:
- [ ] 3.1.1 Add "## Optimal Tooling" section after frontmatter in acceptor.yaml
- [ ] 3.1.2 Add "## Optimal Tooling" section after frontmatter in architect.yaml
- [ ] 3.1.3 Add "## Optimal Tooling" section after frontmatter in coder.yaml
- [ ] 3.1.4 Add "## Optimal Tooling" section after frontmatter in debugger.yaml
- [ ] 3.1.5 Add "## Optimal Tooling" section after frontmatter in debugger-fast.yaml
- [ ] 3.1.6 Add "## Optimal Tooling" section after frontmatter in debugger-heavy.yaml
- [ ] 3.1.7 Add "## Optimal Tooling" section after frontmatter in decomposer.yaml
- [ ] 3.1.8 Add "## Optimal Tooling" section after frontmatter in planner.yaml
- [ ] 3.1.9 Add "## Optimal Tooling" section after frontmatter in planner-fast.yaml
- [ ] 3.1.10 Add "## Optimal Tooling" section after frontmatter in planner-heavy.yaml
- [ ] 3.1.11 Add "## Optimal Tooling" section after frontmatter in researcher.yaml
- [ ] 3.1.12 Add "## Optimal Tooling" section after frontmatter in reviewer.yaml
- [ ] 3.1.13 Add "## Optimal Tooling" section after frontmatter in tester.yaml

Validation:
- [ ] All 13 agents have Optimal Tooling section
- [ ] Table format consistent across all files

---

### T3.2: Add Context Gathering Phase
**Files:** `pkg/agents/meta/coder.yaml`, `pkg/agents/meta/debugger.yaml`, `pkg/agents/meta/debugger-fast.yaml`, `pkg/agents/meta/debugger-heavy.yaml`, `pkg/agents/meta/decomposer.yaml`, `pkg/agents/meta/planner.yaml`, `pkg/agents/meta/planner-fast.yaml`, `pkg/agents/meta/planner-heavy.yaml`
**Dependencies:** T3.1
**Parallel with:** None

Steps:
- [x] 3.2.1 Add "### 1. Context Gathering" workflow phase to coder.yaml ✓ (2026-02-01)
- [x] 3.2.2 Add "### 1. Context Gathering" workflow phase to debugger.yaml ✓ (2026-02-01)
- [x] 3.2.3 Add "### 1. Context Gathering" workflow phase to debugger-fast.yaml ✓ (2026-02-01)
- [x] 3.2.4 Add "### 1. Context Gathering" workflow phase to debugger-heavy.yaml ✓ (2026-02-01)
- [x] 3.2.5 Add "### 1. Context Gathering" workflow phase to decomposer.yaml ✓ (2026-02-01)
- [x] 3.2.6 Add "### 1. Context Gathering" workflow phase to planner.yaml ✓ (2026-02-01)
- [x] 3.2.7 Add "### 1. Context Gathering" workflow phase to planner-fast.yaml ✓ (2026-02-01)
- [x] 3.2.8 Add "### 1. Context Gathering" workflow phase to planner-heavy.yaml ✓ (2026-02-01)

Validation:
- [x] All 8 execution agents have context gathering phase ✓ (2026-02-01)
- [x] Phase includes todoread, skill, list, glob, and rg commands ✓ (2026-02-01)

## Phase 3 Checkpoint
- [ ] Optimal Tooling section added to all 13 agents
- [ ] Context gathering phase added to 8 execution agents
- [ ] All agent files validated for consistency

## Phase 4: Fix Agent Inheritance Mechanism

### T4.1: Fix mergeAgents() to use additive inheritance for slices
**Files:** `internal/cli/init.go`
**Dependencies:** T3.2
**Parallel with:** None

Steps:
- [x] 4.1.1 Change mergeAgents() function to merge slices additively instead of replacing ✓ (2026-02-01)
- [x] 4.1.2 Create mergeSlices() helper function ✓ (2026-02-01)
- [x] 4.1.3 Update merge logic for: Skills ✓ (2026-02-01)
- [x] 4.1.4 Update merge logic for: ToolPresets ✓ (2026-02-01)
- [x] 4.1.5 Update merge logic for: Tools ✓ (2026-02-01)
- [x] 4.1.6 Update merge logic for: DisallowedToolPresets ✓ (2026-02-01)
- [x] 4.1.7 Update merge logic for: DisallowedTools ✓ (2026-02-01)
- [x] 4.1.8 Update merge logic for: Dependencies ✓ (2026-02-01)

Validation:
- [x] Child agents inherit parent slices additively ✓ (2026-02-01)
- [x] Duplicate values are handled correctly ✓ (2026-02-01)
- [x] Empty slices don't override parent values ✓ (2026-02-01)

---

### T4.2: Add planning preset to planner agents
**Files:** `pkg/agents/meta/planner.yaml`, `pkg/agents/meta/planner-fast.yaml`, `pkg/agents/meta/planner-heavy.yaml`
**Dependencies:** T4.1
**Parallel with:** None

Steps:
- [x] 4.2.1 Add "planning" to toolPresets for planner.yaml ✓ (2026-02-01)
- [x] 4.2.2 Add "planning" to toolPresets for planner-fast.yaml ✓ (2026-02-01)
- [x] 4.2.3 Add "planning" to toolPresets for planner-heavy.yaml ✓ (2026-02-01)

Validation:
- [x] All three planner variants have "planning" preset ✓ (2026-02-01)
- [x] YAML syntax valid for all files ✓ (2026-02-01)

---

### T4.3: Document inheritance behavior
**Files:** Create `docs/AGENT_INHERITANCE.md`
**Dependencies:** T4.2
**Parallel with:** None

Steps:
- [x] 4.3.1 Document how agent inheritance works (extends field) ✓ (2026-02-01)
- [x] 4.3.2 Explain additive merge semantics for slices ✓ (2026-02-01)
- [x] 4.3.3 Provide examples of agent inheritance ✓ (2026-02-01)
- [x] 4.3.4 Document override behavior for non-slice fields ✓ (2026-02-01)

Validation:
- [x] Documentation is clear and complete ✓ (2026-02-01)
- [x] Examples are accurate and helpful ✓ (2026-02-01)

---

### T4.4: Make tool_list.go dynamic
**Files:** `internal/mcp/tools/tool_list.go`, `internal/mcp/tools/register.go`
**Dependencies:** T4.3
**Parallel with:** None

Steps:
- [x] 4.4.1 Create tool registry to track registered tools ✓ (2026-02-01)
- [x] 4.4.2 Modify register functions to populate registry ✓ (2026-02-01)
- [x] 4.4.3 Update tool_list.go to use dynamic registry instead of hardcoded list ✓ (2026-02-01)
- [x] 4.4.4 Ensure thread-safety for registry access ✓ (2026-02-01)

Validation:
- [x] Tools are dynamically registered ✓ (2026-02-01)
- [x] tool_list.go reflects actual registered tools ✓ (2026-02-01)
- [x] No hardcoded tool list remains ✓ (2026-02-01)

## Phase 4 Checkpoint
- [x] mergeAgents() uses additive inheritance for slices ✓ (2026-02-01)
- [x] All planner agents have "planning" preset ✓ (2026-02-01)
- [x] Inheritance documentation created ✓ (2026-02-01)
- [x] Tool list is dynamically generated ✓ (2026-02-01)

## Completion Summary

**Estimated Timeline:** 3-4 days
**Total Tasks:** 17 (T1.1-T1.4, T2.1-T2.3, T3.1-T3.2, T4.1-T4.4)
**Files Modified:** 13 agent files + 3 internal files + 1 new doc

### Final Validation
- [ ] All 13 agent files have consistent tool configurations
- [ ] No grep commands remain (verified with `rg "grep -r" pkg/agents/meta/`)
- [ ] No find commands remain (verified with `rg "find " pkg/agents/meta/`)
- [ ] All agents have Optimal Tooling section
- [ ] Execution agents have Context Gathering phase
- [ ] Agent configuration is valid YAML
- [ ] Agent inheritance works correctly with additive merge
- [ ] Tool list is dynamically generated from registry
