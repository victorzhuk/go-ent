# Spec Delta: Marketplace

## ADDED Requirements

### Requirement: Plugin Dependencies
Plugin manifests SHALL support a `dependencies` field listing required plugins.

#### Scenario: Dependency Declaration
- **WHEN** a plugin manifest is created
- **THEN** it MAY include `dependencies: ["skills@go-ent", "hooks@go-ent"]`
- **AND** each dependency SHALL use format `{name}@{org}` or `{name}@{org}:{version}`

### Requirement: Dependency Resolution
The system SHALL resolve and install plugin dependencies recursively.

#### Scenario: Automatic Dependency Installation
- **WHEN** a user installs `agents@go-ent` which depends on `skills@go-ent`
- **THEN** the system SHALL automatically install `skills@go-ent` first
- **AND** it SHALL install in topological order

#### Scenario: Transitive Dependencies
- **WHEN** plugin A depends on B, and B depends on C
- **THEN** the system SHALL install C, then B, then A
- **AND** the install order SHALL respect the dependency graph

### Requirement: Circular Dependency Detection
The system SHALL detect and reject circular dependencies.

#### Scenario: Circular Dependency Rejection
- **WHEN** plugin A depends on B, and B depends on A
- **THEN** the system SHALL detect the circular dependency
- **AND** it SHALL reject the installation with an error message

### Requirement: Version Constraints
Plugin dependencies SHALL support semantic version constraints.

#### Scenario: Version Range Specification
- **WHEN** a manifest specifies `dependencies: ["skills@go-ent:^1.0.0"]`
- **THEN** the system SHALL accept versions `>=1.0.0 and <2.0.0`
- **AND** it SHALL reject incompatible versions

### Requirement: Type-Based Plugin Packages
The go-ent plugin SHALL be split into type-based packages.

#### Scenario: Package Organization
- **WHEN** browsing marketplace plugins
- **THEN** users SHALL find `agents@go-ent`, `skills@go-ent`, `commands@go-ent`, `hooks@go-ent`
- **AND** users SHALL find meta-package `go-ent` that depends on all

#### Scenario: Selective Installation
- **WHEN** a user installs `skills@go-ent` only
- **THEN** only skills SHALL be installed
- **AND** agents, commands, and hooks SHALL NOT be installed

### Requirement: Cross-Plugin References
Agents, commands, and skills SHALL reference components from other packages using fully qualified names.

#### Scenario: Fully Qualified Skill References
- **WHEN** an agent references a skill
- **THEN** it SHALL use format `skills@go-ent:go-code`
- **AND** the system SHALL resolve the reference across packages

#### Scenario: Fully Qualified Agent References
- **WHEN** a command references an agent
- **THEN** it SHALL use format `agents@go-ent:architect`
- **AND** the system SHALL resolve the reference across packages

## MODIFIED Requirements

### Requirement: Plugin Manifest Structure
~~Plugin manifests contain name, version, description, and component references.~~

Plugin manifests SHALL contain name, version, description, component references, and optional dependencies.

#### Scenario: Manifest with Dependencies
- **WHEN** a plugin manifest is parsed
- **THEN** it SHALL include optional `dependencies` field
- **AND** dependencies SHALL be validated for correct format

### Requirement: Plugin Installation
~~The system installs plugins independently without dependency management.~~

The system SHALL install plugins with automatic dependency resolution.

#### Scenario: Dependency-Aware Installation
- **WHEN** installing a plugin with dependencies
- **THEN** the system SHALL check if dependencies are installed
- **AND** it SHALL install missing dependencies automatically
- **AND** it SHALL use correct topological order

## REMOVED Requirements

### Requirement: Monolithic Plugin Packaging
~~Plugins MAY bundle multiple component types (agents, skills, commands, hooks) in a single package.~~

**Reason**: Type-based splitting enables fine-grained installation and better marketplace organization.

**Migration**: Monolithic `go-ent` replaced by meta-package that depends on type-based packages.

## RENAMED Requirements

- FROM: `### Requirement: Plugin ID Format`
- TO: `### Requirement: Plugin ID and Scope Format`

The system SHALL support plugin IDs with optional organization scope: `{name}@{org}` or just `{name}`.
