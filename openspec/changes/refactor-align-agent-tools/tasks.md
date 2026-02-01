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
**Files:** `.claude/agents/acceptor.md`, `.claude/agents/architect.md`, `.claude/agents/coder.md`
**Dependencies:** None
**Parallel with:** T1.2, T1.3

Steps:
- [ ] 1.1.1 Add `list: true` to acceptor.md tools section
- [ ] 1.1.2 Add `todoread: true` to acceptor.md tools section
- [ ] 1.1.3 Add `todowrite: true` to acceptor.md tools section
- [ ] 1.1.4 Add `skill: true` to acceptor.md tools section
- [ ] 1.1.5 Add `list: true` to architect.md tools section
- [ ] 1.1.6 Add `todoread: true` to architect.md tools section
- [ ] 1.1.7 Add `todowrite: true` to architect.md tools section
- [ ] 1.1.8 Add `skill: true` to architect.md tools section
- [ ] 1.1.9 Add `list: true` to coder.md tools section
- [ ] 1.1.10 Add `todoread: true` to coder.md tools section
- [ ] 1.1.11 Add `todowrite: true` to coder.md tools section
- [ ] 1.1.12 Add `skill: true` to coder.md tools section

Validation:
- [ ] YAML syntax valid for all three files
- [ ] All 4 new tools present in each file

---

### T1.2: Update Debugger Agents
**Files:** `.claude/agents/debugger.md`, `.claude/agents/debugger-fast.md`, `.claude/agents/debugger-heavy.md`
**Dependencies:** None
**Parallel with:** T1.1, T1.3

Steps:
- [ ] 1.2.1 Add `list: true` to debugger.md tools section
- [ ] 1.2.2 Add `todoread: true` to debugger.md tools section
- [ ] 1.2.3 Add `todowrite: true` to debugger.md tools section
- [ ] 1.2.4 Add `skill: true` to debugger.md tools section
- [ ] 1.2.5 Add `list: true` to debugger-fast.md tools section
- [ ] 1.2.6 Add `todoread: true` to debugger-fast.md tools section
- [ ] 1.2.7 Add `todowrite: true` to debugger-fast.md tools section
- [ ] 1.2.8 Add `skill: true` to debugger-fast.md tools section
- [ ] 1.2.9 Add `list: true` to debugger-heavy.md tools section
- [ ] 1.2.10 Add `todoread: true` to debugger-heavy.md tools section
- [ ] 1.2.11 Add `todowrite: true` to debugger-heavy.md tools section
- [ ] 1.2.12 Add `skill: true` to debugger-heavy.md tools section

Validation:
- [ ] YAML syntax valid for all three files
- [ ] All 4 new tools present in each file

---

### T1.3: Update Planner and Decomposer Agents
**Files:** `.claude/agents/planner.md`, `.claude/agents/planner-fast.md`, `.claude/agents/planner-heavy.md`, `.claude/agents/decomposer.md`
**Dependencies:** None
**Parallel with:** T1.1, T1.2

Steps:
- [ ] 1.3.1 Add `glob: true` to planner.md tools section
- [ ] 1.3.2 Add `list: true` to planner.md tools section
- [ ] 1.3.3 Add `todoread: true` to planner.md tools section
- [ ] 1.3.4 Add `todowrite: true` to planner.md tools section
- [ ] 1.3.5 Add `skill: true` to planner.md tools section
- [ ] 1.3.6 Repeat additions for planner-fast.md
- [ ] 1.3.7 Repeat additions for planner-heavy.md
- [ ] 1.3.8 Add `list: true` to decomposer.md tools section
- [ ] 1.3.9 Add `todoread: true` to decomposer.md tools section
- [ ] 1.3.10 Add `todowrite: true` to decomposer.md tools section
- [ ] 1.3.11 Add `skill: true` to decomposer.md tools section

Validation:
- [ ] YAML syntax valid for all four files
- [ ] All required tools present in each file

---

