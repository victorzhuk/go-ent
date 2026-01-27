package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var pluginFS interface {
	ReadDir(name string) ([]os.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

func SetPluginFS(fs interface {
	ReadDir(name string) ([]os.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}) {
	pluginFS = fs
}

type agentMeta struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Model        string   `yaml:"model"`
	Color        string   `yaml:"color"`
	Skills       []string `yaml:"skills"`
	Tools        []string `yaml:"tools"`
	Dependencies []string `yaml:"dependencies"`
	Tags         []string `yaml:"tags"`
	Role         string
	Complexity   string
}

func loadAgents() (map[string]*agentMeta, error) {
	agents := make(map[string]*agentMeta)

	entries, err := pluginFS.ReadDir("plugins/go-ent/agents/meta")
	if err != nil {
		return nil, fmt.Errorf("read agents/meta directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join("plugins/go-ent/agents/meta", entry.Name())
		data, err := pluginFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var meta agentMeta
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		for _, tag := range meta.Tags {
			if strings.HasPrefix(tag, "role:") {
				meta.Role = strings.TrimPrefix(tag, "role:")
			}
			if strings.HasPrefix(tag, "complexity:") {
				meta.Complexity = strings.TrimPrefix(tag, "complexity:")
			}
		}

		agents[meta.Name] = &meta
	}

	return agents, nil
}

func loadPrompts() (map[string]string, error) {
	prompts := make(map[string]string)

	entries, err := pluginFS.ReadDir("plugins/go-ent/agents/prompts/agents")
	if err != nil {
		return nil, fmt.Errorf("read agents/prompts/agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join("plugins/go-ent/agents/prompts/agents", entry.Name())
		data, err := pluginFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		prompts[name] = string(data)
	}

	return prompts, nil
}

func loadShared() (string, error) {
	sharedFiles := []string{
		"_principals.md",
		"_judgment.md",
		"_openspec.md",
		"_conventions.md",
		"_handoffs.md",
		"_tooling.md",
	}

	var shared strings.Builder

	for _, filename := range sharedFiles {
		path := filepath.Join("plugins/go-ent/agents/prompts/shared", filename)
		data, err := pluginFS.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", path, err)
		}

		shared.Write(data)
		shared.WriteString("\n\n")
	}

	return shared.String(), nil
}

func loadTemplate(tool string) (*template.Template, error) {
	var templateFile string

	switch tool {
	case "claude":
		templateFile = "plugins/go-ent/agents/templates/claude.yaml.tmpl"
	case "opencode":
		templateFile = "plugins/go-ent/agents/templates/opencode.yaml.tmpl"
	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}

	data, err := pluginFS.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", templateFile, err)
	}

	tpl, err := template.New("agent").Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return tpl, nil
}

func renderAgent(meta *agentMeta, prompt, shared string, tpl *template.Template) (string, error) {
	var result strings.Builder

	if err := tpl.Execute(&result, meta); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	result.WriteString("---\n\n")
	result.WriteString(shared)
	result.WriteString("\n\n")
	result.WriteString(prompt)

	return result.String(), nil
}

func getAgentPath(tool, prefix, name string) string {
	switch tool {
	case "claude":
		return filepath.Join(".claude", "agents", prefix, name+".md")
	case "opencode":
		return filepath.Join(".opencode", "agent", name+".md")
	default:
		return ""
	}
}

func writeFile(path, content string, force, dryRun bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
	}

	if dryRun {
		fmt.Printf("Would create: %s\n", path)
		return nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	fmt.Printf("Created: %s\n", path)
	return nil
}

func copyCommands(tool, prefix string, force, dryRun bool) error {
	entries, err := pluginFS.ReadDir("plugins/go-ent/commands")
	if err != nil {
		return fmt.Errorf("read commands directory: %w", err)
	}

	var targetDir string
	switch tool {
	case "claude":
		targetDir = filepath.Join(".claude", "commands", prefix)
	case "opencode":
		targetDir = filepath.Join(".opencode", "commands", prefix)
	default:
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		srcPath := filepath.Join("plugins/go-ent/commands", entry.Name())
		data, err := pluginFS.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", srcPath, err)
		}

		dstPath := filepath.Join(targetDir, entry.Name())
		if err := writeFile(dstPath, string(data), force, dryRun); err != nil {
			return err
		}
	}

	return nil
}

func copySkills(tool, prefix string, force, dryRun bool) error {
	var walk func(dir string) error

	walk = func(dir string) error {
		entries, err := pluginFS.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("read directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			srcPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if err := walk(srcPath); err != nil {
					return err
				}
				continue
			}

			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			data, err := pluginFS.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("read %s: %w", srcPath, err)
			}

			relPath := strings.TrimPrefix(srcPath, "plugins/go-ent/skills/")

			var targetDir string
			switch tool {
			case "claude":
				targetDir = filepath.Join(".claude", "skills", prefix)
			case "opencode":
				targetDir = filepath.Join(".opencode", "skills", prefix)
			default:
				return fmt.Errorf("unsupported tool: %s", tool)
			}

			dstPath := filepath.Join(targetDir, relPath)
			if err := writeFile(dstPath, string(data), force, dryRun); err != nil {
				return err
			}
		}

		return nil
	}

	return walk("plugins/go-ent/skills")
}

