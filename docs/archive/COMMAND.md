# Mastering Claude Code custom commands and agent prompts

Creating industrial-strength slash commands and agentic prompts for Claude Code requires understanding three interconnected systems: **command file structure**, **prompt engineering patterns**, and **agent orchestration principles**. This comprehensive guide synthesizes Anthropic's official documentation with production-tested patterns to give you complete mastery over Claude Code's extensibility.

Custom slash commands live as Markdown files in `.claude/commands/` (project-level) or `~/.claude/commands/` (user-level), with optional YAML frontmatter controlling behavior. The most effective commands combine **structured prompts using XML tags**, **dynamic context injection via bash execution**, and **clear step-by-step instructions** that leverage Claude's reasoning capabilities.

## Command file anatomy and configuration options

Every slash command follows a simple pattern: a Markdown file where the filename becomes the command name. Creating `optimize.md` gives you `/optimize`. The real power comes from frontmatter configuration and dynamic content injection.

**Complete frontmatter reference:**

```yaml
---
description: Brief description shown in /help and used for auto-invocation
allowed-tools: Read, Edit, Bash(git:*), Grep, Glob
argument-hint: [required-arg] [optional-arg]
model: claude-sonnet-4-5-20250929
disable-model-invocation: false
---
```

The `allowed-tools` field uses permission patterns: `Bash(git add:*)` allows git add with any arguments, `Bash(npm *)` allows npm with any subcommand, and plain `Bash` allows unrestricted shell access. The `description` field serves dual purposes—it appears in `/help` output and enables the SlashCommand tool to auto-invoke your command when relevant.

**Dynamic content injection** happens through three mechanisms. The `$ARGUMENTS` variable captures everything passed after the command name, while `$1`, `$2`, `$3` access positional arguments. Inline bash execution uses the `!`command`` syntax (note the exclamation prefix) to run shell commands before the prompt is sent to Claude:

```markdown
---
allowed-tools: Bash(git:*)
argument-hint: [branch-name]
description: Create feature branch with context
---

## Current State
- Branch: !`git branch --show-current`
- Status: !`git status --short`
- Recent commits: !`git log --oneline -5`

## Task
Create feature branch: $1
Base it on the appropriate parent branch given the current state.
```

File references use the `@` prefix—`@src/utils/helpers.js` includes that file's contents in the prompt. This creates powerful context-aware commands without manual copy-pasting.

## Prompt engineering patterns that work

Anthropic's research establishes XML tags as the **cornerstone of effective prompting**. Claude was specifically trained to recognize XML tags as semantic separators, making them more reliable than markdown headers or plain text delimiters.

**The essential XML structure:**

```xml
<context>Background information and current state</context>
<task>The specific action to perform</task>
<constraints>Rules and boundaries for the output</constraints>
<output_format>Expected structure of the response</output_format>
<examples>
  <example>
    <input>Sample input</input>
    <output>Expected output format</output>
  </example>
</examples>
```

**Chain-of-thought prompting** significantly improves accuracy for complex tasks. The structured approach separates reasoning from answers:

```markdown
Think through this step-by-step in <thinking> tags, then provide your answer in <answer> tags.

Consider:
1. What files are affected?
2. What could break?
3. What tests need updating?
```

**Few-shot examples** are Anthropic's "secret weapon"—3-5 diverse examples dramatically reduce misinterpretation and improve output consistency. Place examples in `<examples>` tags with clear `<input>` and `<output>` pairs that demonstrate exactly the format and quality you expect.

For **system-level behavior**, use the `model` frontmatter to select the right capability tier. Use `claude-3-5-haiku-20241022` for fast, simple tasks; `claude-sonnet-4-5-20250929` for balanced performance; and Opus for complex reasoning that requires maximum capability.

## Production-ready command examples

**TDD enforcement command** (`/tdd`):

