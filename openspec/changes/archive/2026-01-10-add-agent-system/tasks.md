# Tasks: Add Agent System

## 1. Agent Selector
- [x] Create `internal/agent/selector.go`
- [x] Implement `Select(ctx, task) (AgentRole, []Skill, error)`
- [x] Add complexity-based role selection
- [x] Add budget-aware selection
- [x] Unit tests for selection logic

## 2. Complexity Analyzer
- [x] Create `internal/agent/complexity.go`
- [x] Implement `Analyze(task) TaskComplexity`
- [x] Define complexity levels (Trivial to Architectural)
- [x] Add pattern matching rules
- [x] Unit tests for complexity classification

## 3. Delegation Logic
- [x] Create `internal/agent/delegate.go`
- [x] Implement decision matrix from `plugins/go-ent/agents/go-ent:lead.md`
- [x] Add `CanDelegate(from, to AgentRole) bool`
- [x] Add `GetDelegationChain(task) []AgentRole`
- [x] Unit tests for delegation

## 4. Skill Registry
- [x] Create `internal/skill/registry.go`
- [x] Implement `Register(skill Skill) error`
- [x] Implement `MatchForContext(ctx SkillContext) []Skill`
- [x] Add skill loading from markdown
- [x] Unit tests for registry

## 5. Integration
- [x] Update `internal/spec/workflow.go` with CurrentAgent field
- [x] Update `internal/spec/store.go` with AgentsPath(), SkillsPath()
- [x] Integration tests with config system
