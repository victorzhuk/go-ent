# Agent Inheritance

## Overview

Agent inheritance allows you to share common configuration between related agents by defining a base agent and having other agents extend it. This reduces duplication and ensures consistent behavior across agent families.

The go-ent system uses a single-level inheritance model where an agent can extend exactly one base agent. Base agents are defined in `pkg/agents/meta/bases/` directory, while concrete agents (the variants) are in `pkg/agents/meta/`.

**Key benefits:**
- Reduce configuration duplication across related agents
- Establish common baselines for agent families
- Enable consistent tool and skill sets across variants
- Simplify maintenance of shared attributes

## The `extends` Field

To make an agent inherit from a base, add the `extends` field to your agent's YAML file:

```yaml
name: planner
description: Task planner. Breaks features into actionable tasks.
extends: planner  # Extends pkg/agents/meta/bases/planner.yaml
model: main
```

The `extends` value is the base agent's filename (without `.yaml` extension) from the `pkg/agents/meta/bases/` directory.

**Example structure:**
```
pkg/agents/meta/
├── bases/
│   ├── planner.yaml      # Base agent
│   ├── debugger.yaml     # Base agent
│   └── task.yaml         # Base agent
├── planner.yaml          # Extends planner
├── planner-fast.yaml     # Extends planner
├── planner-heavy.yaml    # Extends planner
└── ...
```

## Inheritance Semantics

The `mergeAgents()` function in `internal/cli/init.go` defines how base and variant configurations combine.

### Non-Slice Fields (Override)

Non-slice fields follow a **simple override** pattern: if the variant provides a non-empty value, it replaces the base value. If the variant's value is empty, the base value is used.

**Override behavior applies to:**
- `name` - Agent name
- `description` - Agent description
- `model` - Model (fast/main/heavy)
- `color` - Display color (hex code)
- `role` - Agent role (planning/execution/validation/research)
- `complexity` - Complexity level (simple/standard/heavy)

**Example:**

```yaml
# Base: pkg/agents/meta/bases/planner.yaml
name: ent:planner
description: Task planner. Breaks features into actionable tasks.
model: main
color: "#32CD32"

# Variant: pkg/agents/meta/planner.yaml
name: planner
description: Task planner. Breaks features into actionable tasks.
extends: planner
model: main
color: '#32CD32'

# Result after merge:
name: planner              # Variant overrides base
description: Task planner. Breaks features into actionable tasks.  # Same
model: main                # Same
color: '#32CD32'           # Different format, but same value
```

### Slice Fields (Additive Merge)

Slice fields follow an **additive merge** pattern: values from both base and variant are combined, with automatic deduplication.

**Additive merge applies to:**
- `skills` - Pre-loaded skills
- `tools` - Individual tools
- `toolPresets` - Tool preset groups
- `disallowedToolPresets` - Disallowed tool preset groups
- `disallowedTools` - Individually disallowed tools
- `dependencies` - Agent dependencies

The `mergeSlices()` helper function implements this:
1. Add all base values (maintaining order)
2. Add variant values not already present
3. Return deduplicated result

**Example:**

```yaml
# Base: pkg/agents/meta/bases/planner.yaml
skills:
  - go-arch
  - go-code
toolPresets:
  - readonly
  - serena-analysis

# Variant: pkg/agents/meta/planner.yaml
extends: planner
skills:
  - go-test  # New skill not in base
toolPresets:
  - planning  # New preset
tools:
  - glob
  - list

# Result after merge:
skills:
  - go-arch    # From base
  - go-code    # From base
  - go-test    # From variant (added)
toolPresets:
  - readonly          # From base
  - serena-analysis   # From base
  - planning          # From variant (added)
tools:
  - glob   # From variant
  - list   # From variant
```

## Examples

### Example 1: Basic Inheritance

A planner agent extending the planner base to add specific tools:

