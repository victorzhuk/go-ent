# Change: Align Agent Prompts with Constitutional AI Principles

## Status: complete

## Why
Align agent prompts with Anthropic's Constitutional AI principles to improve judgment, reduce over-application of rules, and enable thoughtful escalation. Current prompts may cause agents to rigidly follow rules without considering context or exercising appropriate judgment.

## What Changes
- Create shared judgment guidance (`_judgment.md`) explaining when to exercise judgment over rules
- Create principal hierarchy framework (`_principals.md`) for conflict resolution
- Update handoff guidelines with irreversible action checkpoints
- Include new guidance in all agent prompts
- Add "thoughtful senior developer" behavioral test criteria

## Impact
- Affected code: Agent prompt files in `plugins/go-ent/agents/prompts/`
- No code changes (prompt engineering only)
- Improves agent decision-making quality
- Reduces unnecessary escalations while maintaining safety
- Aligns with Anthropic's best practices for AI behavior
