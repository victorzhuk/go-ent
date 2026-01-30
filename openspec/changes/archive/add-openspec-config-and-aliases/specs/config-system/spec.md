# Specification: OpenSpec Schema

## ADDED Requirements

### Requirement: Schema Definition
The project SHALL have `openspec/schemas/go-ent/schema.yaml`.

#### Scenario: Schema exists
- **WHEN** checking `openspec/schemas/go-ent/schema.yaml`
- **THEN** file exists
- **AND** defines proposal, specs, design, tasks artifacts
- **AND** defines artifact dependencies

### Requirement: Templates
The project SHALL have Go-specific templates.

#### Scenario: Templates exist
- **WHEN** checking `openspec/schemas/go-ent/templates/`
- **THEN** proposal.md exists
- **AND** spec.md exists
- **AND** design.md exists
- **AND** tasks.md exists

### Requirement: ent init Command
The project SHALL have `ent init` command.

#### Scenario: ent init works
- **WHEN** running `ent init`
- **THEN** `openspec/config.yaml` is created
- **AND** it contains schema: go-ent
- **AND** it contains Go module info from go.mod

## MODIFIED Specifications

None.

## REMOVED Specifications

None.
