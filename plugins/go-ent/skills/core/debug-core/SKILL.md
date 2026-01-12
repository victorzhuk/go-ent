---
name: debug-core
description: "Debugging methodology and techniques. Auto-activates for: troubleshooting, investigating bugs, root cause analysis, reproduction steps."
version: 1.0.0
---

# Debugging Core

## Methodology

### Scientific Approach
1. **Observe** - Gather symptoms and errors
2. **Hypothesize** - Form theories about cause
3. **Test** - Design experiments
4. **Analyze** - Interpret results
5. **Repeat** - Refine hypothesis

### Divide and Conquer
- Binary search through code
- Isolate problematic component
- Reduce input to minimal reproduction

## Reproduction

**Minimal repro process**:
1. Start with failing case
2. Remove unrelated code
3. Use simplest input
4. Document exact steps
5. Verify consistency

**Information needed**:
- Exact error message + stack trace
- Input data + environment
- Steps to reproduce
- Expected vs actual behavior

## Techniques

| Technique | When to Use | Trade-offs |
|-----------|-------------|------------|
| Print debugging | Quick checks, production issues | Can clutter code |
| Debugger (delve, gdb) | Complex flow, variable inspection | Setup overhead |
| Rubber duck | Logic errors, design issues | No tooling needed |
| Binary search | Large codebase, unclear location | Time-consuming |
| Structured logging | Production, distributed systems | Performance impact |

## Common Bug Patterns

| Category | Symptoms | Typical Causes |
|----------|----------|----------------|
| Logic | Wrong result, off-by-one | Incorrect assumptions, missing edge cases |
| Concurrency | Race, deadlock, data race | Unprotected shared state, improper locking |
| Resources | Leaks (memory, files, connections) | Missing cleanup, unclosed resources |
| Integration | Timeouts, version mismatch | API changes, config differences |

## Root Cause Analysis

### 5 Whys
```
Problem: API is slow
Why? → Database queries slow
Why? → No index on queried column
Why? → Index dropped in migration
Why? → Migration auto-generated
Why? → Developer didn't review SQL
Root: Insufficient code review
```

### Fishbone Categories
- **People**: training, experience, communication
- **Process**: workflow, procedures, review
- **Tools**: software, infrastructure, dependencies
- **Environment**: load, configuration, network

## Prevention

**Defensive programming**:
- Validate inputs at boundaries
- Assert preconditions
- Handle errors explicitly
- Check return values

**Testing strategy**:
- Unit tests for logic
- Integration tests for interactions
- Property-based testing for edge cases
- Fuzzing for unexpected inputs

**Observability**:
- Structured logging with correlation IDs
- Application metrics and tracing
- Error tracking (Sentry, Rollbar)
- Alerting on anomalies

## When Stuck

1. Take a break
2. Pair debug with someone
3. Search error messages
4. Check issue trackers
5. Use `git bisect` to find regression
6. Add instrumentation
7. Simplify problem further
