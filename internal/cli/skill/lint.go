package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type LintResult struct {
	Skill  string
	File   string
	Valid  bool
	Issues []skill.ValidationIssue
}

func newLintCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate skill files",
		Long:  `Validate skill files and report issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsPath := getSkillsPath()

			if len(args) == 0 {
				skills, err := findSkillFiles(skillsPath)
				if err != nil {
					return fmt.Errorf("find skill files: %w", err)
				}

				for _, skillFile := range skills {
					if err := lintSkill(skillFile, dryRun); err != nil {
						return err
					}
				}
				return nil
			}

			skillFile := args[0]
			return lintSkill(skillFile, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Check without modifying files")
	return cmd
}

func lintSkill(filePath string, dryRun bool) error {
	parser := skill.NewParser()
	validator := skill.NewValidator()

	meta, err := parser.ParseSkillFile(filePath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	result := validator.Validate(meta, string(content))

	lintResult := LintResult{
		Skill:  meta.Name,
		File:   filePath,
		Valid:  result.ErrorCount() == 0,
		Issues: result.Issues,
	}

	printResult(lintResult)
	return nil
}

func findSkillFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() != "SKILL.md" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func printResult(result LintResult) {
	fmt.Printf("Skill: %s\n", result.Skill)
	fmt.Printf("File:  %s\n", result.File)
	fmt.Printf("Valid: %v\n", result.Valid)

	if len(result.Issues) == 0 {
		fmt.Println("\nNo issues found.")
		return
	}

	fmt.Println("\nIssues:")
	for _, issue := range result.Issues {
		fmt.Printf("  %s\n", issue.String())
	}
}
