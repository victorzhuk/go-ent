# Change: Remove Empty Resources Directory

## Why

The `/cmd/go-ent/internal/resources/` directory is empty and unused, serving no purpose in the codebase. Keeping empty directories creates maintenance confusion and violates clean code principles.

**Current State**:
- Directory exists at `/cmd/go-ent/internal/resources/`
- Contains zero files
- No code references this directory
- No imports of a `resources` package
- Created at some point but never utilized

**Impact of keeping it**:
- Developers may assume it has a purpose
- Could lead to misplaced files
- Adds noise to directory listings
- Violates YAGNI (You Aren't Gonna Need It)

This is a zero-risk cleanup that improves project hygiene.

## What Changes

### 1. Directory Removal

Remove the empty directory:

```
Before:
/cmd/go-ent/internal/
    ├── resources/      ← DELETE (empty)
    ├── tools/
    ├── server/
    └── ...

After:
/cmd/go-ent/internal/
    ├── tools/
    ├── server/
    └── ...
```

### 2. Verification

Confirm no references exist:
- No `import "github.com/victorzhuk/go-ent/cmd/go-ent/internal/resources"`
- No file paths pointing to this directory
- No documentation mentioning it

## Impact

- **Affected specs**: None (cleanup only)
- **Files removed**: 1 empty directory
- **Code changes**: None
- **Breaking changes**: None
- **Dependencies**: None
- **Build system**: No changes

## Success Criteria

1. Directory `/cmd/go-ent/internal/resources/` no longer exists
2. `make build` succeeds without errors
3. `make test` passes all tests
4. No references to `resources` in codebase

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Directory actually needed | None | Verified empty, no code references |
| Future use planned | None | Can be recreated when needed |
