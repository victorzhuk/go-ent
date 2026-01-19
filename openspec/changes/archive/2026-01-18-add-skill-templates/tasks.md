# Tasks: Add Skill Templates

## Status: complete

## Phase 1: CLI Infrastructure (4h)

### 1.1 Create CLI skill command group
- [x] 1.1.1 Create `internal/cli/skill/` directory structure
- [x] 1.1.2 Implement `skill` command group in `internal/cli/skill/root.go`
- [x] 1.1.3 Register skill group in CLI root (`internal/cli/root.go`)
- [x] 1.1.4 Add help text and subcommand documentation

### 1.2 Implement template loader
- [x] 1.2.1 Create `internal/template/loader.go`
- [x] 1.2.2 Implement `LoadTemplates()` to scan `plugins/go-ent/templates/skills/`
- [x] 1.2.3 Implement `LoadTemplate(name)` for single template load
- [x] 1.2.4 Add error handling for missing/invalid templates
- [x] 1.2.5 Add tests for loader functionality

### 1.3 Implement template parser
- [x] 1.3.1 Create `internal/template/parser.go`
- [x] 1.3.2 Parse template `config.yaml` into `TemplateConfig` struct
- [x] 1.3.3 Extract template metadata (name, category, description, version)
- [x] 1.3.4 Parse prompts array from config
- [x] 1.3.5 Add validation for required config fields
- [x] 1.3.6 Add tests for parser with valid/invalid configs

### 1.4 Implement placeholder replacement engine
- [x] 1.4.1 Create `internal/template/generator.go`
- [x] 1.4.2 Implement `ReplacePlaceholders(template, data)` function
- [x] 1.4.3 Support `${PLACEHOLDER}` syntax
- [x] 1.4.4 Handle missing placeholders (keep as-is or error)
- [x] 1.4.5 Add default values for common placeholders
- [x] 1.4.6 Add tests for various placeholder scenarios

### 1.5 Implement post-generation validation
- [x] 1.5.1 Create `internal/cli/skill/validation.go`
- [x] 1.5.2 Integrate with existing skill validator (`internal/skill/validator.go`)
- [x] 1.5.3 Implement `ValidateGeneratedSkill(path)` function
- [x] 1.5.4 Return validation errors with line numbers
- [x] 1.5.5 Add tests for validation scenarios

## Phase 2: Built-in Templates (6h)

### 2.1 Create go-basic template
- [x] 2.1.1 Create `plugins/go-ent/templates/skills/go-basic/` directory
- [x] 2.1.2 Write `template.md` with v2 format
- [x] 2.1.3 Write `config.yaml` with metadata and prompts
- [x] 2.1.4 Validate generated skill passes strict mode
- [x] 2.1.5 Verify quality score >= 90
- [x] 2.1.6 Include 2 examples
- [x] 2.1.7 Handle 3+ edge cases

### 2.2 Create go-complete template
- [x] 2.2.1 Create `plugins/go-ent/templates/skills/go-complete/` directory
- [x] 2.2.2 Write comprehensive `template.md` with all sections
- [x] 2.2.3 Write `config.yaml` with extensive prompts
- [x] 2.2.4 Validate generated skill passes strict mode
- [x] 2.2.5 Verify quality score >= 90
- [x] 2.2.6 Include 3+ examples covering edge cases
- [x] 2.2.7 Handle 5+ edge cases (including delegation)

### 2.3 Create typescript-basic template
- [x] 2.3.1 Create `plugins/go-ent/templates/skills/typescript-basic/` directory
- [x] 2.3.2 Write `template.md` with TypeScript-specific patterns
- [x] 2.3.3 Write `config.yaml` with TS-specific prompts
- [x] 2.3.4 Validate generated skill passes strict mode
- [x] 2.3.5 Verify quality score >= 90
- [x] 2.3.6 Include 2 examples
- [x] 2.3.7 Handle 3+ edge cases

### 2.4 Create database template
- [x] 2.4.1 Create `plugins/go-ent/templates/skills/database/` directory
- [x] 2.4.2 Write `template.md` with SQL/migration patterns
- [x] 2.4.3 Write `config.yaml` with DB-specific prompts
- [x] 2.4.4 Validate generated skill passes strict mode
- [x] 2.4.5 Verify quality score >= 90
- [x] 2.4.6 Include 2 examples
- [x] 2.4.7 Handle 3+ edge cases

### 2.5 Create testing template
- [x] 2.5.1 Create `plugins/go-ent/templates/skills/testing/` directory
- [x] 2.5.2 Write `template.md` with TDD patterns
- [x] 2.5.3 Write `config.yaml` with testing-specific prompts
- [x] 2.5.4 Validate generated skill passes strict mode
- [x] 2.5.5 Verify quality score >= 90
- [x] 2.5.6 Include 2 examples
- [x] 2.5.7 Handle 3+ edge cases

