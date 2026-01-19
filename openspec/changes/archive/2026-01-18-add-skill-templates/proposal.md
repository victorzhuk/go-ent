---
name: add-skill-templates
description: Add interactive CLI skill generator with multiple built-in templates, wizard prompts, and custom template support to streamline skill creation and ensure consistency across the go-ent plugin ecosystem.
version: 1.0.0
author: go-ent
tags: ["cli", "skills", "templates", "developer-experience"]
status: complete
---

# Proposal: Interactive Skill Template Generator

## Summary

Add an interactive CLI skill generator system that provides:
- **Template Generator**: `go-ent skill new <name>` command to scaffold skills from templates
- **Built-in Templates**: Pre-configured templates for different skill types (go, typescript, database, testing, api-design, core, debugging, security, review, arch)
- **Interactive Wizard**: Guided skill creation with prompts for metadata
- **Custom Templates**: Ability to add and manage custom templates
- **Auto-Validation**: Generated skills must pass validation before saving

## Why

### Current Pain Points

1. **Manual Skill Creation**: Developers must manually create skill files, copy examples, or reference existing skills
2. **Inconsistent Structure**: Each new skill may have different structure, missing sections, or formatting issues
3. **High Barrier to Entry**: New contributors must understand the v2 skill format, XML tags, frontmatter options, and validation rules
4. **No Standardization**: No reference starting point for different skill categories (core skills vs language-specific skills)
5. **Validation After Creation**: Developers write skills then discover validation errors through MCP tools rather than being guided upfront

### Goals

1. **Streamline Creation**: Reduce skill creation time from 30+ minutes to under 5 minutes
2. **Guarantee Validity**: Ensure generated skills pass all validation rules out of the box
3. **Best Practices**: Embed research-backed patterns into templates
4. **Flexibility**: Support both built-in and custom templates for different use cases
5. **Developer Experience**: Make skill authoring accessible to new contributors

## What Changes

### New CLI Commands

```bash
# Scaffold new skill with interactive wizard
go-ent skill new <name>

# List available built-in and custom templates
go-ent skill list-templates

# Add custom template from path
go-ent skill add-template <path>

# Show template details
go-ent skill show-template <name>
```

### Built-in Templates

Located at `plugins/go-ent/templates/skills/`:

1. **go-basic**: Minimal Go language skill
2. **go-complete**: Full-featured Go skill with all sections
3. **typescript-basic**: TypeScript language skill
4. **database**: Database operations skill (sql, migrations)
5. **testing**: Testing patterns and test-driven development
6. **api-design**: API design and OpenAPI/GraphQL skills
7. **core-basic**: Core domain skill (architecture, design patterns)
8. **debugging**: Debugging and troubleshooting skill
9. **security**: Security patterns and best practices
10. **review**: Code review quality checks
11. **arch**: Architecture and system design

### Template Structure

Each template contains:
- `template.md`: Skill template with placeholders
- `config.yaml`: Template metadata (category, default values, prompts)

**Example `plugins/go-ent/templates/skills/go-basic/config.yaml`**:
```yaml
name: go-basic
category: language
description: Minimal Go language skill template
author: go-ent
version: 1.0.0
prompts:
  - key: description
    prompt: "Brief description of what this skill does"
    default: "Go coding patterns and best practices"
  - key: triggers
    prompt: "Auto-activation keywords (comma-separated)"
    default: "go,golang"
```

### Interactive Wizard Flow

```bash
$ go-ent skill new my-skill

Select template:
  [1] go-basic          - Minimal Go language skill
  [2] go-complete       - Full-featured Go skill
  [3] typescript-basic  - TypeScript language skill
  [4] database         - Database operations
  [5] testing           - Testing patterns
  [6] api-design        - API design
  [7] core-basic        - Core domain skill
  [8] debugging         - Debugging patterns
  [9] security          - Security patterns
  [10] review           - Code review
  [11] arch             - Architecture design
  [12] custom           - Use custom template
  > 1

Skill name: my-go-patterns
Description (optional): Go patterns for microservices
Version (default: 1.0.0): 
Author (default: your-name): 
Tags (comma-separated, optional): go,microservices

Generating skill from go-basic template...
✓ Created: plugins/go-ent/skills/go/my-go-patterns/SKILL.md
✓ Validated: 0 errors, 0 warnings

Next steps:
- Edit plugins/go-ent/skills/go/my-go-patterns/SKILL.md
- Run: go-ent skill validate my-go-patterns
- Sync to Claude: make skill-sync
```

### Auto-Detection

Templates automatically detect:
- **Output directory**: Language skills go to `plugins/go-ent/skills/<category>/`, core skills to `plugins/go-ent/skills/core/`
- **Category**: Derived from template config or auto-detected from name prefix
- **File naming**: SKILL.md for the skill file

## Impact

### New Code

**CLI Package** (`internal/cli/skill/`):
- `new.go` - Skill new command implementation
- `templates.go` - Template listing and management
- `wizard.go` - Interactive prompt logic
- `validation.go` - Post-generation validation

**Template Package** (`internal/template/`):
- `loader.go` - Load templates from disk
- `parser.go` - Parse template config.yaml
- `generator.go` - Generate skill from template with placeholders
- `placeholder.go` - Placeholder replacement logic

**Build Process**:
- New `make skill-templates` target to build template binaries
- New `make test-templates` target for template testing

### Modified Code

**CLI Root** (`internal/cli/root.go`):
- Add `skill` subcommand group
- Register `new`, `list-templates`, `add-template`, `show-template` commands

**Existing Skills**:
- No changes to existing skills (additive only)
- Existing manual skill creation workflow still supported

### Documentation

