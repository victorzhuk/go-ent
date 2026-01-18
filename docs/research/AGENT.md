# AI Agent Architecture for Claude Code: A Complete Engineering Guide

**Claude Code represents Anthropic's most powerful implementation of agentic coding workflows**, combining a deliberately low-level, unopinionated architecture with sophisticated multi-agent orchestration capabilities. This guide synthesizes official Anthropic documentation, industrial research from 2024-2025, and practical community implementations to provide a comprehensive framework for building AI agents—from single-agent patterns through complex multi-agent systems.

The core insight from Anthropic's research: "The most successful implementations weren't using complex frameworks or specialized libraries. Instead, they were building with simple, composable patterns." This philosophy underlies everything in Claude Code's design.

---

## The agentic loop: foundation of all agent systems

Every agent system rests on a deceptively simple architectural pattern called the **agentic loop**. Claude receives input, requests tool actions, your application executes those tools, results return to Claude, and the cycle repeats until Claude produces a final response.

```
User Input → Claude → Tool Request → Execution → Results → Claude → [repeat] → Final Response
```

Claude Code implements this as a **single-threaded master loop** (codenamed `nO`) with one flat list of messages—no sprawling agent swarms. This architectural decision prevents "chaos of uncontrolled agent proliferation" while still enabling powerful orchestration through controlled subagent spawning.

**Critical implementation details:**
- Automatic context compaction triggers at ~92% of context window usage
- Maximum one subagent branch active at a time to maintain coherence
- Memory persists in simple Markdown documents for transparency and debuggability
- Subagents cannot spawn their own subagents, preventing recursive complexity

This design choice reflects a fundamental tension in agent engineering: **autonomy vs. control**. More autonomous agents can accomplish more complex tasks but become harder to predict, debug, and constrain. Claude Code opts for controlled power.

---

## Six composable patterns from Anthropic's research

Anthropic's seminal "Building Effective Agents" research identifies six foundational patterns that combine to create sophisticated agentic systems. Understanding when—and when not—to use each pattern is essential.

### Prompt chaining decomposes tasks sequentially

**Prompt chaining** breaks complex tasks into a sequence of steps where each LLM call processes the output of the previous one. Use this when tasks "can be easily and cleanly decomposed into fixed subtasks" and you're willing to trade latency for accuracy.

The key enhancement over simple sequential calls: add **programmatic gates** between steps for quality control. Generate marketing copy, verify it meets brand guidelines, then translate—failing the gate triggers regeneration rather than propagating errors downstream.

### Routing directs inputs to specialized handlers

**Routing** classifies inputs and directs them to specialized followup processes. This enables separation of concerns and more focused prompts—customer service queries about refunds go to the refund handler, technical questions to technical support.

A powerful cost optimization: route simple questions to Claude Haiku and complex ones to Claude Sonnet. The router's classification overhead pays for itself through cheaper downstream processing.

### Parallelization exploits independence for speed

**Parallelization** runs LLM calls simultaneously when tasks are independent, with two key variations. **Sectioning** divides work into independent subtasks (run security review, performance review, and style review concurrently). **Voting** runs the same task multiple times for higher confidence through consensus.

Anthropic's Research feature demonstrates the impact: "Parallel tool calling transforms speed and performance... These changes cut research time by **up to 90%** for complex queries."

### Orchestrator-worker handles dynamic complexity

When you can't predict subtasks in advance, the **orchestrator-worker** pattern shines. A central LLM analyzes the task, spawns workers dynamically, and synthesizes their results. The key difference from parallelization: subtasks emerge at runtime rather than being predefined.

Each worker needs clear specifications: "an objective, an output format, guidance on tools and sources, and clear task boundaries." Vague delegation produces vague results.

### Evaluator-optimizer enables iterative refinement

**Evaluator-optimizer** pairs a generator LLM with an evaluator LLM in a feedback loop. The generator produces output, the evaluator scores it against criteria, and the loop continues until quality thresholds are met.

Use this pattern "when we have clear evaluation criteria, and when iterative refinement provides measurable value"—literary translation requiring nuance preservation, or complex searches requiring multiple analysis rounds.

### Tool use is the fundamental building block

**Tool use** transforms a base LLM into an augmented agent capable of interacting with external systems. This isn't just one pattern among equals—it's the foundation underlying all others. Anthropic's guidance: "Agents can handle sophisticated tasks, but their implementation is often straightforward. They are typically just LLMs using tools based on environmental feedback in a loop."

