# Go-Ent Architecture Review & Strategic Recommendations

**Date:** 2026-01-03
**Reviewer:** Claude (AI Architecture Analysis)
**Project:** Go-Ent v2.0 - Enterprise Go Development Toolkit

---

## Executive Summary

Go-Ent is a **spec-driven development toolkit** that provides an MCP server for Claude Code integration, offering enterprise Go project scaffolding, spec management, and workflow automation. After thorough analysis and industry research, this document provides:

1. Architecture analysis (pros/cons)
2. Comparison with similar tools in the ecosystem
3. Template generator vs. prompt-based generation evaluation
4. OpenSpec extension recommendations
5. Strategic implementation plan

**Key Recommendation:** Adopt a **hybrid approach** combining embedded templates for structural consistency with prompt-based generation for business logic. OpenSpec should be extended minimally to support template metadata, not replaced.

---

## Part 1: Current Architecture Analysis

### 1.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Claude Code Plugin                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │ 8 Agents    │ │ 16 Commands │ │ 9 Skills    │            │
│  └─────────────┘ └─────────────┘ └─────────────┘            │
└────────────────────────────┬────────────────────────────────┘
                             │ MCP Protocol (stdio)
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    MCP Server (goent)                        │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Tools: init, list, show, crud, registry, workflow, loop ││
│  └─────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Domain: spec/, store, workflow, registry, loop state    ││
│  └─────────────────────────────────────────────────────────┘│
└────────────────────────────┬────────────────────────────────┘
                             │ File I/O
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    OpenSpec Structure                        │
│  openspec/                                                   │
│  ├── project.md    (conventions)                             │
│  ├── specs/        (current truth - deployed capabilities)  │
│  ├── changes/      (proposals - pending changes)            │
│  └── archive/      (completed changes)                      │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Strengths (Pros)

