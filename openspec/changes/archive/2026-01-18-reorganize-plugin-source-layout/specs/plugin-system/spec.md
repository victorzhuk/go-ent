# Spec Delta: Plugin System

## ADDED Requirements

### Requirement: Plugin Source Separation
The system SHALL organize plugin sources separately from platform-specific templates and build artifacts.

#### Scenario: Source Directory Organization
- **WHEN** a developer inspects the `plugins/` directory
- **THEN** they SHALL see three distinct subdirectories: `sources/`, `platforms/`, and NOT see `generated/` (gitignored)
- **AND** `sources/` SHALL contain platform-agnostic plugin definitions
- **AND** `platforms/` SHALL contain platform-specific templates and configurations

### Requirement: Build Output Directory
The system SHALL generate compiled plugin artifacts to a `dist/` directory at the project root.

#### Scenario: Plugin Compilation
- **WHEN** the system compiles plugins for a target platform
- **THEN** the output SHALL be written to `dist/{platform}/{plugin-name}/`
- **AND** the `dist/` directory SHALL be excluded from version control

#### Scenario: Multi-Platform Build
- **WHEN** plugins are compiled for both Claude and OpenCode
- **THEN** the system SHALL create `dist/claude/go-ent/` and `dist/opencode/go-ent/`
- **AND** each platform output SHALL contain platform-specific configurations

### Requirement: Platform Template Location
The system SHALL store platform-specific templates in `plugins/platforms/{platform}/`.

#### Scenario: Template Resolution
- **WHEN** the adapter needs to load a template for Claude
- **THEN** it SHALL read from `plugins/platforms/claude/templates/`
- **AND** it SHALL NOT read templates from `plugins/sources/`

## MODIFIED Requirements

### Requirement: Plugin Loading
~~The system SHALL scan `plugins/` directory for plugin definitions.~~

The system SHALL scan `plugins/sources/` directory for plugin definitions.

#### Scenario: Plugin Discovery
- **WHEN** the plugin manager initializes
- **THEN** it SHALL scan `plugins/sources/` subdirectories
- **AND** it SHALL load `plugin.yaml` manifest from each plugin
- **AND** it SHALL ignore `plugins/platforms/` and `plugins/generated/`

### Requirement: Adapter Path Resolution
~~The adapter SHALL resolve plugin files relative to `plugins/{plugin-name}/`.~~

The adapter SHALL resolve source files from `plugins/sources/{plugin-name}/` and templates from `plugins/platforms/{platform}/`.

#### Scenario: Claude Adapter Compilation
- **WHEN** Claude adapter compiles the `go-ent` plugin
- **THEN** it SHALL read sources from `plugins/sources/go-ent/`
- **AND** it SHALL read templates from `plugins/platforms/claude/templates/`
- **AND** it SHALL write output to `dist/claude/go-ent/`

#### Scenario: OpenCode Adapter Compilation
- **WHEN** OpenCode adapter compiles the `go-ent` plugin
- **THEN** it SHALL read sources from `plugins/sources/go-ent/`
- **AND** it SHALL read templates from `plugins/platforms/opencode/templates/`
- **AND** it SHALL write output to `dist/opencode/go-ent/`

## REMOVED Requirements

None - this change enhances existing plugin system organization.
