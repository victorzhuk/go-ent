# GoEnt Multi-Agent Spec Manager

## Product Requirements Document (PRD)

**Version**: 3.0  
**Author**: Architecture Team  
**Date**: January 2026  
**Status**: Draft  

---

## 1. Executive Summary

### 1.1 Vision

GoEnt evolves from a simple spec-driven development tool into an **intelligent multi-agent orchestrator** that dynamically selects and coordinates AI agents based on task type, complexity, and context. The system acts as a **Spec Manager** that bridges OpenSpec-style change management with multi-agent execution across different AI coding platforms.

### 1.2 Problem Statement

Current AI-assisted development faces several challenges:

1. **Single-model limitations**: Tasks like "proposal creation" and "code implementation" require different cognitive capabilities
2. **Cost inefficiency**: Using expensive models (Opus 4.5) for simple tasks wastes resources
3. **Manual orchestration**: Developers manually switch between tools (Claude Code, OpenCode) based on task type
4. **Context fragmentation**: Specs, tasks, and implementation details live in disconnected systems
5. **No role specialization**: Same prompts used regardless of whether task needs architect, developer, or reviewer perspective

### 1.3 Solution Overview

GoEnt becomes a **spec action router** that:

- Intercepts spec actions (research, proposal, plan, execute, debug, review)
- Selects optimal agent configuration (model + role + runtime)
- Orchestrates single or parallel agent execution
- Maintains unified context across all operations
- Provides MCP interface for Claude Code and CLI for standalone usage

---

## 2. Target Users

### 2.1 Primary Users

| User Type | Description | Key Needs |
|-----------|-------------|-----------|
| **Senior Go Developer** | Individual using Claude Code for daily work | Automated model selection, cost optimization |
| **Tech Lead** | Manages team workflow and standards | Parallel execution, task distribution |
| **Engineering Manager** | Oversees project delivery | Progress tracking, complexity analysis |

### 2.2 Integration Points

| Platform | Role | Model Support |
|----------|------|---------------|
| **Claude Code** | Primary IDE integration | Opus 4.5, Sonnet 4.5, Haiku 4.5 |
| **OpenCode** | Cost-effective execution | Z.AI GLM 4.7, Kimi K2, Qwen3 |
| **CLI** | Standalone automation | All supported via API |

---

## 3. Core Concepts

### 3.1 Spec Actions

Each user intent maps to a **spec action** that determines agent selection:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        SPEC ACTION TAXONOMY                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  DISCOVERY PHASE                                                    │
│  ├── research     → Explore codebase, understand context            │
│  ├── analyze      → Evaluate complexity, identify risks             │
│  └── retrofit     → Generate specs for existing code                │
│                                                                     │
│  PLANNING PHASE                                                     │
│  ├── proposal     → Create change proposal with rationale           │
│  ├── plan         → Break down into actionable tasks                │
│  ├── design       → Technical architecture decisions                │
│  └── split        → Decompose large changes into phases             │
│                                                                     │
│  EXECUTION PHASE                                                    │
│  ├── implement    → Write code per tasks.md                         │
│  ├── execute      → Run parallel task implementation                │
│  └── scaffold     → Generate boilerplate code                       │
│                                                                     │
│  VALIDATION PHASE                                                   │
│  ├── review       → Code review against standards                   │
│  ├── verify       → Run tests, check requirements                   │
│  ├── debug        → Analyze failures, suggest fixes                 │
│  └── lint         → Style and quality checks                        │
│                                                                     │
│  LIFECYCLE PHASE                                                    │
│  ├── approve      → Request/grant change approval                   │
│  ├── archive      → Complete and archive change                     │
│  └── status       → Report progress and blockers                    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Agent Roles

Agents specialize by role, affecting their prompts and decision-making:

| Role | Focus | Used For |
|------|-------|----------|
| **Product** | User needs, requirements | proposal, research (user-facing) |
| **Architect** | System design, trade-offs | proposal, plan, design, split |
| **Senior** | Implementation patterns | plan, implement, debug |
| **Developer** | Code writing, testing | implement, execute, scaffold |
| **Reviewer** | Quality, standards | review, verify, lint |
| **Ops** | Deployment, monitoring | verify (integration), debug (production) |

