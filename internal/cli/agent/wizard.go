package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/victorzhuk/go-ent/internal/template"
)

// WizardConfig holds configuration from the agent creation wizard
type WizardConfig struct {
	TemplateName string
	Name         string
	Description  string
	Model        string
	Role         string
	Complexity   string
	Skills       string
	Tools        string
	OutputPath   string
}

// PromptModel presents an interactive prompt for selecting an AI model tier
func PromptModel() (string, error) {
	models := []string{"fast", "main", "heavy"}
	var model string
	prompt := &survey.Select{
		Message: "Select model tier:",
		Options: models,
		Default: "main",
		Description: func(value string, index int) string {
			switch value {
			case "fast":
				return "Fast, cost-effective (simple tasks)"
			case "main":
				return "Balanced performance (standard tasks)"
			case "heavy":
				return "Maximum capability (complex tasks)"
			default:
				return ""
			}
		},
	}
	if err := survey.AskOne(prompt, &model); err != nil {
		return "", err
	}
	return model, nil
}

// PromptRole presents an interactive prompt for selecting an agent role
func PromptRole() (string, error) {
	roles := []string{"execution", "planning", "validation", "research"}
	var role string
	prompt := &survey.Select{
		Message: "Select agent role:",
		Options: roles,
		Default: "execution",
		Description: func(value string, index int) string {
			switch value {
			case "execution":
				return "Implements features and writes code"
			case "planning":
				return "Plans tasks and strategies"
			case "validation":
				return "Reviews code and validates quality"
			case "research":
				return "Researches and analyzes codebase"
			default:
				return ""
			}
		},
	}
	if err := survey.AskOne(prompt, &role); err != nil {
		return "", err
	}
	return role, nil
}

// PromptComplexity presents an interactive prompt for selecting task complexity
func PromptComplexity() (string, error) {
	complexities := []string{"simple", "standard", "complex"}
	var complexity string
	prompt := &survey.Select{
		Message: "Select task complexity level:",
		Options: complexities,
		Default: "standard",
		Description: func(value string, index int) string {
			switch value {
			case "simple":
				return "Quick, straightforward tasks"
			case "standard":
				return "Normal development tasks"
			case "complex":
				return "Multi-step, architectural tasks"
			default:
				return ""
			}
		},
	}
	if err := survey.AskOne(prompt, &complexity); err != nil {
		return "", err
	}
	return complexity, nil
}

// RunInteractive runs the interactive agent creation wizard
func RunInteractive(ctx context.Context, name, templateDir string, templates []*template.Template) (*WizardConfig, error) {
	cfg := &WizardConfig{
		Name:         name,
		TemplateName: "standard",
	}

	// Template selection
	if len(templates) > 1 {
		templateNames := make([]string, len(templates))
		for i, t := range templates {
			templateNames[i] = t.Name
		}
		prompt := &survey.Select{
			Message: "Select agent template:",
			Options: templateNames,
			Default: "standard",
		}
		if err := survey.AskOne(prompt, &cfg.TemplateName); err != nil {
			return nil, err
		}
	}

	// Description
	descPrompt := &survey.Input{
		Message: "Agent description:",
		Default: fmt.Sprintf("Custom %s agent", name),
	}
	if err := survey.AskOne(descPrompt, &cfg.Description); err != nil {
		return nil, err
	}

	// Model tier
	model, err := PromptModel()
	if err != nil {
		return nil, err
	}
	cfg.Model = model

	// Role
	role, err := PromptRole()
	if err != nil {
		return nil, err
	}
	cfg.Role = role

	// Complexity
	complexity, err := PromptComplexity()
	if err != nil {
		return nil, err
	}
	cfg.Complexity = complexity

	// Skills
	skillsPrompt := &survey.Input{
		Message: "Skills (comma-separated, leave empty for none):",
		Help:    "e.g., go-code,go-db",
	}
	if err := survey.AskOne(skillsPrompt, &cfg.Skills); err != nil {
		return nil, err
	}

	// Tools
	toolsPrompt := &survey.Input{
		Message: "Additional tools (comma-separated):",
		Default: "todoread,todowrite",
		Help:    "Common tools: todoread, todowrite, skill, list",
	}
	if err := survey.AskOne(toolsPrompt, &cfg.Tools); err != nil {
		return nil, err
	}

	return cfg, nil
}

// RunNonInteractive creates agent config from CLI flags
func RunNonInteractive(name, templateName, description, model, role, complexity, skills, tools string) (*WizardConfig, error) {
	if templateName == "" {
		return nil, fmt.Errorf("--template is required in non-interactive mode")
	}

	if description == "" {
		description = fmt.Sprintf("Custom %s agent", name)
	}

	if model == "" {
		model = "main"
	}

	if role == "" {
		role = "execution"
	}

	if complexity == "" {
		complexity = "standard"
	}

	if tools == "" {
		tools = "todoread,todowrite"
	}

	return &WizardConfig{
		TemplateName: templateName,
		Name:         name,
		Description:  description,
		Model:        model,
		Role:         role,
		Complexity:   complexity,
		Skills:       skills,
		Tools:        tools,
	}, nil
}

// ValidateConfig validates the wizard configuration
func ValidateConfig(cfg *WizardConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if cfg.TemplateName == "" {
		return fmt.Errorf("template name is required")
	}

	// Validate model
	validModels := map[string]bool{"fast": true, "main": true, "heavy": true}
	if !validModels[cfg.Model] {
		return fmt.Errorf("invalid model %q, must be one of: fast, main, heavy", cfg.Model)
	}

	// Validate role
	validRoles := map[string]bool{"execution": true, "planning": true, "validation": true, "research": true}
	if !validRoles[cfg.Role] {
		return fmt.Errorf("invalid role %q, must be one of: execution, planning, validation, research", cfg.Role)
	}

	// Validate complexity
	validComplexity := map[string]bool{"simple": true, "standard": true, "complex": true}
	if !validComplexity[cfg.Complexity] {
		return fmt.Errorf("invalid complexity %q, must be one of: simple, standard, complex", cfg.Complexity)
	}

	return nil
}

// ToTemplateVars converts wizard config to template variables
func (cfg *WizardConfig) ToTemplateVars() map[string]string {
	vars := map[string]string{
		"AGENT_NAME":  cfg.Name,
		"DESCRIPTION": cfg.Description,
		"MODEL":       cfg.Model,
		"ROLE":        cfg.Role,
		"COMPLEXITY":  cfg.Complexity,
		"SKILLS":      strings.TrimSpace(cfg.Skills),
		"TOOLS":       strings.TrimSpace(cfg.Tools),
	}
	return vars
}
