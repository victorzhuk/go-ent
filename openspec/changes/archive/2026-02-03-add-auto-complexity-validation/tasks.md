# Implementation Tasks

## 1. Update Validation Code

- [x] 1.1 Open `internal/cli/init.go` and locate the `validateAgent()` function
- [x] 1.2 Find the `validComplexity` map definition (around line 448)
- [x] 1.3 Add `"auto": true` to the `validComplexity` map
- [x] 1.4 Update the validation error message to include "auto" in the list of valid values
- [x] 1.5 Verify the changes are correctly formatted

## 2. Verify Fix

- [x] 2.1 Run validation command: `make build && ./dist/go-ent validate`
- [x] 2.2 Confirm that debugger agent now passes validation
- [x] 2.3 Verify all other agents continue to validate successfully
- [x] 2.4 Check that the validation count matches expected number of agents

## 3. Test Coverage (if applicable)

- [x] 3.1 Check if there are existing tests for `validateAgent()` function
- [x] 3.2 If tests exist, update test expectations to include "auto" as a valid complexity value
- [x] 3.3 Run tests: `go test ./internal/cli/...`
- [x] 3.4 Verify all tests pass

## 4. Integration Testing

- [x] 4.1 Test that invalid complexity values still produce correct error message
- [x] 4.2 Verify error message lists all valid values: [auto, simple, standard, heavy]
- [x] 4.3 Confirm that the debugger agent's `complexityHints` and `modelMapping` are processed correctly

## 5. Build and Final Verification

- [x] 5.1 Build the project: `make build`
- [x] 5.2 Run validation: `./dist/go-ent validate`
- [x] 5.3 Verify output shows all agents validated successfully
- [x] 5.4 Run all tests: `make test`
- [x] 5.5 Run linter: `make lint`

## Implementation Notes

This is a straightforward bug fix with minimal code changes. The primary focus is ensuring that:

1. The "auto" complexity value is accepted by validation
2. The error message accurately reflects all valid values
3. All existing functionality remains intact (backward compatibility)

No breaking changes or migration is required.