### 2.6 Create api-design template
- [x] 2.6.1 Create `plugins/go-ent/templates/skills/api-design/` directory
- [x] 2.6.2 Write `template.md` with REST/GraphQL patterns
- [x] 2.6.3 Write `config.yaml` with API-specific prompts
- [x] 2.6.4 Validate generated skill passes strict mode
- [x] 2.6.5 Verify quality score >= 90
- [x] 2.6.6 Include 2 examples
- [x] 2.6.7 Handle 3+ edge cases

### 2.7 Create core-basic template
- [x] 2.7.1 Create `plugins/go-ent/templates/skills/core-basic/` directory
- [x] 2.7.2 Write `template.md` with domain/architecture patterns
- [x] 2.7.3 Write `config.yaml` with core-specific prompts
- [x] 2.7.4 Validate generated skill passes strict mode
- [x] 2.7.5 Verify quality score >= 90
- [x] 2.7.6 Include 2 examples
- [x] 2.7.7 Handle 3+ edge cases

### 2.8 Create debugging template
- [x] 2.8.1 Create `plugins/go-ent/templates/skills/debugging-basic/` directory
- [x] 2.8.2 Write `template.md` with troubleshooting patterns
- [x] 2.8.3 Write `config.yaml` with debugging-specific prompts
- [x] 2.8.4 Validate generated skill passes strict mode
- [x] 2.8.5 Verify quality score >= 90
- [x] 2.8.6 Include 2 examples
- [x] 2.8.7 Handle 3+ edge cases

### 2.9 Create security template
- [x] 2.9.1 Create `plugins/go-ent/templates/skills/security/` directory
- [x] 2.9.2 Write `template.md` with security patterns
- [x] 2.9.3 Write `config.yaml` with security-specific prompts
- [x] 2.9.4 Validate generated skill passes strict mode
- [x] 2.9.5 Verify quality score >= 90
- [x] 2.9.6 Include 2 examples
- [x] 2.9.7 Handle 3+ edge cases

### 2.10 Create review template
- [x] 2.10.1 Create `plugins/go-ent/templates/skills/review/` directory
- [x] 2.10.2 Write `template.md` with code review patterns
- [x] 2.10.3 Write `config.yaml` with review-specific prompts
- [x] 2.10.4 Validate generated skill passes strict mode
- [x] 2.10.5 Verify quality score >= 90
- [x] 2.10.6 Include 2 examples
- [x] 2.10.7 Handle 3+ edge cases

### 2.11 Create arch template
- [x] 2.11.1 Create `plugins/go-ent/templates/skills/arch/` directory
- [x] 2.11.2 Write `template.md` with architecture patterns
- [x] 2.11.3 Write `config.yaml` with arch-specific prompts
- [x] 2.11.4 Validate generated skill passes strict mode
- [x] 2.11.5 Verify quality score >= 90
- [x] 2.11.6 Include 2 examples
- [x] 2.11.7 Handle 3+ edge cases

## Phase 3: Interactive Wizard (3h)

### 3.1 Implement prompt system
- [x] 3.1.1 Create `internal/cli/skill/wizard.go`
- [x] 3.1.2 Implement `PromptTemplateSelection()` with list UI
- [x] 3.1.3 Implement `PromptMetadata()` for skill details
- [x] 3.1.4 Integrate `github.com/AlecAivazis/survey/v2` for interactive prompts
- [x] 3.1.5 Add validation for required inputs
- [x] 3.1.6 Add help text for each prompt

### 3.2 Implement template selection UI
- [x] 3.2.1 Format template list with names and descriptions
- [x] 3.2.2 Add numbering for easy selection
- [x] 3.2.3 Support search/filter by name or category
- [x] 3.2.4 Add "custom" option for custom templates
- [x] 3.2.5 Show template details on selection
- [x] 3.2.6 Add tests for selection logic

### 3.3 Implement auto-detection logic
- [x] 3.3.1 Detect output directory from template category
- [x] 3.3.2 Auto-detect category from skill name prefix (go-, typescript-, etc.)
- [x] 3.3.3 Fallback to manual category selection
- [x] 3.3.4 Determine file path: `plugins/go-ent/skills/<category>/<name>/SKILL.md`
- [x] 3.3.5 Add tests for various category detection scenarios

