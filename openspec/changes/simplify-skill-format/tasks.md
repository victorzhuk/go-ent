# Tasks: Simplify Skill Format (v4 Minimal)

## Overview

This breakdown implements the transition from complex v3 skill format to a minimal v4 format that reduces maintenance burden while preserving essential functionality.

**Estimated Effort**: ~16 hours
**Total Tasks**: 12

---

## Task 1: Add v4 Frontmatter Structure

**Status:** pending
**Priority:** high
**Effort:** medium (1-2 hours)
**Files:**
- `internal/skill/parser.go` (lines 12-30)

**Steps:**
1. Add `skillMetaV4` struct for v4 frontmatter parsing:
   ```go
   type skillMetaV4 struct {
       Name        string   `yaml:"name"`
       Description string   `yaml:"description"`
       Triggers    []string `yaml:"triggers"`
   }
   ```
2. Update `detectVersion()` to detect v4 format:
   - Check for simple frontmatter with `triggers` list (not complex object)
   - Check for Markdown sections `## Role`, `## Instructions`, `## Examples`
   - Return "v4" if both conditions met
3. Add `parseFrontmatterV4()` function:
   - Parse only 3 fields (name, description, triggers)
   - Validate required fields (name, description)
   - Convert triggers to `[]string` format

**Verification:**
```bash
# Run parser tests
go test ./internal/skill/... -run TestParser -v

# Verify v4 detection works
```

**Dependencies:** None

---

## Task 2: Implement v4 Body Parsing

**Status:** pending
**Priority:** high
**Effort:** medium (1-2 hours)
**Files:**
- `internal/skill/parser.go` (lines 32-46, 448-481)

**Steps:**
1. Simplify `extractMarkdownSection()` for v4:
   - Already exists, just verify it works with `##` headings
   - Should extract content between `## Role` and next heading
2. Update `SkillMeta` structure:
   - Remove `CoreContent` and `FullContent` (no progressive loading)
   - Add direct fields: `Role`, `Instructions`, `Examples`
   ```go
   type SkillMeta struct {
       Name        string
       Description string
       Triggers    []string
       FilePath    string
       Role        string  // Direct from v4 body
       Instructions string // Direct from v4 body
       Examples    string  // Direct from v4 body
   }
   ```
3. Update `ParseSkillFile()` for v4:
   - Parse frontmatter (3 fields only)
   - Parse body sections in single pass
   - Store directly in SkillMeta (no level indicators)

**Verification:**
```bash
# Test v4 parsing
go test ./internal/skill/... -run TestParseSkillFile -v

# Verify all sections extracted correctly
```

**Dependencies:** Task 1

---

## Task 3: Remove Progressive Loading

**Status:** pending
**Priority:** high
**Effort:** small (30 minutes)
**Files:**
- `internal/skill/parser.go` (lines 88-101, 333-392)
- `internal/skill/registry.go` (lines 794-841)

**Steps:**
1. Remove `LoadLevel` enum:
   - Delete lines 88-101 in parser.go
2. Remove `UpgradeToLevel()` method:
   - Delete lines 333-392 in parser.go
3. Remove `LoadLevel`, `Core`, `Full` fields from `SkillMeta`
4. Remove `UpgradeSkill()` and `PrepareForExecution()` from registry.go
5. Update `Registry` initialization:
   - Remove `LoadMetadata`, `LoadCore`, `LoadExtended` references
6. Update all callers of progressive loading:
   - Change to single-pass loading in `Registry.Load()`

**Verification:**
```bash
# Verify no references to LoadLevel
rg "LoadLevel" internal/skill/ --type go

# Build should succeed
go build ./internal/skill/...
```

**Dependencies:** Task 2

---

## Task 4: Simplify Validation Rules (Reduce to 5 Essential)

**Status:** pending
**Priority:** high
**Effort:** medium (2-3 hours)
**Files:**
- `internal/skill/validator.go` (lines 95-116)
- `internal/skill/rules.go` (all 790 lines)