func printSummary(agentCount, commandCount, skillCount int, tool, prefix string, dryRun bool) {
	var toolName string
	var commandFormat string
	var restartRequired bool

	switch tool {
	case "claude":
		toolName = "Claude Code"
		commandFormat = "/ent:plan"
		restartRequired = true
	case "opencode":
		toolName = "OpenCode"
		commandFormat = "/plan"
		restartRequired = false
	default:
		toolName = tool
		commandFormat = "/ent:plan"
		restartRequired = true
	}

	var agentPath string
	switch tool {
	case "claude":
		agentPath = filepath.Join(".claude", "agents", prefix)
	case "opencode":
		agentPath = filepath.Join(".opencode", "agent")
	}

	var commandPath string
	switch tool {
	case "claude":
		commandPath = filepath.Join(".claude", "commands", prefix)
	case "opencode":
		commandPath = filepath.Join(".opencode", "commands", prefix)
	}

	var skillPath string
	switch tool {
	case "claude":
		skillPath = filepath.Join(".claude", "skills", prefix)
	case "opencode":
		skillPath = filepath.Join(".opencode", "skills", prefix)
	}

	if dryRun {
		fmt.Printf("\n✅ Preview: Would initialize go-ent for %s\n\n", toolName)
		fmt.Println("Would create:")
		fmt.Printf("  %d agents in %s/\n", agentCount, agentPath)
		fmt.Printf("  %d commands in %s/\n", commandCount, commandPath)
		fmt.Printf("  %d skills in %s/\n", skillCount, skillPath)
		return
	}

	fmt.Printf("\n✅ Initialized go-ent for %s\n\n", toolName)
	fmt.Println("Created:")
	fmt.Printf("  %d agents in %s/\n", agentCount, agentPath)
	fmt.Printf("  %d commands in %s/\n", commandCount, commandPath)
	fmt.Printf("  %d skills in %s/\n", skillCount, skillPath)

	fmt.Println("\nNext steps:")
	if restartRequired {
		fmt.Println("  1. Restart Claude Code")
		fmt.Printf("  2. Run: %s \"description\"\n", commandFormat)
	} else {
		fmt.Printf("  1. Run: %s \"description\"\n", commandFormat)
	}
}

type initFlags struct {
	tool   string
	prefix string
	force  bool
	dryRun bool
}

func newInitCmd() *cobra.Command {
	flags := &initFlags{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize go-ent in the current project",
		Long: `Initialize go-ent configuration for the current project.

This command creates the necessary configuration files and agent definitions
for use with Claude Code or OpenCode.

Supported tools:
  claude     - Configure for Claude Code
  opencode   - Configure for OpenCode

Examples:
  ent init --tool=claude
  ent init --tool=opencode
  ent init --tool=claude,opencode
  ent init --tool=claude --prefix=myproject
  ent init --tool=claude --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.tool == "" {
				return errors.New("--tool is required")
			}

			if pluginFS == nil {
				return errors.New("plugin filesystem not initialized")
			}

			agents, err := loadAgents()
			if err != nil {
				return fmt.Errorf("load agents: %w", err)
			}

			prompts, err := loadPrompts()
			if err != nil {
				return fmt.Errorf("load prompts: %w", err)
			}

			shared, err := loadShared()
			if err != nil {
				return fmt.Errorf("load shared: %w", err)
			}

			agentCount := len(agents)

			tools := strings.Split(flags.tool, ",")
			for _, tool := range tools {
				tool = strings.TrimSpace(tool)

				tpl, err := loadTemplate(tool)
				if err != nil {
					return fmt.Errorf("load template for %s: %w", tool, err)
				}

				for name, meta := range agents {
					prompt, ok := prompts[name]
					if !ok {
						return fmt.Errorf("prompt not found for agent: %s", name)
					}

					content, err := renderAgent(meta, prompt, shared, tpl)
					if err != nil {
						return fmt.Errorf("render agent %s: %w", name, err)
					}

					path := getAgentPath(tool, flags.prefix, name)
					if err := writeFile(path, content, flags.force, flags.dryRun); err != nil {
						return err
					}
				}

				entries, err := pluginFS.ReadDir("plugins/go-ent/commands")
				if err != nil {
					return fmt.Errorf("read commands directory: %w", err)
				}
				commandCount := 0
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
						commandCount++
					}
				}

				if err := copyCommands(tool, flags.prefix, flags.force, flags.dryRun); err != nil {
					return fmt.Errorf("copy commands: %w", err)
				}

				var countSkills func(dir string) (int, error)
				countSkills = func(dir string) (int, error) {
					entries, err := pluginFS.ReadDir(dir)
					if err != nil {
						return 0, err
					}
					count := 0
					for _, entry := range entries {
						if entry.IsDir() {
							c, err := countSkills(filepath.Join(dir, entry.Name()))
							if err != nil {
								return 0, err
							}
							count += c
							continue
						}
						if strings.HasSuffix(entry.Name(), ".md") {
							count++
						}
					}
					return count, nil
				}
				skillCount, err := countSkills("plugins/go-ent/skills")
				if err != nil {
					return fmt.Errorf("count skills: %w", err)
				}

				if err := copySkills(tool, flags.prefix, flags.force, flags.dryRun); err != nil {
					return fmt.Errorf("copy skills: %w", err)
				}

				printSummary(agentCount, commandCount, skillCount, tool, flags.prefix, flags.dryRun)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.tool, "tool", "", "Target tool(s) (required, comma-separated: claude,opencode)")
	_ = cmd.MarkFlagRequired("tool")

	cmd.Flags().StringVar(&flags.prefix, "prefix", "ent", "Prefix for configuration directories")
	cmd.Flags().BoolVar(&flags.force, "force", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview changes without writing files")

	return cmd
}
