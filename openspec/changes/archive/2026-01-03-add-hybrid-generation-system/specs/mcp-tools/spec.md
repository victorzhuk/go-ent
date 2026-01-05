## ADDED Requirements

### Requirement: Generation Configuration
The system SHALL support optional `generation.yaml` configuration file in the `openspec/` directory.

#### Scenario: Parse generation config
- **WHEN** `generation.yaml` exists in `openspec/` directory
- **THEN** the system parses and applies configuration settings

#### Scenario: Use defaults when no config
- **WHEN** `generation.yaml` does not exist
- **THEN** the system uses built-in default settings

#### Scenario: Override archetype
- **WHEN** component specifies `archetype` in generation.yaml
- **THEN** that archetype is used instead of auto-detected one

### Requirement: Project Archetypes
The system SHALL provide built-in project archetypes with predefined template sets.

#### Scenario: Standard archetype
- **WHEN** archetype is `standard`
- **THEN** templates for web service with clean architecture are selected

#### Scenario: MCP archetype
- **WHEN** archetype is `mcp`
- **THEN** templates for MCP server plugin are selected

#### Scenario: Custom archetype
- **WHEN** user defines custom archetype in generation.yaml
- **THEN** the custom template set is used for generation

### Requirement: Archetype Discovery
The system SHALL provide a tool to list available project archetypes.

#### Scenario: List all archetypes
- **WHEN** `go_ent_list_archetypes` is called without filter
- **THEN** all built-in and custom archetypes are returned with metadata

#### Scenario: Filter archetypes
- **WHEN** `go_ent_list_archetypes` is called with type filter
- **THEN** only matching archetypes are returned

### Requirement: Spec Analysis
The system SHALL analyze spec files to identify patterns and components.

#### Scenario: Detect CRUD pattern
- **WHEN** spec contains create/read/update/delete scenarios
- **THEN** CRUD pattern is identified with high confidence

#### Scenario: Detect API pattern
- **WHEN** spec contains endpoint/request/response language
- **THEN** API pattern is identified

#### Scenario: Return analysis confidence
- **WHEN** spec is analyzed
- **THEN** confidence score (0.0-1.0) is included in results

### Requirement: Archetype Selection
The system SHALL map spec analysis to recommended archetypes.

#### Scenario: Auto-select archetype
- **WHEN** no archetype is specified in generation.yaml
- **THEN** archetype is selected based on spec analysis

#### Scenario: Explicit override
- **WHEN** archetype is specified in generation.yaml
- **THEN** explicit archetype takes precedence over auto-selection

### Requirement: Component Generation
The system SHALL generate component scaffolds from spec and templates.

#### Scenario: Generate from spec
- **WHEN** `go_ent_generate_component` is called with spec path
- **THEN** component scaffold is generated using selected templates

#### Scenario: Mark extension points
- **WHEN** component is generated
- **THEN** extension points are marked with `@generate:` comments

#### Scenario: Include prompt context
- **WHEN** extension points are generated
- **THEN** relevant spec requirements are included as context

### Requirement: Spec-Driven Generation
The system SHALL generate complete projects from spec files.

#### Scenario: Full project generation
- **WHEN** `go_ent_generate_from_spec` is called
- **THEN** all identified components are generated

#### Scenario: Component integration
- **WHEN** multiple components are generated
- **THEN** integration points between components are created

### Requirement: AI Prompt Templates
The system SHALL provide prompt templates for AI-assisted code generation.

#### Scenario: Load prompt template
- **WHEN** prompt template is requested by type
- **THEN** template is loaded from `prompts/` directory

#### Scenario: Substitute variables
- **WHEN** prompt template is populated
- **THEN** spec content and requirements are substituted

### Requirement: Extension Points
The system SHALL mark extension points in generated code for AI completion.

#### Scenario: Constructor extension
- **WHEN** `@generate:constructor` marker is present
- **THEN** AI can generate dependency injection code

#### Scenario: Methods extension
- **WHEN** `@generate:methods` marker is present
- **THEN** AI can generate business logic methods from spec