```yaml
# Base: pkg/agents/meta/bases/planner.yaml
name: ent:planner
description: Task planner. Breaks features into actionable tasks.
model: main
skills:
  - go-arch
  - go-code
toolPresets:
  - readonly
  - serena-analysis

# Variant: pkg/agents/meta/planner.yaml
name: planner
description: Task planner. Breaks features into actionable tasks.
extends: planner
model: main
toolPresets:
  - planning
tools:
  - glob
  - list
  - todoread
  - todowrite
  - skill

# Result: Agent gets base skills + presets + variant-specific additions
```

### Example 2: Skill Inheritance

A debugger agent adding test and performance skills on top of the base:

```yaml
# Base: pkg/agents/meta/bases/debugger.yaml
name: ent:debugger
description: Standard debugging. Systematic issue investigation and resolution.
model: main
skills:
  - go-code
  - debug-core
toolPresets:
  - editing
  - serena-analysis

# Variant: pkg/agents/meta/debugger.yaml
name: debugger
description: Standard debugging. Systematic issue investigation and resolution.
extends: debugger
skills:
  - go-test      # Added
  - debug-core   # Duplicate (already in base) - deduplicated
  - go-perf      # Added
  - go-code      # Duplicate (already in base) - deduplicated
tools:
  - list
  - todoread
  - todowrite
  - skill

# Result: skills = [go-code, debug-core, go-test, go-perf]
# Duplicates are automatically removed
```

### Example 3: Tool Preset Inheritance

Using tool presets to quickly add groups of related tools:

```yaml
# Base defines standard presets
# pkg/agents/meta/bases/planner.yaml
toolPresets:
  - readonly        # Read, Glob, Grep
  - serena-analysis # Serena semantic analysis tools

# Variant adds planning-specific preset
# pkg/agents/meta/planner.yaml
extends: planner
toolPresets:
  - planning        # Read, Glob, Grep + Task tools + Serena tools

# Result: Agent gets all tools from all three presets
# Tool expansion happens after merging (expandToolPresets())
```

The tool presets are defined in `pkg/agents/presets/tools.yaml`:

```yaml
presets:
  readonly:
    - Read
    - Glob
    - Grep
  planning:
    - Read
    - Glob
    - Grep
    - TaskCreate
    - TaskUpdate
    - TaskList
    - TaskGet
    - mcp__plugin_serena_serena__get_symbols_overview
    # ... more tools
```

## Best Practices

### When to Use Inheritance vs. Standalone Agents

**Use inheritance when:**
- Multiple agents share significant configuration (tools, skills, presets)
- You have agent families (e.g., planner/planner-fast/planner-heavy)
- You want to enforce consistent baselines across variants
- Reducing duplication would simplify maintenance

**Use standalone agents when:**
- The agent is unique with no shared configuration
- The agent configuration is small and simple
- You need complete control over all fields without inheritance complexity
- The agent is a one-off or experimental

### Structuring Base Agents

**Base agent conventions:**
1. **Focus on shared attributes**: Only include fields that truly apply to all variants
2. **Avoid variant-specific details**: Keep base agents generic and minimal
3. **Use tool presets**: Leverage presets for common tool groups
4. **Keep descriptions generic**: Base descriptions should be broad enough to apply to variants
5. **Set sensible defaults**: Model, color, and role in base if consistent across variants

**Example of a well-structured base:**

```yaml
# pkg/agents/meta/bases/planner.yaml
# Good: Minimal, focused on shared attributes
name: ent:planner
description: Task planner. Breaks features into actionable tasks.
model: main
skills:
  - go-arch
  - go-code
toolPresets:
  - readonly
  - serena-analysis
```

```yaml
# Bad: Too many variant-specific details
# pkg/agents/meta/bases/planner.yaml (not recommended)
name: ent:planner
description: Task planner for main model only.
model: main  # This assumes all variants use main - use variant override instead
skills:
  - go-arch
  - go-code
  - go-test  # Not all planners need testing
toolPresets:
  - readonly
  - serena-analysis
  - planning  # Some planners might not need planning tools
tools:
  - glob      # Should be in presets or variant-specific
  - list
  - todoread
```

