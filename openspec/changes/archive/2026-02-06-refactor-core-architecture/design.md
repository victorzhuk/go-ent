## Context

The go-ent codebase emerged from rapid prototyping, resulting in several architectural issues:

**Current State:**
- `spec/boltdb.go` (~300 lines): God object mixing BoltDB storage operations with markdown parsing logic
- `skill/registry.go`: Single file managing skill metadata, runtime execution, and matching logic (~400+ lines)
- Tests: Tightly coupled to implementation details, using concrete types, testing internal functions
- Error handling: Inconsistent patterns, some errors wrapped, some not, varying context levels
- String utilities: Custom functions for what stdlib provides (trim, split, replace patterns)
- Configuration: Deeply nested structures making environment variable mapping complex

**Constraints:**
- No breaking changes to public CLI API or MCP tool interfaces
- Must maintain test coverage during refactoring
- BoltDB data format cannot change (existing user data)
- Cannot introduce new external dependencies
- Must follow Go best practices and Clean Architecture principles

**Stakeholders:**
- Development team: Better maintainability and onboarding
- Users: Stable CLI and MCP interfaces, no migration required
- CI/CD: Must keep tests passing throughout incremental changes

## Goals / Non-Goals

**Goals:**
- Apply Clean Architecture principles with clear domain/usecase/repository/transport layers
- Extract single-responsibility components from god objects
- Define interfaces at consumer boundaries for testability
- Establish consistent error handling patterns throughout codebase
- Replace custom string utilities with stdlib equivalents
- Flatten configuration structure for better readability
- Replace brittle implementation tests with behavior-focused tests
- Maintain backward compatibility for all public APIs

**Non-Goals:**
- Changing BoltDB data schema or storage format
- Modifying CLI command structure or arguments
- Introducing new external dependencies
- Breaking MCP tool interfaces
- Changing the markdown spec file format
- Adding new features or capabilities (this is refactoring only)

## Decisions

### 1. Clean Architecture Layering

**Decision:** Apply strict 4-layer Clean Architecture (Domain → UseCase → Repository → Transport) with dependencies flowing inward.

**Rationale:**
- Separation of concerns makes code easier to understand and test
- Domain layer with zero external deps ensures business logic is testable without mocks
- Interfaces at consumer boundaries allow flexible implementations
- Established pattern with proven benefits for maintainability

**Alternative Considered:** Keep current layered approach but enforce boundaries better
**Rejection:** Current code doesn't have clear layer boundaries; full Clean Architecture provides explicit guardrails

### 2. Interface Definitions at Consumer Boundaries

**Decision:** Define interfaces only where consumed, not where implemented (accept interfaces, return structs).

**Rationale:**
- Consumers control the contract they depend on
- Prevents premature abstraction (YAGNI)
- Enables testing with minimal code changes
- Follows Go standard library patterns

**Example:**
```go
// In skill/usecase package (consumer)
type Store interface {
    FindSkill(ctx context.Context, id string) (*domain.Skill, error)
    ListSkills(ctx context.Context) ([]*domain.Skill, error)
}

// In skill/infrastructure/bolt package (implementation)
type boltStore struct { /* ... */ }
func (s *boltStore) FindSkill(ctx context.Context, id string) (*domain.Skill, error) {
    // implementation
}
```

**Alternative Considered:** Define interfaces in repository package for all implementations
**Rejection:** Violates "interfaces at consumer side" principle, creates unused abstractions

### 3. Incremental Refactoring with Test Safety Net

**Decision:** Refactor incrementally, keeping tests passing after each change using the "Strangler Fig" pattern.

**Rationale:**
- Minimizes risk of breaking changes
- Allows continuous validation
- Easier to revert individual steps
- Maintains confidence throughout the process

**Process:**
1. Extract interface for existing implementation
2. Create new component with new interface
3. Write tests for new component
4. Migrate callers to new component
5. Remove old implementation

**Example from BoltStore:**
```go
// Step 1: Define parsing interface in usecase layer
type Parser interface {
    ParseTask(content []byte) (*domain.Task, error)
}

// Step 2: Extract parser implementation
type taskParser struct{}
func (p *taskParser) ParseTask(content []byte) (*domain.Task, error) { /* ... */ }

// Step 3: Update BoltStore to use parser
type BoltStore struct {
    db     *bolt.DB
    parser Parser
}

// Step 4: Gradually migrate callers to use parser directly
// Step 5: Remove parsing logic from BoltStore
```

