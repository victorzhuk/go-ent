package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectWorkspace(t *testing.T) {
	t.Run("returns empty when no workspace ref", func(t *testing.T) {
		t.Parallel()
		name, err := DetectWorkspace(t.TempDir())
		assert.NoError(t, err)
		assert.Empty(t, name)
	})

	t.Run("reads workspace name from ref file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o750))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, ".claude", "workspace.yaml"),
			[]byte("workspace: my-workspace\n"),
			0o600,
		))

		name, err := DetectWorkspace(dir)
		require.NoError(t, err)
		assert.Equal(t, "my-workspace", name)
	})
}

func TestWorkspaceRegistry(t *testing.T) {
	t.Run("load returns empty when file missing", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		reg, err := LoadRegistry()
		require.NoError(t, err)
		assert.NotNil(t, reg.Workspaces)
		assert.Empty(t, reg.Workspaces)
	})

	t.Run("save and load roundtrip", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		reg := &WorkspaceRegistry{
			Workspaces: map[string]string{
				"team-apps": "/home/user/team",
			},
		}

		require.NoError(t, SaveRegistry(reg))

		loaded, err := LoadRegistry()
		require.NoError(t, err)
		assert.Equal(t, "/home/user/team", loaded.Workspaces["team-apps"])
	})
}

func TestProjectsRegistry(t *testing.T) {
	t.Run("load returns empty when file missing", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())

		reg, err := LoadProjectsRegistry("test-ws")
		require.NoError(t, err)
		assert.Empty(t, reg.Projects)
	})

	t.Run("save and load roundtrip", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())

		reg := &ProjectsRegistry{
			Projects: []ProjectRef{
				{Name: "app-api", Path: "/path/to/api"},
				{Name: "app-worker", Path: "/path/to/worker"},
			},
		}

		require.NoError(t, SaveProjectsRegistry("test-ws", reg))

		loaded, err := LoadProjectsRegistry("test-ws")
		require.NoError(t, err)
		assert.Len(t, loaded.Projects, 2)
		assert.Equal(t, "app-api", loaded.Projects[0].Name)
	})
}

func TestWorkspaceConfig(t *testing.T) {
	t.Run("load returns default when missing", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())

		cfg, err := LoadWorkspaceConfig("test-ws")
		require.NoError(t, err)
		assert.Equal(t, "1.0", cfg.Version)
	})

	t.Run("save and load roundtrip", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())

		cfg := &WorkspaceConfig{
			Version: "1.0",
			Models: map[string]string{
				"fast": "haiku",
			},
		}

		require.NoError(t, SaveWorkspaceConfig("test-ws", cfg))

		loaded, err := LoadWorkspaceConfig("test-ws")
		require.NoError(t, err)
		assert.Equal(t, "haiku", loaded.Models["fast"])
	})
}

func TestSkillsDirs(t *testing.T) {
	t.Run("returns nil for nil workspace", func(t *testing.T) {
		t.Parallel()
		dirs := SkillsDirs(nil)
		assert.Nil(t, dirs)
	})

	t.Run("returns skills dir when exists", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "skills"), 0o750))

		ws := &Workspace{Path: dir}
		dirs := SkillsDirs(ws)
		assert.Len(t, dirs, 1)
		assert.Equal(t, filepath.Join(dir, "skills"), dirs[0])
	})

	t.Run("returns nil when skills dir missing", func(t *testing.T) {
		t.Parallel()
		ws := &Workspace{Path: t.TempDir()}
		dirs := SkillsDirs(ws)
		assert.Nil(t, dirs)
	})
}

func TestWriteWorkspaceRef(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	err := WriteWorkspaceRef(dir, "my-workspace")
	require.NoError(t, err)

	name, err := DetectWorkspace(dir)
	require.NoError(t, err)
	assert.Equal(t, "my-workspace", name)
}