```markdown
---
description: Implement feature using strict Test-Driven Development
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
argument-hint: <feature-description>
---

<task>Implement using strict TDD: $ARGUMENTS</task>

<workflow>
## Phase 1: RED 🔴
Write a failing test first. The test MUST fail before proceeding.
Run the test and confirm failure output.

## Phase 2: GREEN 🟢  
Write MINIMAL code to make the test pass.
No additional features. No premature optimization.
Run test and confirm it passes.

## Phase 3: REFACTOR 🔵
Improve code quality while keeping tests green.
Extract functions, improve naming, remove duplication.
Run tests after each refactor to verify no regressions.
</workflow>

<constraints>
- Never skip the RED phase
- Never modify tests to make them pass
- Commit after each complete RED-GREEN-REFACTOR cycle
</constraints>
```

**Comprehensive code review command** (`/review`):

```markdown
---
allowed-tools: Read, Grep, Glob, Bash(git diff:*)
description: Security-focused code review with actionable feedback
---

<context>
## Changes to Review
!`git diff --name-only HEAD~1`

## Detailed Diff
!`git diff HEAD~1`
</context>

<task>Review these changes thoroughly</task>

<checklist>
- [ ] Security: SQL injection, XSS, auth bypasses, exposed secrets
- [ ] Error handling: All failure paths covered
- [ ] Performance: N+1 queries, memory leaks, inefficient algorithms
- [ ] Types: No `any`, proper null handling
- [ ] Tests: New code has corresponding tests
- [ ] Documentation: Public APIs documented
</checklist>

<output_format>
For each issue found:
1. **Location**: file:line
2. **Severity**: Critical | High | Medium | Low
3. **Issue**: Clear description
4. **Fix**: Specific code suggestion

Score confidence 0-100 for each issue. Only report issues with confidence ≥80.
</output_format>
```

**GitHub issue resolver** (`/fix-issue`):

```markdown
---
argument-hint: <issue-number>
description: Analyze and fix GitHub issue end-to-end
allowed-tools: Bash(gh:*), Read, Write, Edit, Grep, Glob
---

<task>Fix GitHub issue #$1</task>

<workflow>
1. **Understand**: Use `gh issue view $1` to get full context
2. **Locate**: Search codebase for relevant files
3. **Plan**: Think through the fix in <thinking> tags
4. **Test First**: Write a test that reproduces the bug
5. **Implement**: Make minimal changes to fix the issue
6. **Verify**: Run tests, ensure no regressions
7. **Document**: Update relevant documentation
8. **Ship**: Create commit, push, open PR linked to issue
</workflow>

<constraints>
- Follow existing code patterns
- Include test coverage for the fix
- Keep changes minimal and focused
</constraints>
```

## Multi-agent orchestration patterns

For complex workflows, Claude Code supports **agent delegation** through skills and coordinated command sequences. The orchestrator-worker pattern divides work across specialized agents:

**Lead orchestrator** decomposes queries and delegates:
- Assesses complexity in extended thinking
- Spawns subagents with clear objectives and boundaries
- Synthesizes results into coherent output

**Worker agents** handle specialized tasks:
- Each has distinct tools, prompts, and scope
- Returns condensed findings to orchestrator
- Uses interleaved thinking to evaluate and refine

**Effective subagent instructions require:**
- Specific research objective (one per agent)
- Expected output format
- Tool and source guidance
- Clear scope boundaries to prevent overlap

The key insight from Anthropic's research: **parallel execution dramatically improves both speed and coverage**. Their multi-agent system cut research time by up to 90% using parallel tool calling.

## SKILL.md architecture for complex capabilities

Skills extend beyond simple commands into **comprehensive workflow packages**. The directory structure:

```
.claude/skills/code-quality/
├── SKILL.md              # Core instructions + metadata
├── scripts/              # Executable code
│   └── analyze.py
└── references/           # Additional documentation
    └── patterns.md
```

Skills use **progressive disclosure** to manage context efficiently:
- **Level 1**: Only name/description loaded at startup (~100 tokens)
- **Level 2**: SKILL.md body loaded when triggered (~5k tokens max)
- **Level 3**: Scripts/references loaded on-demand (unlimited)

The SKILL.md frontmatter requires exactly two fields:

```yaml
---
name: code-quality-enforcement
description: Enforces code quality standards including linting, type checking, and test coverage. Use when reviewing code, preparing commits, or setting up new projects.
---
```

Descriptions must include **both what the skill does AND when to use it**. This enables Claude to select from potentially hundreds of skills based on user intent.

