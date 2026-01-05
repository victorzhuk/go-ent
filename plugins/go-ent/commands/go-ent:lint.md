---
description: Run Go linters and fix issues
---

# Go Lint

Run linters and analyze issues.

## Steps

1. Run golangci-lint:
   ```bash
   golangci-lint run --timeout 5m
   ```

2. Check enterprise rules:
   ```bash
   # AI-style names
   grep -rn "applicationConfig\|userRepository\|databaseConnection" internal/
   
   # Comment violations
   grep -rn "// Create\|// Get\|// Set\|// Check" internal/ | grep -v "_test.go"
   
   # Error handling
   grep -rn 'return err$' internal/
   
   # Magic numbers
   grep -rn '[0-9][0-9][0-9]' internal/ | grep -v "_test.go" | grep -v "const"
   ```

3. Run goimports:
   ```bash
   goimports -d .
   ```

For each issue: location, type, current code, suggested fix.
