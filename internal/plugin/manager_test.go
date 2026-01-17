package plugin

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRegistry struct{}

func (m *mockRegistry) RegisterSkill(name, path string) error {
	return nil
}

func (m *mockRegistry) RegisterAgent(name, path string) error {
	return nil
}

func (m *mockRegistry) UnregisterSkill(name string) error {
	return nil
}

func (m *mockRegistry) UnregisterAgent(name string) error {
	return nil
}

type mockMarketplace struct{}

func (m *mockMarketplace) Download(ctx context.Context, name, version string) ([]byte, error) {
	return nil, nil
}

func TestManager_Initialize_LogsParseFailures(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	invalidPluginDir := filepath.Join(tmpDir, "invalid-plugin")
	require.NoError(t, os.Mkdir(invalidPluginDir, 0750))

	invalidManifestPath := filepath.Join(invalidPluginDir, ManifestFile)
	require.NoError(t, os.WriteFile(invalidManifestPath, []byte("invalid yaml"), 0600))

	validPluginDir := filepath.Join(tmpDir, "valid-plugin")
	require.NoError(t, os.Mkdir(validPluginDir, 0750))

	validManifestPath := filepath.Join(validPluginDir, ManifestFile)
	validManifest := `name: valid-plugin
version: 1.0.0
description: A valid plugin
author: Test

skills:
  - name: test-skill
    path: skills/test.md
`
	require.NoError(t, os.WriteFile(validManifestPath, []byte(validManifest), 0600))

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	logger := slog.New(logHandler)

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, logger)

	err := m.Initialize(context.Background())
	assert.NoError(t, err)

	plugins := m.List()
	assert.Len(t, plugins, 1)
	assert.Equal(t, "valid-plugin", plugins[0].Name)
}

func TestManager_NewManager_DefaultLogger(t *testing.T) {
	t.Parallel()

	m := NewManager("test-dir", &mockRegistry{}, &mockMarketplace{}, nil)

	assert.NotNil(t, m.logger)
	assert.Equal(t, slog.Default(), m.logger)
}

func TestManager_NewManager_CustomLogger(t *testing.T) {
	t.Parallel()

	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewManager("test-dir", &mockRegistry{}, &mockMarketplace{}, customLogger)

	assert.Equal(t, customLogger, m.logger)
}

func TestManager_Initialize_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: test-plugin
version: 1.0.0
description: Test plugin
author: Test

skills:
  - name: test-skill
    path: skills/test.md

agents:
  - name: test-agent
    path: agents/test.md
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.enabled["test-plugin"] = true

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	plugins := m.List()
	require.Len(t, plugins, 1)
	assert.Equal(t, "test-plugin", plugins[0].Name)
	assert.Equal(t, "1.0.0", plugins[0].Version)
	assert.True(t, plugins[0].Enabled)

	plugin, err := m.Get("test-plugin")
	require.NoError(t, err)
	assert.True(t, plugin.Enabled)
	assert.True(t, plugin.Installed)
}

func TestManager_Initialize_EmptyDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	plugins := m.List()
	assert.Empty(t, plugins)
}

func TestManager_Initialize_WithEnabledState(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "enabled-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: enabled-plugin
version: 1.0.0
description: Test plugin
author: Test
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	pluginDir2 := filepath.Join(tmpDir, "disabled-plugin")
	require.NoError(t, os.Mkdir(pluginDir2, 0750))

	manifestPath2 := filepath.Join(pluginDir2, ManifestFile)
	manifest2 := `name: disabled-plugin
version: 1.0.0
description: Test plugin
author: Test
`
	require.NoError(t, os.WriteFile(manifestPath2, []byte(manifest2), 0600))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.enabled["enabled-plugin"] = true

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	enabledPlugin, err := m.Get("enabled-plugin")
	require.NoError(t, err)
	assert.True(t, enabledPlugin.Enabled)

	disabledPlugin, err := m.Get("disabled-plugin")
	require.NoError(t, err)
	assert.False(t, disabledPlugin.Enabled)
}

func TestManager_Install_AlreadyInstalled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{Name: "test-plugin", Version: "1.0.0", Description: "Test", Author: "Test"},
		Enabled:  true,
	}

	err := m.Install(context.Background(), "test-plugin", "1.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestManager_Install_InvalidClient(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)

	err := m.Install(context.Background(), "test-plugin", "1.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid marketplace client type")
}

func TestManager_Uninstall_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)

	err := m.Uninstall(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_Uninstall_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: test-plugin
version: 1.0.0
description: Test plugin
author: Test
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest:  Manifest{Name: "test-plugin", Version: "1.0.0", Description: "Test", Author: "Test"},
		RootPath:  pluginDir,
		Enabled:   true,
		Installed: true,
	}
	m.enabled["test-plugin"] = true

	err := m.Uninstall(context.Background(), "test-plugin")
	require.NoError(t, err)

	_, exists := m.plugins["test-plugin"]
	assert.False(t, exists)

	_, enabled := m.enabled["test-plugin"]
	assert.False(t, enabled)

	_, err = os.Stat(pluginDir)
	assert.True(t, os.IsNotExist(err))
}

func TestManager_Uninstall_DisabledPlugin(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest:  Manifest{Name: "test-plugin", Version: "1.0.0", Description: "Test", Author: "Test"},
		RootPath:  pluginDir,
		Enabled:   false,
		Installed: true,
	}

	err := m.Uninstall(context.Background(), "test-plugin")
	require.NoError(t, err)

	_, exists := m.plugins["test-plugin"]
	assert.False(t, exists)
}

