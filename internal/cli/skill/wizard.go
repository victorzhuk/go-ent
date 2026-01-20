package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/victorzhuk/go-ent/internal/template"
)

// Interactive wizard for creating skills from templates.

type WizardConfig struct {
	TemplateName string
	Name         string
	Description  string
	OutputPath   string
	Category     string
}

// CategoryPrefixes maps skill name prefixes to their detected categories.
var categoryPrefixes = map[string]string{
	"go-":           "go",
	"typescript-":   "typescript",
	"ts-":           "typescript",
	"python-":       "python",
	"py-":           "python",
	"rust-":         "rust",
	"java-":         "java",
	"js-":           "javascript",
	"javascript-":   "javascript",
	"api-":          "api",
	"rest-":         "api",
	"graphql-":      "api",
	"db-":           "database",
	"sql-":          "database",
	"database-":     "database",
	"test-":         "testing",
	"testing-":      "testing",
	"spec-":         "testing",
	"sec-":          "security",
	"security-":     "security",
	"auth-":         "security",
	"review-":       "review",
	"audit-":        "review",
	"arch-":         "arch",
	"architecture-": "arch",
	"debug-":        "debugging",
	"debugging-":    "debugging",
	"core-":         "core",
	"ops-":          "ops",
	"devops-":       "ops",
}

// DetectCategory determines the skill category from the skill name prefix.
// Returns empty string if no known prefix is found.
func DetectCategory(skillName string) string {
	skillName = strings.ToLower(strings.TrimSpace(skillName))

	for prefix, category := range categoryPrefixes {
		if strings.HasPrefix(skillName, prefix) {
			return category
		}
	}

	return ""
}

// PromptCategory presents an interactive prompt for selecting a skill category.
func PromptCategory() (string, error) {
	categories := []string{
		"go",
		"typescript",
		"javascript",
		"python",
		"rust",
		"java",
		"api",
		"database",
		"testing",
		"security",
		"review",
		"arch",
		"debugging",
		"core",
		"ops",
	}

	q := &survey.Select{
		Message: "Select skill category:",
		Options: categories,
		Help:    "The category determines where the skill will be stored",
	}

	var selected string
	if err := survey.AskOne(q, &selected); err != nil {
		return "", fmt.Errorf("prompt category: %w", err)
	}

	return selected, nil
}

// DetermineOutputPath generates the output path for a skill file based on category and name.
// Returns a path in the format: plugins/go-ent/skills/{category}/{skillName}/SKILL.md
func DetermineOutputPath(category, skillName string) string {
	if category == "" {
		category = "core"
	}
	return fmt.Sprintf("plugins/go-ent/skills/%s/%s/SKILL.md", category, skillName)
}

// PromptTemplateSelection presents an interactive prompt for selecting a template.
// Returns the name of the selected template.
func PromptTemplateSelection(templates []template.Template) (string, error) {
	if len(templates) == 0 {
		return "", fmt.Errorf("no templates available")
	}

	var options []string
	for _, t := range templates {
		option := fmt.Sprintf("%s - %s", t.Name, t.Description)
		options = append(options, option)
	}

	q := &survey.Select{
		Message: "Choose a template:",
		Options: options,
	}

	var selected string
	if err := survey.AskOne(q, &selected); err != nil {
		return "", fmt.Errorf("prompt template: %w", err)
	}

	for _, t := range templates {
		if selected == fmt.Sprintf("%s - %s", t.Name, t.Description) {
			return t.Name, nil
		}
	}

	return "", fmt.Errorf("template not found")
}

// PromptMetadata collects skill metadata through interactive prompts.
// Uses auto-detection for category from the skill name prefix.
func PromptMetadata(name string) (*WizardConfig, error) {
	cfg := &WizardConfig{}

	if name == "" {
		q := &survey.Input{
			Message: "Skill name:",
			Help:    "The name of the skill (e.g., go-payment, typescript-api)",
		}
		if err := survey.AskOne(q, &cfg.Name, survey.WithValidator(survey.Required)); err != nil {
			return nil, fmt.Errorf("prompt name: %w", err)
		}
	} else {
		cfg.Name = name
	}

	cfg.Category = DetectCategory(cfg.Name)

	if cfg.Category == "" {
		category, err := PromptCategory()
		if err != nil {
			return nil, err
		}
		cfg.Category = category
	} else {
		confirmQ := &survey.Confirm{
			Message: fmt.Sprintf("Detected category: %s. Is this correct?", cfg.Category),
			Default: true,
		}
		var confirmed bool
		if err := survey.AskOne(confirmQ, &confirmed); err != nil {
			return nil, fmt.Errorf("confirm category: %w", err)
		}

		if !confirmed {
			category, err := PromptCategory()
			if err != nil {
				return nil, err
			}
			cfg.Category = category
		}
	}

	q := &survey.Input{
		Message: "Description:",
		Help:    "A short description of what this skill does",
	}
	if err := survey.AskOne(q, &cfg.Description, survey.WithValidator(survey.Required)); err != nil {
		return nil, fmt.Errorf("prompt description: %w", err)
	}

	cfg.OutputPath = DetermineOutputPath(cfg.Category, cfg.Name)

	return cfg, nil
}

// GenerateSkill creates a new skill from a template with the provided configuration.
// Loads the template, replaces placeholders, writes the skill file, and validates it.
// Returns the path to the generated skill file.
func GenerateSkill(ctx context.Context, templateDir, templateName string, cfg *WizardConfig) (string, error) {
	tpl, err := template.LoadTemplate(ctx, templateDir, templateName)
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}

	templatePath := tpl.Path + "/template.md"
	templateContent, err := os.ReadFile(templatePath) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("read template file: %w", err)
	}

	data := map[string]string{
		"SKILL_NAME":        cfg.Name,
		"DESCRIPTION":       cfg.Description,
		"SKILL_DESCRIPTION": cfg.Description,
		"CATEGORY":          cfg.Category,
		"VERSION":           "1.0.0",
		"AUTHOR":            "go-ent",
		"TAGS":              cfg.Category,
	}

	generatedContent, err := template.ReplacePlaceholders(string(templateContent), data)
	if err != nil {
		return "", fmt.Errorf("replace placeholders: %w", err)
	}

	outputDir := filepath.Dir(cfg.OutputPath)
	if _, err := os.Stat(cfg.OutputPath); err == nil {
		return "", fmt.Errorf("skill file already exists: %s", cfg.OutputPath)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil { //nolint:gosec
		return "", fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(cfg.OutputPath, []byte(generatedContent), 0644); err != nil { //nolint:gosec
		return "", fmt.Errorf("write skill file: %w", err)
	}

	if err := ValidateGeneratedSkill(cfg.OutputPath); err != nil {
		return "", fmt.Errorf("validate generated skill: %w", err)
	}

	return cfg.OutputPath, nil
}
