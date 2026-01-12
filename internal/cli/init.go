package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	goent "github.com/victorzhuk/go-ent"
	"github.com/victorzhuk/go-ent/internal/model"
	"github.com/victorzhuk/go-ent/internal/toolinit"
)

// newInitCmd creates the init command for setting up tool configurations
func newInitCmd() *cobra.Command {
	var (
		tool          string
		force         bool
		dryRun        bool
		update        bool     // --update flag
		updateFilter  string   // optional filter for update
		modelOverride []string // --model heavy=opus (repeatable)
	)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize tool configuration (Claude Code, OpenCode)",
		Long: `Initialize tool-specific configuration for go-ent plugin.

Generates agents, commands, and skills for the specified tool in the target directory.
Supports both Claude Code (.claude) and OpenCode (.opencode) configurations.

Auto-detection:
  If no tool is specified, detects based on existing directories:
  - .claude/  → claude
  - .opencode/ → opencode
  - none → prompts for selection

Examples:
  # Auto-detect tool based on existing directory
  go-ent init

  # Initialize Claude Code configuration
  go-ent init --tool=claude

  # Initialize OpenCode configuration
  go-ent init --tool=opencode

  # Initialize both tools
  go-ent init --tool=all

  # Initialize in specific directory
  go-ent init /path/to/project --tool=claude

  # Preview changes without writing
  go-ent init --dry-run

  # Force overwrite existing configuration
  go-ent init --force --tool=claude`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine target path
			targetPath := "."
			if len(args) > 0 {
				targetPath = args[0]
			}

			// Parse model overrides from --model flags
			modelOverrides := make(map[string]string)
			for _, override := range modelOverride {
				parts := strings.SplitN(override, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid model override format: %s (expected pattern=model)", override)
				}
				modelOverrides[parts[0]] = parts[1]
			}

			cfg := InitConfig{
				Path:           targetPath,
				Tool:           tool,
				Force:          force,
				DryRun:         dryRun,
				Verbose:        verbose,
				Update:         update,
				UpdateFilter:   updateFilter,
				ModelOverrides: modelOverrides,
			}

			return InitTools(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&tool, "tool", "", "tool to initialize (claude, opencode, all)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing configuration")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing files")
	cmd.Flags().BoolVar(&update, "update", false, "update existing configuration")
	cmd.Flags().StringVar(&updateFilter, "update-filter", "", "filter components to update (agents, skills, commands - comma separated)")
	cmd.Flags().StringArrayVar(&modelOverride, "model", nil, "override model for agents by tag pattern (e.g., heavy=opus, planning:heavy=opus)")

	return cmd
}

// InitConfig holds configuration for the init command
type InitConfig struct {
	Path           string
	Tool           string
	Force          bool
	DryRun         bool
	Verbose        bool
	Update         bool              // --update flag
	UpdateFilter   string            // --update=agents,skills
	ModelOverrides map[string]string // --model heavy=opus
}

// InitTools initializes tool configurations
func InitTools(ctx context.Context, cfg InitConfig) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Auto-detect tool if not specified
	tool := cfg.Tool
	if tool == "" {
		detected, err := detectTool(absPath)
		if err != nil {
			return err
		}
		if detected == "" {
			// No existing config, ask user
			return fmt.Errorf("no tool specified and no existing configuration found\n\nPlease specify --tool=claude or --tool=opencode")
		}
		tool = detected
		if cfg.Verbose {
			fmt.Printf("Auto-detected tool: %s\n", tool)
		}
	}

	// Validate tool
	tool = strings.ToLower(tool)
	if tool != "claude" && tool != "opencode" && tool != "all" {
		return fmt.Errorf("invalid tool: %s (must be claude, opencode, or all)", tool)
	}

	// Generate configurations
	var tools []string
	if tool == "all" {
		tools = []string{"claude", "opencode"}
	} else {
		tools = []string{tool}
	}

	for _, t := range tools {
		if err := generateToolConfig(ctx, absPath, t, cfg); err != nil {
			return fmt.Errorf("generate %s config: %w", t, err)
		}
	}

	return nil
}

// detectTool detects which tool configuration exists
func detectTool(path string) (string, error) {
	claudePath := filepath.Join(path, ".claude")
	opencodePath := filepath.Join(path, ".opencode")

	claudeExists := false
	opencodeExists := false

	if stat, err := os.Stat(claudePath); err == nil && stat.IsDir() {
		claudeExists = true
	}
	if stat, err := os.Stat(opencodePath); err == nil && stat.IsDir() {
		opencodeExists = true
	}

	// Both exist - ambiguous
	if claudeExists && opencodeExists {
		return "", fmt.Errorf("both .claude and .opencode exist\n\nPlease specify --tool=claude or --tool=opencode")
	}

	// Return detected tool
	if claudeExists {
		return "claude", nil
	}
	if opencodeExists {
		return "opencode", nil
	}

	return "", nil
}

