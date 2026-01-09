# Architecture Enhancement Plan: MCP + Code Execution + Spec-Driven Development

**Version**: 1.0
**Author**: Architecture Team
**Date**: January 9, 2026
**Status**: Planning

---

## Executive Summary

**Current State**: go-ent v2.0 is a sophisticated MCP server with spec-driven development, multi-agent orchestration, and advanced planning capabilities.

**Research Insights**: The articles reveal:
1. **Code execution with MCP** reduces token consumption 98.7% by writing code instead of direct tool calls
2. **Spec-driven development** has three maturity levels (spec-first → spec-anchored → spec-as-source)
3. **Dynamic MCPs** enable autonomous tool discovery and composition

**Key Finding**: go-ent already has many of these features planned in v3.0 proposals. The plan below focuses on **enhancing and prioritizing** those proposals based on research insights, not creating new ones.

---

## Analysis of Current Architecture

### Strengths
- ✅ **OpenSpec Integration**: Built-in spec-driven development workflow
- ✅ **Agent System**: Multi-agent orchestration with specialized roles (planner, dev, reviewer, etc.)
- ✅ **Planning Workflow**: Comprehensive `/go-ent:plan-full` with research, design, task decomposition
- ✅ **Registry System**: Centralized task management with dependency tracking
- ✅ **V3.0 Roadmap**: Clear path to execution engine, ACP proxy, plugin system
- ✅ **Code-First Philosophy**: Clean architecture, SOLID principles, simplicity over complexity

### Gaps Identified from Research

| Gap | Current State | Research Insight | Impact |
|------|--------------|------------------|--------|
| **Tool Discovery** | Manual MCP config | Dynamic `mcp-find`, `mcp-add` tools | High token waste loading all tools upfront |
| **Code Execution Mode** | Direct tool calls | Agents write code to call tools | 98.7% token reduction possible |
| **Tool Composition** | Sequential tool calls | Compose new tools from existing ones | Reusability, state management |
| **Spec Anchoring** | Spec-first only | Spec-as-source with spec-anchored | Long-term maintenance, evolution |
| **Adaptive Workflows** | Fixed workflows | Workflow size-adaptive | Overkill for simple tasks |
| **Progressive Disclosure** | Load all definitions | On-demand tool loading | Context efficiency |

---

## Enhancement Plan

### Priority Matrix

| Priority | Enhancement | Impact | Effort | Dependencies |
|----------|-------------|--------|---------|--------------|
| **P0** | Dynamic MCP Tool Discovery | High | Medium | Add ACP proxy mode |
| **P1** | Code Execution Mode (Code-Mode) | Very High | High | Execution engine |
| **P2** | Workflow Size Adaptation | High | Low | Planning commands |
| **P3** | Spec-As-Source Mode | Medium | High | Plugin system + rules engine |
| **P4** | Tool Composition Framework | High | Medium | Code-mode + sandboxing |
| **P5** | State Persistence for Agents | Medium | Medium | Execution engine |

---

## Phase 1: Quick Wins (Week 1-2)

### 1.1 Adaptive Workflow Size Adaptation
**Problem**: Current `/go-ent:plan-full` is overkill for simple tasks (bug fixes, small features)

**Solution**: Add workflow size detection and adaptive execution

**Changes to**: `openspec/changes/add-cli-commands/`

```yaml
# new command: /go-ent:quick <task>
description: "Execute simple tasks (bug fixes, small changes) with minimal workflow"

triggers:
  - "fix bug"
  - "add simple feature"
  - "small change"

workflow:
  1. Analyze task complexity
  2. If complexity == "trivial":
     - Skip research phase
     - Skip design document
     - Direct implementation → test → review
  3. If complexity == "simple":
     - Minimal research (5 min)
     - No design document
     - Tasks only
  4. If complexity == "moderate+":
     - Use full /go-ent:plan-full workflow
```

**Files to modify**:
- `plugins/go-ent/commands/quick.md` (new)
- `internal/agent/complexity.go` (enhance)

**Success Criteria**:
- Bug fixes take <5 min from request to implementation
- Simple features take <15 min
- Trivial tasks don't create verbose markdown files

