# Tasks: Merge Small Packages (Phase 2)

## Task Breakdown

### 1. Move model/category.go to config/
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Copy `internal/model/category.go` to `internal/config/category.go`
2. Change package declaration from `package model` to `package config`
3. Search for imports: `grep -r "internal/model" .`
4. Update imports to use `config` package
5. Delete original file

**Validation:**
- [ ] File moved successfully
- [ ] Package declaration updated
- [ ] Imports updated
- [ ] Build succeeds

**Files Modified:**
- `internal/model/category.go` (moved to `internal/config/`)
- Files importing model package

---

### 2. Move model/config.go to config/
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Task 1

**Steps:**
1. Copy `internal/model/config.go` to `internal/config/model_config.go` (rename to avoid conflict)
2. Change package declaration from `package model` to `package config`
3. Update imports
4. Delete original file

**Validation:**
- [ ] File moved successfully
- [ ] No naming conflicts
- [ ] Build succeeds

**Files Modified:**
- `internal/model/config.go` (moved to `internal/config/model_config.go`)

---

### 3. Move model/defaults.go to config/
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Task 2

**Steps:**
1. Copy `internal/model/defaults.go` to `internal/config/defaults.go`
2. Change package declaration from `package model` to `package config`
3. Update imports
4. Delete original file

**Validation:**
- [ ] File moved successfully
- [ ] Build succeeds

**Files Modified:**
- `internal/model/defaults.go` (moved to `internal/config/`)

---

### 4. Move model/resolver.go to config/
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Task 3

**Steps:**
1. Copy `internal/model/resolver.go` to `internal/config/resolver.go`
2. Change package declaration from `package model` to `package config`
3. Update imports
4. Delete original file

**Validation:**
- [ ] File moved successfully
- [ ] Build succeeds

**Files Modified:**
- `internal/model/resolver.go` (moved to `internal/config/`)

---

### 5. Delete internal/model/ directory
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Tasks 1-4

**Steps:**
1. Verify all files moved
2. Remove `internal/model/` directory
3. Search for any remaining imports: `grep -r "internal/model" .`

**Validation:**
- [ ] Directory deleted
- [ ] No imports remain
- [ ] Build succeeds

**Files Modified:**
- `internal/model/` (deleted)

---

### 6. Move openspec/tracker.go to spec/
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Task 5

**Steps:**
1. Copy `internal/openspec/tracker.go` to `internal/spec/tracker.go`
2. Change package declaration from `package openspec` to `package spec`
3. Search for imports: `grep -r "internal/openspec" .`
4. Update imports to use `spec` package
5. Delete original file

**Validation:**
- [ ] File moved successfully
- [ ] Package declaration updated
- [ ] Imports updated
- [ ] Build succeeds

**Files Modified:**
- `internal/openspec/tracker.go` (moved to `internal/spec/`)
- Files importing openspec package

---

### 7. Delete internal/openspec/ directory
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Task 6

**Steps:**
1. Verify all files moved
2. Remove `internal/openspec/` directory
3. Search for any remaining imports: `grep -r "internal/openspec" .`

**Validation:**
- [ ] Directory deleted
- [ ] No imports remain
- [ ] Build succeeds

**Files Modified:**
- `internal/openspec/` (deleted)

---

### 8. Final verification
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** All previous tasks

**Steps:**
1. Run `make build` - must succeed
2. Run `make test` - all tests must pass
3. Run `make lint` - no lint errors
4. Verify package count reduced by 2
5. Search for dead imports: `grep -r "internal/model\|internal/openspec" .`

**Validation:**
- [ ] Build succeeds
- [ ] Tests pass
- [ ] Lint clean
- [ ] 2 fewer packages
- [ ] No dead imports

**Files Modified:**
- None (verification only)

---

## Task Order

**Sequential:** Tasks must be done in order 1-8 to avoid import conflicts

## Estimated Total Time

- Tasks 1-4: 1.5 hours (model package merge)
- Task 5: 15 minutes (cleanup)
- Tasks 6-7: 1 hour (openspec package merge)
- Task 8: 30 minutes (verification)
- **Total:** ~3 hours

## Testing Strategy

1. **Incremental Builds:** Verify build after each file move
2. **Import Verification:** Check for dead imports after each merge
3. **Final Integration:** Full make build/test/lint cycle
