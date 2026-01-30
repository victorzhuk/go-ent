# Spec Delta: Code Style - Comment Guidelines

## MODIFIED Requirements

### Requirement: Code must not contain WHAT comments

Code SHALL NOT contain comments that simply describe what the code does. Only WHY comments explaining counterintuitive behavior, legacy requirements, or workarounds MUST be allowed.

#### Scenario: Function without WHAT comment

**Given** a well-named function like `validateUserInput()`
**When** reviewing the code
**Then** there is no comment saying "// Validate user input"
**And** the function name clearly communicates its purpose

#### Scenario: Complex logic with WHY comment

**Given** code with counterintuitive behavior
**When** reviewing the code
**Then** a WHY comment explains the reasoning
**And** the comment does NOT simply restate what the code does

**Example acceptable comment:**
```go
// Counterintuitive: zero means unlimited per vendor docs
if limit == 0 {
    return math.MaxInt64
}
```

#### Scenario: Legacy workaround with WHY comment

**Given** code with a workaround for legacy compatibility
**When** reviewing the code
**Then** a WHY comment explains the legacy requirement
**And** indicates when it can be removed

**Example acceptable comment:**
```go
// Required by legacy API - remove after v2 migration
resp.Header.Set("X-Legacy-Token", token)
```

---

### Requirement: Exported APIs must have documentation comments

Public functions, types, and methods SHALL have godoc-style documentation comments for API documentation generation.

#### Scenario: Exported function has documentation

**Given** an exported function `Execute()`
**When** generating API documentation
**Then** a godoc comment describes what the function does
**And** the comment follows godoc conventions

**Example:**
```go
// Execute runs a task with automatic runner and strategy selection.
// Returns the execution result or an error if execution fails.
func (e *Engine) Execute(ctx context.Context, task *Task) (*Result, error)
```

#### Scenario: Unexported function has no documentation

**Given** an unexported function `execute()`
**When** reviewing the code
**Then** there is no documentation comment
**And** the function name is self-documenting

---

### Requirement: Package documentation must be present

Every package SHALL have a package-level documentation comment explaining its purpose and main concepts.

#### Scenario: Package has documentation comment

**Given** a package `internal/execution`
**When** reviewing the package
**Then** a `// Package execution` comment exists
**And** it describes the package's purpose and main types

**Example:**
```go
// Package execution provides task execution engine with multiple
// runtime support, state persistence, and budget tracking.
package execution
```
