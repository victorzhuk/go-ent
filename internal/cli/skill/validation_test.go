package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGeneratedSkill_ValidSkill(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	validSkill := `---
name: test-skill
description: A valid test skill
version: "1.0.0"
author: Test Author
tags: ["test", "validation"]
structure_version: "v2"
---

# Test Skill

<role>
You are a test skill expert.
</role>

<instructions>
Provide helpful testing guidance.
</instructions>

<constraints>
- Follow best practices
- Write clear code
</constraints>

<edge_cases>
If test fails: investigate root cause
</edge_cases>

<examples>
<example>
<input>Test input</input>
<output>Test output</output>
</example>
</examples>

<output_format>
Provide clear, actionable guidance.
</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(validSkill), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.NoError(t, err)
}

func TestValidateGeneratedSkill_FileNotFound(t *testing.T) {
	t.Parallel()

	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.md")

	err := ValidateGeneratedSkill(nonExistentPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestValidateGeneratedSkill_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	invalidYAML := `---
name: test
description: test
invalid: yaml: content:
version: "1.0.0"
---
<role>Test</role>
<instructions>Test</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(invalidYAML), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse skill file")
}

func TestValidateGeneratedSkill_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	require.NoError(t, os.WriteFile(skillPath, []byte(""), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_MissingFrontmatter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	noFrontmatter := `# Test Skill

<role>Test role</role>
<instructions>Test instructions</instructions>
<constraints>- Test constraint</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(noFrontmatter), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_MissingName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	missingName := `---
description: A test skill
version: "1.0.0"
---
<role>Test</role>
<instructions>Test</instructions>
<constraints>- Test</constraints>
<examples>
<example>
<input>test</input>
<output>test</output>
</example>
</examples>
<output_format>Test</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(missingName), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.Error(t, err)
}

func TestValidateGeneratedSkill_CompleteValidSkill(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	completeSkill := `---
name: go-database
description: Database integration patterns with PostgreSQL and pgx
version: "1.0.0"
author: go-ent
tags: ["go", "database", "postgresql", "pgx"]
structure_version: "v2"
---

# Go Database Patterns

<role>
You are an expert Go developer specializing in database integration patterns.
You have extensive experience with PostgreSQL, pgx, and database transaction management.
You follow best practices for connection pooling, query optimization, and error handling.
</role>

<instructions>
Provide guidance on Go database integration including:

1. **Connection Management**
   - Set up connection pools with appropriate limits
   - Configure health checks and timeouts
   - Handle connection lifecycle properly

2. **Query Patterns**
   - Use parameterized queries to prevent SQL injection
   - Implement batch operations for efficiency
   - Use prepared statements for repeated queries

3. **Transaction Management**
   - Handle transactions with proper rollback
   - Implement retry logic for transient failures
   - Use context for cancellation

4. **Error Handling**
   - Wrap database errors with context
   - Handle specific pgx error types
   - Provide meaningful error messages

5. **Performance**
   - Optimize queries with indexes
   - Use connection pooling effectively
   - Implement caching where appropriate
</instructions>

<constraints>
- Always use context.Context for database operations
- Never concatenate strings to build SQL queries
- Always check for errors after database operations
- Close rows objects to prevent connection leaks
- Use tx.Commit() only after all operations succeed
- Implement proper error wrapping with context
</constraints>

<edge_cases>
If connection pool is exhausted: log error and implement retry logic with exponential backoff
If transaction fails due to serialization error: retry transaction up to 3 times
If database is unavailable during startup: implement circuit breaker pattern
</edge_cases>

<examples>
<example>
<input>How do I set up a connection pool with pgx?</input>
<output>
Provide a code example showing proper connection pool setup with pgx.
</output>
</example>

<example>
<input>How do I handle transactions properly?</input>
<output>
Provide a code example showing proper transaction handling with rollback.
</output>
</example>
</examples>

<output_format>
Provide production-ready Go code with:
- Complete, runnable examples
- Proper error handling with context
- Usage of pgx best practices
- Comments explaining key decisions
- Type safety with struct mapping
</output_format>
`

	require.NoError(t, os.WriteFile(skillPath, []byte(completeSkill), 0644))

	err := ValidateGeneratedSkill(skillPath)
	assert.NoError(t, err)
}
