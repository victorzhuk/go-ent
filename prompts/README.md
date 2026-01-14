# Universal Agent Prompt Templates

Industrial-grade, KISS/YAGNI compliant prompt templates for AI agents.

## Design Principles

### KISS - Keep It Simple
- One sentence role definition
- Core protocol ≤5 words
- Max 7 rules per agent
- Minimal workflow phases

### YAGNI - You Aren't Gonna Need It
- No philosophy sections
- No theory explanations
- No verbose examples
- No redundant checklists

### Industrial Standards
- Consistent structure across all agents
- Measurable outputs (confidence levels, hours, counts)
- Clear handoff protocols
- Explicit tool preferences

---

## Template Structure

Every agent follows this structure:

```
# {Name}

{One sentence role definition}

## Core Protocol

**VERB → VERB → VERB → VERB**

---

{{> _shared}}

## Workflow

### 1. {Phase}
{Commands/steps}

### 2. {Phase}
{Commands/steps}

---

## Output

{Minimal output format}

---

## Rules

1. {Rule}
2. {Rule}
...
7. {Rule max}

## Handoff

- @{agent} → {condition}
```

---

## Files

| File | Purpose |
|------|---------|
| `_shared.md` | Common tooling table (partial) |
| `agents.yaml` | Agent definitions + metadata |
| `{name}.md` | Prompt template |

---

## Usage

### With Go text/template

```go
tmpl := template.Must(template.ParseFiles("driver.md", "_shared.md"))
tmpl.Execute(w, data)
```

### With Mustache

```
{{> _shared}}  // includes _shared.md partial
```

### With Jinja2

```
{% include '_shared.md' %}
```

---

## Agent Catalog

| Agent | Model | Purpose |
|-------|-------|---------|
| driver | main | Orchestration |
| coder | main | Implementation |
| debugger | main | Bug investigation |
| debugger-fast | fast | Quick fixes |
| debugger-heavy | heavy | Complex issues |
| planner | main | Task planning |
| reviewer | heavy | Code review |
| researcher | heavy | Root cause analysis |
| architect | heavy | System design |
| decomposer | heavy | Task breakdown |
| acceptor | heavy | Acceptance testing |
| tester | fast | Test execution |
| reproducer | fast | Bug reproduction |

---

## Tool Preferences

All agents use this tooling table:

| Instead of | Use | Reason |
|------------|-----|--------|
| `grep -rn` | `rg -n` | 10x faster |
| `grep "x"` | `rg -tgo "x"` | Go files only |
| `find -name` | `fd` | 5x faster |
| `cat \| grep` | `rg pattern file` | Direct |

---

## Platform Generation

`agents.yaml` defines mappings for:
- **OpenCode**: lowercase tools, YAML arrays
- **Claude Code**: PascalCase tools, CSV strings

Template engine reads `agents.yaml` and generates platform-specific frontmatter.
