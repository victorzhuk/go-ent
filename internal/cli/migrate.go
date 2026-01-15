package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/victorzhuk/go-ent/internal/toolinit"
	"gopkg.in/yaml.v3"
)

type MigrationStatus struct {
	Status     string
	MetaFile   string
	PromptFile string
	LegacyFile string
}

type agentMetaYAML struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Model        string   `yaml:"model"`
	Color        string   `yaml:"color"`
	Skills       []string `yaml:"skills"`
	Tools        []string `yaml:"tools"`
	Dependencies []string `yaml:"dependencies"`
	Tags         []string `yaml:"tags"`
}

func newMigrateCmd() *cobra.Command {
	var (
		check   bool
		execute bool
	)

	cmd := &cobra.Command{
		Use:   "migrate [path]",
		Short: "Migrate legacy single-file agents to split format",
		Long: `Migrate legacy single-file agent files to the new split format.

This command scans for legacy agent files in the agents/ directory and migrates
them to the new split format:
  - meta/{agent-name}.yaml - Agent metadata
  - prompts/agents/{agent-name}.md - Agent prompt content

Subcommands:
  --check    Show migration status without modifying files
  --execute  Perform the migration

Examples:
  # Check migration status
  go-ent migrate --check

  # Execute migration
  go-ent migrate --execute

  # Migrate in specific directory
  go-ent migrate /path/to/project --execute`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !check && !execute {
				return fmt.Errorf("specify either --check or --execute")
			}
			if check && execute {
				return fmt.Errorf("--check and --execute are mutually exclusive")
			}

			projectPath := "."
			if len(args) > 0 {
				projectPath = args[0]
			}

			return MigrateAgents(cmd.Context(), projectPath, check, execute)
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "show migration status only")
	cmd.Flags().BoolVar(&execute, "execute", false, "perform migration")

	return cmd
}

// MigrateAgents handles the migration of legacy agents to split format
func MigrateAgents(ctx context.Context, projectPath string, checkOnly bool, executeMigration bool) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("project directory does not exist: %s", absPath)
	}

	legacyAgents := scanLegacyAgents(absPath)
	if len(legacyAgents) == 0 {
		fmt.Println("✅ No legacy agent files found")
		return nil
	}

	status := checkMigrationStatus(absPath, legacyAgents)

	if checkOnly {
		return printMigrationStatus(status)
	}

	if executeMigration {
		return executeMigrateAgents(absPath, legacyAgents, status)
	}

	return nil
}

func scanLegacyAgents(projectPath string) []string {
	agentsDir := filepath.Join(projectPath, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return nil
	}

	var agents []string
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		agents = append(agents, filepath.Join(agentsDir, entry.Name()))
	}

	return agents
}

func checkMigrationStatus(projectPath string, legacyAgents []string) map[string]*MigrationStatus {
	status := make(map[string]*MigrationStatus)

	for _, agentPath := range legacyAgents {
		baseName := filepath.Base(agentPath)
		agentName := strings.TrimSuffix(baseName, ".md")

		metaFile := filepath.Join(projectPath, "agents", "meta", agentName+".yaml")
		promptFile := filepath.Join(projectPath, "agents", "prompts", "agents", agentName+".md")

		st := &MigrationStatus{
			LegacyFile: agentPath,
			MetaFile:   metaFile,
			PromptFile: promptFile,
		}

		metaExists, _ := os.Stat(metaFile)
		promptExists, _ := os.Stat(promptFile)

		if metaExists != nil && promptExists != nil {
			st.Status = "migrated"
		} else if metaExists != nil || promptExists != nil {
			st.Status = "partially_migrated"
		} else {
			st.Status = "needs_migration"
		}

		status[agentName] = st
	}

	return status
}

