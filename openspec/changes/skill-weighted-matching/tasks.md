# Tasks: Weighted Trigger Matching

## 1. Core Types
- [ ] 1.1 Add `MatchResult` type to `internal/skill/registry.go`
- [ ] 1.2 Add `MatchReason` type
- [ ] 1.3 Update `FindMatchingSkills` signature

## 2. Scoring Algorithm
- [ ] 2.1 Implement `scoreSkill()` with pattern/keyword/file matching
- [ ] 2.2 Implement `matchTrigger()` for explicit triggers
- [ ] 2.3 Implement fallback `matchDescription()` for old format

## 3. Pattern Caching
- [ ] 3.1 Implement regex pattern cache (map[string]*regexp.Regexp)
- [ ] 3.2 Compile patterns once, reuse across queries

## 4. Update Callers
- [ ] 4.1 Update CLI commands to handle MatchResult
- [ ] 4.2 Update tests to expect new return type

## 5. Testing
- [ ] 5.1 Unit tests for scoring algorithm
- [ ] 5.2 Test weighted vs unweighted triggers
- [ ] 5.3 Test pattern/keyword/file matching
- [ ] 5.4 Benchmark matching performance
