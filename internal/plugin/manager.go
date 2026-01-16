package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/victorzhuk/go-ent/internal/marketplace"
)

type Manager struct {
	plugins     map[string]*Plugin
	pluginsDir  string
	enabled     map[string]bool
	mu          sync.RWMutex
	registry    Registry
	marketplace MarketplaceClient
	logger      *slog.Logger
}

type Registry interface {
	RegisterSkill(name, path string) error
	RegisterAgent(name, path string) error
	UnregisterSkill(name string) error
	UnregisterAgent(name string) error
}

type MarketplaceClient interface {
	Download(ctx context.Context, name, version string) ([]byte, error)
}

func NewManager(pluginsDir string, registry Registry, marketplace MarketplaceClient, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}

	return &Manager{
		plugins:     make(map[string]*Plugin),
		pluginsDir:  pluginsDir,
		enabled:     make(map[string]bool),
		registry:    registry,
		marketplace: marketplace,
		logger:      logger,
	}
}

func (m *Manager) Initialize(ctx context.Context) error {
	if err := os.MkdirAll(m.pluginsDir, 0750); err != nil {
		return fmt.Errorf("create plugins directory: %w", err)
	}

	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		return fmt.Errorf("read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(m.pluginsDir, entry.Name())
		manifestPath := filepath.Join(pluginPath, ManifestFile)

		manifest, err := ParseManifest(manifestPath)
		if err != nil {
			m.logger.Warn("failed to parse plugin manifest", "plugin_dir", entry.Name(), "error", err)
			continue
		}

		plugin := &Plugin{
			Manifest:  *manifest,
			RootPath:  pluginPath,
			Installed: true,
			Enabled:   m.enabled[manifest.Name],
		}

		m.plugins[manifest.Name] = plugin
	}

	return nil
}

func (m *Manager) Install(ctx context.Context, name, version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already installed", name)
	}

	client, ok := m.marketplace.(*marketplace.Client)
	if !ok {
		return fmt.Errorf("invalid marketplace client type")
	}

	installer := marketplace.NewInstaller(client, m.pluginsDir)
	if err := installer.Install(ctx, name, version); err != nil {
		return fmt.Errorf("install plugin: %w", err)
	}

	if err := installer.Validate(name); err != nil {
		pluginDir := filepath.Join(m.pluginsDir, name)
		_ = os.RemoveAll(pluginDir)
		return fmt.Errorf("validate plugin: %w", err)
	}

	manifestPath := filepath.Join(m.pluginsDir, name, ManifestFile)
	manifest, err := ParseManifest(manifestPath)
	if err != nil {
		pluginDir := filepath.Join(m.pluginsDir, name)
		_ = os.RemoveAll(pluginDir)
		return fmt.Errorf("parse manifest: %w", err)
	}

	plugin := &Plugin{
		Manifest:  *manifest,
		RootPath:  filepath.Join(m.pluginsDir, name),
		Installed: true,
		Enabled:   true,
	}

	m.plugins[name] = plugin
	m.enabled[name] = true

	return nil
}

func (m *Manager) Uninstall(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	pluginDir := filepath.Join(m.pluginsDir, name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("remove plugin directory: %w", err)
	}

	delete(m.plugins, name)
	delete(m.enabled, name)

	return nil
}

func (m *Manager) Enable(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if plugin.Enabled {
		return nil
	}

	if err := m.loadPlugin(plugin); err != nil {
		return fmt.Errorf("load plugin: %w", err)
	}

	plugin.Enabled = true
	m.enabled[name] = true

	return nil
}

func (m *Manager) Disable(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !plugin.Enabled {
		return nil
	}

	if err := m.unloadPlugin(plugin); err != nil {
		return fmt.Errorf("unload plugin: %w", err)
	}

	plugin.Enabled = false
	delete(m.enabled, name)

	return nil
}

func (m *Manager) List() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(m.plugins))
	for _, p := range m.plugins {
		infos = append(infos, PluginInfo{
			Name:        p.Manifest.Name,
			Version:     p.Manifest.Version,
			Description: p.Manifest.Description,
			Author:      p.Manifest.Author,
			Enabled:     p.Enabled,
			Skills:      len(p.Manifest.Skills),
			Agents:      len(p.Manifest.Agents),
			Rules:       len(p.Manifest.Rules),
		})
	}

	return infos
}

func (m *Manager) Get(name string) (*Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

func (m *Manager) loadPlugin(p *Plugin) error {
	for _, skillRef := range p.Manifest.Skills {
		absPath, err := p.ResolvePath(skillRef.Path)
		if err != nil {
			return fmt.Errorf("resolve skill path %s: %w", skillRef.Path, err)
		}

		if err := m.registry.RegisterSkill(skillRef.Name, absPath); err != nil {
			return fmt.Errorf("register skill %s: %w", skillRef.Name, err)
		}
	}

	for _, agentRef := range p.Manifest.Agents {
		absPath, err := p.ResolvePath(agentRef.Path)
		if err != nil {
			return fmt.Errorf("resolve agent path %s: %w", agentRef.Path, err)
		}

		if err := m.registry.RegisterAgent(agentRef.Name, absPath); err != nil {
			return fmt.Errorf("register agent %s: %w", agentRef.Name, err)
		}
	}

	return nil
}

func (m *Manager) unloadPlugin(p *Plugin) error {
	for _, skillRef := range p.Manifest.Skills {
		if err := m.registry.UnregisterSkill(skillRef.Name); err != nil {
			return fmt.Errorf("unregister skill %s: %w", skillRef.Name, err)
		}
	}

	for _, agentRef := range p.Manifest.Agents {
		if err := m.registry.UnregisterAgent(agentRef.Name); err != nil {
			return fmt.Errorf("unregister agent %s: %w", agentRef.Name, err)
		}
	}

	return nil
}

type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Enabled     bool   `json:"enabled"`
	Skills      int    `json:"skills"`
	Agents      int    `json:"agents"`
	Rules       int    `json:"rules"`
}
