# Prompt Design Principles

This document documents the principles and practices used for designing agent prompts in the go-ent project, following Constitutional AI principles and focused on creating thoughtful, judgment-capable agents.

## Prompt Engineering Philosophy

### Constitutional AI Applied to Agent Prompts

Our prompt design is guided by Constitutional AI principles adapted for autonomous development agents:

- **Judgment over rigidity**: Agents should think like senior developers who understand when to follow guidelines and when to adapt based on context
- **Context awareness**: Rules are starting points, not absolute laws. Real-world situations require thoughtful application
- **Principal hierarchy**: When guidance conflicts, agents follow a clear hierarchy to resolve conflicts
- **Responsibility with checkpoints**: Irreversible actions require explicit verification

### Thoughtful Senior Developer as Target

We design prompts for how senior developers actually work:

- Ask clarifying questions when requirements are ambiguous
- Propose simpler solutions when over-engineering is detected
- Adapt patterns to fit the actual problem, not apply patterns blindly
- Balance best practices with project conventions
- Know when to escalate uncertainty versus make a reasonable judgment call

### Principal Hierarchy for Conflict Resolution

When guidance conflicts, agents follow this hierarchy:

1. **Project conventions** (e.g., AGENTS.md, CODE.md, existing patterns)
2. **User intent** (what the user actually needs, not what they said)
3. **Best practices** (general Go/community guidelines)
4. **Agent specialization** (agent-specific constraints and goals)

### Irreversible Action Checkpoints

Before actions that cannot be easily undone:

- Code deletions: Verify with user or provide clear warning
- Force operations: Explain risks and get confirmation
- Breaking changes: Explicitly flag and verify
- Major refactors: Ensure backup/revert path exists

## Design Principles

### 1. Clear Purpose Statements

Each agent prompt must explicitly state:

- What the agent is responsible for
- What falls outside its scope
- How it should collaborate with other agents

**Example from architect agent:**
```
You are the system architect responsible for high-level design decisions.
You define structure, patterns, and interfaces but do not implement code.
```

**Why this matters:** Prevents scope creep, ensures agents stay focused, reduces conflicts.

### 2. Actionable Guidance

Provide specific behaviors, not abstract theory:

- **Bad:** "Write good code that follows best practices"
- **Good:** "Before writing code: (1) Check existing patterns in the relevant package, (2) Follow AGENTS.md conventions, (3) Run `make lint` and `make test`"

**Why this matters:** Gives agents concrete steps they can execute without ambiguity.

### 3. Good/Bad Examples

Show both correct and incorrect approaches:

```
## Good: Simple solution
func GetUser(id string) (*User, error) {
    return repo.Query(ctx, id)
}

## Bad: Over-engineered for no reason
type UserServiceFactory interface {
    CreateUserService(config Config) (UserService, error)
}
type UserServiceImpl struct { ... }
```

**Why this matters:** Demonstrates the target behavior and common pitfalls.

### 4. Checklists

Provide easy-to-follow verification steps:

```
## Before deploying code:
- [ ] make lint passes
- [ ] make test passes
- [ ] Zero comments except WHY
- [ ] Domain has zero external deps
```

**Why this matters:** Enables agents to systematically verify their work.

### 5. Shared vs. Agent-Specific

- **Shared guidance** (`shared/_*.md`): Used by multiple agents (e.g., judgment, principals)
- **Agent-specific** (`agents/{agent}.md`): Unique behaviors and responsibilities

**Why this matters:** Avoids duplication, ensures consistency, makes updates easier.

### 6. Natural Integration

New guidance must flow with existing instructions:

- Don't bolt on contradictory rules
- Weave new concepts into existing sections
- Cross-reference related guidance

**Why this matters:** Creates coherent, natural-sounding prompts that agents follow holistically.

### 7. No Contradictions

All guidance must be consistent:

- Verify new guidance doesn't conflict with existing rules
- Use principal hierarchy to resolve conflicts
- Document any deliberate trade-offs

