---
name: ent-tools-serena-analysis
description: "Serena semantic analysis tools for deep code understanding. Preloaded by architecture and research agents."
version: "1.0.0"
author: "go-ent"
disable-model-invocation: true
user-invocable: false
allowed-tools:
  - mcp__plugin_serena_serena__find_symbol
  - mcp__plugin_serena_serena__find_referencing_symbols
  - mcp__plugin_serena_serena__get_symbols_overview
  - mcp__plugin_serena_serena__search_for_pattern
  - mcp__plugin_serena_serena__list_dir
  - mcp__plugin_serena_serena__read_file
---

## Role

Tool configuration skill providing Serena MCP semantic analysis capabilities for deep code understanding and architectural analysis.

## Available Tools

**Semantic Analysis (Read-Only):**

- **find_symbol**: Find code symbols (classes, functions, methods) by name path
- **find_referencing_symbols**: Find all references to a symbol across codebase
- **get_symbols_overview**: Get high-level overview of symbols in a file
- **search_for_pattern**: Flexible pattern-based content search with filters
- **list_dir**: Directory structure listing
- **read_file**: Read file contents with Serena's reader

## Tool Usage Patterns

### Symbol Discovery

```
find_symbol:
  name_path_pattern: "User/FindByID"  # Class/Method
  relative_path: "internal/user"       # Restrict search
  include_body: true                   # Get full implementation
  depth: 1                             # Include child symbols
```

### Reference Analysis

```
find_referencing_symbols:
  name_path: "User/Save"
  relative_path: "internal/user/repo.go"
  include_info: true  # Get hover info
```

### File Structure

```
get_symbols_overview:
  relative_path: "internal/user/repo.go"
  depth: 1  # Include immediate children
```

### Pattern Search

```
search_for_pattern:
  substring_pattern: "return err$"  # Unwrapped errors
  relative_path: "internal/"
  restrict_search_to_code_files: true
  paths_include_glob: "**/*.go"
```

## Name Path Patterns

Symbol identification follows path-based naming:

```
MyClass/myMethod     # Method in class
MyClass/myMethod[0]  # First overload
/MyClass/myMethod    # Absolute path (exact match)
myMethod             # Any symbol named myMethod
```

## Best Practices

1. **Start with overview**: Use `get_symbols_overview` to understand file structure
2. **Targeted reads**: Use symbol-level reads instead of full files
3. **Reference tracking**: Use `find_referencing_symbols` to understand usage
4. **Pattern validation**: Use `search_for_pattern` to find anti-patterns
5. **Scope restriction**: Always provide `relative_path` to narrow search

## Read-Only Emphasis

**CRITICAL**: Serena tools are **ONLY** for analysis.

✅ **Use for:**
- Understanding code structure
- Finding symbol definitions
- Tracking symbol references
- Analyzing patterns

❌ **NEVER use for editing:**
- Do NOT use `replace_symbol_body`
- Do NOT use `replace_content`
- Do NOT use `insert_after_symbol`
- Do NOT use `insert_before_symbol`

For editing, use native Claude Code tools (Edit, Write).

## Ideal Use Cases

- Architecture analysis and mapping
- Dependency graphing
- Impact analysis for changes
- Code pattern discovery
- Symbol relationship understanding
- Refactoring planning
