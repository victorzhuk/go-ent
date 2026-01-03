# Go-Ent Architecture Review & Implementation Plan

## Executive Summary

Go-ent is an enterprise Go development toolkit combining spec-driven development (OpenSpec) with MCP-based AI tooling and project scaffolding. This document provides a comprehensive review of current proposals, architecture decisions, and a detailed implementation roadmap.

---

## Part 1: Current State Analysis

### 1.1 Active Proposals Overview

| Proposal | Status | Completion | Blocker |
|----------|--------|------------|---------|
| `add-mcp-generation-tools` | In Progress | ~85% | Testing & polish |
| `add-hybrid-generation-system` | Blocked | 0% | Requires MCP tools |

### 1.2 Proposal: add-mcp-generation-tools

**Purpose**: Add core MCP tools for project generation, spec validation, and change archival.

**Implementation Status**:

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Template Embedding System | ✅ Complete |
| Phase 2 | Validation Tool | ✅ Complete |
| Phase 3 | Archive Tool | ✅ Complete |
| Phase 4 | Plugin Configuration | ⚠️ Partial (needs testing) |
| Phase 5 | Testing & Documentation | ❌ Not Started |

**Completed Tasks**:
- `goent_generate` - Creates Go projects from embedded templates
- `goent_spec_validate` - Validates specs and change proposals
- `goent_spec_archive` - Archives completed changes with delta merging
- Template engine with variable substitution
- Validation rule framework

**Remaining Tasks** (15 items):
```
T4.1.2 - Test plugin installation in Claude Code
T4.1.3 - Update README if installation instructions change
T5.1.1 - Add tests for template engine
T5.1.2 - Add tests for validation rules
T5.1.3 - Add tests for spec merger
T5.1.4 - Add tests for archiver
T5.1.5 - Verify make test passes
T5.2.1 - Add inputSchema to goent_spec_init
T5.2.2 - Add inputSchema to goent_spec_create
T5.2.3 - Add inputSchema to goent_spec_update
T5.2.4 - Add inputSchema to goent_spec_delete
T5.2.5 - Add inputSchema to goent_spec_list
T5.2.6 - Add inputSchema to goent_spec_show
T5.2.7 - Add inputSchema to all registry tools
T5.2.8 - Add inputSchema to all workflow tools
T5.2.9 - Add inputSchema to all loop tools
```

### 1.3 Proposal: add-hybrid-generation-system

**Purpose**: Bridge the gap between rigid template-based generation and inconsistent AI-only generation.

**Key Innovation**: Two-stage generation model:
1. **Stage 1 (Deterministic)**: Templates + variables generate structure
2. **Stage 2 (AI-Driven)**: Client executes prompts to fill business logic

**Design Decisions**:

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Config Location | `openspec/generation.yaml` | Keeps project config together |
| Schema | Minimal with sensible defaults | Reduces complexity |
| Spec Analysis | Pattern-based heuristics + override | Fast, deterministic, flexible |
| Extension Points | Comment-based `@generate:` markers | Non-breaking, contextual |
| AI Prompts | Markdown templates with Go template vars | Easy customization |

**Built-in Archetypes**:
- `standard` - Web service with clean architecture
- `mcp` - MCP server plugin
- `api` - API-only service
- `grpc` - gRPC service
- `worker` - Background worker

**New Tools Planned**:
- `goent_generate_component` - Generate component from spec + templates
- `goent_generate_from_spec` - Analyze spec and generate matching code
- `goent_list_archetypes` - List available project archetypes

---

## Part 2: Architecture Analysis

### 2.1 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Server (stdio)                        │
│                    cmd/goent/main.go                         │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼─────────────────────────────────┐
│                   tools/register.go                            │
│                   (Tool Registration Hub)                      │
└─────────────────────────────────────────────────────────────────┘
        │           │           │           │           │
        ▼           ▼           ▼           ▼           ▼
   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
   │  CRUD   │ │Registry │ │Workflow │ │  Loop   │ │ Generate│
   │  Tools  │ │  Tools  │ │  Tools  │ │  Tools  │ │  Tools  │
   └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
        │           │           │           │           │
        └───────────┴───────────┼───────────┴───────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                 ▼
       ┌──────────┐      ┌──────────┐      ┌──────────┐
       │  spec/   │      │ template/│      │templates/│
       │  store   │      │  engine  │      │ embed.FS │
       │validator │      │          │      │          │
       │ archiver │      │          │      │          │
       └──────────┘      └──────────┘      └──────────┘
