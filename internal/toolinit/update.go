package toolinit

import (
	"context"
	"fmt"
	"strings"

	"github.com/victorzhuk/go-ent/internal/version"
)

// UpdateDiff represents changes between installed and new version
type UpdateDiff struct {
	FromVersion string
	ToVersion   string
	NewFiles    []string
	Modified    []string
	Removed     []string
}

// ComponentFilter specifies which components to update
type ComponentFilter struct {
	Agents   bool
	Commands bool
	Skills   bool
}

// ParseComponentFilter parses a component filter string like "agents,skills"
func ParseComponentFilter(filter string) ComponentFilter {
	cf := ComponentFilter{}
	if filter == "" {
		// Empty filter means update all
		cf.Agents = true
		cf.Commands = true
		cf.Skills = true
		return cf
	}

	parts := strings.Split(filter, ",")
	for _, part := range parts {
		switch strings.TrimSpace(part) {
		case "agents", "agent":
			cf.Agents = true
		case "commands", "command":
			cf.Commands = true
		case "skills", "skill":
			cf.Skills = true
		}
	}

	return cf
}

// CalculateUpdateDiff calculates the difference between installed and new version
func CalculateUpdateDiff(installed *EntInfo, newOps []FileOperation, componentFilter ComponentFilter) *UpdateDiff {
	diff := &UpdateDiff{
		NewFiles: []string{},
		Modified: []string{},
		Removed:  []string{},
	}

	if installed != nil {
		diff.FromVersion = installed.Version
	}

	v := version.Get()
	diff.ToVersion = v.Version

	// Build maps for quick lookup
	installedMap := make(map[string]string) // path -> hash
	if installed != nil {
		for _, entry := range installed.Components.Agents {
			if componentFilter.Agents {
				installedMap[entry.Name] = entry.Hash
			}
		}
		for _, entry := range installed.Components.Commands {
			if componentFilter.Commands {
				installedMap[entry.Name] = entry.Hash
			}
		}
		for _, entry := range installed.Components.Skills {
			if componentFilter.Skills {
				installedMap[entry.Name] = entry.Hash
			}
		}
	}

	newMap := make(map[string]string) // path -> hash
	for _, op := range newOps {
		// Apply component filter (with ent/ namespace)
		shouldInclude := false
		if strings.HasPrefix(op.Path, "agents/ent/") || strings.HasPrefix(op.Path, "agent/ent/") {
			shouldInclude = componentFilter.Agents
		} else if strings.HasPrefix(op.Path, "commands/ent/") || strings.HasPrefix(op.Path, "command/ent/") {
			shouldInclude = componentFilter.Commands
		} else if strings.HasPrefix(op.Path, "skills/ent/") || strings.HasPrefix(op.Path, "skill/ent/") {
			shouldInclude = componentFilter.Skills
		}

		if !shouldInclude {
			continue
		}

		newHash := HashContent(op.Content)
		newMap[op.Path] = newHash

		if oldHash, exists := installedMap[op.Path]; exists {
			// File exists - check if modified
			if oldHash != newHash {
				diff.Modified = append(diff.Modified, op.Path)
			}
		} else {
			// New file
			diff.NewFiles = append(diff.NewFiles, op.Path)
		}
	}

	// Find removed files
	for path := range installedMap {
		if _, exists := newMap[path]; !exists {
			diff.Removed = append(diff.Removed, path)
		}
	}

	return diff
}

// IsEmpty returns true if the diff has no changes
func (d *UpdateDiff) IsEmpty() bool {
	return len(d.NewFiles) == 0 && len(d.Modified) == 0 && len(d.Removed) == 0
}

