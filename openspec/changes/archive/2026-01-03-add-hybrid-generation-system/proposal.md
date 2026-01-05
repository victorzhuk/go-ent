# Change: Add Hybrid Generation System

## Why

The current go-ent architecture has a critical gap: templates exist but aren't embedded or processed, and there's no bridge between specs and code generation. Industry research shows:

1. **Template-only generation is rigid**: Traditional scaffolding tools (Cookiecutter, Yeoman) create identical projects but can't adapt to context or generate business logic.

2. **AI-only generation is inconsistent**: Pure LLM generation produces variable outputs, may hallucinate APIs, and lacks structural guarantees.

3. **Hybrid is the emerging best practice**: "Hybrid systems deliver the best of both worlds - use Prompt Templates for structure and consistency, while AI Agents handle learning, analysis, and adaptation."

Currently, go-ent cannot:
- Generate projects from specs without manual template selection
- Bridge the gap between requirements and implementation
- Leverage AI for context-aware code generation while maintaining consistency

## What Changes

### 1. Generation Configuration

Add `generation.yaml` to OpenSpec structure:
```yaml
openspec/
├── generation.yaml     # NEW: Project generation settings
├── project.md          # Existing
└── specs/              # Existing
```

### 2. New MCP Tools

| Tool | Purpose |
|------|---------|
| `go_ent_generate_component` | Generate a component from spec + templates |
| `go_ent_generate_from_spec` | Analyze spec and generate matching code |
| `go_ent_list_archetypes` | List available project archetypes |

### 3. Spec-to-Template Mapping

Create mapping system that:
- Analyzes spec requirements
- Selects appropriate templates
- Identifies extension points for AI generation
- Generates component scaffold

### 4. AI Prompt Templates

Create prompt templates in `prompts/` directory:
- `prompts/usecase.md` - Generate use case implementation
- `prompts/handler.md` - Generate API handler
- `prompts/repository.md` - Generate repository implementation

## Impact

- **Affected specs**: `mcp-tools` (new tools)
- **New OpenSpec file**: `generation.yaml` (optional)
- **New directory**: `prompts/` for AI prompt templates
- **Affected code**:
  - `cmd/go-ent/internal/tools/` - new tool handlers
  - `cmd/go-ent/internal/generation/` - new package
- **Breaking changes**: None (additive only)
- **Dependencies**: Requires `add-mcp-generation-tools` to be completed first

## Success Criteria

1. `go_ent_generate_component` creates component scaffold from spec
2. `generation.yaml` allows project-specific generation config
3. AI prompts fill in business logic after template generation
4. Generated code follows enterprise patterns from skills

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| AI hallucination in generated code | Medium | Templates provide structure, AI fills details only |
| Generation config complexity | Low | Keep minimal, sensible defaults |
| Spec analysis inaccuracy | Medium | Start with explicit mapping, add heuristics later |
