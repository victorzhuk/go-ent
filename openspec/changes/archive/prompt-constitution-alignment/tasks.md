# Tasks: Align Agent Prompts with Constitutional AI Principles

## Status: complete

## 1. Create judgment guidance
- [x] 1.1 Create `plugins/go-ent/agents/prompts/shared/_judgment.md`
- [x] 1.2 Document when to exercise judgment over strict rule following
- [x] 1.3 Add "thoughtful senior developer" test criteria
- [x] 1.4 Include examples of appropriate judgment calls
- [x] 1.5 Define boundaries (when NOT to deviate from rules)

## 2. Create principal hierarchy
- [x] 2.1 Create `plugins/go-ent/agents/prompts/shared/_principals.md`
- [x] 2.2 Define hierarchy: Project conventions > User intent > Best practices
- [x] 2.3 Add conflict resolution guidance
- [x] 2.4 Document when to ask vs when to decide
- [x] 2.5 Include escalation criteria for ambiguous situations

## 3. Update handoff guidelines
- [x] 3.1 Update `plugins/go-ent/agents/prompts/shared/_handoffs.md`
- [x] 3.2 Add irreversible action checkpoints (e.g., before deletions, force-push)
- [x] 3.3 Define escalation triggers for uncertain situations
- [x] 3.4 Clarify handoff vs escalation distinction
- [x] 3.5 Add integration with judgment/principals guidance
- [x] 3.6 Include practical examples of checkpoint application

## 4. Update agent prompts
- [x] 4.1 Include `_judgment.md` in architect agent prompt
- [x] 4.2 Include `_judgment.md` in dev agent prompt
- [x] 4.3 Include `_judgment.md` in planner agent prompt
- [x] 4.4 Include new principals in all agent prompts
- [x] 4.5 Test agent behavior with updated prompts (manual verification)

## 5. Documentation
- [x] 5.1 Document prompt design principles in `docs/PROMPT_DESIGN.md`
- [x] 5.2 Add examples of good vs bad judgment calls
- [x] 5.3 Create testing checklist for prompt changes