func TestManager_Enable_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)

	err := m.Enable(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_Enable_AlreadyEnabled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
		},
		Enabled:   true,
		Installed: true,
	}

	err := m.Enable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.True(t, plugin.Enabled)
}

func TestManager_Enable_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	skillsDir := filepath.Join(pluginDir, "skills")
	require.NoError(t, os.Mkdir(skillsDir, 0750))
	skillPath := filepath.Join(skillsDir, "test.md")
	require.NoError(t, os.WriteFile(skillPath, []byte("test skill content"), 0600))

	agentsDir := filepath.Join(pluginDir, "agents")
	require.NoError(t, os.Mkdir(agentsDir, 0750))
	agentPath := filepath.Join(agentsDir, "test.md")
	require.NoError(t, os.WriteFile(agentPath, []byte("test agent content"), 0600))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
			Skills:      []SkillRef{{Name: "test-skill", Path: "skills/test.md"}},
			Agents:      []AgentRef{{Name: "test-agent", Path: "agents/test.md"}},
		},
		RootPath:  pluginDir,
		Enabled:   false,
		Installed: true,
	}

	err := m.Enable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.True(t, plugin.Enabled)
	assert.True(t, m.enabled["test-plugin"])
}

func TestManager_Enable_WithSkillsOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	skillsDir := filepath.Join(pluginDir, "skills")
	require.NoError(t, os.Mkdir(skillsDir, 0750))
	skillPath := filepath.Join(skillsDir, "test.md")
	require.NoError(t, os.WriteFile(skillPath, []byte("test skill content"), 0600))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
			Skills:      []SkillRef{{Name: "test-skill", Path: "skills/test.md"}},
		},
		RootPath:  pluginDir,
		Enabled:   false,
		Installed: true,
	}

	err := m.Enable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.True(t, plugin.Enabled)
}

func TestManager_Enable_WithAgentsOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	agentsDir := filepath.Join(pluginDir, "agents")
	require.NoError(t, os.Mkdir(agentsDir, 0750))
	agentPath := filepath.Join(agentsDir, "test.md")
	require.NoError(t, os.WriteFile(agentPath, []byte("test agent content"), 0600))

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
			Agents:      []AgentRef{{Name: "test-agent", Path: "agents/test.md"}},
		},
		RootPath:  pluginDir,
		Enabled:   false,
		Installed: true,
	}

	err := m.Enable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.True(t, plugin.Enabled)
}

func TestManager_Disable_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)

	err := m.Disable(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_Disable_AlreadyDisabled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
		},
		Enabled:   false,
		Installed: true,
	}

	err := m.Disable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.False(t, plugin.Enabled)
}

func TestManager_Disable_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
			Skills:      []SkillRef{{Name: "test-skill", Path: "skills/test.md"}},
			Agents:      []AgentRef{{Name: "test-agent", Path: "agents/test.md"}},
		},
		Enabled:   true,
		Installed: true,
	}
	m.enabled["test-plugin"] = true

	err := m.Disable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.False(t, plugin.Enabled)
	_, exists := m.enabled["test-plugin"]
	assert.False(t, exists)
}

func TestManager_Disable_WithSkillsOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
			Skills:      []SkillRef{{Name: "test-skill", Path: "skills/test.md"}},
		},
		Enabled:   true,
		Installed: true,
	}
	m.enabled["test-plugin"] = true

	err := m.Disable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.False(t, plugin.Enabled)
}

func TestManager_Disable_WithAgentsOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
			Agents:      []AgentRef{{Name: "test-agent", Path: "agents/test.md"}},
		},
		Enabled:   true,
		Installed: true,
	}
	m.enabled["test-plugin"] = true

	err := m.Disable(context.Background(), "test-plugin")
	require.NoError(t, err)

	plugin := m.plugins["test-plugin"]
	assert.False(t, plugin.Enabled)
}

func TestManager_Get_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)

	plugin, err := m.Get("non-existent")
	assert.Error(t, err)
	assert.Nil(t, plugin)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_Get_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	expectedManifest := Manifest{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test plugin",
		Author:      "Test",
		Skills:      []SkillRef{{Name: "test-skill", Path: "skills/test.md"}},
		Agents:      []AgentRef{{Name: "test-agent", Path: "agents/test.md"}},
	}

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest:  expectedManifest,
		RootPath:  "/tmp/test-plugin",
		Enabled:   true,
		Installed: true,
	}

	plugin, err := m.Get("test-plugin")
	require.NoError(t, err)
	assert.NotNil(t, plugin)
	assert.Equal(t, "test-plugin", plugin.Manifest.Name)
	assert.Equal(t, "1.0.0", plugin.Manifest.Version)
	assert.Equal(t, "Test plugin", plugin.Manifest.Description)
	assert.Equal(t, "Test", plugin.Manifest.Author)
	assert.Equal(t, "/tmp/test-plugin", plugin.RootPath)
	assert.True(t, plugin.Enabled)
	assert.True(t, plugin.Installed)
	assert.Len(t, plugin.Manifest.Skills, 1)
	assert.Len(t, plugin.Manifest.Agents, 1)
}

func TestManager_Get_DisabledPlugin(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, nil)
	m.plugins["test-plugin"] = &Plugin{
		Manifest: Manifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Author:      "Test",
		},
		Enabled:   false,
		Installed: true,
	}

	plugin, err := m.Get("test-plugin")
	require.NoError(t, err)
	assert.False(t, plugin.Enabled)
}
