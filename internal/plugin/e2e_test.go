package plugin

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type e2eTrackingRegistry struct {
	mu                 sync.Mutex
	registeredSkills   map[string]string
	registeredAgents   map[string]string
	unregisteredSkills map[string]bool
	unregisteredAgents map[string]bool
	skillReg           *skill.Registry
	agentReg           *agent.Registry
}

func (r *e2eTrackingRegistry) RegisterSkill(name, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.registeredSkills == nil {
		r.registeredSkills = make(map[string]string)
	}
	r.registeredSkills[name] = path
	if r.skillReg != nil {
		return r.skillReg.RegisterSkill(name, path)
	}
	return nil
}

func (r *e2eTrackingRegistry) RegisterAgent(name, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.registeredAgents == nil {
		r.registeredAgents = make(map[string]string)
	}
	r.registeredAgents[name] = path
	if r.agentReg != nil {
		return r.agentReg.RegisterAgent(name, path)
	}
	return nil
}

func (r *e2eTrackingRegistry) UnregisterSkill(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.unregisteredSkills == nil {
		r.unregisteredSkills = make(map[string]bool)
	}
	r.unregisteredSkills[name] = true
	if r.skillReg != nil {
		return r.skillReg.UnregisterSkill(name)
	}
	return nil
}

func (r *e2eTrackingRegistry) UnregisterAgent(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.unregisteredAgents == nil {
		r.unregisteredAgents = make(map[string]bool)
	}
	r.unregisteredAgents[name] = true
	if r.agentReg != nil {
		return r.agentReg.UnregisterAgent(name)
	}
	return nil
}

func createSamplePluginInDir(t *testing.T, pluginsDir string) {
	pluginDir := filepath.Join(pluginsDir, "sample-e2e-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	skillsDir := filepath.Join(pluginDir, "skills", "sample-skill")
	require.NoError(t, os.MkdirAll(skillsDir, 0750))

	skillPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: sample-skill
description: Sample skill for E2E testing. Auto-activates for: testing, e2e, sample.
---
# Sample Skill

This is a sample skill for E2E testing.

## Purpose

Provides sample functionality for testing.

## Usage

Used to verify plugin loading and skill registration.
`
	require.NoError(t, os.WriteFile(skillPath, []byte(skillContent), 0600))

	agentsDir := filepath.Join(pluginDir, "agents")
	require.NoError(t, os.Mkdir(agentsDir, 0750))

	agentPath := filepath.Join(agentsDir, "sample-agent.md")
	agentContent := `---
name: sample-agent
description: Sample agent for E2E testing
model: claude-3-opus
color: "#FF6B6B"
skills:
  - sample-skill
tools:
  file-read: true
  file-write: true
---
# Sample Agent

This is a sample agent for E2E testing.

## Purpose

Provides sample agent functionality for testing.

## Usage

Used to verify plugin loading and agent registration.

## Capabilities

- Can use sample-skill
- Has file-read and file-write tools
`
	require.NoError(t, os.WriteFile(agentPath, []byte(agentContent), 0600))

	manifestPath := filepath.Join(pluginDir, ManifestFile)
	manifestContent := `name: sample-e2e-plugin
version: 1.0.0
description: Sample plugin for E2E testing
author: E2E Test

skills:
  - name: sample-skill
    path: skills/sample-skill/SKILL.md

agents:
  - name: sample-agent
    path: agents/sample-agent.md
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifestContent), 0600))
}

func TestPluginE2E_FullLifecycle(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	createSamplePluginInDir(t, tmpDir)

	pluginDir := filepath.Join(tmpDir, "sample-e2e-plugin")
	_, err := os.Stat(pluginDir)
	assert.NoError(t, err, "plugin directory should exist")

	skillReg := skill.NewRegistry()
	agentReg := agent.NewRegistry()

	tracking := &e2eTrackingRegistry{
		skillReg: skillReg,
		agentReg: agentReg,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, nil, logger)

	err = m.Initialize(context.Background())
	require.NoError(t, err, "initialize manager")

	plugins := m.List()
	require.Len(t, plugins, 1, "should have exactly one plugin")
	assert.Equal(t, "sample-e2e-plugin", plugins[0].Name, "plugin name should match")
	assert.Equal(t, "1.0.0", plugins[0].Version, "plugin version should match")
	assert.Equal(t, "Sample plugin for E2E testing", plugins[0].Description, "description should match")
	assert.Equal(t, "E2E Test", plugins[0].Author, "author should match")
	assert.False(t, plugins[0].Enabled, "plugin should be disabled after initialize")
	assert.Equal(t, 1, plugins[0].Skills, "should report 1 skill")
	assert.Equal(t, 1, plugins[0].Agents, "should report 1 agent")

	err = m.Enable(context.Background(), "sample-e2e-plugin")
	require.NoError(t, err, "enable plugin")

	enabledPlugins := m.List()
	require.Len(t, enabledPlugins, 1)
	assert.True(t, enabledPlugins[0].Enabled, "plugin should be enabled after Enable call")

	// Test persistence by creating new manager instance
	m = NewManager(tmpDir, tracking, nil, logger)

	ctx := context.Background()

	err = m.Initialize(ctx)
	require.NoError(t, err)

	plugin, err := m.Get("sample-e2e-plugin")
	require.NoError(t, err)
	assert.True(t, plugin.Enabled)
	assert.True(t, plugin.Installed)

	err = m.Disable(ctx, "sample-e2e-plugin")
	require.NoError(t, err)

	plugin, err = m.Get("sample-e2e-plugin")
	require.NoError(t, err)
	assert.False(t, plugin.Enabled)
	assert.True(t, plugin.Installed)

	err = m.Enable(ctx, "sample-e2e-plugin")
	require.NoError(t, err)

	plugin, err = m.Get("sample-e2e-plugin")
	require.NoError(t, err)
	assert.True(t, plugin.Enabled)
	assert.True(t, plugin.Installed)

	err = m.Uninstall(ctx, "sample-e2e-plugin")
	require.NoError(t, err)

	_, err = m.Get("sample-e2e-plugin")
	assert.Error(t, err)

	plugins = m.List()
	assert.Empty(t, plugins)
}