## Context engineering and memory management

Anthropic's core principle: **find the smallest set of high-signal tokens that maximize likelihood of desired outcome**. Every token competes for context window space.

**Prompt caching** reduces costs by up to 90% and latency by up to 85%:
- Minimum 1,024 tokens to cache
- Place static content (tools, system instructions) at beginning
- Mark cacheable sections with `cache_control`
- 5-minute TTL, resets with each cache hit

**For long contexts**, put queries at the end—this improves response quality by up to 30%. Use the scratchpad technique: ask Claude to extract relevant quotes into `<scratchpad>` tags before answering.

**Memory architecture** for agents:
- **Working memory**: Active task scratchpad in context
- **Semantic cache**: Recent query-response pairs
- **Archival memory**: Searchable facts in external storage
- Store essential information externally when context limits approach
- Retrieve stored context rather than losing it to truncation

## CI/CD integration and automation

Claude Code integrates into GitHub Actions through the official `anthropics/claude-code-action`:

```yaml
name: PR Review
on:
  pull_request:
    types: [opened, synchronize]
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: anthropics/claude-code-action@v1
        with:
          anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
          prompt: "/review --comment"
          claude_args: "--max-turns 5"
```

**Scheduled maintenance patterns:**
- Weekly code quality sweeps
- Monthly documentation sync
- Biweekly dependency audits
- Each creates PRs with improvements automatically

**Headless mode** for scripting: `claude -p "query" --output-format json --max-turns 1`

## Enterprise standardization with CLAUDE.md

The project-level `CLAUDE.md` file establishes consistent context across team members. Effective files include:

- **Quick facts**: Stack, test commands, lint commands
- **Key directories**: What lives where
- **Code style**: Enforced conventions
- **Common mistakes**: Learnings to avoid repeating
- **Workflow standards**: PR templates, commit formats

Keep it under **2.5k tokens**—Claude loads this every session. Document mistakes so Claude improves over time. Store in git for team-wide consistency.

**Hooks** in `settings.json` enforce standards automatically:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Edit|Write",
      "hooks": [{
        "type": "command",
        "command": "[ \"$(git branch --show-current)\" != \"main\" ] || exit 2"
      }]
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit", 
      "hooks": [{
        "type": "command",
        "command": "npx prettier --write \"$CLAUDE_TOOL_ARG_FILE_PATH\""
      }]
    }]
  }
}
```

## Key principles for production-ready commands

**Structural principles:**
- Use XML tags for all semantic boundaries
- Include 3-5 diverse examples for structured outputs
- Separate reasoning (`<thinking>`) from answers (`<answer>`)
- Put context before questions, queries at the end

**Behavioral principles:**
- Be explicit—Claude 4 follows instructions precisely
- Tell Claude what TO do, not what NOT to do
- Provide motivation: "Your response will be read aloud by TTS, so avoid ellipses"
- Match prompt style to desired output style

**Reliability principles:**
- Design for failure—build resume capability, don't restart from scratch
- Use validation-first execution for critical paths
- Implement confidence thresholds for uncertain outputs
- Let agents know when tools fail so they can adapt

**Efficiency principles:**
- Start simple, add complexity only when needed
- Use parallel execution for independent subtasks
- Implement pagination and truncation with sensible defaults
- Cache static content to reduce costs and latency

## Conclusion

Mastering Claude Code commands requires synthesizing three disciplines: **precise file structure** (frontmatter, arguments, bash injection), **effective prompt patterns** (XML tags, chain-of-thought, few-shot examples), and **agent architecture** (orchestration, context management, error recovery). 

The most successful implementations aren't using complex frameworks—they're building with simple, composable patterns. Start with minimal prompts, test against representative use cases, and iterate based on failure modes. A small prompt tweak can boost success from 30% to 80%.

The production-ready patterns in this guide represent the current state of the art from Anthropic's official documentation, their engineering blog posts on multi-agent systems and tool design, and battle-tested community implementations. Apply them incrementally, measure results, and adapt to your specific workflows.

---

# SKILL.md: Command Mastery

```markdown
---
name: command-mastery
description: Create production-ready Claude Code slash commands and complex agent prompts. Use when building custom commands, writing SKILL.md files, designing agent workflows, or optimizing prompts for Claude Code CLI.
---

