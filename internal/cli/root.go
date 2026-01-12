package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/version"
)

var (
	cfgFile string
	verbose bool
)

// NewRootCmd creates the root command for the go-ent CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go-ent",
		Short: "Enterprise Go development toolkit",
		Long: `go-ent is an enterprise Go development toolkit with multi-agent workflows,
spec-driven development, and intelligent task execution.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .go-ent/config.yaml)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newAgentCmd())
	cmd.AddCommand(newSkillCmd())
	cmd.AddCommand(newSpecCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newModelCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			fmt.Printf("go-ent %s\n", version.String())
			fmt.Printf("  go: %s\n", v.GoVersion)
			if v.VCSRef != "unknown" && v.VCSRef != "" {
				fmt.Printf("  ref: %s\n", v.VCSRef)
			}
		},
	}
}

func newRunCmd() *cobra.Command {
	var (
		agent    string
		taskType string
		files    []string
		strategy string
		budget   int
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "run <task-description>",
		Short: "Execute a task with automatic agent selection",
		Long: `Execute a task with automatic agent selection based on complexity analysis.

The run command analyzes the task description and selects the appropriate agent
(architect, dev, tester, etc.) and model (opus, sonnet, haiku) based on:
  - Task complexity
  - Required skills
  - Budget constraints
  - Task type (feature, bugfix, refactor, etc.)

Examples:
  # Simple task with auto-selection
  go-ent run "add logging to user service"

  # Override agent selection
  go-ent run --agent architect "design new API endpoint"

  # Specify task type
  go-ent run --type bugfix "fix memory leak in cache"

  # Multiple files
  go-ent run --files repo.go,service.go "refactor user repository"

  # Dry run (selection only, no execution)
  go-ent run --dry-run "implement rate limiting"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := RunConfig{
				Task:     args[0],
				TaskType: taskType,
				Files:    files,
				Agent:    agent,
				Strategy: strategy,
				Budget:   budget,
				DryRun:   dryRun,
				Verbose:  verbose,
			}
			return Run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "override agent selection (architect, dev, tester, etc.)")
	cmd.Flags().StringVar(&taskType, "type", "", "task type (feature, bugfix, refactor, test, documentation, architecture)")
	cmd.Flags().StringSliceVar(&files, "files", nil, "files involved in the task")
	cmd.Flags().StringVar(&strategy, "strategy", "", "execution strategy (quick, standard, thorough)")
	cmd.Flags().IntVar(&budget, "budget", 0, "maximum token budget (0 = unlimited)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show agent selection without executing")

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
