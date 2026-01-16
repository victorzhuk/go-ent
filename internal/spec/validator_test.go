package spec

//nolint:gosec // test file with necessary file operations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRequirementHasScenario(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantIssues int
	}{
		{
			name: "requirement with scenario",
			content: `### Requirement: User can login

#### Scenario: Valid credentials
- WHEN user enters valid credentials
- THEN user is logged in
`,
			wantIssues: 0,
		},
		{
			name: "requirement without scenario",
			content: `### Requirement: User can login

Some description here but no scenario.
`,
			wantIssues: 1,
		},
		{
			name: "multiple requirements with one missing scenario",
			content: `### Requirement: User can login

#### Scenario: Valid credentials
- WHEN user enters valid credentials
- THEN user is logged in

### Requirement: User can logout

No scenario here.
`,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
			}

			issues := validateRequirementHasScenario(ctx)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestValidateScenarioFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantIssues int
	}{
		{
			name: "correct scenario format",
			content: `#### Scenario: Valid input
- WHEN user enters valid data
- THEN data is accepted
`,
			wantIssues: 0,
		},
		{
			name: "bullet with bold scenario",
			content: `- **Scenario: Valid input**
- WHEN user enters valid data
`,
			wantIssues: 1,
		},
		{
			name: "wrong header level (3 hashtags)",
			content: `### Scenario: Valid input
- WHEN user enters valid data
`,
			wantIssues: 1,
		},
		{
			name: "wrong header level (5 hashtags)",
			content: `##### Scenario: Valid input
- WHEN user enters valid data
`,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
			}

			issues := validateScenarioFormat(ctx)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestValidateDeltaOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantIssues int
	}{
		{
			name: "valid ADDED operation",
			content: `## ADDED Requirements

### Requirement: New feature
`,
			wantIssues: 0,
		},
		{
			name: "valid MODIFIED operation",
			content: `## MODIFIED Requirements

### Requirement: Updated feature
`,
			wantIssues: 0,
		},
		{
			name: "invalid operation",
			content: `## UPDATED Requirements

### Requirement: Updated feature
`,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
			}

			issues := validateDeltaOperations(ctx)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestValidateRequirementFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantIssues int
	}{
		{
			name: "correct requirement format",
			content: `### Requirement: User can login
`,
			wantIssues: 0,
		},
		{
			name: "bullet with bold requirement",
			content: `- **Requirement: User can login**
`,
			wantIssues: 1,
		},
		{
			name: "wrong header level (2 hashtags)",
			content: `## Requirement: User can login
`,
			wantIssues: 1,
		},
		{
			name: "wrong header level (4 hashtags)",
			content: `#### Requirement: User can login
`,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
			}

			issues := validateRequirementFormat(ctx)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestValidateSpec(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	validSpec := `# Test Spec

### Requirement: User authentication

#### Scenario: Login with valid credentials
- WHEN user enters valid username and password
- THEN user is authenticated
`

	invalidSpec := `# Test Spec

### Requirement: User authentication

No scenarios here.
`

	validPath := filepath.Join(tmpDir, "valid.md")
	require.NoError(t, os.WriteFile(validPath, []byte(validSpec), 0600))

	invalidPath := filepath.Join(tmpDir, "invalid.md")
	require.NoError(t, os.WriteFile(invalidPath, []byte(invalidSpec), 0600))

	validator := NewValidator()

	t.Run("valid spec", func(t *testing.T) {
		result, err := validator.ValidateSpec(validPath, false)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Issues)
	})

	t.Run("invalid spec", func(t *testing.T) {
		result, err := validator.ValidateSpec(invalidPath, false)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Issues)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := validator.ValidateSpec("/nonexistent.md", false)
		assert.Error(t, err)
	})
}

func TestValidateChange(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create valid change structure
	validChange := filepath.Join(tmpDir, "valid-change")
	require.NoError(t, os.MkdirAll(filepath.Join(validChange, "specs"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(validChange, "proposal.md"), []byte("# Proposal"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(validChange, "tasks.md"), []byte("# Tasks"), 0600))

	deltaSpec := `## ADDED Requirements

### Requirement: New feature

#### Scenario: Basic usage
- WHEN feature is used
- THEN it works
`
	require.NoError(t, os.WriteFile(filepath.Join(validChange, "specs", "feature.md"), []byte(deltaSpec), 0600))

	// Create invalid change (missing proposal.md)
	invalidChange := filepath.Join(tmpDir, "invalid-change")
	require.NoError(t, os.MkdirAll(filepath.Join(invalidChange, "specs"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(invalidChange, "tasks.md"), []byte("# Tasks"), 0600))

	validator := NewValidator()

	t.Run("valid change", func(t *testing.T) {
		result, err := validator.ValidateChange(validChange, false)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("missing proposal.md", func(t *testing.T) {
		result, err := validator.ValidateChange(invalidChange, false)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Greater(t, len(result.Issues), 0)

		// Check that at least one issue is about missing proposal.md
		hasProposalIssue := false
		for _, issue := range result.Issues {
			if issue.RuleID == "change-structure" && issue.Message == "Missing proposal.md" {
				hasProposalIssue = true
				break
			}
		}
		assert.True(t, hasProposalIssue)
	})
}

func TestValidationResult(t *testing.T) {
	t.Parallel()

	result := &ValidationResult{
		Issues: []ValidationIssue{
			{Severity: SeverityError, Message: "Error 1"},
			{Severity: SeverityError, Message: "Error 2"},
			{Severity: SeverityWarning, Message: "Warning 1"},
		},
	}

	assert.Equal(t, 2, result.ErrorCount())
	assert.Equal(t, 1, result.WarningCount())
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
