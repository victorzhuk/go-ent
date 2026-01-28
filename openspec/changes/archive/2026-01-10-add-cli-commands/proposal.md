# Proposal: Add CLI Commands

**Status:** ARCHIVED
**Archived:** 2026-01-10

## Overview

Add standalone CLI commands (`go-ent run`, `go-ent status`, etc.) for non-MCP usage. Enables go-ent to work outside of Claude Code environment for automation and CI/CD.

## Rationale

### Problem
GoEnt only works via MCP (Claude Code) - can't use it standalone for scripts, CI/CD, or local development.

### Solution
Add CLI commands that wrap the execution engine and spec management:
```
go-ent run <task>              # Execute with agent selection
go-ent status                  # Show execution status
go-ent agent list/info         # Agent management
go-ent skill list/info         # Skill management
go-ent spec init/list/show     # Spec management
go-ent config show/set/init    # Config management
```

**Phase 1.1 Enhancement**: Add adaptive workflow commands:
```
go-ent quick <description>     # Fast path: Simple tasks (Haiku, <5 min)
go-ent go <description>        # Unified workflow router (analyzes complexity)
```

**Phase 5.1 Enhancement**: Unified command with automatic workflow selection:
```
go-ent go <description>
  → Analyzes task complexity
  → Routes to /go-ent:quick OR /go-ent:plan automatically
  → Provides estimates and recommendations
```

## Key Components

1. `internal/cli/root.go` - Root command and CLI framework
2. `internal/cli/run.go` - Execute tasks
3. `internal/cli/agent.go` - Agent commands
4. `internal/cli/spec.go` - Spec commands
5. `internal/cli/config.go` - Config commands
6. **`internal/cli/quick.go`** - **Fast path for simple tasks (Phase 1.1)**
7. **`internal/cli/go.go`** - **Unified workflow router (Phase 5.1)**
8. **`internal/workflow/router.go`** - **Complexity analysis and routing logic**

## Dependencies

- Requires: P0-P4 (execution engine)
- Can develop in parallel with P5 (agent-mcp-tools)

## Success Criteria

- [ ] `go-ent run <task>` executes with agent selection
- [ ] `go-ent spec list` works like MCP tool
- [ ] `go-ent config show` displays current config
- [ ] All commands have `--help` text
- [ ] `go-ent quick <task>` completes simple tasks in <5 minutes
- [ ] `go-ent go <task>` automatically routes to correct workflow
- [ ] Workflow router achieves >90% accuracy in complexity classification

## Phase 1.1 Enhancement: Quick Command for Simple Tasks

Provides a fast path for trivial tasks that don't require full OpenSpec workflow.

**Purpose**: Reduce overhead for simple tasks (linting, formatting, quick fixes) by using lightweight execution.

**Implementation**:
```go
// internal/cli/quick.go
func Quick(description string) error {
    // 1. Analyze task complexity
    complexity := analyzer.Analyze(description)
    if complexity > ComplexityThreshold {
        return fmt.Errorf("task too complex, use 'go-ent go' instead")
    }

    // 2. Select fast agent (Haiku)
    agent := "haiku"

    // 3. Execute without OpenSpec proposal
    result := executor.ExecuteDirect(agent, description)

    // 4. Display result
    return displayResult(result)
}
```

**Usage Examples**:
```bash
# Formatting
go-ent quick "format all Go files"

# Linting
go-ent quick "run golangci-lint and fix auto-fixable issues"

# Quick fix
go-ent quick "fix the typo in README.md line 42"

# Documentation
go-ent quick "add godoc comments to exported functions in repo.go"
```

**Characteristics**:
- **No proposal creation**: Executes directly without OpenSpec overhead
- **Fast model**: Uses Haiku for speed and cost savings
- **Time limit**: Target <5 minutes execution
- **No planning phase**: Skips research, decomposition, complexity analysis
- **Immediate feedback**: Streams output in real-time

**Complexity Threshold**:
```go
// internal/workflow/complexity.go
type ComplexityScore int

const (
    ComplexityTrivial   ComplexityScore = 1  // <5 min, Haiku
    ComplexitySimple    ComplexityScore = 2  // <15 min, Sonnet
    ComplexityModerate  ComplexityScore = 3  // <1 hour, plan mode
    ComplexityComplex   ComplexityScore = 4  // >1 hour, full workflow
)

func Analyze(description string) ComplexityScore {
    indicators := []Indicator{
        {Pattern: "format|lint|fix typo", Score: ComplexityTrivial},
        {Pattern: "multiple files|refactor|design", Score: ComplexityComplex},
        // ... more rules
    }
    // Score based on keywords, file count, impact
}
```

## Phase 5.1 Enhancement: Unified Workflow Router

Automatically routes tasks to appropriate workflow based on complexity analysis.

**Purpose**: Single entry point that intelligently decides between quick execution and full planning workflow.

