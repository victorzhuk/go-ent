---
description: Execute tasks from OpenSpec change proposal
allowed-tools: Read, Bash, Edit, mcp__plugin_serena_serena, mcp__go_ent__go_ent_registry_next, mcp__go_ent__go_ent_registry_update
---

# Apply Change

Input: `$ARGUMENTS` (change-id or empty for auto)

## Actions

1. **Get next task** (registry-aware):
   ```
   If $ARGUMENTS is empty:
     Use go_ent_registry_next with path="." to get recommended task
   Else:
     Use go_ent_registry_next with path=".", change_id=$ARGUMENTS
   ```

2. **Load change context**:
   ```bash
   cat openspec/changes/{change-id}/tasks.md
   cat openspec/changes/{change-id}/proposal.md
   cat openspec/changes/{change-id}/design.md (if exists)
   ```

3. **Mark task as in_progress**:
   ```
   Use go_ent_registry_update:
     task_id="{change-id}/{task-num}"
     status="in_progress"
   ```

4. **Execute task**:
   - Use Serena for code context
   - Implement following task requirements
   - Validate: `go build && go test -race`

5. **Mark task complete**:
   ```
   Use go_ent_registry_update:
     task_id="{change-id}/{task-num}"
     status="completed"
   ```

   Also update tasks.md checkbox:
   ```markdown
   - [x] X.Y Task description
   ```

6. **Report progress**:
   ```
   ✅ Task complete: {change-id}/{task-num}
   Progress: 80% (8/10)

   Next recommended: {next-task-id} (priority: {priority})
   Reason: {reason from registry_next}

   Continue: /go-ent:apply (auto-pick next)
   ```

7. **If all tasks done**, update proposal status to COMPLETE and suggest `/go-ent:archive`.

## Notes

- Registry integration provides smart "next task" selection based on priority and dependencies
- Cross-change dependencies are respected
- Tasks blocked by dependencies are automatically skipped
