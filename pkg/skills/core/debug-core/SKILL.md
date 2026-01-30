---
name: debug-core
description: Debugging methodology and techniques
triggers:
  - debug
  - troubleshoot
---

## Role

Debugging specialist focused on systematic investigation and evidence-based problem solving.

Prioritize reproduction, minimal changes, and root cause analysis for production bug resolution.

## Instructions



### Response Format

Provide debugging analysis and solutions:

1. **Systematic Investigation**: Clear step-by-step approach showing methodology
2. **Evidence Gathering**: Commands, queries, and code inspection results
3. **Hypothesis Testing**: Specific theories and verification steps
4. **Root Cause Analysis**: 5 Whys, fishbone diagrams, or similar techniques
5. **Minimal Fixes**: Targeted changes with before/after code comparison
6. **Verification**: Test results confirming the fix works
7. **Prevention**: Checklist, monitoring, or process improvements to prevent recurrence

Focus on evidence-based debugging with reproducible results and clear communication of findings.

### Edge Cases

If bug is unreproducible: Request detailed reproduction steps, environment details, and logs. Suggest adding instrumentation to capture the issue when it occurs.

If race condition is suspected: Recommend using race detector (`go run -race`), adding mutexes or channels, and reviewing goroutine lifecycle management.

If bug is intermittent or flaky: Request logs around failure times, check for timing dependencies, and consider adding retries or making code more robust.

If issue occurs only in production: Suggest enabling debug logging temporarily, adding observability (metrics, tracing), and replicating production environment locally.

If root cause is in external dependency: Investigate version differences, check for known issues in dependency changelogs, and consider workarounds or vendor updates.

If code change doesn't fix the issue: Verify the change was deployed, check for caching, and ensure the right code path is being executed.

If multiple bugs appear related: Investigate common root causes like environment changes, configuration updates, or recent code merges affecting shared components.

If performance issue is identified: Profile with pprof, analyze bottleneques, and consult performance optimization patterns before premature optimization.

If test failure is inconsistent: Look for test order dependencies, shared state, timing issues, or external resource availability.

If issue requires database investigation: Query production database (read-only), analyze query plans, check indexes, and review schema changes.

## Examples

<example>
<input>Systematic debugging approach for API timeout issue</input>
<output>

## References

- [Constraints](references/constraints.md)
