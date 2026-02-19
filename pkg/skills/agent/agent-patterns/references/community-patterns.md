## Core Composable Patterns (Anthropic's 6 Patterns)

### 1. Prompt Chaining
- Break complex tasks into sequential steps
- Each step's output feeds the next step's input
- Add validation gates between steps
- Best for: document generation, multi-stage analysis

### 2. Routing
- Classify input and route to specialized handlers
- Use complexity scoring (1-10) to choose agent variant
- Fast path for simple tasks, heavy path for complex ones
- Best for: customer support, code review triage

### 3. Parallelization
- Run independent subtasks concurrently
- Aggregate results after all complete
- Use when subtasks don't depend on each other
- Best for: multi-file analysis, batch processing

### 4. Orchestrator-Workers
- Orchestrator breaks task into subtasks
- Workers execute independently and report back
- Orchestrator synthesizes results
- Best for: complex coding tasks, research

### 5. Evaluator-Optimizer
- Generate initial output, then evaluate quality
- Iterate until quality threshold met
- Use separate models/prompts for generation and evaluation
- Best for: code quality, content refinement

### 6. Autonomous Agent (Tool-Use Loop)
- Agent decides actions, uses tools, observes results
- Loops until task complete or limit reached
- Requires guardrails: max iterations, allowed tools, timeout
- Best for: debugging, research, complex problem solving

## Multi-Agent Architecture

### Driver-Orchestrator Pattern
```
Driver (read-only, planning)
├── Coder (implementation)
├── Tester (validation)
├── Reviewer (quality)
├── Debugger (troubleshooting)
└── Planner (design, specs)
```

- Driver NEVER modifies files directly
- Driver delegates to specialized agents
- Each agent has specific skills and tool access
- Use complexity scoring to choose fast vs heavy agents

### Agent Design Principles
- Single Responsibility: Each agent has one clear purpose
- Explicit delegation: Use platform-native syntax (@agent, task tool)
- Skill loading: Agents reference SKILL.md files for expertise
- Context management: Pass only relevant context to each agent
- Output contracts: Define expected output format per agent

## Prompt Engineering for Agents
- System prompt defines agent personality, capabilities, constraints
- Use XML tags for structured sections
- Include examples of expected behavior
- Define tool usage patterns explicitly
- Set clear boundaries: what the agent should and shouldn't do
- Include error handling instructions

## Workflow Management
- Use proposal/spec documents for complex features
- Track tasks with checkboxes and status updates
- Implement phase-based execution: Assess → Plan → Implement → Validate → Complete
- Archive completed work for future reference
- Use ADRs for architectural decisions
