package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/template"
)

// Template management commands: list, add, and show templates.

type templateSource string

const (
	sourceBuiltIn templateSource = "built-in"
	sourceCustom  templateSource = "custom"
)

type templateDisplay struct {
	*template.Template
	Source templateSource
}

func newListTemplatesCmd() *cobra.Command {
	var category string
	var showBuiltIn bool
	var showCustom bool

	cmd := &cobra.Command{
		Use:   "list-templates",
		Short: "List all available skill templates",
		Long: `Display a table of all available skill templates with their metadata.

Shows built-in and custom templates with name, category, source, and description.
Templates are sorted by source (built-in first) then alphabetically by name.

Examples:
  # List all templates
  ent skill list-templates

  # Filter by category
  ent skill list-templates --category go

  # Show only built-in templates
  ent skill list-templates --built-in

  # Show only custom templates
  ent skill list-templates --custom`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTemplates(category, showBuiltIn, showCustom)
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter templates by category (e.g., go, typescript, testing)")
	cmd.Flags().BoolVar(&showBuiltIn, "built-in", false, "Show only built-in templates")
	cmd.Flags().BoolVar(&showCustom, "custom", false, "Show only custom templates")

	return cmd
}

// listTemplates lists available templates with optional filtering.
// Supports filtering by category and source (built-in/custom).
func listTemplates(category string, showBuiltIn, showCustom bool) error {
	ctx := context.Background()

	builtInDir := getTemplatesPath()
	if envDir := os.Getenv("GO_ENT_TEMPLATE_DIR"); envDir != "" {
		builtInDir = envDir
	}
	customDir := getCustomTemplatesPath()

	if customDir != "" {
		if _, err := os.Stat(customDir); os.IsNotExist(err) {
			if err := os.MkdirAll(customDir, 0755); err != nil { //nolint:gosec
				return fmt.Errorf("create custom templates directory: %w", err)
			}
		}
	}

	var templates []templateDisplay

	if showBuiltIn || (!showBuiltIn && !showCustom) {
		builtIn, err := template.LoadTemplates(ctx, builtInDir)
		if err == nil {
			for _, tpl := range builtIn {
				templates = append(templates, templateDisplay{
					Template: tpl,
					Source:   sourceBuiltIn,
				})
			}
		}
	}

	if showCustom || (!showBuiltIn && !showCustom) {
		custom, err := template.LoadTemplates(ctx, customDir)
		if err == nil {
			for _, tpl := range custom {
				templates = append(templates, templateDisplay{
					Template: tpl,
					Source:   sourceCustom,
				})
			}
		}
	}

	if category != "" {
		var filtered []templateDisplay
		for _, tpl := range templates {
			if strings.EqualFold(tpl.Category, category) {
				filtered = append(filtered, tpl)
			}
		}
		templates = filtered
	}

	if len(templates) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "No templates found")
		return nil
	}

	sort.Slice(templates, func(i, j int) bool {
		if templates[i].Source != templates[j].Source {
			return templates[i].Source == sourceBuiltIn
		}
		return templates[i].Name < templates[j].Name
	})

	return printTemplatesTable(templates)
}

// printTemplatesTable prints templates in a tabular format.
func printTemplatesTable(templates []templateDisplay) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tCATEGORY\tSOURCE\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "----\t--------\t------\t-----------")

	for _, t := range templates {
		desc := truncateString(t.Description, 50)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Name, t.Category, t.Source, desc)
	}

	return w.Flush()
}

func getTemplatesPath() string {
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

func getCustomTemplatesPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".go-ent", "templates", "skills")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func newAddTemplateCmd() *cobra.Command {
	var destDir string

	cmd := &cobra.Command{
		Use:   "add-template <path>",
		Short: "Add a custom skill template to the registry",
		Long: `Add a custom template from a local directory to the template registry.

The template directory must contain:
  • template.md: The skill template with v2 format
  • config.yaml: Template metadata and prompt configuration

Validation:
  • Template structure is validated before adding
  • template.md must pass skill validation
  • config.yaml must be valid YAML

Destination:
  By default, templates are added to the user templates directory
  (~/.go-ent/templates/skills/). Use --built-in flag to add to the
  built-in directory (requires write permissions to plugins/go-ent/templates/skills/).

Examples:
  # Add template from local directory (uses user templates directory)
  ent skill add-template /path/to/my-template

  # Add template to built-in directory
  ent skill add-template /path/to/my-template --built-in /path/to/go-ent/plugins/go-ent/templates/skills/`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcPath := args[0]
			return addTemplate(srcPath, destDir)
		},
	}

	cmd.Flags().StringVar(&destDir, "built-in", "", "Add to built-in templates directory (path to plugins/go-ent/templates/skills/)")
	cmd.Flags().StringVar(&destDir, "custom", "", "Add to custom templates directory (path to user templates)")

	return cmd
}

func newShowTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-template <name>",
		Short: "Show detailed information about a skill template",
		Long: `Display detailed information about a skill template including
metadata, configuration prompts, and a preview of the template content.

Shows template details such as:
  • Name, category, and version
  • Author and description
  • Template source (built-in or custom)
  • Configuration prompts with defaults
  • Template preview (first 20 lines)

Searches both built-in and custom template directories. Shows the first
matching template found.

Examples:
  # Show details about a built-in template
  ent skill show-template go-complete

  # Show details about a custom template
  ent skill show-template my-custom-template`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return showTemplate(name)
		},
	}

	return cmd
}

