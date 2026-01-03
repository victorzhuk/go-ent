---
description: Archive completed OpenSpec change
allowed-tools: Read, Bash, Edit
---

# Archive Change

Input: `$ARGUMENTS` (change-id)

## Actions

1. **Verify completion**:
   ```bash
   cat openspec/changes/{id}/proposal.md   # Status: COMPLETE
   cat openspec/changes/{id}/tasks.md      # All [x]
   ```

2. **Final validation**:
   ```bash
   go build ./...
   go test ./... -race
   golangci-lint run
   ```

3. **Update status**:
   ```
   **Status:** COMPLETE → ARCHIVED
   **Archived:** {date}
   ```

4. **Move to archive**:
   ```bash
   mv openspec/changes/{id} openspec/archive/
   ```

5. **Report**:
   ```
   ═══════════════════════════════════════════
   📦 CHANGE ARCHIVED
   ═══════════════════════════════════════════
     ID: {id}
     Location: openspec/archive/{id}/

   💡 SUGGESTED GIT
     git add -A
     git commit -m "feat({id}): {title}"
   ═══════════════════════════════════════════
   ```
