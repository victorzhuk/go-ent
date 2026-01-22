# Spec: Enhanced Validation Error Messages

## ADDED Requirements

### REQ-VAL-001: Validation errors include actionable suggestions

Validation errors must provide clear guidance on how to fix the issue.

#### Scenario: Missing name field
**WHEN** skill is missing the `name` field in frontmatter
**THEN** error message includes suggestion to add name field
**AND** error message includes example of correct format

#### Scenario: Invalid name format
**WHEN** skill name contains uppercase letters or spaces
**THEN** error message includes suggestion to use lowercase and hyphens
**AND** error message includes example of valid name format

#### Scenario: Missing description field
**WHEN** skill is missing the `description` field in frontmatter
**THEN** error message includes suggestion to add description
**AND** error message includes example showing what to include

### REQ-VAL-002: Validation errors include code examples

Validation errors for XML structure issues must include correct XML examples.

#### Scenario: Missing examples section
**WHEN** skill is missing `<examples>` section
**THEN** error message includes suggestion to add examples section
**AND** error message includes complete XML example with proper structure

#### Scenario: Missing role section
**WHEN** skill is missing `<role>` section
**THEN** error message includes suggestion to add role section
**AND** error message includes example role definition

#### Scenario: Missing instructions section
**WHEN** skill is missing `<instructions>` section
**THEN** error message includes suggestion to add instructions
**AND** error message includes example instructions structure

### REQ-VAL-003: CLI displays suggestions prominently

The CLI must format error messages to make suggestions easily visible.

#### Scenario: Single error with suggestion
**WHEN** validation finds one error with suggestion
**THEN** CLI displays error message
**AND** CLI displays suggestion with visual indicator (💡)
**AND** CLI displays example with proper indentation

#### Scenario: Multiple errors with suggestions
**WHEN** validation finds multiple errors with suggestions
**THEN** CLI displays each error separately
**AND** CLI displays each suggestion clearly
**AND** suggestions are visually distinct from error messages

#### Scenario: Error without suggestion (backward compatibility)
**WHEN** validation error has empty suggestion field
**THEN** CLI displays error message normally
**AND** CLI does not display empty suggestion section

### REQ-VAL-004: All validation rules provide suggestions

Every validation rule must provide helpful suggestions.

#### Scenario: Check SK001-SK009 coverage
**WHEN** running validation on skills with various issues
**THEN** SK001 (name-required) provides suggestion and example
**AND** SK002 (name-format) provides suggestion and example
**AND** SK003 (description-required) provides suggestion and example
**AND** SK004 (examples-section) provides suggestion and example
**AND** SK005 (role-section) provides suggestion and example
**AND** SK006 (instructions-section) provides suggestion and example
**AND** SK007 (constraints-section) provides suggestion and example
**AND** SK008 (output-format) provides suggestion and example
**AND** SK009 (edge-cases) provides suggestion and example

## MODIFIED Requirements

None - this is purely additive functionality.

## REMOVED Requirements

None - no existing functionality is removed.