---

## Claude Code's three-tier extension system

Claude Code provides three distinct mechanisms for customization, each serving different purposes: **Skills** (model-invoked capabilities), **Commands** (user-invoked shortcuts), and **Subagents** (specialized AI assistants).

### Skills: domain knowledge Claude discovers automatically

Skills are **model-invoked**—Claude autonomously decides when to use them based on task requirements and skill descriptions. Each skill lives in a directory containing a `SKILL.md` file with YAML frontmatter.

**Required SKILL.md structure:**
```markdown
---
name: testing-patterns
description: Jest testing patterns for this project. Use when writing tests, creating mocks, or following TDD workflow.
allowed-tools: Read, Grep, Glob, Bash(npm:*)
---

# Testing Patterns

## Test Structure
- Use `describe` blocks for grouping
- Follow AAA pattern: Arrange, Act, Assert

## Mocking
- Use factory functions: `getMockUser(overrides)`
- Mock external dependencies, not internal modules
```

**Frontmatter field constraints:**
| Field | Required | Limits |
|-------|----------|--------|
| `name` | Yes | Lowercase, hyphens, max 64 chars |
| `description` | Yes | Max 1024 chars; must explain WHAT it does and WHEN to use it |
| `allowed-tools` | No | Comma-separated list restricting tool access |
| `model` | No | Specify `claude-sonnet-4-20250514` or similar |

Skills use a **progressive disclosure** architecture to manage context efficiently:
- **Level 1**: Only name and description load at startup (~100 tokens)
- **Level 2**: Full SKILL.md body loads when Claude determines relevance (<5k tokens)
- **Level 3+**: Referenced files (FORMS.md, scripts/) load only when needed

**Storage locations** determine scope:
- `~/.claude/skills/` → Personal skills, all projects
- `.claude/skills/` → Project skills, shared via git

### Commands: user-invoked shortcuts for common tasks

Slash commands provide explicit user control through Markdown files that become available in the `/command` menu. They support argument interpolation and inline bash execution.

**Command file with frontmatter:**
```markdown
---
description: Analyze and fix a GitHub issue
allowed-tools: Bash(git:*,gh:*), Read, Write, Edit
argument-hint: [issue-number]
---

Please analyze and fix GitHub issue #$1.

Current branch: !`git branch --show-current`
Recent commits: !`git log --oneline -5`

Follow these steps:
1. Use `gh issue view $1` to get details
2. Search codebase for relevant files
3. Implement changes following TDD
4. Create PR with descriptive message
```

**Variable substitution:**
- `$ARGUMENTS` → All arguments as single string
- `$1`, `$2`, `$3` → Individual positional arguments
- `!`backticks`` → Execute bash and insert output

**Storage locations:**
- `.claude/commands/` → Project commands, shown as "(project)"
- `~/.claude/commands/` → Personal commands, shown as "(user)"

### Subagents: isolated specialists with their own context

Subagents are **specialized AI assistants** that run in separate context windows with custom system prompts, specific tool access, and independent permissions. This isolation prevents information overload in the main conversation.

**Subagent definition file:**
```markdown
---
name: code-reviewer
description: Reviews code for quality, security, and conventions. Use proactively after code changes.
model: opus
tools: Read, Grep, Glob
permissionMode: default
skills: testing-patterns
---

You are a senior code reviewer focusing on quality, security, and best practices.

## Review Process
1. Run `git diff` to see recent changes
2. Focus on modified files only
3. Begin review immediately

## Checklist
- [ ] No TypeScript `any` types
- [ ] Proper error handling present
- [ ] Tests included for new code
- [ ] No exposed secrets
```

**Configuration options:**
| Field | Values | Purpose |
|-------|--------|---------|
| `model` | `sonnet`, `opus`, `haiku`, `'inherit'` | Which Claude model to use |
| `permissionMode` | `default`, `acceptEdits`, `bypassPermissions`, `plan` | How to handle permissions |
| `tools` | Comma-separated list | Which tools the subagent can access |
| `skills` | Comma-separated list | Skills to auto-load on startup |