### 3.3 Runtime Environments

| Runtime | Best For | Cost | Speed |
|---------|----------|------|-------|
| **Claude Code + Opus 4.5** | Complex reasoning, architecture | $$$ | Medium |
| **Claude Code + Sonnet 4.5** | Balanced tasks | $$ | Fast |
| **OpenCode + GLM 4.7** | Bulk implementation | $ | Fast |
| **OpenCode + Kimi K2** | Long-context tasks | $ | Medium |

---

## 4. Agent Selection Matrix

### 4.1 Default Mappings

```yaml
# Action → Agent Configuration
action_matrix:
  # Discovery
  research:
    primary:
      runtime: claude-code
      model: sonnet-4.5
      role: architect
    fallback:
      runtime: opencode
      model: glm-4.7
      role: senior

  # Planning
  proposal:
    primary:
      runtime: claude-code
      model: opus-4.5
      roles: [product, architect]  # Multi-role
    approval_required: true

  plan:
    primary:
      runtime: claude-code
      model: opus-4.5
      roles: [architect, senior]
    output: tasks.md

  design:
    primary:
      runtime: claude-code
      model: opus-4.5
      role: architect
    output: design.md

  # Execution
  implement:
    primary:
      runtime: claude-code
      model: sonnet-4.5
      role: senior
    fallback:
      runtime: opencode
      model: glm-4.7
      role: developer

  execute:
    # Parallel execution for multiple tasks
    strategy: parallel
    max_workers: 3
    primary:
      runtime: opencode
      model: glm-4.7
      role: developer

  # Validation
  review:
    primary:
      runtime: claude-code
      model: opus-4.5
      role: reviewer
    checklist:
      - enterprise_standards
      - solid_principles
      - error_handling

  debug:
    primary:
      runtime: claude-code
      model: sonnet-4.5
      role: senior
    context:
      - error_logs
      - stack_traces
      - recent_changes
```

### 4.2 Selection Algorithm

```
┌─────────────────────────────────────────────────────────────────────┐
│                    AGENT SELECTION FLOW                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  INPUT: spec_action, change_context, user_preferences               │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ 1. COMPLEXITY ANALYSIS                                       │   │
│  │    - LOC changed estimate                                    │   │
│  │    - Files affected count                                    │   │
│  │    - Cross-domain dependencies                               │   │
│  │    - Risk level (breaking changes, security, etc.)           │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ 2. CAPABILITY MATCHING                                       │   │
│  │    - Action requirements → Model capabilities                │   │
│  │    - Context size → Model context window                     │   │
│  │    - Role requirements → Agent specialization                │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ 3. COST OPTIMIZATION                                         │   │
│  │    - Budget constraints                                      │   │
│  │    - Token estimation                                        │   │
│  │    - Fallback to cheaper options if low complexity           │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ 4. EXECUTION STRATEGY                                        │   │
│  │    - Single agent vs. multi-agent                            │   │
│  │    - Sequential vs. parallel                                 │   │
│  │    - Approval gates                                          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  OUTPUT: AgentConfig[]                                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 5. API Specification

### 5.1 MCP Server Tools

The MCP server exposes these tools for Claude Code integration:

#### 5.1.1 Spec Management Tools

```typescript
// Initialize spec folder in project
interface SpecInit {
  tool: "spec_init";
  params: {
    path: string;           // Project root path
    template?: string;      // Optional template (go-clean, microservice, etc.)
  };
  returns: {
    created: string[];      // Created files/folders
    project_yaml: string;   // Path to project.yaml
  };
}

// List specs, changes, or tasks
interface SpecList {
  tool: "spec_list";
  params: {
    type: "specs" | "changes" | "tasks" | "all";
    status?: "draft" | "active" | "pending_approval" | "archived";
    change_id?: string;     // Filter by change
  };
  returns: {
    items: SpecItem[];
    summary: { total: number; by_status: Record<string, number> };
  };
}

// Show detailed content
interface SpecShow {
  tool: "spec_show";
  params: {
    type: "spec" | "change" | "task";
    id: string;             // Full or partial ID
    include?: ("proposal" | "tasks" | "design" | "specs")[];
  };
  returns: {
    content: Record<string, string>;
    metadata: SpecMetadata;
    related_files: string[];
  };
}

