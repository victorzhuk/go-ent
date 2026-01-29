---
name: ent-tools-readonly
description: "Read-only tool guidance for analysis and exploration. Preloaded by planning and research agents."
version: "1.0.0"
author: "go-ent"
disable-model-invocation: true
user-invocable: false
allowed-tools:
  - Read
  - Glob
  - Grep
---

## Role

Tool configuration skill providing read-only access for safe code analysis and exploration.

## Available Tools

This skill grants access to:

- **Read**: Read file contents (always before any analysis)
- **Glob**: Find files by pattern (e.g., `**/*.go`, `internal/**/repo.go`)
- **Grep**: Search file contents with regex patterns

## Tool Usage Patterns

### File Reading

```
Read a file before analyzing or discussing it
Always read the current version, never rely on memory
```

### File Discovery

```
Use Glob for pattern-based file finding:
  **/*.go - All Go files recursively
  internal/*/repo.go - Repository files in immediate subdirs
```

### Content Search

```
Use Grep for content-based search:
  Pattern: "func New"
  Output mode: files_with_matches (default) or content
  Context: -A/-B/-C flags for surrounding lines
```

## Constraints

- **Cannot** modify files (no Write or Edit access)
- **Cannot** execute commands (no Bash access)
- Focus on analysis, understanding, and planning
- Safe for exploratory work without side effects

## Best Practices

1. **Read before analysis**: Always read files being discussed
2. **Pattern-based search**: Use Glob for file discovery
3. **Content search**: Use Grep for code pattern analysis
4. **Minimize reads**: Read only what's necessary for the task

This toolset is ideal for:
- Planning and design work
- Code analysis and review
- Architecture exploration
- Requirement gathering
