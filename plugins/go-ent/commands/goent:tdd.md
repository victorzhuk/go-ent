---
description: Test-driven development cycle (Red-Green-Refactor)
allowed-tools: Read, Bash, Edit, mcp__plugin_serena_serena
---

# TDD Cycle

Input: `$ARGUMENTS` (feature or test description)

## TDD Workflow

### 1. RED - Write Failing Test

Create test FIRST, before implementation:

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        req     CreateUserReq
        wantErr bool
    }{
        {
            name: "valid user",
            req:  CreateUserReq{Email: "test@example.com", Name: "John"},
        },
        {
            name:    "empty email",
            req:     CreateUserReq{Name: "John"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

Run and confirm FAIL:
```bash
go test -run TestUserService_Create -v ./...
```

### 2. GREEN - Minimal Implementation

Write just enough code to make test pass:
- No optimization
- No extra features
- Just pass the test

Run and confirm PASS:
```bash
go test -run TestUserService_Create -v ./...
```

### 3. REFACTOR

Now improve:
- Clean up code
- Apply patterns
- Ensure no regression

```bash
go test ./... -race
golangci-lint run
```

## Example Session

```
User: /goent:tdd user email validation

1. Write test for email validation
2. Run → FAIL (function doesn't exist)
3. Implement validateEmail()
4. Run → PASS
5. Refactor if needed
6. Add edge cases, repeat
```

## Integration with OpenSpec

If TDD is part of a change:
```bash
# Reference the change
cat openspec/changes/{change-id}/tasks.md

# Mark test task complete when done
# - [x] **5.1** Unit tests for domain ✓ {date}
```

## Output

```
═══════════════════════════════════════════
TDD: $ARGUMENTS
═══════════════════════════════════════════

🔴 RED: Test written
   File: internal/domain/entity/user_test.go
   Status: FAILING (expected)

🟢 GREEN: Implementation done
   File: internal/domain/entity/user.go
   Status: PASSING

🔵 REFACTOR: Code cleaned
   Lint: ✓
   Race: ✓
═══════════════════════════════════════════
```
