# Tasks: Standardize Command Naming to goent:* Prefix

**Change ID:** refactor-standardize-goent-commands
**Status:** COMPLETED

## Phase 1: Cleanup & Preparation

**Objective:** Remove obsolete stub commands and validate current state

- [x] **T001** Delete obsolete stub commands
  - Delete `plugins/go-ent/commands/checklist.md`
  - Delete `plugins/go-ent/commands/review.md`
  - Delete `plugins/go-ent/commands/test.md`
  - Verify no references to these commands exist in codebase
  - **Validation:** `grep -r "checklist\|review\|test" plugins/go-ent/ README.md AGENTS.md` returns no command references
  - **Est:** 15min

- [x] **T002** Audit current command references in documentation
  - Search README.md for all command invocations
  - Search AGENTS.md for all command invocations
  - Search TRANSFORMATION.md for all command invocations
  - Create list of files requiring updates
  - **Validation:** Reference list created in `refactor-notes.md`
  - **Est:** 20min

**Checkpoint 1:** `go build ./...` passes, 3 files deleted, reference audit complete

---

## Phase 2: Rename Utility Commands

**Objective:** Standardize utility commands to goent:* prefix

- [x] **T003** Rename init.md to goent:init.md
  - `mv plugins/go-ent/commands/init.md plugins/go-ent/commands/goent:init.md`
  - Update YAML frontmatter `name:` field to `goent:init`
  - Update `description:` if it references command name
  - **Validation:** Execute `/goent:init --help` successfully
  - **Est:** 10min
  - **Depends on:** T001

- [x] **T004** Rename scaffold.md to goent:scaffold.md
  - `mv plugins/go-ent/commands/scaffold.md plugins/go-ent/commands/goent:scaffold.md`
  - Update YAML frontmatter `name:` field to `goent:scaffold`
  - Update `description:` if it references command name
  - **Validation:** Execute `/goent:scaffold --help` successfully
  - **Est:** 10min
  - **Depends on:** T001

- [x] **T005** Rename lint.md to goent:lint.md
  - `mv plugins/go-ent/commands/lint.md plugins/go-ent/commands/goent:lint.md`
  - Update YAML frontmatter `name:` field to `goent:lint`
  - Update `description:` if it references command name
  - **Validation:** Execute `/goent:lint --help` successfully
  - **Est:** 10min
  - **Depends on:** T001

**Checkpoint 2:** All 3 utility commands execute with goent:* prefix

---

## Phase 3: Consolidate Planning Workflow

**Objective:** Standardize planning sub-commands and consolidate plan workflow

### Sub-phase 3a: Rename Planning Sub-Commands

- [x] **T006** Rename clarify.md to goent:clarify.md
  - `mv plugins/go-ent/commands/clarify.md plugins/go-ent/commands/goent:clarify.md`
  - Update YAML frontmatter `name:` field to `goent:clarify`
  - **Validation:** Execute `/goent:clarify <change-id>` successfully
  - **Est:** 10min
  - **Depends on:** T001

- [x] **T007** Rename research.md to goent:research.md
  - `mv plugins/go-ent/commands/research.md plugins/go-ent/commands/goent:research.md`
  - Update YAML frontmatter `name:` field to `goent:research`
  - **Validation:** Execute `/goent:research <change-id>` successfully
  - **Est:** 10min
  - **Depends on:** T001

- [x] **T008** Rename decompose.md to goent:decompose.md
  - `mv plugins/go-ent/commands/decompose.md plugins/go-ent/commands/goent:decompose.md`
  - Update YAML frontmatter `name:` field to `goent:decompose`
  - **Validation:** Execute `/goent:decompose <change-id>` successfully
  - **Est:** 10min
  - **Depends on:** T001

- [x] **T009** Rename analyze.md to goent:analyze.md
  - `mv plugins/go-ent/commands/analyze.md plugins/go-ent/commands/goent:analyze.md`
  - Update YAML frontmatter `name:` field to `goent:analyze`
  - **Validation:** Execute `/goent:analyze <change-id>` successfully
  - **Est:** 10min
  - **Depends on:** T001

### Sub-phase 3b: Consolidate Plan Commands

- [x] **T010** Update plan-full.md orchestration calls
  - Open `plugins/go-ent/commands/plan-full.md`
  - Find all internal command calls: `/clarify`, `/research`, `/decompose`, `/analyze`
  - Replace with: `/goent:clarify`, `/goent:research`, `/goent:decompose`, `/goent:analyze`
  - Remove `/checklist` call (deleted in T001)
  - **Validation:** All command calls use `goent:` prefix
  - **Est:** 20min
  - **Depends on:** T006, T007, T008, T009

- [x] **T011** Rename plan-full.md to goent:plan.md
  - `mv plugins/go-ent/commands/plan-full.md plugins/go-ent/commands/goent:plan.md`
  - Update YAML frontmatter `name:` field to `goent:plan`
  - Update `description:` to reflect this is the primary planning command
  - **Validation:** Execute `/goent:plan "test feature"` successfully
  - **Est:** 10min
  - **Depends on:** T010

