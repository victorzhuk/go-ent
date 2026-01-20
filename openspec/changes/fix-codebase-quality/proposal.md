# Proposal: Fix Codebase Quality Issues

## Status: draft

## Problem Statement

During verification of the skill-migration proposal, pre-existing codebase quality issues were discovered that prevent clean verification results:

- 256 lint issues across codebase
- Race conditions in background agent tests
- Nil pointer dereference panic in AST operations
- Template validation warnings

These issues are unrelated to skill-migration work but block verification cycles.

## Proposed Solution

### Phase 1: Fix Lint Issues
- Address 108 errcheck issues
- Fix 7 gocritic issues
- Resolve 130 gosec security warnings
- Fix 4 staticcheck issues
- Remove 7 unused variables

### Phase 2: Fix Test Failures
- Fix race conditions in internal/agent/background (9 tests)
- Fix nil pointer dereference panic in internal/ast

### Phase 3: Fix Template Validation
- Address template quality score assertion failures
- Fix SK010 strict validation warnings

## Files Affected
- internal/agent/background/ - race conditions
- internal/ast/ - nil pointer panic
- Multiple packages - lint issues
- plugins/go-ent/skills/**/*.json - template validation

## Success Criteria
- [ ] make lint passes with 0 errors
- [ ] make test passes with 0 failures
- [ ] make validate-templates passes
- [ ] All race conditions fixed
- [ ] All gosec security warnings resolved

## Dependencies
None - can proceed independently
