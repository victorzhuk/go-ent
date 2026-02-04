# Complexity Validation Fix

## ADDED Requirements

### Requirement: Valid Complexity Values
The agent metadata validation SHALL accept "auto" as a valid complexity value.

#### Scenario: Validation accepts "auto" complexity
- **WHEN** validating an agent with `complexity: auto`
- **THEN** validation passes without error
- **AND** "auto" is recognized as a valid complexity value

#### Scenario: Validation accepts all valid complexity values
- **WHEN** validating agents with complexity values of "auto", "simple", "standard", or "heavy"
- **THEN** all agents pass validation
- **AND** no validation errors occur

#### Scenario: Validation rejects invalid complexity values
- **WHEN** validating an agent with an invalid complexity value (e.g., "medium", "high")
- **THEN** validation fails with an appropriate error message
- **AND** the error message lists all valid values: [auto, simple, standard, heavy]

### Requirement: Debugger Agent Validation
The debugger agent with `complexity: auto` SHALL pass validation.

#### Scenario: Debugger agent validates successfully
- **WHEN** running validation on the debugger agent metadata
- **THEN** validation passes
- **AND** the agent's `complexity: auto` field is accepted
- **AND** the agent's `complexityHints` and `modelMapping` fields are processed correctly