- Update `docs/SKILL-AUTHORING.md` with template generator guide
- Add `docs/TEMPLATE-CREATION.md` for custom template authoring
- Update CLAUDE.md to reference template generator

### No Breaking Changes

- Existing skill creation workflow unchanged
- Backward compatible with manual skill authoring
- Existing MCP tools unchanged
- No configuration changes required

## Alternatives Considered

### 1. Static Templates Only (current approach)
- ❌ Requires manual copy-paste and editing
- ❌ No interactive guidance
- ❌ Easy to miss required sections
- ❌ No validation feedback during creation

### 2. Web-Based Generator
- ❌ Requires additional infrastructure
- ❌ Not aligned with CLI-first workflow
- ❌ Harder to integrate with go-ent build system
- ✅ Better visual preview (but not essential)

### 3. Language-Specific Generators (e.g., `go-ent go-skill new`)
- ❌ More commands to maintain
- ❌ Harder to share templates across languages
- ❌ Inconsistent CLI experience
- ✅ Shorter commands (but `skill new` is acceptable)

### 4. AI-Powered Generator (ask Claude to generate skill)
- ❌ Unreliable without structured guidance
- ❌ Requires validation cycle anyway
- ❌ Doesn't teach authors the format
- ✅ More flexible (but quality varies)

### 5. Chosen: Interactive CLI Generator with Templates
- ✅ Guided experience ensures completeness
- ✅ Immediate validation feedback
- ✅ Easy to learn format through templates
- ✅ Extensible with custom templates
- ✅ Aligned with CLI-first workflow

## Implementation Details

### Phase 1: CLI Infrastructure (4h)
- Implement `skill` command group in CLI
- Create template loader and parser
- Implement placeholder replacement engine
- Add post-generation validation

### Phase 2: Built-in Templates (6h)
- Create 11 built-in templates
- Each template includes:
  - Complete v2 skill structure
  - Appropriate role and instructions
  - 2-3 example input/output pairs
  - Relevant constraints and edge cases
  - Template config.yaml

### Phase 3: Interactive Wizard (3h)
- Implement prompt logic with validation
- Add template selection UI
- Implement auto-detection of category and output path
- Add progress indicators and success messages

### Phase 4: Custom Template Support (2h)
- Implement `add-template` command
- Validate custom templates
- Add `list-templates` with custom templates
- Add `show-template` for template details

### Phase 5: Testing (3h)
- Unit tests for template loading
- Integration tests for CLI commands
- Test all built-in templates generate valid skills
- Test custom template addition flow

### Phase 6: Documentation (2h)
- Update SKILL-AUTHORING.md
- Create TEMPLATE-CREATION.md
- Update AGENTS.md with new commands

**Total Estimate**: ~20 hours

### Template Quality Standards

Each built-in template must:
1. Pass all validation rules in strict mode
2. Have quality score >= 90
3. Include at least 2 examples
4. Handle at least 3 edge cases
5. Follow research-backed patterns from `docs/research/SKILL.md`

### Placeholder Syntax

Templates use `${PLACEHOLDER}` syntax for dynamic content:

```markdown
---
name: ${SKILL_NAME}
description: "${SKILL_DESCRIPTION}"
version: ${SKILL_VERSION}
author: ${SKILL_AUTHOR}
tags: ${SKILL_TAGS}
---

# ${SKILL_NAME}

<role>
${ROLE_DEFINITION}
</role>
```

**Placeholders**:
- `${SKILL_NAME}` - Derived from command argument or prompt
- `${SKILL_DESCRIPTION}` - User input or template default
- `${SKILL_VERSION}` - User input (default: "1.0.0")
- `${SKILL_AUTHOR}` - User input (default: git user)
- `${SKILL_TAGS}` - User input (default: template tags)
- `${ROLE_DEFINITION}` - Template-specific role
- Template-specific placeholders defined in config.yaml

## Success Criteria

### Phase 1: CLI Infrastructure ✅
- `go-ent skill new test-skill` creates skill file
- Template loads from disk without errors
- Placeholder replacement works correctly
- Validation runs after generation

### Phase 2: Built-in Templates ✅
- All 11 templates created
- Each template generates valid skill
- Quality score >= 90 for all templates

### Phase 3: Interactive Wizard ✅
- Prompts for missing required fields
- Template selection UI works
- Auto-detects category and output path
- Shows success message with next steps

### Phase 4: Custom Templates ✅
- `add-template` validates and registers custom template
- Custom templates appear in `list-templates`
- Custom templates generate valid skills

### Phase 5: Testing ✅
- Unit tests cover 80%+ code
- Integration tests verify end-to-end flow
- All built-in templates tested

### Phase 6: Documentation ✅
- SKILL-AUTHORING.md updated
- TEMPLATE-CREATION.md created
- AGENTS.md references new commands

## Dependencies

- Go prompt library: `github.com/AlecAivazis/survey/v2` (for interactive prompts)
- Existing skill parser: `internal/skill/parser.go`
- Existing validator: `internal/skill/validator.go`
- Existing template patterns from existing skills

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Template quality inconsistent | Medium | Quality score threshold (>=90), manual review |
| Wizard UX confusing | Low | User testing, clear prompts, help text |
| Custom template validation too strict | Low | Permissive validation mode for custom templates |
| Placeholder replacement bugs | Medium | Comprehensive unit tests, edge case coverage |
| Category auto-detection fails | Low | Fallback to manual selection option |

## Future Enhancements

1. **Template Marketplace**: Share community templates
2. **Skill Composition**: Combine multiple templates
3. **Versioned Templates**: Support multiple template versions
4. **Template Diff**: Show changes when updating templates
5. **AI-Assisted**: Suggest template based on description
6. **Template Linter**: Validate templates against best practices
