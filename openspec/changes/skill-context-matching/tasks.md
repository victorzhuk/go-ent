# Tasks: Context-Aware Matching

## 1. Add MatchContext struct
- [ ] 1.1 Define MatchContext type
- [ ] 1.2 Update FindMatchingSkills signature

## 2. Implement context boosting
- [ ] 2.1 Implement file-type boosting logic
- [ ] 2.2 Implement task-type classification
- [ ] 2.3 Implement skill affinity scoring

## 3. Update callers
- [ ] 3.1 Provide context where available
- [ ] 3.2 Add graceful degradation if no context

## 4. Testing
- [ ] 4.1 Test context boosting scenarios
- [ ] 4.2 Test without context (backward compat)
