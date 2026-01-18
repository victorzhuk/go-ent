# Tasks: Enhanced Validation Rules

## 1. SK010: Example Diversity

### 1.1 Implement diversity scoring
- [ ] 1.1.1 Create `calculateDiversityScore()` to check input complexity variety
- [ ] 1.1.2 Check behavior variety (success/error/edge)
- [ ] 1.1.3 Check data type variety
- [ ] 1.1.4 Return 0.0-1.0 score

### 1.2 Implement SK010 rule
- [ ] 1.2.1 Create `checkExampleDiversity()` function
- [ ] 1.2.2 Return warning if diversity <50%
- [ ] 1.2.3 Include actionable suggestion and example
- [ ] 1.2.4 No error if <3 examples (count check is SK004)

## 2. SK011: Instruction Conciseness

### 2.1 Implement token counting
- [ ] 2.1.1 Create `countTokens()` function (words * 1.3 approximation)
- [ ] 2.1.2 Handle empty strings
- [ ] 2.1.3 Verify reasonably accurate (±10%)

### 2.2 Implement SK011 rule
- [ ] 2.2.1 Create `checkInstructionConcise()` function
- [ ] 2.2.2 Warn at 5k tokens
- [ ] 2.2.3 Critical warning at 8k tokens
- [ ] 2.2.4 Include suggestion to reduce content

## 3. SK012: Trigger Explicit

### 3.1 Implement SK012 rule
- [ ] 3.1.1 Create `checkTriggerExplicit()` function
- [ ] 3.1.2 Detect explicit triggers
- [ ] 3.1.3 Return info warning if using description-based triggers
- [ ] 3.1.4 Return warning if no triggers at all
- [ ] 3.1.5 Include example of explicit trigger format

## 4. SK013: Redundancy Detection

### 4.1 Implement overlap calculation
- [ ] 4.1.1 Create `calculateOverlap()` for skill similarity
- [ ] 4.1.2 Create `calculateTriggerOverlap()` to compare trigger sets
- [ ] 4.1.3 Create `calculateTextSimilarity()` to compare descriptions
- [ ] 4.1.4 Implement weighted average (70% triggers, 30% description)

### 4.2 Implement SK013 rule
- [ ] 4.2.1 Create `checkRedundancy()` function
- [ ] 4.2.2 Compare skill with all others
- [ ] 4.2.3 Return warning if overlap >70%
- [ ] 4.2.4 Identify most overlapping skill
- [ ] 4.2.5 Include suggestion to merge or differentiate

## 5. Integration

### 5.1 Register new rules in validator
- [ ] 5.1.1 Register SK010-SK012 in standard rules
- [ ] 5.1.2 Register SK013 for ValidateWithContext()
- [ ] 5.1.3 Document all rules with IDs and descriptions

### 5.2 Update CLI to show new warnings
- [ ] 5.2.1 Display new warnings with proper formatting
- [ ] 5.2.2 Make info-level warnings visually distinct from errors
- [ ] 5.2.3 Display suggestions and examples clearly

## 6. Testing

### 6.1 Unit tests for new rules
- [ ] 6.1.1 Test SK010 with diverse and non-diverse examples
- [ ] 6.1.2 Test SK011 with various token counts
- [ ] 6.1.3 Test SK012 with explicit/description/no triggers
- [ ] 6.1.4 Test SK013 with high/low overlap pairs
- [ ] 6.1.5 Cover all edge cases

### 6.2 Integration tests
- [ ] 6.2.1 Test full validation with all new rules
- [ ] 6.2.2 Test ValidateWithContext for SK013
- [ ] 6.2.3 Verify warnings don't block validation
- [ ] 6.2.4 Test with real skill files
