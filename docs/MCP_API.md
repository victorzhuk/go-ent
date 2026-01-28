# MCP Tools API Reference

Complete reference for go-ent MCP tools.

---

## Table of Contents

- [Overview](#overview)
- [Tool Categories](#tool-categories)
- [Spec Management Tools](#spec-management-tools)
- [Registry Tools](#registry-tools)
- [Workflow Tools](#workflow-tools)
- [Error Handling](#error-handling)
- [Examples](#examples)

---

## Overview

go-ent exposes **30+ MCP tools** for spec-driven development. Tools are organized by category:

- **Spec Management** (8 tools): Create, read, update, delete specs and changes
- **Registry** (6 tools): Task management and dependency tracking
- **Workflow** (3 tools): Workflow state management
- **Loop** (4 tools): Autonomous execution loops
- **Generation** (4 tools): Code generation from templates

### Tool Naming Convention

Tools follow the pattern: `<category>_<action>`

Examples:
- `spec_init` - Initialize openspec folder
- `registry_list` - List registry tasks
- `workflow_start` - Start workflow

---

## Tool Categories

### Quick Reference

| Category | Tools | Purpose |
|----------|-------|---------|
| `spec_*` | 8 | Spec and change management |
| `registry_*` | 6 | Task registry operations |
| `workflow_*` | 3 | Workflow state management |
| `loop_*` | 4 | Autonomous loop control |
| `generate_*` | 4 | Code generation |

---

## Spec Management Tools

### `spec_init`

Initialize openspec folder structure in a project.

**Parameters**:
```json
{
  "path": "string (required)",
  "name": "string",
  "module": "string",
  "description": "string",
  "conventions": {
    "key": "value"
  }
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Initialized openspec at /path/to/project/openspec"
    }
  ]
}
```

**Example**:
```javascript
{
  "path": "/home/user/myproject",
  "name": "My Project",
  "module": "github.com/user/myproject",
  "description": "A Go microservice",
  "conventions": {
    "error_handling": "wrap with context",
    "logging": "zerolog"
  }
}
```

**Creates**:
```
myproject/
└── openspec/
    ├── config.yaml
    ├── specs/
    ├── changes/
    └── changes/archive/
```

---

### `spec_list`

List specs, changes, or tasks.

**Parameters**:
```json
{
  "path": "string (required)",
  "type": "string (required)",  // "specs", "changes", or "tasks"
  "change_id": "string"          // required if type="tasks"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Found 3 specs:\n1. api/handlers\n2. domain/models\n3. infra/database"
    }
  ]
}
```

**Example**:
```javascript
// List all specs
{ "path": "/path/to/project", "type": "specs" }

// List active changes
{ "path": "/path/to/project", "type": "changes" }

// List tasks in a change
{
  "path": "/path/to/project",
  "type": "tasks",
  "change_id": "add-auth"
}
```

---

### `spec_show`

Show detailed content of a spec, change, or task.

**Parameters**:
```json
{
  "path": "string (required)",
  "type": "string (required)",  // "spec", "change", or "task"
  "id": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "# API Handlers\n\n## Overview\n..."
    }
  ]
}
```

**Example**:
```javascript
// Show spec
{
  "path": "/path/to/project",
  "type": "spec",
  "id": "api/handlers"
}

// Show change proposal
{
  "path": "/path/to/project",
  "type": "change",
  "id": "add-auth"
}
```

---

### `spec_create`

Create a new spec or change.

**Parameters**:
```json
{
  "path": "string (required)",
  "type": "string (required)",  // "spec" or "change"
  "id": "string (required)",
  "content": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Created spec at openspec/specs/api/handlers.md"
    }
  ]
}
```

**Example**:
```javascript
{
  "path": "/path/to/project",
  "type": "change",
  "id": "add-auth-middleware",
  "content": "# Add Auth Middleware\n\n## Problem\n..."
}
```

**Creates**:
```
openspec/changes/add-auth-middleware/
└── proposal.md
```

---

### `spec_update`

Update existing spec or change.

**Parameters**:
```json
{
  "path": "string (required)",
  "type": "string (required)",
  "id": "string (required)",
  "content": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Updated spec api/handlers"
    }
  ]
}
```

---

### `spec_delete`

Delete a spec or change.

**Parameters**:
```json
{
  "path": "string (required)",
  "type": "string (required)",
  "id": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Deleted change add-auth-middleware"
    }
  ]
}
```

---

### `spec_validate`

Validate spec or change format and content.

**Parameters**:
```json
{
  "path": "string (required)",
  "id": "string",          // validate specific item
  "strict": "boolean"      // strict validation mode
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Validation passed:\n✓ All specs valid\n✓ All changes valid\n✓ No broken references"
    }
  ]
}
```

**Validates**:
- YAML frontmatter syntax
- Required fields present
- File structure
- Cross-references
- Task format (if strict mode)

---

### `spec_archive`

Archive a completed change.

**Parameters**:
```json
{
  "path": "string (required)",
  "change_id": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Archived change add-auth-middleware to openspec/changes/archive/2026-01-28-add-auth-middleware/"
    }
  ]
}
```

**Process**:
1. Validates all tasks complete
2. Merges delta specs into main specs
3. Moves change to archive with timestamp
4. Updates change index

---

## Registry Tools

### `registry_list`

List all tasks in the registry.

**Parameters**:
```json
{
  "path": "string (required)",
  "filters": {
    "status": "string",      // "pending", "in_progress", "completed"
    "priority": "string",    // "low", "medium", "high"
    "change_id": "string"
  }
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Tasks:\n1. [pending] add-auth/1: Implement JWT validation\n2. [in_progress] add-auth/2: Add middleware\n3. [completed] add-auth/3: Write tests"
    }
  ]
}
```

---

### `registry_next`

Get next recommended task(s).

**Parameters**:
```json
{
  "path": "string (required)",
  "count": "number"        // default: 1
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Next task:\nID: add-auth/1\nTitle: Implement JWT validation\nPriority: high\nDependencies: none"
    }
  ]
}
```

**Algorithm**:
1. Filter by status (pending only)
2. Check dependencies satisfied
3. Sort by priority
4. Return top N

---

### `registry_update`

Update task status or metadata.

**Parameters**:
```json
{
  "path": "string (required)",
  "task_id": "string (required)",
  "updates": {
    "status": "string",      // "pending", "in_progress", "completed"
    "priority": "string",
    "assignee": "string",
    "notes": "string"
  }
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Updated task add-auth/1: status=completed"
    }
  ]
}
```

---

### `registry_deps`

Manage task dependencies.

**Parameters**:
```json
{
  "path": "string (required)",
  "task_id": "string (required)",
  "operation": "string (required)",  // "add", "remove", "list"
  "dependency_id": "string"          // required for add/remove
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Dependencies for add-auth/2:\n- add-auth/1 (completed)"
    }
  ]
}
```

---

### `registry_sync`

Sync registry with tasks.md files in changes.

**Parameters**:
```json
{
  "path": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Synced 12 tasks from 3 changes"
    }
  ]
}
```

**Process**:
1. Scan all `changes/*/tasks.md`
2. Parse task definitions
3. Update registry
4. Preserve status and metadata

---

## Workflow Tools

### `workflow_state`

Get or update workflow state.

**Parameters**:
```json
{
  "path": "string (required)",
  "change_id": "string (required)",
  "state": "string"        // if provided, updates state
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Workflow state: review\nPhase: draft → review → implement → archive"
    }
  ]
}
```

**States**:
- `draft` - Proposal being written
- `review` - Awaiting approval
- `revise` - Needs changes
- `implement` - Tasks being executed
- `archive` - Completed

---

### `workflow_start`

Start planning workflow for a change.

**Parameters**:
```json
{
  "path": "string (required)",
  "change_id": "string (required)",
  "description": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Started workflow for add-auth\nPhase 1: Clarify requirements\nNext: Ask clarifying questions"
    }
  ]
}
```

---

### `workflow_complete`

Mark workflow phase complete.

**Parameters**:
```json
{
  "path": "string (required)",
  "change_id": "string (required)",
  "phase": "string (required)"
}
```

**Returns**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Completed phase: clarify\nNext phase: research"
    }
  ]
}
```

---

## Error Handling

### Error Response Format

```json
{
  "content": [
    {
      "type": "text",
      "text": "Error: <error message>"
    }
  ],
  "isError": true
}
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `path is required` | Missing path parameter | Provide project path |
| `openspec not found` | No openspec folder | Run `spec_init` first |
| `change not found` | Invalid change ID | Check `spec_list` for valid IDs |
| `task has dependencies` | Blocking tasks not complete | Complete dependencies first |
| `invalid state transition` | Workflow state error | Follow workflow order |

---

## Examples

### Complete Workflow Example

```javascript
// 1. Initialize project
{
  "tool": "spec_init",
  "params": {
    "path": "/home/user/myapi",
    "name": "My API",
    "module": "github.com/user/myapi"
  }
}

// 2. Create change
{
  "tool": "spec_create",
  "params": {
    "path": "/home/user/myapi",
    "type": "change",
    "id": "add-auth",
    "content": "# Add Authentication\n\n## Problem\nNo user authentication...\n\n## Solution\nJWT middleware..."
  }
}

// 3. Start workflow
{
  "tool": "workflow_start",
  "params": {
    "path": "/home/user/myapi",
    "change_id": "add-auth",
    "description": "Add JWT authentication"
  }
}

// 4. Get next task
{
  "tool": "registry_next",
  "params": {
    "path": "/home/user/myapi",
    "count": 1
  }
}

// 5. Update task status
{
  "tool": "registry_update",
  "params": {
    "path": "/home/user/myapi",
    "task_id": "add-auth/1",
    "updates": {
      "status": "completed"
    }
  }
}

// 6. Archive when done
{
  "tool": "spec_archive",
  "params": {
    "path": "/home/user/myapi",
    "change_id": "add-auth"
  }
}
```

---

## See Also

- [OpenSpec Workflow](./OPENSPEC_WORKFLOW.md) - How to use OpenSpec
- [CLI Reference](./CLI_REFERENCE.md) - CLI commands that use these tools
- [Commands Reference](./COMMANDS_REFERENCE.md) - Slash commands
- [Architecture](./ARCHITECTURE.md) - MCP server architecture

---

**Version:** v0.3.0
**Last updated:** 2026-01-28
