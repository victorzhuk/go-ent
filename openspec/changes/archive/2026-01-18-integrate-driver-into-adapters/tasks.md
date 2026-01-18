# Tasks: Integrate Driver into Adapters

**Change ID**: integrate-driver-into-adapters
**Total Estimated**: ~8 hours

---

## Phase 1: Prompt Consolidation (3h)

- [ ] **1.1** Create base and driver prompt templates (1h)
  - Files: `plugins/sources/go-ent/agents/prompts/_base.md`, `_driver.md`
  - Extract universal agent patterns from existing prompts
  - Define orchestrator capabilities in `_driver.md`
  - Dependencies: none

- [ ] **1.2** Migrate content from /prompts/ (1h)
  - Files: Copy content from `/prompts/agents/driver.md` and `/prompts/shared/tooling.md`
  - Integrate into `_driver.md` and `shared/_tooling.md`
  - Preserve orchestration patterns
  - Dependencies: 1.1

- [ ] **1.3** Update existing agent prompts (1h)
  - Files: `plugins/sources/go-ent/agents/prompts/agents/*.md`
  - Reference `_base.md` for common patterns
  - Add driver handoff instructions where needed
  - Dependencies: 1.1

## Phase 2: Platform Driver Implementation (3h)

- [ ] **2.1** Implement Claude driver (1.5h)
  - Files: `plugins/platforms/claude/driver.go`
  - Claude-specific agent delegation logic
  - Task routing and context management
  - Dependencies: 1.3

- [ ] **2.2** Implement OpenCode driver (1.5h)
  - Files: `plugins/platforms/opencode/driver.go`
  - OpenCode-specific orchestration patterns
  - Platform-specific workflow management
  - Dependencies: 1.3
  - Parallel with: 2.1

## Phase 3: Adapter Integration (1.5h)

- [ ] **3.1** Update Claude adapter (0.75h)
  - Files: `internal/toolinit/claude.go`
  - Integrate driver capabilities
  - Load driver prompts during compilation
  - Dependencies: 2.1

- [ ] **3.2** Update OpenCode adapter (0.75h)
  - Files: `internal/toolinit/opencode.go`
  - Integrate driver capabilities
  - Load driver prompts during compilation
  - Dependencies: 2.2
  - Parallel with: 3.1

## Phase 4: Cleanup & Validation (0.5h)

- [ ] **4.1** Delete legacy prompts directory (0.25h)
  - Files: Delete `/prompts/` directory
  - Verify no references remain
  - Dependencies: 3.1, 3.2

- [ ] **4.2** Update documentation (0.25h)
  - Files: `docs/DEVELOPMENT.md`, `README.md`
  - Document driver pattern
  - Update prompt location references
  - Dependencies: 4.1

---

## Critical Path

```
1.1 → 1.2 → 1.3 → [2.1, 2.2] → [3.1, 3.2] → 4.1 → 4.2
```

**Parallel work:**
- Tasks 2.1 and 2.2 can run in parallel
- Tasks 3.1 and 3.2 can run in parallel
