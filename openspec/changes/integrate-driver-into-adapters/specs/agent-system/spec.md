# Spec Delta: Agent System

## ADDED Requirements

### Requirement: Base Agent Template
The system SHALL provide a `_base.md` template defining universal patterns for all agents.

#### Scenario: Agent Prompt Generation
- **WHEN** the system generates any agent prompt
- **THEN** it SHALL include or inherit from `_base.md`
- **AND** the base SHALL define common tooling, conventions, and behaviors

### Requirement: Driver/Orchestrator Template
The system SHALL provide a `_driver.md` template defining orchestration capabilities.

#### Scenario: Orchestrator Prompt Composition
- **WHEN** a driver agent prompt is generated
- **THEN** it SHALL include orchestration patterns from `_driver.md`
- **AND** it SHALL define task routing, delegation, and context management

### Requirement: Platform-Specific Driver Implementation
Each platform adapter SHALL implement driver/orchestration logic specific to that platform's capabilities.

#### Scenario: Claude Driver Behavior
- **WHEN** Claude adapter compiles the driver agent
- **THEN** it SHALL use Claude-specific delegation mechanisms
- **AND** it SHALL load driver logic from `plugins/platforms/claude/driver.go`

#### Scenario: OpenCode Driver Behavior
- **WHEN** OpenCode adapter compiles the driver agent
- **THEN** it SHALL use OpenCode-specific workflow patterns
- **AND** it SHALL load driver logic from `plugins/platforms/opencode/driver.go`

### Requirement: Single Prompt Source Location
The system SHALL maintain all agent prompts in `plugins/sources/{plugin}/agents/prompts/`.

#### Scenario: Prompt Discovery
- **WHEN** the adapter needs to load agent prompts
- **THEN** it SHALL read from `plugins/sources/go-ent/agents/prompts/`
- **AND** it SHALL NOT read from any other prompt directory
- **AND** the `/prompts/` legacy directory SHALL NOT exist

## MODIFIED Requirements

### Requirement: Agent Prompt Organization
~~Agent prompts MAY be stored in multiple locations (`/prompts/` and `plugins/*/agents/prompts/`).~~

Agent prompts SHALL be stored exclusively in `plugins/sources/{plugin}/agents/prompts/`.

#### Scenario: Prompt Consolidation
- **WHEN** a developer looks for agent prompts
- **THEN** they SHALL find all prompts in `plugins/sources/go-ent/agents/prompts/`
- **AND** there SHALL be a clear hierarchy: `_base.md`, `_driver.md`, `shared/`, `agents/`

### Requirement: Orchestration Pattern
~~Orchestration is defined independently from platform adapters.~~

Orchestration logic SHALL be implemented within platform adapters as platform-specific behavior.

#### Scenario: Platform-Specific Orchestration
- **WHEN** an orchestration workflow executes
- **THEN** the behavior SHALL match the target platform's capabilities
- **AND** the platform adapter SHALL determine routing and delegation strategies

## REMOVED Requirements

### Requirement: Legacy Prompts Directory
~~The system MAY maintain prompts in `/prompts/` for universal templates.~~

**Reason**: Duplication and confusion eliminated by consolidating into plugin structure.

**Migration**: All content from `/prompts/` integrated into `plugins/sources/go-ent/agents/prompts/`.
