---
name: ent-workflow
description: OpenSpec workflow and agent delegation patterns. Spec-driven development with artifact creation, task tracking, handoffs, and validation.
triggers:
  - ent-workflow
  - openspec
  - change proposal
  - spec workflow
  - artifact workflow
  - handoffs
  - delegation
---

## Role

Guide agents through OpenSpec CLI workflows and proper agent delegation patterns.

## OpenSpec CLI Workflow

OpenSpec uses **artifacts** to structure work:
- **proposal.md** - What and why (requirements, scope, approach)
- **tasks.md** - Implementation checklist
- **specs/** - Delta specs (changes to main specs)
- **implementation** - The actual code

Always use the `openspec` CLI via Bash, not MCP tools.

### Creating Changes

```bash
# Initialize OpenSpec in a project (if needed)
openspec init

# Create a new change
openspec new change add-feature-name

# Fast-forward: create all artifacts in one go
# This creates proposal.md, tasks.md, and any needed specs
openspec ff add-feature-name

# OR: Incremental artifact creation
openspec continue add-feature-name  # Creates next artifact
```

### Working with Artifacts

```bash
# Show change status and artifacts
openspec show add-feature-name

# Show specific artifact
openspec show add-feature-name --artifact proposal
openspec show add-feature-name --artifact tasks

# Get instructions for creating next artifact
openspec instructions add-feature-name

# Validate change structure
openspec validate add-feature-name
```

### Viewing Changes

```bash
# List all active changes
openspec list

# List specs
openspec list --specs

# Show change with all artifacts
openspec show add-feature-name --format markdown

# Check artifact completion status
openspec status add-feature-name
```

### Implementation

```bash
# Apply/implement tasks (if you have an apply command)
# Note: Task implementation is typically done manually,
# guided by tasks.md

# Sync delta specs to main specs (after implementation)
openspec sync add-feature-name
```

### Archiving

```bash
# Verify before archive (checks all artifacts exist and are valid)
openspec validate add-feature-name

# Archive completed change
# This moves the change to archive/ and can update main specs
openspec archive add-feature-name
```

### Schema Management

```bash
# List available workflow schemas
openspec schemas

# Show template paths for a schema
openspec templates --schema default
```

## Workflow Patterns

### Pattern 1: Quick Feature (Fast-Forward)

```bash
# 1. Create change and all artifacts at once
openspec new change add-login-validation
openspec ff add-login-validation

# 2. Review generated artifacts
openspec show add-login-validation

# 3. Implement tasks from tasks.md

# 4. Archive when done
openspec archive add-login-validation
```

### Pattern 2: Incremental (Step-by-Step)

```bash
# 1. Create change directory
openspec new change refactor-auth

# 2. Create proposal first
openspec continue refactor-auth
# [Edit proposal.md]

# 3. Create tasks
openspec continue refactor-auth
# [Review tasks.md]

# 4. Create specs if needed
openspec continue refactor-auth

# 5. Implement

# 6. Archive
openspec archive refactor-auth
```

### Pattern 3: Spec Updates

```bash
# 1. Create change for spec updates
openspec new change update-api-spec

# 2. Fast-forward to create artifacts
openspec ff update-api-spec

# 3. Edit delta specs in openspec/changes/update-api-spec/specs/

# 4. Sync to main specs
openspec sync update-api-spec

# 5. Archive
openspec archive update-api-spec
```

## Agent Delegation Patterns

Guidance for when to delegate between agents, escalation patterns, and safety checkpoints before irreversible operations.

### When to Delegate

**By Agent Role:**
- **Planning agents** → Create tasks, then delegate to execution agents
- **Execution agents** → Implement code, delegate testing to tester
- **Validation agents** → Review implementation, report to coordination

### Safety Checkpoints

**Before Irreversible Operations:**
1. Verify backups exist
2. Confirm user intent for destructive ops
3. Run validation before deployment
4. Test rollback procedures
5. Document decision and reasoning

### Delegation Responsibilities

**Coordinating Agent (Planner/Architect):**
- Creates comprehensive task breakdown
- Assigns tasks to appropriate agents
- Monitors progress and resolves blockers
- Ensures consistency across agents

**Execution Agent (Coder):**
- Implements assigned tasks
- Follows project conventions
- Delegates testing to tester
- Reports completion/blockers

**Validation Agent (Reviewer/Tester):**
- Validates implementation correctness
- Runs test suites
- Identifies issues and reports back
- Approves or requests changes

## Common Pitfalls

### OpenSpec
1. **Don't duplicate OpenSpec functionality** - Always use `openspec` CLI, never reimplement its features
2. **Don't skip validation** - Always run `openspec validate` before archiving
3. **Don't manually move archives** - Use `openspec archive` to ensure proper cleanup
4. **Don't forget to sync specs** - Use `openspec sync` to update main specs from deltas

### Agent Delegation
1. **Don't work in isolation** - Coordinate with other agents via shared task list
2. **Don't skip handoffs** - Explicitly delegate to next agent with clear context
3. **Don't assume state** - Always verify current state before acting
4. **Don't ignore blockers** - Escalate promptly when stuck

## Examples

### OpenSpec Workflow

<example>
<input>Create a new change for adding user authentication</input>
<output>
I'll create a new OpenSpec change for user authentication:

```bash
openspec new change add-user-authentication
```

Now I'll fast-forward to create all artifacts:

```bash
openspec ff add-user-authentication
```

This creates:
- `openspec/changes/add-user-authentication/proposal.md`
- `openspec/changes/add-user-authentication/tasks.md`
- Any needed specs in `openspec/changes/add-user-authentication/specs/`

Let me show the current status:

```bash
openspec show add-user-authentication
```
</output>
</example>

<example>
<input>What's the status of my current changes?</input>
<output>
Let me list all active changes:

```bash
openspec list
```

To see artifact completion status for a specific change:

```bash
openspec status change-name
```

To view a specific change with all artifacts:

```bash
openspec show change-name --format markdown
```
</output>
</example>

<example>
<input>How do I archive a completed change?</input>
<output>
First, validate the change has all required artifacts:

```bash
openspec validate change-name
```

Then archive it:

```bash
openspec archive change-name
```

This will:
1. Verify all artifacts are complete
2. Move the change to `openspec/changes/archive/`
3. Optionally sync delta specs to main specs
</output>
</example>

### Agent Delegation

<example>
<input>Planner has broken down implementation into 5 tasks</input>
<output>
As planner, I've created the task breakdown. Now I'll delegate to coder:

Handoff to `/ent:coder`:
- Context: Implement user authentication feature
- Tasks: See `openspec/changes/add-auth/tasks.md`
- Dependencies: None (first implementation phase)
- Expected outcome: All tasks completed, tests passing

After implementation, delegate to `/ent:tester` for validation.
</output>
</example>

## References

- OpenSpec CLI: `openspec --help`
- Change commands: `openspec change --help`
- Artifact workflow: `openspec instructions --help`
- Schemas: `openspec schemas`