func printMigrationStatus(status map[string]*MigrationStatus) error {
	fmt.Printf("\n╔════════════════════════════════════════╗\n")
	fmt.Printf("║  Agent Migration Status              ║\n")
	fmt.Printf("╚════════════════════════════════════════╝\n\n")

	total := len(status)
	migrated := 0
	partial := 0
	needs := 0

	fmt.Printf("%-20s %-20s %s\n", "Agent", "Status", "Details")
	fmt.Printf("%s\n", strings.Repeat("-", 80))

	for name, st := range status {
		switch st.Status {
		case "migrated":
			migrated++
		case "partially_migrated":
			partial++
		case "needs_migration":
			needs++
		}

		var details string
		switch st.Status {
		case "migrated":
			details = "✓ Already in split format"
		case "partially_migrated":
			var missing []string
			if _, err := os.Stat(st.MetaFile); os.IsNotExist(err) {
				missing = append(missing, "meta")
			}
			if _, err := os.Stat(st.PromptFile); os.IsNotExist(err) {
				missing = append(missing, "prompt")
			}
			details = fmt.Sprintf("⚠ Partial: missing %s", strings.Join(missing, ", "))
		case "needs_migration":
			details = "→ Needs migration"
		}

		fmt.Printf("%-20s %-20s %s\n", name, st.Status, details)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total:    %d\n", total)
	fmt.Printf("  Migrated: %d\n", migrated)
	fmt.Printf("  Partial:  %d\n", partial)
	fmt.Printf("  Needs:    %d\n", needs)

	if needs > 0 {
		fmt.Printf("\nRun 'go-ent migrate --execute' to migrate agents\n")
	}

	return nil
}

func executeMigrateAgents(projectPath string, legacyAgents []string, status map[string]*MigrationStatus) error {
	var toMigrate []string
	for _, st := range status {
		if st.Status == "needs_migration" || st.Status == "partially_migrated" {
			toMigrate = append(toMigrate, st.LegacyFile)
		}
	}

	if len(toMigrate) == 0 {
		fmt.Println("✅ All agents are already migrated")
		return nil
	}

	backupDir := filepath.Join(projectPath, ".ent-backup", time.Now().Format("20060102-150405"))
	if err := createBackup(backupDir, toMigrate); err != nil {
		fmt.Printf("⚠️  Backup failed (continuing anyway): %v\n", err)
	} else {
		fmt.Printf("📦 Backup created: %s\n", backupDir)
	}

	for _, agentPath := range toMigrate {
		if err := performMigration(agentPath, projectPath); err != nil {
			return fmt.Errorf("migrate %s: %w", agentPath, err)
		}
	}

	fmt.Printf("\n✅ Migration complete\n")
	return nil
}

func performMigration(agentPath, projectPath string) error {
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return fmt.Errorf("read agent file: %w", err)
	}

	metadata, body, err := toolinit.ParseFrontmatter(string(content))
	if err != nil {
		return fmt.Errorf("parse frontmatter: %w", err)
	}

	if len(metadata) == 0 {
		return fmt.Errorf("no frontmatter found")
	}

	baseName := filepath.Base(agentPath)
	agentName := strings.TrimSuffix(baseName, ".md")

	meta := agentMetaYAML{
		Name:         agentName,
		Skills:       []string{},
		Tools:        []string{},
		Dependencies: []string{},
		Tags:         []string{},
	}

	if name, ok := metadata["name"].(string); ok {
		meta.Name = name
	}
	if desc, ok := metadata["description"].(string); ok {
		meta.Description = desc
	}
	if model, ok := metadata["model"].(string); ok {
		meta.Model = model
	}
	if color, ok := metadata["color"].(string); ok {
		meta.Color = color
	}

	if skills, ok := metadata["skills"].([]interface{}); ok {
		for _, s := range skills {
			if str, ok := s.(string); ok {
				meta.Skills = append(meta.Skills, str)
			}
		}
	}

	if tools, ok := metadata["tools"].([]interface{}); ok {
		for _, t := range tools {
			if str, ok := t.(string); ok {
				meta.Tools = append(meta.Tools, str)
			}
		}
	}

	if tags, ok := metadata["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				meta.Tags = append(meta.Tags, tagStr)
			}
		}
	}

	if deps, ok := metadata["dependencies"].([]interface{}); ok {
		for _, d := range deps {
			if str, ok := d.(string); ok {
				meta.Dependencies = append(meta.Dependencies, str)
			}
		}
	}

	inferredDeps := inferDependencies(body)
	for _, dep := range inferredDeps {
		if !contains(meta.Dependencies, dep) {
			meta.Dependencies = append(meta.Dependencies, dep)
		}
	}

	if len(meta.Dependencies) == 0 {
		meta.Dependencies = nil
	}

	metaPath := filepath.Join(projectPath, "agents", "meta", agentName+".yaml")
	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		return fmt.Errorf("create meta dir: %w", err)
	}

	metaYAML, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaYAML, 0644); err != nil {
		return fmt.Errorf("write meta file: %w", err)
	}

	promptPath := filepath.Join(projectPath, "agents", "prompts", "agents", agentName+".md")
	if err := os.MkdirAll(filepath.Dir(promptPath), 0755); err != nil {
		return fmt.Errorf("create prompt dir: %w", err)
	}

	if err := os.WriteFile(promptPath, []byte(body+"\n"), 0644); err != nil {
		return fmt.Errorf("write prompt file: %w", err)
	}

	return nil
}

func inferDependencies(body string) []string {
	re := regexp.MustCompile(`@ent:(\w+)`)
	matches := re.FindAllStringSubmatch(body, -1)

	seen := make(map[string]bool)
	var deps []string

	for _, match := range matches {
		if len(match) > 1 {
			agentName := match[1]
			if !seen[agentName] {
				seen[agentName] = true
				deps = append(deps, agentName)
			}
		}
	}

	return deps
}

func createBackup(backupDir string, files []string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	for _, file := range files {
		src, err := os.Open(file)
		if err != nil {
			continue
		}
		defer src.Close()

		dstPath := filepath.Join(backupDir, filepath.Base(file))
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			continue
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			continue
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
