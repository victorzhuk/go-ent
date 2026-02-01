# Spec: Dynamic Tool Registry

## Purpose
Define requirements for a dynamic tool registry system that generates tool lists based on agent roles and categories.

## Requirements

### Requirement: ToolRegistry interface
The system SHALL provide a `ToolRegistry` type that supports dynamic tool list generation from metadata.

#### Scenario: Create tool registry from presets
- **GIVEN** tool presets are loaded from `tools.yaml`
- **WHEN** `NewToolRegistry(presets)` is called
- **THEN** a ToolRegistry SHALL be returned with all preset tools registered

#### Scenario: Registry stores tool metadata
- **GIVEN** a ToolRegistry with registered tools
- **WHEN** querying tool information
- **THEN** each tool SHALL have Name, Category, and Description metadata

### Requirement: Category-based tool lookup
The system SHALL support retrieving tools by their functional category.

#### Scenario: Get tools by category
- **GIVEN** a ToolRegistry with tools categorized as "read", "write", "search", "analysis"
- **WHEN** `GetToolsByCategory("read")` is called
- **THEN** all tools with category "read" SHALL be returned (e.g., `Read`, `Glob`, `Grep`)

#### Scenario: Get tools by search category
- **GIVEN** a ToolRegistry with search tools
- **WHEN** `GetToolsByCategory("search")` is called
- **THEN** search tools like `Grep`, `Glob`, `WebSearch` SHALL be returned

#### Scenario: Get tools by analysis category
- **GIVEN** a ToolRegistry with analysis tools
- **WHEN** `GetToolsByCategory("analysis")` is called
- **THEN** analysis tools like Serena MCP tools SHALL be returned

#### Scenario: Empty result for unknown category
- **GIVEN** a ToolRegistry
- **WHEN** `GetToolsByCategory("unknown")` is called
- **THEN** an empty slice SHALL be returned

### Requirement: Role-based tool assignment
The system SHALL support generating appropriate tool lists based on agent roles.

#### Scenario: Get tools for planning role
- **GIVEN** a ToolRegistry
- **WHEN** `GetToolsForRole("planning")` is called
- **THEN** tools SHALL include read and analysis categories (e.g., `Read`, `Glob`, `Grep`, Serena analysis tools)

#### Scenario: Get tools for execution role
- **GIVEN** a ToolRegistry
- **WHEN** `GetToolsForRole("execution")` is called
- **THEN** tools SHALL include read, write, and search categories (e.g., `Read`, `Write`, `Edit`, `Bash`)

#### Scenario: Get tools for research role
- **GIVEN** a ToolRegistry
- **WHEN** `GetToolsForRole("research")` is called
- **THEN** tools SHALL include read and search categories (e.g., `Read`, `Grep`, `Glob`, `WebSearch`, `WebFetch`)

#### Scenario: Get tools for validation role
- **GIVEN** a ToolRegistry
- **WHEN** `GetToolsForRole("validation")` is called
- **THEN** tools SHALL include read and analysis categories

#### Scenario: Default tools for unknown role
- **GIVEN** a ToolRegistry
- **WHEN** `GetToolsForRole("unknown")` is called
- **THEN** read category tools SHALL be returned as safe default

### Requirement: Tool registration
The system SHALL support registering tools with metadata dynamically.

#### Scenario: Register new tool
- **GIVEN** a ToolRegistry
- **WHEN** `RegisterTool(name, category, description)` is called
- **THEN** the tool SHALL be available in subsequent category queries

#### Scenario: Register duplicate tool ignored
- **GIVEN** a ToolRegistry with existing tool "Read"
- **WHEN** `RegisterTool("Read", ...)` is called again
- **THEN** the registration SHALL be ignored or update existing metadata
- **AND** no duplicate entries SHALL exist

### Requirement: Integration with agent loading
The system SHALL integrate ToolRegistry with `loadAgents()` for dynamic tool list generation.

#### Scenario: Generate tools from role during load
- **GIVEN** an agent with `role: execution` and no explicit tools
- **WHEN** the agent is loaded via `loadAgents()`
- **THEN** appropriate tools for execution role SHALL be generated dynamically

#### Scenario: Explicit tools override dynamic generation
- **GIVEN** an agent with explicit `tools: ["Read", "Write"]` and `role: execution`
- **WHEN** the agent is loaded
- **THEN** explicit tools SHALL take precedence over dynamic generation

### Requirement: Tool categories definition
The system SHALL define standard tool categories in the registry.

#### Scenario: Read category tools
- **GIVEN** the ToolRegistry
- **THEN** "read" category SHALL include: `Read`, `Glob`, `Grep`, `List`, `TodoRead`

#### Scenario: Write category tools
- **GIVEN** the ToolRegistry
- **THEN** "write" category SHALL include: `Write`, `Edit`, `Bash`, `TodoWrite`

#### Scenario: Search category tools
- **GIVEN** the ToolRegistry
- **THEN** "search" category SHALL include: `Grep`, `Glob`, `WebSearch`, `WebFetch`

#### Scenario: Analysis category tools
- **GIVEN** the ToolRegistry
- **THEN** "analysis" category SHALL include Serena MCP tools: `get_symbols_overview`, `find_symbol`, `search_for_pattern`, etc.

#### Scenario: Skill category tools
- **GIVEN** the ToolRegistry
- **THEN** "skill" category SHALL include: `Skill`
