---
name: go-review
description: "Code review patterns and quality checks. Auto-activates for: code review, quality checks, PR review, architecture validation."
---

# Go Code Review

## Checklist

### 1. Architecture
```
Transport → UseCase → Domain ← Repository ← Infrastructure
```
- Domain has ZERO external deps, NO struct tags
- Interfaces defined at consumer side

### 2. Naming
```go
// ❌ REJECT
applicationConfiguration := config.Load()

// ✅ ACCEPT
cfg := config.Load()
```

### 3. Comments
```go
// ❌ REJECT - explains what
// Create a new user
user := NewUser(name)

// ✅ ACCEPT - explains why (rare)
// Required by legacy API
resp.Header.Set("X-Legacy-Token", token)
```

### 4. Error Handling
```go
// ❌ REJECT
return fmt.Errorf("Failed to query user: %w", err)
return err

// ✅ ACCEPT
return fmt.Errorf("query user %s: %w", id, err)
```

## Review Commands

```bash
# Architecture violations
grep -r "import.*transport" internal/domain/

# AI-style names
grep -rn "applicationConfig\|userRepository" internal/

# Comment violations
grep -rn "// Create\|// Get\|// Set" internal/ | grep -v "_test.go"

# Error handling
grep -rn 'return err$' internal/
```

## Serena

```
mcp__serena__find_symbol(name: "UserRepository")
mcp__serena__find_referencing_symbols(symbol: "CreateUser")
mcp__serena__get_project_structure()
```

## Output Format

```markdown
## Code Review

### 🚨 Critical (Must Fix)
- [FILE:LINE] Issue description

### ⚠️ Warnings
...

### ✅ Well Done
- Good error handling in X
```

## Confidence Filter

Only report >= 80%:
- 95-100%: Bugs, security, Go idiom violations
- 85-94%: Quality issues, anti-patterns
- <80%: Skip — subjective preferences