// Create new spec/change/task
interface SpecCreate {
  tool: "spec_create";
  params: {
    type: "spec" | "change" | "task";
    id: string;
    content: string;
    parent_change?: string; // For tasks
  };
  returns: {
    path: string;
    id: string;
    validation: ValidationResult;
  };
}

// Update existing item
interface SpecUpdate {
  tool: "spec_update";
  params: {
    type: "spec" | "change" | "task";
    id: string;
    content?: string;       // Full replacement
    patch?: SpecPatch;      // Partial update
  };
  returns: {
    path: string;
    changes: string[];
  };
}

// Validate spec/change
interface SpecValidate {
  tool: "spec_validate";
  params: {
    id: string;
    strict?: boolean;       // Fail on warnings
  };
  returns: {
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationWarning[];
  };
}

// Archive completed change
interface SpecArchive {
  tool: "spec_archive";
  params: {
    change_id: string;
    reason?: string;
  };
  returns: {
    archived_path: string;
    specs_updated: string[];
  };
}
```

#### 5.1.2 Agent Orchestration Tools

```typescript
// Execute spec action with agent selection
interface AgentExecute {
  tool: "agent_execute";
  params: {
    action: SpecAction;
    change_id?: string;
    task_ids?: string[];
    options?: ExecutionOptions;
  };
  returns: {
    execution_id: string;
    agents: AgentConfig[];
    status: "queued" | "running" | "completed" | "failed";
    results?: ExecutionResult[];
  };
}

type SpecAction =
  | "research" | "analyze" | "retrofit"
  | "proposal" | "plan" | "design" | "split"
  | "implement" | "execute" | "scaffold"
  | "review" | "verify" | "debug" | "lint"
  | "approve" | "archive" | "status";

interface ExecutionOptions {
  model_override?: string;      // Force specific model
  role_override?: string;       // Force specific role
  runtime_override?: string;    // Force specific runtime
  parallel?: boolean;           // Enable parallel execution
  max_workers?: number;         // Parallel worker limit
  dry_run?: boolean;            // Preview without execution
  budget_limit?: number;        // Cost limit in USD
}

// Get execution status
interface AgentStatus {
  tool: "agent_status";
  params: {
    execution_id: string;
  };
  returns: {
    status: ExecutionStatus;
    progress: { completed: number; total: number };
    current_agent?: AgentInfo;
    results: StepResult[];
    cost_estimate: CostEstimate;
  };
}

// Cancel running execution
interface AgentCancel {
  tool: "agent_cancel";
  params: {
    execution_id: string;
    reason?: string;
  };
  returns: {
    cancelled: boolean;
    partial_results?: ExecutionResult[];
  };
}

// Get agent configuration for action
interface AgentConfig {
  tool: "agent_config";
  params: {
    action: SpecAction;
    change_id?: string;
    preview?: boolean;          // Return config without executing
  };
  returns: {
    primary: AgentSpec;
    fallback?: AgentSpec;
    strategy: "single" | "multi" | "parallel";
    estimated_cost: CostEstimate;
    estimated_time: string;
  };
}

interface AgentSpec {
  runtime: "claude-code" | "opencode" | "cli";
  model: string;
  role: AgentRole;
  skills: string[];
  context_files: string[];
}
```

#### 5.1.3 Skill Management Tools

```typescript
// List available skills
interface SkillList {
  tool: "skill_list";
  params: {
    category?: "review" | "patterns" | "testing" | "architecture";
    role?: AgentRole;
  };
  returns: {
    skills: SkillInfo[];
  };
}

interface SkillInfo {
  id: string;
  name: string;
  description: string;
  triggers: string[];
  category: string;
  applicable_roles: AgentRole[];
}

// Get skill details
interface SkillShow {
  tool: "skill_show";
  params: {
    skill_id: string;
  };
  returns: {
    content: string;
    examples: SkillExample[];
    dependencies: string[];
  };
}
```

### 5.2 CLI Commands

```bash
# Spec Management
goent spec init [--template <template>]
goent spec list [--type <type>] [--status <status>]
goent spec show <id> [--include <components>]
goent spec create <type> <id> [--from <file>]
goent spec validate <id> [--strict]
goent spec archive <change-id>

