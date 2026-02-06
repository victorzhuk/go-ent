# Testing

Testing conventions and patterns for go-ent.

## Overview

This document defines testing patterns used throughout the go-ent codebase.

## Table-Driven Tests

Use table-driven tests with `t.Run()` and `t.Parallel()`:

```go
func TestParseTask(t *testing.T) {
    tests := []struct {
        name    string
        content []byte
        want    *domain.Task
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid task",
            content: []byte("# Task\nMetadata: value"),
            want:    &domain.Task{ID: "task-id"},
            wantErr: false,
        },
        {
            name:    "invalid format",
            content: []byte("not markdown"),
            want:    nil,
            wantErr: true,
            errMsg:  "invalid format",
        },
        {
            name:    "empty content",
            content: []byte(""),
            want:    nil,
            wantErr: true,
            errMsg:  "empty content",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            parser := NewTaskParser()
            got, err := parser.Parse(tt.content)

            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want.ID, got.ID)
        })
    }
}
```

## Test Organization

- Unit tests: `*_test.go` in same package
- Integration tests: `*_integration_test.go`
- Use `testify/assert` and `testify/require`
- Use `t.Parallel()` for concurrent tests

## Behavior-Focused Testing

Test observable behavior, not implementation details:

```go
// GOOD - Test behavior through interface
func TestTaskUseCase_Execute(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*mockRepository)
        req     CreateUserRequest
        want    *CreateUserResponse
        wantErr bool
    }{
        {
            name: "successful creation",
            setup: func(m *mockRepository) {
                m.ExpectSave(nil)
            },
            req: CreateUserRequest{Name: "John"},
            want: &CreateUserResponse{ID: "user-123"},
        },
    }
    // ...
}

// BAD - Testing internal implementation
func TestBoltStore_parseTask(t *testing.T) {
    store := NewBoltStore(db)
    task := store.parseTask(mdContent)  // Testing private method
}
```

## Test Helpers

### Mock Implementations

Create mock implementations for testing:

```go
type mockSkill struct {
    name        string
    description string
    canHandle   func(ctx domain.SkillContext) bool
    result      domain.SkillResult
    err         error
}

func (m *mockSkill) Name() string {
    return m.name
}

func (m *mockSkill) Description() string {
    return m.description
}

func (m *mockSkill) CanHandle(ctx domain.SkillContext) bool {
    return m.canHandle(ctx)
}

func (m *mockSkill) Execute(ctx context.Context, req domain.SkillRequest) (domain.SkillResult, error) {
    if m.err != nil {
        return domain.SkillResult{}, m.err
    }
    return m.result, nil
}
```

### Temporary Directories

Use `t.TempDir()` for file-based tests:

```go
func TestLoad(t *testing.T) {
    tmpDir := t.TempDir()

    skillPath := filepath.Join(tmpDir, "skill1", "SKILL.md")
    require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0o750))
    require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0o600))

    r := NewRegistry()
    err := r.Load(tmpDir)
    require.NoError(t, err)
}
```

## Common Patterns

### Error Testing

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid input",
            input:   "valid",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            wantErr: true,
            errMsg:  "cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            err := Validate(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
                return
            }

            require.NoError(t, err)
        })
    }
}
```

### Setup and Teardown

```go
func TestWithSetup(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(t *testing.T) *repository
        teardown func(t *testing.T, repo *repository)
    }{
        {
            name: "test case",
            setup: func(t *testing.T) *repository {
                repo := NewRepository(t.TempDir())
                // Add test data
                return repo
            },
            teardown: func(t *testing.T, repo *repository) {
                // Cleanup if needed
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := tt.setup(t)
            if tt.teardown != nil {
                defer tt.teardown(t, repo)
            }

            // Test logic
        })
    }
}
```

## Best Practices

- Use `t.Parallel()` for independent test cases
- Use `require` for setup/failure conditions (fail fast)
- Use `assert` for assertions (continue on failure to see all issues)
- Test behavior through interfaces, not internal implementation
- Use `t.TempDir()` instead of manual temp directory creation
- Clean up resources with `defer` or `t.Cleanup()`
- Keep test cases focused and independent
- Use descriptive test case names