**Why this matters:** Contradictions cause unpredictable agent behavior.

### 8. Testability

Guidance should produce verifiable behavior:

- Define what "good" looks like for each guideline
- Include success criteria
- Provide test scenarios

**Why this matters:** Enables validation that prompts work as intended.

## File Organization

```
plugins/go-ent/agents/prompts/
├── shared/              # Common guidance (included by multiple agents)
│   ├── _judgment.md     # When to exercise judgment
│   ├── _principals.md   # Principal hierarchy for conflicts
│   ├── _handoffs.md     # When/how to hand off between agents
│   ├── _conventions.md  # Go code conventions
│   └── _tooling.md      # Tool usage guidelines
└── agents/              # Agent-specific prompts
    ├── architect.md     # System design
    ├── coder.md         # Implementation
    ├── planner.md       # Task breakdown
    ├── tester.md        # Testing strategy
    └── ...
```

### Shared Files

- **`_judgment.md`**: When to follow rules vs. apply judgment
- **`_principals.md`**: Hierarchy for resolving conflicting guidance
- **`_handoffs.md`**: When/how to delegate or escalate
- **`_conventions.md`**: Go coding standards and patterns
- **`_tooling.md`**: MCP tool usage best practices

### Agent-Specific Files

Each `agents/{agent}.md` file:

1. Defines the agent's core responsibility
2. Includes relevant shared guidance
3. Adds agent-specific behaviors
4. Provides examples specific to that role

## Adding New Guidance

### Step-by-Step Process

1. **Check existing shared files** - Look for related guidance in `shared/`
2. **Determine scope**:
   - Applies to multiple agents → Add to or create shared file
   - Specific to one agent → Add to `agents/{agent}.md`
3. **Draft the guidance** using design principles above
4. **Cross-reference** - Link to related files where appropriate
5. **Update agent prompts** - Include new shared files in relevant agents
6. **Test with examples** - Verify behavior with agent-specific scenarios
7. **Check for contradictions** - Ensure no conflicts with existing guidance

### Decision Tree

```
New guidance needed?
├─ Does it apply to multiple agents?
│  ├─ Yes → shared/{topic}.md
│  └─ No → agents/{agent}.md
├─ Is it about when to follow rules?
│  └─ Yes → Add to _judgment.md
├─ Is it about conflict resolution?
│  └─ Yes → Add to _principals.md
└─ Is it agent behavior?
   └─ Yes → Add to agents/{agent}.md
```

## Best Practices

### Use Concise, Natural Language

Avoid verbose AI-style explanations:

- **Bad:** "In accordance with established best practices and community guidelines, the implementation should..."
- **Good:** "Follow AGENTS.md conventions. Before coding, check existing patterns in the package."

### Include Real-World Examples

Use actual Go code patterns from the project:

```
## Example: Repository pattern
See internal/domain/user/repo.go for the pattern to follow.
Private models with tags, public domain entities.
```

### Focus on Decision Frameworks

Instead of just rules, provide frameworks for making decisions:

```
## Should I create an interface?
- Multiple implementations needed now? Yes → interface
- Testing requires mock? Consider real implementation first
- "For future flexibility"? No, YAGNI
```

### Emphasize Context Awareness

Teach agents to read the room:

```
## Adapting to context
- Prototype: Faster decisions, more acceptable to skip tests
- Production: Follow all conventions, comprehensive testing
- Simple feature: Don't over-engineer
- Complex feature: More thorough design and testing
```

### Test by Simulating Agent Behavior

Before deploying, mentally simulate:

- "What would this agent do if asked X?"
- "Would it make good judgment calls?"
- "Would it know when to escalate?"

### Iterate Based on Real Usage

- Collect examples of good and poor decisions
- Update prompts based on patterns
- Document learnings back to this file

## Good vs. Bad Judgment Calls

### Architect Agent

