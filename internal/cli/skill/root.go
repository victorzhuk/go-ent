package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage skills",
		Long:  `List, inspect, and validate skills.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newLintCmd())

	return cmd
}

func newListCmd() *cobra.Command {
	var format string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list [query]",
		Short: "List all available skills",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsPath := getSkillsPath()
			registry := skill.NewRegistry()

			if err := registry.Load(skillsPath); err != nil {
				return fmt.Errorf("load skills: %w", err)
			}

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			if query == "" {
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
			}

			matches := registry.FindMatchingSkills(query)
			if len(matches) == 0 {
				_, _ = fmt.Fprintln(os.Stderr, "No matching skills found")
				return nil
			}

			switch format {
			case "table":
				return printMatchesTable(matches, verbose)
			case "detailed":
				return printMatchesDetailed(matches, verbose)
			default:
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, detailed)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show match reasons")

	return cmd
}

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about a skill",
		Args:  cobra.ExactArgs(1),
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
		return "pkg/skills"
	}

	exeDir := filepath.Dir(exe)
	path := filepath.Join(exeDir, "..", "plugins", "go-ent", "skills")

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "pkg/skills"
}

func printSkillsTable(skills []*skilldomain.Info) error {
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

func printSkillsDetailed(skills []*skilldomain.Info) error {
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

func printSkillInfo(meta *skilldomain.Info) error {
	fmt.Printf("# Skill: %s\n\n", meta.Name)
	fmt.Printf("## Description\n\n%s\n", meta.Description)

	if len(meta.Triggers) > 0 {
		fmt.Printf("\n## Auto-Activation Triggers\n\n")
		for _, trigger := range meta.Triggers {
			fmt.Printf("- %s\n", trigger)
		}
	}

	content, err := os.ReadFile(meta.FilePath) //nolint:gosec // controlled path from skill registry
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

func printMatchesTable(matches []skill.MatchResult, verbose bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "SCORE\tNAME\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "-----\t----\t-----------")

	for _, m := range matches {
		if m.Skill == nil {
			continue
		}
		desc := extractShortDescription(m.Skill.Description)
		score := int(m.Score * 100)
		_, _ = fmt.Fprintf(w, "%d%%\t%s\t%s\n", score, m.Skill.Name, desc)
	}

	return w.Flush()
}

func printMatchesDetailed(matches []skill.MatchResult, verbose bool) error {
	for i, m := range matches {
		if m.Skill == nil {
			continue
		}
		if i > 0 {
			fmt.Println()
		}

		score := int(m.Score * 100)
		fmt.Printf("# %s (%d%% match)\n\n", m.Skill.Name, score)
		fmt.Printf("%s\n", m.Skill.Description)

		if verbose && len(m.MatchedBy) > 0 {
			fmt.Printf("\n**Match Reasons**:\n")
			for _, reason := range m.MatchedBy {
				weight := int(reason.Weight * 100)
				fmt.Printf("- %s: %s (weight: %d%%)\n", reason.Type, reason.Value, weight)
			}
		}

		if len(m.Skill.Triggers) > 0 {
			fmt.Printf("\n**Triggers**: %s\n", strings.Join(m.Skill.Triggers, ", "))
		}

		fmt.Printf("\n**Location**: %s\n", m.Skill.FilePath)
	}

	return nil
}