**Alternative Considered:** Complete rewrite of problematic files
**Rejection:** Too risky, harder to validate, potential for missed functionality

### 4. Behavior-Focused Testing Strategy

**Decision:** Rewrite tests to focus on observable behavior using interfaces, not implementation details.

**Rationale:**
- Tests survive refactoring
- More meaningful validation of actual requirements
- Easier to understand test intent
- Reduces test brittleness

**Pattern:**
```go
// Before (brittle)
func TestBoltStore_ParseTask(t *testing.T) {
    store := NewBoltStore(db)
    task := store.parseTask(mdContent)
    assert.Equal(t, "task-id", task.ID) // Testing internal method
}

// After (behavior-focused)
func TestTaskParser_ParseTask(t *testing.T) {
    tests := []struct {
        name    string
        content []byte
        want    *domain.Task
        wantErr bool
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
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewTaskParser()
            got, err := parser.Parse(tt.content)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want.ID, got.ID)
        })
    }
}
```

**Alternative Considered:** Keep existing tests but add new behavior tests
**Rejection:** Doubles test maintenance burden, old tests may give false confidence

### 5. Consistent Error Handling Pattern

**Decision:** Establish and enforce error handling conventions across all packages.

**Pattern:**
```go
// Package-level errors in errors.go
var (
    ErrSkillNotFound = errors.New("skill not found")
    ErrInvalidTask   = errors.New("invalid task format")
)

// Error wrapping with context
func (s *boltStore) FindSkill(ctx context.Context, id string) (*domain.Skill, error) {
    var m skillModel
    err := s.db.View(/* ... */)
    if err != nil {
        if errors.Is(err, bolt.ErrBucketNotFound) {
            return nil, fmt.Errorf("skill %s: %w", id, ErrSkillNotFound)
        }
        return nil, fmt.Errorf("query skill %s: %w", id, err)
    }
    return toDomain(&m), nil
}
```

**Conventions:**
- Lowercase error messages
- No trailing punctuation
- Wrap with `%w` for error checking
- Include context (what operation, what identifier)
- Use `errors.Is()` for error type checking

**Alternative Considered:** Use error wrapping library like `pkg/errors`
**Rejection:** Go 1.13+ error wrapping is sufficient, extra dependency not needed

### 6. Configuration Flattening

**Decision:** Flatten nested configuration structures to 1-2 levels max.

**Before:**
```go
type Config struct {
    Database struct {
        BoltDB struct {
            Path string
            Mode string
        }
        Postgres struct {
            Host string
            Port int
        }
    }
    Server struct {
        HTTP struct {
            Host string
            Port int
        }
    }
}
```

**After:**
```go
type Config struct {
    BoltDBPath   string `env:"BOLTDB_PATH" envDefault:"./ent.db"`
    BoltDBMode   string `env:"BOLTDB_MODE" envDefault:"0600"`
    PGHost       string `env:"PG_HOST" envDefault:"localhost"`
    PGPort       int    `env:"PG_PORT" envDefault:"5432"`
    HTTPHost     string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
    HTTPPort     int    `env:"HTTP_PORT" envDefault:"8080"`
}
```

**Rationale:**
- Easier to read and understand
- Simpler environment variable mapping
- Reduces pointer dereferencing
- More idiomatic Go

**Alternative Considered:** Keep nested but provide accessor methods
**Rejection:** Adds complexity without benefit

### 7. Dependency Injection Pattern

**Decision:** Use constructor injection with struct composition for clean dependency graph.