### T1.4: Update Remaining Agents
**Files:** `.claude/agents/reproducer.md`, `.claude/agents/researcher.md`, `.claude/agents/reviewer.md`, `.claude/agents/task-fast.md`, `.claude/agents/task-heavy.md`, `.claude/agents/tester.md`
**Dependencies:** T1.1, T1.2, T1.3
**Parallel with:** None

Steps:
- [ ] 1.4.1 Add tools to reproducer.md (list, todoread, todowrite, skill)
- [ ] 1.4.2 Add tools to researcher.md (list, webfetch, websearch, todoread, todowrite, skill)
- [ ] 1.4.3 Add tools to reviewer.md (list, todoread, skill)
- [ ] 1.4.4 Add tools to task-fast.md (list, todoread, skill)
- [ ] 1.4.5 Add tools to task-heavy.md (list, todoread, todowrite, skill)
- [ ] 1.4.6 Add tools to tester.md (list, todoread, skill)

Validation:
- [ ] All 6 files updated with correct tool sets
- [ ] YAML syntax valid for all files

## Phase 1 Checkpoint
- [ ] All 14 agent files have required tools added
- [ ] No YAML syntax errors
- [ ] Tool counts: 48 total tool additions across all agents

## Phase 2: Replace Commands

### T2.1: Replace grep with rg in Execution Agents
**Files:** `.claude/agents/coder.md`, `.claude/agents/debugger.md`, `.claude/agents/debugger-fast.md`, `.claude/agents/debugger-heavy.md`
**Dependencies:** T1.4
**Parallel with:** T2.2

Steps:
- [ ] 2.1.1 Replace `grep -rn "func New" internal/repository/` with `rg -tgo "func New" internal/repository/` in coder.md
- [ ] 2.1.2 Replace `grep -rn "error message" internal/` with `rg -n "error message" internal/` in debugger.md files
- [ ] 2.1.3 Replace `grep -r "error\|panic" logs/` with `rg "error|panic" logs/` in debugger.md files
- [ ] 2.1.4 Replace `grep -A 10 "panic"` with `rg -A 10 "panic"` in debugger.md files
- [ ] 2.1.5 Replace `grep -C 3 "error"` with `rg -C 3 "error"` in debugger.md files

Validation:
- [ ] No grep commands remain in execution agent prompts
- [ ] All replacements use correct rg syntax

---

### T2.2: Replace grep with rg in Analysis Agents
**Files:** `.claude/agents/acceptor.md`, `.claude/agents/architect.md`, `.claude/agents/planner.md`, `.claude/agents/planner-fast.md`, `.claude/agents/planner-heavy.md`, `.claude/agents/reviewer.md`
**Dependencies:** T1.4
**Parallel with:** T2.1

Steps:
- [ ] 2.2.1 Replace `grep -rn "WHEN.*THEN" openspec/` with `rg -n "WHEN.*THEN" openspec/` in acceptor.md
- [ ] 2.2.2 Replace `grep -r "type.*Repository" internal/` with `rg -tgo "type.*Repository" internal/` in architect.md
- [ ] 2.2.3 Replace `grep -rn "func New" internal/repository/` with `rg -tgo "func New" internal/repository/` in planner.md files
- [ ] 2.2.4 Replace `grep -r "import.*transport" internal/domain/` with `rg "import.*transport" internal/domain/` in reviewer.md
- [ ] 2.2.5 Replace `grep -rn "applicationConfig\|userRepository" internal/` with `rg -n "applicationConfig|userRepository" internal/` in reviewer.md
- [ ] 2.2.6 Replace `grep -rn "// Create\|// Get\|// Set" internal/` with `rg -n "// Create|// Get|// Set" internal/` in reviewer.md

Validation:
- [ ] No grep commands remain in analysis agent prompts
- [ ] All replacements use correct rg syntax with file type filters where appropriate

---

### T2.3: Replace find with fd in Planner Agents
**Files:** `.claude/agents/planner.md`, `.claude/agents/planner-fast.md`, `.claude/agents/planner-heavy.md`
**Dependencies:** T2.1, T2.2
**Parallel with:** None

