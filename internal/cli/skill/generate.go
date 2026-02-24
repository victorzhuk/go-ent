package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/generator"
)

var (
	generateToolsFlag []string
	generateNameFlag  string
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate skills for configured tools",
		Long: `Generate skill output for configured tools.

This command reads skill sources from pkg/skills/ (embedded in the binary)
and generates tool-specific output in .claude/skills/ or .opencode/skills/.

For Claude: keeps all fields (version, author, tags, etc.)
For OpenCode: strips Claude-specific fields, keeping only name, description, and triggers

Examples:
  ent skill generate                       # Generate all skills for detected runtime
  ent skill generate --tools=claude        # Generate for Claude only
  ent skill generate --name=go/go-code     # Generate specific skill
`,
		RunE: runSkillGenerate,
	}

	cmd.Flags().StringSliceVar(&generateToolsFlag, "tools", nil, "Target tools (claude,opencode)")
	cmd.Flags().StringVar(&generateNameFlag, "name", "", "Generate specific skill by category/name (e.g., go/go-code)")

	return cmd
}

func runSkillGenerate(cmd *cobra.Command, args []string) error {
	tools := generateToolsFlag
	if len(tools) == 0 {
		if detected := config.DetectRuntime("."); detected != "" {
			tools = []string{detected}
		}
	}

	if len(tools) == 0 {
		tools = []string{"claude"}
	}

	var targets []generator.Target
	for _, tool := range tools {
		switch tool {
		case "claude":
			targets = append(targets, generator.NewClaudeTarget(".claude/skills"))
		case "opencode":
			targets = append(targets, generator.NewOpenCodeTarget(".opencode/skills"))
		case "openspec":
			continue
		default:
			return fmt.Errorf("unknown tool: %s", tool)
		}
	}

	if len(targets) == 0 {
		return fmt.Errorf("no valid tools configured")
	}

	if generateNameFlag != "" {
		parts := filepath.SplitList(generateNameFlag)
		if len(parts) != 2 {
			category := filepath.Dir(generateNameFlag)
			name := filepath.Base(generateNameFlag)
			if category == "." || category == "" {
				return fmt.Errorf("invalid skill name format, expected: category/name (e.g., go/go-code)")
			}
			parts = []string{category, name}
		}

		category, name := parts[0], parts[1]
		if err := generateSkill(targets, category, name); err != nil {
			return fmt.Errorf("generate skill %s/%s: %w", category, name, err)
		}
	} else {
		if err := generateAllSkills(targets); err != nil {
			return fmt.Errorf("generate skills: %w", err)
		}
	}

	fmt.Println("\n✅ Skill generation complete!")
	return nil
}

func generateAllSkills(targets []generator.Target) error {
	skills, err := generator.ListSkills("skills")
	if err != nil {
		return fmt.Errorf("list skills: %w", err)
	}

	count := 0
	for category, names := range skills {
		for _, name := range names {
			if err := generateSkill(targets, category, name); err != nil {
				return fmt.Errorf("generate %s/%s: %w", category, name, err)
			}
			count++
		}
	}

	fmt.Printf("Generated %d skills\n", count)
	return nil
}

func generateSkill(targets []generator.Target, category, name string) error {
	skill, err := generator.LoadSkillSource("skills", category, name)
	if err != nil {
		return fmt.Errorf("load skill: %w", err)
	}

	for _, target := range targets {
		output, err := target.GenerateSkill(skill)
		if err != nil {
			return fmt.Errorf("generate %s target: %w", target.Name(), err)
		}

		outputPath := target.SkillOutputPath(category, name)
		if err := writeSkillOutput(outputPath, output); err != nil {
			return fmt.Errorf("write %s output: %w", target.Name(), err)
		}

		fmt.Printf("Generated %s/%s → %s\n", category, name, outputPath)
	}

	return nil
}

func writeSkillOutput(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
