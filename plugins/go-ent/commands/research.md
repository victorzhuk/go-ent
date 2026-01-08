---
description: Structured research phase for unknowns and technology decisions
argument-hint: <change-id> [topic]
---

# Research Phase

Investigate unknowns and technical questions before implementation.

## Input

- Change ID: $ARGUMENTS (required)
- Optional: Specific topics to research

## Path Resolution

Change directory: `openspec/changes/$ARGUMENTS/`

For the steps below, `$CHANGE_ROOT` refers to `openspec/changes/$ARGUMENTS/`.

## When to Use

Run research when proposal contains:
- "NEEDS CLARIFICATION" markers
- "TBD" items in design decisions
- Multiple technology options to evaluate
- Unclear integration patterns
- Unknown performance/security implications

## Steps

1. Validate change exists: `openspec show $ARGUMENTS`
2. Resolve change directory path (see Path Resolution above)
3. Scan all change artifacts for unknowns:
   - Items marked "NEEDS CLARIFICATION"
   - "TBD" in design.md
   - "TODO: research" comments
   - Questions in proposal without answers
3. Categorize research areas:
   - Technology evaluation
   - Integration patterns
   - External dependencies
   - Performance/security analysis
4. For each unknown, structure research:
   - Question/goal
   - Options evaluated
   - Decision criteria
   - Recommendation with rationale
5. Track completion status
6. Write `research.md` to change directory

## Research Template

```markdown
# Research: <change-id>

Last Updated: YYYY-MM-DD

## Status
- [ ] Technology evaluation complete
- [ ] Integration patterns identified
- [ ] Unknowns resolved
- [ ] External dependencies documented
- [ ] Performance/security reviewed

## Unknowns from Proposal

### 1. <Unknown Name>
**Question**: What technology/approach should we use for X?

**Options Evaluated**:
| Option | Pros | Cons | Decision |
|--------|------|------|----------|
| Library A | Active, 10k stars | Complex API | - |
| Library B | Simple | Unmaintained | - |
| Custom | Full control | More code | SELECTED |

**Research Findings**:
- Library A requires X which conflicts with Y
- Custom implementation ~200 LOC
- Performance similar across options

**Recommendation**: Use custom implementation because:
1. Simple use case doesn't justify dependency
2. Full control over behavior
3. No version conflicts

**References**:
- [Library A docs](https://example.com)
- [Blog: Custom vs Library](https://example.com)

---

### 2. <Another Unknown>
**Question**: How to integrate with external API?

**Options**: REST vs GraphQL vs gRPC

**Research Needed**:
- [ ] API documentation review
- [ ] Latency benchmarks
- [ ] SDK quality assessment
- [ ] Authentication mechanism

**Status**: IN PROGRESS

---

## Technology Decisions

| Decision | Outcome | Rationale |
|----------|---------|-----------|
| OTP Library | pquerna/otp | Active, well-tested, TOTP/HOTP support |
| SMS Provider | TBD | Blocked: needs cost/reliability comparison |
| Database | PostgreSQL (existing) | No change needed |

## Integration Patterns Discovered

### Pattern 1: External API with Circuit Breaker
```
Repository → Adapter → External API
               ↓
        Circuit Breaker
               ↓
          Fallback Cache
```

**Source**: Found in `internal/repository/payment/` (similar pattern)

## Open Questions

Questions still needing human input:

1. **Budget constraints for SMS provider?**
   - Research complete, awaiting business decision
   - Options: Twilio ($X), AWS SNS ($Y), MessageBird ($Z)

2. **Rate limiting strategy?**
   - Technical options identified
   - Need product decision on limits
```

## Research Guidelines

1. **Be thorough**: Evaluate 2-3 options minimum
2. **Document sources**: Link to docs, blogs, examples
3. **Quantify when possible**: Performance numbers, costs, maintenance burden
4. **Check existing patterns**: Search codebase first (`rg <keyword>`)
5. **Track status**: Mark items complete only when decided
6. **Escalate blockers**: Identify what needs human decision

## Output

Creates `$CHANGE_ROOT/research.md`

Output file: `openspec/changes/$ARGUMENTS/research.md`

## Integration

Use before:
- `/go-ent:decompose` - Resolve unknowns before task breakdown
- `/openspec:apply` - Complete research before implementation

Blocks implementation if critical unknowns remain unresolved.