Steps:
- [ ] 2.3.1 Replace `find internal -type d -depth 2` with `fd -t d --max-depth 2 internal`
- [ ] 2.3.2 Replace `find . -name "*.go" -type f` with `fd -e go`
- [ ] 2.3.3 Replace `find internal -type d` with `fd -t d internal`
- [ ] 2.3.4 Replace `find . -name "*.md"` with `fd -e md`

Validation:
- [ ] No find commands remain in planner agent prompts
- [ ] All fd commands use correct syntax

## Phase 2 Checkpoint
- [ ] All grep commands replaced with rg (47+ replacements)
- [ ] All find commands replaced with fd (4+ replacements)
- [ ] File type filtering applied where appropriate

## Phase 3: Documentation & Validation

### T3.1: Add Optimal Tooling Section
**Files:** All 14 agent files in `.claude/agents/`
**Dependencies:** T2.3
**Parallel with:** None

Steps:
- [ ] 3.1.1 Add "## Optimal Tooling" section after frontmatter in acceptor.md
- [ ] 3.1.2 Add "## Optimal Tooling" section after frontmatter in architect.md
- [ ] 3.1.3 Add "## Optimal Tooling" section after frontmatter in coder.md
- [ ] 3.1.4 Add "## Optimal Tooling" section after frontmatter in debugger.md
- [ ] 3.1.5 Add "## Optimal Tooling" section after frontmatter in debugger-fast.md
- [ ] 3.1.6 Add "## Optimal Tooling" section after frontmatter in debugger-heavy.md
- [ ] 3.1.7 Add "## Optimal Tooling" section after frontmatter in decomposer.md
- [ ] 3.1.8 Add "## Optimal Tooling" section after frontmatter in planner.md
- [ ] 3.1.9 Add "## Optimal Tooling" section after frontmatter in planner-fast.md
- [ ] 3.1.10 Add "## Optimal Tooling" section after frontmatter in planner-heavy.md
- [ ] 3.1.11 Add "## Optimal Tooling" section after frontmatter in reproducer.md
- [ ] 3.1.12 Add "## Optimal Tooling" section after frontmatter in researcher.md
- [ ] 3.1.13 Add "## Optimal Tooling" section after frontmatter in reviewer.md
- [ ] 3.1.14 Add "## Optimal Tooling" section after frontmatter in task-fast.md
- [ ] 3.1.15 Add "## Optimal Tooling" section after frontmatter in task-heavy.md
- [ ] 3.1.16 Add "## Optimal Tooling" section after frontmatter in tester.md

Validation:
- [ ] All 14 agents have Optimal Tooling section
- [ ] Table format consistent across all files

---

### T3.2: Add Context Gathering Phase
**Files:** `.claude/agents/coder.md`, `.claude/agents/debugger.md`, `.claude/agents/debugger-fast.md`, `.claude/agents/debugger-heavy.md`, `.claude/agents/decomposer.md`, `.claude/agents/planner.md`, `.claude/agents/planner-fast.md`, `.claude/agents/planner-heavy.md`, `.claude/agents/reproducer.md`, `.claude/agents/task-heavy.md`
**Dependencies:** T3.1
**Parallel with:** None

Steps:
- [ ] 3.2.1 Add "### 1. Context Gathering" workflow phase to coder.md
- [ ] 3.2.2 Add "### 1. Context Gathering" workflow phase to debugger.md
- [ ] 3.2.3 Add "### 1. Context Gathering" workflow phase to debugger-fast.md
- [ ] 3.2.4 Add "### 1. Context Gathering" workflow phase to debugger-heavy.md
- [ ] 3.2.5 Add "### 1. Context Gathering" workflow phase to decomposer.md
- [ ] 3.2.6 Add "### 1. Context Gathering" workflow phase to planner.md
- [ ] 3.2.7 Add "### 1. Context Gathering" workflow phase to planner-fast.md
- [ ] 3.2.8 Add "### 1. Context Gathering" workflow phase to planner-heavy.md
- [ ] 3.2.9 Add "### 1. Context Gathering" workflow phase to reproducer.md
- [ ] 3.2.10 Add "### 1. Context Gathering" workflow phase to task-heavy.md

