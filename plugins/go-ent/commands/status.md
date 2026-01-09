---
description: View status of all OpenSpec changes
allowed-tools: Read, Bash, mcp__go_ent__registry_list, mcp__go_ent__registry_next
---

# OpenSpec Status

## Actions

1. Check if registry exists and get stats:
   ```
   Use registry_list with path="."
   Use registry_next with path=".", count=1
   ```

2. If registry exists, display registry summary first.

3. Then list changes:
   ```bash
   ls openspec/changes/
   ls openspec/archive/ 2>/dev/null || true
   ```

Parse each change's `proposal.md` and `tasks.md`.

## Output (with registry)

```
═══════════════════════════════════════════
            OPENSPEC STATUS
═══════════════════════════════════════════

📊 REGISTRY SUMMARY
  Total tasks:  37
  Completed:    22 (59%)
  In Progress:  2
  Blocked:      3
  Next:         add-auth/2.1 (priority: high)

📋 IN PROGRESS
  add-user-auth    ████████░░ 80%  (8/10)

📝 PROPOSED
  add-notifications Not started   (0/12)

✅ COMPLETE (ready to archive)
  update-api-v2    ██████████ 100%

📦 ARCHIVED (recent)
  add-health-checks  2024-01-01

═══════════════════════════════════════════

Commands:
  /go-ent:registry next     Get next task
  /go-ent:apply {id}        Continue work
  /go-ent:archive {id}      Archive complete
```

## Output (no registry)

If registry doesn't exist, show traditional status without registry summary.