**Built-in subagents provide common patterns:**
- **Explore Subagent** (Haiku): Fast, read-only codebase exploration
- **Plan Subagent** (Sonnet): Research during planning before presenting options
- **General-purpose Subagent** (Sonnet): Complex multi-step tasks requiring exploration and modification

To encourage proactive use, include phrases like "use PROACTIVELY" or "MUST BE USED" in subagent descriptions.

---

## CLAUDE.md: the highest leverage customization point

`CLAUDE.md` is a special file automatically pulled into context when starting a conversation—making it **the single most important place to customize Claude Code behavior**.

**Recommended structure follows WHY-WHAT-HOW:**
```markdown
# Project Overview
Brief description of what this codebase does and why

# Tech Stack
- Framework: Next.js 14 with App Router
- Language: TypeScript 5.x strict mode
- Testing: Vitest + React Testing Library
- Styling: Tailwind CSS

# Key Commands
- `npm run dev`: Start development server
- `npm run test`: Run test suite
- `npm run typecheck`: Type checking

# Code Style
- Use ES modules (import/export), not CommonJS
- Prefer functional components with hooks
- Destructure imports when possible
- Follow patterns in `src/components/Button.tsx`

# Workflow
- IMPORTANT: Run typecheck after making code changes
- Prefer running single tests, not the whole suite
- Commit atomic changes with descriptive messages
```

**Critical insight from Anthropic research**: Frontier LLMs can follow ~150-200 instructions with reasonable consistency, but Claude Code's system prompt already contains ~50 instructions. **Less is more**—keep CLAUDE.md focused on universally applicable guidance.

**Multiple CLAUDE.md files combine hierarchically:**
1. Project root `CLAUDE.md` (shared via git, recommended)
2. Project root `CLAUDE.local.md` (gitignored, personal preferences)
3. Parent directories (for monorepos)
4. Child directories (loaded on-demand when exploring those areas)
5. Home folder `~/.claude/CLAUDE.md` (applies to all sessions)

Use the `#` key during coding to add instructions that Claude will incorporate into CLAUDE.md. Add emphasis ("IMPORTANT", "YOU MUST") to improve adherence to critical instructions.

---

## Tool design principles for effective agents

Anthropic's core guidance: "Put as much effort into the Agent-Computer Interface (ACI) as you would in Human-Computer Interface (HCI)." Poor tool descriptions send agents down completely wrong paths.

**Good tool description (from Anthropic documentation):**
```json
{
  "name": "get_stock_price",
  "description": "Retrieves the current stock price for a given ticker symbol. The ticker symbol must be a valid symbol for a publicly traded company on a major US stock exchange like NYSE or NASDAQ. The tool will return the latest trade price in USD. It should be used when the user asks about the current or most recent price of a specific stock. It will not provide any other information about the stock or company.",
  "input_schema": {
    "properties": {
      "ticker": {
        "type": "string",
        "description": "The stock ticker symbol, e.g. AAPL for Apple Inc."
      }
    }
  }
}
```

**Bad tool description:**
```json
{
  "name": "get_stock_price",
  "description": "Gets the stock price for a ticker."
}
```

The difference: **3-4 sentences minimum** explaining what the tool does, important caveats, limitations, and when to use it versus alternatives.

**Tool permission patterns for subagents:**
| Agent Type | Recommended Tools |
|------------|-------------------|
| Read-only reviewers | `Read, Grep, Glob` |
| Research agents | `Read, Grep, Glob, WebFetch, WebSearch` |
| Code writers | `Read, Write, Edit, Bash, Glob, Grep` |
| Documentation agents | `Read, Write, Edit, Glob, Grep, WebFetch` |

---

## Context engineering for long-horizon tasks

Context management becomes critical for agents handling complex, multi-step tasks. Claude Code implements several strategies automatically, but understanding them helps you design better workflows.

**Automatic compaction** triggers at ~92% context window usage. The system summarizes older content while preserving recent, critical information. Design workflows assuming this will happen—don't rely on perfect recall of early conversation content.

**Subagent isolation** prevents context pollution. Each subagent operates in its own context window, so specialized research doesn't overload the main conversation with irrelevant details. Results return compressed.

**Memory externalization** stores important information in Markdown files that can be re-read when needed. From Anthropic: "Agents summarize completed work phases and store essential information in external memory before proceeding to new tasks."

**Extended thinking triggers** increase Claude's reasoning budget:
```
"think" < "think hard" < "think harder" < "ultrathink"
```