- [x] **T012** Delete plan.md (functionality merged)
  - Delete `plugins/go-ent/commands/plan.md`
  - Verify plan-full.md contains equivalent or superior functionality
  - **Validation:** File deleted, no references remain
  - **Est:** 5min
  - **Depends on:** T011

**Checkpoint 3:** Planning workflow executes end-to-end with new command names

---

## Phase 4: Documentation Updates

**Objective:** Update all documentation to reference new command names

- [x] **T013** Update README.md command references
  - Replace all `/init` with `/goent:init`
  - Replace all `/scaffold` with `/goent:scaffold`
  - Replace all `/lint` with `/goent:lint`
  - Replace all `/plan` or `/plan-full` with `/goent:plan`
  - Update command table/list with all 15 goent:* commands
  - **Validation:** `grep -E "/(init|scaffold|lint|plan|plan-full|clarify|research|decompose|analyze)\b" README.md` returns no matches
  - **Est:** 30min
  - **Depends on:** T011, T012

- [x] **T014** Update AGENTS.md command references
  - Find all command references in agent instructions
  - Update to use goent:* prefix
  - Verify agent delegation logic still valid
  - **Validation:** All command references use goent:* prefix
  - **Est:** 20min
  - **Depends on:** T011, T012

- [x] **T015** Update TRANSFORMATION.md command references
  - Find all command references
  - Update to use goent:* prefix
  - Update workflow examples
  - **Validation:** All command references use goent:* prefix
  - **Est:** 15min
  - **Depends on:** T011, T012

- [x] **T016** Check and update plugin hooks
  - Review `plugins/go-ent/hooks/hooks.json`
  - Verify no hard-coded command name dependencies
  - Update if necessary
  - **Validation:** Hooks execute without errors
  - **Est:** 15min
  - **Depends on:** T011, T012

- [x] **T017** Update agent prompt files
  - Search `plugins/go-ent/agents/*.md` for command references
  - Update any references to old command names
  - **Validation:** `grep -r "/(init|scaffold|lint|plan|plan-full|clarify|research|decompose|analyze)\b" plugins/go-ent/agents/` returns no matches
  - **Est:** 20min
  - **Depends on:** T011, T012

**Checkpoint 4:** All documentation updated and validated

---

## Phase 5: Final Validation & Testing

**Objective:** Comprehensive testing of refactored command structure

- [x] **T018** Execute full command inventory test
  - List all 15 commands: `ls plugins/go-ent/commands/goent:*.md`
  - Execute each command with `--help` or minimal args
  - Verify no errors
  - **Validation:** All 15 commands execute successfully
  - **Est:** 30min
  - **Depends on:** T013, T014, T015, T016, T017

- [x] **T019** Execute end-to-end workflow test
  - Create test project: `/goent:init test-refactor-validation`
  - Create test plan: `/goent:plan "Add test feature"`
  - Verify planning workflow: clarify → research → decompose → analyze
  - Execute task: `/goent:apply`
  - Archive change: `/goent:archive`
  - **Validation:** Complete workflow executes without errors
  - **Est:** 45min
  - **Depends on:** T018

- [x] **T020** Verify zero broken references
  - Run comprehensive grep: `grep -r "/(init|scaffold|lint|plan|plan-full|clarify|research|decompose|analyze|checklist|review|test)\b" plugins/go-ent/ README.md AGENTS.md TRANSFORMATION.md`
  - Verify only valid references (e.g., in historical context or comments)
  - **Validation:** No active broken command references found
  - **Est:** 15min
  - **Depends on:** T018

- [x] **T021** Update change status to COMPLETE
  - Run final build: `go build ./...`
  - Run final test: `go test ./...`
  - Update `proposal.md` status to COMPLETE
  - **Validation:** All builds pass, status updated
  - **Est:** 10min
  - **Depends on:** T019, T020

**Checkpoint 5:** All tests pass, 100% goent:* standardization achieved

---

## Summary

**Total Tasks:** 21
**Estimated Time:** ~5-6 hours
**Complexity:** MEDIUM
**Dependencies:** Sequential phases with internal parallelization possible

**Critical Path:**
T001 → T010 → T011 → T012 → T013-T017 → T018 → T019 → T020 → T021

**Parallel Opportunities:**
- T003, T004, T005 can run in parallel (after T001)
- T006, T007, T008, T009 can run in parallel (after T001)
- T013, T014, T015, T016, T017 can run in parallel (after T012)

**Risk Areas:**
- **T010:** Orchestration call updates - must be precise
- **T011:** Renaming plan-full to primary plan - critical for workflow
- **T019:** End-to-end test - validates entire refactor

**Success Criteria:**
- 15 commands, all with `goent:*` prefix
- Zero broken references
- Complete workflow test passes
- Documentation fully updated
