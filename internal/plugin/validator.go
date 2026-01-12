package plugin

import (
	"fmt"
	"regexp"
	"sync"
)

type Validator struct {
	mu               sync.RWMutex
	installedPlugins map[string]*Plugin
	loadedSkills     map[string]string
	loadedAgents     map[string]string
	loadedRules      map[string]string
}

func NewValidator() *Validator {
	return &Validator{
		installedPlugins: make(map[string]*Plugin),
		loadedSkills:     make(map[string]string),
		loadedAgents:     make(map[string]string),
		loadedRules:      make(map[string]string),
	}
}

func (v *Validator) ValidateManifest(m *Manifest) error {
	if err := v.validateManifestSchema(m); err != nil {
		return err
	}

	if err := v.validateVersion(m.Version); err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidatePlugin(p *Plugin, existingPlugins map[string]*Plugin) error {
	if err := v.ValidateManifest(&p.Manifest); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	if err := v.checkConflicts(p, existingPlugins); err != nil {
		return err
	}

	if err := v.checkDependencies(p); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateManifestSchema(m *Manifest) error {
	// Delegate basic validation to the manifest itself
	if err := m.Validate(); err != nil {
		return err
	}

	// Additional validator-specific checks
	if !isValidName(m.Name) {
		return fmt.Errorf("name contains invalid characters (use only lowercase letters, numbers, hyphens)")
	}

	return nil
}

func (v *Validator) validateVersion(version string) error {
	regex := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$`)
	if !regex.MatchString(version) {
		return fmt.Errorf("invalid version format (expected semver: x.y.z)")
	}
	return nil
}

func (v *Validator) checkConflicts(p *Plugin, existingPlugins map[string]*Plugin) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for _, skillRef := range p.Manifest.Skills {
		if existingPlugin, exists := v.loadedSkills[skillRef.Name]; exists && existingPlugin != p.Manifest.Name {
			return fmt.Errorf("skill conflict: %s already provided by plugin %s", skillRef.Name, existingPlugin)
		}
	}

	for _, agentRef := range p.Manifest.Agents {
		if existingPlugin, exists := v.loadedAgents[agentRef.Name]; exists && existingPlugin != p.Manifest.Name {
			return fmt.Errorf("agent conflict: %s already provided by plugin %s", agentRef.Name, existingPlugin)
		}
	}

	for _, ruleRef := range p.Manifest.Rules {
		if existingPlugin, exists := v.loadedRules[ruleRef.Name]; exists && existingPlugin != p.Manifest.Name {
			return fmt.Errorf("rule conflict: %s already provided by plugin %s", ruleRef.Name, existingPlugin)
		}
	}

	return nil
}

func (v *Validator) checkDependencies(p *Plugin) error {
	return nil
}

func (v *Validator) RegisterPlugin(p *Plugin) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, skillRef := range p.Manifest.Skills {
		v.loadedSkills[skillRef.Name] = p.Manifest.Name
	}

	for _, agentRef := range p.Manifest.Agents {
		v.loadedAgents[agentRef.Name] = p.Manifest.Name
	}

	for _, ruleRef := range p.Manifest.Rules {
		v.loadedRules[ruleRef.Name] = p.Manifest.Name
	}

	v.installedPlugins[p.Manifest.Name] = p
}

func (v *Validator) UnregisterPlugin(p *Plugin) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, skillRef := range p.Manifest.Skills {
		if pluginName, exists := v.loadedSkills[skillRef.Name]; exists && pluginName == p.Manifest.Name {
			delete(v.loadedSkills, skillRef.Name)
		}
	}

	for _, agentRef := range p.Manifest.Agents {
		if pluginName, exists := v.loadedAgents[agentRef.Name]; exists && pluginName == p.Manifest.Name {
			delete(v.loadedAgents, agentRef.Name)
		}
	}

	for _, ruleRef := range p.Manifest.Rules {
		if pluginName, exists := v.loadedRules[ruleRef.Name]; exists && pluginName == p.Manifest.Name {
			delete(v.loadedRules, ruleRef.Name)
		}
	}

	delete(v.installedPlugins, p.Manifest.Name)
}

func isValidName(name string) bool {
	regex := regexp.MustCompile(`^[a-z0-9-]+$`)
	return regex.MatchString(name)
}
