package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Loader struct {
	pluginsDir string
}

func NewLoader(pluginsDir string) *Loader {
	return &Loader{
		pluginsDir: pluginsDir,
	}
}

func (l *Loader) LoadPlugin(name string) (*Plugin, error) {
	pluginDir := filepath.Join(l.pluginsDir, name)
	manifestPath := filepath.Join(pluginDir, ManifestFile)

	manifest, err := ParseManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	plugin := &Plugin{
		Manifest:  *manifest,
		RootPath:  pluginDir,
		Installed: true,
		Enabled:   true,
	}

	if err := l.validatePlugin(plugin); err != nil {
		return nil, fmt.Errorf("validate plugin: %w", err)
	}

	return plugin, nil
}

func (l *Loader) LoadAll() ([]*Plugin, error) {
	var plugins []*Plugin

	entries, err := os.ReadDir(l.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		plugin, err := l.LoadPlugin(entry.Name())
		if err != nil {
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func (l *Loader) validatePlugin(p *Plugin) error {
	for _, skillRef := range p.Manifest.Skills {
		absPath := filepath.Join(p.RootPath, skillRef.Path)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("skill file not found: %s", skillRef.Path)
		}

		if !strings.HasSuffix(skillRef.Path, ".md") {
			return fmt.Errorf("skill file must be .md: %s", skillRef.Path)
		}
	}

	for _, agentRef := range p.Manifest.Agents {
		absPath := filepath.Join(p.RootPath, agentRef.Path)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("agent file not found: %s", agentRef.Path)
		}

		if !strings.HasSuffix(agentRef.Path, ".md") {
			return fmt.Errorf("agent file must be .md: %s", agentRef.Path)
		}
	}

	for _, ruleRef := range p.Manifest.Rules {
		absPath := filepath.Join(p.RootPath, ruleRef.Path)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("rule file not found: %s", ruleRef.Path)
		}

		if !strings.HasSuffix(ruleRef.Path, ".yaml") && !strings.HasSuffix(ruleRef.Path, ".yml") {
			return fmt.Errorf("rule file must be .yaml or .yml: %s", ruleRef.Path)
		}
	}

	return nil
}
