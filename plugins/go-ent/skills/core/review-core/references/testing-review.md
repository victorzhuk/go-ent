# Test Coverage Review

Example showing test coverage review for a user registration feature.

## Example

<example>
<input>Test coverage review for user registration feature</input>
<output>
## Review Feedback: PR #156 - Add User Registration

### Testing Concerns (P1 - Request Changes)

**⚠️ 1. Missing integration tests for edge cases**

```go
// Only unit tests present
func TestUserRegistration_Valid(t *testing.T) {
    // ... happy path tests
}
```

**Issue**: Registration involves database, email service, and validation. Unit tests alone don't catch integration issues.

**Fix**: Add integration tests:
```go
func TestUserRegistration_Integration(t *testing.T) {
    ctx := context.Background()

    // Test duplicate email
    err := service.Register(ctx, &RegisterRequest{Email: "exists@example.com"})
    if !errors.Is(err, ErrEmailExists) {
        t.Errorf("Expected ErrEmailExists, got %v", err)
    }

    // Test email sending
    // Test transaction rollback on failure
    // Test concurrent registrations
}
```

---

**⚠️ 2. No tests for concurrent scenarios**

```go
func (s *Service) Register(ctx context.Context, req *RegisterRequest) error {
    // ... validation and creation
}
```

**Issue**: Race conditions possible when multiple users register with same email simultaneously.

**Fix**: Add concurrency test:
```go
func TestUserRegistration_Concurrent(t *testing.T) {
    ctx := context.Background()

    const n = 10
    errChan := make(chan error, n)

    for i := 0; i < n; i++ {
        go func() {
            errChan <- service.Register(ctx, &RegisterRequest{
                Email:    "duplicate@example.com",
                Password: "secure123",
            })
        }()
    }

    successCount := 0
    errorCount := 0
    for i := 0; i < n; i++ {
        if err := <-errChan; err == nil {
            successCount++
        } else {
            errorCount++
        }
    }

    // Should only succeed once
    if successCount != 1 {
        t.Errorf("Expected 1 success, got %d", successCount)
    }
}
```

---

**⚠️ 3. Missing failure scenarios**

Current tests cover:
- ✓ Valid registration
- ✗ Invalid email format
- ✗ Weak password
- ✗ Duplicate email
- ✗ Database errors
- ✗ Email service failures

**Fix**: Add table-driven test for failure cases:
```go
func TestUserRegistration_FailureCases(t *testing.T) {
    tests := []struct {
        name    string
        req     *RegisterRequest
        wantErr error
    }{
        {
            name:    "empty email",
            req:     &RegisterRequest{Email: ""},
            wantErr: ErrValidation,
        },
        {
            name:    "invalid email format",
            req:     &RegisterRequest{Email: "not-an-email"},
            wantErr: ErrValidation,
        },
        {
            name:    "password too short",
            req:     &RegisterRequest{Email: "test@example.com", Password: "123"},
            wantErr: ErrWeakPassword,
        },
        {
            name:    "password too long",
            req:     &RegisterRequest{Email: "test@example.com", Password: string(make([]byte, 129))},
            wantErr: ErrValidation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            err := service.Register(ctx, tt.req)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("Expected %v, got %v", tt.wantErr, err)
            }
        })
    }
}
```

---

### Code Quality Issues (P2 - Suggestions)

**💡 4. Test naming could be more descriptive**

Current: `TestUserRegistration_Valid`
Suggestion: `TestUserRegistration_Success_ValidInput`

**💡 5. Missing test setup helpers**

```go
// Repeated test setup
func TestUserRegistration_Valid(t *testing.T) {
    db := setupDB(t)
    defer db.Close()
    service := NewService(db, mockEmailer)
    // ...
}

func TestUserRegistration_Duplicate(t *testing.T) {
    db := setupDB(t)
    defer db.Close()
    service := NewService(db, mockEmailer)
    // ...
}
```

**Fix**: Extract to helper:
```go
func newTestService(t *testing.T) (*Service, *sql.DB, *MockEmailer) {
    db := setupDB(t)
    emailer := &MockEmailer{}
    return NewService(db, emailer), db, emailer
}
```

---

**💡 6. No benchmarks for performance-critical paths**

```go
func BenchmarkUserRegistration(b *testing.B) {
    service := newTestService(b)
    req := &RegisterRequest{
        Email:    "test@example.com",
        Password: "securepassword123",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Clear DB or use transaction
        _ = service.Register(context.Background(), req)
    }
}
```

---

### Positive Feedback

✅ Unit tests cover happy path
✅ Uses table-driven tests where appropriate
✅ Test data is clear and readable
✅ Mock email service used properly

---

### Summary

**Action**: Request changes (test coverage gaps)

**Blocking**: Integration tests, concurrent scenarios, failure cases

**Estimated time to fix**: 3-4 hours

**Notes**: Registration is critical user flow. Need confidence that it works correctly under load and error conditions. Add integration tests with testcontainers for realistic environment.
</output>
</example>