Each phrase maps to increasing compute allocation for complex reasoning tasks.

---

## Multi-agent orchestration patterns

For complex tasks requiring heavy parallelization, information exceeding single context windows, or numerous specialized tools, multi-agent systems become necessary.

### The orchestrator-worker pattern

Anthropic's Research feature demonstrates the canonical pattern: "A lead agent coordinates the process while delegating to specialized subagents that operate in parallel."

**Scaling effort to query complexity:**
- **Simple fact-finding**: 1 agent, 3-10 tool calls
- **Direct comparisons**: 2-4 subagents, 10-15 calls each
- **Complex research**: 10+ subagents with divided responsibilities

**Critical delegation requirements** for each subagent:
1. Clear objective (what to accomplish)
2. Output format specification (how to return results)
3. Guidance on tools and sources (what resources to use)
4. Clear task boundaries (what NOT to do)

### Hub-and-spoke coordination

A central orchestrator routes tasks to specialists:
```
User Request → Orchestrator
    ├── Code Agent (implementation)
    ├── Review Agent (quality checks)
    ├── Research Agent (documentation)
    └── Test Agent (verification)
            ↓
    Results → Orchestrator → Synthesis → User
```

**Key design constraint**: Subagents cannot spawn their own subagents. This prevents recursive complexity explosions.

### Sequential orchestration for refinement

```
Agent 1 (Generate) → Agent 2 (Validate) → Agent 3 (Refine)
```

Each agent receives the previous agent's output, performs a specific transformation, and passes results forward. This pattern excels at iterative quality improvement.

---

## XML structuring for agent prompts

XML tags provide "boundary markers" for structured information, improving semantic precision and parsing accuracy while reducing ambiguity.

**Effective XML structure for agent instructions:**
```xml
<system_instructions>
  You are a code review specialist focusing on security and performance.
</system_instructions>

<context>
  <codebase>TypeScript React application</codebase>
  <recent_changes>!`git diff HEAD~5`</recent_changes>
</context>

<task>
  <description>Review the recent changes for security vulnerabilities</description>
  <requirements>
    <item>Check for exposed secrets</item>
    <item>Verify input validation</item>
    <item>Assess XSS prevention</item>
  </requirements>
</task>

<output_format>
  <type>Markdown</type>
  <sections>Summary, Critical Issues, Recommendations</sections>
</output_format>
```

**Best practices from Anthropic:**
- Use consistent tag names throughout prompts
- Nest tags for hierarchical content
- Reference tags in instructions: "Using the contract in `<contract>` tags..."
- Combine with other techniques like few-shot examples

---

## Chain-of-thought patterns for reasoning

Complex agent tasks benefit from structured reasoning patterns that make the agent's thought process explicit and correctable.

### ReAct: reasoning plus acting

The ReAct pattern interleaves reasoning with action:
```
Question: {input}
Thought: I should check the current branch first
Action: execute_command
Action Input: git branch --show-current
Observation: feature/auth-refactor
Thought: Now I'll look at recent changes on this branch
Action: execute_command
Action Input: git log --oneline -5
Observation: [commit history]
Thought: I now have enough context to proceed
Final Answer: [synthesized response]
```

This pattern creates an audit trail of the agent's decision-making, making failures debuggable.

### Tree-of-thought for complex reasoning

For problems requiring exploration of multiple solution paths, **tree-of-thought** structures reasoning as a tree where each node represents an intermediate thought. The agent uses breadth-first or depth-first search to explore possibilities, evaluating each reasoning step before committing.

---

## TDD workflows with AI agents

Test-driven development adapts naturally to agent workflows, providing both verification and guardrails.

**Red-Green-Refactor with agents:**

**Phase 1 (RED):** Ask the agent to write failing tests based on specifications. Run tests to confirm they fail. Commit tests before any implementation.

**Phase 2 (GREEN):** Instruct the agent to implement *only enough code* to pass tests. Critical constraint: "Do NOT modify the tests." The agent iterates until all tests pass.

**Phase 3 (REFACTOR):** Ask the agent to "clean up logic but keep all tests green." Run tests after each refactor step.

