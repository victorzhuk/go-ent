# Tasks: Research-Aligned Quality Scoring

## Status: in_progress

## 1. Foundation

### 1.1 Define new score types
- [x] 1.1.1 Create `QualityScore` struct with Total, Structure, Content, Examples, Triggers, Conciseness
- [x] 1.1.2 Define `StructureScore`, `ContentScore`, `ExamplesScore` sub-types
- [x] 1.1.3 Document all score fields

## 2. Scoring Implementation

### 2.1 Implement structure scoring
- [x] 2.1.1 Create `calculateStructureScore()` function
- [x] 2.1.2 Award points: Role: 4pts, Instructions: 4pts, Constraints: 3pts, Constraints: 3pts, Examples: 3pts, OutputFormat: 3pts, EdgeCases: 3pts
- [x] 2.1.3 Verify total max 20 points

### 2.2 Implement content scoring
- [x] 2.2.1 Implement `scoreRoleClarity()` for expertise, domain, behavior (0-8pts)
- [x] 2.2.2 Implement `scoreInstructions()` for actionability, specificity, structure (0-9pts)
- [x] 2.2.3 Implement `scoreConstraints()` for positive/negative rules, specificity (0-8pts)
- [x] 2.2.4 Verify total max 25 points

### 2.3 Implement examples scoring
- [x] 2.3.1 Implement count scoring: 0=0pts, 1=3pts, 2=6pts, 3-5=10pts, >5=8pts
- [x] 2.3.2 Implement diversity scoring: different input types, behaviors (0-8pts)
- [x] 2.3.3 Implement edge case scoring: 2pts per edge case, max 4pts
- [x] 2.3.4 Implement format scoring: input/output pairs + XML structure (0-3pts)
- [x] 2.3.5 Verify total max 25 points

### 2.4 Implement triggers scoring\n- [x] 2.4.1 Implement explicit triggers scoring: 10pts base + 3pts for weights + 2pts for diversity\n- [x] 2.4.2 Implement description-based fallback: 5pts max\n- [x] 2.4.3 Verify total max 15 points

### 2.5 Implement conciseness scoring
- [x] 2.5.1 Create token counting function (words * 1.3 approximation)
- [x] 2.5.2 Implement scoring curve: <3k=15pts, 3-5k=10pts, 5-8k=5pts, >8k=0pts
- [x] 2.5.3 Verify total max 15 points

## 3. Integration

### 3.1 Update CalculateQualityScore function
- [x] 3.1.1 Call all sub-scorers
- [x] 3.1.2 Return complete QualityScore struct
- [x] 3.1.3 Verify total equals sum of all categories

### 3.2 Update CLI output formatter
- [x] 3.2.1 Display total score and breakdown
- [x] 3.2.2 Show visual bars for each category
- [x] 3.2.3 Provide recommendations based on low scores
- [x] 3.2.4 Match design spec format

### 3.3 Add skill analysis command
- [x] 3.3.1 Create `cmd/go-ent/skill_analyze.go`
- [x] 3.3.2 Implement `skill analyze --all` for all skills
- [x] 3.3.3 Generate distribution report
- [x] 3.3.4 Identify common issues
- [x] 3.3.5 Export to JSON/CSV

## 4. Testing

### 4.1 Unit tests for all scorers
- [x] 4.1.1 Test structure scoring with various sections
- [x] 4.1.2 Test content scoring with quality variations
- [x] 4.1.3 Test examples scoring with different counts/diversity
- [x] 4.1.4 Test triggers scoring with explicit/description-based
- [x] 4.1.5 Test conciseness scoring with various token counts
- [x] 4.1.6 Cover all edge cases

### 4.2 Integration tests with real skills
- [x] 4.2.1 Load real skills from testdata/
- [x] 4.2.2 Verify scores are in valid ranges
- [x] 4.2.3 Verify breakdown sums to total
- [x] 4.2.4 Compare scores with expectations

### 4.3 Benchmark performance
- [x] 4.3.1 Benchmark scoring individual skills
- [x] 4.3.2 Benchmark batch scoring (100+ skills)
- [x] 4.3.3 Verify <10ms per skill
- [x] 4.3.4 Identify optimization opportunities

## 5. Migration

### 5.1 Analyze existing skills
- [x] 5.1.1 Run analyzer on all existing skills
- [x] 5.1.2 Generate distribution report
- [x] 5.1.3 Identify skills needing improvement
- [x] 5.1.4 Create prioritized improvement list

### 5.2 Update documentation
- [x] 5.2.1 Document new scoring breakdown
- [x] 5.2.2 Explain scoring criteria for each category
- [x] 5.2.3 Provide optimization tips
- [x] 5.2.4 Include score interpretation guide
