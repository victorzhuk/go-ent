package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage skills and create skills from templates",
		Long: `Skill management commands for creating, listing, and inspecting skills.

The skill command group provides:
  • Scaffold new skills from built-in or custom templates
  • List available skills and templates
  • Inspect skill details and documentation
  • Add custom templates to extend the system

Templates: Speed up skill creation with pre-built templates for different
languages and domains. Templates handle structure, examples, and best practices
so you can focus on the skill content.

Available subcommands:
  new              - Create a new skill from a template
  list             - List all available skills
  info             - Show detailed information about a skill
  list-templates   - List all available skill templates
  add-template     - Add a custom template to the registry
  show-template    - Display template details and preview

Examples:
  # List all available skills
  go-ent skill list

  # Show detailed info about a skill
  go-ent skill info go-code

  # Create a new skill using interactive wizard
  go-ent skill new go-payment

  # Create a skill in non-interactive mode
  go-ent skill new go-api --template go-basic --description "REST API skill"

  # List all available templates
  go-ent skill list-templates

  # Filter templates by category
  go-ent skill list-templates --category go

  # Show template details
  go-ent skill show-template go-complete

  # Add a custom template
  go-ent skill add-template /path/to/my-template`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newSkillCmd())
	cmd.AddCommand(newListTemplatesCmd())
	cmd.AddCommand(newAddTemplateCmd())
	cmd.AddCommand(newShowTemplateCmd())

	return cmd
}

func newListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available skills",
		Long: `Display a table of all available skills with their descriptions.

Shows all installed skills in the skills directory. Use --format to control
output style. The table format shows concise information, while detailed
format includes triggers and file locations.

Examples:
  # List all skills in table format (default)
  go-ent skill list

  # Show detailed information for all skills
  go-ent skill list --format detailed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsPath := getSkillsPath()
			registry := skill.NewRegistry()

			if err := registry.Load(skillsPath); err != nil {
				return fmt.Errorf("load skills: %w", err)
			}

			skills := registry.All()
			if len(skills) == 0 {
				_, _ = fmt.Fprintln(os.Stderr, "No skills found")
				return nil
			}

			switch format {
			case "table":
				return printSkillsTable(skills)
			case "detailed":
				return printSkillsDetailed(skills)
			default:
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, detailed)")

	return cmd
}

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about a skill",
		Long: `Display detailed information about a specific skill including
description, auto-activation triggers, and full documentation.

Shows the skill's metadata, triggers, and documentation content from
the SKILL.md file. Useful for understanding what a skill does and
when it activates automatically.

Examples:
  # Show information about the go-code skill
  go-ent skill info go-code

  # Show information about a specific skill
  go-ent skill info typescript-basic`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			skillsPath := getSkillsPath()
			registry := skill.NewRegistry()

			if err := registry.Load(skillsPath); err != nil {
				return fmt.Errorf("load skills: %w", err)
			}

			meta, err := registry.Get(name)
			if err != nil {
				return fmt.Errorf("skill not found: %s", name)
			}

			return printSkillInfo(meta)
		},
	}
}

func getSkillsPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "plugins/go-ent/skills"
	}

	exeDir := filepath.Dir(exe)
	path := filepath.Join(exeDir, "..", "plugins", "go-ent", "skills")

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "plugins/go-ent/skills"
}

func printSkillsTable(skills []skill.SkillMeta) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_ = w.Flush()

	_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "----\t-----------")

	for _, s := range skills {
		desc := extractShortDescription(s.Description)
		_, _ = fmt.Fprintf(w, "%s\t%s\n", s.Name, desc)
	}

	return nil
}

func printSkillsDetailed(skills []skill.SkillMeta) error {
	for i, s := range skills {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("# %s\n\n", s.Name)
		fmt.Printf("%s\n", s.Description)

		if len(s.Triggers) > 0 {
			fmt.Printf("\n**Triggers**: %s\n", strings.Join(s.Triggers, ", "))
		}

		fmt.Printf("\n**Location**: %s\n", s.FilePath)
	}

	return nil
}

func printSkillInfo(meta *skill.SkillMeta) error {
	fmt.Printf("# Skill: %s\n\n", meta.Name)
	fmt.Printf("## Description\n\n%s\n", meta.Description)

	if len(meta.Triggers) > 0 {
		fmt.Printf("\n## Auto-Activation Triggers\n\n")
		for _, trigger := range meta.Triggers {
			fmt.Printf("- %s\n", trigger)
		}
	}

	content, err := os.ReadFile(meta.FilePath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		inFrontmatter := false
		foundStart := false
		var contentLines []string

		for _, line := range lines {
			if strings.TrimSpace(line) == "---" {
				if !foundStart {
					foundStart = true
					inFrontmatter = true
					continue
				}
				if inFrontmatter {
					inFrontmatter = false
					continue
				}
			}
			if !inFrontmatter && foundStart {
				contentLines = append(contentLines, line)
			}
		}

		if len(contentLines) > 0 {
			fmt.Printf("\n## Documentation\n\n%s\n", strings.Join(contentLines, "\n"))
		}
	}

	fmt.Printf("\n**File**: %s\n", meta.FilePath)

	return nil
}

func extractShortDescription(desc string) string {
	if idx := strings.Index(desc, "Auto-activates for:"); idx != -1 {
		desc = strings.TrimSpace(desc[:idx])
	}

	if len(desc) > 70 {
		return desc[:67] + "..."
	}

	return desc
}
