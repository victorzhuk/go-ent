# Spec: Planning Preset Addition

## Purpose
Define requirements for automatically assigning the planning tool preset to agents with planning role.

## Requirements

### Requirement: Auto-assign planning preset to planning role agents
The system SHALL automatically add the `planning` tool preset to agents with `role: planning` during `expandToolPresets()`.

#### Scenario: Planning agent without planning preset
- **GIVEN** an agent with `role: planning` and toolPresets `["readonly", "serena-analysis"]`
- **WHEN** `expandToolPresets()` is called
- **THEN** the agent SHALL have toolPresets `["readonly", "serena-analysis", "planning"]`

#### Scenario: Planning agent already has planning preset
- **GIVEN** an agent with `role: planning` and toolPresets `["planning", "readonly"]`
- **WHEN** `expandToolPresets()` is called
- **THEN** the agent SHALL still have toolPresets `["planning", "readonly"]` without duplication

#### Scenario: Non-planning agent unchanged
- **GIVEN** an agent with `role: execution` and toolPresets `["editing"]`
- **WHEN** `expandToolPresets()` is called
- **THEN** the agent SHALL have toolPresets `["editing"]` without planning preset added

#### Scenario: Agent without role unchanged
- **GIVEN** an agent without `role` field and toolPresets `["readonly"]`
- **WHEN** `expandToolPresets()` is called
- **THEN** the agent SHALL have toolPresets `["readonly"]` without planning preset added

### Requirement: Planning preset tools
The system SHALL expand the `planning` preset to include task management and analysis tools.

#### Scenario: Planning preset expansion
- **GIVEN** the `planning` preset is defined in `tools.yaml`
- **WHEN** an agent with `planning` preset is processed
- **THEN** the agent SHALL receive tools: `Read`, `Glob`, `Grep`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, and Serena analysis tools

#### Scenario: Planning agents receive task tools
- **GIVEN** a planning role agent
- **WHEN** the agent configuration is generated
- **THEN** the agent SHALL have access to `TaskCreate`, `TaskUpdate`, `TaskList`, and `TaskGet` tools

### Requirement: Planning preset disallowed
The system SHALL respect `disallowedToolPresets` to exclude planning tools when explicitly configured.

#### Scenario: Planning agent with disallowed planning preset
- **GIVEN** an agent with `role: planning` and `disallowedToolPresets: ["planning"]`
- **WHEN** `expandToolPresets()` is called
- **THEN** the planning preset SHALL NOT be added to the agent
- **AND** no planning preset tools SHALL be in the final tool list

### Requirement: Planning preset validation
The system SHALL validate the `planning` preset is a recognized preset type.

#### Scenario: Valid planning preset reference
- **GIVEN** an agent with `toolPresets: ["planning"]`
- **WHEN** `validateAgent()` is called
- **THEN** no error SHALL be returned for the planning preset

### Requirement: Tool preset validation
The system SHALL recognize `planning` as a valid tool preset.

#### Scenario: Validate planning preset in agent config
- **GIVEN** an agent YAML with `toolPresets: ["planning"]`
- **WHEN** the agent is validated
- **THEN** validation SHALL pass without "unknown tool preset" error
