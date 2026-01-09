# Proposal: Add Execution Engine

## Overview

Implement the execution engine that runs agents on different runtimes: Claude Code (MCP), OpenCode (native), and CLI (standalone). Supports Single, Multi-agent, and Parallel execution strategies.

## Rationale

### Problem
No runtime abstraction - can't execute tasks via OpenCode or CLI, only MCP.

### Solution
- **Runner interface**: Abstract runtime execution (Claude Code, OpenCode, CLI)
- **Execution strategies**: Single, Multi (conversation), Parallel (with dependency graph)
- **Budget tracking**: Monitor and enforce spending limits
- **Result aggregation**: Collect outputs from parallel executions

## Key Components

1. `internal/execution/engine.go` - Main orchestration engine
2. `internal/execution/claude.go` - Claude Code MCP runner
3. `internal/execution/opencode.go` - **OpenCode native runner (required for v3.0)**
4. `internal/execution/cli.go` - CLI standalone runner
5. `internal/execution/strategy.go` - Execution strategy implementations
6. `internal/execution/codemode.go` - **Code-mode tool: JavaScript sandbox for dynamic tool composition**
7. `internal/execution/sandbox.go` - **Security sandbox for untrusted code execution**
8. `internal/tool/composer.go` - **Tool composition registry for storing composed tools**

## Dependencies

- Requires: P0-P3 (all foundation)
- Blocks: P5 (agent-mcp-tools), P6 (cli-commands)

## Success Criteria

- [ ] Claude Code runner executes via MCP
- [ ] OpenCode runner executes via native API
- [ ] CLI runner executes standalone
- [ ] Parallel execution with dependency graph works
- [ ] Budget tracking prevents overspending
- [ ] Code-mode tool enables JavaScript-based tool composition
- [ ] Sandbox isolates untrusted code execution
- [ ] Composed tools persist and can be reused across sessions

## Phase 2 Enhancement: Code-Mode and Tool Composition

### Code-Mode Tool

Implements a JavaScript sandbox for dynamic tool composition at runtime, inspired by Docker's code-mode pattern.

**Purpose**: Allow agents to programmatically create and modify tools on-the-fly without editing Go code.

**Implementation**:
```go
// internal/execution/codemode.go
type CodeMode struct {
    sandbox  *Sandbox
    composer *tool.Composer
}

// Execute JavaScript code in isolated sandbox
func (c *CodeMode) Execute(code string, context map[string]any) (any, error)

// Register composed tool for future use
func (c *CodeMode) RegisterTool(name string, code string) error
```

**Example Use Case**:
```javascript
// Agent composes a new tool dynamically
const autoFix = async (file, lintErrors) => {
  const content = await readFile(file);
  for (const error of lintErrors) {
    // Apply automated fix logic
    content = applyFix(content, error);
  }
  await writeFile(file, content);
  return { fixed: lintErrors.length };
};
```

**MCP Tool**: `code_mode_execute(code: string, context: object) -> result`

### Security Sandbox

Isolates untrusted code execution to prevent system compromise.

**Features**:
- Resource limits (CPU, memory, time)
- Filesystem access restrictions (read-only project files)
- Network access control (optional allow-list)
- API whitelisting (only allowed MCP tools)

**Implementation**:
```go
// internal/execution/sandbox.go
type Sandbox struct {
    limits   ResourceLimits
    allowFS  []string  // Allowed file paths
    allowAPI []string  // Allowed API calls
}

type ResourceLimits struct {
    MaxMemoryMB  int
    MaxCPUTime   time.Duration
    MaxExecTime  time.Duration
}
```

### Tool Composition Registry

Stores composed tools for cross-session reuse.

**Schema**:
```go
// internal/tool/composer.go
type ComposedTool struct {
    ID          string
    Name        string
    Description string
    Code        string    // JavaScript implementation
    Scope       string    // "project" | "global"
    Created     time.Time
    UsageCount  int
    LastUsed    time.Time
}

func (c *Composer) Save(tool *ComposedTool) error
func (c *Composer) Load(name string) (*ComposedTool, error)
func (c *Composer) List() []ComposedTool
```

**Storage**: `.go-ent/composed-tools/{name}.json`

**Benefits**:
- 80%+ reuse of successful compositions
- Persistent skill development across sessions
- Shareable tool definitions within teams

## Impact

**Performance**:
- Code-mode adds ~5ms overhead per execution
- Sandbox prevents runaway processes
- Composed tools reduce token usage by 70-90%

**Security**:
- Isolated execution environment
- No access to system commands without explicit permission
- Audit trail for all composed tool executions