### 3.4 Implement skill generation flow
- [x] 3.4.1 Create output directory structure
- [x] 3.4.2 Load selected template
- [x] 3.4.3 Replace placeholders with user input
- [x] 3.4.4 Write generated skill to file
- [x] 3.4.5 Run validation on generated skill
- [x] 3.4.6 Display validation results
- [x] 3.4.7 Show success message with next steps
- [x] 3.4.8 Add tests for full generation flow

### 3.5 Implement `new` command
- [x] 3.5.1 Create `internal/cli/skill/new.go`
- [x] 3.5.2 Implement `NewSkillCmd` cobra command
- [x] 3.5.3 Handle `<name>` positional argument
- [x] 3.5.4 Integrate wizard for interactive prompts
- [x] 3.5.5 Add flags for non-interactive mode (`--template`, `--description`, etc.)
- [x] 3.5.6 Register command in skill group
- [x] 3.5.7 Add tests for command execution

## Phase 4: Custom Template Support (2h)

### 4.1 Implement `list-templates` command
- [x] 4.1.1 Create `internal/cli/skill/templates.go`
- [x] 4.1.2 Implement `ListTemplatesCmd` cobra command
- [x] 4.1.3 Display built-in templates
- [x] 4.1.4 Display custom templates
- [x] 4.1.5 Show template metadata (name, category, description)
- [x] 4.1.6 Add flags for filtering (`--category`, `--built-in`, `--custom`)
- [x] 4.1.7 Register command in skill group
- [x] 4.1.8 Add tests for listing scenarios

### 4.2 Implement `add-template` command
- [x] 4.2.1 Extend `internal/cli/skill/templates.go`
- [x] 4.2.2 Implement `AddTemplateCmd` cobra command
- [x] 4.2.3 Validate template path exists
- [x] 4.2.4 Validate template structure (template.md + config.yaml)
- [x] 4.2.5 Validate template passes validation
- [x] 4.2.6 Copy to `plugins/go-ent/templates/skills/` or user templates dir
- [x] 4.2.7 Register command in skill group
- [x] 4.2.8 Add tests for adding valid/invalid templates

### 4.3 Implement `show-template` command
- [x] 4.3.1 Extend `internal/cli/skill/templates.go`
- [x] 4.3.2 Implement `ShowTemplateCmd` cobra command
- [x] 4.3.3 Load and display template details
- [x] 4.3.4 Show template config metadata
- [x] 4.3.5 Show template preview (first 20 lines)
- [x] 4.3.6 Add tests for display scenarios

### 4.4 Add custom template directory support
- [x] 4.4.1 Support `~/.go-ent/templates/skills/` for user templates
- [x] 4.4.2 Scan both built-in and custom template directories
- [x] 4.4.3 Merge template lists with source indicators
- [x] 4.4.4 Add tests for mixed template sources

## Phase 5: Testing (3h)

### 5.1 Unit tests for template infrastructure
- [x] 5.1.1 Test `LoadTemplates()` with empty, single, multiple templates
- [x] 5.1.2 Test `LoadTemplate(name)` with valid/invalid names
- [x] 5.1.3 Test `ParseConfig()` with valid/invalid YAML
- [x] 5.1.4 Test `ReplacePlaceholders()` with various scenarios
- [x] 5.1.5 Test `ValidateGeneratedSkill()` with valid/invalid skills
- [x] 5.1.6 Test `DetectCategory()` with various names
- [x] 5.1.7 Achieve 80%+ code coverage

### 5.2 Integration tests for CLI commands
- [x] 5.2.1 Test `go-ent skill new` with valid name and template
- [x] 5.2.2 Test `go-ent skill new` with invalid template name
- [x] 5.2.3 Test `go-ent skill new` non-interactive mode with flags
- [x] 5.2.4 Test `go-ent skill list-templates` with filters
- [x] 5.2.5 Test `go-ent skill add-template` with valid template
- [x] 5.2.6 Test `go-ent skill add-template` with invalid template
- [x] 5.2.7 Test `go-ent skill show-template` for various templates
- [x] 5.2.8 Test file system operations (create, read, write)

### 5.3 Template validation tests
- [x] 5.3.1 Generate skills from all 11 built-in templates
- [x] 5.3.2 Validate all generated skills pass strict mode
- [x] 5.3.3 Verify quality score >= 90 for all templates
- [x] 5.3.4 Test placeholder replacement with edge cases
- [x] 5.3.5 Test with missing placeholders (should error)
- [x] 5.3.6 Test with extra placeholders (should keep as-is)

### 5.4 End-to-end tests
- [x] 5.4.1 Test complete workflow: select template → fill prompts → generate → validate
- [x] 5.4.2 Test with custom template added via `add-template`
- [x] 5.4.3 Test auto-detection of category for various skill names
- [x] 5.4.4 Test error handling (missing template, invalid config, validation failure)
- [x] 5.4.5 Test file already exists scenario (should error with clear message)