**Current 20+ rules:**
1. validateFrontmatter
2. validateNameFormat
3. validateVersion
4. validateXMLTags
5. validateRoleSection
6. validateInstructionsSection
7. validateExamples
8. validateConstraints
9. validateEdgeCases
10. validateOutputFormat
11. checkTriggerExplicit
12. checkExampleDiversity
13. checkInstructionConcise
14. checkRedundancy

**New 5 essential rules:**
1. `validateFrontmatter` - Required fields (name, description, triggers)
2. `validateRoleSection` - Role heading exists and has content
3. `validateInstructionsSection` - Instructions heading exists and has content
4. `validateExamplesSection` - Examples heading exists and has code blocks
5. `validateTriggers` - Triggers list is non-empty

**Steps:**
1. Update `NewValidator()` to use only 5 rules:
   ```go
   return &Validator{
       rules: []ValidationRule{
           validateFrontmatterV4,
           validateRoleSectionV4,
           validateInstructionsSectionV4,
           validateExamplesSectionV4,
           validateTriggersV4,
       },
   }
   ```
2. Create simplified validation functions in rules.go:
   - Keep only essential checks
   - Remove XML tag validation (no longer needed)
   - Remove edge cases, output format validation
3. Remove outdated rules:
   - Delete `validateXMLTags`, `validateEdgeCases`, `validateOutputFormat`
   - Delete `checkExampleDiversity`, `checkInstructionConcise`, `checkRedundancy`
4. Update `ValidationContext` structure if needed:
   - May not need `Strict` field anymore

**Verification:**
```bash
# Run validation tests
go test ./internal/skill/... -run TestValidator -v

# Verify only 5 rules exist
rg "func validate" internal/skill/rules.go --type go
```

**Dependencies:** Task 3

---

## Task 5: Remove Quality Scoring System

**Status:** pending
**Priority:** medium
**Effort:** small (1 hour)
**Files:**
- `internal/skill/scorer.go` (delete entire file, 618 lines)
- `internal/skill/validator.go` (lines 44-49, 96, 135-137, 166, 192)
- `internal/skill/registry.go` (lines 213, 223-224, 384-391, 743-781, 783-792)

**Steps:**
1. Delete `internal/skill/scorer.go` completely
2. Remove `QualityScore` type from validator.go (lines 44-49)
3. Remove `QualityScore` from `ValidationResult`:
   - Remove `Score *QualityScore` field
4. Remove `scorer` from `Registry` struct (line 213)
5. Remove `scorer` initialization in `NewRegistry()` (line 223-224)
6. Remove quality scoring in `Registry.Load()`:
   - Remove `meta.QualityScore = r.scorer.Score(...)` (lines 384-391)
7. Remove quality reporting methods:
   - Delete `ValidateAll()` quality scoring (lines 753-780)
   - Delete `GetQualityReport()` (lines 783-792)

**Verification:**
```bash
# Verify no scorer references
rg "QualityScore" internal/skill/ --type go
rg "scorer" internal/skill/ --type go

# Build should succeed
go build ./internal/skill/...
```

**Dependencies:** Task 4

---

## Task 6: Remove Overlap Detection

**Status:** pending
**Priority:** low
**Effort:** small (30 minutes)
**Files:**
- `internal/skill/overlap.go` (delete entire file, 125 lines)
- `internal/skill/validator.go` (lines 86, 112-114, 170-202)
- `internal/skill/rules.go` (lines 753-789)

**Steps:**
1. Delete `internal/skill/overlap.go` completely
2. Remove `ValidationRuleWithContext` type from validator.go (line 86)
3. Remove `rulesWithContext` from `Validator` struct (line 112-114)
4. Remove `ValidateWithContext()` method (lines 170-202)
5. Delete `checkRedundancy()` function from rules.go (lines 753-789)

