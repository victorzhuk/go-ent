# Tasks: Align plugins/go-ent with FLOW.md Architecture

## 1. Command Analysis & Planning
- [ ] Audit all 17 existing commands for usage patterns
- [ ] Map current commands to FLOW.md commands
- [ ] Design migration path for deprecated commands
- [ ] Create feature flag system (FLOW_MODE=enabled|legacy)
- [ ] Document breaking changes

## 2. Core Commands Implementation

### 2.1 /task Command
- [ ] Create plugins/go-ent/commands/task.md <!-- depends: 3.8 -->
- [ ] Implement task selection (from registry or by ID)
- [ ] Load task context (spec, design.md, proposal.md)
- [ ] Implement local execution (Claude Code)
- [ ] Implement ACP delegation (OpenCode)
- [ ] Update tasks.md checkbox on completion
- [ ] Regenerate state.md after task completion
- [ ] Add --model flag for model override
- [ ] Add --local flag to force Claude Code execution

### 2.2 /bug Command
- [ ] Create plugins/go-ent/commands/bug.md
- [ ] Implement bug description parsing
- [ ] Trigger @reproducer agent first
- [ ] Implement smoke/heavy escalation logic
- [ ] Create regression test requirement
- [ ] Integrate with /task for fix execution

### 2.3 /plan Command Updates
- [ ] Update plugins/go-ent/commands/plan.md <!-- depends: 3.4 -->
- [ ] Integrate with @planner-smoke for triage
- [ ] Add --architecture flag to force @architect
- [ ] Improve OpenSpec output format
- [ ] Add task decomposition step

## 3. Planning Agents (Claude Code)

### 3.1 @planner-smoke
- [ ] Create plugins/go-ent/agents/planning/planner-smoke.md
- [ ] Model: claude-haiku-4-5-20251001
- [ ] Decision: needs_architecture (true/false)
- [ ] Max tokens: 1024
- [ ] Trigger conditions documented

### 3.2 @architect (Refactor)
- [ ] Move agents/architect.md → agents/planning/architect.md
- [ ] Model: claude-opus-4-5-20250514
- [ ] Enable extended thinking (32K budget)
- [ ] Output: Architecture Decision Record format
- [ ] Max tokens: 16384

### 3.3 @planner (Refactor)
- [ ] Move agents/planner.md → agents/planning/planner.md
- [ ] Model: claude-sonnet-4-5-20250929
- [ ] Output: OpenSpec feature format
- [ ] Max tokens: 8192
- [ ] Integrate with @decomposer

### 3.4 @decomposer
- [ ] Create agents/planning/decomposer.md
- [ ] Model: claude-sonnet-4-5-20250929
- [ ] Output: Task list with dependencies
- [ ] Task sizing: 1-4 hours each
- [ ] Max tokens: 8192

## 4. Execution Agents (OpenCode)

### 4.1 Triage Agents
- [ ] Create agents/execution/task-smoke.md
- [ ] Create agents/execution/debugger-smoke.md
- [ ] Model: glm-4.7 for both
- [ ] Decision: PROCEED | ESCALATE
- [ ] Escalation triggers documented

### 4.2 Heavy Agents
- [ ] Create agents/execution/task-heavy.md
- [ ] Create agents/execution/debugger-heavy.md
- [ ] Model: kimi-k2-thinking-turbo
- [ ] Deep reasoning instructions
- [ ] Complex case patterns

### 4.3 Core Execution Agents
- [ ] Create agents/execution/coder.md (from dev.md)
- [ ] Create agents/execution/reviewer.md (refactor existing)
- [ ] Create agents/execution/tester.md (refactor existing)
- [ ] Model: glm-4.7 for all
- [ ] Stack-specific patterns

### 4.4 Acceptance & Quality
- [ ] Create agents/execution/acceptor.md
- [ ] Model: glm-4.7
- [ ] Acceptance checklist
- [ ] Final verification logic

### 4.5 Bug Workflow Agents
- [ ] Create agents/execution/reproducer.md
- [ ] Create agents/execution/researcher.md
- [ ] Model: glm-4.7
- [ ] Reproduction steps format
- [ ] Research questions template

## 5. ACP Integration

### 5.1 Client Implementation
- [ ] Create internal/acp/client.go
- [ ] Implement Connect(endpoint) error
- [ ] Implement ExecuteTask(task, model) (Response, error)
- [ ] Implement ExecuteBug(description, model) (Response, error)
- [ ] Add health check and fallback logic

### 5.2 Protocol Implementation
- [ ] Define ACP request/response format
- [ ] Implement task serialization
- [ ] Implement context passing (spec, design, etc.)
- [ ] Add timeout and retry logic
- [ ] Implement streaming response handling

### 5.3 Model Selection
- [ ] Create internal/acp/models.go
- [ ] Define ModelSpec (name, provider, limits)
- [ ] Implement SelectModel(complexity) ModelSpec
- [ ] Primary: GLM 4.7 config
- [ ] Heavy: Kimi K2 config
- [ ] Fallback: Claude models