# Agent Execution
goent run <action> [change-id] [options]
goent run proposal --change add-auth
goent run plan --change add-auth --model opus-4.5
goent run execute --change add-auth --parallel --max-workers 3
goent run review --change add-auth --checklist enterprise

# Execution Management
goent status <execution-id>
goent cancel <execution-id>
goent logs <execution-id> [--follow]

# Configuration
goent config show
goent config set <key> <value>
goent config agent-matrix show
goent config agent-matrix edit
```

### 5.3 Slash Commands for Claude Code

```bash
# Discovery
/goent:research <topic>           # Explore codebase
/goent:analyze <change-id>        # Complexity analysis
/goent:retrofit <path>            # Generate specs for existing code

# Planning
/goent:proposal <description>     # Create change proposal
/goent:plan <change-id>           # Generate tasks.md
/goent:design <change-id>         # Create design.md
/goent:split <change-id>          # Split large changes

# Execution
/goent:implement <change-id>      # Implement single task
/goent:execute <change-id>        # Parallel execution
/goent:scaffold <type> <name>     # Generate boilerplate

# Validation
/goent:review <change-id>         # Code review
/goent:verify <change-id>         # Run verification
/goent:debug <error-context>      # Debug assistance
/goent:lint [path]                # Style check

# Lifecycle
/goent:approve <change-id>        # Request approval
/goent:archive <change-id>        # Archive change
/goent:status [change-id]         # Progress report
```

---

## 6. Execution Workflows

### 6.1 Proposal Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                      PROPOSAL WORKFLOW                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  USER: /goent:proposal "Add two-factor authentication"              │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 1: Context Gathering                                    │   │
│  │   Agent: Sonnet 4.5 / Architect                              │   │
│  │   Actions:                                                   │   │
│  │   - Read existing auth specs                                 │   │
│  │   - Analyze codebase structure                               │   │
│  │   - Identify affected components                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 2: Proposal Generation                                  │   │
│  │   Agent: Opus 4.5 / Product + Architect                      │   │
│  │   Outputs:                                                   │   │
│  │   - openspec/changes/add-2fa/proposal.md                     │   │
│  │   - openspec/changes/add-2fa/specs/auth/spec.md (delta)      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 3: Validation                                           │   │
│  │   Agent: Sonnet 4.5 / Reviewer                               │   │
│  │   Checks:                                                    │   │
│  │   - Spec format compliance                                   │   │
│  │   - No conflicts with existing specs                         │   │
│  │   - All requirements have scenarios                          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  OUTPUT: Change folder created, ready for /goent:plan               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.2 Parallel Execution Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                   PARALLEL EXECUTION WORKFLOW                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  USER: /goent:execute add-2fa --parallel --max-workers 3            │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 1: Task Analysis                                        │   │
│  │   Read tasks.md, identify independent tasks                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 2: Dependency Graph                                     │   │
│  │                                                              │   │
│  │   Task 1.1 (DB Schema) ──┐                                   │   │
│  │   Task 1.2 (OTP Model)  ─┼──► Task 2.1 (OTP Service)         │   │
│  │                          │                                   │   │
│  │   Task 3.1 (UI Component) ──► Task 3.2 (Integration)         │   │
│  │                                                              │   │
│  │   Independent groups: [1.1, 1.2, 3.1], [2.1], [3.2]          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 3: Parallel Dispatch                                    │   │
│  │                                                              │   │
│  │   Worker 1 (OpenCode/GLM 4.7) ──► Task 1.1                   │   │
│  │   Worker 2 (OpenCode/GLM 4.7) ──► Task 1.2                   │   │
│  │   Worker 3 (OpenCode/GLM 4.7) ──► Task 3.1                   │   │
│  │                                                              │   │
│  │   [Wait for completion]                                      │   │
│  │                                                              │   │
│  │   Worker 1 (OpenCode/GLM 4.7) ──► Task 2.1                   │   │
│  │                                                              │   │
│  │   [Wait for completion]                                      │   │
│  │                                                              │   │
│  │   Worker 1 (OpenCode/GLM 4.7) ──► Task 3.2                   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STEP 4: Merge & Verify                                       │   │
│  │   Agent: Sonnet 4.5 / Senior                                 │   │
│  │   - Merge parallel outputs                                   │   │
│  │   - Resolve conflicts                                        │   │
│  │   - Run integration tests                                    │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│                          ▼                                          │
│  OUTPUT: All tasks completed, tasks.md updated                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 7. Data Models

### 7.1 Core Types

```go
package domain