**Pattern:**
```go
// Domain layer (no external deps)
type Task struct {
    ID      string
    Name    string
    Status  TaskStatus
}

// Repository implementation
type boltTaskRepo struct {
    db *bolt.DB
}

func NewBoltTaskRepo(db *bolt.DB) Repository {
    return &boltTaskRepo{db: db}
}

// UseCase
type taskUseCase struct {
    repo Repository
    log  *slog.Logger
}

func NewTaskUseCase(repo Repository, log *slog.Logger) UseCase {
    return &taskUseCase{repo: repo, log: log}
}

// Transport
type TaskHandler struct {
    uc UseCase
}

func NewTaskHandler(uc UseCase) *TaskHandler {
    return &TaskHandler{uc: uc}
}

// Application wiring
func NewApplication(cfg *Config, log *slog.Logger) (*Application, error) {
    db, err := bolt.Open(cfg.BoltDBPath, cfg.BoltDBMode, nil)
    if err != nil {
        return nil, fmt.Errorf("open bolt db: %w", err)
    }

    taskRepo := repository.NewBoltTaskRepo(db)
    taskUC := usecase.NewTaskUseCase(taskRepo, log)
    taskHandler := transport.NewTaskHandler(taskUC)

    return &Application{taskHandler: taskHandler}, nil
}
```

**Rationale:**
- Explicit dependencies
- Easy to test with mocks/fakes
- Clear dependency graph
- Follows Go conventions

**Alternative Considered:** Use dependency injection framework (wire, dig, fx)
**Rejection:** Adds complexity and learning curve, simple constructors are sufficient

## Risks / Trade-offs

### Risk: Breaking Existing Functionality During Refactoring
**Impact:** High - could break user workflows or data
**Mitigation:**
- Incremental refactoring with tests passing at each step
- Comprehensive integration tests before and after
- Feature flags to switch between old and new implementations
- Rollback plan for each major change

### Risk: Test Refactoring Overwhelms Development Time
**Impact:** Medium - slower progress on new features
**Mitigation:**
- Refactor tests alongside code (not as separate phase)
- Focus on high-impact, brittle tests first
- Accept that some low-risk tests may remain temporarily implementation-focused

### Risk: Over-Engineering During Clean Architecture Application
**Impact:** Medium - unnecessary complexity
**Mitigation:**
- Start with existing patterns, only extract where needed
- Don't create interfaces "for future flexibility" (YAGNI)
- Code review focus on simplicity and necessity

### Risk: Performance Regression from Layer Indirection
**Impact:** Low to Medium - may affect startup time or memory
**Mitigation:**
- Benchmark before and after refactoring
- Optimize hot paths after correctness established
- Accept small performance cost for maintainability gains

### Risk: Configuration Breaking Changes
**Impact:** Medium - existing deployments may fail
**Mitigation:**
- Maintain backward compatibility with environment variable names
- Add deprecation warnings before removing old variables
- Document migration path clearly
- Test with existing production config if available

### Trade-off: More Files/Code vs Better Organization
**Decision:** Accept more files for better organization
**Rationale:**
- Smaller files are easier to understand
- Clearer boundaries reduce cognitive load
- Better test organization
- Follows standard Go project layouts

## Migration Plan

### Phase 1: Foundation (Week 1-2)
**Goal:** Establish patterns and tooling

**Tasks:**
1. Define error handling conventions in `CODE.md`
2. Create common domain types if missing
3. Set up test utilities for table-driven tests
4. Add linter rules for consistent patterns

**Validation:**
- Linter passes with new rules
- Documentation updated

### Phase 2: Extract Task Parsing (Week 2-3)
**Goal:** Separate parsing from storage in `spec/boltdb.go`

**Steps:**
1. Create `spec/parser/` package with `TaskParser` interface and implementation
2. Write behavior-focused tests for `TaskParser`
3. Update `BoltStore` to accept `TaskParser` via constructor
4. Migrate callers to use `TaskParser` directly where appropriate
5. Remove parsing logic from `BoltStore`

**Validation:**
- All tests pass
- `BoltStore` reduced from ~300 to ~100 lines
- Task parsing tests use table-driven pattern

### Phase 3: Split Skill Registry (Week 3-4)
**Goal:** Decompose `skill/registry.go` into focused components

**Steps:**
1. Create `skill/domain/` package for core skill types
2. Create `skill/repository/` for storage operations
3. Create `skill/matcher/` for skill matching logic
4. Create `skill/runtime/` for skill execution
5. Update `registry.go` to compose these components
6. Write behavior-focused tests for each component

**Validation:**
- `registry.go` reduced to coordination (~50 lines)
- Each component has clear responsibility
- Tests focus on behavior, not implementation

