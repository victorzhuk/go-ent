# Spec Delta: Template System

## ADDED Requirements

### Requirement: Template Engine
The system SHALL use Go's `text/template` package for prompt template processing.

#### Scenario: Template Parsing
- **WHEN** the system loads a template file
- **THEN** it SHALL parse using `text/template`
- **AND** it SHALL validate template syntax at compile time

### Requirement: Template Inheritance
The system SHALL support template inheritance using `{{define}}`, `{{template}}`, and `{{block}}` directives.

#### Scenario: Base Template Extension
- **WHEN** an agent template extends a base template
- **THEN** it SHALL use `{{template "base" .}}` to inherit
- **AND** it SHALL override blocks with `{{define "blockname"}}`
- **AND** the final output SHALL be a composed template

#### Scenario: Multi-Level Inheritance
- **WHEN** a template inherits from a base that inherits from another
- **THEN** the system SHALL resolve the inheritance chain
- **AND** blocks SHALL override correctly at each level

### Requirement: Custom Template Functions
The system SHALL provide custom template functions for common operations.

#### Scenario: File Inclusion
- **WHEN** a template uses `{{include "path/to/file"}}`
- **THEN** the system SHALL read the file contents
- **AND** it SHALL insert the contents at that location

#### Scenario: Text Indentation
- **WHEN** a template uses `{{indent 4 .Content}}`
- **THEN** the system SHALL indent each line by 4 spaces

#### Scenario: Default Values
- **WHEN** a template uses `{{default "fallback" .Value}}`
- **THEN** it SHALL use `.Value` if present, otherwise "fallback"

### Requirement: Backward Compatibility
The system SHALL process `.md` files without template syntax as plain text.

#### Scenario: Plain Markdown Files
- **WHEN** a prompt file has `.md` extension
- **THEN** it SHALL be read as plain text
- **AND** no template processing SHALL be applied
- **AND** simple `{{include}}` directives SHALL still work

#### Scenario: Template File Processing
- **WHEN** a prompt file has `.tmpl` or `.md.tmpl` extension
- **THEN** it SHALL be processed as a Go template
- **AND** all template directives SHALL be evaluated

### Requirement: Compile-Time Validation
The system SHALL validate templates at compile time, not runtime.

#### Scenario: Syntax Error Detection
- **WHEN** a template has invalid syntax
- **THEN** the system SHALL detect the error during compilation
- **AND** it SHALL report the filename and line number
- **AND** it SHALL NOT proceed with generation

#### Scenario: Missing Template Reference
- **WHEN** a template references a non-existent template
- **THEN** the system SHALL detect the error during compilation
- **AND** it SHALL report which template is missing

## MODIFIED Requirements

### Requirement: Template Processing
~~The system SHALL process templates using custom string replacement for `{{include}}` directives.~~

The system SHALL process templates using Go's `text/template` engine with custom functions.

#### Scenario: Include Directive
- **WHEN** a template contains `{{include "file.md"}}`
- **THEN** it SHALL be processed as a custom template function
- **AND** the function SHALL read and return the file contents

### Requirement: Template Data Context
~~Templates SHALL be static text files without data context.~~

Templates SHALL receive structured data via context objects.

#### Scenario: Data-Driven Templates
- **WHEN** a template is executed
- **THEN** it SHALL receive a data context (e.g., `.Role`, `.Description`)
- **AND** it SHALL access fields using dot notation (e.g., `{{.Role}}`)

## REMOVED Requirements

None - this change enhances existing template system without removing features.
