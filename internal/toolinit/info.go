package toolinit

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EntInfo tracks the installed version and components
type EntInfo struct {
	Version     string            `yaml:"version"`
	VCSRef      string            `yaml:"vcs_ref"`
	InstalledAt time.Time         `yaml:"installed_at"`
	Components  ComponentManifest `yaml:"components"`
}

// ComponentManifest tracks installed components with their hashes
type ComponentManifest struct {
	Agents   []ComponentEntry `yaml:"agents,omitempty"`
	Commands []ComponentEntry `yaml:"commands,omitempty"`
	Skills   []ComponentEntry `yaml:"skills,omitempty"`
}

// ComponentEntry represents a single installed component
type ComponentEntry struct {
	Name string `yaml:"name"`
	Hash string `yaml:"hash"`
}

// LoadEntInfo loads the ent.info.yaml file from a directory
func LoadEntInfo(dir string) (*EntInfo, error) {
	infoPath := filepath.Join(dir, "ent.info.yaml")

	data, err := os.ReadFile(infoPath) // #nosec G304 -- controlled config/template file path
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No info file exists
		}
		return nil, fmt.Errorf("read info file: %w", err)
	}

	var info EntInfo
	if err := yaml.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parse info file: %w", err)
	}

	return &info, nil
}

// SaveEntInfo saves the ent.info.yaml file to a directory
func SaveEntInfo(dir string, info *EntInfo) error {
	infoPath := filepath.Join(dir, "ent.info.yaml")

	data, err := yaml.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal info: %w", err)
	}

	if err := os.WriteFile(infoPath, data, 0600); err != nil {
		return fmt.Errorf("write info file: %w", err)
	}

	return nil
}

// HashContent computes SHA256 hash of content
func HashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", hash)
}

// BuildComponentManifest builds a manifest from file operations
func BuildComponentManifest(ops []FileOperation) ComponentManifest {
	manifest := ComponentManifest{
		Agents:   []ComponentEntry{},
		Commands: []ComponentEntry{},
		Skills:   []ComponentEntry{},
	}

	for _, op := range ops {
		entry := ComponentEntry{
			Name: op.Path,
			Hash: HashContent(op.Content),
		}

		switch {
		case strings.HasPrefix(op.Path, "agents/ent/") || strings.HasPrefix(op.Path, "agent/ent/"):
			manifest.Agents = append(manifest.Agents, entry)
		case strings.HasPrefix(op.Path, "commands/ent/") || strings.HasPrefix(op.Path, "command/ent/"):
			manifest.Commands = append(manifest.Commands, entry)
		case strings.HasPrefix(op.Path, "skills/ent/") || strings.HasPrefix(op.Path, "skill/ent/"):
			manifest.Skills = append(manifest.Skills, entry)
		}
	}

	return manifest
}
