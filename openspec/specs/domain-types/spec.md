# Domain-Types Specification

### Requirement: Agent Role Classification

The system SHALL define six agent roles with distinct specializations.

**Level**: MUST

#### Scenario: Product Agent Role
**Given** the system needs to understand user requirements
**When** classifying agent specialization
**Then** the `Product` role SHALL be available for requirements analysis and product decisions

#### Scenario: Architect Agent Role
**Given** the system needs technical design decisions
**When** classifying agent specialization
**Then** the `Architect` role SHALL be available for system design and architecture

#### Scenario: Senior Developer Agent Role
**Given** the system needs complex implementation
**When** classifying agent specialization
**Then** the `Senior` role SHALL be available for advanced coding and debugging

#### Scenario: Developer Agent Role
**Given** the system needs standard implementation
**When** classifying agent specialization
**Then** the `Developer` role SHALL be available for routine coding and testing

#### Scenario: Reviewer Agent Role
**Given** the system needs code quality enforcement
**When** classifying agent specialization
**Then** the `Reviewer` role SHALL be available for code review and standards validation

#### Scenario: Ops Agent Role
**Given** the system needs deployment and production support
**When** classifying agent specialization
**Then** the `Ops` role SHALL be available for infrastructure and monitoring

---

### Requirement: Runtime Environment Classification

The system SHALL support three runtime environments for agent execution.

**Level**: MUST

#### Scenario: Claude Code Runtime
**Given** the system integrates with Claude Code MCP
**When** selecting execution runtime
**Then** the `ClaudeCode` runtime SHALL be available

#### Scenario: OpenCode Runtime
**Given** the system integrates with OpenCode native API
**When** selecting execution runtime
**Then** the `OpenCode` runtime SHALL be available

#### Scenario: CLI Runtime
**Given** the system provides standalone CLI execution
**When** selecting execution runtime
**Then** the `CLI` runtime SHALL be available

---

### Requirement: Spec Action Taxonomy

The system SHALL classify all operations into a five-phase action taxonomy.

**Level**: MUST

#### Scenario: Discovery Phase Actions
**Given** the system performs codebase exploration
**When** classifying the operation
**Then** actions `research`, `analyze`, `retrofit` SHALL belong to the Discovery phase

#### Scenario: Planning Phase Actions
**Given** the system performs design and planning
**When** classifying the operation
**Then** actions `proposal`, `plan`, `design`, `split` SHALL belong to the Planning phase

#### Scenario: Execution Phase Actions
**Given** the system performs code generation
**When** classifying the operation
**Then** actions `implement`, `execute`, `scaffold` SHALL belong to the Execution phase

#### Scenario: Validation Phase Actions
**Given** the system performs quality checks
**When** classifying the operation
**Then** actions `review`, `verify`, `debug`, `lint` SHALL belong to the Validation phase

#### Scenario: Lifecycle Phase Actions
**Given** the system manages change lifecycle
**When** classifying the operation
**Then** actions `approve`, `archive`, `status` SHALL belong to the Lifecycle phase

---

### Requirement: Execution Strategy Classification

The system SHALL support three execution strategies for task coordination.

**Level**: MUST

#### Scenario: Single Agent Execution
**Given** a task can be completed by one agent
**When** selecting execution strategy
**Then** the `Single` strategy SHALL execute one agent sequentially

#### Scenario: Multi-Agent Conversation
**Given** a task requires agent collaboration
**When** selecting execution strategy
**Then** the `Multi` strategy SHALL enable agent handoff and conversation

#### Scenario: Parallel Agent Execution
**Given** a task has independent subtasks
**When** selecting execution strategy
**Then** the `Parallel` strategy SHALL execute multiple agents concurrently

---

### Requirement: Skill Interface Contract

The system SHALL define a Skill interface for reusable agent capabilities.

**Level**: MUST

#### Scenario: Skill Identity
**Given** a skill is registered in the system
**When** querying the skill
**Then** it SHALL provide a unique name and human-readable description

#### Scenario: Skill Applicability
**Given** a skill receives a context
**When** evaluating applicability
**Then** it SHALL indicate whether it can handle the context

#### Scenario: Skill Execution
**Given** a skill is applicable to a request
**When** executing the skill
**Then** it SHALL return a result or error

---

### Requirement: Domain Error Classification

The system SHALL provide typed errors for domain-specific failures.

**Level**: MUST

#### Scenario: Agent Not Found Error
**Given** the system attempts to resolve an unknown agent
**When** the agent does not exist
**Then** an `ErrAgentNotFound` error SHALL be returned

#### Scenario: Runtime Unavailable Error
**Given** the system attempts to use a runtime
**When** the runtime is not available
**Then** an `ErrRuntimeUnavailable` error SHALL be returned

#### Scenario: Invalid Action Error
**Given** the system receives an action
**When** the action is not in the taxonomy
**Then** an `ErrInvalidAction` error SHALL be returned

---

### Requirement: Zero External Dependencies

The domain package SHALL have zero external dependencies.

**Level**: MUST

#### Scenario: Pure Domain Logic
**Given** the domain package is being built
**When** analyzing dependencies
**Then** only standard library imports SHALL be present

#### Scenario: No Framework Coupling
**Given** the domain types are defined
**When** inspecting struct tags
**Then** no framework-specific tags (like `gorm`, `json`) SHALL be used on domain entities
