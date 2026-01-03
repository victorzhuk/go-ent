# Tasks: Add Hybrid Generation System

## Dependencies
- **BLOCKED BY**: `add-mcp-generation-tools` must be completed first
- T1.1 → T1.2, T1.3
- T2.1 → T2.2
- T1.3, T2.2 → T3.1

## Phase 1: Generation Configuration

### T1.1: Define generation.yaml schema
- **Story**: specs/mcp-tools/spec.md#Generation Configuration
- **Files**: cmd/goent/internal/generation/config.go
- **Depends**: None
- **Parallel**: No
- [ ] 1.1.1 Define `GenerationConfig` struct
- [ ] 1.1.2 Define `Archetype` struct with template lists
- [ ] 1.1.3 Define `ComponentConfig` struct with spec reference
- [ ] 1.1.4 Implement YAML parsing for generation.yaml
- [ ] 1.1.5 Implement defaults when file doesn't exist

### T1.2: Create archetype registry
- **Story**: specs/mcp-tools/spec.md#Project Archetypes
- **Files**: cmd/goent/internal/generation/archetypes.go
- **Depends**: T1.1
- **Parallel**: No
- [ ] 1.2.1 Define built-in archetypes (standard, mcp, api, grpc, worker)
- [ ] 1.2.2 Implement archetype resolution from config
- [ ] 1.2.3 Implement template list generation per archetype
- [ ] 1.2.4 Add archetype validation

### T1.3: Implement goent_list_archetypes tool
- **Story**: specs/mcp-tools/spec.md#Archetype Discovery
- **Files**: cmd/goent/internal/tools/archetypes.go
- **Depends**: T1.2
- **Parallel**: Yes (with T2.x)
- [ ] 1.3.1 Define `ListArchetypesInput` (optional filter)
- [ ] 1.3.2 Define inputSchema for MCP registration
- [ ] 1.3.3 Implement handler returning archetype metadata
- [ ] 1.3.4 Register tool in register.go

## Phase 2: Spec Analysis

### T2.1: Create spec analyzer
- **Story**: specs/mcp-tools/spec.md#Spec Analysis
- **Files**: cmd/goent/internal/generation/analyzer.go
- **Depends**: None
- **Parallel**: Yes (with T1.x)
- [ ] 2.1.1 Define `SpecAnalysis` struct (capabilities, patterns, components)
- [ ] 2.1.2 Implement requirement parsing from spec.md
- [ ] 2.1.3 Implement pattern detection (CRUD, API, async, etc.)
- [ ] 2.1.4 Implement component identification from requirements
- [ ] 2.1.5 Generate template recommendations

### T2.2: Implement spec-to-archetype mapping
- **Story**: specs/mcp-tools/spec.md#Archetype Selection
- **Files**: cmd/goent/internal/generation/mapper.go
- **Depends**: T2.1
- **Parallel**: No
- [ ] 2.2.1 Define mapping rules (patterns → archetypes)
- [ ] 2.2.2 Implement scoring algorithm for archetype selection
- [ ] 2.2.3 Support explicit override via generation.yaml
- [ ] 2.2.4 Generate component list from spec analysis

## Phase 3: Component Generation

### T3.1: Implement goent_generate_component tool
- **Story**: specs/mcp-tools/spec.md#Component Generation
- **Files**: cmd/goent/internal/tools/generate_component.go
- **Depends**: T1.3, T2.2
- **Parallel**: No
- [ ] 3.1.1 Define `GenerateComponentInput` (spec_path, component_name, output_dir)
- [ ] 3.1.2 Define inputSchema for MCP registration
- [ ] 3.1.3 Implement spec analysis integration
- [ ] 3.1.4 Implement template selection based on analysis
- [ ] 3.1.5 Generate component scaffold from templates
- [ ] 3.1.6 Mark extension points for AI generation
- [ ] 3.1.7 Register tool in register.go

### T3.2: Implement goent_generate_from_spec tool
- **Story**: specs/mcp-tools/spec.md#Spec-Driven Generation
- **Files**: cmd/goent/internal/tools/generate_from_spec.go
- **Depends**: T3.1
- **Parallel**: No
- [ ] 3.2.1 Define `GenerateFromSpecInput` (spec_path, output_dir, options)
- [ ] 3.2.2 Implement full project generation from spec
- [ ] 3.2.3 Iterate through identified components
- [ ] 3.2.4 Generate each component using T3.1
- [ ] 3.2.5 Create integration points between components
- [ ] 3.2.6 Register tool in register.go

## Phase 4: AI Prompt Integration

### T4.1: Create prompt template system
- **Story**: specs/mcp-tools/spec.md#AI Prompt Templates
- **Files**: prompts/*.md, cmd/goent/internal/generation/prompts.go
- **Depends**: None
- **Parallel**: Yes (independent)
- [ ] 4.1.1 Create prompts/ directory structure
- [ ] 4.1.2 Write usecase.md prompt template
- [ ] 4.1.3 Write handler.md prompt template
- [ ] 4.1.4 Write repository.md prompt template
- [ ] 4.1.5 Implement prompt template loading
- [ ] 4.1.6 Implement variable substitution in prompts

### T4.2: Create extension point markers
- **Story**: specs/mcp-tools/spec.md#Extension Points
- **Files**: templates/**/*.tmpl
- **Depends**: T4.1
- **Parallel**: No
- [ ] 4.2.1 Define extension point syntax (e.g., `// @generate:usecase`)
- [ ] 4.2.2 Add markers to relevant templates
- [ ] 4.2.3 Implement extension point detection
- [ ] 4.2.4 Generate AI prompts at extension points

## Phase 5: Testing and Documentation

### T5.1: Add tests for generation system
- **Files**: cmd/goent/internal/generation/*_test.go
- **Depends**: All implementation tasks
- [ ] 5.1.1 Add tests for config parsing
- [ ] 5.1.2 Add tests for spec analyzer
- [ ] 5.1.3 Add tests for archetype selection
- [ ] 5.1.4 Add tests for component generation
- [ ] 5.1.5 Add integration tests for full workflow

### T5.2: Update documentation
- **Files**: README.md, openspec/AGENTS.md
- **Depends**: All implementation tasks
- [ ] 5.2.1 Document generation.yaml format
- [ ] 5.2.2 Document new MCP tools
- [ ] 5.2.3 Add examples for hybrid generation workflow
- [ ] 5.2.4 Update AGENTS.md with generation commands