#### ✅ **1. Brownfield-First Design**
OpenSpec's separation of `specs/` (truth) from `changes/` (proposals) is perfectly suited for evolving projects. This aligns with [OpenSpec's core philosophy](https://github.com/Fission-AI/OpenSpec) as "brownfield-first" - unlike GitHub Spec-Kit which targets greenfield projects.

#### ✅ **2. MCP Protocol Adoption**
Using MCP (Model Context Protocol) positions the project well for the future. MCP has become the "de-facto standard" for AI tool integration, with adoption by OpenAI, Google, VS Code, and AWS.

#### ✅ **3. Comprehensive Workflow System**
The three-stage workflow (Draft → Review/Approve → Implement → Archive) with explicit approval gates prevents premature implementation and maintains human oversight.

#### ✅ **4. Multi-Tool AI Compatibility**
OpenSpec's AGENTS.md pattern allows different team members to use Claude Code, Cursor, CodeBuddy, or any AGENTS.md-compatible tool while sharing the same specs.

#### ✅ **5. Registry-Based Task Coordination**
Cross-change dependency tracking and priority-based task recommendation (`/gt:registry next`) enables sophisticated project management.

#### ✅ **6. Self-Correcting Autonomous Loop**
The `/go-ent:loop` command with retry logic and error categorization represents cutting-edge agentic patterns.

#### ✅ **7. Clean Separation of Concerns**
- Domain logic in `internal/spec/`, `internal/template/`, `internal/generation/`
- Tool handlers in `cmd/go-ent/internal/tools/`
- Templates as reference patterns (not runtime dependencies)

### 1.3 Weaknesses (Cons)

#### ❌ **1. Templates Not Embedded - Critical Gap**
15 template files exist in `/templates/` but are never embedded in the binary. The pending `add-mcp-generation-tools` change addresses this but is unimplemented.

```go
// MISSING: This file doesn't exist yet
//go:embed **/*.tmpl
var TemplateFS embed.FS
```

#### ❌ **2. No Input Schemas for MCP Tools**
Current tools lack JSON Schema definitions, which MCP 2025 spec recommends for better AI integration:

```go
// Current (tools/init.go) - no inputSchema
s.HandleTool("go_ent_spec_init", initHandler)

// Should be:
s.HandleTool("go_ent_spec_init", initHandler, mcp.WithInputSchema(initSchema))
```

#### ❌ **3. Hardcoded Plugin Path**
`plugin.json` contains absolute path `/home/zhuk/Projects/own/go-ent/dist/go-ent` making marketplace distribution impossible.

#### ❌ **4. Missing Validation Tool**
AGENTS.md extensively documents `openspec validate [change-id] --strict` but no MCP tool implements this. Validation is critical for the workflow.

#### ❌ **5. No Archive Automation**
Stage 3 (Archive) of the OpenSpec workflow requires manual file operations. No tool automates delta merging.

#### ❌ **6. Template Syntax Mismatch**
Templates use custom `{{MODULE_PATH}}` syntax instead of Go's `text/template` `{{.ModulePath}}`, requiring conversion.

#### ❌ **7. Limited Project Archetypes**
Only two project types (standard, mcp). No support for:
- API-only services
- gRPC services
- CLI tools
- Worker/queue processors

#### ❌ **8. No Live Spec Exploration**
Unlike Kiro which provides IDE-integrated spec visualization, go-ent requires file reading to explore specs.

### 1.4 Architecture Debt Assessment

| Category | Severity | Description | Pending Fix? |
|----------|----------|-------------|--------------|
| Template embedding | 🔴 Critical | Core feature unusable | Yes (T1.1) |
| Plugin path | 🔴 Critical | Distribution blocked | Yes (T4.1) |
| Validation tool | 🟠 High | Workflow incomplete | Yes (T2.2) |
| Archive tool | 🟠 High | Manual process required | Yes (T3.3) |
| Input schemas | 🟡 Medium | AI integration degraded | Yes (T5.2) |
| Template syntax | 🟡 Medium | Conversion needed | Yes (T1.3) |
| Project types | 🟢 Low | Limited but functional | No |

---

## Part 2: Industry Comparison

### 2.1 Similar Tools Landscape (2025)

| Tool | Category | Strengths | Weaknesses |
|------|----------|-----------|------------|
| **GitHub Spec-Kit** | SDD | Battle-tested, constitution-driven | Greenfield only |
| **OpenSpec** | SDD | Brownfield-first, lightweight | Less ecosystem |
| **Kiro (AWS)** | Agentic IDE | Full IDE, multimodal | Proprietary |
| **BMAD Method** | SDD | 19 specialized agents | Heavy overhead |
| **Tessl** | SDD | Cutting-edge | Beta only |
| **Ent Framework** | Code Gen | Type-safe, schema-first | DB-only |
| **Cookiecutter** | Scaffolding | 6000+ templates | No AI integration |
| **Yeoman** | Scaffolding | Flexible automation | Dated, complex |
| **Copier** | Scaffolding | Template updates | No AI integration |

### 2.2 Positioning Analysis

Go-Ent occupies a unique niche:

```
                    Greenfield ◄──────────────────► Brownfield
                         │                              │
                    Spec-Kit                       OpenSpec
                    Kiro                           go-ent
                         │                              │
            ─────────────┼──────────────────────────────┼─────────────
                         │                              │
     Template-Based ◄────┼────────────────────────────►│ AI-Prompt-Based
                         │                              │
                    Cookiecutter                   Pure LLM
                    Yeoman                         Copilot
                         │                              │
                         ▼                              ▼
              go-ent targets HERE ──────────────────►
              (Hybrid: Templates + AI for brownfield)
```

### 2.3 Competitive Advantages

1. **Go-Specific Expertise**: 9 specialized skills for Go development (go-arch, go-api, go-test, etc.)
2. **MCP-First Design**: Native MCP server vs. CLI wrappers
3. **OpenSpec Integration**: Built-in brownfield change management
4. **Enterprise Patterns**: Clean Architecture, SOLID enforcement
5. **Self-Correction**: Autonomous loop with error recovery

### 2.4 Competitive Gaps

1. **No IDE Integration**: Unlike Kiro, no visual spec exploration
2. **Fewer Project Types**: Cookiecutter has 6000+ templates
3. **Single Language**: Go-only vs. polyglot tools
4. **No Cloud Backend**: Local-only vs. cloud-powered validation

---

## Part 3: Template Generator vs. Prompts Evaluation

### 3.1 The Core Question

> Should go-ent use embedded templates (like Cookiecutter) or rely on AI prompts (like pure Copilot) for code generation?

### 3.2 Template-Based Generation

**How it works:**
```
Templates (embedded) + Variables (user input) = Generated Code
```

**Advantages:**
- ✅ **Consistency**: Every project follows identical structure
- ✅ **Speed**: Instant generation, no AI latency
- ✅ **Determinism**: Same inputs → same outputs
- ✅ **Offline**: Works without internet/API
- ✅ **Reviewability**: Templates are auditable code

**Disadvantages:**
- ❌ **Rigidity**: Limited to predefined patterns
- ❌ **Maintenance**: Templates need updates
- ❌ **No Context**: Can't adapt to existing code
- ❌ **Boilerplate**: Generates everything, even unused

### 3.3 Prompt-Based Generation

**How it works:**
```
Natural Language Description + AI Model = Generated Code
```

**Advantages:**
- ✅ **Flexibility**: Any pattern, any language
- ✅ **Context-Aware**: Can read existing code
- ✅ **Adaptive**: Matches project conventions
- ✅ **Interactive**: Iterative refinement

**Disadvantages:**
- ❌ **Inconsistency**: Different outputs each time
- ❌ **Hallucination**: May generate non-existent APIs
- ❌ **Security**: May introduce vulnerabilities
- ❌ **Latency**: Requires API calls
- ❌ **Cost**: Token consumption

### 3.4 Hybrid Approach (Recommended)

**Research shows:** "Hybrid systems deliver the best of both worlds - use Prompt Templates for structure and consistency, while AI Agents handle learning, analysis, and adaptation."

**Recommended Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                    go-ent Generation Pipeline                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐    ┌─────────────────┐    ┌─────────────┐ │
│  │ Templates   │ ─► │ Structure       │ ─► │ Skeleton    │ │
│  │ (Embedded)  │    │ Generation      │    │ Project     │ │
│  └─────────────┘    └─────────────────┘    └──────┬──────┘ │
│                                                    │        │
│  ┌─────────────┐    ┌─────────────────┐    ┌──────▼──────┐ │
│  │ Specs       │ ─► │ AI Prompt       │ ─► │ Business    │ │
│  │ (OpenSpec)  │    │ Generation      │    │ Logic       │ │
│  └─────────────┘    └─────────────────┘    └──────┬──────┘ │
│                                                    │        │
│                                            ┌──────▼──────┐ │
│                                            │ Complete    │ │
│                                            │ Project     │ │
│                                            └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**Layer Responsibilities:**

| Layer | Source | Responsibility | Example |
|-------|--------|----------------|---------|
| **Structure** | Templates | Directory layout, config files, boilerplate | `go.mod`, `Makefile`, `Dockerfile` |
| **Contracts** | OpenSpec | Interfaces, API schemas, types | `internal/domain/entity/*.go` |
| **Logic** | AI Prompts | Business rules, handlers, implementations | `internal/usecase/*.go` |

### 3.5 Decision: Templates ARE Needed

**Verdict:** Template generator is essential, but with enhanced role.

**Reasons:**
1. **Consistency Mandate**: Enterprise projects need identical structures
2. **Offline Support**: Can't depend on AI for basic scaffolding
3. **Speed**: Initial project setup should be instant
4. **Auditability**: Security teams can review templates

**But Enhanced with:**
1. **Spec-Driven Extension Points**: Templates include placeholders for spec-generated code
2. **AI-Powered Customization**: Post-generation refinement via prompts
3. **Dynamic Templates**: Template selection based on spec analysis

---

## Part 4: OpenSpec Extension Analysis

### 4.1 Current OpenSpec Capabilities

```yaml
openspec/
├── project.md              # Project conventions
├── specs/                  # Deployed capabilities (requirements + scenarios)
│   └── {capability}/
│       ├── spec.md         # What the system does
│       └── design.md       # How it's built
├── changes/                # Proposed changes
│   └── {change-id}/
│       ├── proposal.md     # Why + what
│       ├── tasks.md        # Implementation steps
│       ├── design.md       # Technical decisions
│       └── specs/          # Delta changes
└── archive/                # Completed changes
```

### 4.2 What's Missing for Generation?

| Gap | Description | Current Workaround |
|-----|-------------|-------------------|
| **Template Metadata** | No spec-to-template mapping | Manual selection |
| **Generation Config** | No per-project generation settings | Hardcoded defaults |
| **Component Registry** | No list of generatable components | Ad-hoc discovery |
| **Generation History** | No record of what was generated | Git commits |
| **Customization Points** | No spec-defined extension hooks | Edit after generate |

### 4.3 Extension Options

#### Option A: Extend OpenSpec (Recommended)

Add new file types to OpenSpec structure:

```yaml
openspec/
├── project.md              # Existing
├── generation.yaml         # NEW: Generation configuration
├── components/             # NEW: Component definitions
│   ├── user-service.yaml
│   └── order-api.yaml
├── specs/                  # Existing
└── changes/                # Existing
```

**generation.yaml:**
```yaml
defaults:
  go_version: "1.25"
  project_type: standard

archetypes:
  api-service:
    templates: [go.mod, main, config, api-handler]
    skip: [grpc, worker]

  mcp-server:
    templates: [go.mod, main, mcp-server]

components:
  - name: user-service
    archetype: api-service
    spec: specs/user-management/spec.md
```

**Pros:**
- ✅ Keeps everything in one place
- ✅ Leverages existing OpenSpec tooling
- ✅ Natural extension of spec-driven approach

**Cons:**
- ❌ Adds complexity to OpenSpec structure
- ❌ May conflict with upstream OpenSpec updates

#### Option B: Separate Generation Config

Keep generation config outside OpenSpec:

```yaml
go-ent.yaml                 # Generation config
openspec/                   # Pure specs
templates/                  # Pure templates
```

**Pros:**
- ✅ Clear separation of concerns
- ✅ No OpenSpec modifications needed

**Cons:**
- ❌ Two config systems to maintain
- ❌ Less cohesive developer experience

#### Option C: Inline in project.md

Extend `project.md` with generation sections:

```markdown
# Project: MyService

## Conventions
...existing...

## Generation
- Archetype: api-service
- Go Version: 1.25
- Components:
  - user-service (specs/user-management)
  - order-api (specs/orders)
```

**Pros:**
- ✅ Minimal file changes
- ✅ Single source of truth

**Cons:**
- ❌ Mixes concerns in one file
- ❌ Harder to parse programmatically

### 4.4 Recommendation: Option A with Constraints

**Decision:** Extend OpenSpec minimally with `generation.yaml`

**Constraints:**
1. **Single File**: Only add `generation.yaml`, no new directories
2. **Optional**: Generation config is optional, defaults work without it
3. **Backwards Compatible**: Existing OpenSpec projects work unchanged
4. **Upstream Safe**: Extensions in `.goent/` namespace if needed

---

## Part 5: Strategic Implementation Plan

### 5.1 Phased Roadmap

```
Phase 0: Foundation (Current)         Phase 1: Generation
─────────────────────────────         ──────────────────
✅ MCP server running                 ⬜ Embed templates
✅ Spec/task management               ⬜ Template engine
✅ Workflow/registry tools            ⬜ go_ent_generate tool
❌ Templates not embedded             ⬜ Project archetypes
❌ No validation tool
❌ No archive tool

Phase 2: Validation                   Phase 3: Hybrid Generation
───────────────────                   ────────────────────────
⬜ Validation rules                   ⬜ generation.yaml support
⬜ go_ent_spec_validate                ⬜ Spec-to-template mapping
⬜ Strict mode                        ⬜ AI prompt integration
                                      ⬜ Component generation

Phase 4: Archive                      Phase 5: Advanced
────────────────                      ──────────────────
⬜ Spec merger                        ⬜ Template versioning
⬜ go_ent_spec_archive                 ⬜ Migration generation
⬜ Delta application                  ⬜ IDE integration
```

### 5.2 Immediate Priorities (Next 2 Weeks)

**Priority 1: Complete `add-mcp-generation-tools`**

This existing proposal covers critical gaps:
- T1.x: Template embedding system
- T2.x: Validation tool
- T3.x: Archive tool
- T4.x: Plugin path fix

**Priority 2: Fix Plugin Distribution**

Without this, the tool can't be installed:
```json
// Change from:
"command": "/home/zhuk/Projects/own/go-ent/dist/go-ent"
// To:
"command": "${pluginDir}/dist/go-ent"
```

**Priority 3: Add Input Schemas**

For MCP 2025 compliance and better AI integration.

### 5.3 New Proposal: Hybrid Generation System

A new OpenSpec change should be created:

```markdown
# Change: Add Hybrid Generation System

## Why
Template-only generation creates rigid structures. AI-only generation
is inconsistent. A hybrid approach provides:
- Templates for structure (consistent, fast, auditable)
- AI prompts for logic (flexible, context-aware)
- Specs as the bridge between them

## What Changes
1. Add `generation.yaml` to OpenSpec structure
2. Create component registry in specs
3. Implement `go_ent_generate_component` tool
4. Add spec-to-code prompt templates

## Impact
- Extends OpenSpec (additive)
- New MCP tool
- New generation workflow
```

### 5.4 Proposal: Template Versioning

For long-term maintainability:

```markdown
# Change: Add Template Versioning

## Why
As go-ent evolves, templates change. Projects generated with old
templates need upgrade paths.

## What Changes
1. Add version metadata to templates
2. Create template changelog
3. Implement `go_ent_upgrade` tool
4. Generate migration scripts

## Impact
- Template format change
- New MCP tool
- New upgrade workflow
```

### 5.5 Do NOT Extend: What to Avoid

❌ **Don't create a new spec format** - OpenSpec works well
❌ **Don't add visual IDE** - Out of scope for MCP server
❌ **Don't support other languages** - Go focus is strength
❌ **Don't replace templates with pure AI** - Consistency matters
❌ **Don't add cloud backend** - Keep it simple and local

---

## Part 6: Final Recommendations

### 6.1 Summary of Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Template generator needed?** | ✅ Yes | Structure consistency, speed, auditability |
| **Prompts needed?** | ✅ Yes | Business logic, context-awareness |
| **Extend OpenSpec?** | ✅ Minimally | Add `generation.yaml` only |
| **New spec format?** | ❌ No | OpenSpec works well |
| **Template versioning?** | ⬜ Future | After core generation works |

### 6.2 Action Items

1. **Immediate**: Complete `add-mcp-generation-tools` change
2. **Short-term**: Create `add-hybrid-generation` proposal
3. **Medium-term**: Add `generation.yaml` support
4. **Long-term**: Template versioning and migrations

### 6.3 Success Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Time to scaffold project | Manual | < 5 seconds |
| Template coverage | 0% embedded | 100% embedded |
| Spec validation | Manual | Automated |
| Archive process | Manual | One command |
| Plugin installable | ❌ No | ✅ Yes |

---

## Appendix: Research Sources

### MCP & AI Tools
- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)
- [Thoughtworks: MCP Impact 2025](https://www.thoughtworks.com/en-us/insights/blog/generative-ai/model-context-protocol-mcp-impact-2025)
- [VS Code MCP Support GA](https://github.blog/changelog/2025-07-14-model-context-protocol-mcp-support-in-vs-code-is-generally-available/)

### Spec-Driven Development
- [GitHub: Spec-Driven Development Toolkit](https://github.blog/ai-and-ml/generative-ai/spec-driven-development-with-ai-get-started-with-a-new-open-source-toolkit/)
- [Red Hat: SDD Improves AI Coding Quality](https://developers.redhat.com/articles/2025/10/22/how-spec-driven-development-improves-ai-coding-quality)
- [Martin Fowler: SDD Tools Comparison](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)
- [OpenSpec GitHub](https://github.com/Fission-AI/OpenSpec)

### Code Generation
- [Ent Framework](https://entgo.io/)
- [JetBrains: Go Ecosystem 2025](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/)
- [Medium: AI Code Generation with Templates](https://medium.com/@Neopric/using-generative-ai-as-a-code-generator-with-user-defined-templates-in-software-development-d3b3db0d4f0f)

### Scaffolding Tools
- [Cookiecutter vs Yeoman](https://www.cookiecutter.io/article-post/compare-cookiecutter-to-yeoman)
- [OpsLevel: Choosing Scaffolders](https://www.opslevel.com/resources/cookiecutter-vs-yeoman-choosing-the-right-scaffolder-for-your-service)

---

*This document should be reviewed and updated as the project evolves.*