**Verification:**
```bash
# Verify no overlap references
rg "overlap" internal/skill/ --type go
rg "ValidateWithContext" internal/skill/ --type go

# Build should succeed
go build ./internal/skill/...
```

**Dependencies:** Task 5

---

## Task 7: Remove Auto-Fixing System

**Status:** pending
**Priority:** low
**Effort:** small (30 minutes)
**Files:**
- `internal/skill/fixer.go` (delete entire file, 819 lines)
- `internal/skill/validator.go` (no changes needed)
- `internal/skill/registry.go` (no changes needed)

**Steps:**
1. Delete `internal/skill/fixer.go` completely
2. Verify no references to `Fixer` type in other files

**Verification:**
```bash
# Verify no fixer references
rg "Fixer" internal/skill/ --type go

# Build should succeed
go build ./internal/skill/...
```

**Dependencies:** Task 4

---

## Task 8: Simplify Registry

**Status:** pending
**Priority:** high
**Effort:** medium (1-2 hours)
**Files:**
- `internal/skill/registry.go` (lines 208-226, 363-407, 726-781)

**Steps:**
1. Update `Registry` struct:
   ```go
   type Registry struct {
       skills        []SkillMeta
       runtimeSkills map[string]domain.Skill
       parser        *Parser
       validator     *Validator
       // Remove: scorer (deleted)
   }
   ```
2. Update `NewRegistry()`:
   ```go
   func NewRegistry() *Registry {
       return &Registry{
           skills:        make([]SkillMeta, 0),
           runtimeSkills: make(map[string]domain.Skill),
           parser:        NewParser(),
           validator:     NewValidator(),
           // Remove: scorer: NewQualityScorer()
       }
   }
   ```
3. Simplify `Load()` method:
   - Remove quality scoring (already done in Task 5)
   - Single-pass parsing (already done in Task 2)
   - Remove `resolveLoadOrder()` for dependencies (v4 has no dependencies)
4. Remove dependency resolution:
   - Delete `resolveLoadOrder()` method (lines 286-361)
5. Simplify `ValidateAll()`:
   - Remove quality scoring (already done in Task 5)
   - Just aggregate validation results
6. Remove `GetQualityReport()` (already done in Task 5)

**Verification:**
```bash
# Run registry tests
go test ./internal/skill/... -run TestRegistry -v

# Verify no quality scoring
rg "quality|score|scorer" internal/skill/registry.go --type go -i
```

**Dependencies:** Task 5, Task 6

---

## Task 9: Create Skill Conversion Script

**Status:** pending
**Priority:** high
**Effort:** medium (2-3 hours)
**Files:**
- `cmd/skill-convert/main.go` (new file)

**Steps:**
1. Create new command: `cmd/skill-convert/main.go`
2. Parse all existing SKILL.md files in `.claude/skills/`
3. For each file:
   - Detect current format (v1, v2, or v3)
   - Extract name and description from frontmatter
   - Convert triggers:
     - v3: Flatten `triggers.keywords` to simple list
     - v2: Use explicit triggers or description-based extraction
     - v1: Extract from description "Auto-activates for:"
   - Convert body sections:
     - v3: Markdown sections already correct
     - v2/v1: Convert XML tags to Markdown sections
4. Write v4 format:
   ```yaml
   ---
   name: {name}
   description: {description}
   triggers:
     - {trigger1}
     - {trigger2}
   ---

   ## Role

   {role_content}

   ## Instructions

   {instructions_content}

   ## Examples

   {examples_content}
   ```
5. Add command-line flags:
   - `--path`: Skills directory path
   - `--dry-run`: Show changes without writing
   - `--backup`: Create backup before converting

**Verification:**
```bash
# Build conversion tool
go build -o bin/skill-convert ./cmd/skill-convert

# Test dry-run
./bin/skill-convert --path .claude/skills --dry-run

# Run actual conversion
./bin/skill-convert --path .claude/skills --backup
```

