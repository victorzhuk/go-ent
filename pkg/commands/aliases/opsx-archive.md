---
name: opsx:archive
description: Archive completed OpenSpec change (alias for ent:archive)
---

# OpenSpec: Archive Change

**Alias for `/ent:archive`** - OpenSpec-compatible command name.

Archive a completed change after deployment verification.

## Usage

```
/opsx:archive <change-id>
```

Same workflow as `/ent:archive`:
1. Verify all tasks complete
2. Merge delta specs to main specs
3. Move from active/ to archive/
4. Update registry

## Examples

```
/opsx:archive add-user-auth
/opsx:archive fix-memory-leak-123
```

## See Also

- `/ent:archive` - Primary command name
- `/opsx:new` - Create new change
- `/opsx:apply` - Execute tasks
