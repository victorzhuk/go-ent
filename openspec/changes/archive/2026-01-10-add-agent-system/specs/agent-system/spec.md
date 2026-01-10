## ADDED Requirements

### Requirement: Agent Selection

The system SHALL automatically select the optimal agent based on task complexity.

#### Scenario: Trivial task routing
- **WHEN** task is analyzed and classified as Trivial
- **THEN** Haiku model with minimal scaffolding is selected
- **AND** execution proceeds without delegation

#### Scenario: Simple task routing
- **WHEN** task is classified as Simple (basic implementation, known patterns)
- **THEN** Dev agent with Haiku is selected
- **AND** relevant skills are attached

#### Scenario: Moderate task routing
- **WHEN** task requires moderate complexity (multi-file, integration)
- **THEN** Dev agent with Sonnet is selected
- **AND** multiple skills may be combined

#### Scenario: Complex task routing
- **WHEN** task is architecturally complex
- **THEN** Architect agent with Opus is selected for design
- **AND** delegation to Dev is prepared after design approval

#### Scenario: Architectural task routing
- **WHEN** task affects system architecture or has cross-cutting concerns
- **THEN** Architect agent with Opus is selected
- **AND** Lead agent may orchestrate if multiple agents needed

### Requirement: Task Complexity Analysis

The system SHALL classify task complexity using pattern matching and heuristics.

#### Scenario: Keyword-based classification
- **WHEN** task description contains keywords like "architecture", "design system"
- **THEN** complexity level Architectural is assigned

#### Scenario: Scope-based classification
- **WHEN** task affects > 5 files or > 3 packages
- **THEN** complexity level is elevated to Moderate or Complex

#### Scenario: Pattern recognition
- **WHEN** task matches known simple patterns (add CRUD endpoint, fix typo)
- **THEN** complexity level Simple or Trivial is assigned

#### Scenario: Complexity thresholds
- **THEN** complexity levels are defined as:
  - Trivial: Single-file, < 10 lines changed
  - Simple: 1-2 files, known patterns
  - Moderate: 3-5 files, some integration
  - Complex: > 5 files or architectural decisions
  - Architectural: System-wide changes, design decisions

### Requirement: Agent Delegation

The system SHALL enforce delegation rules based on agent roles.

#### Scenario: Lead delegates to specialists
- **WHEN** Lead agent delegates to Dev, Tester, or Reviewer
- **THEN** delegation is allowed
- **AND** delegation chain is tracked

#### Scenario: Architect delegates design
- **WHEN** Architect completes design
- **THEN** Architect can delegate implementation to Dev
- **AND** design document is passed as context

#### Scenario: Invalid delegation blocked
- **WHEN** Dev attempts to delegate to Architect
- **THEN** delegation is rejected with error
- **AND** message suggests proper flow (escalate to Lead)

#### Scenario: Delegation chain tracking
- **WHEN** task is delegated multiple times
- **THEN** full chain is preserved (Lead → Architect → Dev)
- **AND** each agent's contribution is tracked

### Requirement: Skill Registry

The system SHALL load and match skills from markdown files.

#### Scenario: Load skills from directory
- **WHEN** system starts
- **THEN** skills are loaded from `plugins/go-ent/skills/*.md`
- **AND** skill metadata is parsed from frontmatter

#### Scenario: Skill auto-activation
- **WHEN** task context matches skill activation rules
- **THEN** relevant skills are automatically attached to agent
- **AND** example: "go-code" skill activates for `.go` file edits

#### Scenario: Context-based skill matching
- **WHEN** agent is selected for task
- **THEN** skills matching task context are retrieved
- **AND** example: API design task → go-api skill

#### Scenario: Multiple skill combination
- **WHEN** complex task requires multiple skills
- **THEN** compatible skills are combined
- **AND** example: go-code + go-test for TDD workflow

### Requirement: Agent Role Definitions

The system SHALL define clear capabilities and constraints for each agent role.

#### Scenario: Lead agent capabilities
- **THEN** Lead can: orchestrate, delegate, review progress
- **AND** Lead cannot: implement directly (delegates instead)

#### Scenario: Architect agent capabilities
- **THEN** Architect can: design systems, make architectural decisions, create specs
- **AND** Architect cannot: implement (delegates to Dev)

#### Scenario: Dev agent capabilities
- **THEN** Dev can: implement features, write code, integrate components
- **AND** Dev should: follow designs from Architect

#### Scenario: Tester agent capabilities
- **THEN** Tester can: write tests, run TDD cycles, verify implementations
- **AND** Tester uses lightweight model (Haiku) for efficiency

#### Scenario: Reviewer agent capabilities
- **THEN** Reviewer can: review code, check standards, identify issues
- **AND** Reviewer uses high-quality model (Opus) for thoroughness

### Requirement: Budget-Aware Selection

The system SHALL consider cost when selecting agents and models.

#### Scenario: Budget constraint enforcement
- **WHEN** budget is limited and task is Simple
- **THEN** cheaper model (Haiku) is preferred over Sonnet
- **AND** cost estimate is provided

#### Scenario: Quality vs cost trade-off
- **WHEN** task is Critical AND budget allows
- **THEN** higher quality model (Opus) is selected
- **AND** cost increase is justified in rationale

#### Scenario: Budget exceeded warning
- **WHEN** task execution would exceed budget
- **THEN** warning is shown before proceeding
- **AND** alternative cheaper approach is suggested

### Requirement: Agent Context Preparation

The system SHALL prepare appropriate context for each agent.

#### Scenario: Agent receives role-specific prompt
- **WHEN** agent is selected
- **THEN** role-specific system prompt is included
- **AND** example: Dev agent gets implementation guidelines

#### Scenario: Skills are injected into context
- **WHEN** relevant skills are identified
- **THEN** skill prompts are appended to agent context
- **AND** skills are clearly marked in prompt

#### Scenario: Task history is included
- **WHEN** task is part of delegation chain
- **THEN** previous agent outputs are included
- **AND** context from Architect design is available to Dev
