# Skills Full Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate 70 SKILL.md files and the go-ent validator to the official Claude Code skills standard.

**Architecture:** Three phases — (1) validator/parser changes with transition mode, (2) skill content migrations in parallel batches, (3) strict-mode lockdown and renames. The parser's `detectVersion` must be updated first because every migrated skill fails parsing until it accepts the new body sections.

**Tech Stack:** Go 1.23+, `internal/skill/` package, `pkg/skills/` content, `pkg/templates/`, `github.com/stretchr/testify/assert`

---

## Phase 1 — Validator & Parser (Tasks 1–9)

All code changes. TDD throughout. Must complete before Phase 2.

---

### Task 1: Add `Status` to `SkillMeta` and `skillMetaV4`

**Files:**
- Modify: `internal/skill/parser.go`
- Modify: `internal/skill/parser_test.go`

**Step 1: Write the failing test**

Add to `internal/skill/parser_test.go`:

```go
func TestParseSkillFile_ParsesStatus(t *testing.T) {
    t.Parallel()

    content := "---\nname: go-api\ndescription: test\nstatus: production\n---\n\n## Overview\n\nTest overview with enough content here.\n\n## When to Use\n\n- Use this\n\n## Implementation\n\n### Core\n\nDo this.\n"
    f := writeTempSkill(t, content)

    p := NewParser()
    meta, err := p.ParseSkillFile(f)
    assert.NoError(t, err)
    assert.Equal(t, "production", meta.Status)
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/skill/... -run TestParseSkillFile_ParsesStatus -v
```
Expected: `FAIL — meta.Status == ""`

**Step 3: Add `Status` to `SkillMeta` and `skillMetaV4`**

In `internal/skill/parser.go`, add to `SkillMeta`:
```go
Status string
```

Add to `skillMetaV4`:
```go
Status string `yaml:"status"`
```

In `ParseSkillFile`, after building `result`:
```go
result.Status = v4Meta.Status
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/skill/... -run TestParseSkillFile_ParsesStatus -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/parser.go internal/skill/parser_test.go
git commit -m "feat(skill): add Status field to SkillMeta and v4 frontmatter parser"
```

---

### Task 2: Add new body sections to `SkillMeta` and parser

**Files:**
- Modify: `internal/skill/parser.go`
- Modify: `internal/skill/parser_test.go`

**Step 1: Write failing tests**

Add to `internal/skill/parser_test.go`:

```go
func TestParseSkillFile_ParsesNewSections(t *testing.T) {
    t.Parallel()

    content := "---\nname: go-api\ndescription: test\nstatus: production\n---\n\n## Overview\n\nSpec-first API design.\n\n## When to Use\n\n- Designing a REST API\n\n## Implementation\n\n### Spec-First\n\nWrite OpenAPI first.\n\n## Common Mistakes\n\n- Mixing business logic into handlers\n\n## Resources\n\n- [`references/patterns.md`](references/patterns.md)\n"
    f := writeTempSkill(t, content)

    p := NewParser()
    meta, err := p.ParseSkillFile(f)
    assert.NoError(t, err)
    assert.Contains(t, meta.Overview, "Spec-first API design")
    assert.Contains(t, meta.WhenToUse, "Designing a REST API")
    assert.Contains(t, meta.Implementation, "Write OpenAPI first")
    assert.Contains(t, meta.CommonMistakes, "Mixing business logic")
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestParseSkillFile_ParsesNewSections -v
```
Expected: compile error — `meta.Overview` undefined

**Step 3: Add fields and parsing**

In `internal/skill/parser.go`, add to `SkillMeta`:
```go
Overview       string
WhenToUse      string
Implementation string
CommonMistakes string
```

In `ParseSkillFile`, add to `result` construction:
```go
Overview:       p.extractMarkdownSection(contentStr, "Overview"),
WhenToUse:      p.extractMarkdownSection(contentStr, "When to Use"),
Implementation: p.extractMarkdownSection(contentStr, "Implementation"),
CommonMistakes: p.extractMarkdownSection(contentStr, "Common Mistakes"),
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestParseSkillFile_ParsesNewSections -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/parser.go internal/skill/parser_test.go
git commit -m "feat(skill): add Overview, WhenToUse, Implementation, CommonMistakes sections to parser"
```

---

### Task 3: Update `detectVersion` to accept new body format

This is the highest-risk task. The current check requires `## Role + ## Instructions + ## Examples`. Migrated skills have none of these. Update detection to accept either format.

**Files:**
- Modify: `internal/skill/parser.go`
- Modify: `internal/skill/parser_test.go`

**Step 1: Write failing test for new format detection**

Add to `internal/skill/parser_test.go`:

```go
func TestDetectVersion_NewFormatAccepted(t *testing.T) {
    t.Parallel()

    p := NewParser()

    tests := []struct {
        name        string
        frontmatter string
        content     string
        wantVersion string
    }{
        {
            name:        "new format: Overview + When to Use + Implementation",
            frontmatter: "name: go-api\ndescription: test\nstatus: production",
            content:     "## Overview\n\ntest\n\n## When to Use\n\n- use it\n\n## Implementation\n\n### Core\n\ndo it",
            wantVersion: "v4",
        },
        {
            name:        "old format: triggers + Role + Instructions + Examples",
            frontmatter: "name: go-api\ndescription: test\ntriggers:\n  - go api",
            content:     "## Role\n\ntest\n\n## Instructions\n\ntest\n\n## Examples\n\ntest",
            wantVersion: "v4",
        },
        {
            name:        "neither format",
            frontmatter: "name: go-api",
            content:     "## Random\n\ntest",
            wantVersion: "unknown",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got := p.detectVersion(tt.content, tt.frontmatter)
            assert.Equal(t, tt.wantVersion, got)
        })
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestDetectVersion_NewFormatAccepted -v
```
Expected: new format case returns "unknown", not "v4"

