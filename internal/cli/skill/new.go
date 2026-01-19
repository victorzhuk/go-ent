package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/template"
)

// Skill creation command and helpers.

func newSkillCmd() *cobra.Command {
	var (
		templateName   string
		description    string
		category       string
		author         string
		tags           string
		nonInteractive bool
	)

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new skill from a template",
		Long: `Create a new skill from a built-in or custom template.

Interactive mode (default):
  Prompts for template selection, description, and category.
  Automatically detects category from skill name prefix (e.g., go-payment -> go)

Non-interactive mode:
  Use flags to specify all options without prompts.
  Requires --template flag at minimum.

Auto-detection:
  The command automatically detects category from skill name prefix:
  • go-*          -> go
  • typescript-*   -> typescript
  • javascript-*   -> javascript
  • test-*         -> testing
  • api-*          -> api
  • security-*     -> security
  • review-*       -> review
  • arch-*         -> arch
  • debug-*        -> debugging
  • core-*         -> core

Examples:
  # Interactive mode with auto-detected category
  go-ent skill new go-payment

  # Interactive mode with manual template selection
  go-ent skill new my-skill --template go-basic

  # Non-interactive mode with all flags
  go-ent skill new go-api \
    --template go-complete \
    --description "REST API skill with best practices" \
    --category go \
    --author "Your Name" \
    --tags "api,rest,http"

  # Non-interactive mode (category auto-detected)
  go-ent skill new test-helper \
    --template testing \
    --description "Testing helper skill"

Environment variables:
  GO_ENT_TEMPLATE_DIR - Override built-in templates directory
  GO_ENT_SKILLS_DIR  - Override output skills directory`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			templateDir := getTemplateDir()
			if envDir := os.Getenv("GO_ENT_TEMPLATE_DIR"); envDir != "" {
				templateDir = envDir
			}
			templates, err := template.LoadTemplates(ctx, templateDir)
			if err != nil {
				return fmt.Errorf("load templates: %w", err)
			}

			if len(templates) == 0 {
				return fmt.Errorf("no templates found in %s", templateDir)
			}

			var cfg *WizardConfig

			if nonInteractive {
				cfg, err = runNonInteractive(name, templateName, description, category, author, tags)
			} else {
				cfg, err = runInteractive(ctx, name, templateDir, templates, templateName, description, category)
			}

			if err != nil {
				return err
			}

			skillsDir := getSkillsPath()
			if envDir := os.Getenv("GO_ENT_SKILLS_DIR"); envDir != "" {
				skillsDir = envDir
			}
			cfg.OutputPath = filepath.Join(skillsDir, filepath.Join(strings.Split(cfg.OutputPath, string(filepath.Separator))[3:]...))

			outputPath, err := GenerateSkill(ctx, templateDir, cfg.TemplateName, cfg)
			if err != nil {
				return fmt.Errorf("generate skill: %w", err)
			}

			fmt.Printf("\n✓ Skill created successfully!\n\n")
			fmt.Printf("  Location: %s\n", outputPath)
			fmt.Printf("  Name:     %s\n", cfg.Name)
			fmt.Printf("  Category: %s\n\n", cfg.Category)
			fmt.Printf("Next steps:\n")
			fmt.Printf("  1. Review and customize the generated skill file\n")
			fmt.Printf("  2. Test with: go-ent skill info %s\n", cfg.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "", "Template name (non-interactive mode)")
	cmd.Flags().StringVar(&description, "description", "", "Skill description (required in non-interactive mode)")
	cmd.Flags().StringVar(&category, "category", "", "Skill category (auto-detected from name prefix if omitted)")
	cmd.Flags().StringVar(&author, "author", "", "Skill author")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags (e.g., 'api,rest,http')")
	cmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "Run in non-interactive mode (requires --template and --description)")

	return cmd
}

// runInteractive runs the interactive wizard for creating a skill.
// Prompts for template selection and metadata.
func runInteractive(ctx context.Context, name, templateDir string, templates []*template.Template, templateName, description, category string) (*WizardConfig, error) {
	var cfg *WizardConfig
	var err error

	selectedTemplate := templateName
	if selectedTemplate == "" {
		templatesSlice := make([]template.Template, len(templates))
		for i, t := range templates {
			templatesSlice[i] = *t
		}
		selectedTemplate, err = PromptTemplateSelection(templatesSlice)
		if err != nil {
			return nil, fmt.Errorf("select template: %w", err)
		}
	}

	cfg, err = PromptMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("prompt metadata: %w", err)
	}

	cfg.TemplateName = selectedTemplate

	if description != "" {
		cfg.Description = description
	}

	if category != "" {
		cfg.Category = category
		cfg.OutputPath = DetermineOutputPath(category, cfg.Name)
	}

	return cfg, nil
}

// runNonInteractive creates a skill configuration without prompting.
// Validates that required flags are provided and uses auto-detection for category.
func runNonInteractive(name, templateName, description, category, author, tags string) (*WizardConfig, error) {
	if templateName == "" {
		return nil, fmt.Errorf("--template flag is required in non-interactive mode")
	}

	if description == "" {
		return nil, fmt.Errorf("--description flag is required in non-interactive mode")
	}

	cfg := &WizardConfig{
		Name:         name,
		TemplateName: templateName,
		Description:  description,
	}

	detectedCategory := DetectCategory(name)
	if category == "" {
		if detectedCategory == "" {
			return nil, fmt.Errorf("--category flag required when category cannot be detected from name")
		}
		cfg.Category = detectedCategory
	} else {
		cfg.Category = category
	}

	cfg.OutputPath = DetermineOutputPath(cfg.Category, cfg.Name)

	return cfg, nil
}

func getTemplateDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "plugins/go-ent/templates/skills"
	}

	exeDir := filepath.Dir(exe)
	path := filepath.Join(exeDir, "..", "plugins", "go-ent", "templates", "skills")

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "plugins/go-ent/templates/skills"
}
