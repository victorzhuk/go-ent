# Spec: Additive Inheritance for Agent Configuration

## Purpose
Define requirements for implementing additive inheritance in agent configuration merging, allowing agents to extend base configurations without replacing slice fields.

## Requirements

### Requirement: mergeSlices helper function
The system SHALL provide a `mergeSlices()` helper function that merges two string slices additively without duplicates.

#### Scenario: Merge two slices with no overlap
- **WHEN** `mergeSlices(["a", "b"], ["c", "d"])` is called
- **THEN** the result SHALL be `["a", "b", "c", "d"]`

#### Scenario: Merge two slices with duplicates
- **WHEN** `mergeSlices(["a", "b", "c"], ["b", "c", "d"])` is called
- **THEN** the result SHALL be `["a", "b", "c", "d"]` with no duplicates

#### Scenario: Merge with empty base slice
- **WHEN** `mergeSlices([], ["a", "b"])` is called
- **THEN** the result SHALL be `["a", "b"]`

#### Scenario: Merge with empty variant slice
- **WHEN** `mergeSlices(["a", "b"], [])` is called
- **THEN** the result SHALL be `["a", "b"]`

#### Scenario: Merge two empty slices
- **WHEN** `mergeSlices([], [])` is called
- **THEN** the result SHALL be `[]`

#### Scenario: Preserve order from base then variant
- **WHEN** `mergeSlices(["first", "second"], ["third", "fourth"])` is called
- **THEN** base items SHALL appear before variant items in result

### Requirement: Additive inheritance in mergeAgents
The system SHALL modify `mergeAgents()` to use additive merging for slice fields (Skills, Tools, ToolPresets) instead of replacement.

#### Scenario: Agent extends base with additional skills
- **GIVEN** a base agent with skills `["go-code", "go-arch"]`
- **WHEN** a variant agent extends base with skills `["go-db"]`
- **THEN** the merged agent SHALL have skills `["go-code", "go-arch", "go-db"]`

#### Scenario: Agent extends base with overlapping skills
- **GIVEN** a base agent with skills `["go-code", "go-arch"]`
- **WHEN** a variant agent extends base with skills `["go-arch", "go-db"]`
- **THEN** the merged agent SHALL have skills `["go-code", "go-arch", "go-db"]` without duplication

#### Scenario: Agent extends base with additional tools
- **GIVEN** a base agent with tools `["Read", "Write"]`
- **WHEN** a variant agent extends base with tools `["Edit", "Bash"]`
- **THEN** the merged agent SHALL have tools `["Read", "Write", "Edit", "Bash"]`

#### Scenario: Agent extends base with additional tool presets
- **GIVEN** a base agent with toolPresets `["readonly"]`
- **WHEN** a variant agent extends base with toolPresets `["editing"]`
- **THEN** the merged agent SHALL have toolPresets `["readonly", "editing"]`

#### Scenario: Non-slice fields still use replacement
- **GIVEN** a base agent with name `"base"`, description `"Base agent"`, model `"main"`
- **WHEN** a variant agent extends base with name `"variant"`, model `"fast"`
- **THEN** the merged agent SHALL have name `"variant"`, description `"Base agent"`, model `"fast"`

### Requirement: Backward compatibility
The system SHALL ensure existing agent definitions continue to work after the inheritance change.

#### Scenario: Existing agent with extends field loads successfully
- **GIVEN** an existing agent YAML with `extends: planner`
- **WHEN** the agent is loaded via `loadAgents()`
- **THEN** no error SHALL be returned
- **AND** the agent SHALL have combined tools from base and variant

#### Scenario: Agent without extends field unchanged
- **GIVEN** an agent YAML without `extends` field
- **WHEN** the agent is loaded via `loadAgents()`
- **THEN** the agent SHALL have exactly the tools defined in its YAML