**Implementation**:
```go
// internal/cli/go.go
func Go(description string) error {
    // 1. Analyze task complexity
    score := router.Analyze(description)

    // 2. Display analysis
    fmt.Printf("Task complexity: %s\n", score.Level())
    fmt.Printf("Estimated time: %s\n", score.EstimatedTime())
    fmt.Printf("Recommended workflow: %s\n", score.RecommendedWorkflow())

    // 3. Ask user confirmation (optional --auto flag)
    if !autoConfirm {
        confirmed := askConfirmation()
        if !confirmed {
            return nil
        }
    }

    // 4. Route to appropriate workflow
    switch score.Level() {
    case ComplexityTrivial:
        return executeQuick(description)
    case ComplexitySimple:
        return executeWithAgent(description, "sonnet")
    case ComplexityModerate, ComplexityComplex:
        return executePlanMode(description)
    }
}
```

**Routing Logic**:
```go
// internal/workflow/router.go
type Router struct {
    analyzer *ComplexityAnalyzer
    config   *Config
}

type RoutingDecision struct {
    Workflow         string          // "quick" | "agent" | "plan"
    Agent            string          // "haiku" | "sonnet" | "opus"
    EstimatedTime    time.Duration
    Confidence       float64         // 0.0-1.0
    Reasoning        []string        // Why this decision
}

func (r *Router) Analyze(description string) RoutingDecision {
    indicators := r.extractIndicators(description)

    // Complexity signals
    complexity := 0
    if indicators.FileCount > 5 { complexity += 2 }
    if indicators.HasDesignPhase { complexity += 3 }
    if indicators.HasBreakingChange { complexity += 2 }
    if indicators.HasUnknowns { complexity += 1 }

    // Choose workflow
    if complexity <= 2 {
        return RoutingDecision{
            Workflow: "quick",
            Agent: "haiku",
            EstimatedTime: 5 * time.Minute,
            Confidence: 0.9,
        }
    } else if complexity <= 4 {
        return RoutingDecision{
            Workflow: "agent",
            Agent: "sonnet",
            EstimatedTime: 15 * time.Minute,
            Confidence: 0.8,
        }
    } else {
        return RoutingDecision{
            Workflow: "plan",
            Agent: "opus",
            EstimatedTime: 60 * time.Minute,
            Confidence: 0.95,
        }
    }
}
```

**Usage Examples**:
```bash
# Automatic routing
go-ent go "fix linting errors"
# → Routes to quick (Haiku, <5 min)

go-ent go "add rate limiting to API handlers"
# → Routes to agent (Sonnet, <15 min)

go-ent go "redesign authentication system with OAuth2"
# → Routes to plan (Opus, full workflow)

# Force specific workflow
go-ent go --workflow=quick "add new feature"
go-ent go --workflow=plan "simple fix"  # Override if needed

# Auto-confirm routing decision
go-ent go --auto "refactor database layer"
```

**User Experience**:
```
$ go-ent go "add logging to all API endpoints"

Analyzing task...
┌────────────────────────────────────────┐
│ Task Complexity Analysis               │
├────────────────────────────────────────┤
│ Complexity:     Moderate               │
│ Files affected: ~8 (estimated)         │
│ Breaking change: No                    │
│ Design required: No                    │
│                                        │
│ Recommended:    Agent workflow (Sonnet)│
│ Estimated time: 12-18 minutes          │
│ Confidence:     87%                    │
│                                        │
│ Reasoning:                             │
│ • Multiple files to modify             │
│ • Repetitive pattern (good for agent)  │
│ • No complex design decisions          │
└────────────────────────────────────────┘

Proceed with agent workflow? [Y/n]:
```

**Configuration**:
```yaml
# .go-ent/workflow.yaml
workflow:
  # Complexity thresholds
  thresholds:
    quick_max_files: 3
    quick_max_minutes: 5
    agent_max_files: 10
    agent_max_minutes: 20

  # Default preferences
  defaults:
    auto_confirm: false
    prefer_quick: true  # Favor quick over agent when uncertain

  # Override rules
  overrides:
    - pattern: "security|auth|crypto"
      force_workflow: plan
      reason: "Security changes require full review"

    - pattern: "breaking"
      force_workflow: plan
      reason: "Breaking changes need careful planning"
```

## Impact

**Performance**:
- Quick workflow: 70-90% faster than full workflow for simple tasks
- Routing accuracy: >90% with proper configuration
- Cost savings: Haiku execution ~20x cheaper than Opus

**User Experience**:
- Single entry point: `go-ent go` for all tasks
- Automatic optimization: System chooses best workflow
- Transparency: Clear explanation of routing decision
- Flexibility: Can override automatic routing if needed

**Metrics** (Phase 5.2 integration):
- Track routing accuracy over time
- Measure actual vs estimated execution time
- Identify tasks that should have different routing