### Phase 4: Standardize Error Handling (Week 4-5)
**Goal:** Apply consistent error patterns across codebase

**Steps:**
1. Audit all packages for error handling patterns
2. Create `errors.go` files in packages lacking them
3. Update error wrapping to include context
4. Replace custom error types with package-level errors where appropriate
5. Add tests for error behavior (`errors.Is()`, `errors.As()`)

**Validation:**
- Linter check for consistent error patterns
- All errors provide context
- Error types documented in package docs

### Phase 5: Replace Custom String Functions (Week 5)
**Goal:** Use stdlib string functions

**Steps:**
1. Find all custom string utility functions
2. Identify stdlib equivalents (`strings`, `path`, `filepath`)
3. Replace custom functions
4. Remove unused utility code
5. Update tests accordingly

**Validation:**
- No custom string utilities remain
- All tests pass
- Code simpler and more idiomatic

### Phase 6: Simplify Configuration (Week 5-6)
**Goal:** Flatten configuration structures

**Steps:**
1. Audit all configuration structures
2. Design flat configuration with clear environment variable mapping
3. Update configuration loading logic
4. Add backward compatibility layer if needed
5. Update documentation
6. Test with existing environment variables

**Validation:**
- All configuration structures 1-2 levels max
- Environment variable mapping clear and documented
- Existing deployments continue to work

### Phase 7: Clean Architecture Application (Week 6-7)
**Goal:** Ensure all packages follow Clean Architecture principles

**Steps:**
1. Audit each package for layer violations
2. Extract domain logic where mixed with infrastructure
3. Define interfaces at consumer boundaries
4. Update dependency injection to use constructors
5. Add integration tests for use cases

**Validation:**
- Each package has clear layer (domain/usecase/repository/transport)
- Interfaces at consumer side
- Domain has zero external dependencies
- Integration tests cover use cases

### Phase 8: Test Refactoring (Week 7-8)
**Goal:** Replace brittle tests with behavior-focused tests

**Steps:**
1. Identify tests tied to implementation details
2. Refactor to test observable behavior
3. Use interfaces and table-driven patterns
4. Add integration tests where missing
5. Remove unused tests

**Validation:**
- No tests test internal functions
- All tests use table-driven pattern
- Test coverage maintained or improved
- Tests survive refactoring

### Phase 9: Documentation and Cleanup (Week 8)
**Goal:** Update documentation and remove deprecated code

**Steps:**
1. Update `CODE.md` with new architecture patterns
2. Add architecture diagrams
3. Document error handling conventions
4. Remove deprecated code and features
5. Update `AGENTS.md` and other documentation

**Validation:**
- Documentation reflects new architecture
- No TODO or deprecated code comments
- New developer onboarding smooth

### Rollback Strategy

For each phase:
1. Keep old implementation alongside new until tests pass
2. Use Git branches for each phase
3. Ability to revert individual phases independently
4. Tag commits before each major change
5. Document rollback procedure in README

If a phase fails:
1. Revert that phase's changes
2. Analyze failure
3. Adjust approach
4. Retry with modifications

## Open Questions

1. **BoltDB vs other storage:** Should we plan to migrate away from BoltDB in the future?
   - **Decision:** Out of scope for this refactoring. BoltDB is working; migration is a separate initiative.

2. **Skill execution model:** Should skill runtime support async execution?
   - **Decision:** Keep current synchronous model for now. Async is a feature, not refactoring scope.

3. **Configuration validation:** How strict should validation be? Should we fail fast on invalid config?
   - **Decision:** Validate on load, fail fast with clear error messages. Add `--validate-config` flag for pre-flight checks.

4. **Test coverage target:** What minimum coverage percentage is acceptable after refactoring?
   - **Decision:** Maintain current coverage (~70%), aim for 75% in critical paths. Quality over quantity.

5. **Backward compatibility duration:** How long to maintain deprecated environment variables?
   - **Decision:** One release cycle (~2 weeks) with deprecation warnings, then remove.

6. **Performance budgets:** Should we set performance budgets for refactored components?
   - **Decision:** Yes. Benchmark critical paths before refactoring, ensure no regression >10% post-refactoring.