**Prompt template for TDD:**
```markdown
I'm implementing: [feature name]

Business rules:
- [Rule 1]
- [Rule 2]

Please follow TDD strictly:
1. Write ONE failing test
2. Show me the test, wait for approval
3. Implement code to pass
4. Run tests, show results
5. Move to next test

NEVER modify tests to make them pass.
One behavior per test.
Start with high-value behavior, not edge cases.
```

---

## Error handling and guardrails

Agents are stateful and errors compound. Without effective mitigations, minor failures become catastrophic.

**Core guardrail strategies:**
- **Iteration limits**: Prevent infinite loops (`max_iterations = 10`)
- **Sandboxed environments**: Run in containerized environments
- **Permission systems**: Write operations and risky Bash commands require explicit approval
- **Checkpointing**: Build systems that can resume from known-good states

**When to escalate to humans:**
- Exceeding failure thresholds (retry limits)
- High-risk actions (irreversible, large financial impact)
- Low confidence scores
- Unrecoverable errors

**Anthropic's philosophy**: "Let the agent know when a tool is failing and let it adapt. Combine AI adaptability with deterministic safeguards (retry logic, checkpoints)."

---

## Production directory structure

A complete Claude Code project setup:
```
your-project/
├── CLAUDE.md                      # Project memory (required)
├── CLAUDE.local.md                # Personal overrides (gitignored)
├── .mcp.json                      # MCP server configuration
├── .claude/
│   ├── settings.json              # Hooks, environment, permissions
│   ├── settings.local.json        # Personal overrides
│   │
│   ├── agents/                    # Custom subagents
│   │   ├── code-reviewer.md       
│   │   ├── debugger.md
│   │   └── research-agent.md
│   │
│   ├── commands/                  # Slash commands
│   │   ├── commit.md
│   │   ├── review.md
│   │   └── tdd-cycle.md
│   │
│   ├── skills/                    # Agent skills
│   │   ├── testing-patterns/
│   │   │   └── SKILL.md
│   │   └── api-design/
│   │       ├── SKILL.md
│   │       └── schemas/
│   │           └── openapi.yaml
│   │
│   └── hooks/                     # Event-driven automation
│       └── pre-commit.sh
│
└── .github/
    └── workflows/
        └── claude-review.yml      # CI integration
```

---

## Key repositories and resources

| Purpose | Repository |
|---------|------------|
| Official skill template | github.com/anthropics/skills |
| Comprehensive setup example | github.com/ChrisWiles/claude-code-showcase |
| Production slash commands (57+) | github.com/wshobson/commands |
| Multi-agent orchestration | github.com/wshobson/agents |
| 100+ specialized subagents | github.com/VoltAgent/awesome-claude-code-subagents |
| Skill factory/builder | github.com/alirezarezvani/claude-code-skill-factory |
| Curated collections | github.com/hesreallyhim/awesome-claude-code |

**Official Anthropic documentation:**
- Agent design patterns: anthropic.com/research/building-effective-agents
- Multi-agent systems: anthropic.com/engineering/multi-agent-research-system
- Claude Code best practices: anthropic.com/engineering/claude-code-best-practices
- Agent skills: code.claude.com/docs/en/skills
- Subagents: code.claude.com/docs/en/sub-agents
- Slash commands: code.claude.com/docs/en/slash-commands

---

## Conclusion: principles for effective agent engineering

**Start simple and add complexity only when needed.** Anthropic's consistent guidance: "We recommend finding the simplest solution possible." A well-crafted CLAUDE.md file often outperforms elaborate multi-agent systems for most tasks.

**Design tools as carefully as interfaces.** The Agent-Computer Interface deserves the same attention as Human-Computer Interface design. Every tool description should be 3-4 sentences minimum, explaining not just what the tool does but when to use it and what limitations apply.

**Use subagents for isolation, not just parallelization.** The primary benefit of subagents isn't speed—it's keeping the main conversation context clean and focused. Use them when specialized tasks would pollute the main context with irrelevant details.

**Build for failure recovery.** Agents will fail. Design checkpointing, clear handoff protocols, and human escalation paths. Let agents know when tools fail so they can adapt rather than spiral.

**Progressive disclosure manages complexity.** Skills load metadata first, then full content, then supporting files—only what's needed. Apply this principle to your own designs: surface just enough information for the current decision.

The most powerful pattern isn't any single technique—it's the disciplined composition of simple patterns into coherent systems. Master the six composable patterns, understand when each applies, and combine them thoughtfully for your specific use cases.