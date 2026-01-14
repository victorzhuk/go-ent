# Agent Refactoring Guide

## Summary of Changes

This guide documents the required changes to align all agents with best practices established in the driver orchestrator.

---

## 1. Tool Configuration Changes

### Add Missing Tools (All Agents)

```yaml
tools:
  # ... existing tools ...
  todoread: true      # ADD - Read task state
  todowrite: true     # ADD - Update task progress
  skill: true         # ADD - Load SKILL.md files
  list: true          # ADD - Directory listing
```

### Agents That Need TODO Tools

| Agent | Add todoread | Add todowrite | Reason |
|-------|--------------|---------------|--------|
| acceptor | ✓ | ✓ | Track acceptance state |
| architect | ✓ | ✓ | Track design decisions |
| coder | ✓ | ✓ | Track implementation progress |
| debugger | ✓ | ✓ | Track investigation |
| debugger-fast | ✓ | ✓ | Track quick fixes |
| debugger-heavy | ✓ | ✓ | Track complex debugging |
| decomposer | ✓ | ✓ | Task breakdown tracking |
| planner | ✓ | ✓ | Planning progress |
| planner-heavy | ✓ | ✓ | Complex planning |
| reproducer | ✓ | ✓ | Bug reproduction tracking |
| researcher | ✓ | ✓ | Research findings |
| task-fast | ✓ | ✗ | Read-only assessment |
| task-heavy | ✓ | ✓ | Complex task analysis |

---

## 2. Command Replacements

### Replace grep with rg

**Find and replace in all agent files:**

| Original | Replace With | Context |
|----------|--------------|---------|
| `grep -rn "pattern"` | `rg -n "pattern"` | General search |
| `grep -r "pattern"` | `rg "pattern"` | Recursive search |
| `grep -rn "pattern" internal/` | `rg -n "pattern" internal/` | Directory search |
| `grep -r "import.*transport" internal/domain/` | `rg "import.*transport" internal/domain/` | Pattern in dir |
| `grep -rn "func New" internal/repository/` | `rg -n "func New" internal/repository/` | Function search |
| `grep -A 10 "panic"` | `rg -A 10 "panic"` | Context after |
| `grep -C 3 "error"` | `rg -C 3 "error"` | Context around |

**File type filtering (use rg -t):**

| Original | Replace With |
|----------|--------------|
| `grep -rn "pattern" *.go` | `rg -tgo "pattern"` |
| `grep -rn "pattern" --include="*.go"` | `rg -tgo "pattern"` |
| `grep -rn "pattern" --include="*.ts"` | `rg -tts "pattern"` |

### Replace find with fd

| Original | Replace With |
|----------|--------------|
| `find internal -type d -depth 2` | `fd -t d --max-depth 2 internal` |
| `find . -name "*.go" -type f` | `fd -e go` |
| `find internal -type d` | `fd -t d internal` |
| `find . -name "*.md"` | `fd -e md` |

### Replace cat | grep with direct rg

| Original | Replace With |
|----------|--------------|
| `cat file.go \| grep "pattern"` | `rg -n "pattern" file.go` |
| `cat logs/* \| grep "error"` | `rg "error" logs/` |

---

## 3. Add Optimal Tooling Section

Add this section to **ALL agents** after the frontmatter:

```markdown
## Optimal Tooling

| Instead of | Use | Reason |
|------------|-----|--------|
| `grep -rn` | `rg -n` | 10x faster, respects .gitignore |
| `grep -r "pattern"` | `rg -tgo "pattern"` | File type filtering |
| `find . -name` | `fd` | 5x faster |
| `cat file \| grep` | `rg -n pattern file` | Direct search |
```

---

## 4. Add Context Gathering Phase

Add context gathering as **first step** in workflow for execution agents:

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

---

## 5. Add TODO Tracking

### At Workflow Start

```bash
todowrite create "{agent}: {task-description}"
```

### During Work

```bash
todowrite update "{step}: in_progress"
todowrite add-context "{key}: {value}"
```

### At Completion

```bash
todowrite update "{task}: ✅ completed"
```

---

## 6. Specific Agent Changes

### acceptor.md

```diff
tools:
  read: true
  bash: true
  grep: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true

- grep -rn "WHEN.*THEN" openspec/
+ rg -n "WHEN.*THEN" openspec/
```

