# Design: Align Agent Tool Configurations

## Context

The go-ent agent system uses YAML-based metadata definitions in `pkg/agents/meta/*.yaml` to configure agents. These definitions support inheritance through the `extends` field, allowing agents to build upon base configurations. Tool presets in `pkg/agents/presets/tools.yaml` provide reusable tool groupings.

**Current State:**
- Agent metadata uses simple field replacement during inheritance (not additive)
- Tool presets are expanded via `expandToolPresets()` which merges preset tools into a set
- The `planning` preset exists in `tools.yaml` but isn't fully integrated
- 14 agent files need tool alignment and grep→rg pattern updates

**Constraints:**
- Must maintain backward compatibility with existing agent definitions
- Tool expansion happens at initialization time, not runtime
- Agent metadata is embedded via `pkg.FS` (embed.FS)

## Goals / Non-Goals

**Goals:**
- Implement additive inheritance for slice fields (skills, tools, toolPresets)
- Add `planning` preset to appropriate agents during initialization
- Support dynamic tool list generation from registry metadata
- Update all 14 agents with consistent tool configurations
- Replace grep patterns with rg equivalents in agent prompts

**Non-Goals:**
- Runtime tool list modification (tools are expanded at init time)
- Changing the agent file format or schema
- Removing existing tools or breaking existing workflows
- Adding new tool presets beyond `planning`

## Decisions

### 1. Additive Inheritance for Slice Fields

**Decision:** Modify `mergeAgents()` to merge slice fields additively instead of replacing them.

**Rationale:**
- Current behavior: `variant.Skills` completely replaces `base.Skills`
- Desired behavior: `variant.Skills` extends `base.Skills` (base + variant)
- This allows base agents to define common tools/skills while variants add specialized ones

**Implementation:**
```go
func mergeSlices(base, variant []string) []string {
    seen := make(map[string]bool)
    result := make([]string, 0, len(base)+len(variant))
    
    for _, item := range base {
        if !seen[item] {
            seen[item] = true
            result = append(result, item)
        }
    }
    
    for _, item := range variant {
        if !seen[item] {
            seen[item] = true
            result = append(result, item)
        }
    }
    
    return result
}
```

**Usage in mergeAgents:**
```go
merged.Skills = mergeSlices(base.Skills, variant.Skills)
merged.Tools = mergeSlices(base.Tools, variant.Tools)
merged.ToolPresets = mergeSlices(base.ToolPresets, variant.ToolPresets)
```

**Alternative Considered:**
- Keep replacement behavior and duplicate base definitions in each variant
- Rejected: Violates DRY principle, increases maintenance burden

### 2. Planning Preset Assignment

**Decision:** Add `planning` preset to agents with `role: planning` during `expandToolPresets()`.

**Rationale:**
- Planning agents need task management tools (TaskCreate, TaskUpdate, etc.)
- The `planning` preset already exists in `tools.yaml` with appropriate tools
- Automatic assignment ensures consistency and reduces manual configuration

**Implementation:**
```go
func expandToolPresets(meta *agentMeta, presets *toolPresets) {
    // Auto-assign planning preset to planning role agents
    if meta.Role == "planning" {
        hasPlanning := false
        for _, p := range meta.ToolPresets {
            if p == "planning" {
                hasPlanning = true
                break
            }
        }
        if !hasPlanning {
            meta.ToolPresets = append(meta.ToolPresets, "planning")
        }
    }
    
    // ... existing preset expansion logic
}
```

**Alternative Considered:**
- Manually add `planning` to each planner agent's YAML
- Rejected: Easy to forget, automatic assignment is more maintainable

### 3. Dynamic Tool Registry

**Decision:** Generate tool lists dynamically from agent metadata in `loadAgents()`.

**Rationale:**
- Current: Tools are defined statically in agent YAML or via presets
- Desired: Registry can introspect available tools and generate lists
- Enables "give me all analysis tools" or "give me all editing tools" queries