type SpecAction string

const (
    ActionResearch   SpecAction = "research"
    ActionAnalyze    SpecAction = "analyze"
    ActionRetrofit   SpecAction = "retrofit"
    ActionProposal   SpecAction = "proposal"
    ActionPlan       SpecAction = "plan"
    ActionDesign     SpecAction = "design"
    ActionSplit      SpecAction = "split"
    ActionImplement  SpecAction = "implement"
    ActionExecute    SpecAction = "execute"
    ActionScaffold   SpecAction = "scaffold"
    ActionReview     SpecAction = "review"
    ActionVerify     SpecAction = "verify"
    ActionDebug      SpecAction = "debug"
    ActionLint       SpecAction = "lint"
    ActionApprove    SpecAction = "approve"
    ActionArchive    SpecAction = "archive"
    ActionStatus     SpecAction = "status"
)

type AgentRole string

const (
    RoleProduct   AgentRole = "product"
    RoleArchitect AgentRole = "architect"
    RoleSenior    AgentRole = "senior"
    RoleDeveloper AgentRole = "developer"
    RoleReviewer  AgentRole = "reviewer"
    RoleOps       AgentRole = "ops"
)

type Runtime string

const (
    RuntimeClaudeCode Runtime = "claude-code"
    RuntimeOpenCode   Runtime = "opencode"
    RuntimeCLI        Runtime = "cli"
)

type AgentConfig struct {
    Runtime      Runtime   `json:"runtime"`
    Model        string    `json:"model"`
    Role         AgentRole `json:"role"`
    Skills       []string  `json:"skills,omitempty"`
    ContextFiles []string  `json:"context_files,omitempty"`
}

type ExecutionStrategy string

const (
    StrategySingle   ExecutionStrategy = "single"
    StrategyMulti    ExecutionStrategy = "multi"
    StrategyParallel ExecutionStrategy = "parallel"
)

type ActionMapping struct {
    Action          SpecAction        `json:"action"`
    Primary         AgentConfig       `json:"primary"`
    Fallback        *AgentConfig      `json:"fallback,omitempty"`
    Strategy        ExecutionStrategy `json:"strategy"`
    ApprovalRequired bool             `json:"approval_required"`
    MaxWorkers      int               `json:"max_workers,omitempty"`
    Checklist       []string          `json:"checklist,omitempty"`
}

type Execution struct {
    ID          string            `json:"id"`
    Action      SpecAction        `json:"action"`
    ChangeID    string            `json:"change_id,omitempty"`
    TaskIDs     []string          `json:"task_ids,omitempty"`
    Agents      []AgentConfig     `json:"agents"`
    Strategy    ExecutionStrategy `json:"strategy"`
    Status      ExecutionStatus   `json:"status"`
    Steps       []ExecutionStep   `json:"steps"`
    StartedAt   time.Time         `json:"started_at"`
    CompletedAt *time.Time        `json:"completed_at,omitempty"`
    Cost        CostEstimate      `json:"cost"`
}

type ExecutionStatus string

const (
    StatusQueued    ExecutionStatus = "queued"
    StatusRunning   ExecutionStatus = "running"
    StatusCompleted ExecutionStatus = "completed"
    StatusFailed    ExecutionStatus = "failed"
    StatusCancelled ExecutionStatus = "cancelled"
)

type ExecutionStep struct {
    ID        string          `json:"id"`
    Agent     AgentConfig     `json:"agent"`
    TaskID    string          `json:"task_id,omitempty"`
    Status    ExecutionStatus `json:"status"`
    Output    string          `json:"output,omitempty"`
    Error     string          `json:"error,omitempty"`
    StartedAt time.Time       `json:"started_at"`
    Duration  time.Duration   `json:"duration"`
    Tokens    TokenUsage      `json:"tokens"`
}