## 6. Escalation System

### 6.1 Rules Engine
- [ ] Create internal/agent/escalation.go
- [ ] Define EscalationRule struct
- [ ] Implement ShouldEscalate(task, history) bool
- [ ] Add complexity scoring
- [ ] Add retry count tracking

### 6.2 Triggers
- [ ] Keyword triggers (race condition, deadlock, etc.)
- [ ] Complexity threshold (>0.8)
- [ ] Retry count (>2 failed attempts)
- [ ] Task metadata signals
- [ ] Performance degradation patterns

### 6.3 Integration
- [ ] Wire escalation into /task command
- [ ] Wire escalation into /bug command
- [ ] Add metrics for escalation rates
- [ ] Log escalation decisions

## 7. Command Deprecation

### 7.1 Mark as Deprecated (Add warnings)
- [ ] Add deprecation notice to analyze.md
- [ ] Add deprecation notice to clarify.md
- [ ] Add deprecation notice to research.md
- [ ] Add deprecation notice to decompose.md
- [ ] Add deprecation notice to gen.md
- [ ] Add deprecation notice to scaffold.md
- [ ] Add deprecation notice to tdd.md
- [ ] Add deprecation notice to lint.md
- [ ] Add deprecation notice to loop.md
- [ ] Add deprecation notice to loop-cancel.md

### 7.2 Migration Guides
- [ ] Create docs/migration/analyze-to-plan.md
- [ ] Create docs/migration/clarify-to-plan.md
- [ ] Create docs/migration/gen-to-task.md
- [ ] Create docs/migration/tdd-to-task.md
- [ ] Create docs/migration/loop-to-task.md

### 7.3 Archive (After transition period)
- [ ] Move deprecated commands to commands/archive/
- [ ] Update command registry
- [ ] Remove from documentation

## 8. Skill Updates

### 8.1 Skill Auto-Activation
- [ ] Update skills to specify executor (claude-code | opencode | both)
- [ ] Add model preference per skill
- [ ] Update go-code.md with GLM 4.7 patterns
- [ ] Update go-test.md with testcontainers for GLM
- [ ] Update go-review.md for confidence filtering

### 8.2 New Skills
- [ ] Create skill: task-execution.md (orchestration)
- [ ] Create skill: bug-workflow.md (debugging flow)
- [ ] Create skill: escalation-analysis.md (when to escalate)

## 9. Configuration

### 9.1 Model Configuration
- [ ] Create config/models.yaml format
- [ ] Add GLM 4.7 provider config
- [ ] Add Kimi K2 provider config
- [ ] Add fallback chain (GLM → Kimi → Claude)
- [ ] Add cost tracking per model

### 9.2 Feature Flags
- [ ] Add FLOW_MODE flag (enabled | legacy)
- [ ] Add ACP_ENABLED flag
- [ ] Add ESCALATION_ENABLED flag
- [ ] Add fallback behavior configs

## 10. Documentation

### 10.1 Update Existing
- [ ] Update docs/FLOW.md with implementation status <!-- depends: 2.9, 4.20, 5.3 -->
- [ ] Update openspec/AGENTS.md with new agents
- [ ] Update README.md with new commands
- [ ] Update DEVELOPMENT.md with FLOW mode

### 10.2 New Documentation
- [ ] Create docs/acp-integration.md
- [ ] Create docs/model-selection.md
- [ ] Create docs/escalation-rules.md
- [ ] Create docs/opencode-setup.md

## 11. Testing

### 11.1 Unit Tests
- [ ] Test ACP client connection/fallback
- [ ] Test model selection logic
- [ ] Test escalation rules
- [ ] Test command routing (new vs deprecated)

### 11.2 Integration Tests
- [ ] Test /plan → /task workflow
- [ ] Test /bug workflow with escalation
- [ ] Test ACP delegation (mocked OpenCode)
- [ ] Test fallback to Claude Code

### 11.3 End-to-End Tests
- [ ] Test full feature development (plan → tasks)
- [ ] Test bug fix workflow
- [ ] Test smoke → heavy escalation
- [ ] Test with real OpenCode instance (if available)

## 12. Rollout

### 12.1 Alpha (Internal)
- [ ] Enable FLOW_MODE for go-ent development
- [ ] Test on real features
- [ ] Gather metrics (escalation rate, completion rate)
- [ ] Fix critical issues

### 12.2 Beta (Opt-in)
- [ ] Document opt-in process
- [ ] Add telemetry for usage patterns
- [ ] Create feedback channel
- [ ] Address user feedback

### 12.3 GA (General Availability)
- [ ] Set FLOW_MODE=enabled as default
- [ ] Announce deprecations
- [ ] Provide 1-month transition period
- [ ] Archive old commands
