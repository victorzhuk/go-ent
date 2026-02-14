package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncProject(t *testing.T) {
	t.Run("indexes specs and changes", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))

		wsPath := filepath.Join(tmp, "workspace")
		projectPath := filepath.Join(tmp, "project")

		require.NoError(t, os.MkdirAll(filepath.Join(wsPath, "openspec", "specs"), 0o750))
		require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "openspec", "specs", "auth"), 0o750))
		require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "openspec", "specs", "data"), 0o750))
		require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "openspec", "changes", "add-feature"), 0o750))

		require.NoError(t, os.WriteFile(
			filepath.Join(projectPath, "openspec", "specs", "auth", "spec.md"),
			[]byte("# Authentication Spec"),
			0o600,
		))

		ws := &Workspace{Name: "test-ws", Path: wsPath}
		project := ProjectRef{Name: "my-app", Path: projectPath}

		specs, changes, pulled, err := SyncProject(ws, project)
		require.NoError(t, err)
		assert.Equal(t, 2, specs)
		assert.Equal(t, 1, changes)
		assert.Equal(t, 0, pulled)
	})

	t.Run("pulls workspace specs with ws prefix", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))

		wsPath := filepath.Join(tmp, "workspace")
		projectPath := filepath.Join(tmp, "project")

		wsSpecDir := filepath.Join(wsPath, "openspec", "specs", "shared-model")
		require.NoError(t, os.MkdirAll(wsSpecDir, 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(wsSpecDir, "spec.md"),
			[]byte("# Shared Data Model"),
			0o600,
		))

		require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "openspec", "specs"), 0o750))

		ws := &Workspace{Name: "test-ws", Path: wsPath}
		project := ProjectRef{Name: "my-app", Path: projectPath}

		_, _, pulled, err := SyncProject(ws, project)
		require.NoError(t, err)
		assert.Equal(t, 1, pulled)

		data, err := os.ReadFile(filepath.Join(projectPath, "openspec", "specs", "ws-shared-model", "spec.md"))
		require.NoError(t, err)
		assert.Equal(t, "# Shared Data Model", string(data))
	})

	t.Run("handles missing openspec dirs", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))

		ws := &Workspace{Name: "test-ws", Path: filepath.Join(tmp, "workspace")}
		project := ProjectRef{Name: "my-app", Path: filepath.Join(tmp, "project")}

		specs, changes, pulled, err := SyncProject(ws, project)
		require.NoError(t, err)
		assert.Equal(t, 0, specs)
		assert.Equal(t, 0, changes)
		assert.Equal(t, 0, pulled)
	})
}

func TestExtractTitle(t *testing.T) {
	t.Run("extracts title from markdown", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "spec.md")
		require.NoError(t, os.WriteFile(path, []byte("# My Spec Title\n\nContent here"), 0o600))

		title := ExtractTitle(path)
		assert.Equal(t, "My Spec Title", title)
	})

	t.Run("returns empty for missing file", func(t *testing.T) {
		t.Parallel()
		title := ExtractTitle("/nonexistent/spec.md")
		assert.Empty(t, title)
	})

	t.Run("returns empty for no heading", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "spec.md")
		require.NoError(t, os.WriteFile(path, []byte("No heading here"), 0o600))

		title := ExtractTitle(path)
		assert.Empty(t, title)
	})
}
