
## OpenSpec Workflow

Use OpenSpec CLI for spec-driven development. Artifacts structure work: **proposal.md** (requirements) → **tasks.md** (checklist) → **specs/** (delta specs) → **implementation**.

### Quick Commands

```bash
# Create change and all artifacts
openspec new change feature-name
openspec ff feature-name

# Show status
openspec list
openspec show feature-name

# Validate and archive
openspec validate feature-name
openspec archive feature-name
```

### Workflow Patterns

**Fast-Forward (recommended):**
```bash
openspec new change feature-name
openspec ff feature-name  # Creates all artifacts
# Implement tasks
openspec archive feature-name
```

**Incremental:**
```bash
openspec new change feature-name
openspec continue feature-name  # Creates next artifact
# Repeat continue for each artifact
openspec archive feature-name
```

## Agent Delegation

### When to Delegate
- **Planning → Execution**: After creating task breakdown
- **Execution → Testing**: After implementing features
- **Testing → Review**: After running test suites
- **Review → Coordination**: After validation complete

### Agent Selection Guide

**Before delegating, match the task type to the right agent:**

| Task Type | Agent | Model Tier |
|-----------|-------|------------|
| Write new code / modify code | @ent/coder | main |
| Design architecture / API boundaries | @ent/architect | heavy |
| Break feature into tasks | @ent/planner | main (auto-routes by complexity) |
| Quick triage of simple task | @ent/planner-fast | fast |
| Deep architectural planning | @ent/planner-heavy | heavy |
| Fix a bug / investigate failure | @ent/debugger | main (auto-routes by complexity) |
| Fix obvious single-line bug | @ent/debugger-fast | fast |
| Debug concurrency / perf issue | @ent/debugger-heavy | heavy |
| Write or fix tests specifically | @ent/tester | main |
| Review completed code | @ent/reviewer | heavy |
| Read-only code investigation | @ent/researcher | heavy |
| Break large task into subtasks | @ent/decomposer | main |
| Verify requirements are met | @ent/acceptor | main |

**Key routing rules:**
- Do NOT use @ent/coder for debugging — use @ent/debugger
- Do NOT use @ent/coder for planning — use @ent/planner
- Do NOT use @ent/coder for test-only tasks — use @ent/tester
- Use fast-tier agents for simple tasks to save cost and latency
- Use heavy-tier agents for architecture and review to get better reasoning

### Safety Checkpoints
Before irreversible operations:
1. Verify backups exist
2. Confirm user intent for destructive ops
3. Run validation before deployment
4. Test rollback procedures
5. Document decision and reasoning

See full guidance in `ent-workflow` skill
