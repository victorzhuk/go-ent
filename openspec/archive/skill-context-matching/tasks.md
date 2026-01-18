# Tasks: Context-Aware Matching

## Status: complete

## 1. Add MatchContext struct
- [x] 1.1 Define MatchContext type
- [x] 1.2 Update FindMatchingSkills signature

## 2. Implement context boosting
- [x] 2.1 Implement file-type boosting logic
- [x] 2.2 Implement task-type classification
- [x] 2.3 Implement skill affinity scoring

## 3. Update callers
- [x] 3.1 Provide context where available
- [x] 3.2 Add graceful degradation if no context

**Analysis:**
- No production callers of `FindMatchingSkills` exist yet
- Only test files (internal/skill/registry_test.go) use it
- Tests already use context properly
- Graceful degradation fully implemented (lines 163-165, 257-259, 305-307, 358)
- Backward compatible with nil/empty context via variadic parameter
- Ready for future integration

## 4. Testing
- [x] 4.1 Test context boosting scenarios
- [x] 4.2 Test without context (backward compat)
