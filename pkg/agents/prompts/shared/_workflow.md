
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

### Safety Checkpoints
Before irreversible operations:
1. Verify backups exist
2. Confirm user intent for destructive ops
3. Run validation before deployment
4. Test rollback procedures
5. Document decision and reasoning

See full guidance in `ent-workflow` skill
