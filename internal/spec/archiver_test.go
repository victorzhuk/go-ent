package spec

//nolint:gosec // test file with necessary file operations

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create mock openspec structure
	specPath := filepath.Join(tmpDir, "openspec")
	changesPath := filepath.Join(specPath, "changes")
	archivePath := filepath.Join(changesPath, "archive")
	specsPath := filepath.Join(specPath, "specs")

	require.NoError(t, os.MkdirAll(changesPath, 0750))
	require.NoError(t, os.MkdirAll(archivePath, 0750))
	require.NoError(t, os.MkdirAll(specsPath, 0750))

	// Create a test change
	changeID := "test-feature"
	changePath := filepath.Join(changesPath, changeID)
	require.NoError(t, os.MkdirAll(changePath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "tasks.md"), []byte("# Tasks"), 0600))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	t.Run("archive with skipSpecs", func(t *testing.T) {
		result, err := archiver.Archive(changeID, true, false)
		require.NoError(t, err)
		assert.Equal(t, changeID, result.ChangeID)
		assert.False(t, result.DryRun)
		assert.Empty(t, result.UpdatedSpecs)
		assert.Empty(t, result.Errors)

		// Verify change was moved to archive with date prefix
		today := time.Now().Format("2006-01-02")
		expectedArchive := filepath.Join(archivePath, today+"-"+changeID)
		assert.Equal(t, expectedArchive, result.ArchivePath)

		_, err = os.Stat(expectedArchive)
		assert.NoError(t, err, "Change should be archived")

		_, err = os.Stat(changePath)
		assert.True(t, os.IsNotExist(err), "Original change should be removed")
	})
}

func TestArchive_DryRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create mock openspec structure
	specPath := filepath.Join(tmpDir, "openspec")
	changesPath := filepath.Join(specPath, "changes")

	require.NoError(t, os.MkdirAll(changesPath, 0750))

	// Create a test change
	changeID := "test-feature"
	changePath := filepath.Join(changesPath, changeID)
	require.NoError(t, os.MkdirAll(changePath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal"), 0600))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	result, err := archiver.Archive(changeID, true, true)
	require.NoError(t, err)
	assert.True(t, result.DryRun)

	// Verify change was NOT moved in dry run
	_, err = os.Stat(changePath)
	assert.NoError(t, err, "Original change should still exist in dry run")

	_, err = os.Stat(result.ArchivePath)
	assert.True(t, os.IsNotExist(err), "Archive should not be created in dry run")
}

func TestArchive_NonexistentChange(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openspec")
	require.NoError(t, os.MkdirAll(specPath, 0750))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	_, err := archiver.Archive("nonexistent", false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "change not found")
}

func TestMergeDeltas(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create mock openspec structure
	specPath := filepath.Join(tmpDir, "openspec")
	changesPath := filepath.Join(specPath, "changes")
	specsPath := filepath.Join(specPath, "specs")
	authSpecPath := filepath.Join(specsPath, "auth")

	require.NoError(t, os.MkdirAll(changesPath, 0750))
	require.NoError(t, os.MkdirAll(authSpecPath, 0750))

	// Create base spec
	baseSpec := `# Authentication Specification

### Requirement: User login

#### Scenario: Valid credentials
- WHEN user provides valid credentials
- THEN user is authenticated
`
	require.NoError(t, os.WriteFile(filepath.Join(authSpecPath, "spec.md"), []byte(baseSpec), 0600))

	// Create a change with delta
	changeID := "add-2fa"
	changePath := filepath.Join(changesPath, changeID)
	changeDeltaPath := filepath.Join(changePath, "specs", "auth")
	require.NoError(t, os.MkdirAll(changeDeltaPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "tasks.md"), []byte("# Tasks"), 0600))

	deltaSpec := `## ADDED Requirements

### Requirement: Two-factor authentication

#### Scenario: Enable 2FA
- WHEN user enables 2FA
- THEN user account requires 2FA
`
	require.NoError(t, os.WriteFile(filepath.Join(changeDeltaPath, "spec.md"), []byte(deltaSpec), 0600))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	result, err := archiver.Archive(changeID, false, false)
	require.NoError(t, err)
	assert.Contains(t, result.UpdatedSpecs, "auth")

	// Verify spec was updated
	mergedSpec, err := os.ReadFile(filepath.Join(authSpecPath, "spec.md")) // #nosec G304 -- test file
	require.NoError(t, err)

	mergedContent := string(mergedSpec)
	assert.Contains(t, mergedContent, "User login", "Original requirement should remain")
	assert.Contains(t, mergedContent, "Two-factor authentication", "New requirement should be added")
}

