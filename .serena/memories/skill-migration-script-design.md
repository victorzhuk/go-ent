# Skill Migration Script Design: v3 → v4

## Overview

One-time conversion tool to migrate skills from v3 format to v4 minimal format.

## Command Interface

```bash
# Convert single skill
ent skill migrate v3-to-v4 --input path/to/SKILL.md --output path/to/output/

# Dry run (show changes without writing)
ent skill migrate v3-to-v4 --input path/to/SKILL.md --dry-run

# Convert all skills in directory
ent skill migrate v3-to-v4 --all --skills-dir ./pkg/skills

# Create backups
ent skill migrate v3-to-v4 --all --backup --skills-dir ./pkg/skills

# Verbose output
ent skill migrate v3-to-v4 --input path/to/SKILL.md -v
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Migration Pipeline                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   Detect     │───▶│   Parse      │───▶│  Transform   │      │
│  │   Version    │    │   v3 Content │    │   to v4      │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│         │                   │                   │               │
│         ▼                   ▼                   ▼               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │ Error if not │    │ Extract:     │    │ Apply rules: │      │
│  │ v3 or v2     │    │ - frontmatter│    │ - flatten    │      │
│  │              │    │ - sections   │    │   triggers   │      │
│  │              │    │ - xml tags   │    │ - move heavy │      │
│  │              │    │              │    │   sections   │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│                                                 │               │
│                                                 ▼               │
│                                        ┌──────────────┐        │
│                                        │   Generate   │        │
│                                        │   Output     │        │
│                                        └──────────────┘        │
│                                                 │               │
│                                                 ▼               │
│                                        ┌──────────────┐        │
│                                        │   Validate   │        │
│                                        │   v4 Format  │        │
│                                        └──────────────┘        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Version Detector

```go
type VersionDetector struct{}

func (d *VersionDetector) Detect(content string) (string, error) {
    // Check for v3 markers:
    // - triggers: in frontmatter (not just triggers: list)
    // - ## Role, ## Instructions (Markdown sections)
    // - No XML tags like <role>, <instructions>
    
    // Check for v2 markers:
    // - <role>, <instructions>, <examples> XML tags
    // - triggers might be in body or frontmatter
    
    // Check for v1 markers:
    // - Simple frontmatter (name, description only)
    // - No triggers field
    
    return "v3" | "v2" | "v1" | "unknown", nil
}
```

### 2. v3 Parser

```go
type V3Parser struct{}

type V3Skill struct {
    // Frontmatter
    Name         string
    Description  string
    Version      string
    Author       string
    Tags         []string
    Triggers     V3Triggers
    Category     string
    QualityScore int
    
    // Body sections
    Role         string
    Instructions string
    Constraints  string
    EdgeCases    string
    Examples     string
    OutputFormat string
}

type V3Triggers struct {
    Keywords    []string
    FilePattern string
    Weight      float64
}

func (p *V3Parser) Parse(content string) (*V3Skill, error) {
    // 1. Extract frontmatter (YAML between ---)
    // 2. Parse YAML into V3Skill fields
    // 3. Extract Markdown sections (## Role, ## Instructions, etc)
    // 4. Return populated V3Skill
}
```

### 3. Transformer

```go
type Transformer struct {
    // Configuration
    MoveThreshold int  // Min items to move to references (default: 5)
    MaxRefDepth     int  // Max depth for references (default: 1)
}

type V4Skill struct {
    Name         string
    Description  string
    Triggers     []string
    Role         string
    Instructions string
    Examples     string
    References   []Reference
}

type Reference struct {
    Path    string  // e.g., "references/constraints.md"
    Title   string
    Content string
}

func (t *Transformer) Transform(v3 *V3Skill) (*V4Skill, error) {
    v4 := &V4Skill{
        Name:         v3.Name,
        Description:  cleanDescription(v3.Description),
        Triggers:     flattenTriggers(v3.Triggers),
        Role:         v3.Role,
        Instructions: mergeInstructions(v3.Instructions, v3.OutputFormat),
        Examples:     v3.Examples,
    }
    
    // Handle Constraints
    if shouldMoveToReferences(v3.Constraints, t.MoveThreshold) {
        v4.References = append(v4.References, Reference{
            Path:    "references/constraints.md",
            Title:   "Constraints",
            Content: v3.Constraints,
        })
    } else if v3.Constraints != "" {
        // Append to Instructions
        v4.Instructions += "\n\n### Constraints\n\n" + v3.Constraints
    }
    
    // Handle Edge Cases
    if shouldMoveToReferences(v3.EdgeCases, t.MoveThreshold) {
        v4.References = append(v4.References, Reference{
            Path:    "references/edge-cases.md",
            Title:   "Edge Cases",
            Content: v3.EdgeCases,
        })
    }
    
    return v4, nil
}