Validation:
- [ ] All 10 execution agents have context gathering phase
- [ ] Phase includes todoread, skill, list, glob, and rg commands

## Phase 3 Checkpoint
- [ ] Optimal Tooling section added to all 14 agents
- [ ] Context gathering phase added to 10 execution agents
- [ ] All agent files validated for consistency

## Phase 4: Fix Agent Inheritance Mechanism

### T4.1: Fix mergeAgents() to use additive inheritance for slices
**Files:** `internal/cli/init.go`
**Dependencies:** T3.2
**Parallel with:** None

Steps:
- [ ] 4.1.1 Change mergeAgents() function to merge slices additively instead of replacing
- [ ] 4.1.2 Create mergeSlices() helper function
- [ ] 4.1.3 Update merge logic for: Skills
- [ ] 4.1.4 Update merge logic for: ToolPresets
- [ ] 4.1.5 Update merge logic for: Tools
- [ ] 4.1.6 Update merge logic for: DisallowedToolPresets
- [ ] 4.1.7 Update merge logic for: DisallowedTools
- [ ] 4.1.8 Update merge logic for: Dependencies

Validation:
- [ ] Child agents inherit parent slices additively
- [ ] Duplicate values are handled correctly
- [ ] Empty slices don't override parent values

---

### T4.2: Add planning preset to planner agents
**Files:** `pkg/agents/meta/planner.yaml`, `pkg/agents/meta/planner-fast.yaml`, `pkg/agents/meta/planner-heavy.yaml`
**Dependencies:** T4.1
**Parallel with:** None

Steps:
- [ ] 4.2.1 Add "planning" to toolPresets for planner.yaml
- [ ] 4.2.2 Add "planning" to toolPresets for planner-fast.yaml
- [ ] 4.2.3 Add "planning" to toolPresets for planner-heavy.yaml

Validation:
- [ ] All three planner variants have "planning" preset
- [ ] YAML syntax valid for all files

---

### T4.3: Document inheritance behavior
**Files:** Create `docs/AGENT_INHERITANCE.md`
**Dependencies:** T4.2
**Parallel with:** None

Steps:
- [ ] 4.3.1 Document how agent inheritance works (extends field)
- [ ] 4.3.2 Explain additive merge semantics for slices
- [ ] 4.3.3 Provide examples of agent inheritance
- [ ] 4.3.4 Document override behavior for non-slice fields

Validation:
- [ ] Documentation is clear and complete
- [ ] Examples are accurate and helpful

---

### T4.4: Make tool_list.go dynamic
**Files:** `internal/mcp/tools/tool_list.go`, `internal/mcp/tools/register.go`
**Dependencies:** T4.3
**Parallel with:** None

Steps:
- [ ] 4.4.1 Create tool registry to track registered tools
- [ ] 4.4.2 Modify register functions to populate registry
- [ ] 4.4.3 Update tool_list.go to use dynamic registry instead of hardcoded list
- [ ] 4.4.4 Ensure thread-safety for registry access

Validation:
- [ ] Tools are dynamically registered
- [ ] tool_list.go reflects actual registered tools
- [ ] No hardcoded tool list remains

## Phase 4 Checkpoint
- [ ] mergeAgents() uses additive inheritance for slices
- [ ] All planner agents have "planning" preset
- [ ] Inheritance documentation created
- [ ] Tool list is dynamically generated

## Completion Summary

**Estimated Timeline:** 3-4 days
**Total Tasks:** 20 (T1.1-T1.4, T2.1-T2.3, T3.1-T3.2, T4.1-T4.4)
**Files Modified:** 14 agent files + 3 internal files + 1 new doc

### Final Validation
- [ ] All 14 agent files have consistent tool configurations
- [ ] No grep commands remain (verified with `rg "grep -r" .claude/agents/`)
- [ ] No find commands remain (verified with `rg "find " .claude/agents/`)
- [ ] All agents have Optimal Tooling section
- [ ] Execution agents have Context Gathering phase
- [ ] Agent configuration is valid YAML
- [ ] Agent inheritance works correctly with additive merge
- [ ] Tool list is dynamically generated from registry
