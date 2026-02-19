## TDD Cycle
1. **RED**: Write a failing test that describes desired behavior
2. **GREEN**: Write minimal code to make the test pass
3. **REFACTOR**: Clean up while keeping tests green
4. Repeat

## Testing Pyramid
```
       /  E2E  \         Few, slow, expensive
      / Integration \     Some, moderate speed
     /    Unit Tests  \   Many, fast, cheap
```

## Test Qualities (FIRST)
- **F**ast: Unit tests run in milliseconds
- **I**ndependent: Tests don't depend on each other
- **R**epeatable: Same result every time
- **S**elf-validating: Pass or fail, no manual inspection
- **T**imely: Written before or alongside production code

## What to Test
- Business logic and domain rules (always)
- Edge cases and error conditions (always)
- Integration points: DB, APIs, queues (integration tests)
- Critical user workflows (E2E tests)
- Do NOT test: framework internals, trivial getters/setters, private methods

## Patterns
- **Arrange-Act-Assert**: Setup → Execute → Verify
- **Given-When-Then**: Context → Action → Outcome (BDD style)
- **Test Doubles**: Stubs (return values), Mocks (verify calls), Fakes (simplified impls)
- **Builder Pattern**: Build complex test objects with sensible defaults
- **Object Mother / Factory**: Centralize test data creation

## Anti-Patterns to Avoid
- Testing implementation details instead of behavior
- Brittle tests that break on refactoring
- Slow test suites (parallelism, minimize I/O in unit tests)
- Test interdependence (shared mutable state)
- Excessive mocking (if you mock everything, you test nothing)