### architect.md

```diff
tools:
  read: true
  glob: true
  grep: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true

- grep -r "type.*Repository" internal/
+ rg -tgo "type.*Repository" internal/
```

### coder.md

```diff
tools:
  read: true
  write: true
  edit: true
  bash: true
  glob: true
  grep: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true

- grep -rn "func New" internal/repository/
+ rg -tgo "func New" internal/repository/
```

### debugger.md, debugger-fast.md, debugger-heavy.md

```diff
tools:
  # ... existing ...
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true

- grep -rn "error message" internal/
+ rg -n "error message" internal/

- grep -r "error\|panic" logs/
+ rg "error|panic" logs/

- git diff HEAD~5 -- internal/
+ git diff HEAD~5 -- internal/   # (git is fine)
```

### decomposer.md

```diff
tools:
  read: true
  grep: true
  glob: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true
```

### planner.md, planner-fast.md, planner-heavy.md

```diff
tools:
  read: true
  grep: true
+ glob: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true

- find internal -type d -depth 2
+ fd -t d --max-depth 2 internal

- grep -rn "func New" internal/repository/
+ rg -tgo "func New" internal/repository/
```

### reproducer.md

```diff
tools:
  read: true
  write: true
  bash: true
  grep: true
  glob: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true

- grep -r "error\|panic" logs/
+ rg "error|panic" logs/
```

### researcher.md

```diff
tools:
  read: true
  bash: true
  grep: true
  glob: true
+ list: true
+ webfetch: true      # ADD - Fetch external docs
+ websearch: true     # ADD - Search external info
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true
```

### reviewer.md

```diff
tools:
  read: true
  grep: true
  glob: true
  bash: true
+ list: true
+ todoread: true
+ skill: true
  mcp__plugin_serena_serena: true

- grep -r "import.*transport" internal/domain/
+ rg "import.*transport" internal/domain/

- grep -rn "applicationConfig\|userRepository" internal/
+ rg -n "applicationConfig|userRepository" internal/

- grep -rn "// Create\|// Get\|// Set" internal/
+ rg -n "// Create|// Get|// Set" internal/
```

### task-fast.md

```diff
tools:
  read: true
  grep: true
  glob: true
+ list: true
+ todoread: true
+ skill: true
  mcp__plugin_serena_serena: true

- cat openspec/changes/{change-id}/proposal.md
+ read openspec/changes/{change-id}/proposal.md
```

### task-heavy.md

```diff
tools:
  read: true
  grep: true
  glob: true
+ list: true
+ todoread: true
+ todowrite: true
+ skill: true
  mcp__plugin_serena_serena: true
```

### tester.md

```diff
tools:
  read: true
  bash: true
  grep: true
  glob: true
+ list: true
+ todoread: true
+ skill: true
```

---

## 7. MCP Tool Naming

The current format `mcp__plugin_serena_serena` appears to be a plugin-specific naming convention. Verify this matches your OpenCode MCP configuration.

Alternative formats depending on setup:
- `mcp__serena` (if using direct MCP)
- `mcp__plugin_serena_serena` (if using plugin wrapper)

---

## 8. Model Tiers

Current model assignments look correct:
- `fast` - Quick triage, simple tasks
- `main` - Standard implementation
- `heavy` - Complex analysis, architecture

Ensure these map to actual models in your `opencode.json`:

```json
{
  "models": {
    "fast": "anthropic/claude-haiku-4-5",
    "main": "anthropic/claude-sonnet-4-5-20250929",
    "heavy": "anthropic/claude-opus-4-5-20251101"
  }
}
```

---

## 9. Checklist for Each Agent

When refactoring each agent, verify:

- [ ] Added `todoread: true` (if reads task state)
- [ ] Added `todowrite: true` (if updates progress)
- [ ] Added `skill: true`
- [ ] Added `list: true`
- [ ] Replaced all `grep` with `rg`
- [ ] Replaced all `find` with `fd`
- [ ] Replaced `cat file | grep` with `rg -n`
- [ ] Added "Optimal Tooling" section
- [ ] Added context gathering as first workflow step
- [ ] Added TODO tracking (create/update calls)
- [ ] Verified MCP tool name matches config
