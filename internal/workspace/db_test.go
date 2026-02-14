package workspace

import (
	"path/filepath"
	"testing"
	"time"

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

	t.Run("project CRUD", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		db, err := OpenDB("test-ws")
		require.NoError(t, err)
		defer db.Close()

		meta := &ProjectMeta{
			Name:        "app-api",
			Path:        "/path/to/api",
			SyncedAt:    time.Now(),
			SpecCount:   5,
			ChangeCount: 2,
		}

		require.NoError(t, db.PutProject(meta))

		projects, err := db.ListProjects()
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, "app-api", projects[0].Name)
		assert.Equal(t, 5, projects[0].SpecCount)
	})

	t.Run("spec indexing", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		db, err := OpenDB("test-ws")
		require.NoError(t, err)
		defer db.Close()

		require.NoError(t, db.PutSpec(&SpecMeta{
			ID:      "auth",
			Project: "app-api",
			Title:   "Authentication",
		}))
		require.NoError(t, db.PutSpec(&SpecMeta{
			ID:      "data-model",
			Project: "app-api",
			Title:   "Data Model",
		}))
		require.NoError(t, db.PutSpec(&SpecMeta{
			ID:      "billing",
			Project: "app-billing",
			Title:   "Billing",
		}))

		specs, err := db.ListSpecs("app-api")
		require.NoError(t, err)
		assert.Len(t, specs, 2)
	})
}

func TestWorkspaceDBPath(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")
	path := WorkspaceDBPath("my-ws")
	assert.Equal(t, filepath.Join("/tmp/test-cache", "go-ent", "workspaces", "my-ws", "workspace.db"), path)
}