**Dependencies:** Task 3

---

## Task 10: Convert All Existing Skills to v4

**Status:** pending
**Priority:** high
**Effort:** medium (2-3 hours)
**Files:**
- `.claude/skills/**/SKILL.md` (multiple files)

**Steps:**
1. Find all SKILL.md files:
   ```bash
   find .claude/skills -name "SKILL.md"
   ```
2. Run conversion script from Task 9
3. Manual review of converted skills:
   - Verify frontmatter has only 3 fields
   - Check triggers are simple lists (not objects)
   - Verify body uses `## Role`, `## Instructions`, `## Examples`
   - Remove unused frontmatter fields:
     - version, author, license, compatibility, tags, quality_score, category
   - Remove XML tags if present
   - Remove empty triggers: `{}`
4. Update `.opencode/skills/` if they exist

**Skills to convert (examples):**
- `.claude/skills/ent/go/go-code/SKILL.md`
- `.claude/skills/ent/go/go-api/SKILL.md`
- `.claude/skills/ent/go/go-arch/SKILL.md`
- `.claude/skills/ent/core/api-design/SKILL.md`
- `.claude/skills/ent/core/arch-core/SKILL.md`
- ... (all other skills)

**Verification:**
```bash
# Verify all skills have v4 format
rg "^name:" .claude/skills --type md -l | wc -l
rg "^description:" .claude/skills --type md -l | wc -l
rg "^triggers:" .claude/skills --type md -l | wc -l

# Verify no v3 frontmatter fields
rg "^(version|author|license|compatibility|tags|quality_score|category):" .claude/skills --type md

# Verify no XML tags
rg "<(role|instructions|examples|constraints|edge_cases|output_format)>" .claude/skills --type md

# Run validation
go run ./build/ent skill validate
```

**Dependencies:** Task 9

---

## Task 11: Update MCP Skill Tools

**Status:** pending
**Priority:** high
**Effort:** small (1 hour)
**Files:**
- `internal/mcp/tools/skill_validate.go` (lines 62-94)
- `internal/mcp/tools/skill_quality.go` (delete entire file, 111 lines)
- `internal/mcp/server.go` (registration calls)

**Steps:**
1. Update `skill_validate.go`:
   - Remove `Score` field from `SkillValidateOutput`:
     ```go
     type SkillValidateOutput struct {
         Valid  bool                    `json:"valid"`
         Issues []skill.ValidationIssue `json:"issues"`
         // Remove: Score float64
     }
     ```
   - Remove quality score from `formatValidationOutput()` (line 92)
   - Remove score from `CallToolResult` (line 64)
2. Delete `skill_quality.go` completely:
   - Quality reporting tool no longer needed
3. Update tool registration in MCP server:
   - Remove `registerSkillQuality()` call
   - Keep `registerSkillValidate()`

**Verification:**
```bash
# Build MCP server
go build ./cmd/ent

# Test skill_validate tool
echo '{"name":"go-code"}' | ./build/ent mcp call skill_validate

# Verify skill_quality tool removed
echo '{}' | ./build/ent mcp list | grep -i quality
# Should return nothing
```

**Dependencies:** Task 5, Task 8

---

## Task 12: Update Tests and Documentation

**Status:** pending
**Priority:** medium
**Effort:** medium (2-3 hours)
**Files:**
- `internal/skill/parser_test.go` (update existing tests)
- `internal/skill/validator_test.go` (update existing tests)
- `internal/skill/registry_test.go` (update existing tests)
- `internal/skill/fixer_test.go` (delete file)
- `internal/skill/overlap_test.go` (delete file)
- `internal/skill/scorer_test.go` (delete file)
- `internal/skill/scorer_bench_test.go` (delete file)
- `internal/skill/scorer_integration_test.go` (delete file)
- `internal/skill/progressive_load_test.go` (delete file)
- `internal/skill/load_extended_test.go` (delete file)
- `docs/INDEX.md` (update skill format documentation)

