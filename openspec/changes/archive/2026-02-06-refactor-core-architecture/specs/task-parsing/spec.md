# Task-Parsing Specification

## ADDED Requirements

### Requirement: Parse Single Task from Markdown Content

The parser SHALL extract a single task definition from markdown content.

**Level**: MUST

#### Scenario: Parse valid task with all fields
**Given** markdown content contains a valid task definition
**When** parsing the markdown content
**Then** the parser SHALL return a task with ID, name, status, and metadata populated
**And** the task ID SHALL match the identifier in the markdown

#### Scenario: Parse task with minimal required fields
**Given** markdown content contains a task with only ID and name
**When** parsing the markdown content
**Then** the parser SHALL return a task with ID and name populated
**And** optional fields SHALL have sensible defaults

#### Scenario: Fail on invalid markdown format
**Given** markdown content does not contain valid task syntax
**When** parsing the markdown content
**Then** the parser SHALL return an error
**And** the error message SHALL indicate the validation failure

### Requirement: Parse Multiple Tasks from Single Markdown File

The parser SHALL extract multiple task definitions from a single markdown file.

**Level**: MUST

#### Scenario: Parse file with two tasks
**Given** markdown content contains two task definitions separated by headers
**When** parsing the markdown content
**Then** the parser SHALL return a slice containing two tasks
**And** each task SHALL have correct ID and metadata

#### Scenario: Parse file with no tasks
**Given** markdown content contains no task definitions
**When** parsing the markdown content
**Then** the parser SHALL return an empty slice
**And** no error SHALL be returned

#### Scenario: Fail on duplicate task IDs
**Given** markdown content contains two tasks with the same ID
**When** parsing the markdown content
**Then** the parser SHALL return an error
**And** the error SHALL indicate duplicate task ID

### Requirement: Extract Task Metadata from Markdown Front Matter

The parser SHALL extract structured metadata from task markdown front matter.

**Level**: MUST

#### Scenario: Parse task with metadata block
**Given** markdown content contains task with YAML front matter metadata
**When** parsing the markdown content
**Then** the parser SHALL populate task metadata fields from the front matter
**And** metadata SHALL include priority, assignee, and deadline if present

#### Scenario: Parse task without metadata
**Given** markdown content contains task without front matter
**When** parsing the markdown content
**Then** the parser SHALL return a task with empty metadata
**And** no error SHALL be returned

#### Scenario: Fail on invalid YAML metadata
**Given** markdown content contains task with malformed YAML front matter
**When** parsing the markdown content
**Then** the parser SHALL return an error
**And** the error SHALL indicate YAML parsing failure

### Requirement: Validate Task ID Format

The parser SHALL validate that task IDs follow the required format.

**Level**: MUST

#### Scenario: Accept valid task ID
**Given** task ID follows kebab-case format with alphanumeric characters and hyphens
**When** validating the task ID
**Then** the parser SHALL accept the ID as valid
**And** the task SHALL be created successfully

#### Scenario: Reject empty task ID
**Given** task ID is empty or not provided
**When** validating the task ID
**Then** the parser SHALL return an error
**And** the error SHALL indicate task ID is required

#### Scenario: Reject invalid task ID format
**Given** task ID contains spaces or special characters
**When** validating the task ID
**Then** the parser SHALL return an error
**And** the error SHALL indicate invalid ID format

### Requirement: Parse Task Status

The parser SHALL parse and normalize task status values.

**Level**: MUST

#### Scenario: Parse known task status
**Given** markdown contains task status "todo", "in-progress", or "done"
**When** parsing the markdown content
**Then** the parser SHALL set the task status to the corresponding enum value
**And** the status SHALL be properly normalized

#### Scenario: Default status when not provided
**Given** markdown does not contain task status
**When** parsing the markdown content
**Then** the parser SHALL default the task status to "todo"
**And** no error SHALL be returned

#### Scenario: Fail on unknown task status
**Given** markdown contains unrecognized task status value
**When** parsing the markdown content
**Then** the parser SHALL return an error
**And** the error SHALL indicate invalid status value

### Requirement: Extract Task Dependencies

The parser SHALL extract and validate task dependency references.

**Level**: MUST

#### Scenario: Parse task with valid dependencies
**Given** markdown contains task with dependency list referencing other task IDs
**When** parsing the markdown content
**Then** the parser SHALL populate the task dependencies slice
**And** each dependency SHALL reference a valid task ID format

#### Scenario: Parse task with no dependencies
**Given** markdown contains task without dependency list
**When** parsing the markdown content
**Then** the parser SHALL return a task with empty dependencies slice
**And** no error SHALL be returned

#### Scenario: Validate dependency format
**Given** markdown contains dependency with invalid ID format
**When** parsing the markdown content
**Then** the parser SHALL return an error
**And** the error SHALL indicate invalid dependency format
