---
name: opsx:new
description: Start a new OpenSpec change (alias for ent:plan)
---

# OpenSpec: New Change

**Alias for `/ent:plan`** - OpenSpec-compatible command name.

Create a complete OpenSpec change proposal with research and task breakdown.

## Usage

```
/opsx:new <description>
```

Same workflow as `/ent:plan`:
1. Clarify requirements
2. Research codebase
3. Design solution
4. Decompose into tasks

## Examples

```
/opsx:new Add user authentication system
/opsx:new Fix memory leak in agent execution
/opsx:new Refactor database layer for better testability
```

## See Also

- `/ent:plan` - Primary command name
- `/opsx:apply` - Execute tasks
- `/opsx:archive` - Archive completed change
