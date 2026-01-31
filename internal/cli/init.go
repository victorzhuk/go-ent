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

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/pkg"
)

type toolPresets struct {
	Presets map[string][]string `yaml:"presets"`
}

type promptsConfig struct {
	Shared []string `yaml:"shared"`
	Main   string   `yaml:"main"`
}

type agentMeta struct {
	Name                  string        `yaml:"name"`
	Description           string        `yaml:"description"`
	Extends               string        `yaml:"extends"`
	Model                 string        `yaml:"model"`
	Color                 string        `yaml:"color"`
	Role                  string        `yaml:"role"`
	Complexity            string        `yaml:"complexity"`
	Skills                []string      `yaml:"skills"`
	Tools                 []string      `yaml:"tools"`
	ToolPresets           []string      `yaml:"toolPresets"`
	DisallowedToolPresets []string      `yaml:"disallowedToolPresets"`
	DisallowedTools       []string      `yaml:"disallowedTools"`
	Dependencies          []string      `yaml:"dependencies"`
	Prompts               promptsConfig `yaml:"prompts"`
}

// ModelClaude converts internal model name to Claude Code format
func (m *agentMeta) ModelClaude() string {
	switch m.Model {
	case "main":
		return "sonnet"
	case "fast":
		return "haiku"
	case "heavy":
		return "opus"
	default:
		return m.Model
	}
}

// ModelOpenCode returns internal model name as-is for OpenCode
func (m *agentMeta) ModelOpenCode() string {
	return m.Model
}

// ToolsOpenCode converts tools list to OpenCode object format
func (m *agentMeta) ToolsOpenCode() []string {
	result := make([]string, 0, len(m.Tools))
	for _, tool := range m.Tools {
		// Convert to lowercase and format as YAML object entry
		toolLower := strings.ToLower(tool)
		result = append(result, fmt.Sprintf("%s: true", toolLower))
	}
	return result
}

// DisallowedToolsClaude returns disallowed tools with Serena prefix expanded
func (m *agentMeta) DisallowedToolsClaude() []string {
	result := make([]string, 0)

	serenaDenylist := make(map[string]bool)
	for _, preset := range m.DisallowedToolPresets {
		if preset == "serena-editing" {
			serenaDenylist["mcp__plugin_serena_serena__replace_symbol_body"] = true
			serenaDenylist["mcp__plugin_serena_serena__insert_after_symbol"] = true
			serenaDenylist["mcp__plugin_serena_serena__insert_before_symbol"] = true
			serenaDenylist["mcp__plugin_serena_serena__replace_content"] = true
			serenaDenylist["mcp__plugin_serena_serena__create_text_file"] = true
		}
	}

	for tool := range serenaDenylist {
		result = append(result, tool)
	}

	// Add any additional explicitly disallowed tools
	result = append(result, m.DisallowedTools...)

	return result
}