func flattenTriggers(v3 V3Triggers) []string {
    // Return keywords only
    // Remove file_pattern (category inference)
    // Remove weight (computed at runtime)
    return v3.Keywords
}

func cleanDescription(desc string) string {
    // Remove "Auto-activates for:" prefix and content
    // Truncate to 256 chars if needed
}

func shouldMoveToReferences(content string, threshold int) bool {
    // Count items (bullet points, paragraphs, etc)
    // Return true if count >= threshold
}
```

### 4. Output Generator

```go
type OutputGenerator struct{}

func (g *OutputGenerator) Generate(skill *V4Skill) (map[string]string, error) {
    files := make(map[string]string)
    
    // Generate SKILL.md
    var sb strings.Builder
    
    // Frontmatter
    sb.WriteString("---\n")
    sb.WriteString(fmt.Sprintf("name: %s\n", skill.Name))
    sb.WriteString(fmt.Sprintf("description: %s\n", skill.Description))
    sb.WriteString("triggers:\n")
    for _, t := range skill.Triggers {
        sb.WriteString(fmt.Sprintf("  - %s\n", t))
    }
    sb.WriteString("---\n\n")
    
    // Role
    sb.WriteString("## Role\n\n")
    sb.WriteString(skill.Role)
    sb.WriteString("\n\n")
    
    // Instructions
    sb.WriteString("## Instructions\n\n")
    sb.WriteString(skill.Instructions)
    sb.WriteString("\n\n")
    
    // Examples
    sb.WriteString("## Examples\n\n")
    sb.WriteString(skill.Examples)
    
    // References section (if any)
    if len(skill.References) > 0 {
        sb.WriteString("\n\n## References\n\n")
        for _, ref := range skill.References {
            sb.WriteString(fmt.Sprintf("- [%s](%s)\n", ref.Title, ref.Path))
        }
    }
    
    files["SKILL.md"] = sb.String()
    
    // Generate reference files
    for _, ref := range skill.References {
        var refContent strings.Builder
        refContent.WriteString(fmt.Sprintf("# %s\n\n", ref.Title))
        refContent.WriteString(ref.Content)
        files[ref.Path] = refContent.String()
    }
    
    return files, nil
}
```

### 5. Writer

```go
type Writer struct {
    OutputDir string
    Backup    bool
    DryRun    bool
}

func (w *Writer) Write(files map[string]string) error {
    if w.DryRun {
        // Print what would be written
        for path, content := range files {
            fmt.Printf("=== %s ===\n%s\n\n", path, content[:min(500, len(content))])
        }
        return nil
    }
    
    if w.Backup {
        // Create backup of existing files
        backupDir := fmt.Sprintf("%s.backup.%d", w.OutputDir, time.Now().Unix())
        // Copy existing to backup
    }
    
    // Write files
    for path, content := range files {
        fullPath := filepath.Join(w.OutputDir, path)
        dir := filepath.Dir(fullPath)
        
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("create dir %s: %w", dir, err)
        }
        
        if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
            return fmt.Errorf("write %s: %w", path, err)
        }
    }
    
    return nil
}
```

## Conversion Rules Reference

| v3 Element | v4 Action | Details |
|------------|-----------|---------|
| `name` | Keep | Validate format |
| `description` | Clean | Remove "Auto-activates for:" prefix, truncate to 256 chars |
| `version` | Remove | Not needed |
| `author` | Remove | Not needed |
| `license` | Remove | Not needed |
| `tags` | Remove | Not needed |
| `category` | Infer | From path: `skills/{category}/` |
| `quality_score` | Remove | Not needed |
| `compatibility` | Remove | Not needed |
| `triggers.keywords` | Flatten | `triggers: [item1, item2]` |
| `triggers.file_pattern` | Remove | Category inference |
| `triggers.weight` | Remove | Runtime computation |
| `## Role` | Keep | Required section |
| `## Instructions` | Keep | Required section |
| `## Constraints` | Conditional | Move to `references/constraints.md` if >5 items |
| `## Edge Cases` | Conditional | Move to `references/edge-cases.md` if >5 items |
| `## Examples` | Keep | Required section |
| `## Output Format` | Merge | Add as `### Response Format` under Instructions |
| `<role>` XML | Convert | Extract content to `## Role` |
| `<instructions>` XML | Convert | Extract content to `## Instructions` |
| `<constraints>` XML | Convert | Same as ## Constraints |
| `<examples>` XML | Convert | Extract to `## Examples` |
| `references/` | Validate | Check structure, no frontmatter |

