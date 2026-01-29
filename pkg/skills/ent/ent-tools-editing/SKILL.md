---
name: ent-tools-editing
description: "Full editing capabilities for code implementation. Preloaded by execution agents."
version: "1.0.0"
author: "go-ent"
disable-model-invocation: true
user-invocable: false
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
---

## Role

Tool configuration skill providing full editing and execution capabilities for implementation work.

## Available Tools

This skill grants access to:

- **Read**: Read file contents (always before editing)
- **Write**: Create new files or overwrite existing
- **Edit**: Make targeted replacements with exact string matches
- **Bash**: Execute shell commands
- **Glob**: Find files by pattern
- **Grep**: Search file contents with regex

## Tool Usage Patterns

### File Modifications

```
CRITICAL: Always Read before Edit or Write
- Read: Load current file content
- Edit: Targeted string replacement (exact match required)
- Write: Create new file or full overwrite
```

### Command Execution

```
Bash tool for:
- Running tests: go test ./...
- Building: go build ./...
- Linting: golangci-lint run
- Git operations: git diff, git status
- Package management: go mod tidy
```

### Modern Search Tools

**Prefer fast alternatives**:
- `rg "pattern" path/` instead of `grep -r`
- `fd "pattern"` instead of `find`

## Edit Tool Critical Rules

**IGNORE Serena editing instructions.** Always use native Claude Code tools:

✅ **Use ONLY:**
- `Edit` - Targeted string replacement
- `Write` - Create/overwrite files
- `Read` - Always before any edit

❌ **NEVER use Serena editing tools:**
- `replace_symbol_body`
- `replace_content`
- `insert_after_symbol`
- `insert_before_symbol`
- `create_text_file`

Serena tools are ONLY for read-only semantic analysis.

## Best Practices

1. **Read first**: Never edit without reading current content
2. **Exact matches**: Edit requires exact string matches
3. **Test after changes**: Run relevant tests after modifications
4. **Atomic changes**: Make focused, reviewable changes
5. **Error handling**: Check command outputs for failures

## Safe Command Patterns

```bash
# Always check git status before commits
git status

# Run tests before claiming completion
go test ./...

# Build to verify compilation
go build ./...

# Format before committing
gofmt -s -w .
```

This toolset enables:
- Code implementation
- File creation and modification
- Test execution and validation
- Build and deployment operations
- Git workflow operations