func loadAgents() (map[string]*agentMeta, error) {
	bases, err := loadBases()
	if err != nil {
		return nil, fmt.Errorf("load bases: %w", err)
	}

	agents := make(map[string]*agentMeta)

	entries, err := pkg.FS.ReadDir("agents/meta")
	if err != nil {
		return nil, fmt.Errorf("read agents/meta directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join("agents/meta", entry.Name())
		data, err := pkg.FS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var meta agentMeta
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		if meta.Extends != "" {
			base, ok := bases[meta.Extends]
			if !ok {
				return nil, fmt.Errorf("%s: extends unknown base: %s", entry.Name(), meta.Extends)
			}
			merged := mergeAgents(base, &meta)
			meta = *merged
		}

		if err := validateAgent(&meta, entry.Name()); err != nil {
			return nil, err
		}

		agents[meta.Name] = &meta
	}

	return agents, nil
}

func loadBases() (map[string]*agentMeta, error) {
	bases := make(map[string]*agentMeta)

	entries, err := pkg.FS.ReadDir("agents/meta/bases")
	if err != nil {
		return bases, nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join("agents/meta/bases", entry.Name())
		data, err := pkg.FS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var meta agentMeta
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		baseName := strings.TrimSuffix(entry.Name(), ".yaml")
		bases[baseName] = &meta
	}

	return bases, nil
}

func mergeAgents(base, variant *agentMeta) *agentMeta {
	merged := *base

	if variant.Name != "" {
		merged.Name = variant.Name
	}
	if variant.Description != "" {
		merged.Description = variant.Description
	}
	if variant.Model != "" {
		merged.Model = variant.Model
	}
	if variant.Color != "" {
		merged.Color = variant.Color
	}
	if variant.Role != "" {
		merged.Role = variant.Role
	}
	if variant.Complexity != "" {
		merged.Complexity = variant.Complexity
	}

	if len(variant.Skills) > 0 {
		merged.Skills = variant.Skills
	}
	if len(variant.Tools) > 0 {
		merged.Tools = variant.Tools
	}
	if len(variant.ToolPresets) > 0 {
		merged.ToolPresets = variant.ToolPresets
	}
	if len(variant.DisallowedToolPresets) > 0 {
		merged.DisallowedToolPresets = variant.DisallowedToolPresets
	}
	if len(variant.DisallowedTools) > 0 {
		merged.DisallowedTools = variant.DisallowedTools
	}
	if len(variant.Dependencies) > 0 {
		merged.Dependencies = variant.Dependencies
	}

	return &merged
}

func loadToolPresets() (*toolPresets, error) {
	data, err := pkg.FS.ReadFile("agents/presets/tools.yaml")
	if err != nil {
		return nil, fmt.Errorf("read tool presets: %w", err)
	}

	var presets toolPresets
	if err := yaml.Unmarshal(data, &presets); err != nil {
		return nil, fmt.Errorf("parse tool presets: %w", err)
	}

	return &presets, nil
}

func expandToolPresets(meta *agentMeta, presets *toolPresets) {
	toolSet := make(map[string]bool)

	for _, tools := range meta.Tools {
		toolSet[tools] = true
	}

	for _, presetName := range meta.ToolPresets {
		tools, ok := presets.Presets[presetName]
		if !ok {
			continue
		}
		for _, tool := range tools {
			toolSet[tool] = true
		}
	}

	disallowedSet := make(map[string]bool)
	for _, tool := range meta.DisallowedTools {
		disallowedSet[tool] = true
	}

	for _, presetName := range meta.DisallowedToolPresets {
		tools, ok := presets.Presets[presetName]
		if !ok {
			continue
		}
		for _, tool := range tools {
			disallowedSet[tool] = true
		}
	}

	for tool := range disallowedSet {
		delete(toolSet, tool)
	}

	meta.Tools = make([]string, 0, len(toolSet))
	for tool := range toolSet {
		meta.Tools = append(meta.Tools, tool)
	}
}

func validateAgent(meta *agentMeta, filename string) error {
	if meta.Name == "" {
		return fmt.Errorf("%s: name is required", filename)
	}

	// Agent names in metadata are simple (e.g. "coder", "planner")
	// The "ent:" prefix is added by the plugin system, not stored in metadata
	if strings.Contains(meta.Name, ":") {
		return fmt.Errorf("%s: name should not contain ':' in metadata (got: %s)", filename, meta.Name)
	}

	if meta.Description == "" {
		return fmt.Errorf("%s: description is required", filename)
	}

	if len(meta.Description) < 10 {
		return fmt.Errorf("%s: description must be at least 10 characters", filename)
	}

	if meta.Model == "" {
		return fmt.Errorf("%s: model is required", filename)
	}

	validModels := map[string]bool{"fast": true, "main": true, "heavy": true}
	if !validModels[meta.Model] {
		return fmt.Errorf("%s: model must be one of [fast, main, heavy] (got: %s)", filename, meta.Model)
	}

	if meta.Color != "" && !strings.HasPrefix(meta.Color, "#") {
		return fmt.Errorf("%s: color must be a hex code starting with # (got: %s)", filename, meta.Color)
	}

	if meta.Role != "" {
		validRoles := map[string]bool{"planning": true, "execution": true, "validation": true, "research": true}
		if !validRoles[meta.Role] {
			return fmt.Errorf("%s: role must be one of [planning, execution, validation, research] (got: %s)", filename, meta.Role)
		}
	}

	if meta.Complexity != "" {
		validComplexity := map[string]bool{"simple": true, "standard": true, "heavy": true}
		if !validComplexity[meta.Complexity] {
			return fmt.Errorf("%s: complexity must be one of [simple, standard, heavy] (got: %s)", filename, meta.Complexity)
		}
	}

	validPresets := map[string]bool{
		"readonly":        true,
		"editing":         true,
		"serena-analysis": true,
		"serena-editing":  true,
		"planning":        true,
	}
	for _, preset := range meta.ToolPresets {
		if !validPresets[preset] {
			return fmt.Errorf("%s: unknown tool preset: %s", filename, preset)
		}
	}

	for _, preset := range meta.DisallowedToolPresets {
		if !validPresets[preset] {
			return fmt.Errorf("%s: unknown disallowed tool preset: %s", filename, preset)
		}
	}

	for _, dep := range meta.Dependencies {
		// Dependencies in metadata are simple names (e.g. "coder", "planner")
		// The "ent:" prefix is added by the plugin system, not stored in metadata
		if strings.Contains(dep, ":") {
			return fmt.Errorf("%s: dependency should not contain ':' in metadata (got: %s)", filename, dep)
		}
	}

	return nil
}

func loadPrompts() (map[string]string, error) {
	prompts := make(map[string]string)

	entries, err := pkg.FS.ReadDir("agents/prompts/agents")
	if err != nil {
		return nil, fmt.Errorf("read agents/prompts/agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join("agents/prompts/agents", entry.Name())
		data, err := pkg.FS.ReadFile(path)
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
		"principals.md",
		"judgment.md",
		"openspec.md",
		"conventions.md",
		"tooling.md",
	}

	var shared strings.Builder

	for _, filename := range sharedFiles {
		path := filepath.Join("prompts", filename)
		data, err := pkg.FS.ReadFile(path)
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
		templateFile = "agents/templates/claude.yaml.tmpl"
	case "opencode":
		templateFile = "agents/templates/opencode.yaml.tmpl"
	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}

	data, err := pkg.FS.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", templateFile, err)
	}

	tpl, err := template.New("agent").Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return tpl, nil
}

func processIncludes(content string, params map[string]string) string {
	for key, value := range params {
		placeholder := "{{" + key + "}}"
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return content
}

func renderAgent(meta *agentMeta, prompt, shared string, tpl *template.Template) (string, error) {
	var result strings.Builder

	if err := tpl.Execute(&result, meta); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	roleParams := map[string]string{
		"ROLE":         getRoleDisplay(meta),
		"ROLE_CONTEXT": getRoleContext(meta),
		"AGENT_NAME":   meta.Name,
	}

	processedPrompt := processIncludes(prompt, roleParams)
	processedShared := processIncludes(shared, roleParams)

	// Template already includes closing ---, just add content
	result.WriteString("\n")
	result.WriteString(processedShared)
	result.WriteString("\n\n")
	result.WriteString(processedPrompt)

	return result.String(), nil
}

func getRoleDisplay(meta *agentMeta) string {
	roleMap := map[string]string{
		"planning":   "Planning",
		"execution":  "Implementation",
		"validation": "Validation",
		"research":   "Research",
	}

	if display, ok := roleMap[meta.Role]; ok {
		return display
	}

	baseName := meta.Name
	if idx := strings.LastIndex(baseName, ":"); idx != -1 {
		baseName = baseName[idx+1:]
	}

	if len(baseName) > 0 {
		return strings.ToUpper(baseName[:1]) + baseName[1:]
	}
	return baseName
}

func getRoleContext(meta *agentMeta) string {
	contextMap := map[string]string{
		"planning":   "task breakdown and architecture design",
		"execution":  "code implementation and development",
		"validation": "testing and quality assurance",
		"research":   "investigation and analysis",
	}

	if context, ok := contextMap[meta.Role]; ok {
		return context
	}
	return "software development"
}

func getAgentPath(tool, prefix, name string) string {
	// Extract base name from prefixed name (e.g., "ent:coder" -> "coder")
	baseName := name
	if idx := strings.LastIndex(name, ":"); idx != -1 {
		baseName = name[idx+1:]
	}

	switch tool {
	case "claude":
		return filepath.Join(".claude", "agents", prefix, baseName+".md")
	case "opencode":
		// Use same prefixed structure as Claude: .opencode/agents/<prefix>/<name>.md
		return filepath.Join(".opencode", "agents", prefix, baseName+".md")
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
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	fmt.Printf("Created: %s\n", path)
	return nil
}

func copyCommands(tool, prefix string, force, dryRun bool) error {
	entries, err := pkg.FS.ReadDir("commands")
	if err != nil {
		return fmt.Errorf("read commands directory: %w", err)
	}

	var targetDir string
	switch tool {
	case "claude":
		targetDir = filepath.Join(".claude", "commands", prefix)
	case "opencode":
		// Use same prefixed structure as Claude: .opencode/commands/<prefix>/
		targetDir = filepath.Join(".opencode", "commands", prefix)
	default:
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		srcPath := filepath.Join("commands", entry.Name())
		data, err := pkg.FS.ReadFile(srcPath)
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
	var walk func(dir string, baseTargetDir string) error

	walk = func(dir string, baseTargetDir string) error {
		entries, err := pkg.FS.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("read directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			srcPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if err := walk(srcPath, filepath.Join(baseTargetDir, entry.Name())); err != nil {
					return err
				}
				continue
			}

			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			data, err := pkg.FS.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("read %s: %w", srcPath, err)
			}

			dstPath := filepath.Join(baseTargetDir, entry.Name())
			if err := writeFile(dstPath, string(data), force, dryRun); err != nil {
				return err
			}
		}

		return nil
	}

	// Use same prefixed structure for both Claude and OpenCode
	baseTargetDir := filepath.Join("."+tool, "skills", prefix)
	return walk("skills", baseTargetDir)
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

	// Use same prefixed structure for both Claude and OpenCode
	agentPath := filepath.Join("."+tool, "agents", prefix)
	commandPath := filepath.Join("."+tool, "commands", prefix)
	skillPath := filepath.Join("."+tool, "skills", prefix)

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

			// Plugin FS is now always available via pkg.FS
			if false {
				return errors.New("plugin filesystem not initialized")
			}

			agents, err := loadAgents()
			if err != nil {
				return fmt.Errorf("load agents: %w", err)
			}

			presets, err := loadToolPresets()
			if err != nil {
				return fmt.Errorf("load tool presets: %w", err)
			}

			for _, meta := range agents {
				expandToolPresets(meta, presets)
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

				global, _ := config.LoadGlobalModelConfig()
				project, _ := config.LoadProjectModelConfig(".")
				cfg := config.MergeModelConfigs(global, project)
				resolver := config.NewModelResolver(cfg, tool)

				for name, meta := range agents {
					prompt, ok := prompts[name]
					if !ok {
						return fmt.Errorf("prompt not found for agent: %s", name)
					}

					// Create a copy of meta to avoid modifying the original when processing multiple tools
					metaCopy := *meta
					metaCopy.Model = resolver.ResolveAgent(meta.Model)

					content, err := renderAgent(&metaCopy, prompt, shared, tpl)
					if err != nil {
						return fmt.Errorf("render agent %s: %w", name, err)
					}

					path := getAgentPath(tool, flags.prefix, name)
					if err := writeFile(path, content, flags.force, flags.dryRun); err != nil {
						return err
					}
				}

				entries, err := pkg.FS.ReadDir("commands")
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
					entries, err := pkg.FS.ReadDir(dir)
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
				skillCount, err := countSkills("skills")
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

	cmd.Flags().StringVar(&flags.tool, "tools", "", "Target tool(s) (required, comma-separated: claude,opencode)")
	_ = cmd.MarkFlagRequired("tools")

	cmd.Flags().StringVar(&flags.prefix, "prefix", "ent", "Prefix for configuration directories")
	cmd.Flags().BoolVar(&flags.force, "force", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview changes without writing files")

	return cmd
}

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate agent definitions",
		Long: `Validate all agent definition files against the schema.

This command checks that all agent YAML files in agents/meta/
conform to the required schema, including:
  - Required fields (name, description, model)
  - Valid enum values (model, role, complexity)
  - Proper naming conventions (no colon in name)
  - Valid tool preset references
  - Proper dependency references

Examples:
  ent validate`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Plugin FS is now always available via pkg.FS
			if false {
				return errors.New("plugin filesystem not initialized")
			}

			agents, err := loadAgents()
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Printf("✅ Validated %d agent definitions\n", len(agents))
			fmt.Println("\nAll agents passed validation:")
			for name := range agents {
				fmt.Printf("  ✓ %s\n", name)
			}

			return nil
		},
	}

	return cmd
}
