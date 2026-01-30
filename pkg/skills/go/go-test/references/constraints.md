# Constraints

- Include table-driven tests with t.Run() and t.Parallel()
- Include testify/assert and testify/require for assertions
- Include proper test setup/teardown with testcontainers for integration tests
- Include mocks using testify/mock for usecase layer tests
- Include race detection with -race flag for concurrent code
- Include coverage reporting for measuring test coverage
- Exclude testing private implementation details directly
- Exclude global mutable state in tests (use test-scoped state)
- Exclude sleeps or time.Sleep in tests (use timeouts or channels)
- Exclude duplicate test logic across multiple files (extract helpers)
- Bound to layer-specific testing: domain (pure unit), usecase (mock repos), repository (testcontainers), transport (httptest)
- Follow TDD principles when adding new features
- Use descriptive test names that explain what is being tested