type TokenUsage struct {
    Input    int `json:"input"`
    Output   int `json:"output"`
    Cached   int `json:"cached"`
    Thinking int `json:"thinking,omitempty"`
}

type CostEstimate struct {
    EstimatedUSD float64    `json:"estimated_usd"`
    ActualUSD    float64    `json:"actual_usd,omitempty"`
    Breakdown    []CostItem `json:"breakdown,omitempty"`
}

type CostItem struct {
    Model    string  `json:"model"`
    Tokens   int     `json:"tokens"`
    CostUSD  float64 `json:"cost_usd"`
}
```

### 7.2 Spec Types

```go
package domain

type Spec struct {
    ID          string    `json:"id"`
    Path        string    `json:"path"`
    Capability  string    `json:"capability"`
    Content     string    `json:"content"`
    Requirements []Requirement `json:"requirements"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Requirement struct {
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Level     string     `json:"level"` // SHALL, MUST, SHOULD
    Scenarios []Scenario `json:"scenarios"`
}

type Scenario struct {
    Name   string   `json:"name"`
    When   string   `json:"when"`
    Then   string   `json:"then"`
    Given  string   `json:"given,omitempty"`
}

type Change struct {
    ID          string       `json:"id"`
    Status      ChangeStatus `json:"status"`
    Proposal    string       `json:"proposal"`    // proposal.md content
    Tasks       []Task       `json:"tasks"`
    Design      string       `json:"design,omitempty"` // design.md content
    SpecDeltas  []SpecDelta  `json:"spec_deltas"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
    ArchivedAt  *time.Time   `json:"archived_at,omitempty"`
}

type ChangeStatus string

const (
    StatusDraft           ChangeStatus = "draft"
    StatusPendingApproval ChangeStatus = "pending_approval"
    StatusApproved        ChangeStatus = "approved"
    StatusImplementing    ChangeStatus = "implementing"
    StatusCompleted       ChangeStatus = "completed"
    StatusArchived        ChangeStatus = "archived"
    StatusRejected        ChangeStatus = "rejected"
)

type Task struct {
    ID          string     `json:"id"`
    Section     string     `json:"section"`
    Description string     `json:"description"`
    Status      TaskStatus `json:"status"`
    Assignee    string     `json:"assignee,omitempty"`
    Subtasks    []Task     `json:"subtasks,omitempty"`
}

type TaskStatus string

const (
    TaskOpen       TaskStatus = "open"
    TaskInProgress TaskStatus = "in_progress"
    TaskCompleted  TaskStatus = "completed"
    TaskDeferred   TaskStatus = "deferred"
)

type SpecDelta struct {
    SpecID   string              `json:"spec_id"`
    Path     string              `json:"path"`
    Added    []Requirement       `json:"added,omitempty"`
    Modified []RequirementDelta  `json:"modified,omitempty"`
    Removed  []string            `json:"removed,omitempty"`
}

type RequirementDelta struct {
    ID     string `json:"id"`
    Before string `json:"before"`
    After  string `json:"after"`
}
```

---

## 8. Configuration

### 8.1 Project Configuration

```yaml
# .goent/config.yaml

project:
  name: "my-service"
  type: "go-microservice"
  repository: "github.com/org/my-service"

# Tech stack detection (auto-generated or manual)
tech_stack:
  language: "go"
  version: "1.23"
  framework: "stdlib"
  database: ["postgresql", "redis"]
  messaging: ["kafka"]

# Build and test commands
commands:
  build: "make build"
  test: "make test"
  lint: "make lint"
  format: "make fmt"

# OpenSpec paths
openspec:
  root: "openspec"
  specs: "openspec/specs"
  changes: "openspec/changes"
  archive: "openspec/archive"

# Agent configuration
agents:
  # Default model preferences
  models:
    high_reasoning: "claude-opus-4-5"
    balanced: "claude-sonnet-4-5"
    fast: "glm-4.7"
  
  # Runtime preferences
  runtimes:
    planning: "claude-code"
    implementation: "opencode"
    review: "claude-code"
  
  # Cost limits
  budget:
    per_execution: 1.0  # USD
    per_day: 10.0       # USD
    warn_threshold: 0.5 # USD

