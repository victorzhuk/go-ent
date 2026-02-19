## XML Structure for Agent Prompts
```xml
<agent_identity>
You are a specialized Go backend developer agent...
</agent_identity>

<capabilities>
- Read and analyze Go source code
- Implement features following clean architecture
- Write comprehensive tests
</capabilities>

<constraints>
- Never modify files outside the assigned scope
- Always run tests after changes
- Follow existing code conventions
</constraints>

<workflow>
1. Understand the task requirements
2. Analyze existing code structure
3. Plan implementation approach
4. Implement changes incrementally
5. Write/update tests
6. Verify all tests pass
</workflow>

<examples>
<example>
<task>Add input validation to CreateUser endpoint</task>
<approach>
1. Check existing validation patterns
2. Define validation rules
3. Implement at handler boundary
4. Add test cases for validation
</approach>
</example>
</examples>
```

## Chain-of-Thought Prompting
- Ask agents to "think step by step" for complex reasoning
- Use `<thinking>` tags to separate reasoning from output
- Break complex problems into explicit substeps
- Show the reasoning process, not just the answer

## Few-Shot Patterns
- Provide 2-3 examples of input → output
- Examples should cover different cases (happy path, edge case, error)
- Format examples consistently with the expected output format
- Place examples near the end of the prompt, before the actual task

## Context Engineering
- Front-load the most important context
- Use summaries for large codebases — don't dump entire files
- Reference specific files/functions by path
- Provide relevant error messages and stack traces verbatim
- Include dependency versions and environment details

## Anti-Patterns to Avoid
- Vague instructions ("make it better")
- Contradictory constraints
- Too many instructions (cognitive overload)
- Assuming context the agent doesn't have
- Not specifying output format
- Over-reliance on system prompts without examples

## Evaluation & Iteration
- Test prompts with diverse inputs
- Measure output quality consistently
- A/B test prompt variations
- Monitor failure modes and add guardrails
- Version control your prompts alongside code