**Good Judgment:**
- Suggests a simple struct over a complex pattern when requirements are straightforward
- Recognizes that a prototype doesn't need perfect abstraction layers
- Asks clarifying questions when requirements seem contradictory

**Bad Judgment:**
- Insists on implementing the Repository pattern for a one-off data access scenario
- Adds "future-proof" interfaces that no one will use
- Makes assumptions about scale without asking

### Coder (Dev) Agent

**Good Judgment:**
- Adds minimal tests covering happy path for prototype code
- Skips over-engineering when a simple function meets the need
- Runs `make fmt` and `make lint` before committing

**Bad Judgment:**
- Creates an interface "for testing" when concrete type would work
- Adds abstraction layers "in case we need them later"
- Writes tests for private implementation details instead of behavior

### Planner Agent

**Good Judgment:**
- Breaks down a feature into 5 coherent tasks instead of 20 micro-tasks
- Combines related steps that are better done together
- Identifies when to escalate for requirements clarification

**Bad Judgment:**
- Creates 50 micro-tasks for a simple feature
- Breaks down design and implementation into separate agents when they're tightly coupled
- Proceeds without clarification when requirements are clearly ambiguous

### Learning from These Examples

**Context matters:**
- Prototype vs. production
- Simple vs. complex
- Internal tool vs. customer-facing

**Consider user intent:**
- What do they actually need vs. what they said
- Timeline constraints
- Resource constraints

**Apply principals:**
- Project conventions first
- Then user intent
- Then best practices
- Finally, agent-specific goals

**When uncertain:**
- Ask rather than assume
- Escalate appropriately
- Document the question and decision

## Testing Checklist for Prompt Changes

### Before Deploying Prompt Changes

**Content Review:**
- [ ] Read the modified prompt aloud - does it flow naturally?
- [ ] Check for contradictions with existing instructions
- [ ] Verify all examples are agent-relevant and accurate
- [ ] Ensure new guidance integrates smoothly with existing text

**Testing Scenarios:**
- [ ] Test with typical task for this agent
- [ ] Test with edge case or unusual request
- [ ] Test with conflict between two guidelines
- [ ] Test with situation requiring judgment
- [ ] Test with situation requiring escalation

**Validation:**
- [ ] Document what behavior should change
- [ ] Get feedback from other developers if possible
- [ ] Create rollback plan if new prompts cause issues

### Testing Scenario Template

For each scenario, capture:

```
Scenario: [Brief description]
Input: [What you'd tell the agent]
Expected output: [What the agent should do]
Actual output: [What the agent did]
Pass/Fail: [ ]
Notes: [Observations, learnings]
```

### After Deployment

**Monitoring:**
- [ ] Monitor agent behavior for unexpected changes
- [ ] Collect examples of good decisions
- [ ] Collect examples of poor decisions
- [ ] Track situations where agent escalates unnecessarily

**Iteration:**
- [ ] Update prompts based on real usage patterns
- [ ] Add new examples from actual work
- [ ] Refine guidance based on feedback
- [ ] Document learnings back to this file

**Rollback Criteria:**

Consider rollback if:

- Agent behavior becomes unpredictable
- New guidance conflicts with established patterns
- Agent starts making poor judgment calls
- Escalation frequency increases significantly
- Other developers report confusion

## References

### Project Files

- `CLAUDE.md` - Core development principles and conventions
- `AGENTS.md` - Build commands, CLI usage, testing patterns
- `CODE.md` - Complete code examples and patterns
- `plugins/go-ent/agents/prompts/shared/` - Shared prompt guidance
- `plugins/go-ent/agents/prompts/agents/` - Agent-specific prompts

### External Resources

- **Constitutional AI**: Anthropic's research on alignment and governance
- **Prompt Engineering**: Best practices for LLM prompt design
- **Clean Architecture**: Robert C. Martin's layered architecture principles
- **Domain-Driven Design**: Bounded contexts and domain modeling

---

**Version:** 1.0
**Last Updated:** 2025-01-22
**Maintained By:** Development team
