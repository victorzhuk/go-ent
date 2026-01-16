package spec

//nolint:gosec // test file with necessary file operations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDeltaSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		wantAdded    int
		wantModified int
		wantRemoved  int
		wantRenamed  int
	}{
		{
			name: "ADDED requirements",
			content: `## ADDED Requirements

### Requirement: New feature A

#### Scenario: Basic usage
- WHEN user uses feature A
- THEN it works

### Requirement: New feature B

#### Scenario: Basic usage
- WHEN user uses feature B
- THEN it works
`,
			wantAdded: 2,
		},
		{
			name: "MODIFIED requirements",
			content: `## MODIFIED Requirements

### Requirement: Updated feature

#### Scenario: Enhanced usage
- WHEN user uses enhanced feature
- THEN it works better
`,
			wantModified: 1,
		},
		{
			name: "REMOVED requirements",
			content: `## REMOVED Requirements

### Requirement: Deprecated feature

**Reason**: No longer needed
`,
			wantRemoved: 1,
		},
		{
			name: "RENAMED requirements",
			content: `## RENAMED Requirements

- FROM: ` + "`### Requirement: Old name`" + `
- TO: ` + "`### Requirement: New name`" + `
`,
			wantRenamed: 1,
		},
		{
			name: "mixed delta operations",
			content: `## ADDED Requirements

### Requirement: New feature

#### Scenario: Basic usage
- WHEN user uses feature
- THEN it works

## MODIFIED Requirements

### Requirement: Updated feature

#### Scenario: Enhanced usage
- WHEN user uses feature
- THEN it works better

## REMOVED Requirements

### Requirement: Deprecated feature

**Reason**: No longer needed

## RENAMED Requirements

- FROM: ` + "`### Requirement: Old name`" + `
- TO: ` + "`### Requirement: New name`" + `
`,
			wantAdded:    1,
			wantModified: 1,
			wantRemoved:  1,
			wantRenamed:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			delta, err := ParseDeltaSpec(tt.content)
			require.NoError(t, err)
			assert.Len(t, delta.Added, tt.wantAdded)
			assert.Len(t, delta.Modified, tt.wantModified)
			assert.Len(t, delta.Removed, tt.wantRemoved)
			assert.Len(t, delta.Renamed, tt.wantRenamed)
		})
	}
}

func TestMergeDeltas_ADDED(t *testing.T) {
	t.Parallel()

	baseSpec := `# Test Spec

### Requirement: Existing feature

#### Scenario: Basic usage
- WHEN feature is used
- THEN it works
`

	delta := &DeltaSpec{
		Added: []Requirement{
			{
				Name: "New feature",
				Content: `### Requirement: New feature

#### Scenario: New usage
- WHEN new feature is used
- THEN it works`,
			},
		},
	}

	result, err := MergeDeltas(baseSpec, delta)
	require.NoError(t, err)

	// Check that original requirement still exists
	assert.Contains(t, result, "Requirement: Existing feature")

	// Check that new requirement was added
	assert.Contains(t, result, "Requirement: New feature")
}

func TestMergeDeltas_MODIFIED(t *testing.T) {
	t.Parallel()

	baseSpec := `# Test Spec

### Requirement: Feature to modify

#### Scenario: Old scenario
- WHEN old behavior
- THEN old result

### Requirement: Other feature

#### Scenario: Should not change
- WHEN other feature is used
- THEN it works
`

	delta := &DeltaSpec{
		Modified: []Requirement{
			{
				Name: "Feature to modify",
				Content: `### Requirement: Feature to modify

#### Scenario: New scenario
- WHEN new behavior
- THEN new result`,
			},
		},
	}

	result, err := MergeDeltas(baseSpec, delta)
	require.NoError(t, err)

	// Check that modified requirement has new content
	assert.Contains(t, result, "New scenario")
	assert.NotContains(t, result, "Old scenario")

	// Check that other requirement is unchanged
	assert.Contains(t, result, "Other feature")
	assert.Contains(t, result, "Should not change")
}