// showTemplate displays detailed information about a template.
// Searches built-in directory first, then custom directory.
func showTemplate(name string) error {
	ctx := context.Background()

	builtInDir := getTemplatesPath()
	if envDir := os.Getenv("GO_ENT_TEMPLATE_DIR"); envDir != "" {
		builtInDir = envDir
	}
	builtInTpl, err := template.LoadTemplate(ctx, builtInDir, name)
	if err == nil {
		return printTemplateDetails(builtInTpl, sourceBuiltIn)
	}

	customDir := getCustomTemplatesPath()
	if customDir != "" {
		customTpl, err := template.LoadTemplate(ctx, customDir, name)
		if err == nil {
			return printTemplateDetails(customTpl, sourceCustom)
		}
	}

	return fmt.Errorf("template not found: %s", name)
}

// printTemplateDetails displays detailed template information.
// Shows metadata, configuration prompts, and template preview.
func printTemplateDetails(tpl *template.Template, source templateSource) error {
	fmt.Printf("# Template: %s\n\n", tpl.Name)

	fmt.Printf("## Metadata\n\n")
	fmt.Printf("  **Name:** %s\n", tpl.Name)
	fmt.Printf("  **Category:** %s\n", tpl.Category)
	fmt.Printf("  **Description:** %s\n", tpl.Description)
	fmt.Printf("  **Version:** %s\n", tpl.Version)
	if tpl.Author != "" {
		fmt.Printf("  **Author:** %s\n", tpl.Author)
	}
	fmt.Printf("  **Source:** %s\n", source)
	fmt.Printf("  **Path:** %s\n", tpl.Path)

	configPath := filepath.Join(tpl.Path, "config.yaml")
	cfg, err := template.ParseConfig(configPath)
	if err == nil && len(cfg.Prompts) > 0 {
		fmt.Printf("\n## Configuration Prompts\n\n")
		for _, p := range cfg.Prompts {
			required := ""
			if p.Required {
				required = " (required)"
			}
			fmt.Printf("  **%s**%s\n", p.Key, required)
			if p.Prompt != "" {
				fmt.Printf("    %s\n", p.Prompt)
			}
			if p.Default != "" {
				fmt.Printf("    Default: %s\n", p.Default)
			}
			fmt.Println()
		}
	}

	mdPath := filepath.Join(tpl.Path, "template.md")
	content, err := os.ReadFile(mdPath) //nolint:gosec
	if err == nil {
		fmt.Printf("## Template Preview (first 20 lines)\n\n")
		fmt.Println("```")

		lines := strings.Split(string(content), "\n")
		previewLines := lines
		if len(lines) > 20 {
			previewLines = lines[:20]
		}

		for _, line := range previewLines {
			fmt.Println(line)
		}

		if len(lines) > 20 {
			fmt.Printf("\n... (%d more lines)\n", len(lines)-20)
		}

		fmt.Println("```")
	}

	return nil
}

// addTemplate adds a custom template to the template registry.
// Validates template structure and contents before copying.
func addTemplate(srcPath, destDir string) error {
	absSrcPath, err := filepath.Abs(srcPath)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}

	if _, err := os.Stat(absSrcPath); os.IsNotExist(err) {
		return fmt.Errorf("template directory not found: %s", absSrcPath)
	}
	if !isDirectory(absSrcPath) {
		return fmt.Errorf("template path is not a directory: %s", absSrcPath)
	}

	configPath := filepath.Join(absSrcPath, "config.yaml")
	mdPath := filepath.Join(absSrcPath, "template.md")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("template missing required file: config.yaml")
	}
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		return fmt.Errorf("template missing required file: template.md")
	}

	cfg, err := template.ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("parse template config: %w", err)
	}

	if err := validateTemplate(absSrcPath, cfg); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	if destDir == "" {
		destDir = getCustomTemplatesPath()
	}

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil { //nolint:gosec
			return fmt.Errorf("create destination directory: %w", err)
		}
	}

	destPath := filepath.Join(destDir, cfg.Name)

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("template '%s' already exists in destination: %s", cfg.Name, destPath)
	}

	if err := copyDirectory(absSrcPath, destPath); err != nil {
		return fmt.Errorf("copy template: %w", err)
	}

	fmt.Printf("Template '%s' added successfully to: %s\n", cfg.Name, destPath)
	return nil
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func validateTemplate(tplPath string, cfg *template.TemplateConfig) error {
	mdPath := filepath.Join(tplPath, "template.md")

	parser := skill.NewParser()
	validator := skill.NewValidator()

	meta, err := parser.ParseSkillFile(mdPath)
	if err != nil {
		return fmt.Errorf("parse template.md: %w", err)
	}

	content, err := os.ReadFile(mdPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read template.md: %w", err)
	}

	result := validator.Validate(meta, string(content))

	if !result.Valid || len(result.Issues) > 0 {
		var errors []string
		for _, issue := range result.Issues {
			if issue.Severity == "error" {
				loc := issue.Rule
				if issue.Line > 0 {
					loc = fmt.Sprintf("%s:%d", loc, issue.Line)
				}
				errors = append(errors, fmt.Sprintf("  [%s] %s: %s", issue.Severity, loc, issue.Message))
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("validation issues:\n%s", strings.Join(errors, "\n"))
		}
	}

	return nil
}

func copyDirectory(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read source directory: %w", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil { //nolint:gosec
		return fmt.Errorf("create destination directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil { //nolint:gosec
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
