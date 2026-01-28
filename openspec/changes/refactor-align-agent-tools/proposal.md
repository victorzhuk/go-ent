# Proposal: Align Agent Tool Configurations

## Status: pending

## Why

**Current Problem:**

The go-ent agent system has 14+ agents with inconsistent tool configurations, causing inefficiencies and outdated patterns:

- **12 agents missing critical tools**: `todoread`, `todowrite`, `skill`, and `list` tools are absent from most agent configurations
- **Outdated search patterns**: 47+ instances of `grep` commands in agent prompts that should use `rg` (ripgrep) for 10x faster searches
- **No optimal tooling guidance**: Agents lack the "Optimal Tooling" section that guides efficient command usage
- **Missing context gathering**: Workflows don't start with standardized context gathering phases

**Quantified Impact:**
- 14 agent files require updates (`.claude/agents/*.md`)
- ~47 grep patterns need replacement with rg equivalents
- 12 agents need tool additions (4 tools each = 48 tool additions)
- Estimated time waste: 15-20% slower searches, inconsistent task tracking

## What Changes

### Before

```yaml
# .claude/agents/coder.md (current)
tools:
  read: true
  write: true
  edit: true
  bash: true
  glob: true
  grep: true
  mcp__plugin_serena_serena: true
```

Agent prompts use:
```bash
grep -rn "func New" internal/repository/
grep -r "type.*Repository" internal/
```

### After

```yaml
# .claude/agents/coder.md (updated)
tools:
  read: true
  write: true
  edit: true
  bash: true
  glob: true
  grep: true
  list: true           # ADD - Directory listing
  todoread: true       # ADD - Read task state
  todowrite: true      # ADD - Update task progress
  skill: true          # ADD - Load SKILL.md files
  mcp__plugin_serena_serena: true
```

Agent prompts use:
```bash
rg -tgo "func New" internal/repository/
rg -tgo "type.*Repository" internal/
```

### Key Components

| File | Change Description |
|------|-------------------|
| `.claude/agents/acceptor.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/architect.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/coder.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/debugger.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/debugger-fast.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/debugger-heavy.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/decomposer.md` | Add `list`, `todoread`, `todowrite`, `skill` tools |
| `.claude/agents/planner.md` | Add `glob`, `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg, find with fd |
| `.claude/agents/planner-fast.md` | Add `glob`, `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg, find with fd |
| `.claude/agents/planner-heavy.md` | Add `glob`, `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg, find with fd |
| `.claude/agents/reproducer.md` | Add `list`, `todoread`, `todowrite`, `skill` tools; replace grep with rg |
| `.claude/agents/researcher.md` | Add `list`, `webfetch`, `websearch`, `todoread`, `todowrite`, `skill` tools |
| `.claude/agents/reviewer.md` | Add `list`, `todoread`, `skill` tools; replace grep with rg |
| `.claude/agents/task-fast.md` | Add `list`, `todoread`, `skill` tools |
| `.claude/agents/task-heavy.md` | Add `list`, `todoread`, `todowrite`, `skill` tools |
| `.claude/agents/tester.md` | Add `list`, `todoread`, `skill` tools |

**New Section Added to All Agents:**
```markdown
## Optimal Tooling

| Instead of | Use | Reason |
|------------|-----|--------|
| `grep -rn` | `rg -n` | 10x faster, respects .gitignore |
| `grep -r "pattern"` | `rg -tgo "pattern"` | File type filtering |
| `find . -name` | `fd` | 5x faster |
| `cat file | grep` | `rg -n pattern file` | Direct search |
```

**New Workflow Phase Added:**
```markdown
### 1. Context Gathering

```bash
# Check current task state
todoread

# Load relevant skill
skill {skill-name}

# Explore project structure
list internal
glob "**/*.go"

# Search with rg (not grep)
rg -tgo "pattern" internal/
```
```

## Impact

**Breaking Changes:** None - these are agent configuration improvements only

**Performance Improvements:**
- Search operations: 10x faster with rg vs grep
- File finding: 5x faster with fd vs find
- Task tracking: Consistent TODO state management

**Developer Experience:**
- Consistent tooling across all agents
- Clear guidance on optimal command usage
- Standardized context gathering workflow

## Success Criteria

- [ ] All 14 agent files updated with missing tools
- [ ] All grep commands replaced with rg equivalents
- [ ] All find commands replaced with fd equivalents
- [ ] "Optimal Tooling" section added to all agents
- [ ] Context gathering phase added to execution agents
- [ ] Agent configuration validated (no YAML syntax errors)
- [ ] Sample agent invocations tested successfully

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Tool availability varies by environment | Low | rg and fd are standard in modern dev environments; fallback to grep/find documented |
| Agent behavior changes unexpectedly | Low | Changes are additive (new tools) or performance-only (rg vs grep) |
| MCP tool naming mismatch | Medium | Verify `mcp__plugin_serena_serena` matches actual OpenCode MCP config |
| Breaking existing workflows | Low | No removal of existing tools, only additions and pattern replacements |

## Alternatives Considered

1. **Keep grep/find** - Rejected: 10x performance improvement with rg/fd is significant
2. **Add tools incrementally** - Rejected: Inconsistent tooling causes confusion; better to align all at once
3. **Create base agent template** - Considered for future: Inheritance would reduce duplication but requires larger refactoring

## Related Documentation

- `docs/REFACTORING_GUIDE.md` - Detailed command replacement patterns
- `docs/tools/claude-code-extension-guide.md` - Agent format specification
- `plugins/go-ent/agents/presets/tools.yaml` - Tool preset definitions
