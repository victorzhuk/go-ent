# Proposal: Add Context Memory System

## Why

Agents currently lose context between sessions and cannot learn from successful patterns. Every task starts from scratch, repeating discovery of codebase patterns, architectural decisions, and proven solutions.

Inspired by:
- [Claude-Flow](https://github.com/ruvnet/claude-flow) - Dual memory (AgentDB vectors + ReasoningBank SQLite)
- [Oh-My-OpenCode](https://github.com/code-yeongyu/oh-my-opencode) - Context-aware parallelization

## What Changes

- **Session Memory**: Persist important context across agent invocations within a session
- **Project Memory**: Store codebase patterns, conventions, and decisions in SQLite
- **Pattern Learning**: Capture successful task completions as reusable patterns
- **Semantic Search**: Find relevant past context using embedding-based similarity
- **Memory Compression**: Quantize/summarize old memories to manage storage

## Impact

- Affected specs: memory-system (new capability)
- Affected code: internal/memory/, cmd/mcp/
- Dependencies: Requires add-config-system (completed)

## Key Benefits

1. **Reduced Token Usage**: Avoid re-discovering known patterns
2. **Faster Onboarding**: New sessions inherit project knowledge
3. **Improved Consistency**: Apply proven patterns automatically
4. **Learning**: Agents get better at project-specific tasks over time
