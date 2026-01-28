# Init Command Model Resolution Fix

## Problem
After restoring the `init` command, model category names like `heavy` were being passed directly to the output templates instead of being resolved to actual model IDs like `claude-opus-4-5-20251101`.

## Root Cause
The `init` command was using the raw `model` value from agent meta YAML files (e.g., `model: heavy`) without resolving it through the model configuration system.

## Solution
Integrated the existing model resolution system (`internal/model/resolver.go`) into the `init` command:

1. Added import for `"github.com/victorzhuk/go-ent/internal/model"`

2. In `newInitCmd()`, before rendering each agent:
   - Load global model config: `global, _ := model.LoadGlobal()`
   - Load project model config: `project, _ := model.LoadProject(".")`
   - Merge configs: `cfg := model.Merge(global, project)`
   - Create resolver for the current tool: `resolver := model.NewResolver(cfg, tool)`
   - Resolve model before rendering: `meta.Model = resolver.ResolveAgent(meta.Model)`

## Result
- Claude Code agents with `model: heavy` now output `model: claude-opus-4-5-20250514`
- OpenCode agents with `model: heavy` now output `model: kimi-for-coding/kimi-k2-thinking`
- Category names are properly resolved to actual model IDs for each runtime

## Files Modified
- `internal/cli/init.go` - Added model resolution logic
