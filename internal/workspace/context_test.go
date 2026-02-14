package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateContextPrompt(t *testing.T) {
	t.Run("returns empty for nil workspace", func(t *testing.T) {
		t.Parallel()
		content, err := GenerateContextPrompt(nil)
		assert.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("includes workspace name and projects", func(t *testing.T) {
		t.Parallel()
		ws := &Workspace{
			Name: "enterprise-apps",
			Path: t.TempDir(),
			Projects: []ProjectRef{
				{Name: "app-api", Description: "REST API"},
				{Name: "app-worker", Description: "Background jobs"},
			},
		}

		content, err := GenerateContextPrompt(ws)
		require.NoError(t, err)
		assert.Contains(t, content, "enterprise-apps")
		assert.Contains(t, content, "app-api")
		assert.Contains(t, content, "app-worker")
		assert.Contains(t, content, "REST API")
	})

	t.Run("includes workspace specs", func(t *testing.T) {
		wsPath := t.TempDir()
		specDir := filepath.Join(wsPath, "openspec", "specs", "auth")
		require.NoError(t, os.MkdirAll(specDir, 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(specDir, "spec.md"),
			[]byte("# Authentication Architecture"),
			0o600,
		))

		ws := &Workspace{
			Name: "team-ws",
			Path: wsPath,
		}

		content, err := GenerateContextPrompt(ws)
		require.NoError(t, err)
		assert.Contains(t, content, "auth")
		assert.Contains(t, content, "Authentication Architecture")
	})
}

func TestWriteContextPrompt(t *testing.T) {
	t.Run("writes _workspace.md", func(t *testing.T) {
		t.Parallel()
		outDir := filepath.Join(t.TempDir(), "output")

		ws := &Workspace{
			Name: "test-ws",
			Path: t.TempDir(),
			Projects: []ProjectRef{
				{Name: "app-1"},
			},
		}

		err := WriteContextPrompt(outDir, ws)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(outDir, "_workspace.md"))
		require.NoError(t, err)
		assert.True(t, strings.Contains(string(data), "test-ws"))
	})

	t.Run("no-op for nil workspace", func(t *testing.T) {
		t.Parallel()
		outDir := filepath.Join(t.TempDir(), "output")

		err := WriteContextPrompt(outDir, nil)
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(outDir, "_workspace.md"))
		assert.True(t, os.IsNotExist(err))
	})
}