```

### 2.2 Current Tool Inventory

| Tool | Package | Purpose |
|------|---------|---------|
| `goent_spec_init` | tools/init.go | Initialize OpenSpec in project |
| `goent_spec_create` | tools/crud.go | Create spec/change |
| `goent_spec_update` | tools/crud.go | Update spec/change |
| `goent_spec_delete` | tools/crud.go | Delete spec/change |
| `goent_spec_list` | tools/list.go | List specs/changes |
| `goent_spec_show` | tools/show.go | Show spec/change details |
| `goent_generate` | tools/generate.go | Generate project from templates |
| `goent_spec_validate` | tools/validate.go | Validate specs/changes |
| `goent_spec_archive` | tools/archive.go | Archive completed changes |
| `goent_registry_*` | tools/registry.go | Task registry management |
| `goent_workflow_*` | tools/workflow.go | Workflow state management |
| `goent_loop_*` | tools/loop.go | Autonomous loop execution |

### 2.3 Data Flow: Generation Pipeline

```
User Request
    │
    ▼
┌──────────────────┐
│ Read generation  │  ← openspec/generation.yaml (optional)
│ config           │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Analyze spec     │  ← openspec/specs/*/spec.md
│ Identify patterns│
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Select archetype │  ← Pattern matching + explicit override
│ & templates      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Generate scaffold│  ← templates/*.tmpl
│ with extension   │
│ points           │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Return files +   │  → Generated code + AI prompts
│ prompt templates │
└──────────────────┘
```

---

## Part 3: Implementation Roadmap

### Phase A: Complete add-mcp-generation-tools (Priority: HIGH)

**Goal**: Finish remaining tasks and archive the change.

#### A.1: Plugin Testing (2 tasks)
```
[ ] T4.1.2 - Test plugin installation in Claude Code
    - Install plugin via Claude Code marketplace
    - Verify goent binary resolves correctly
    - Test all MCP tools work via Claude Code

[ ] T4.1.3 - Update README if needed
    - Document installation steps
    - Add troubleshooting section
```

#### A.2: Unit Tests (5 tasks)
```
[ ] T5.1.1 - Template engine tests
    - Test variable substitution
    - Test ProcessAll with different project types
    - Test error handling for missing templates

[ ] T5.1.2 - Validation rules tests
    - Test each validation rule in isolation
    - Test strict vs normal mode
    - Test edge cases (empty files, malformed headers)

[ ] T5.1.3 - Spec merger tests
    - Test ADDED operation
    - Test MODIFIED operation (full requirement replacement)
    - Test REMOVED operation
    - Test RENAMED operation
    - Test multi-requirement deltas

[ ] T5.1.4 - Archiver tests
    - Test successful archive
    - Test dry-run mode
    - Test skip-specs mode
    - Test validation failure blocking

[ ] T5.1.5 - Run make test
    - Ensure all tests pass
    - Check for race conditions
    - Verify coverage meets standards
```

#### A.3: Input Schemas (9 tasks)

Add JSON input schemas to all existing tools for better MCP client compatibility:

```
[ ] T5.2.1 - goent_spec_init
    Schema: { path: string (required) }

[ ] T5.2.2 - goent_spec_create
    Schema: { path: string, type: "spec"|"change", id: string, content?: string }

[ ] T5.2.3 - goent_spec_update
    Schema: { path: string, type: "spec"|"change", id: string, content: string }

[ ] T5.2.4 - goent_spec_delete
    Schema: { path: string, type: "spec"|"change", id: string }

[ ] T5.2.5 - goent_spec_list
    Schema: { path: string, type?: "spec"|"change"|"all", status?: string }

[ ] T5.2.6 - goent_spec_show
    Schema: { path: string, type: "spec"|"change", id: string, format?: "md"|"json" }

[ ] T5.2.7 - Registry tools
    - goent_registry_list: { path: string, change?: string, status?: string }
    - goent_registry_next: { path: string, count?: number }
    - goent_registry_update: { path: string, task_id: string, status?: string, priority?: string }
    - goent_registry_deps: { path: string, task_id: string, operation: "show"|"add"|"remove", dep_id?: string }
    - goent_registry_sync: { path: string }
    - goent_registry_init: { path: string }

[ ] T5.2.8 - Workflow tools
    - goent_workflow_start: { path: string, change_id: string, phase?: string }
    - goent_workflow_approve: { path: string, wait_point: string }
    - goent_workflow_status: { path: string }

[ ] T5.2.9 - Loop tools
    - goent_loop_start: { path: string, task: string, max_iterations?: number }
    - goent_loop_cancel: { path: string }
    - goent_loop_status: { path: string }
```

#### A.4: Archive Change

After all tasks complete:
```bash
goent_spec_archive --id add-mcp-generation-tools
```

---

### Phase B: Implement add-hybrid-generation-system (Priority: MEDIUM)

**Dependency**: Phase A must be complete first.

#### B.1: Generation Configuration (Phase 1)

```
[ ] T1.1.1 - Define GenerationConfig struct
    - defaults: { go_version, archetype }
    - archetypes: map[string]ArchetypeConfig
    - components: []ComponentConfig

[ ] T1.1.2 - Define Archetype struct
    - name: string
    - description: string
    - templates: []string
    - skip: []string

[ ] T1.1.3 - Define ComponentConfig struct
    - name: string
    - spec: string (path to spec.md)
    - archetype: string (override)
    - output: string (output directory)

[ ] T1.1.4 - Implement YAML parsing
    - Parse openspec/generation.yaml
    - Handle missing file gracefully

[ ] T1.1.5 - Implement defaults
    - go_version: current runtime version
    - archetype: "standard"
```

#### B.2: Archetype Registry (Phase 1)

```
[ ] T1.2.1 - Define built-in archetypes
    - standard: web service with clean architecture
    - mcp: MCP server plugin
    - api: API-only service
    - grpc: gRPC service
    - worker: Background worker

[ ] T1.2.2 - Archetype resolution
    - Merge built-in with custom archetypes
    - Handle archetype inheritance (future)

[ ] T1.2.3 - Template list generation
    - Given archetype, return list of templates
    - Filter by skip list

[ ] T1.2.4 - Archetype validation
    - Ensure referenced templates exist
    - Warn on missing templates
```

#### B.3: goent_list_archetypes Tool (Phase 1)

```
[ ] T1.3.1 - Define ListArchetypesInput
    - filter?: string (type filter)

[ ] T1.3.2 - Define inputSchema

[ ] T1.3.3 - Implement handler
    - Return all archetypes with metadata
    - Apply filter if provided

[ ] T1.3.4 - Register tool
```

#### B.4: Spec Analyzer (Phase 2)

```
[ ] T2.1.1 - Define SpecAnalysis struct
    - capabilities: []string
    - patterns: []PatternMatch
    - components: []Component
    - confidence: float64

[ ] T2.1.2 - Implement requirement parsing
    - Parse ### Requirement: headers
    - Extract requirement text
    - Parse #### Scenario: sections

[ ] T2.1.3 - Implement pattern detection
    - CRUD: create, read, update, delete
    - API: endpoint, request, response
    - Async: queue, worker, background
    - Auth: authenticate, authorize, permission
    - Storage: repository, database, persist

[ ] T2.1.4 - Component identification
    - Group related requirements
    - Identify boundaries

[ ] T2.1.5 - Generate recommendations
    - Score archetypes by pattern match
    - Return top recommendation with confidence
```

#### B.5: Spec-to-Archetype Mapper (Phase 2)

```
[ ] T2.2.1 - Define mapping rules
    - Pattern → Archetype scoring matrix
    - Threshold for recommendation

[ ] T2.2.2 - Implement scoring algorithm
    - Weight patterns by relevance
    - Combine scores for final ranking

[ ] T2.2.3 - Support explicit override
    - generation.yaml takes precedence
    - Log when override differs from recommendation

[ ] T2.2.4 - Generate component list
    - For each identified component
    - Return archetype + templates
```

#### B.6: goent_generate_component Tool (Phase 3)

```
[ ] T3.1.1 - Define GenerateComponentInput
    - spec_path: string
    - component_name: string
    - output_dir: string
    - archetype?: string (override)

[ ] T3.1.2 - Define inputSchema

[ ] T3.1.3 - Spec analysis integration
    - Call analyzer
    - Get patterns and components

[ ] T3.1.4 - Template selection
    - Based on analysis or override
    - Get template list from archetype

[ ] T3.1.5 - Generate scaffold
    - Process templates
    - Output to target directory

[ ] T3.1.6 - Mark extension points
    - Add @generate: comments
    - Include spec context

[ ] T3.1.7 - Register tool
```

#### B.7: goent_generate_from_spec Tool (Phase 3)

```
[ ] T3.2.1 - Define GenerateFromSpecInput
    - spec_path: string
    - output_dir: string
    - options?: GenerationOptions

[ ] T3.2.2 - Implement full generation
    - Analyze spec
    - Identify all components

[ ] T3.2.3 - Iterate components
    - For each component
    - Call generate_component

[ ] T3.2.4 - Generate integration
    - Wire components together
    - Create main.go entry point

[ ] T3.2.5 - Create integration points
    - Dependency injection
    - Configuration setup

[ ] T3.2.6 - Register tool
```

#### B.8: AI Prompt Templates (Phase 4)

```
[ ] T4.1.1 - Create prompts/ directory

[ ] T4.1.2 - Write usecase.md
    - Context section
    - Spec content variable
    - Requirements list
    - Instructions for implementation

[ ] T4.1.3 - Write handler.md
    - HTTP handler generation
    - Request/response handling
    - Error handling patterns

[ ] T4.1.4 - Write repository.md
    - Data access patterns
    - CRUD operations
    - Transaction handling

[ ] T4.1.5 - Implement prompt loading
    - Read from prompts/
    - Support custom prompts

[ ] T4.1.6 - Variable substitution
    - {{.SpecContent}}
    - {{.Requirements}}
    - {{.ExistingCode}}
    - {{.ProjectName}}
    - {{.Conventions}}
```

#### B.9: Extension Point System (Phase 4)

```
[ ] T4.2.1 - Define extension point syntax
    - // @generate:<type>
    - Types: constructor, methods, validation, handlers, tests

[ ] T4.2.2 - Add markers to templates
    - Update relevant .tmpl files
    - Include context comments

[ ] T4.2.3 - Implement detection
    - Parse generated files
    - Find @generate markers

[ ] T4.2.4 - Generate AI prompts
    - At each extension point
    - Include relevant context
```

#### B.10: Testing (Phase 5)

```
[ ] T5.1.1 - Config parsing tests
[ ] T5.1.2 - Spec analyzer tests
[ ] T5.1.3 - Archetype selection tests
[ ] T5.1.4 - Component generation tests
[ ] T5.1.5 - Integration tests
```

---

## Part 4: Risk Assessment

### 4.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Input schema changes break existing clients | Low | Medium | Version schemas, deprecate gracefully |
| Spec analysis misidentifies patterns | Medium | Low | Allow explicit override, show confidence |
| Extension points clutter code | Low | Medium | Clear syntax, can be stripped |
| AI generates incorrect code | Medium | Medium | Templates provide structure, human review |
| Plugin path resolution fails | Medium | High | Test on multiple platforms |

### 4.2 Dependencies

```
add-mcp-generation-tools
    └── add-hybrid-generation-system (BLOCKED)
            └── Future: multi-language support
            └── Future: custom template repos
```

---

## Part 5: Quality Criteria

### 5.1 Acceptance Criteria for Phase A

1. All unit tests pass with `make test`
2. Plugin installs and works in Claude Code
3. All MCP tools have input schemas
4. Generated projects build successfully
5. Validation catches documented error cases
6. Archive properly merges deltas into specs

### 5.2 Acceptance Criteria for Phase B

1. `generation.yaml` is optional (defaults work)
2. Spec analysis identifies 80%+ of patterns correctly
3. All built-in archetypes generate valid projects
4. Extension points are clearly marked
5. AI prompts include sufficient context
6. Component generation creates compilable code

---

## Part 6: Recommended Next Actions

### Immediate (Today)

1. **Start Phase A.2**: Write unit tests for template engine
2. **Review**: Verify all Phase 1-3 tasks are actually complete
3. **Test**: Try plugin installation manually

### Short-term (This Week)

1. Complete all Phase A tasks
2. Archive `add-mcp-generation-tools`
3. Begin Phase B.1 (Generation Configuration)

### Medium-term (Next 2 Weeks)

1. Complete Phase B.1-B.4
2. Implement core generation tools
3. Create initial prompt templates

---

## Appendix: File Inventory

### A.1 Implementation Files

| File | Purpose | Status |
|------|---------|--------|
| `cmd/goent/main.go` | MCP server entry | ✅ |
| `cmd/goent/internal/tools/register.go` | Tool registration | ✅ |
| `cmd/goent/internal/tools/generate.go` | Project generation | ✅ |
| `cmd/goent/internal/tools/validate.go` | Spec validation | ✅ |
| `cmd/goent/internal/tools/archive.go` | Change archival | ✅ |
| `cmd/goent/internal/template/engine.go` | Template processing | ✅ |
| `cmd/goent/internal/spec/validator.go` | Validation logic | ✅ |
| `cmd/goent/internal/spec/rules.go` | Validation rules | ✅ |
| `cmd/goent/internal/spec/merger.go` | Delta merging | ✅ |
| `cmd/goent/internal/spec/archiver.go` | Archive logic | ✅ |
| `cmd/goent/templates/embed.go` | Template embedding | ✅ |
| `cmd/goent/internal/generation/*.go` | Hybrid gen (planned) | ❌ |

### A.2 Spec Files

| File | Purpose |
|------|---------|
| `openspec/specs/cli-build/spec.md` | CLI build requirements |
| `openspec/specs/ci-pipeline/spec.md` | CI/CD requirements |
| `openspec/changes/add-mcp-generation-tools/` | Active change |
| `openspec/changes/add-hybrid-generation-system/` | Blocked change |

---

*Generated: 2026-01-03*
*Next Review: After Phase A completion*
