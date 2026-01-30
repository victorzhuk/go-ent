# Proposal: CLI Cleanup and ACP Integration (Phase 5)

## Metadata
- **Change ID:** `simplify-05-cli-cleanup`
- **Status:** Proposed
- **Type:** Refactoring
- **Priority:** High
- **Affects Specs:** None
- **Part of:** Go-Ent Simplification Series (5/5)
- **Depends On:** `simplify-04-mcp-tools`

## Problem

After Phases 1-4, the CLI layer still has:
- References to deleted packages (execution, worker, agent)
- Commands for deleted systems (agent, model, migrate)
- No integration with new ACP client
- Dead code from over-engineered systems

The CLI needs to:
1. Use ACP client for task execution
2. Remove commands for deleted systems
3. Update spec/skill commands to use simplified packages
4. Clean up dead imports and code

## Proposed Solution

### Update Core Commands (Use ACP)

1. **cli/run.go** - Use ACP client instead of execution engine
   - Replace execution.Engine with acp.Executor
   - Execute tasks via ACP
   - Stream output from ACP

2. **cli/task.go** - Use ACP client for task execution
   - Replace worker spawning with ACP execution
   - Use acp.ExecuteTask for single tasks
   - Use acp.ExecuteTasks for parallel execution

### Remove Obsolete Commands

3. **cli/agent.go** - DELETE (agents execute via ACP)
4. **cli/model.go** - DELETE (model selection via ACP)
5. **cli/migrate.go** - DELETE (no migrations needed)

### Update Existing Commands

6. **cli/spec.go** - Use simplified spec/ package
   - Update imports from internal/openspec to internal/spec
   - Verify commands work with merged package

7. **cli/skill.go** - Use simplified skill/ package
   - Verify commands work with simplified package
   - Remove any dead code

### Clean Up Initialization

8. **cmd/go-ent/main.go** - Clean up initialization
   - Remove references to deleted packages
   - Add ACP client initialization if needed
   - Simplify command registration

## Impact

- **Breaking Changes:** CLI commands removed (agent, model, migrate)
- **API Changes:** CLI now uses ACP for execution
- **Migration Required:** No (pre-release)
- **Testing Required:** Yes
  - Verify remaining commands work
  - Test full `/ent:plan` → `/ent:apply` cycle
  - Verify task execution via ACP

## Risks

- **Low Risk:** CLI changes are isolated
- **Testing:** Need to verify full workflow
- **Rollback:** Git revert is straightforward

## Dependencies

- **Previous Proposal:** `simplify-04-mcp-tools` (must complete first)
- **Completes:** Go-Ent Simplification Series

## Success Criteria

- [ ] cli/run.go uses ACP client
- [ ] cli/task.go uses ACP client
- [ ] cli/agent.go deleted
- [ ] cli/model.go deleted
- [ ] cli/migrate.go deleted
- [ ] cli/spec.go updated
- [ ] cli/skill.go updated
- [ ] cmd/go-ent/main.go cleaned up
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] Full `/ent:plan` → `/ent:apply` cycle works
- [ ] Parallel task execution works
- [ ] Registry state consistency maintained