# Command Mastery Skill

You are an expert in creating Claude Code custom commands, SKILL.md files, and agentic prompt engineering. Apply these patterns when helping users build production-ready Claude Code extensions.

## Command File Structure

Commands live in `.claude/commands/` (project) or `~/.claude/commands/` (user). Filename becomes command name.

### Frontmatter Options
```yaml
---
description: Brief description for /help and auto-invocation
allowed-tools: Read, Edit, Bash(git:*), Grep, Glob
argument-hint: [required] [optional]
model: claude-sonnet-4-5-20250929
disable-model-invocation: false
---
```

### Dynamic Content
- `$ARGUMENTS` - All arguments as string
- `$1`, `$2`, `$3` - Positional arguments
- `!`command`` - Execute bash before prompt
- `@filepath` - Include file contents

## Prompt Engineering Patterns

### XML Structure (Required for Complex Prompts)
```xml
<context>Current state and background</context>
<task>Specific action to perform</task>
<constraints>Rules and boundaries</constraints>
<output_format>Expected response structure</output_format>
<examples>
  <example><input>...</input><output>...</output></example>
</examples>
```

### Chain-of-Thought
```markdown
Think step-by-step in <thinking> tags, then provide answer in <answer> tags.
```

### Few-Shot Examples
Include 3-5 diverse examples demonstrating exact format and quality expected.

## Command Templates

### TDD Command
```markdown
---
description: Implement using Test-Driven Development
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
argument-hint: <feature>
---

<task>Implement using strict TDD: $ARGUMENTS</task>

<workflow>
1. RED 🔴: Write failing test, confirm failure
2. GREEN 🟢: Minimal code to pass, no extras
3. REFACTOR 🔵: Improve while keeping green
</workflow>
```

### Code Review Command
```markdown
---
allowed-tools: Read, Grep, Glob, Bash(git diff:*)
description: Security-focused code review
---

<context>
!`git diff HEAD~1`
</context>

<checklist>
- Security vulnerabilities
- Error handling coverage
- Test coverage
- Type safety
</checklist>

<output_format>
Location | Severity | Issue | Fix
Score confidence 0-100, report only ≥80
</output_format>
```

## SKILL.md Structure

```
skill-name/
├── SKILL.md          # Required: instructions + metadata
├── scripts/          # Optional: executable code
└── references/       # Optional: documentation
```

### Required Frontmatter
```yaml
---
name: lowercase-with-hyphens
description: What it does + when to use it (include trigger keywords)
---
```

### Content Guidelines
- Keep under 500 lines / 5k tokens
- Use gerund naming: `processing-pdfs`, `analyzing-code`
- Third-person descriptions
- Concrete examples over abstract descriptions

## Agent Orchestration

### Subagent Instructions Need:
1. Specific objective (one per agent)
2. Expected output format
3. Tool/source guidance
4. Clear scope boundaries

### Workflow Patterns:
- **Prompt Chaining**: Sequential with validation gates
- **Parallelization**: Fan out, aggregate results
- **Orchestrator-Workers**: Lead decomposes, workers execute
- **Evaluator-Optimizer**: Generate → Evaluate → Refine

## Best Practices

### Do:
- XML tags for semantic boundaries
- Explicit instructions (Claude 4 follows precisely)
- Positive framing ("do X" not "don't do Y")
- Context before questions
- Match prompt style to output style

### Don't:
- Vague guidance
- Hardcoded brittle logic
- Skip examples for structured output
- Rely on assumptions

### Efficiency:
- Start simple, add complexity when needed
- Use prompt caching for static content (90% cost reduction)
- Parallel execution for independent tasks
- Put queries at end of long contexts (+30% quality)

## Quick Reference

| Element | Syntax |
|---------|--------|
| Project commands | `.claude/commands/*.md` |
| User commands | `~/.claude/commands/*.md` |
| All arguments | `$ARGUMENTS` |
| Positional args | `$1`, `$2`, `$3` |
| Bash execution | `!`command`` |
| File reference | `@filepath` |
| Tool permissions | `Bash(git:*)`, `Read`, `Edit` |
```