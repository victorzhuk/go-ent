package plugin

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/marketplace"
)

type trackingRegistry struct {
	mu                 sync.Mutex
	registeredSkills   map[string]string
	registeredAgents   map[string]string
	unregisteredSkills map[string]bool
	unregisteredAgents map[string]bool
}

func (r *trackingRegistry) RegisterSkill(name, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.registeredSkills == nil {
		r.registeredSkills = make(map[string]string)
	}
	r.registeredSkills[name] = path
	return nil
}

func (r *trackingRegistry) RegisterAgent(name, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.registeredAgents == nil {
		r.registeredAgents = make(map[string]string)
	}
	r.registeredAgents[name] = path
	return nil
}

func (r *trackingRegistry) UnregisterSkill(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.unregisteredSkills == nil {
		r.unregisteredSkills = make(map[string]bool)
	}
	r.unregisteredSkills[name] = true
	return nil
}

func (r *trackingRegistry) UnregisterAgent(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.unregisteredAgents == nil {
		r.unregisteredAgents = make(map[string]bool)
	}
	r.unregisteredAgents[name] = true
	return nil
}

func TestPluginManager_LoadsPluginSuccessfully(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	skillsDir := filepath.Join(pluginDir, "skills", "test-skill")
	require.NoError(t, os.MkdirAll(skillsDir, 0750))
	skillPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: "Test skill for integration test"
---
# Test Skill
This is a test skill.
`
	require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

	agentsDir := filepath.Join(pluginDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0750))
	agentPath := filepath.Join(agentsDir, "test-agent.md")
	agentContent := `---
name: test-agent
description: "Test agent for integration test"
---
# Test Agent
This is a test agent.
`
	require.NoError(t, os.WriteFile(agentPath, []byte(agentContent), 0600))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: test-plugin
version: 1.0.0
description: Test plugin for integration
author: Test Author

skills:
  - name: test-skill
    path: skills/test-skill/SKILL.md

agents:
  - name: test-agent
    path: agents/test-agent.md
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	tracking := &trackingRegistry{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, &marketplace.Client{}, logger)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	plugins := m.List()
	require.Len(t, plugins, 1)
	assert.Equal(t, "test-plugin", plugins[0].Name)
	assert.Equal(t, "1.0.0", plugins[0].Version)
	assert.Equal(t, "Test plugin for integration", plugins[0].Description)
	assert.Equal(t, "Test Author", plugins[0].Author)
	assert.False(t, plugins[0].Enabled)
	assert.Equal(t, 1, plugins[0].Skills)
	assert.Equal(t, 1, plugins[0].Agents)

	err = m.Enable(context.Background(), "test-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Contains(t, tracking.registeredSkills, "test-skill")
	assert.Contains(t, tracking.registeredAgents, "test-agent")
	assert.Equal(t, filepath.Join(pluginDir, "skills/test-skill/SKILL.md"), tracking.registeredSkills["test-skill"])
	assert.Equal(t, filepath.Join(pluginDir, "agents/test-agent.md"), tracking.registeredAgents["test-agent"])
	tracking.mu.Unlock()

	enabledPlugin, err := m.Get("test-plugin")
	require.NoError(t, err)
	assert.True(t, enabledPlugin.Enabled)
}

func TestPluginManager_EnableDisableCycle(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "cycle-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	skillsDir := filepath.Join(pluginDir, "skills", "cycle-skill")
	require.NoError(t, os.MkdirAll(skillsDir, 0750))
	skillPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: cycle-skill
description: "Cycle test skill"
---
# Cycle Skill
Test skill for enable/disable cycle.
`
	require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: cycle-plugin
version: 1.0.0
description: Plugin for enable/disable cycle test
author: Test Author

skills:
  - name: cycle-skill
    path: skills/cycle-skill/SKILL.md
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	tracking := &trackingRegistry{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, &marketplace.Client{}, logger)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	err = m.Enable(context.Background(), "cycle-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Contains(t, tracking.registeredSkills, "cycle-skill")
	_, skillRegistered := tracking.registeredSkills["cycle-skill"]
	assert.True(t, skillRegistered)
	tracking.mu.Unlock()

	plugin, err := m.Get("cycle-plugin")
	require.NoError(t, err)
	assert.True(t, plugin.Enabled)

	err = m.Disable(context.Background(), "cycle-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Contains(t, tracking.unregisteredSkills, "cycle-skill")
	_, skillUnregistered := tracking.unregisteredSkills["cycle-skill"]
	assert.True(t, skillUnregistered)
	tracking.mu.Unlock()

	plugin, err = m.Get("cycle-plugin")
	require.NoError(t, err)
	assert.False(t, plugin.Enabled)

	err = m.Enable(context.Background(), "cycle-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Equal(t, 1, len(tracking.registeredSkills))
	tracking.mu.Unlock()

	plugin, err = m.Get("cycle-plugin")
	require.NoError(t, err)
	assert.True(t, plugin.Enabled)
}

func TestPluginManager_LoadsMultiplePlugins(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	createPlugin := func(name, skillName, agentName string) {
		pluginDir := filepath.Join(tmpDir, name)
		require.NoError(t, os.Mkdir(pluginDir, 0750))

		skillsDir := filepath.Join(pluginDir, "skills", skillName)
		require.NoError(t, os.MkdirAll(skillsDir, 0750))
		skillPath := filepath.Join(skillsDir, "SKILL.md")
		skillContent := `---
name: ` + skillName + `
description: "Test skill"
---
# Test Skill
`
		require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

		agentsDir := filepath.Join(pluginDir, "agents")
		require.NoError(t, os.MkdirAll(agentsDir, 0750))
		agentPath := filepath.Join(agentsDir, agentName+".md")
		agentContent := `---
name: ` + agentName + `
description: "Test agent"
---
# Test Agent
`
		require.NoError(t, os.WriteFile(agentPath, []byte(agentContent), 0600))

		manifestPath := filepath.Join(pluginDir, ManifestFile)
		manifest := `name: ` + name + `
version: 1.0.0
description: Test plugin
author: Test

skills:
  - name: ` + skillName + `
    path: skills/` + skillName + `/SKILL.md

agents:
  - name: ` + agentName + `
    path: agents/` + agentName + `.md
`
		require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))
	}

	createPlugin("plugin-a", "skill-a", "agent-a")
	createPlugin("plugin-b", "skill-b", "agent-b")
	createPlugin("plugin-c", "skill-c", "agent-c")

	tracking := &trackingRegistry{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, &marketplace.Client{}, logger)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	plugins := m.List()
	require.Len(t, plugins, 3)

	pluginNames := make(map[string]bool)
	for _, p := range plugins {
		pluginNames[p.Name] = true
	}
	assert.True(t, pluginNames["plugin-a"])
	assert.True(t, pluginNames["plugin-b"])
	assert.True(t, pluginNames["plugin-c"])

	for _, pluginName := range []string{"plugin-a", "plugin-b", "plugin-c"} {
		err := m.Enable(context.Background(), pluginName)
		require.NoError(t, err, "should enable %s", pluginName)
	}

	tracking.mu.Lock()
	assert.Equal(t, 3, len(tracking.registeredSkills))
	assert.Equal(t, 3, len(tracking.registeredAgents))
	assert.Contains(t, tracking.registeredSkills, "skill-a")
	assert.Contains(t, tracking.registeredSkills, "skill-b")
	assert.Contains(t, tracking.registeredSkills, "skill-c")
	assert.Contains(t, tracking.registeredAgents, "agent-a")
	assert.Contains(t, tracking.registeredAgents, "agent-b")
	assert.Contains(t, tracking.registeredAgents, "agent-c")
	tracking.mu.Unlock()
}

func TestPluginManager_InvalidManifestSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	validPluginDir := filepath.Join(tmpDir, "valid-plugin")
	require.NoError(t, os.Mkdir(validPluginDir, 0750))

	skillsDir := filepath.Join(validPluginDir, "skills", "valid-skill")
	require.NoError(t, os.MkdirAll(skillsDir, 0750))
	skillPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: valid-skill
description: "Valid skill"
---
# Valid Skill
`
	require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

	validManifestPath := filepath.Join(validPluginDir, ManifestFile)
	validManifest := `name: valid-plugin
version: 1.0.0
description: Valid plugin
author: Test

skills:
  - name: valid-skill
    path: skills/valid-skill/SKILL.md
`
	require.NoError(t, os.WriteFile(validManifestPath, []byte(validManifest), 0600))

	invalidPluginDir := filepath.Join(tmpDir, "invalid-plugin")
	require.NoError(t, os.Mkdir(invalidPluginDir, 0750))

	invalidManifestPath := filepath.Join(invalidPluginDir, ManifestFile)
	invalidManifest := `name: invalid-plugin
version: 1.0.0
description: "Invalid YAML: [unclosed bracket
author: Test
`
	require.NoError(t, os.WriteFile(invalidManifestPath, []byte(invalidManifest), 0600))

	tracking := &trackingRegistry{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, &marketplace.Client{}, logger)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	plugins := m.List()
	require.Len(t, plugins, 1)
	assert.Equal(t, "valid-plugin", plugins[0].Name)
	assert.Equal(t, 1, plugins[0].Skills)

	err = m.Enable(context.Background(), "valid-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Equal(t, 1, len(tracking.registeredSkills))
	assert.Contains(t, tracking.registeredSkills, "valid-skill")
	tracking.mu.Unlock()

	_, err = m.Get("invalid-plugin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPluginManager_EnableDisableWithMultipleResources(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "multi-resource-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	for i := 1; i <= 3; i++ {
		skillName := fmt.Sprintf("skill-%d", i)
		skillsDir := filepath.Join(pluginDir, "skills", skillName)
		require.NoError(t, os.MkdirAll(skillsDir, 0750))
		skillPath := filepath.Join(skillsDir, "SKILL.md")
		skillContent := fmt.Sprintf(`---
name: %s
description: "Test skill %d"
---
# Skill %d
`, skillName, i, i)
		require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

		agentName := fmt.Sprintf("agent-%d", i)
		agentsDir := filepath.Join(pluginDir, "agents")
		require.NoError(t, os.MkdirAll(agentsDir, 0750))
		agentPath := filepath.Join(agentsDir, agentName+".md")
		agentContent := fmt.Sprintf(`---
name: %s
description: "Test agent %d"
---
# Agent %d
`, agentName, i, i)
		require.NoError(t, os.WriteFile(agentPath, []byte(agentContent), 0600))
	}

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: multi-resource-plugin
version: 1.0.0
description: Plugin with multiple resources
author: Test

skills:
  - name: skill-1
    path: skills/skill-1/SKILL.md
  - name: skill-2
    path: skills/skill-2/SKILL.md
  - name: skill-3
    path: skills/skill-3/SKILL.md

agents:
  - name: agent-1
    path: agents/agent-1.md
  - name: agent-2
    path: agents/agent-2.md
  - name: agent-3
    path: agents/agent-3.md
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	tracking := &trackingRegistry{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, &marketplace.Client{}, logger)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	err = m.Enable(context.Background(), "multi-resource-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Equal(t, 3, len(tracking.registeredSkills))
	assert.Equal(t, 3, len(tracking.registeredAgents))
	for i := 1; i <= 3; i++ {
		skillName := fmt.Sprintf("skill-%d", i)
		agentName := fmt.Sprintf("agent-%d", i)
		assert.Contains(t, tracking.registeredSkills, skillName)
		assert.Contains(t, tracking.registeredAgents, agentName)
	}
	tracking.mu.Unlock()

	err = m.Disable(context.Background(), "multi-resource-plugin")
	require.NoError(t, err)

	tracking.mu.Lock()
	assert.Equal(t, 3, len(tracking.unregisteredSkills))
	assert.Equal(t, 3, len(tracking.unregisteredAgents))
	for i := 1; i <= 3; i++ {
		skillName := fmt.Sprintf("skill-%d", i)
		agentName := fmt.Sprintf("agent-%d", i)
		assert.Contains(t, tracking.unregisteredSkills, skillName)
		assert.Contains(t, tracking.unregisteredAgents, agentName)
	}
	tracking.mu.Unlock()

	plugin, err := m.Get("multi-resource-plugin")
	require.NoError(t, err)
	assert.False(t, plugin.Enabled)
}

func TestPluginManager_ConcurrentEnableDisable(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pluginDir := filepath.Join(tmpDir, "concurrent-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	skillsDir := filepath.Join(pluginDir, "skills", "concurrent-skill")
	require.NoError(t, os.MkdirAll(skillsDir, 0750))
	skillPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: concurrent-skill
description: "Concurrent test skill"
---
# Concurrent Skill
`
	require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifest := `name: concurrent-plugin
version: 1.0.0
description: Plugin for concurrent test
author: Test

skills:
  - name: concurrent-skill
    path: skills/concurrent-skill/SKILL.md
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0600))

	tracking := &trackingRegistry{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, &marketplace.Client{}, logger)

	err := m.Initialize(context.Background())
	require.NoError(t, err)

	errors := make(chan error, 10)
	for i := 0; i < 5; i++ {
		go func() {
			errors <- m.Enable(context.Background(), "concurrent-plugin")
		}()
	}

	for i := 0; i < 5; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	tracking.mu.Lock()
	assert.Equal(t, 1, len(tracking.registeredSkills))
	tracking.mu.Unlock()

	for i := 0; i < 5; i++ {
		go func() {
			errors <- m.Disable(context.Background(), "concurrent-plugin")
		}()
	}

	for i := 0; i < 5; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	tracking.mu.Lock()
	assert.Equal(t, 1, len(tracking.unregisteredSkills))
	tracking.mu.Unlock()
}