**Steps:**

### Test Updates:
1. Delete removed functionality tests:
   - Delete `fixer_test.go`
   - Delete `overlap_test.go`
   - Delete `scorer_test.go`, `scorer_bench_test.go`, `scorer_integration_test.go`
   - Delete `progressive_load_test.go`
   - Delete `load_extended_test.go`
2. Update `parser_test.go`:
   - Add tests for v4 format parsing
   - Remove tests for progressive loading
   - Update fixtures to use v4 format
3. Update `validator_test.go`:
   - Remove tests for deleted rules (quality scoring, overlap, edge cases, output format)
   - Add tests for 5 essential rules
4. Update `registry_test.go`:
   - Remove tests for quality scoring
   - Remove tests for progressive loading
   - Update to test single-pass loading

### Documentation Updates:
1. Update `docs/INDEX.md`:
   - Document v4 skill format
   - Remove v3 format documentation
   - Update examples to show v4 format
   - Remove progressive loading section
   - Remove quality scoring section
2. Create migration guide (optional):
   - Document v3 → v4 conversion process
   - List breaking changes

**Verification:**
```bash
# Run all tests
go test ./internal/skill/... -v -race

# Verify no test files for removed features
ls internal/skill/*test.go | grep -E "(fixer|overlap|scorer|progressive|extended)"

# Verify test count reduced
go test ./internal/skill/... -list . | wc -l
```

**Dependencies:** Task 11

---

## Task Dependencies Graph

```
Task 1 (v4 Frontmatter)
    ↓
Task 2 (v4 Body Parsing)
    ↓
Task 3 (Remove Progressive Loading)
    ↓
Task 4 (Simplify Validation) ──→ Task 5 (Remove Quality Scoring) ──→
                    │                                         │
                    ↓                                         ↓
                Task 6 (Remove Overlap)               Task 8 (Simplify Registry)
                    │                                         │
                    ↓                                         ↓
                Task 7 (Remove Fixer) ───────────────────→ Task 11 (Update MCP Tools)
                    │                                         │
                    ↓                                         ↓
                Task 9 (Conversion Script) ──────────────→ Task 10 (Convert All Skills)
                    │                                         │
                    ↓                                         ↓
                ──────────────────────────────────────────→ Task 12 (Update Tests & Docs)
```

---

## Success Criteria

- [ ] Parser handles v4 format correctly
- [ ] All skills converted to v4 format
- [ ] Validation reduced to 5 essential rules
- [ ] Progressive loading completely removed
- [ ] Quality scoring system removed
- [ ] Overlap detection removed
- [ ] Auto-fixing system removed
- [ ] Registry simplified
- [ ] MCP skill tools updated
- [ ] All tests updated and passing
- [ ] Documentation updated
- [ ] `go test ./... -race` passes
- [ ] `go build ./...` succeeds

---

## Risk Mitigation

**Risk:** Breaking existing skill functionality
**Mitigation:** Keep v3 parsing support during transition period, remove after all skills converted

**Risk:** Test coverage gaps
**Mitigation:** Update tests incrementally with each task, run after each task completion

**Risk:** Documentation lag
**Mitigation:** Update docs in Task 12 after all implementation complete

---

## Notes

1. **Backward Compatibility:** Consider keeping v3 parsing as fallback during transition period (optional)
2. **Git History:** Skills should be converted in a separate branch for easy rollback
3. **Backup Strategy:** Task 9 includes `--backup` flag for safety
4. **Testing Strategy:** Run tests after each task, not just at end
5. **Documentation Priority:** Update docs immediately after Task 12 completion

---

## References

- Proposal: `openspec/changes/simplify-skill-format/proposal.md`
- Parser: `internal/skill/parser.go`
- Validator: `internal/skill/validator.go`
- Registry: `internal/skill/registry.go`
- Example skills: `.claude/skills/ent/go/go-code/SKILL.md`