---

### 1.2 Progressive Disclosure for MCP Tools
**Problem**: All MCP tools loaded upfront, consuming tokens

**Solution**: Implement tool search and lazy loading

**Changes to**: `internal/mcp/server/`

```go
type ToolRegistry struct {
    allTools    map[string]*mcp.Tool
    activeTools map[string]*mcp.Tool

    lazyLoader func(pattern string) ([]*mcp.Tool, error)
}

func (r *ToolRegistry) FindTools(pattern string) ([]*mcp.Tool, error) {
    results := make([]*mcp.Tool, 0)
    for _, tool := range r.allTools {
        if matches(pattern, tool.Name, tool.Description) {
            results = append(results, tool)
        }
    }
    return results, nil
}

func (r *ToolRegistry) LazyLoad(tools []string) error {
    for _, name := range tools {
        if _, loaded := r.activeTools[name]; !loaded {
            tool := r.allTools[name]
            r.activeTools[name] = tool
        }
    }
    return nil
}
```

**New MCP Tools**:
- `tool_find(pattern string) []Tool`
- `tool_describe(name string) ToolDefinition`

**MCP Server Changes**:
```go
func (s *Server) handleToolFind(args map[string]any) (*mcp.CallToolResult, error) {
    pattern := args["pattern"].(string)
    tools := s.toolRegistry.FindTools(pattern)

    return mcp.NewToolResultJSON(tools), nil
}
```

**Success Criteria**:
- Tool definitions consume 10% of previous tokens
- Agent can discover tools via `tool_find` rather than listing all
- Tools only loaded when explicitly requested

---

## Phase 2: Code Execution Mode (Week 3-5)

### 2.1 Code-Mode Tool for Tool Composition
**Problem**: Direct tool calls limit composability and state management

**Solution**: Add `code_mode` tool that generates and executes code

**Based on**: P4 (execution engine) + Docker's code-mode approach

**New MCP Tool**:
```yaml
name: code_mode
description: |
  Execute JavaScript code that can call other MCP tools.
  Use for complex workflows, data transformation, state management.

parameters:
  - name: code
    type: string
    description: |
      JavaScript code. Available functions:
      - mcp_call(tool_name, params): Call any MCP tool
      - mcp_list(): List all tools
      - console.log(): Output for review (intermediate results)

  - name: servers
    type: array
    description: MCP servers available to code (default: all)

returns:
  - output: Console.log output
  - tokens_used: Token usage
  - execution_time: Duration
```

**Implementation**: `internal/execution/codemode.go`

```go
type CodeModeRunner struct {
    toolRegistry *ToolRegistry
    sandbox      *Sandbox
}

func (r *CodeModeRunner) Execute(ctx context.Context, req CodeModeRequest) (*CodeModeResult, error) {
    // 1. Validate code for security
    if err := r.validate(req.Code); err != nil {
        return nil, fmt.Errorf("validation: %w", err)
    }

    // 2. Create sandboxed execution environment
    env := r.sandbox.Create()
    env.Set("mcp_call", func(tool string, params map[string]any) (any, error) {
        return r.toolRegistry.Call(ctx, tool, params)
    })
    env.Set("console.log", func(args ...any) {
        // Capture logs, don't send to LLM
        r.captureLogs(args...)
    })

    // 3. Execute JavaScript in sandbox
    result, err := env.Run(req.Code)

    // 4. Return only final output
    return &CodeModeResult{
        Output:        result,
        Logs:          r.getLogs(),
        TokensUsed:    r.calculateTokens(req.Code, result),
        ExecutionTime: time.Since(start),
    }, err
}
```

