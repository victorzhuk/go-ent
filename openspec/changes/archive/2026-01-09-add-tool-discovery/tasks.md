# Tasks: Add Tool Discovery System

## 1. TF-IDF Search Implementation
- [x] 1.1 Create `internal/mcp/tools/search.go`
- [x] 1.2 Implement term extraction with stopword filtering
- [x] 1.3 Implement TF-IDF scoring algorithm
- [x] 1.4 Add `BuildDocument()` helper for tool indexing
- [x] 1.5 Unit tests for search accuracy

## 2. Tool Registry Implementation
- [x] 2.1 Create `internal/mcp/tools/discovery.go`
- [x] 2.2 Implement `ToolRegistry` with lazy loading
- [x] 2.3 Add `Register()`, `Find()`, `Describe()` methods
- [x] 2.4 Add `Load()` method for dynamic activation
- [x] 2.5 Add `Active()` and `All()` query methods
- [x] 2.6 Implement `BuildIndex()` for TF-IDF
- [x] 2.7 Add thread-safety with RWMutex

## 3. Discovery MCP Tools
- [x] 3.1 Create `internal/mcp/tools/meta.go`
- [x] 3.2 Implement `tool_find` with query and limit parameters
- [x] 3.3 Implement `tool_describe` with detailed metadata output
- [x] 3.4 Implement `tool_load` for dynamic activation
- [x] 3.5 Implement `tool_active` to list loaded tools
- [x] 3.6 Add formatted output with markdown

## 4. Integration
- [x] 4.1 Update `register.go` to initialize ToolRegistry
- [x] 4.2 Call `registerMetaTools()` with registry
- [x] 4.3 Build search index after registration
- [x] 4.4 Add logging for tool count and initialization
- [x] 4.5 Verify build passes

## 5. Tool Renaming (Breaking)
- [x] 5.1 Rename spec tools (`spec_*`)
- [x] 5.2 Rename registry tools (`registry_*`)
- [x] 5.3 Rename workflow tools (`workflow_*`)
- [x] 5.4 Rename loop tools (`loop_*`)
- [x] 5.5 Rename generate tools (`generate*`)
- [x] 5.6 Rename agent_execute tool
- [x] 5.7 Update `openspec/AGENTS.md`
- [x] 5.8 Update `plugins/go-ent/commands/*.md`
- [x] 5.9 Update `plugins/go-ent/agents/*.md`

## 6. Testing & Validation
- [x] 6.1 Test `tool_find` with various queries
- [x] 6.2 Test `tool_describe` for all tools
- [x] 6.3 Test `tool_load` dynamic activation
- [x] 6.4 Test `tool_active` listing
- [x] 6.5 Verify search accuracy >80%
- [x] 6.6 Measure token reduction for simple tasks
- [x] 6.7 Test thread-safety under concurrent access

## 7. Documentation
- [x] 7.1 Add usage examples to proposal.md
- [x] 7.2 Document tool search queries
- [x] 7.3 Add migration guide for tool name changes
- [x] 7.4 Update CHANGELOG with breaking changes
- [x] 7.5 Add architecture diagrams