func TestMergeDeltas_NewSpec(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create mock openspec structure
	specPath := filepath.Join(tmpDir, "openspec")
	changesPath := filepath.Join(specPath, "changes")
	specsPath := filepath.Join(specPath, "specs")

	require.NoError(t, os.MkdirAll(changesPath, 0750))
	require.NoError(t, os.MkdirAll(specsPath, 0750))

	// Create a change with delta for a new spec
	changeID := "add-notifications"
	changePath := filepath.Join(changesPath, changeID)
	changeDeltaPath := filepath.Join(changePath, "specs", "notifications")
	require.NoError(t, os.MkdirAll(changeDeltaPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "tasks.md"), []byte("# Tasks"), 0600))

	deltaSpec := `## ADDED Requirements

### Requirement: Email notifications

#### Scenario: Send email
- WHEN event occurs
- THEN email is sent
`
	require.NoError(t, os.WriteFile(filepath.Join(changeDeltaPath, "spec.md"), []byte(deltaSpec), 0600))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	result, err := archiver.Archive(changeID, false, false)
	require.NoError(t, err)
	assert.Contains(t, result.UpdatedSpecs, "notifications")

	// Verify new spec was created
	newSpecPath := filepath.Join(specsPath, "notifications", "spec.md")
	_, err = os.Stat(newSpecPath)
	assert.NoError(t, err, "New spec should be created")

	newSpec, err := os.ReadFile(newSpecPath) // #nosec G304 -- test file
	require.NoError(t, err)
	assert.Contains(t, string(newSpec), "Email notifications")
}

func TestValidateBeforeArchive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	specPath := filepath.Join(tmpDir, "openspec")
	changesPath := filepath.Join(specPath, "changes")

	require.NoError(t, os.MkdirAll(changesPath, 0750))

	// Create a valid change
	changeID := "valid-change"
	changePath := filepath.Join(changesPath, changeID)
	changeDeltaPath := filepath.Join(changePath, "specs", "test")
	require.NoError(t, os.MkdirAll(changeDeltaPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "tasks.md"), []byte("# Tasks"), 0600))

	validDelta := `## ADDED Requirements

### Requirement: Test feature

#### Scenario: Basic usage
- WHEN feature is used
- THEN it works
`
	require.NoError(t, os.WriteFile(filepath.Join(changeDeltaPath, "spec.md"), []byte(validDelta), 0600))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	result, err := archiver.ValidateBeforeArchive(changeID, false)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Issues)
}

func TestMergeDeltas_DryRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create mock openspec structure
	specPath := filepath.Join(tmpDir, "openspec")
	changesPath := filepath.Join(specPath, "changes")
	specsPath := filepath.Join(specPath, "specs")
	authSpecPath := filepath.Join(specsPath, "auth")

	require.NoError(t, os.MkdirAll(changesPath, 0750))
	require.NoError(t, os.MkdirAll(authSpecPath, 0750))

	// Create base spec
	baseSpec := `# Authentication Specification

### Requirement: User login
`
	require.NoError(t, os.WriteFile(filepath.Join(authSpecPath, "spec.md"), []byte(baseSpec), 0600))

	// Create a change with delta
	changeID := "test-change"
	changePath := filepath.Join(changesPath, changeID)
	changeDeltaPath := filepath.Join(changePath, "specs", "auth")
	require.NoError(t, os.MkdirAll(changeDeltaPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(changePath, "tasks.md"), []byte("# Tasks"), 0600))

	deltaSpec := `## ADDED Requirements

### Requirement: New feature
`
	require.NoError(t, os.WriteFile(filepath.Join(changeDeltaPath, "spec.md"), []byte(deltaSpec), 0600))

	store := NewStore(tmpDir)
	archiver := NewArchiver(store)

	result, err := archiver.Archive(changeID, false, true)
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	assert.Contains(t, result.UpdatedSpecs, "auth")

	// Verify spec was NOT actually updated in dry run
	spec, err := os.ReadFile(filepath.Join(authSpecPath, "spec.md")) // #nosec G304 -- test file
	require.NoError(t, err)
	assert.Equal(t, baseSpec, string(spec), "Spec should not be modified in dry run")
}