**Security Sandbox** (inspired by Docker's approach):
```go
type Sandbox struct {
    timeout       time.Duration
    memoryLimit   int64
    diskReadPath  string // restricted to workspace/
    diskWritePath string // restricted to workspace/
    network       bool    // false by default
}

func (s *Sandbox) validate(code string) error {
    // 1. No file system access outside workspace
    // 2. No network calls
    // 3. No subprocess execution
    // 4. No eval/exec
    // 5. Max execution time
    // 6. Max memory
    return nil
}
```

**Example Usage**:

```
USER: "Add rate limiting to all 15 API endpoints"

AGENT (using code_mode):
```javascript
const endpoints = await mcp_call('api_list_endpoints');

const implementations = [];
for (const ep of endpoints) {
  const impl = await mcp_call('code_generate', {
    template: 'rate-limited-endpoint',
    context: { endpoint: ep }
  });
  implementations.push(impl);
  console.log(`Generated rate limiter for ${ep}`);
}

console.log(`Generated ${implementations.length} rate limiters`);
return implementations;
```

RESULT: Only console.log output shown to LLM, not intermediate tool calls
```

**Success Criteria**:
- Token usage reduced 90%+ for multi-tool workflows
- Complex workflows (5+ tool calls) use code_mode automatically
- Security sandbox prevents unsafe operations
- State persists across tool calls within code_mode

**Integration**: Connects to P4 (execution engine) and P7 (ACP proxy mode)

---

### 2.2 Tool Composition Registry
**Problem**: Composable tools can't be reused across sessions

**Solution**: Persist composed tools from code_mode sessions

**New Capability**: `tool-composition` spec in `openspec/specs/`

```yaml
openspec/specs/tool-composition/spec.md

## ADDED Requirements

### Requirement: Tool Composition
The system SHALL allow composing new tools from existing MCP tools.

#### Scenario: Save composed tool
- **WHEN** agent executes code_mode successfully
- **AND** agent requests save
- **THEN** tool is saved to workspace/.goent/composed/
- **AND** tool is available in future sessions

#### Scenario: Load composed tool
- **WHEN** agent lists available tools
- **THEN** composed tools are included
- **AND** can be called directly like built-in tools
```

**Implementation**: `internal/tool/composer.go`

```go
type ComposedTool struct {
    ID          string
    Name        string
    Description string
    Code        string // JavaScript code from code_mode
    Source      string // Which session created this
    CreatedAt   time.Time
    UsageCount  int
}

type Composer struct {
    repo        ToolRepository
    validator   *SandboxValidator
}

func (c *Composer) Save(tool ComposedTool) error {
    if err := c.validator.Validate(tool.Code); err != nil {
        return fmt.Errorf("invalid code: %w", err)
    }
    return c.repo.Save(tool)
}

func (c *Composer) List() ([]*ComposedTool, error) {
    return c.repo.ListAll()
}
```

**Workspace Structure**:
```
project/
  .goent/
    composed/
      ├── fetch-and-analyze-sheet.tool.json
      ├── bulk-rate-limiter.tool.json
      └── migration-validator.tool.json
    workspace/
      ├── leads.csv  # persisted state
      └── cache.json
```

**Success Criteria**:
- Agents can save and reuse composed tools
- 80%+ of complex workflows use composed tools after first creation
- Tool repository accessible across sessions

---

## Phase 3: Dynamic MCP Integration (Week 6-7)

### 3.1 MCP Smart Search
**Problem**: Manual MCP configuration, no autonomous tool discovery

**Solution**: Integrate with Docker MCP Gateway for smart search

**Based on**: P7 (ACP proxy mode) + Docker's `mcp-find`, `mcp-add`

**New MCP Tools**:
```yaml
name: mcp_find
description: Search available MCP servers by name or description
parameters:
  - name: query
    type: string
    required: true
  - name: category
    type: string
    enum: [database, version-control, cloud, api, monitoring]

name: mcp_add
description: Add MCP server to current session
parameters:
  - name: server_id
    type: string
    required: true
  - name: config
    type: object
    description: Server-specific configuration
```

**Integration with Docker MCP Gateway**:
```go
// internal/mcp/gateway.go
type GatewayClient struct {
    endpoint string
    catalog *MCPCatalog
}

func (g *GatewayClient) Find(ctx context.Context, query string) ([]*MCPServer, error) {
    // Call Docker MCP Gateway /api/servers endpoint
    resp, err := g.client.R().Get("/api/servers", func(r *gorequest.SuperAgent) {
        r.Param("q", query)
    })

    var servers []*MCPServer
    resp.JSON(&servers)
    return servers, err
}

func (g *GatewayClient) Add(ctx context.Context, serverID string, config map[string]any) error {
    // Add server to active session via gateway
    _, err := g.client.R().Post("/api/session/servers", map[string]any{
        "id":     serverID,
        "config": config,
    })
    return err
}
```

**Configuration**:
```yaml
# .goent/mcp.yaml
mcp_gateway:
  url: http://localhost:8080

catalogs:
  - name: docker
    url: https://hub.docker.com/mcp
  - name: community
    url: https://github.com/modelcontextprotocol/servers
```

**Success Criteria**:
- Agents can discover and add MCP servers autonomously
- No manual configuration for 90% of common MCP servers
- Docker MCP Gateway integration working

---

### 3.2 Dynamic Tool Selection
**Problem**: Too many tools loaded, context bloat

**Solution**: Only load tools relevant to current task

**Changes to**: `internal/mcp/server/`

```go
type ToolSelector struct {
    analyzer    *TaskAnalyzer
    registry    *ToolRegistry
    cache       map[string][]string // task_type -> tool_names
}

func (s *ToolSelector) SelectForTask(task Task) ([]*mcp.Tool, error) {
    // 1. Analyze task type
    taskType := s.analyzer.Classify(task)

    // 2. Check cache
    if tools, ok := s.cache[taskType]; ok {
        return s.loadTools(tools)
    }

    // 3. Use embedding similarity if not cached
    tools := s.findSimilarTools(task.Description)

    // 4. Cache result
    s.cache[taskType] = toolNames(tools)

    return tools, nil
}
```

**Success Criteria**:
- Average 10 tools loaded per session (down from 50+)
- Task classification accuracy >90%
- Cache hit rate >80%

---

## Phase 4: Spec-Driven Development Enhancements (Week 8-10)

### 4.1 Spec Anchoring Mode
**Problem**: Specs deleted after task completion, can't evolve with project

**Solution**: Keep specs active and link to code files

**Based on**: Research article's spec-anchored + spec-as-source levels

**Changes to**: `openspec/` workflow

```yaml
# .goent/spec-anchoring.yaml
mode: anchored  # options: first, anchored, source

anchoring_rules:
  - match: { complexity: "trivial" }
    mode: first  # Create spec for task, delete after completion

  - match: { complexity: "simple" }
    mode: anchored  # Keep spec, update on changes

  - match: { complexity: "moderate+" }
    mode: source  # Spec is source of truth, code generated
```

**Spec File Format** (enhanced):
```markdown
# Spec: API Rate Limiting

## Metadata
status: active
version: 2
linked_files:
  - internal/api/rate_limiter.go
  - internal/api/middleware.go
last_reviewed: 2026-01-09
evolution_history:
  - version: 1
    date: 2026-01-05
    changes: Initial implementation
  - version: 2
    date: 2026-01-09
    changes: Added Redis support
```

**New Command**: `/go-ent:evolve <spec-id>`

```markdown
Updates spec based on code changes or new requirements:

1. Analyze current code linked to spec
2. Detect drift (code != spec)
3. Propose spec updates
4. Update spec with approval
5. Generate code changes if mode=source
```

**Success Criteria**:
- Specs persist beyond task completion for moderate+ complexity
- Spec drift detected and flagged automatically
- Evolution history tracked in spec files

---

### 4.2 Spec As-Source Mode (Optional)
**Problem**: Code divergence from spec over time

**Solution**: Code is generated from spec, never manually edited (for spec-as-source mode)

**Implementation**: Requires plugin system (P6) + rules engine

```yaml
# spec-as-source/config.yaml
enabled: true

files_with_source_marker:
  pattern: "// GENERATED FROM SPEC - DO NOT EDIT"
  enforce: true

generation_rules:
  - spec: openspec/specs/auth/spec.md
    files:
      - internal/domain/user.go
      - internal/repository/user_repo.go
      - internal/usecase/auth.go

  - spec: openspec/specs/api/spec.md
    files:
      - internal/transport/http/handler.go
```

**Workflow**:
1. User edits spec.md
2. `/go-ent:generate-from-spec auth`
3. Code regenerated
4. Validation runs
5. Tests verify

**Success Criteria**:
- Code generation from spec works reliably
- Tests pass after regeneration
- 95%+ code coverage from spec scenarios

---

## Phase 5: Integration & Polish (Week 11-12)

### 5.1 Unified Workflow Command
**Problem**: Multiple commands (`/plan`, `/quick`, `/execute`, `/loop`) confusing

**Solution**: Single smart command that adapts to task complexity

```yaml
# plugins/go-ent/commands/go.md
name: /go-ent:go <task>
description: Universal command - adapts workflow to task complexity

workflow:
  1. Analyze task:
     - Size (trivial, simple, moderate, complex)
     - Type (bug, feature, refactor, review)
     - Risk level (low, medium, high)

  2. Select workflow:
     - trivial/bug → quick workflow (direct fix)
     - simple/feature → plan workflow (research + tasks)
     - moderate/feature → full workflow (research + design + tasks)
     - complex/architectural → plan-full (multi-agent planning)

  3. Execute selected workflow
```

**Implementation**: `internal/workflow/router.go`

```go
type Router struct {
    complexityAnalyzer *ComplexityAnalyzer
    quickWorkflow     *QuickWorkflow
    planWorkflow       *PlanWorkflow
    fullWorkflow       *FullWorkflow
}

func (r *Router) Route(task Task) (Workflow, error) {
    analysis := r.complexityAnalyzer.Analyze(task)

    switch {
    case analysis.Size == ComplexityTrivial && analysis.Type == TypeBug:
        return r.quickWorkflow, nil

    case analysis.Size == ComplexitySimple && analysis.Type == TypeFeature:
        return r.planWorkflow, nil

    case analysis.Size >= ComplexityModerate:
        return r.fullWorkflow, nil

    default:
        return r.planWorkflow, nil
    }
}
```

**Success Criteria**:
- Single `/go-ent:go` command works for 95%+ of requests
- Workflow selection accuracy >90%
- User satisfaction with automatic routing

---

### 5.2 Performance Monitoring
**Problem**: Can't measure token savings from code_mode and dynamic tool loading

**Solution**: Add metrics and dashboard

**New MCP Tool**: `metrics_show`

```yaml
name: metrics_show
description: Show execution metrics and cost analysis

returns:
  - sessions: Session statistics
  - token_usage: By tool, by model, by workflow
  - cost_breakdown: Per-session and total
  - optimization_score: (expected - actual) / expected
```

**Metrics Storage**: `internal/metrics/store.go`

```go
type Metric struct {
    SessionID    string
    Workflow     string
    ToolsUsed    []string
    TokensIn     int
    TokensOut    int
    CostUSD      float64
    Duration     time.Duration
    UsedCodeMode bool
    Timestamp    time.Time
}

type Store struct {
    db *sql.DB
}

func (s *Store) Record(m Metric) error {
    // Store metrics for analysis
}

func (s *Store) Analyze(period time.Duration) (*Analysis, error) {
    // Calculate optimization scores, trends
}
```

**Success Criteria**:
- Token savings measured and reported
- Cost reduction targets tracked
- Performance regression detected

---

## Dependency Graph

```
┌─────────────────────────────────────────────────────────────────┐
│                      PHASE 1: Quick Wins                      │
├─────────────────────────────────────────────────────────────────┤
│ 1.1 Adaptive Workflow Size      1.2 Progressive Disclosure │
│ └─ Changes: add-cli-commands    └─ Changes: mcp/server    │
│ └─ New: /go-ent:quick          └─ New: tool_find, tool_load│
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    PHASE 2: Code Execution                    │
├─────────────────────────────────────────────────────────────────┤
│ 2.1 Code-Mode Tool         2.2 Tool Composition           │
│ └─ Requires: P4 (exec)     └─ Requires: 2.1             │
│ └─ New: code_mode           └─ New: composer tool        │
│ └─ New: sandbox.go          └─ Spec: tool-composition     │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                  PHASE 3: Dynamic MCP                         │
├─────────────────────────────────────────────────────────────────┤
│ 3.1 MCP Smart Search         3.2 Dynamic Tool Selection    │
│ └─ Requires: P7 (ACP)       └─ Requires: 1.2             │
│ └─ New: mcp_find, mcp_add   └─ New: ToolSelector         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│              PHASE 4: Spec-Driven Enhancements               │
├─────────────────────────────────────────────────────────────────┤
│ 4.1 Spec Anchoring          4.2 Spec As-Source            │
│ └─ Requires: existing spec  └─ Requires: P6 (plugin)      │
│ └─ New: /go-ent:evolve     └─ New: generation rules      │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                  PHASE 5: Integration & Polish               │
├─────────────────────────────────────────────────────────────────┤
│ 5.1 Unified Workflow        5.2 Performance Monitoring       │
│ └─ Requires: Phases 1-4    └─ Requires: All phases       │
│ └─ New: /go-ent:go         └─ New: metrics_show          │
└─────────────────────────────────────────────────────────────────┘
```

---

## Success Metrics

### Phase 1 (Week 2)
- [ ] Simple tasks (<10 LOC) complete in <5 min
- [ ] Tool definitions in initial context <50% of previous
- [ ] Agent can discover tools via `tool_find`

### Phase 2 (Week 5)
- [ ] Token usage reduced 90% for workflows with 5+ tool calls
- [ ] Code-mode security sandbox passes all tests
- [ ] Composed tools saved and reused in 80%+ of sessions

### Phase 3 (Week 7)
- [ ] 90%+ of common MCP servers discoverable without config
- [ ] Average 10 tools loaded per session (down from 50+)
- [ ] Docker MCP Gateway integration working end-to-end

### Phase 4 (Week 10)
- [ ] Spec drift detection accuracy >90%
- [ ] Evolution history tracked for all moderate+ specs
- [ ] Spec-as-source code generation reliability >95%

### Phase 5 (Week 12)
- [ ] Single `/go-ent:go` command handles 95%+ of requests
- [ ] Workflow routing accuracy >90%
- [ ] Token savings measured: >80% reduction baseline
- [ ] Cost reduction >60% for typical workflows

---

## Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|---------|------------|
| Code-mode security vulnerabilities | Medium | High | Sandbox with network/process/disk restrictions, code validation, rate limiting |
| Spec-as-source reliability | High | Medium | Start with spec-anchored, spec-as-source opt-in only |
| Tool composition complexity | Medium | Medium | Simple composition first, limit to single-file compositions |
| Docker MCP Gateway dependency | Low | Medium | Fallback to local MCP catalog if gateway unavailable |
| Workflow routing errors | Medium | Low | A/B testing, manual override option, feedback loop |

---

## Recommendations

1. **Phase 1 First**: Implement adaptive workflow and progressive disclosure for quick wins
2. **Prioritize Code-Mode**: Highest ROI from research insights (98.7% token reduction)
3. **Spec Anchoring > Spec-As-Source**: More practical for real-world use
4. **Integrate with P7 (ACP Proxy)**: Dynamic MCP needs ACP integration anyway
5. **Measure Everything**: Add metrics from Day 1 to track improvements

---

## Open Questions

1. **Sandbox Technology**: Should we use gVisor, gVisor-lite, or custom sandbox for code-mode?
2. **Spec Persistence**: Where to store anchored specs? In-repo `.goent/specs/` or external DB?
3. **Docker Gateway**: Should we implement our own gateway or depend on Docker's official gateway?
4. **Tool Composition Limits**: Max composition depth? Max composed tools per session?
5. **Workflow Routing Feedback**: How to collect user feedback on routing decisions?

---

## References

1. [Anthropic: Code Execution with MCP](https://www.anthropic.com/engineering/code-execution-with-mcp)
2. [Martin Fowler: Understanding Spec-Driven Development](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)
3. [Docker: Dynamic MCPs](https://www.docker.com/blog/dynamic-mcps-stop-hardcoding-your-agents-world/)
4. [OpenSpec AGENTS.md](../openspec/AGENTS.md)
5. [Migration Plan v3.0](./MIGRATION_PLAN_GOENT_V3.md)
