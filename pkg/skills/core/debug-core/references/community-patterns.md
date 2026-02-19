## Process
1. **Reproduce**: Create a reliable, minimal reproduction
2. **Isolate**: Narrow down to the smallest failing unit
3. **Diagnose**: Understand WHY it fails, not just WHERE
4. **Fix**: Apply the minimal correct fix
5. **Verify**: Confirm the fix works AND doesn't break other things
6. **Prevent**: Add tests and guards to prevent regression

## Reproduction
- Get the exact error message, stack trace, and logs
- Determine the minimal steps to reproduce
- Identify what changed: recent commits, config changes, dependency updates
- Check environment differences: local vs CI vs production

## Isolation Techniques
- Binary search: disable half the code, which half fails?
- Comment out sections to find the breaking change
- Use git bisect for regression hunting
- Create a minimal reproduction in an isolated environment
- Check: is it the code, data, config, or environment?

## Diagnosis Tools
- **Debuggers**: Step through code, inspect state (delve, pdb, node --inspect)
- **Logging**: Add targeted debug logging around the suspicious area
- **Tracing**: Follow request flow through distributed systems
- **Profiling**: CPU, memory, goroutine dumps for performance issues
- **Network**: curl, tcpdump, browser DevTools Network tab

## Common Bug Categories
- Off-by-one errors: boundary conditions, loop bounds
- Race conditions: concurrent access without proper synchronization
- Null/nil references: missing nil checks, optional chaining
- State mutations: unexpected side effects, shared mutable state
- Configuration: wrong environment, missing env vars, stale cache
- Dependency issues: version conflicts, breaking changes

## Prevention
- Write a test that catches the bug BEFORE fixing it
- Add assertions for invariants that were violated
- Improve error messages to make future debugging easier
- Document the root cause and fix in the commit message
- Consider if similar bugs exist elsewhere in the codebase