## Error Handling

```go
type MigrationError struct {
    File    string
    Phase   string // "detect", "parse", "transform", "generate", "write"
    Message string
    Cause   error
}

func (e *MigrationError) Error() string {
    return fmt.Sprintf("migration failed for %s at phase %s: %s", e.File, e.Phase, e.Message)
}
```

## Reporting

```go
type MigrationReport struct {
    TotalFiles    int
    Successful    int
    Failed        int
    Skipped       int // Already v4 or unknown format
    
    Details []FileReport
}

type FileReport struct {
    SourcePath   string
    OutputPath   string
    Status       string // "success", "failed", "skipped"
    VersionFrom  string // "v3", "v2", "v1"
    Changes      []string // List of transformations applied
    Error        error
}

func (r *MigrationReport) Print() {
    fmt.Printf("Migration Report\n")
    fmt.Printf("================\n")
    fmt.Printf("Total:   %d\n", r.TotalFiles)
    fmt.Printf("Success: %d\n", r.Successful)
    fmt.Printf("Failed:  %d\n", r.Failed)
    fmt.Printf("Skipped: %d\n", r.Skipped)
    
    if r.Failed > 0 {
        fmt.Printf("\nFailed Files:\n")
        for _, d := range r.Details {
            if d.Status == "failed" {
                fmt.Printf("  - %s: %v\n", d.SourcePath, d.Error)
            }
        }
    }
}
```

## Testing Strategy

### Unit Tests

```go
func TestVersionDetector(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected string
    }{
        {
            name:     "v3 with markdown sections",
            content:  "---\nname: test\ntriggers:\n  keywords: [a, b]\n---\n\n## Role\n...",
            expected: "v3",
        },
        {
            name:     "v2 with xml tags",
            content:  "---\nname: test\n---\n\n<role>...</role>",
            expected: "v2",
        },
    }
    // ...
}

func TestTransformer(t *testing.T) {
    tests := []struct {
        name     string
        input    *V3Skill
        expected *V4Skill
    }{
        {
            name: "flatten triggers",
            input: &V3Skill{
                Triggers: V3Triggers{
                    Keywords:    []string{"a", "b"},
                    FilePattern: "*.go",
                    Weight:      0.8,
                },
            },
            expected: &V4Skill{
                Triggers: []string{"a", "b"},
            },
        },
    }
    // ...
}
```

### Integration Tests

```go
func TestMigrateEndToEnd(t *testing.T) {
    // Create temp dir with v3 skill
    // Run migration
    // Verify output matches expected v4 format
    // Validate with v4 validator
}
```

## Implementation File Structure

```
cmd/skill-convert/
├── main.go              # CLI entry point
├── detector.go          # Version detection
├── parser_v3.go         # v3 parsing
├── parser_v2.go         # v2 parsing (for reference)
├── transformer.go       # v3 → v4 transformation
├── generator.go         # v4 output generation
├── writer.go            # File writing with backup
└── report.go            # Migration reporting

internal/skill/migrate/
├── migrate.go           # Public API for programmatic use
└── migrate_test.go      # Tests
```

## Usage in OpenSpec Tasks

The migration script will be used in:

- **Task 9**: Create Skill Conversion Script
- **Task 10**: Convert All Existing Skills

```bash
# Task 9: Build the tool
go build -o bin/skill-convert ./cmd/skill-convert

# Task 10: Run conversion
./bin/skill-convert --all --skills-dir ./pkg/skills --backup

# Verify
./bin/skill-convert --validate --skills-dir ./pkg/skills
```
