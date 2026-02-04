# Change: Add "auto" to Valid Complexity Values

## Why

The agent metadata validation in `internal/cli/init.go` doesn't recognize "auto" as a valid complexity value. However, `pkg/agents/meta/debugger.yaml` uses `complexity: auto` for dynamic model selection based on complexity hints and model mapping.

The validation function `validateAgent()` only allows "simple", "standard", and "heavy" as valid complexity values, causing validation to fail for the debugger agent.

## What Changes

- Update `validComplexity` map in `internal/cli/init.go` to include "auto" as a valid complexity value
- Update validation error message to reflect the new list of valid values
- Update test expectations if any tests verify the error message

## Impact

- **Affected specs**: fix-validation (NEW)
- **Affected code**: `internal/cli/init.go` - `validateAgent()` function (line 447-452)
- **Breaking change**: None
- **Migration**: None - this is a bug fix
