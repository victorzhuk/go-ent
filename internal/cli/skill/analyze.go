package skill

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
)

func newAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze",
		Short: "List and analyze skills in registry",
		Long:  `List all skills and their basic validation status.`,
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

			sort.Slice(skills, func(i, j int) bool {
				return skills[i].Name < skills[j].Name
			})

			fmt.Printf("Found %d skills\n\n", len(skills))
			for _, s := range skills {
				fmt.Printf("• %s\n", s.Name)
				fmt.Printf("  Description: %s\n", s.Description)
				fmt.Printf("  Category: %s\n", s.Category)
				fmt.Printf("  Triggers: %d\n", len(s.Triggers))
				fmt.Println()
			}

			return nil
		},
	}
}