## Phase 6: Documentation (2h)

### 6.1 Update SKILL-AUTHORING.md
- [x] 6.1.1 Add "Creating Skills with Templates" section
- [x] 6.1.2 Document `go-ent skill new` command
- [x] 6.1.3 Document interactive wizard flow
- [x] 6.1.4 Document template selection
- [x] 6.1.5 Add example usage with screenshots
- [x] 6.1.6 Document non-interactive mode flags

### 6.2 Create TEMPLATE-CREATION.md
- [x] 6.2.1 Create `docs/TEMPLATE-CREATION.md`
- [x] 6.2.2 Document template structure (template.md + config.yaml)
- [x] 6.2.3 Explain config.yaml schema and options
- [x] 6.2.4 Document placeholder syntax
- [x] 6.2.5 Provide template creation best practices
- [x] 6.2.6 Include example custom template
- [x] 6.2.7 Document `go-ent skill add-template` usage

### 6.3 Update AGENTS.md
- [x] 6.3.1 Add skill commands to AGENTS.md
- [x] 6.3.2 Document `go-ent skill new`
- [x] 6.3.3 Document `go-ent skill list-templates`
- [x] 6.3.4 Document `go-ent skill add-template`
- [x] 6.3.5 Document `go-ent skill show-template`
- [x] 6.3.6 Update CLI command reference section

### 6.4 Add inline documentation
- [x] 6.4.1 Add godoc comments to all public functions
- [x] 6.4.2 Add usage examples in package docs
- [x] 6.4.3 Add comments explaining template design decisions
- [x] 6.4.4 Update CLI command help texts

### 6.5 Update Makefile
- [x] 6.5.1 Add `make test-templates` target
- [x] 6.5.2 Add `make validate-templates` target
- [x] 6.5.3 Update main Makefile with new targets

## Verification

### Pre-Merge Checklist
- [x] All tasks in Phases 1-6 completed
- [x] Unit tests pass with 80%+ coverage (template package: 86.5%)
- [x] Integration tests pass
- [x] All built-in templates generate valid skills (quality >= 90)
- [x] Quality score >= 90 for all templates
- [x] Documentation complete (README, SKILL-AUTHORING.md, TEMPLATE-CREATION.md)
- [x] Build succeeds
- [x] Pre-existing test/lint failures in other packages (not related to this feature)
- [x] No breaking changes to existing workflow
- [x] Backward compatible with manual skill creation

### Manual Testing
- [x] Test `go-ent skill new my-go-skill` with go-basic template
- [x] Test `go-ent skill new my-ts-skill` with typescript-basic template
- [x] Test `go-ent skill list-templates` shows all 11 templates
- [x] Test `go-ent skill show-template go-complete` displays details
- [x] Test `go-ent skill add-template` with custom template
- [x] Verify generated skill validates successfully
- [x] Verify generated skill has correct structure
- [x] Test auto-detection of category for various names
- [x] Test error handling with missing template
- [x] Test with skill name that already exists

## Verification Summary

All phases completed successfully:

**Phase 1: CLI Infrastructure** ✅
- Implemented template loader, parser, generator
- Added placeholder replacement engine
- Integrated validation with skill validator
- Non-interactive mode with flags implemented

**Phase 2: Built-in Templates** ✅
- All 11 templates created and validated
- Quality scores: 90-97 across all templates
- Templates cover go, typescript, database, testing, api-design, core, debugging, security, review, arch
- Each template includes 2-3 examples and handles edge cases

**Phase 3: Interactive Wizard** ✅
- Template selection UI implemented
- Auto-detection of category and output path
- Progress indicators and success messages
- Non-interactive mode with `--template` flag

**Phase 4: Custom Template Support** ✅
- `list-templates` command with filtering
- `add-template` command for custom templates
- `show-template` for template details
- Support for both built-in and custom templates

**Phase 5: Testing** ✅
- Unit tests: 86.5% coverage in template package
- Integration tests: All commands tested
- Template validation: All 11 templates generate valid skills (quality >= 90)
- End-to-end tests: Complete workflow validated

**Phase 6: Documentation** ✅
- SKILL-AUTHORING.md updated with template guide
- TEMPLATE-CREATION.md created with template authoring guide
- AGENTS.md updated with new CLI commands
- Inline documentation added to all code

**Notes:**
- Pre-existing test/lint failures in other packages (domain, spec) are unrelated to this feature
- Build succeeds: `make build` completes without errors
- All skill generator commands work as specified
- No breaking changes to existing workflow