// FormatDiff formats the diff for display
func (d *UpdateDiff) FormatDiff() string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════╗\n")
	if d.FromVersion != "" {
		sb.WriteString(fmt.Sprintf("║  go-ent Update: %s → %s", d.FromVersion, d.ToVersion))
	} else {
		sb.WriteString(fmt.Sprintf("║  go-ent Install: %s", d.ToVersion))
	}
	// Pad to width
	padding := 44 - len(fmt.Sprintf("  go-ent Update: %s → %s", d.FromVersion, d.ToVersion))
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString("║\n")
	sb.WriteString("╚════════════════════════════════════════════╝\n\n")

	if d.IsEmpty() {
		sb.WriteString("✅ No changes\n")
		return sb.String()
	}

	sb.WriteString("📊 Changes:\n\n")

	// Group by component type
	agents := struct{ new, mod, rem []string }{}
	commands := struct{ new, mod, rem []string }{}
	skills := struct{ new, mod, rem []string }{}

	for _, path := range d.NewFiles {
		if strings.Contains(path, "agent") {
			agents.new = append(agents.new, path)
		} else if strings.Contains(path, "command") {
			commands.new = append(commands.new, path)
		} else if strings.Contains(path, "skill") {
			skills.new = append(skills.new, path)
		}
	}

	for _, path := range d.Modified {
		if strings.Contains(path, "agent") {
			agents.mod = append(agents.mod, path)
		} else if strings.Contains(path, "command") {
			commands.mod = append(commands.mod, path)
		} else if strings.Contains(path, "skill") {
			skills.mod = append(skills.mod, path)
		}
	}

	for _, path := range d.Removed {
		if strings.Contains(path, "agent") {
			agents.rem = append(agents.rem, path)
		} else if strings.Contains(path, "command") {
			commands.rem = append(commands.rem, path)
		} else if strings.Contains(path, "skill") {
			skills.rem = append(skills.rem, path)
		}
	}

	// Format agents
	if len(agents.new)+len(agents.mod)+len(agents.rem) > 0 {
		sb.WriteString("Agents:\n")
		for _, path := range agents.new {
			sb.WriteString(fmt.Sprintf("  + %s (NEW)\n", path))
		}
		for _, path := range agents.mod {
			sb.WriteString(fmt.Sprintf("  ~ %s (MODIFIED)\n", path))
		}
		for _, path := range agents.rem {
			sb.WriteString(fmt.Sprintf("  - %s (REMOVED)\n", path))
		}
		sb.WriteString("\n")
	}

	// Format commands
	if len(commands.new)+len(commands.mod)+len(commands.rem) > 0 {
		sb.WriteString("Commands:\n")
		for _, path := range commands.new {
			sb.WriteString(fmt.Sprintf("  + %s (NEW)\n", path))
		}
		for _, path := range commands.mod {
			sb.WriteString(fmt.Sprintf("  ~ %s (MODIFIED)\n", path))
		}
		for _, path := range commands.rem {
			sb.WriteString(fmt.Sprintf("  - %s (REMOVED)\n", path))
		}
		sb.WriteString("\n")
	}

	// Format skills
	if len(skills.new)+len(skills.mod)+len(skills.rem) > 0 {
		sb.WriteString("Skills:\n")
		for _, path := range skills.new {
			sb.WriteString(fmt.Sprintf("  + %s (NEW)\n", path))
		}
		for _, path := range skills.mod {
			sb.WriteString(fmt.Sprintf("  ~ %s (MODIFIED)\n", path))
		}
		for _, path := range skills.rem {
			sb.WriteString(fmt.Sprintf("  - %s (REMOVED)\n", path))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ShouldUpdate checks if an update is needed
func ShouldUpdate(installed *EntInfo) bool {
	if installed == nil {
		return true // No installation, should install
	}

	v := version.Get()

	// Always update if current version is "dev"
	if v.Version == "dev" || installed.Version == "dev" {
		return true
	}

	// Update if versions differ
	return installed.Version != v.Version
}

// PerformUpdate executes an update with the given configuration.
//
// NOTE: This is currently a stub implementation that validates update prerequisites
// but does not perform the actual update. The actual update logic is handled by the
// CLI init command with the --update flag, which calls adapter.Generate() directly.
//
// This function exists to:
//  1. Validate that an update is needed (via ShouldUpdate)
//  2. Provide a future extension point for programmatic updates
//  3. Ensure updates go through the proper CLI workflow with user interaction
//
// To update plugins, use: go-ent init --update
func PerformUpdate(ctx context.Context, adapter Adapter, cfg *GenerateConfig, componentFilter ComponentFilter) error {
	targetDir := adapter.TargetDir()

	// Load existing info
	installed, err := LoadEntInfo(cfg.Path + "/" + targetDir)
	if err != nil {
		return fmt.Errorf("load installed info: %w", err)
	}

	// Check if update is needed
	if !ShouldUpdate(installed) && !cfg.Force {
		fmt.Println("✅ Already up to date")
		return nil
	}

	// Actual update must be performed through CLI to ensure proper user interaction,
	// backup creation, and error handling.
	return fmt.Errorf("update must be called through CLI init command with --update flag")
}
