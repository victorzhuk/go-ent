---
name: go-review
description: "Code review for Go enterprise standards, SOLID, Clean Architecture. Auto-activates for: code review, PR review, quality checks, Go validation."
allowed-tools: Read, Grep, Glob, Bash
---

# Go Code Review

## Review Checklist

### 1. Architecture (Dependencies flow inward)
```
Transport → UseCase → Domain ← Repository ← Infrastructure
```
- Domain has ZERO external deps, NO struct tags
- Interfaces defined at consumer side
- Private models in repository with mappers

### 2. Naming (Concise, no AI-style)
```go
// ❌ REJECT
applicationConfiguration := config.Load()
userRepositoryInstance := userRepo.New(pool)

// ✅ ACCEPT  
cfg := config.Load()
repo := userRepo.New(pool)
```

### 3. Comments (ZERO explaining WHAT)
```go
// ❌ REJECT - explains what
// Create a new user
user := NewUser(name)

// ✅ ACCEPT - explains why (rare)
// Required by legacy API - remove after v2 migration
resp.Header.Set("X-Legacy-Token", token)
```

### 4. Error Handling
```go
// ❌ REJECT
return fmt.Errorf("Failed to query user: %w", err)
return err  // no context

// ✅ ACCEPT
return fmt.Errorf("query user %s: %w", id, err)
```

### 5. Patterns
- Constructor: `New()` public, `new*()` private
- Context as first parameter
- Happy path left (early returns)
- Structs private by default

## Review Commands

```bash
# Architecture violations
grep -r "import.*transport" internal/domain/

# AI-style names
grep -rn "applicationConfig\|userRepository\|databaseConnection" internal/

# Comment violations  
grep -rn "// Create\|// Get\|// Set\|// Check" internal/ | grep -v "_test.go"

# Error handling
grep -rn 'return err$' internal/
```

## Output Format

```markdown
## Code Review

### 🚨 Critical (Must Fix)
- [FILE:LINE] Issue description
  ```go
  // Current → Suggested
  ```

### ⚠️ Warnings | 💡 Suggestions | ✅ Well Done
```