### Common Pitfalls to Avoid

**1. Overriding required fields:**

```yaml
# Variant: Don't leave critical fields empty
name: ""  # Bad - empty name won't override base
description: ""  # Bad - empty description won't override base
extends: planner

# Correct: Override with actual values
name: planner-fast
description: Fast task planner for quick assessments.
extends: planner
```

**2. Duplicate values in slices (harmless but messy):**

```yaml
# Variant: Duplicates are deduplicated automatically, but avoid them
extends: planner
skills:
  - go-arch    # Already in base - redundant
  - go-code    # Already in base - redundant
  - go-test    # Good - only unique additions

# Better: Only list additions
extends: planner
skills:
  - go-test
```

**3. Incorrect base name reference:**

```yaml
# Bad: Using full path or wrong name
extends: pkg/agents/meta/bases/planner  # Wrong
extends: planner.yaml                    # Wrong (no .yaml extension)

# Correct: Use filename without extension
extends: planner
```

**4. Missing base file:**

The system validates that `extends` references exist. If you reference a non-existent base:

```
Error: planner-fast.yaml: extends unknown base: planner-base
```

Ensure the base file exists in `pkg/agents/meta/bases/`.

**5. Confusing model with role:**

```yaml
# Model is about capability level (fast/main/heavy)
model: main

# Role is about functional area (planning/execution/validation/research)
role: planning

# These are different - don't confuse them
```

## Implementation Details

### Merge Algorithm

The `mergeAgents()` function in `internal/cli/init.go` (lines 199-229):

```go
func mergeAgents(base, variant *agentMeta) *agentMeta {
    merged := *base

    // Non-slice fields: override if variant value is non-empty
    if variant.Name != "" {
        merged.Name = variant.Name
    }
    // ... similar for other non-slice fields

    // Slice fields: additive merge with deduplication
    merged.Skills = mergeSlices(base.Skills, variant.Skills)
    merged.Tools = mergeSlices(base.Tools, variant.Tools)
    merged.ToolPresets = mergeSlices(base.ToolPresets, variant.ToolPresets)
    merged.DisallowedToolPresets = mergeSlices(base.DisallowedToolPresets, variant.DisallowedToolPresets)
    merged.DisallowedTools = mergeSlices(base.DisallowedTools, variant.DisallowedTools)
    merged.Dependencies = mergeSlices(base.Dependencies, variant.Dependencies)

    return &merged
}
```

The `mergeSlices()` helper (lines 178-197):

```go
func mergeSlices(base, variant []string) []string {
    seen := make(map[string]bool)
    var result []string

    // Add base values first (maintains order)
    for _, v := range base {
        if !seen[v] {
            seen[v] = true
            result = append(result, v)
        }
    }

    // Add variant values (only if not seen)
    for _, v := range variant {
        if !seen[v] {
            seen[v] = true
            result = append(result, v)
        }
    }

    return result
}
```

### Tool Expansion Flow

After merging, tool presets are expanded:

1. Load agent (with inheritance merged)
2. Load tool presets from `pkg/agents/presets/tools.yaml`
3. Call `expandToolPresets()`:
   - Add individual tools from `tools` field
   - Add tools from all `toolPresets`
   - Remove tools in `disallowedTools`
   - Remove tools from all `disallowedToolPresets`
4. Result: final tool list for the agent

## See Also

- [Agent Format Specification](docs/tools/claude-code-extension-guide.md) - Complete agent YAML format
- [Tool Presets](pkg/agents/presets/tools.yaml) - Available tool preset definitions
- [Refactoring Guide](docs/REFACTORING_GUIDE.md) - Command replacement patterns and best practices
- [Source Code](internal/cli/init.go) - `mergeAgents()` and `mergeSlices()` implementation
