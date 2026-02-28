package workspace

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceDB(t *testing.T) {
	t.Run("open and close", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		db, err := OpenDB("test-ws")
		require.NoError(t, err)
		require.NotNil(t, db)
		require.NoError(t, db.Close())
	})

	t.Run("put project", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		db, err := OpenDB("test-ws")
		require.NoError(t, err)
		defer func() { require.NoError(t, db.Close()) }()

		meta := &ProjectMeta{
			Name:        "app-api",
			Path:        "/path/to/api",
			SpecCount:   5,
			ChangeCount: 2,
		}

		require.NoError(t, db.PutProject(meta))
	})

	t.Run("put spec", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		db, err := OpenDB("test-ws")
		require.NoError(t, err)
		defer func() { require.NoError(t, db.Close()) }()

		require.NoError(t, db.PutSpec(&SpecMeta{
			ID:      "auth",
			Project: "app-api",
			Title:   "Authentication",
		}))
	})
}

func TestWorkspaceDBPath(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")
	path := WorkspaceDBPath("my-ws")
	assert.Equal(t, filepath.Join("/tmp/test-cache", "go-ent", "workspaces", "my-ws", "workspace.db"), path)
}
