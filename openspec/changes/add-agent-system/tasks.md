# Tasks: Add Agent System

## 1. Agent Selector
- [ ] Create `internal/agent/selector.go`
- [ ] Implement `Select(ctx, task) (AgentRole, []Skill, error)`
- [ ] Add complexity-based role selection
- [ ] Add budget-aware selection
- [ ] Unit tests for selection logic

## 2. Complexity Analyzer
- [ ] Create `internal/agent/complexity.go`
- [ ] Implement `Analyze(task) TaskComplexity`
- [ ] Define complexity levels (Trivial to Architectural)
- [ ] Add pattern matching rules
- [ ] Unit tests for complexity classification

## 3. Delegation Logic
- [ ] Create `internal/agent/delegate.go`
- [ ] Implement decision matrix from `plugins/go-ent/agents/go-ent:lead.md`
- [ ] Add `CanDelegate(from, to AgentRole) bool`
- [ ] Add `GetDelegationChain(task) []AgentRole`
- [ ] Unit tests for delegation

## 4. Skill Registry
- [ ] Create `internal/skill/registry.go`
- [ ] Implement `Register(skill Skill) error`
- [ ] Implement `MatchForContext(ctx SkillContext) []Skill`
- [ ] Add skill loading from markdown
- [ ] Unit tests for registry

## 5. Integration
- [ ] Update `internal/spec/workflow.go` with CurrentAgent field
- [ ] Update `internal/spec/store.go` with AgentsPath(), SkillsPath()
- [ ] Integration tests with config system
