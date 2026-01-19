# skill-management Specification

## Purpose
TBD - created by archiving change add-skill-templates. Update Purpose after archive.
## Requirements
### Requirement: Generate skill from built-in template
The system SHALL generate a skill file at `plugins/go-ent/skills/<category>/<name>/SKILL.md` with all placeholders replaced and validates the generated skill when the user runs `go-ent skill new <name> --template <template-name>`.

#### Scenario: Successful skill generation from template
- **WHEN** user runs `go-ent skill new my-go-skill --template go-basic`
- **THEN** system generates skill at `plugins/go-ent/skills/go/my-go-skill/SKILL.md`
- **AND** all placeholders are replaced with provided values
- **AND** validation runs on generated skill
- **AND** success message shows validation results

### Requirement: Interactive wizard for skill creation
The system SHALL prompt the user interactively for template selection and skill metadata when `go-ent skill new <name>` is run without flags.

#### Scenario: Interactive skill creation
- **WHEN** user runs `go-ent skill new my-skill`
- **THEN** system prompts for template selection from list
- **AND** prompts for skill description (optional)
- **AND** prompts for version (default: 1.0.0)
- **AND** prompts for author (default: git user)
- **AND** prompts for tags (optional)
- **AND** prompts for template-specific fields from config.yaml

### Requirement: Auto-detect skill category
The system SHALL auto-detect the category from the skill name pattern `<category>-<name>` and place the skill in the appropriate directory.

#### Scenario: Category auto-detection from skill name
- **WHEN** user runs `go-ent skill new go-new-pattern`
- **THEN** system detects category as "go"
- **AND** places skill at `plugins/go-ent/skills/go/go-new-pattern/SKILL.md`

#### Scenario: Category auto-detection for TypeScript skill
- **WHEN** user runs `go-ent skill new typescript-util`
- **THEN** system detects category as "typescript"
- **AND** places skill at `plugins/go-ent/skills/typescript/typescript-util/SKILL.md`

#### Scenario: Fallback to manual category selection
- **WHEN** skill name does not match category pattern
- **THEN** system prompts for manual category selection

### Requirement: Validate generated skill
The system SHALL run validation and report errors, warnings, and quality score after skill generation completes.

#### Scenario: Successful generation with validation
- **WHEN** skill generation completes
- **THEN** system runs validation on generated skill
- **AND** displays number of errors and warnings
- **AND** displays quality score (0-100)
- **AND** suggests next steps if validation passes

### Requirement: List available templates
The system SHALL display all available templates with name, category, description, and source when user runs `go-ent skill list-templates`.

#### Scenario: List all templates
- **WHEN** user runs `go-ent skill list-templates`
- **THEN** system displays all built-in and custom templates
- **AND** shows template name, category, description, and source

### Requirement: Filter templates by category
The system SHALL display only templates matching the specified category when user runs `go-ent skill list-templates --category <category>`.

#### Scenario: Filter templates by category
- **WHEN** user runs `go-ent skill list-templates --category go`
- **THEN** system displays only templates with category "go"

### Requirement: Add custom template
The system SHALL validate the template structure and copy it to the templates directory when user runs `go-ent skill add-template <path>`.

#### Scenario: Add valid custom template
- **WHEN** user runs `go-ent skill add-template /path/to/template`
- **AND** template has valid structure (template.md + config.yaml)
- **AND** template passes validation
- **THEN** system copies template to templates directory
- **AND** reports success

#### Scenario: Reject invalid template
- **WHEN** user runs `go-ent skill add-template /path/to/template`
- **AND** template is missing required files or fails validation
- **THEN** system displays specific validation errors
- **AND** does not copy template

### Requirement: Show template details
The system SHALL display template metadata and preview when user runs `go-ent skill show-template <name>`.

#### Scenario: Show template details
- **WHEN** user runs `go-ent skill show-template go-basic`
- **THEN** system displays template metadata (name, category, description, version, author)
- **AND** shows first 20 lines of template.md
- **AND** lists config prompts

### Requirement: Placeholder replacement
The system SHALL replace all placeholders with provided values during skill generation.

#### Scenario: Replace standard placeholders
- **WHEN** template contains `${SKILL_NAME}`, `${SKILL_DESCRIPTION}`, `${SKILL_VERSION}`, `${SKILL_AUTHOR}`, `${SKILL_TAGS}`
- **AND** user provides values via prompts or flags
- **THEN** system replaces all placeholders with provided values

#### Scenario: Replace custom placeholders
- **WHEN** template contains custom placeholders defined in config.yaml
- **AND** user provides values for custom placeholders
- **THEN** system replaces all custom placeholders with provided values

### Requirement: Non-interactive mode
The system SHALL generate skill without interactive prompts when all required values are provided via flags.

#### Scenario: Non-interactive skill creation
- **WHEN** user runs `go-ent skill new my-skill --template go-basic --description "My skill" --version "1.0.0" --author "user" --tags "tag1,tag2"`
- **THEN** system generates skill without any prompts
- **AND** uses provided values and template defaults

### Requirement: Template validation
The system SHALL validate template structure and quality when templates are loaded or added.

#### Scenario: Validate template structure
- **WHEN** template is loaded or added
- **THEN** system validates template.md exists and is readable
- **AND** validates config.yaml exists and is readable
- **AND** validates config.yaml has required fields (name, category, description)

#### Scenario: Validate built-in template quality
- **WHEN** built-in template is generated for testing
- **THEN** system validates template passes skill validation
- **AND** verifies quality score >= 90

### Requirement: Error handling for invalid input
The system SHALL display clear error messages with line numbers and suggestions when validation fails during skill creation.

#### Scenario: Invalid input during skill creation
- **WHEN** user provides invalid input during skill creation
- **AND** validation fails
- **THEN** system displays clear error message explaining the issue
- **AND** shows line number for validation errors
- **AND** does not create skill file
- **AND** suggests corrective action

### Requirement: Prevent overwriting existing skills
The system SHALL prevent overwriting existing skills and display an error message when attempting to create a duplicate skill.

#### Scenario: Skill already exists
- **WHEN** user attempts to create skill with name that already exists
- **THEN** system displays error "Skill already exists at <path>"
- **AND** suggests using different name or deleting existing skill
- **AND** does not overwrite existing skill

### Requirement: Built-in templates quality standards
The system SHALL ensure all built-in templates meet quality standards.

#### Scenario: Built-in template validation
- **WHEN** built-in template is generated
- **THEN** template must pass all validation rules in strict mode
- **AND** must have quality score >= 90
- **AND** must include at least 2 examples
- **AND** must handle at least 3 edge cases
- **AND** must follow v2 skill format with all XML sections

### Requirement: Template config schema
The system SHALL support the following config.yaml schema:

```yaml
name: string              # Template identifier (required)
category: string          # Template category (required)
description: string       # Template description (required)
author: string           # Template author (optional, default: go-ent)
version: string          # Template version (optional, default: 1.0.0)
prompts:                # Array of prompts (optional)
  - key: string         # Placeholder name (required)
    prompt: string       # Prompt text to display (required)
    default: string     # Default value (optional)
    required: boolean   # Whether input is required (optional, default: false)
```

#### Scenario: Parse template config with all fields
- **WHEN** system parses template config.yaml
- **THEN** system successfully reads name, category, description fields
- **AND** optionally reads author, version, and prompts array
- **AND** validates all required fields are present