// generateToolConfig generates configuration for a specific tool
func generateToolConfig(ctx context.Context, path, tool string, cfg InitConfig) error {
	// Create adapter
	var adapter toolinit.Adapter
	switch tool {
	case "claude":
		adapter = toolinit.NewClaudeAdapter()
	case "opencode":
		adapter = toolinit.NewOpenCodeAdapter()
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}

	// Load model configuration
	globalModelCfg, _ := model.LoadGlobal()
	projectModelCfg, _ := model.LoadProject(path)
	modelCfg := model.Merge(globalModelCfg, projectModelCfg)
	if modelCfg == nil {
		modelCfg = model.DefaultConfig()
	}

	// Prepare generation config
	genCfg := &toolinit.GenerateConfig{
		Path:           path,
		PluginFS:       goent.PluginFS,
		Force:          cfg.Force,
		DryRun:         cfg.DryRun,
		ModelOverrides: cfg.ModelOverrides,
		ModelConfig:    modelCfg,
	}

	// Print banner
	if cfg.DryRun {
		fmt.Printf("\n╔════════════════════════════════════════╗\n")
		fmt.Printf("║  DRY RUN: %s Configuration Preview  ║\n", strings.ToUpper(tool))
		fmt.Printf("╚════════════════════════════════════════╝\n\n")
	} else {
		fmt.Printf("\n╔════════════════════════════════════════╗\n")
		fmt.Printf("║  Initializing %s Configuration     ║\n", strings.ToUpper(tool))
		fmt.Printf("╚════════════════════════════════════════╝\n\n")
	}

	// Check existing configuration
	targetDir := filepath.Join(path, adapter.TargetDir())
	existingInfo, _ := toolinit.LoadEntInfo(targetDir)

	// Handle update mode
	if cfg.Update {
		if existingInfo == nil {
			fmt.Println("⚠️  No existing installation found - performing fresh install")
		} else if !toolinit.ShouldUpdate(existingInfo) && !cfg.Force {
			fmt.Printf("✅ Already up to date (version %s)\n", existingInfo.Version)
			return nil
		}
	} else if stat, err := os.Stat(targetDir); err == nil && stat.IsDir() {
		if !cfg.Force {
			return fmt.Errorf("%s already exists\n\nUse --force to overwrite or --update to update", targetDir)
		}
		if cfg.Verbose {
			fmt.Printf("⚠️  Overwriting existing configuration at %s\n\n", targetDir)
		}
	}

	// Generate configuration
	if err := adapter.Generate(ctx, genCfg); err != nil {
		return err
	}

	// Print success message
	if cfg.DryRun {
		fmt.Printf("\n✅ Preview complete\n")
		fmt.Printf("   Run without --dry-run to create files\n\n")
	} else {
		fmt.Printf("\n✅ Configuration created successfully\n")
		fmt.Printf("   Location: %s\n\n", targetDir)

		// Print summary
		if err := printSummary(adapter, targetDir); err != nil {
			if cfg.Verbose {
				fmt.Printf("⚠️  Could not generate summary: %v\n", err)
			}
		}

		// Print next steps
		printNextSteps(tool)
	}

	return nil
}

// printSummary prints a summary of generated files
func printSummary(adapter toolinit.Adapter, targetDir string) error {
	// Count files in each directory
	counts := map[string]int{
		"commands": 0,
		"command":  0,
		"agents":   0,
		"agent":    0,
		"skills":   0,
		"skill":    0,
	}

	for dir := range counts {
		dirPath := filepath.Join(targetDir, dir)
		if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
			count, _ := countFiles(dirPath)
			counts[dir] = count
		}
	}

	// Determine which directories exist
	var commandDir, agentDir, skillDir string
	if counts["commands"] > 0 {
		commandDir = "commands"
	} else if counts["command"] > 0 {
		commandDir = "command"
	}
	if counts["agents"] > 0 {
		agentDir = "agents"
	} else if counts["agent"] > 0 {
		agentDir = "agent"
	}
	if counts["skills"] > 0 {
		skillDir = "skills"
	} else if counts["skill"] > 0 {
		skillDir = "skill"
	}

	fmt.Printf("📊 Summary:\n")
	if commandDir != "" {
		fmt.Printf("   Commands: %d file(s) in %s/\n", counts[commandDir], commandDir)
	}
	if agentDir != "" {
		fmt.Printf("   Agents:   %d file(s) in %s/\n", counts[agentDir], agentDir)
	}
	if skillDir != "" {
		fmt.Printf("   Skills:   %d file(s) in %s/\n", counts[skillDir], skillDir)
	}
	fmt.Println()

	return nil
}

// countFiles counts files in a directory recursively
func countFiles(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// printNextSteps prints next steps for the user
func printNextSteps(tool string) {
	fmt.Printf("📖 Next Steps:\n\n")

	switch tool {
	case "claude":
		fmt.Printf("   1. Restart Claude Code to load the plugin\n")
		fmt.Printf("   2. Use /plan to create change proposals\n")
		fmt.Printf("   3. Available agents:\n")
		fmt.Printf("      • @planner-smoke - Quick triage (Haiku)\n")
		fmt.Printf("      • @architect - System design (Opus)\n")
		fmt.Printf("      • @planner - Detailed planning (Sonnet)\n")
		fmt.Printf("      • @decomposer - Task breakdown (Sonnet)\n")
		fmt.Printf("\n   4. For execution, use OpenCode with /task and /bug\n\n")

	case "opencode":
		fmt.Printf("   1. Restart OpenCode to load the plugin\n")
		fmt.Printf("   2. Use /task to execute tasks from registry\n")
		fmt.Printf("   3. Use /bug to fix bugs with TDD workflow\n")
		fmt.Printf("   4. Available agents:\n")
		fmt.Printf("      • @task-smoke - Simple tasks (GLM 4.7)\n")
		fmt.Printf("      • @task-heavy - Complex tasks (Kimi K2)\n")
		fmt.Printf("      • @coder - Implementation (GLM 4.7)\n")
		fmt.Printf("      • @debugger-smoke - Simple debugging (GLM 4.7)\n")
		fmt.Printf("      • @debugger-heavy - Complex debugging (Kimi K2)\n")
		fmt.Printf("      • ...and 5 more execution agents\n")
		fmt.Printf("\n   5. For planning, use Claude Code with /plan\n\n")
	}

	fmt.Printf("💡 Tip: Use --verbose flag to see detailed output\n")
	fmt.Printf("💡 Tip: Run 'go-ent init --dry-run' to preview changes\n\n")
}
