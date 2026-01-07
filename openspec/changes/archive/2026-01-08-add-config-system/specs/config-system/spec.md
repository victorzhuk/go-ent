# Capability: Configuration System

## Overview

Project-level configuration system for agent preferences, runtime selection, budget limits, and model mappings via `.go-ent/config.yaml`.

---

## ADDED Requirements

### Requirement: Configuration File Location

The system SHALL load configuration from `.go-ent/config.yaml` in the project root.

**Level**: MUST

#### Scenario: Standard Location
**Given** a project root directory
**When** loading configuration
**Then** the system SHALL attempt to read `.go-ent/config.yaml`

#### Scenario: Missing Configuration
**Given** no `.go-ent/config.yaml` exists
**When** loading configuration
**Then** the system SHALL use default values without error

---

### Requirement: Agent Role Configuration

The system SHALL allow configuration of model and skills per agent role.

**Level**: MUST

#### Scenario: Role Model Assignment
**Given** a config file with agents.roles.architect.model = "opus"
**When** loading configuration
**Then** the architect role SHALL use the opus model

#### Scenario: Role Skills Assignment
**Given** a config file with agents.roles.senior.skills = ["go-code", "go-db"]
**When** loading configuration
**Then** the senior role SHALL have go-code and go-db skills enabled

---

### Requirement: Runtime Preference Configuration

The system SHALL allow configuration of preferred and fallback runtimes.

**Level**: MUST

#### Scenario: Preferred Runtime
**Given** a config file with runtime.preferred = "opencode"
**When** selecting a runtime
**Then** the system SHALL prefer the opencode runtime

#### Scenario: Fallback Runtimes
**Given** a config file with runtime.fallback = ["claude-code", "cli"]
**When** the preferred runtime is unavailable
**Then** the system SHALL try fallback runtimes in order

---

### Requirement: Budget Limit Configuration

The system SHALL enforce budget limits at daily, monthly, and per-task levels.

**Level**: MUST

#### Scenario: Daily Budget Limit
**Given** a config file with budget.daily = 10.0
**When** tracking daily spending
**Then** the system SHALL enforce a $10 daily limit

#### Scenario: Per-Task Budget Limit
**Given** a config file with budget.per_task = 1.0
**When** executing a task
**Then** the system SHALL enforce a $1 task limit

---

### Requirement: Model Name Mapping

The system SHALL map friendly model names to actual model IDs.

**Level**: MUST

#### Scenario: Model Alias Resolution
**Given** a config file with models.opus = "claude-opus-4-5-20251101"
**When** resolving the "opus" alias
**Then** the system SHALL return "claude-opus-4-5-20251101"

---

### Requirement: Environment Variable Override

The system SHALL support environment variable overrides for configuration values.

**Level**: SHOULD

#### Scenario: Budget Override
**Given** an environment variable `GOENT_BUDGET_DAILY=5.0`
**When** loading configuration
**Then** the daily budget SHALL be 5.0 regardless of config file

---

### Requirement: Configuration Validation

The system SHALL validate configuration on load and return clear errors for invalid values.

**Level**: MUST

#### Scenario: Invalid Agent Role
**Given** a config file with an unknown agent role
**When** validating configuration
**Then** the system SHALL return a validation error

#### Scenario: Negative Budget
**Given** a config file with budget.daily = -10.0
**When** validating configuration
**Then** the system SHALL return a validation error
