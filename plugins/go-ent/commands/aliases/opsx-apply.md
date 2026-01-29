---
name: opsx:apply
description: Execute OpenSpec tasks (alias for ent:apply)
---

# OpenSpec: Apply Tasks

**Alias for `/ent:apply`** - OpenSpec-compatible command name.

Execute tasks from the OpenSpec change registry with TDD and validation.

## Usage

```
/opsx:apply
```

Same workflow as `/ent:apply`:
1. Fetch next available task from registry
2. Execute with test-driven development
3. Validate implementation
4. Mark complete and continue

## See Also

- `/ent:apply` - Primary command name
- `/opsx:new` - Create new change
- `/opsx:archive` - Archive completed change
