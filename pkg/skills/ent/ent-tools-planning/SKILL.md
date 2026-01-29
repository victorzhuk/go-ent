---
name: ent-tools-planning
description: "Combined planning toolset: read-only exploration + task management + semantic analysis. Preloaded by planner agents."
version: "1.0.0"
author: "go-ent"
disable-model-invocation: true
user-invocable: false
allowed-tools:
  - Read
  - Glob
  - Grep
  - TodoRead
  - TodoWrite
  - Skill
  - List
  - mcp__plugin_serena_serena__find_symbol
  - mcp__plugin_serena_serena__find_referencing_symbols
  - mcp__plugin_serena_serena__get_symbols_overview
  - mcp__plugin_serena_serena__search_for_pattern
  - mcp__plugin_serena_serena__list_dir
  - mcp__plugin_serena_serena__read_file
---

## Role

Comprehensive tool configuration for planning agents combining read-only exploration, semantic analysis, and task management capabilities.

## Available Tools

### File Operations (Read-Only)
- **Read**: Read file contents
- **Glob**: Find files by pattern
- **Grep**: Search file contents

### Semantic Analysis (Serena)
- **find_symbol**: Locate code symbols by name
- **find_referencing_symbols**: Track symbol usage
- **get_symbols_overview**: Understand file structure
- **search_for_pattern**: Pattern-based code search
- **list_dir**: Directory exploration
- **read_file**: Serena file reader

### Task Management
- **TodoRead**: Read task list
- **TodoWrite**: Create and update tasks
- **List**: View available resources

### Skill Invocation
- **Skill**: Invoke other skills for specialized guidance

## Planning Workflow

### Phase 1: Discovery

```
1. Use Glob/fd to discover relevant files
2. Use get_symbols_overview for high-level structure
3. Use find_symbol for detailed symbol analysis
4. Use search_for_pattern to find existing patterns
```

### Phase 2: Analysis

```
1. Read key files identified in discovery
2. Use find_referencing_symbols for impact analysis
3. Identify architectural boundaries and dependencies
4. Map data flow and control flow
```

### Phase 3: Task Breakdown

```
1. TodoWrite to create task structure
2. Break work into phases and milestones
3. Identify dependencies between tasks
4. Estimate complexity and effort
```

### Phase 4: Documentation

```
1. Document architectural decisions
2. Create implementation plan with file references
3. List prerequisites and assumptions
4. Identify risks and mitigation strategies
```

## Modern Search Tools

**Prefer fast alternatives in Bash:**
- `rg "pattern" path/` (10-100x faster than grep)
- `fd "pattern"` (faster than find)

## Task Management Patterns

```
TodoWrite for:
- Creating phase-based task hierarchies
- Breaking down features into steps
- Tracking dependencies and blockers
- Organizing work by architectural layer
```

## Best Practices

1. **Start broad**: Overview before detail
2. **Semantic first**: Use Serena tools for code understanding
3. **Pattern awareness**: Search for existing patterns before designing new ones
4. **Task structure**: Create clear, actionable task breakdowns
5. **Documentation**: Reference specific files and line numbers
6. **Skill delegation**: Use Skill tool for specialized analysis

## No Modification

This toolset is **read-only** for safe planning:

❌ Cannot Write files
❌ Cannot Edit code
❌ Cannot Execute commands (except via Bash for search tools)

Focus on understanding, analysis, and structured planning.

## Ideal For

- Feature planning and task breakdown
- Architecture analysis and design
- Impact assessment for changes
- Dependency mapping
- Risk identification
- Effort estimation
- Implementation roadmap creation
