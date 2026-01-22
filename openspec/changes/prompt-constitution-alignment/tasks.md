# Tasks: Align Agent Prompts with Constitutional AI Principles

## 1. Create judgment guidance
- [ ] 1.1 Create `plugins/go-ent/agents/prompts/shared/_judgment.md`
- [ ] 1.2 Document when to exercise judgment over strict rule following
- [ ] 1.3 Add "thoughtful senior developer" test criteria
- [ ] 1.4 Include examples of appropriate judgment calls
- [ ] 1.5 Define boundaries (when NOT to deviate from rules)

## 2. Create principal hierarchy
- [ ] 2.1 Create `plugins/go-ent/agents/prompts/shared/_principals.md`
- [ ] 2.2 Define hierarchy: Project conventions > User intent > Best practices
- [ ] 2.3 Add conflict resolution guidance
- [ ] 2.4 Document when to ask vs when to decide
- [ ] 2.5 Include escalation criteria for ambiguous situations

## 3. Update handoff guidelines
- [ ] 3.1 Update `plugins/go-ent/agents/prompts/shared/_handoffs.md`
- [ ] 3.2 Add irreversible action checkpoints (e.g., before deletions, force-push)
- [ ] 3.3 Define escalation triggers for uncertain situations
- [ ] 3.4 Clarify handoff vs escalation distinction

## 4. Update agent prompts
- [ ] 4.1 Include `_judgment.md` in architect agent prompt
- [ ] 4.2 Include `_judgment.md` in dev agent prompt
- [ ] 4.3 Include `_judgment.md` in planner agent prompt
- [ ] 4.4 Include new principals in all agent prompts
- [ ] 4.5 Test agent behavior with updated prompts (manual verification)

## 5. Documentation
- [ ] 5.1 Document prompt design principles in `docs/PROMPT_DESIGN.md`
- [ ] 5.2 Add examples of good vs bad judgment calls
- [ ] 5.3 Create testing checklist for prompt changes
