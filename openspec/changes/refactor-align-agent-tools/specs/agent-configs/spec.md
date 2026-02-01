# Spec: Agent File Updates

## ADDED Requirements

### Requirement: Add missing tools to agent configurations
All agent files SHALL include the `list`, `todoread`, `todowrite`, and `skill` tools.

#### Scenario: Coder agent has all required tools
- **GIVEN** the coder agent configuration
- **WHEN** tools are listed
- **THEN** the agent SHALL have: `read`, `write`, `edit`, `bash`, `glob`, `grep`, `list`, `todoread`, `todowrite`, `skill`, and `mcp__plugin_serena_serena`

#### Scenario: Planner agent has all required tools
- **GIVEN** the planner agent configuration
- **WHEN** tools are listed
- **THEN** the agent SHALL have: `read`, `glob`, `grep`, `list`, `todoread`, `todowrite`, `skill`, and task management tools

#### Scenario: Debugger agent has all required tools
- **GIVEN** the debugger agent configuration
- **WHEN** tools are listed
- **THEN** the agent SHALL have editing tools plus `list`, `todoread`, `todowrite`, `skill`

#### Scenario: Researcher agent has web tools
- **GIVEN** the researcher agent configuration
- **WHEN** tools are listed
- **THEN** the agent SHALL have: `list`, `webfetch`, `websearch`, `todoread`, `todowrite`, `skill`

#### Scenario: Reviewer agent has required tools
- **GIVEN** the reviewer agent configuration
- **WHEN** tools are listed
- **THEN** the agent SHALL have: `read`, `glob`, `grep`, `list`, `todoread`, `skill`

### Requirement: Replace grep with rg patterns
All agent prompts SHALL use `rg` (ripgrep) instead of `grep` for search operations.

#### Scenario: Grep -rn replaced with rg -n
- **GIVEN** an agent prompt containing search examples
- **WHEN** the pattern `grep -rn "func New"` is found
- **THEN** it SHALL be replaced with `rg -n "func New"`

#### Scenario: Grep with file type filtering
- **GIVEN** an agent prompt with file-specific searches
- **WHEN** the pattern `grep -r "type.*Repository"` is found
- **THEN** it SHALL be replaced with `rg -tgo "type.*Repository"`

#### Scenario: All grep occurrences replaced
- **GIVEN** all agent prompt files
- **WHEN** searching for `grep` commands
- **THEN** no `grep` examples SHALL remain (except in documentation explaining the replacement)

### Requirement: Replace find with fd patterns
Planner agents SHALL use `fd` instead of `find` for file discovery.

#### Scenario: Find replaced with fd
- **GIVEN** a planner agent prompt with file finding examples
- **WHEN** the pattern `find . -name "*.go"` is found
- **THEN** it SHALL be replaced with `fd -e go`

#### Scenario: Find with type filtering
- **GIVEN** a planner agent prompt
- **WHEN** the pattern `find . -type f -name "*.yaml"` is found
- **THEN** it SHALL be replaced with `fd -e yaml`

### Requirement: Add Optimal Tooling section
All agent files SHALL include an "Optimal Tooling" section documenting efficient command replacements.

#### Scenario: Optimal Tooling section present
- **GIVEN** any agent configuration file
- **WHEN** the file is read
- **THEN** it SHALL contain a section titled "## Optimal Tooling"

#### Scenario: Tooling table format
- **GIVEN** the Optimal Tooling section
- **THEN** it SHALL contain a table with columns: "Instead of", "Use", "Reason"

#### Scenario: Grep to rg mapping documented
- **GIVEN** the Optimal Tooling section
- **THEN** it SHALL document: `grep -rn` → `rg -n` (10x faster, respects .gitignore)

#### Scenario: File type filtering documented
- **GIVEN** the Optimal Tooling section
- **THEN** it SHALL document: `grep -r "pattern"` → `rg -tgo "pattern"`

#### Scenario: Find to fd mapping documented
- **GIVEN** the Optimal Tooling section
- **THEN** it SHALL document: `find . -name` → `fd` (5x faster)

#### Scenario: Cat grep pattern documented
- **GIVEN** the Optimal Tooling section
- **THEN** it SHALL document: `cat file | grep` → `rg -n pattern file`

### Requirement: Add Context Gathering workflow phase
Execution agents SHALL include a "Context Gathering" phase in their workflow.

#### Scenario: Context Gathering section present
- **GIVEN** an execution role agent (coder, debugger, tester)
- **WHEN** the agent prompt is read
- **THEN** it SHALL contain a "### 1. Context Gathering" or similar section

#### Scenario: TodoRead in context gathering
- **GIVEN** the Context Gathering section
- **THEN** it SHALL document using `todoread` to check current task state

#### Scenario: Skill loading in context gathering
- **GIVEN** the Context Gathering section
- **THEN** it SHALL document using `skill {skill-name}` to load relevant skills

#### Scenario: Directory exploration in context gathering
- **GIVEN** the Context Gathering section
- **THEN** it SHALL document using `list internal` and `glob "**/*.go"` for exploration

#### Scenario: Search with rg in context gathering
- **GIVEN** the Context Gathering section
- **THEN** it SHALL document using `rg -tgo "pattern" internal/` for searching

### Requirement: Agent file validation
All updated agent files SHALL pass YAML syntax validation.

#### Scenario: Valid YAML syntax
- **GIVEN** an updated agent file
- **WHEN** YAML parsing is attempted
- **THEN** no syntax errors SHALL be returned

#### Scenario: Valid tool references
- **GIVEN** an updated agent file
- **WHEN** tools are validated
- **THEN** all referenced tools SHALL be recognized by the system

#### Scenario: No duplicate tool entries
- **GIVEN** an updated agent file
- **WHEN** tools are listed
- **THEN** no tool SHALL appear more than once

## MODIFIED Requirements

### Requirement: Agent initialization includes updated agents
The `ent init` command SHALL generate agent files with all tool updates.

#### Scenario: Initialize creates updated coder agent
- **GIVEN** a fresh project
- **WHEN** `ent init --tool=claude` is executed
- **THEN** the generated coder agent SHALL include `list`, `todoread`, `todowrite`, `skill` tools

#### Scenario: Initialize creates updated planner agent
- **GIVEN** a fresh project
- **WHEN** `ent init --tool=claude` is executed
- **THEN** the generated planner agent SHALL include task management tools and `list`, `skill` tools

#### Scenario: Force flag regenerates with updates
- **GIVEN** an existing project with old agent files
- **WHEN** `ent init --tool=claude --force` is executed
- **THEN** all agent files SHALL be regenerated with updated tool configurations
