# Tasks: Weighted Trigger Matching

## 1. Core Types
- [x] 1.1 Add `MatchResult` type to `internal/skill/registry.go`
- [x] 1.2 Add `MatchReason` type
- [x] 1.3 Update `FindMatchingSkills` signature

## 2. Scoring Algorithm
- [x] 2.1 Implement `scoreSkill()` with pattern/keyword/file matching
- [x] 2.2 Implement `matchTrigger()` for explicit triggers
- [x] 2.3 Implement fallback `matchDescription()` for old format

## 3. Pattern Caching
- [x] 3.1 Implement regex pattern cache (map[string]*regexp.Regexp)
- [x] 3.2 Compile patterns once, reuse across queries

## 4. Update Callers
- [x] 4.1 Update CLI commands to handle MatchResult
- [x] 4.2 Update tests to expect new return type

## 5. Testing
- [x] 5.1 Unit tests for scoring algorithm
- [x] 5.2 Test weighted vs unweighted triggers
- [x] 5.3 Test pattern/keyword/file matching
- [x] 5.4 Benchmark matching performance
