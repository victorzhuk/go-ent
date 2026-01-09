# MCP Tools - Delta Spec

## ADDED Requirements

### Requirement: Tool Discovery via Search

The system SHALL provide a `tool_find` MCP tool that searches for tools by keyword query using TF-IDF relevance scoring.

#### Scenario: Search for registry tools
- **WHEN** agent calls `tool_find` with query "registry list tasks"
- **THEN** return tools matching keywords (e.g., `registry_list`, `registry_next`)
- **AND** rank results by TF-IDF relevance score
- **AND** indicate which tools are currently active

#### Scenario: Limit search results
- **WHEN** agent calls `tool_find` with limit parameter
- **THEN** return at most `limit` results
- **AND** return highest scoring tools first

---

### Requirement: Tool Metadata Retrieval

The system SHALL provide a `tool_describe` MCP tool that returns detailed metadata for a specific tool.

#### Scenario: Describe a tool
- **WHEN** agent calls `tool_describe` with tool name
- **THEN** return tool name, description, and input schema
- **AND** indicate active status
- **AND** include category and keywords if available

#### Scenario: Tool not found
- **WHEN** agent calls `tool_describe` with non-existent tool name
- **THEN** return error message
- **AND** suggest using `tool_find` to discover available tools

---

### Requirement: Dynamic Tool Loading

The system SHALL provide a `tool_load` MCP tool that activates tools dynamically.

#### Scenario: Load multiple tools
- **WHEN** agent calls `tool_load` with array of tool names
- **THEN** activate each tool by registering with MCP server
- **AND** mark tools as active in registry
- **AND** return confirmation with total active count

#### Scenario: Load already active tool
- **WHEN** agent calls `tool_load` with tool already active
- **THEN** skip registration (idempotent operation)
- **AND** do not return error

#### Scenario: Load non-existent tool
- **WHEN** agent calls `tool_load` with invalid tool name
- **THEN** return error indicating tool not found
- **AND** do not activate any tools

---

### Requirement: Active Tool Listing

The system SHALL provide a `tool_active` MCP tool that lists currently active tools.

#### Scenario: List active tools
- **WHEN** agent calls `tool_active`
- **THEN** return names and descriptions of all active tools
- **AND** include count of active tools

#### Scenario: No active tools
- **WHEN** agent calls `tool_active` with no tools loaded
- **THEN** return empty list
- **AND** message indicating no tools are active

---

### Requirement: TF-IDF Search Index

The system SHALL build a TF-IDF search index from tool metadata at startup.

#### Scenario: Index tool metadata
- **WHEN** server initializes
- **THEN** extract terms from tool names and descriptions
- **THEN** calculate term frequency (TF) for each document
- **AND** calculate inverse document frequency (IDF) across all tools
- **AND** store term-document matrix for search

#### Scenario: Search with TF-IDF scoring
- **WHEN** agent performs search
- **THEN** tokenize query into terms
- **AND** calculate TF-IDF score for each tool
- **AND** return tools ranked by score descending

---

### Requirement: Tool Name Shortening

The system SHALL rename all tools from `go_ent_*` prefix to shorter names.

#### Scenario: Tool name migration
- **WHEN** existing tool is renamed
- **THEN** update tool registration with new name
- **AND** update all documentation references
- **AND** mark as breaking change in CHANGELOG

#### Scenario: Backward compatibility
- **WHEN** client uses old `go_ent_*` tool name
- **THEN** return error indicating tool not found
- **AND** suggest new tool name in error message