func TestMergeDeltas_REMOVED(t *testing.T) {
	t.Parallel()

	baseSpec := `# Test Spec

### Requirement: Keep this

#### Scenario: Should stay
- WHEN used
- THEN works

### Requirement: Remove this

#### Scenario: Should be gone
- WHEN used
- THEN works

### Requirement: Also keep this

#### Scenario: Should stay
- WHEN used
- THEN works
`

	delta := &DeltaSpec{
		Removed: []RemovedRequirement{
			{
				Name:   "Remove this",
				Reason: "No longer needed",
			},
		},
	}

	result, err := MergeDeltas(baseSpec, delta)
	require.NoError(t, err)

	// Check that removed requirement is gone
	assert.NotContains(t, result, "Remove this")
	assert.NotContains(t, result, "Should be gone")

	// Check that other requirements still exist
	assert.Contains(t, result, "Keep this")
	assert.Contains(t, result, "Also keep this")
}

func TestMergeDeltas_RENAMED(t *testing.T) {
	t.Parallel()

	baseSpec := `# Test Spec

### Requirement: Old name

#### Scenario: Usage
- WHEN used
- THEN works
`

	delta := &DeltaSpec{
		Renamed: []RenamedRequirement{
			{
				FromName: "Old name",
				ToName:   "New name",
			},
		},
	}

	result, err := MergeDeltas(baseSpec, delta)
	require.NoError(t, err)

	// Check that requirement was renamed
	assert.NotContains(t, result, "Old name")
	assert.Contains(t, result, "New name")

	// Check that content is preserved
	assert.Contains(t, result, "Usage")
}

func TestMergeDeltas_Complex(t *testing.T) {
	t.Parallel()

	baseSpec := `# Test Spec

### Requirement: To rename

#### Scenario: Original
- WHEN used
- THEN works

### Requirement: To modify

#### Scenario: Old
- WHEN old
- THEN old

### Requirement: To remove

#### Scenario: Gone
- WHEN gone
- THEN gone

### Requirement: To keep

#### Scenario: Stay
- WHEN stay
- THEN stay
`

	delta := &DeltaSpec{
		Renamed: []RenamedRequirement{
			{FromName: "To rename", ToName: "Renamed requirement"},
		},
		Modified: []Requirement{
			{
				Name: "To modify",
				Content: `### Requirement: To modify

#### Scenario: New
- WHEN new
- THEN new`,
			},
		},
		Removed: []RemovedRequirement{
			{Name: "To remove", Reason: "Deprecated"},
		},
		Added: []Requirement{
			{
				Name: "New requirement",
				Content: `### Requirement: New requirement

#### Scenario: Fresh
- WHEN fresh
- THEN fresh`,
			},
		},
	}

	result, err := MergeDeltas(baseSpec, delta)
	require.NoError(t, err)

	// Check renames
	assert.NotContains(t, result, "To rename")
	assert.Contains(t, result, "Renamed requirement")

	// Check modifications
	assert.Contains(t, result, "To modify")
	assert.Contains(t, result, "New")
	assert.NotContains(t, result, "Old")

	// Check removals
	assert.NotContains(t, result, "To remove")
	assert.NotContains(t, result, "Gone")

	// Check additions
	assert.Contains(t, result, "New requirement")
	assert.Contains(t, result, "Fresh")

	// Check kept requirement
	assert.Contains(t, result, "To keep")
	assert.Contains(t, result, "Stay")
}

func TestRenameRequirement(t *testing.T) {
	t.Parallel()

	content := `### Requirement: Original name

Some content here
`

	result := renameRequirement(content, "Original name", "New name")
	assert.Contains(t, result, "### Requirement: New name")
	assert.NotContains(t, result, "Original name")
	assert.Contains(t, result, "Some content here")
}

func TestRemoveRequirement(t *testing.T) {
	t.Parallel()

	content := `### Requirement: Keep this

Content 1

### Requirement: Remove this

Content 2

### Requirement: Also keep

Content 3
`

	result := removeRequirement(content, "Remove this", "No longer needed")

	assert.NotContains(t, result, "Remove this")
	assert.NotContains(t, result, "Content 2")
	assert.Contains(t, result, "Keep this")
	assert.Contains(t, result, "Also keep")
	assert.Contains(t, result, "Content 1")
	assert.Contains(t, result, "Content 3")
}

func TestAppendRequirement(t *testing.T) {
	t.Parallel()

	content := "# Spec\n\n### Requirement: Existing\n"
	newReq := "### Requirement: New\n\nContent"

	result := appendRequirement(content, newReq)

	lines := strings.Split(result, "\n")

	// Find the new requirement
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Requirement: New") {
			found = true
			break
		}
	}
	assert.True(t, found, "New requirement should be appended")

	// Original content should still be present
	assert.Contains(t, result, "Requirement: Existing")
}
