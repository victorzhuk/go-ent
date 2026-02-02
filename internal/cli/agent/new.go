package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/template"
)

func newAgentCmd() *cobra.Command {
	var (
		templateName   string
		description    string
		model          string
		role           string
		complexity     string
		skills         string
		tools          string
		nonInteractive bool
	)

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new agent from a template",
		Long: `Create a new agent from a built-in or custom template.

Interactive mode (default):
  Prompts for template selection, description, model, role, and other options.

Non-interactive mode:
  Use flags to specify all options without prompts.
  Requires --template flag at minimum.

Examples:
  # Interactive mode
  ent agent new my-agent

  # Non-interactive mode with all flags
  ent agent new api-reviewer \
    --template standard \
    --description "Reviews API design and implementation" \
    --model main \
    --role review \
    --complexity standard \
    --skills "go-api,security" \
    --tools "todoread,todowrite,skill" \
    --non-interactive

Environment variables:
  GO_ENT_AGENT_TEMPLATE_DIR - Override built-in agent templates directory
  GO_ENT_AGENTS_DIR         - Override output agents directory`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			templateDir := getAgentTemplateDir()
			if envDir := os.Getenv("GO_ENT_AGENT_TEMPLATE_DIR"); envDir != "" {
				templateDir = envDir
			}

			templates, err := template.LoadTemplates(ctx, templateDir)
			if err != nil {
				return fmt.Errorf("load agent templates: %w", err)
			}

			if len(templates) == 0 {
				return fmt.Errorf("no agent templates found in %s", templateDir)
			}

			var cfg *WizardConfig

			if nonInteractive {
				cfg, err = RunNonInteractive(name, templateName, description, model, role, complexity, skills, tools)
			} else {
				cfg, err = RunInteractive(ctx, name, templateDir, templates)
			}

			if err != nil {
				return fmt.Errorf("configure agent: %w", err)
			}

			// Validate configuration
			if err := ValidateConfig(cfg); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			// Load template
			tpl, err := template.LoadTemplate(ctx, templateDir, cfg.TemplateName)
			if err != nil {
				return fmt.Errorf("load template: %w", err)
			}

			// Determine output directories
			metaDir := filepath.Join("pkg", "agents", "meta")
			promptsDir := filepath.Join("pkg", "agents", "prompts", "agents")

			if envDir := os.Getenv("GO_ENT_AGENTS_DIR"); envDir != "" {
				metaDir = envDir
			}

			// Create directories
			if err := os.MkdirAll(metaDir, 0o755); err != nil {
				return fmt.Errorf("create meta directory: %w", err)
			}
			if err := os.MkdirAll(promptsDir, 0o755); err != nil {
				return fmt.Errorf("create prompts directory: %w", err)
			}

			// Prepare template data
			data := cfg.ToTemplateVars()

			// Add YAML-formatted lists for skills and tools
			data["SKILLS_YAML"] = formatYAMLList(cfg.Skills)
			data["TOOLS_YAML"] = formatYAMLList(cfg.Tools)

			// Render agent.yaml
			agentYamlPath := filepath.Join(metaDir, name+".yaml")
			agentTemplate := filepath.Join(tpl.Path, "agent.yaml.tmpl")
			if err := renderTemplateFile(agentTemplate, agentYamlPath, data); err != nil {
				return fmt.Errorf("render agent.yaml: %w", err)
			}

			// Render prompt.md
			promptPath := filepath.Join(promptsDir, name+".md")
			promptTemplate := filepath.Join(tpl.Path, "prompt.md.tmpl")
			if err := renderTemplateFile(promptTemplate, promptPath, data); err != nil {
				return fmt.Errorf("render prompt.md: %w", err)
			}

			fmt.Printf("✓ Agent %q created successfully!\n\n", name)
			fmt.Printf("Files created:\n")
			fmt.Printf("  - %s\n", agentYamlPath)
			fmt.Printf("  - %s\n\n", promptPath)
			fmt.Printf("Next steps:\n")
			fmt.Printf("  1. Edit the prompt in %s\n", promptPath)
			fmt.Printf("  2. Customize %s if needed\n", agentYamlPath)
			fmt.Printf("  3. Run 'ent agent validate' to verify\n")

			return nil
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "standard", "Template to use")
	cmd.Flags().StringVar(&description, "description", "", "Agent description")
	cmd.Flags().StringVar(&model, "model", "main", "Model tier (haiku/main/opus)")
	cmd.Flags().StringVar(&role, "role", "execution", "Agent role (execution/planning/review/research)")
	cmd.Flags().StringVar(&complexity, "complexity", "standard", "Task complexity (simple/standard/complex)")
	cmd.Flags().StringVar(&skills, "skills", "", "Comma-separated skill names")
	cmd.Flags().StringVar(&tools, "tools", "todoread,todowrite", "Comma-separated tool names")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode")

	return cmd
}

func getAgentTemplateDir() string {
	return filepath.Join("pkg", "templates", "agents")
}

func renderTemplateFile(templatePath, outputPath string, data map[string]string) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	rendered, err := template.ReplacePlaceholders(string(content), data)
	if err != nil {
		return fmt.Errorf("replace placeholders: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func formatYAMLList(csvList string) string {
	if csvList == "" {
		return ""
	}

	items := strings.Split(csvList, ",")
	var result strings.Builder

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result.WriteString("\n    - ")
			result.WriteString(item)
		}
	}

	return result.String()
}
