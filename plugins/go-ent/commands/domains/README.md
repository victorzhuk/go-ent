# Domains

Domain knowledge files that provide context-specific guidance for flows.

## Available Domains

### openspec.md
OpenSpec-specific rules for change management and task tracking.

**When to use**: Flows that work with OpenSpec change proposals, tasks, and registry.

**Contains**:
- OpenSpec file structure (`openspec/changes/{id}/`)
- Change ID format and registry operations
- Task tracking in `tasks.md`
- Spec validation commands
- Proposal and design templates

**Used by flows**:
- `flows/plan.md` - Creating OpenSpec changes
- `flows/task.md` - Executing OpenSpec tasks

### generic.md
Generic development patterns that apply across projects.

**When to use**: Flows that don't depend on OpenSpec-specific conventions.

**Contains**:
- Standard project structure patterns
- Code standards and naming conventions
- Testing patterns (TDD, table-driven tests)
- Build and validation commands
- Code review processes
- Agent escalation patterns
- Common Go patterns (repository, service, factory)
- Security and observability patterns

**Used by flows**:
- `flows/bug.md` - Debugging workflow

## Adding New Domains

To add a new domain:

1. Create a new file in this directory: `domains/{domain-name}.md`
2. Add domain-specific patterns and rules
3. Update flow files to include: `{{include "domains/{domain-name}.md"}}`

## Domain Structure

Each domain file should follow this structure:

```markdown
# Domain: {Domain Name}

Brief description of what this domain covers.

## Category 1
...
## Category 2
...
```

Categories can include:
- File structure
- Commands and operations
- Validation rules
- Templates
- Patterns
- Best practices