# Enterprise standards for review
standards:
  enabled: true
  checklists:
    - "clean-architecture"
    - "solid-principles"
    - "error-handling"
    - "testing-patterns"
```

### 8.2 Agent Matrix Override

```yaml
# .goent/agent-matrix.yaml

# Override default action mappings
overrides:
  proposal:
    primary:
      runtime: claude-code
      model: claude-opus-4-5
      roles: [product, architect]
    # Skip approval for small changes
    approval_required: false
    approval_threshold_loc: 100  # Only require approval for >100 LOC

  execute:
    # Use cheaper model for implementation
    primary:
      runtime: opencode
      model: glm-4.7
      role: developer
    # Limit parallelism
    max_workers: 2

  review:
    # Custom checklist
    checklist:
      - "clean-architecture"
      - "go-idioms"
      - "domain-driven-design"
```

---

## 9. Success Metrics

### 9.1 Efficiency Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time to proposal | < 5 min | From command to validated proposal |
| Time to plan | < 10 min | From proposal to tasks.md |
| Parallel execution speedup | 2-3x | vs. sequential execution |
| Cost per change | < $5 | Average across all actions |

### 9.2 Quality Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| First-pass review approval | > 80% | Changes passing initial review |
| Spec compliance | 100% | All outputs follow OpenSpec format |
| Test coverage generated | > 70% | For implemented tasks |

### 9.3 Adoption Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Daily active users | Track growth | MCP tool invocations |
| Commands per session | > 5 | User engagement |
| Agent selection satisfaction | > 4/5 | User feedback |

---

## 10. Future Considerations

### 10.1 Phase 2 Features

- **Learning agent selection**: ML-based optimization of agent selection
- **Custom agent definitions**: User-defined roles and prompts
- **Team collaboration**: Multi-user approval workflows
- **CI/CD integration**: GitHub Actions for spec validation

### 10.2 Phase 3 Features

- **Cross-repository specs**: Monorepo and multi-repo support
- **Spec versioning**: Git-like history for specs
- **Analytics dashboard**: Cost, time, and quality metrics
- **Plugin marketplace**: Community agents and skills

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **Spec Action** | A user intent mapped to agent orchestration (e.g., proposal, plan, execute) |
| **Agent** | An AI model instance with specific role and configuration |
| **Runtime** | The execution environment (Claude Code, OpenCode, CLI) |
| **Role** | Agent specialization affecting prompts (Product, Architect, Senior, etc.) |
| **Change** | An OpenSpec change proposal with tasks and spec deltas |
| **Spec Delta** | Additions, modifications, or removals to an existing spec |

---

## Appendix B: Example Session

```bash
# 1. Create proposal using Opus 4.5 with Product+Architect roles
$ goent run proposal "Add rate limiting to API endpoints"
[Opus 4.5/Product+Architect] Analyzing codebase...
[Opus 4.5/Product+Architect] Creating proposal...
✓ Created openspec/changes/add-rate-limiting/
  - proposal.md
  - specs/api/spec.md (delta)

# 2. Generate tasks using Opus 4.5 with Architect+Senior roles  
$ goent run plan --change add-rate-limiting
[Opus 4.5/Architect+Senior] Breaking down proposal...
✓ Created openspec/changes/add-rate-limiting/tasks.md
  - 3 sections, 12 tasks

# 3. Execute in parallel using GLM 4.7 with Developer role
$ goent run execute --change add-rate-limiting --parallel
[GLM 4.7/Developer] Worker 1: Task 1.1 - Add rate limit config...
[GLM 4.7/Developer] Worker 2: Task 1.2 - Create middleware...
[GLM 4.7/Developer] Worker 3: Task 2.1 - Add Redis client...
✓ Completed 12/12 tasks
  Cost: $0.42 | Time: 3m 24s

# 4. Review using Opus 4.5 with Reviewer role
$ goent run review --change add-rate-limiting
[Opus 4.5/Reviewer] Reviewing against enterprise standards...
✓ Review complete: 2 suggestions, 0 blockers

# 5. Archive
$ goent run archive --change add-rate-limiting
✓ Archived to openspec/archive/2026-01-04-add-rate-limiting/
✓ Updated openspec/specs/api/spec.md
```
