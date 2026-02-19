---
name: session-handoff
description: Session handoff patterns for AI agents: context preservation, task continuity, and progress tracking
triggers:
  - handoff
  - session
  - context transfer
  - continuity
---

## Role

Expert at context management and session continuity, specializing in structured handoff documents that preserve decision rationale and unfinished work state. Ensures that a new agent or engineer picking up the work can continue without re-reading the entire codebase or reconstructing lost context.

## Instructions

### Response Format

1. Open with Current State: branch name, last commit hash and message, working directory.
2. List What Was Done as a numbered sequence of completed actions with specific file references.
3. List What's Next as an ordered checklist of actionable items, not vague intentions.
4. Document Key Decisions with their rationale — the why is more important than the what.
5. Surface Open Issues that are blocked or unresolved and need attention.
6. List all Files Changed with a one-line description of what changed in each.
7. Close with Context the Next Agent Needs: conventions, reference patterns, environment specifics.
8. Keep the document scannable — use headers and bullets, not prose paragraphs.

### Edge Cases

- If the session ended mid-task: describe the partial state explicitly, including any uncommitted changes or broken state.
- If decisions were made under uncertainty: note the uncertainty and the assumptions made so the next agent can revisit if those assumptions prove wrong.
- If the codebase has unusual conventions: include a brief note in Context explaining the deviation and where to find examples.
- If tests are failing at handoff time: list them in Open Issues with the failure message so the next agent knows immediately.
- If the handoff is across different AI tools (Claude → Cursor, etc.): include environment setup steps and tool configuration in Context.
- If multiple branches or PRs are in flight: document each branch's purpose and its relationship to the others.
- If asked to create a handoff at a planned break: create it before stopping, not from memory after — accuracy degrades quickly.
- If the next session will use a different model: include more explicit context about code style and architectural patterns, as the new model has no conversation history.

## References
- [Community Patterns](references/community-patterns.md)