func TestPluginE2E_MultiplePlugins(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	plugin1Dir := filepath.Join(tmpDir, "plugin-one")
	require.NoError(t, os.Mkdir(plugin1Dir, 0750))

	skills1Dir := filepath.Join(plugin1Dir, "skills", "skill-one")
	require.NoError(t, os.MkdirAll(skills1Dir, 0750))
	skill1Path := filepath.Join(skills1Dir, "SKILL.md")
	skill1Content := `---
name: skill-one
description: First skill. Auto-activates for: one.
---
# Skill One
`
	require.NoError(t, os.WriteFile(skill1Path, []byte(skill1Content), 0600))

	agents1Dir := filepath.Join(plugin1Dir, "agents")
	require.NoError(t, os.Mkdir(agents1Dir, 0750))
	agent1Path := filepath.Join(agents1Dir, "agent-one.md")
	agent1Content := `---
name: agent-one
description: First agent
model: claude-3-opus
---
# Agent One
`
	require.NoError(t, os.WriteFile(agent1Path, []byte(agent1Content), 0600))

	manifest1Path := filepath.Join(plugin1Dir, ManifestFile)
	manifest1 := `name: plugin-one
version: 1.0.0
description: First plugin
author: Test

skills:
  - name: skill-one
    path: skills/skill-one/SKILL.md

agents:
  - name: agent-one
    path: agents/agent-one.md
`
	require.NoError(t, os.WriteFile(manifest1Path, []byte(manifest1), 0600))

	plugin2Dir := filepath.Join(tmpDir, "plugin-two")
	require.NoError(t, os.Mkdir(plugin2Dir, 0750))

	skills2Dir := filepath.Join(plugin2Dir, "skills", "skill-two")
	require.NoError(t, os.MkdirAll(skills2Dir, 0750))
	skill2Path := filepath.Join(skills2Dir, "SKILL.md")
	skill2Content := `---
name: skill-two
description: Second skill. Auto-activates for: two.
---
# Skill Two
`
	require.NoError(t, os.WriteFile(skill2Path, []byte(skill2Content), 0600))

	manifest2Path := filepath.Join(plugin2Dir, ManifestFile)
	manifest2 := `name: plugin-two
version: 1.0.0
description: Second plugin
author: Test

skills:
  - name: skill-two
    path: skills/skill-two/SKILL.md
`
	require.NoError(t, os.WriteFile(manifest2Path, []byte(manifest2), 0600))

	skillReg := skill.NewRegistry()
	agentReg := agent.NewRegistry()

	tracking := &e2eTrackingRegistry{
		skillReg: skillReg,
		agentReg: agentReg,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	m := NewManager(tmpDir, tracking, nil, logger)

	ctx := context.Background()

	err := m.Initialize(ctx)
	require.NoError(t, err)

	plugins := m.List()
	require.Len(t, plugins, 2, "should have two plugins")

	err = m.Enable(ctx, "plugin-one")
	require.NoError(t, err, "enable plugin-one")

	_, err = skillReg.Get("skill-one")
	assert.NoError(t, err, "skill-one should be found")

	_, err = agentReg.Get("agent-one")
	assert.NoError(t, err, "agent-one should be found")

	_, err = skillReg.Get("skill-two")
	assert.Error(t, err, "skill-two should not be found yet")

	err = m.Enable(ctx, "plugin-two")
	require.NoError(t, err, "enable plugin-two")

	_, err = skillReg.Get("skill-two")
	assert.NoError(t, err, "skill-two should be found")

	tracking.mu.Lock()
	assert.Equal(t, 2, len(tracking.registeredSkills), "should have 2 registered skills")
	assert.Equal(t, 1, len(tracking.registeredAgents), "should have 1 registered agent")
	tracking.mu.Unlock()

	err = m.Disable(ctx, "plugin-one")
	require.NoError(t, err, "disable plugin-one")

	_, err = skillReg.Get("skill-one")
	assert.Error(t, err, "skill-one should not be found after disable")

	_, err = agentReg.Get("agent-one")
	assert.Error(t, err, "agent-one should not be found after disable")

	_, err = skillReg.Get("skill-two")
	assert.NoError(t, err, "skill-two should still be found")
}