**Implementation:**
```go
// ToolRegistry provides dynamic tool list generation
type ToolRegistry struct {
    presets map[string][]string
    tools   map[string]ToolInfo
}

type ToolInfo struct {
    Name        string
    Category    string // "read", "write", "search", "analysis"
    Description string
}

func (r *ToolRegistry) GetToolsByCategory(category string) []string {
    var result []string
    for name, info := range r.tools {
        if info.Category == category {
            result = append(result, name)
        }
    }
    return result
}

func (r *ToolRegistry) GetToolsForRole(role string) []string {
    switch role {
    case "planning":
        return r.GetToolsByCategory("read")
    case "execution":
        return append(r.GetToolsByCategory("read"), r.GetToolsByCategory("write")...)
    case "research":
        return append(r.GetToolsByCategory("read"), r.GetToolsByCategory("search")...)
    default:
        return r.GetToolsByCategory("read")
    }
}
```

**Alternative Considered:**
- Static tool lists only
- Rejected: Dynamic generation enables more flexible agent configuration

### 4. Agent File Updates

**Decision:** Update all 14 agent files in `.claude/agents/*.md` with:
- Missing tools: `list`, `todoread`, `todowrite`, `skill`
- Replace grep with rg patterns
- Add "Optimal Tooling" section
- Add context gathering workflow phase

**Rationale:**
- Consistent tooling across all agents improves predictability
- rg/fd provide 10x performance improvement
- Standardized workflows reduce cognitive load

**Implementation Pattern:**
```yaml
# Before
tools:
  read: true
  write: true
  bash: true

# After
tools:
  read: true
  write: true
  bash: true
  list: true
  todoread: true
  todowrite: true
  skill: true
```

## Risks / Trade-offs

**[Risk]** Additive inheritance changes existing behavior for agents that relied on replacement
→ **Mitigation:** Audit all agents that use `extends` to ensure they don't depend on replacement semantics

**[Risk]** Planning preset auto-assignment might add unwanted tools to some agents
→ **Mitigation:** Agents can use `disallowedToolPresets` to exclude specific presets

**[Risk]** Dynamic tool registry adds complexity
→ **Mitigation:** Start with simple category-based filtering, expand only if needed

**[Risk]** grep→rg replacement might break on systems without rg installed
→ **Mitigation:** Document rg as required dependency; provide fallback guidance

**[Trade-off]** Additive inheritance vs. explicit configuration
- Additive: Less verbose, DRY principle
- Explicit: More predictable, easier to understand at a glance
- **Resolution:** Use additive with clear documentation

## Migration Plan

### Phase 1: Update Agent Files (14 files)
1. Update each `.claude/agents/*.md` with missing tools
2. Replace grep patterns with rg equivalents
3. Add "Optimal Tooling" section
4. Add context gathering phase

### Phase 2: Implement Additive Inheritance
1. Add `mergeSlices()` function to `internal/cli/init.go`
2. Update `mergeAgents()` to use additive merging for slice fields
3. Test with existing agent definitions

### Phase 3: Add Planning Preset
1. Update `expandToolPresets()` to auto-assign `planning` to planning role agents
2. Verify planner agents get task management tools

### Phase 4: Dynamic Tool Registry
1. Add `ToolRegistry` type with category-based lookups
2. Integrate with `loadAgents()` for dynamic tool list generation
3. Update agent metadata to support tool categories

### Rollback Strategy
- All changes are additive (no tool removals)
- Can revert individual phases independently
- Agent files can be regenerated with `--force` flag

## Open Questions

1. **Should we validate that all agents have required tools after expansion?**
   - Option A: Add validation in `validateAgent()`
   - Option B: Trust the preset system

2. **How should we handle tool ordering?**
   - Option A: Preserve base order, append variant order
   - Option B: Alphabetical sort after merge
   - Option C: Don't guarantee order (use sets)

3. **Should we support tool categories in YAML?**
   - Option A: Add `toolCategories` field to agent metadata
   - Option B: Keep categories internal to registry only

4. **Do we need a migration command for existing projects?**
   - Option A: `ent migrate agents` to update existing agent files
   - Option B: Document manual update process