**Step 3: Update `detectVersion`**

In `internal/skill/parser.go`, replace `detectVersion`:

```go
func (p *Parser) detectVersion(content, frontmatter string) string {
    hasNewSections := strings.Contains(content, "## Overview") &&
        strings.Contains(content, "## When to Use") &&
        strings.Contains(content, "## Implementation")

    hasOldSections := strings.Contains(content, "## Role") &&
        strings.Contains(content, "## Instructions") &&
        strings.Contains(content, "## Examples")

    hasTriggers := strings.Contains(frontmatter, "triggers:")

    if hasNewSections || (hasTriggers && hasOldSections) {
        return "v4"
    }

    return "unknown"
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestDetectVersion_NewFormatAccepted -v
```
Expected: `PASS`

**Step 5: Run full suite — no regressions**

```bash
go test ./internal/skill/... -v
```
Expected: all existing tests still pass

**Step 6: Commit**

```bash
git add internal/skill/parser.go internal/skill/parser_test.go
git commit -m "feat(skill): accept new body sections (Overview/WhenToUse/Implementation) as v4 format"
```

---

### Task 4: Add `validateStatus` rule

**Files:**
- Modify: `internal/skill/rules.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write failing test**

Add to `internal/skill/validator_test.go`:

```go
func TestValidateStatus(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        status   string
        wantErr  bool
        wantWarn bool
    }{
        {"production is valid", "production", false, false},
        {"draft is valid", "draft", false, false},
        {"deprecated is valid", "deprecated", false, false},
        {"delegated is valid", "delegated", false, false},
        {"empty status is error", "", true, false},
        {"invalid status is error", "wip", true, false},
        {"uppercase invalid", "Production", true, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := &ValidationContext{
                Meta:    &SkillMeta{Name: "test-skill", Status: tt.status},
                Content: "---\nstatus: " + tt.status + "\n---",
                Lines:   splitLines("---\nstatus: " + tt.status + "\n---"),
            }
            issues := validateStatusV4(ctx)
            if tt.wantErr {
                assert.True(t, hasSeverity(issues, SeverityError), "expected error for status %q", tt.status)
            } else {
                assert.Empty(t, issues, "unexpected issues for status %q", tt.status)
            }
        })
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestValidateStatus -v
```
Expected: compile error — `validateStatusV4` undefined

**Step 3: Implement `validateStatusV4` in `rules.go`**

Add at bottom of `internal/skill/rules.go`:

```go
var validStatuses = map[string]bool{
    "production": true,
    "draft":      true,
    "deprecated": true,
    "delegated":  true,
}

// validateStatusV4 checks that the status field is present and valid.
func validateStatusV4(ctx *ValidationContext) []ValidationIssue {
    status := strings.TrimSpace(ctx.Meta.Status)
    if status == "" {
        return []ValidationIssue{{
            Rule:       "status",
            Severity:   SeverityError,
            Message:    "missing required field: status",
            Suggestion: "Add a 'status' field to frontmatter",
            Example:    "status: production  # production | draft | deprecated | delegated",
            Line:       1,
        }}
    }

    if !validStatuses[status] {
        return []ValidationIssue{{
            Rule:       "status",
            Severity:   SeverityError,
            Message:    fmt.Sprintf("invalid status %q: must be one of: production, draft, deprecated, delegated", status),
            Suggestion: "Use one of the valid status values",
            Example:    "status: draft",
            Line:       findLineNumber(ctx.Lines, "status:"),
        }}
    }

    return nil
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestValidateStatus -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/rules.go internal/skill/validator_test.go
git commit -m "feat(skill): add validateStatusV4 rule (production|draft|deprecated|delegated)"
```

---

### Task 5: Add `validateDescriptionQuality` rule

**Files:**
- Modify: `internal/skill/rules.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write failing tests**

Add to `internal/skill/validator_test.go`:

```go
func TestValidateDescriptionQuality(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        desc     string
        wantErr  bool
        wantWarn bool
    }{
        {
            name: "good third-person description",
            desc: "This skill should be used when the user asks to design a REST API spec-first or mentions OpenAPI, ogen, or gRPC.",
        },
        {
            name: "use when trigger phrase",
            desc: "Use when implementing any feature or bugfix, before writing implementation code.",
        },
        {
            name: "asks to trigger phrase",
            desc: "Use when the user asks to debug a goroutine leak or profile memory allocations.",
        },
        {
            name:    "exceeds 1024 chars",
            desc:    strings.Repeat("a", 1025),
            wantErr: true,
        },
        {
            name:    "no trigger signal",
            desc:    "Spec-first API design with OpenAPI and ogen.",
            wantErr: true,
        },
        {
            name:     "workflow summary warning",
            desc:     "Use when designing APIs — first write the spec, then generate code, then implement handlers.",
            wantWarn: true,
        },
        {
            name:     "first-person warning",
            desc:     "I can help you design REST APIs when asked.",
            wantWarn: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := &ValidationContext{
                Meta:    &SkillMeta{Name: "test-skill", Description: tt.desc},
                Content: "",
                Lines:   []string{},
            }
            issues := validateDescriptionQualityV4(ctx)
            if tt.wantErr {
                assert.True(t, hasSeverity(issues, SeverityError), "expected error")
            } else if !tt.wantWarn {
                for _, i := range issues {
                    assert.NotEqual(t, SeverityError, i.Severity, "unexpected error: %s", i.Message)
                }
            }
            if tt.wantWarn {
                assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
            }
        })
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestValidateDescriptionQuality -v
```
Expected: compile error — `validateDescriptionQualityV4` undefined

**Step 3: Implement in `rules.go`**

Add at bottom of `internal/skill/rules.go`:

```go
var descTriggerSignals = []string{"when", "use when", "asks to", "should be used"}
var descWorkflowWords = []string{"first,", "then,", "step 1", " → ", "->"}

// validateDescriptionQualityV4 checks description follows official third-person trigger-specific style.
func validateDescriptionQualityV4(ctx *ValidationContext) []ValidationIssue {
    desc := ctx.Meta.Description
    var issues []ValidationIssue

    if len(desc) > 1024 {
        issues = append(issues, ValidationIssue{
            Rule:       "description-quality",
            Severity:   SeverityError,
            Message:    fmt.Sprintf("description too long: %d chars (max: 1024)", len(desc)),
            Suggestion: "Shorten description to under 1024 characters",
            Example:    "This skill should be used when the user asks to design a REST API or mentions OpenAPI.",
        })
    }

    hasTrigger := false
    lowerDesc := strings.ToLower(desc)
    for _, signal := range descTriggerSignals {
        if strings.Contains(lowerDesc, signal) {
            hasTrigger = true
            break
        }
    }
    if !hasTrigger {
        issues = append(issues, ValidationIssue{
            Rule:       "description-quality",
            Severity:   SeverityError,
            Message:    "description missing trigger signal: must contain 'when', 'use when', 'asks to', or 'should be used'",
            Suggestion: "Rewrite as third-person trigger-specific prose",
            Example:    "This skill should be used when the user asks to design a REST API or work with OpenAPI.",
        })
    }

    for _, word := range descWorkflowWords {
        if strings.Contains(lowerDesc, word) {
            issues = append(issues, ValidationIssue{
                Rule:       "description-quality",
                Severity:   SeverityWarning,
                Message:    "description may summarise workflow — Claude may shortcut skill body instead of reading it",
                Suggestion: "Describe WHEN to use the skill, not HOW it works",
            })
            break
        }
    }

    if strings.HasPrefix(desc, "I ") || strings.HasPrefix(desc, "I'") {
        issues = append(issues, ValidationIssue{
            Rule:       "description-quality",
            Severity:   SeverityWarning,
            Message:    "description uses first-person — rewrite in third person",
            Suggestion: "Start with 'This skill should be used when...' or 'Use when...'",
        })
    }

    return issues
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestValidateDescriptionQuality -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/rules.go internal/skill/validator_test.go
git commit -m "feat(skill): add validateDescriptionQualityV4 (trigger signal, length, workflow summary)"
```

---

### Task 6: Add `validateEmptyPlaceholders` rule

Fixes the `debug-core` critical bug: empty `<tag></tag>` accepted by current validator.

**Files:**
- Modify: `internal/skill/rules.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write failing test**

Add to `internal/skill/validator_test.go`:

```go
func TestValidateEmptyPlaceholders(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        content string
        wantErr bool
    }{
        {
            name:    "empty example input tag",
            content: "<example><input></input><output>result</output></example>",
            wantErr: true,
        },
        {
            name:    "empty output tag",
            content: "<example><input>query</input><output></output></example>",
            wantErr: true,
        },
        {
            name:    "whitespace-only tag",
            content: "<example><input>   </input><output>result</output></example>",
            wantErr: true,
        },
        {
            name:    "self-closing tag pattern",
            content: "## Examples\n\n<example><input/></example>",
            wantErr: true,
        },
        {
            name:    "no placeholder tags",
            content: "## Implementation\n\n### Core\n\nWrite a function.\n",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := &ValidationContext{
                Meta:    &SkillMeta{Name: "test-skill"},
                Content: tt.content,
                Lines:   splitLines(tt.content),
            }
            issues := validateEmptyPlaceholdersV4(ctx)
            if tt.wantErr {
                assert.True(t, hasSeverity(issues, SeverityError), "expected error")
            } else {
                assert.Empty(t, issues)
            }
        })
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestValidateEmptyPlaceholders -v
```
Expected: compile error — `validateEmptyPlaceholdersV4` undefined

**Step 3: Implement in `rules.go`**

Add at bottom of `internal/skill/rules.go`:

```go
var emptyTagPattern = regexp.MustCompile(`<(\w+)>\s*</\1>|<\w+\s*/>`)

// validateEmptyPlaceholdersV4 detects empty XML placeholder tags left from templates.
func validateEmptyPlaceholdersV4(ctx *ValidationContext) []ValidationIssue {
    if emptyTagPattern.MatchString(ctx.Content) {
        return []ValidationIssue{{
            Rule:       "empty-placeholders",
            Severity:   SeverityError,
            Message:    "skill contains empty XML placeholder tags — fill in or remove",
            Suggestion: "Replace empty tags with real content, or remove the section entirely",
            Example:    "Remove <input></input> and add actual example input text",
        }}
    }
    return nil
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestValidateEmptyPlaceholders -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/rules.go internal/skill/validator_test.go
git commit -m "fix(skill): add validateEmptyPlaceholdersV4 to catch empty XML template tags"
```

---

### Task 7: Add `validateNameMatchesDir` rule

**Files:**
- Modify: `internal/skill/rules.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write failing test**

Add to `internal/skill/validator_test.go`:

```go
func TestValidateNameMatchesDir(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        skillName string
        filePath string
        wantErr  bool
    }{
        {
            name:      "name matches directory",
            skillName: "go-api",
            filePath:  "/repo/pkg/skills/go/go-api/SKILL.md",
            wantErr:   false,
        },
        {
            name:      "name does not match directory",
            skillName: "different-name",
            filePath:  "/repo/pkg/skills/go/go-api/SKILL.md",
            wantErr:   true,
        },
        {
            name:      "no file path - skip check",
            skillName: "go-api",
            filePath:  "",
            wantErr:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := &ValidationContext{
                FilePath: tt.filePath,
                Meta:     &SkillMeta{Name: tt.skillName, FilePath: tt.filePath},
                Content:  "",
                Lines:    []string{},
            }
            issues := validateNameMatchesDirV4(ctx)
            if tt.wantErr {
                assert.True(t, hasSeverity(issues, SeverityError), "expected error")
            } else {
                assert.Empty(t, issues)
            }
        })
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestValidateNameMatchesDir -v
```
Expected: compile error — `validateNameMatchesDirV4` undefined

**Step 3: Implement in `rules.go`**

Add import `"path/filepath"` to rules.go imports. Add at bottom:

```go
// validateNameMatchesDirV4 checks that the skill's name field matches its containing directory name.
func validateNameMatchesDirV4(ctx *ValidationContext) []ValidationIssue {
    if ctx.FilePath == "" {
        return nil
    }

    dirName := filepath.Base(filepath.Dir(ctx.FilePath))
    if ctx.Meta.Name != "" && ctx.Meta.Name != dirName {
        return []ValidationIssue{{
            Rule:       "name-matches-dir",
            Severity:   SeverityError,
            Message:    fmt.Sprintf("name %q does not match directory name %q", ctx.Meta.Name, dirName),
            Suggestion: "Rename the skill's 'name' field to match its directory, or rename the directory",
            Example:    fmt.Sprintf("name: %s", dirName),
            Line:       findLineNumber(ctx.Lines, "name:"),
        }}
    }

    return nil
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestValidateNameMatchesDir -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/rules.go internal/skill/validator_test.go
git commit -m "feat(skill): add validateNameMatchesDirV4 rule"
```

---

### Task 8: Add `validateResourceLinks` rule

**Files:**
- Modify: `internal/skill/rules.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write failing test**

Add to `internal/skill/validator_test.go`:

```go
func TestValidateResourceLinks(t *testing.T) {
    t.Parallel()

    // Create a temp directory with a real references file
    dir := t.TempDir()
    skillFile := filepath.Join(dir, "SKILL.md")
    refFile := filepath.Join(dir, "references", "patterns.md")
    require.NoError(t, os.MkdirAll(filepath.Dir(refFile), 0o755))
    require.NoError(t, os.WriteFile(refFile, []byte("# Patterns"), 0o644))

    tests := []struct {
        name     string
        filePath string
        content  string
        wantErr  bool
    }{
        {
            name:     "existing reference file",
            filePath: skillFile,
            content:  "## Resources\n\n- [`references/patterns.md`](references/patterns.md) — patterns\n",
            wantErr:  false,
        },
        {
            name:     "missing reference file",
            filePath: skillFile,
            content:  "## Resources\n\n- [`references/missing.md`](references/missing.md) — does not exist\n",
            wantErr:  true,
        },
        {
            name:     "no resources section",
            filePath: skillFile,
            content:  "## Implementation\n\n### Core\n\nDo this.\n",
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            ctx := &ValidationContext{
                FilePath: tt.filePath,
                Meta:     &SkillMeta{Name: "test-skill", FilePath: tt.filePath},
                Content:  tt.content,
                Lines:    splitLines(tt.content),
            }
            issues := validateResourceLinksV4(ctx)
            if tt.wantErr {
                assert.True(t, hasSeverity(issues, SeverityError), "expected error")
            } else {
                assert.Empty(t, issues)
            }
        })
    }
}
```

Also add `require` import to test file: `"github.com/stretchr/testify/require"` and `"os"`, `"path/filepath"`.

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestValidateResourceLinks -v
```
Expected: compile error — `validateResourceLinksV4` undefined

**Step 3: Implement in `rules.go`**

Add imports `"os"` and `"path/filepath"` if not already present. Add at bottom:

```go
var resourceLinkPattern = regexp.MustCompile(`\[.*?\]\((.*?)\)`)

// validateResourceLinksV4 checks that all links in ## Resources section point to existing files.
func validateResourceLinksV4(ctx *ValidationContext) []ValidationIssue {
    if ctx.FilePath == "" {
        return nil
    }

    resourcesSection := extractSectionFromContent(ctx.Content, "Resources")
    if resourcesSection == "" {
        return nil
    }

    skillDir := filepath.Dir(ctx.FilePath)
    var issues []ValidationIssue

    for _, match := range resourceLinkPattern.FindAllStringSubmatch(resourcesSection, -1) {
        link := match[1]
        if strings.HasPrefix(link, "http") {
            continue
        }
        target := filepath.Join(skillDir, link)
        if _, err := os.Stat(target); os.IsNotExist(err) {
            issues = append(issues, ValidationIssue{
                Rule:       "resource-links",
                Severity:   SeverityError,
                Message:    fmt.Sprintf("resource link %q points to missing file", link),
                Suggestion: "Create the file or remove the broken link",
                Example:    fmt.Sprintf("mkdir -p %s && touch %s", filepath.Dir(target), target),
            })
        }
    }

    return issues
}

// extractSectionFromContent extracts a markdown section by name.
func extractSectionFromContent(content, section string) string {
    heading := "## " + section
    idx := strings.Index(content, heading)
    if idx == -1 {
        return ""
    }
    after := content[idx+len(heading):]
    if next := strings.Index(after, "\n## "); next != -1 {
        return after[:next]
    }
    return after
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestValidateResourceLinks -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/skill/rules.go internal/skill/validator_test.go
git commit -m "feat(skill): add validateResourceLinksV4 to verify resources section links exist"
```

---

### Task 9: Add transition-mode validator and wire new rules

The `NewValidator()` stays unchanged (backward compat). `NewTransitionValidator()` adds new rules and converts old rules to warnings. After all 70 skills are migrated, `NewValidator()` will be updated to use strict rules.

**Files:**
- Modify: `internal/skill/validator.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write failing tests**

Add to `internal/skill/validator_test.go`:

```go
func TestNewTransitionValidator_NewRulesEnforced(t *testing.T) {
    t.Parallel()

    // Missing status should be an error in transition validator
    meta := &SkillMeta{
        Name:         "test-skill",
        Description:  "Spec-first API design — no trigger signal here",
        Triggers:     []string{"test"},
        Role:         "Expert Go developer focused on clean architecture and best practices.",
        Instructions: "### Core\n\nDo things following standard patterns and best practices here.",
        Examples:     "### Example 1: Basic\n\n**Input**: test\n\n**Output**: result",
        Status:       "",
    }

    v := NewTransitionValidator()
    result := v.Validate(meta, "")

    ruleNames := make([]string, 0, len(result.Issues))
    for _, i := range result.Issues {
        ruleNames = append(ruleNames, i.Rule)
    }

    assert.Contains(t, ruleNames, "status", "should flag missing status")
    assert.Contains(t, ruleNames, "description-quality", "should flag missing trigger signal")
}

func TestNewTransitionValidator_OldRoleRuleIsWarning(t *testing.T) {
    t.Parallel()

    // A new-format skill (no ## Role) should get a WARNING not an error in transition mode
    meta := &SkillMeta{
        Name:           "test-skill",
        Description:    "This skill should be used when the user asks to test things.",
        Status:         "production",
        Role:           "",   // no role section
        Instructions:   "",   // no old instructions
        Overview:       "Test overview content with enough words to be meaningful.",
        WhenToUse:      "Use when testing.",
        Implementation: "### Core\n\nWrite tests.",
    }

    v := NewTransitionValidator()
    result := v.Validate(meta, "")

    for _, i := range result.Issues {
        if i.Rule == "role-section" {
            assert.Equal(t, SeverityWarning, i.Severity, "role-section should be a warning in transition mode, not error")
        }
    }
    // Must still be valid (no errors from role absence)
    assert.True(t, result.Valid)
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/skill/... -run TestNewTransitionValidator -v
```
Expected: compile error — `NewTransitionValidator` undefined

**Step 3: Implement `NewTransitionValidator` in `validator.go`**

Add after `NewValidator()`:

```go
// transitionRoleSectionV4 warns (rather than errors) when ## Role section is missing.
// Used during migration from old format to new format.
func transitionRoleSectionV4(ctx *ValidationContext) []ValidationIssue {
    issues := validateRoleSectionV4(ctx)
    for i := range issues {
        if issues[i].Severity == SeverityError {
            issues[i].Severity = SeverityWarning
            issues[i].Message = "(transition) " + issues[i].Message
        }
    }
    return issues
}

// transitionInstructionsSectionV4 warns (rather than errors) on ## Instructions absence.
func transitionInstructionsSectionV4(ctx *ValidationContext) []ValidationIssue {
    issues := validateInstructionsSectionV4(ctx)
    for i := range issues {
        if issues[i].Severity == SeverityError {
            issues[i].Severity = SeverityWarning
            issues[i].Message = "(transition) " + issues[i].Message
        }
    }
    return issues
}

// transitionExamplesSectionV4 warns (rather than errors) on ## Examples absence.
func transitionExamplesSectionV4(ctx *ValidationContext) []ValidationIssue {
    issues := validateExamplesSectionV4(ctx)
    for i := range issues {
        if issues[i].Severity == SeverityError {
            issues[i].Severity = SeverityWarning
            issues[i].Message = "(transition) " + issues[i].Message
        }
    }
    return issues
}

// NewTransitionValidator creates a validator used during migration.
// Old-format rules (Role, Instructions, Examples) emit warnings instead of errors.
// New rules (status, description-quality, empty-placeholders, name-matches-dir, resource-links)
// are enforced as errors. After all skills are migrated, replace NewValidator() with this rule set.
func NewTransitionValidator() *Validator {
    return &Validator{
        rules: []ValidationRule{
            validateFrontmatterV4,
            validateNameFormatV4,
            validateNameMatchesDirV4,
            validateStatusV4,
            validateDescriptionQualityV4,
            validateEmptyPlaceholdersV4,
            validateResourceLinksV4,
            transitionRoleSectionV4,
            transitionInstructionsSectionV4,
            transitionExamplesSectionV4,
            validateReferencesV4,
        },
    }
}
```

**Step 4: Run to verify it passes**

```bash
go test ./internal/skill/... -run TestNewTransitionValidator -v
```
Expected: `PASS`

**Step 5: Run full suite**

```bash
go test ./internal/skill/... -v 2>&1 | tail -20
```
Expected: all tests pass

**Step 6: Commit**

```bash
git add internal/skill/validator.go internal/skill/validator_test.go
git commit -m "feat(skill): add NewTransitionValidator with new rules enforced, old rules as warnings"
```

---

### Task 10: Add `standard` canonical template and update wizard

**Files:**
- Create: `pkg/templates/standard/template.md`
- Create: `pkg/templates/README.md`
- Modify: `internal/cli/skill/new.go` — add trigger phrases prompt to `PromptMetadata`

**Step 1: Create canonical template**

Create `pkg/templates/standard/template.md`:

```markdown
---
name: ${SKILL_NAME}
description: >
  This skill should be used when the user asks to ${TRIGGER_PHRASE_1},
  ${TRIGGER_PHRASE_2}, or mentions ${TRIGGER_KEYWORD_1}, ${TRIGGER_KEYWORD_2}.
status: draft
---

# ${SKILL_TITLE}

## Overview
${ONE_PARAGRAPH_WHAT_AND_WHY}

## When to Use
- ${USE_CASE_1}
- ${USE_CASE_2}
- ${USE_CASE_3}

**Not for:** ${ANTI_USE_CASE} — use [${OTHER_SKILL}](../${OTHER_SKILL}/) instead.

## Quick Reference
| Task | Approach |
|------|----------|
| ${TASK_1} | ${APPROACH_1} |
| ${TASK_2} | ${APPROACH_2} |

## Implementation

### ${SUBSECTION_1}
${IMPERATIVE_INSTRUCTIONS}

\`\`\`go
${CODE_EXAMPLE}
\`\`\`

### ${SUBSECTION_2}
${IMPERATIVE_INSTRUCTIONS}

## Common Mistakes
- **${MISTAKE_1}**: ${FIX_1}
- **${MISTAKE_2}**: ${FIX_2}

## Resources
- [`references/patterns.md`](references/patterns.md) — ${PATTERNS_DESCRIPTION}
- [`examples/`](examples/) — ${EXAMPLES_DESCRIPTION}
```

**Step 2: Create template variable documentation**

Create `pkg/templates/README.md`:

```markdown
# Skill Templates

## Available Templates

| Template | Purpose |
|---|---|
| `standard` | New skills following official Claude Code schema |
| `go-basic` | Go language skills (lightweight) |
| `go-complete` | Go language skills (comprehensive) |
| `api-design` | API design skills |
| `arch` | Architecture skills |
| `testing` | Testing/QA skills |
| `security` | Security skills |

## Template Variables (standard template)

| Variable | Required | Description |
|---|---|---|
| `${SKILL_NAME}` | Yes | kebab-case, must match directory name |
| `${SKILL_TITLE}` | Yes | Title case display name |
| `${TRIGGER_PHRASE_1}` | Yes | Verb phrase: "design a REST API", "debug a goroutine leak" |
| `${TRIGGER_PHRASE_2}` | Yes | Second verb trigger |
| `${TRIGGER_KEYWORD_1}` | Yes | Noun keyword: "OpenAPI", "gRPC" |
| `${TRIGGER_KEYWORD_2}` | Yes | Second noun keyword |
| `${ONE_PARAGRAPH_WHAT_AND_WHY}` | Yes | Imperative prose, ≤150 words |
| `${ANTI_USE_CASE}` | Recommended | What this skill is NOT for |
| `${OTHER_SKILL}` | Recommended | Alternative skill for the anti-use case |

## Writing Effective Descriptions

Descriptions are the primary activation signal. Rules:
1. Third person: "This skill should be used when..."
2. Include specific trigger phrases (verb form: "design a...", "debug a...")
3. Include trigger keywords (noun form: "OpenAPI", "goroutine")
4. Do NOT summarise the workflow (Claude may shortcut)
5. Max 1024 characters

Bad: "Spec-first API design with OpenAPI and ogen."
Good: "This skill should be used when the user asks to design a REST API spec-first, generate Go server code from OpenAPI, work with ogen, or define gRPC services."
```

**Step 3: Find `PromptMetadata` in wizard**

```bash
grep -n "PromptMetadata\|PromptTemplate" /home/zhuk/Projects/own/go-ent/internal/cli/skill/*.go
```

Read the file containing `PromptMetadata` to understand its signature before modifying.

**Step 4: Update `PromptMetadata` to prompt for trigger phrases**

In `internal/cli/skill/wizard.go` (or wherever `PromptMetadata` lives), add prompts for:
- "Trigger phrase 1 (verb form, e.g. 'design a REST API'):"
- "Trigger phrase 2:"
- "Trigger keyword 1 (noun, e.g. 'OpenAPI'):"
- "Anti-use case (what this skill is NOT for):"

Store in `WizardConfig.TriggerPhrase1`, `TriggerPhrase2`, `TriggerKeyword1`, `AntiUseCase` and substitute into the standard template.

**Step 5: Run tests**

```bash
go test ./internal/cli/skill/... -v 2>&1 | tail -20
```
Expected: all pass

**Step 6: Commit**

```bash
git add pkg/templates/standard/ pkg/templates/README.md internal/cli/skill/
git commit -m "feat(template): add standard canonical template and wizard trigger phrase prompts"
```

---

## Phase 2 — Skill Content Migration (Tasks 11–15)

Tasks 11–15 are content-only changes. They can be executed in parallel by separate subagents once Phase 1 is done. Each subagent should run `go test ./internal/skill/...` and `go ent skill validate` after each batch to verify zero errors.

**Validator to use during migration:** `NewTransitionValidator()` — old sections are warnings, new rules are enforced.

---

### Task 11: Migrate Batch 1 — Stubs (10 skills)

**Target skills:** All skills with Tier 4 designation (agent-patterns, microservices, message-queues, mcp-development, prompt-engineering, graphql, grpc, and others with <50 lines)

**Per-skill steps:**
1. Add `status: draft` to frontmatter
2. Remove `triggers:` array from frontmatter
3. Rewrite `description` to third-person trigger-specific style (≤200 chars, include "when")
4. Rename `## Role` → `## Overview` (merge content into one paragraph)
5. Rename `## Instructions` → `## Implementation`
6. Remove `## Examples` section (stubs have no real examples; section removed cleanly)
7. Add minimal `## When to Use` section (3 bullets)
8. Run: `go ent skill validate pkg/skills/agent/ pkg/skills/backend/` — zero errors

**Example before (agent-patterns):**

```yaml
---
name: agent-patterns
description: AI agent orchestration patterns, multi-agent systems, agent communication
triggers:
  - agent pattern
  - orchestration
  - multi-agent
---

## Role

You are an expert in designing and orchestrating AI agent systems.

## Instructions

- Agent factory pattern
- Message queue for communication
- State machine for agent states
```

**Example after (agent-patterns):**

```yaml
---
name: agent-patterns
description: >
  This skill should be used when the user asks to design a multi-agent system,
  orchestrate AI agents, or implement agent communication patterns.
status: draft
---

## Overview

Patterns for designing AI agent systems, including orchestration, communication, and lifecycle management.

## When to Use

- Designing a multi-agent workflow
- Choosing a communication pattern between agents (message queue, direct call, pub/sub)
- Implementing a supervisor or factory pattern for agents

**Not for:** single-agent tool use or prompt engineering — use prompt-engineering for that.

## Implementation

### Core Patterns

- **Factory**: create agents dynamically from a specification
- **Supervisor**: parent agent restarts failed child agents
- **State machine**: model agent lifecycle with explicit states
- **Message queue**: decouple agents via async message passing

## Common Mistakes

- **Tight coupling between agents**: use message queues or events, not direct calls
- **No liveness detection**: implement heartbeat or deadline monitoring

## Resources

See references/ when this skill is promoted to production status.
```

**Commit after each category:**

```bash
git add pkg/skills/agent/ pkg/skills/backend/
git commit -m "chore(skills): migrate batch 1 stub skills to official schema (status: draft)"
```

---

### Task 12: Migrate Batch 2 — lang/ (17 skills)

**Target:** All 17 lang/ skills (flutter, python-core, python-django, python-fastapi, react-core, react-nextjs, rust-core, rust-async, rust-gtk4, rust-web, nodejs-backend, nodejs-nestjs, php-core, php-laravel, tailwind, typescript, vue, web-design)

**Per-skill steps (same as Batch 1 but also fix `rust-async` description length):**
1. Remove `triggers:`, add `status: draft` (most lang skills) or `status: production` (rust-async, typescript)
2. Rewrite description — third-person, trigger-specific, ≤200 chars
3. Rename sections: `## Role` → `## Overview`, `## Instructions` → `## Implementation`
4. Add `## When to Use`
5. Keep body content mostly as-is (don't rewrite technical content, just restructure)
6. Run: `go ent skill validate pkg/skills/lang/` — zero errors

**rust-async special case** — description was 250+ chars, rewrite to:
```yaml
description: >
  This skill should be used when the user asks to implement async Rust with Tokio,
  use channels or select!, spawn tasks, or implement graceful shutdown patterns.
```

**Commit:**

```bash
git add pkg/skills/lang/
git commit -m "chore(skills): migrate batch 2 lang/ skills to official schema"
```

---

### Task 13: Migrate Batch 3 — core/ (22 skills)

**Target:** All 22 core/ skills. Most are Tier 2–3, with production-level content.

**Special cases:**
- `debug-core`: Fix P0 blocker — remove empty `<example><input></input><output></output></example>` tags, add real debugging workflow in `## Implementation`
- `tdd`: Expand from 40 lines to 80+ lines with concrete red-green-refactor example
- `arch-core`, `security-core`, `planning`: Add `references/` directory with heavy content moved out

**Per-skill steps:**
1. Remove `triggers:`, add `status: production` (Tier 1–2) or `status: draft` (Tier 3 stubs)
2. Rewrite description
3. Full section restructure (`## Role` → `## Overview`, add `## When to Use`, `## Common Mistakes`)
4. Imperative form throughout
5. For Tier 1–2 with >1500 words: create `references/` subdirectory, move heavy content

**debug-core fix:**
Remove the entire `## Examples` section with its empty tags. Add to `## Implementation`:

```markdown
### Debugging Workflow

Given a bug report:

1. **Reproduce** — write a failing test or script that triggers the bug reliably
2. **Isolate** — binary search: comment half the code path, identify the minimal reproduction
3. **Hypothesise** — form one specific hypothesis about root cause
4. **Verify** — either the fix makes the test pass, or you reject the hypothesis
5. **Fix** — implement minimal change, verify all related tests pass

Never skip step 1. A bug you can't reproduce is a bug you can't fix.
```

**Commit (by sub-batch):**

```bash
git add pkg/skills/core/
git commit -m "chore(skills): migrate batch 3 core/ skills to official schema, fix debug-core P0"
```

---

### Task 14: Migrate Batch 4 — infra/ + qa/ (11 skills)

**Target:** infra/ (database-design, docker-devops, observability, performance, postgresql, redis) and qa/ (qa-api, qa-browser, qa-flutter, qa-performance, qa-visual)

**Special cases:**
- `postgresql`: Dense, excellent content — add `references/` for extended SQL patterns
- `qa-api`: Strong content, add `examples/` with sample Hurl test file

**Per-skill steps:** Same as Batch 3.

**Commit:**

```bash
git add pkg/skills/infra/ pkg/skills/qa/
git commit -m "chore(skills): migrate batch 4 infra/ and qa/ skills to official schema"
```

---

### Task 15: Migrate Batch 5 — go/ + ent/ (20 skills, full treatment)

**Target:** All 17 go/ skills + 3 ent/ skills. These are Tier 1–2, highest quality, most used.

**Full treatment for each:**
1. All previous steps
2. Create `references/` subdirectory with detailed patterns moved from body
3. Create `examples/` with at least one working runnable file

**Example for go-arch:**
- `references/adr-template.md` — full ADR template with example
- `references/ddd-patterns.md` — aggregate, value object, domain event examples
- `examples/clean-arch-layout.md` — annotated directory structure

**Example for ent-foundation:**
- `references/decision-framework.md` — full decision-making framework
- `references/naming-conventions.md` — complete naming guide with examples

**Example for go-perf:**
- `references/profiling-guide.md` — pprof workflow, flamegraph interpretation
- `examples/benchmark_test.go` — working benchmark with realistic load

**Commit:**

```bash
git add pkg/skills/go/ pkg/skills/ent/
git commit -m "chore(skills): migrate batch 5 go/ and ent/ skills to official schema with references/ and examples/"
```

---

## Phase 3 — Lockdown (Tasks 16–17)

### Task 16: Promote `NewTransitionValidator` to `NewValidator`

After all 70 skills pass `NewTransitionValidator` with zero errors, update `NewValidator` to use the new rule set.

**Files:**
- Modify: `internal/skill/validator.go`
- Modify: `internal/skill/validator_test.go`

**Step 1: Write test verifying all real skills pass new validator**

Add to `internal/skill/validator_test.go`:

```go
func TestAllSkillsPassNewValidator(t *testing.T) {
    t.Parallel()

    p := NewParser()
    v := NewValidator() // will use new rules after Task 16

    err := filepath.WalkDir("../../../pkg/skills", func(path string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() || d.Name() != "SKILL.md" {
            return err
        }
        t.Run(path, func(t *testing.T) {
            t.Parallel()
            meta, parseErr := p.ParseSkillFile(path)
            assert.NoError(t, parseErr, "should parse without error")
            if parseErr != nil {
                return
            }
            content, _ := os.ReadFile(path)
            result := v.Validate(meta, string(content))
            assert.True(t, result.Valid, "skill %s should be valid: %v", path, result.Issues)
            assert.Zero(t, result.ErrorCount(), "skill %s should have no errors", path)
        })
        return nil
    })
    assert.NoError(t, err)
}
```

**Step 2: Run to verify it fails** (skills not yet all migrated or validator not yet updated)

```bash
go test ./internal/skill/... -run TestAllSkillsPassNewValidator -v 2>&1 | grep FAIL | head -20
```

**Step 3: Update `NewValidator()` to use new rule set**

In `internal/skill/validator.go`, replace `NewValidator()` body:

```go
func NewValidator() *Validator {
    return &Validator{
        rules: []ValidationRule{
            validateFrontmatterV4,
            validateNameFormatV4,
            validateNameMatchesDirV4,
            validateStatusV4,
            validateDescriptionQualityV4,
            validateEmptyPlaceholdersV4,
            validateResourceLinksV4,
            validateReferencesV4,
        },
    }
}
```

Remove the three `transition*` wrappers and `NewTransitionValidator()` entirely.

**Step 4: Run full suite**

```bash
go test ./... -v 2>&1 | grep -E "^(ok|FAIL|---)" | tail -30
```
Expected: all pass

**Step 5: Commit**

```bash
git add internal/skill/validator.go internal/skill/validator_test.go
git commit -m "feat(skill): promote new validator rules to default, remove transition mode"
```

---

### Task 17: Rename 6 skills (directory renames)

**Renames:**

| Old path | New path |
|---|---|
| `pkg/skills/core/arch-core/` | `pkg/skills/core/architecture/` |
| `pkg/skills/core/debug-core/` | `pkg/skills/core/debugging/` |
| `pkg/skills/core/security-core/` | `pkg/skills/core/security/` |
| `pkg/skills/lang/flutter/` | `pkg/skills/lang/flutter-core/` |
| `pkg/skills/lang/vue/` | `pkg/skills/lang/vue-core/` |
| `pkg/skills/lang/tailwind/` | `pkg/skills/lang/tailwind-css/` |

**Step 1: Rename directories and update `name:` fields**

```bash
cd pkg/skills/core
mv arch-core architecture
mv debug-core debugging
mv security-core security

cd ../lang
mv flutter flutter-core
mv vue vue-core
mv tailwind tailwind-css
```

Update `name:` in each `SKILL.md` to match new directory name.

**Step 2: Find and update any cross-references**

```bash
rg "arch-core\|debug-core\|security-core\|flutter\b\|vue\b\|tailwind\b" pkg/skills/ --include="*.md" -l
```

Update any `[text](../arch-core/)` style links in the found files.

**Step 3: Run validator on renamed skills**

```bash
go ent skill validate pkg/skills/core/architecture/ pkg/skills/core/debugging/ pkg/skills/core/security/
```
Expected: zero errors (name now matches directory)

**Step 4: Run full test suite**

```bash
go test ./... 2>&1 | tail -10
```
Expected: all pass

**Step 5: Commit**

```bash
git add pkg/skills/
git commit -m "refactor(skills): rename 6 skills to match category naming conventions"
```

---

## Verification Checklist

After all 17 tasks complete:

- [ ] `go test ./...` passes with zero failures
- [ ] `go ent skill validate pkg/skills/` reports zero errors across all 70 skills
- [ ] `go ent skill validate pkg/skills/` reports ≤5 warnings total (only legitimate ones)
- [ ] `pkg/templates/standard/template.md` exists and renders correctly
- [ ] `pkg/templates/README.md` documents all variables
- [ ] All 6 renames complete: old paths no longer exist
- [ ] `debug-core` (now `debugging`) has no empty XML placeholder tags
- [ ] `ent-foundation`, `go-arch`, `go-perf` have `references/` directories
- [ ] `go-api` has `examples/` directory with working example
