package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

// RunConfig holds configuration for the run command.
type RunConfig struct {
	Task     string
	TaskType string
	Files    []string
	Agent    string
	Strategy string
	Budget   int
	DryRun   bool
	Verbose  bool
}

// Run executes a task with automatic agent selection.
func Run(ctx context.Context, cfg RunConfig) error {
	if cfg.Task == "" {
		return fmt.Errorf("task description required")
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzing task: %s\n", cfg.Task)
	}

	taskType := parseTaskType(cfg.TaskType)
	task := agent.Task{
		Description: cfg.Task,
		Type:        taskType,
		Files:       cfg.Files,
		Metadata:    make(map[string]interface{}),
	}

	selector := agent.NewSelector(agent.Config{
		MaxBudget:  cfg.Budget,
		StrictMode: false,
	}, &mockRegistry{})

	result, err := selector.Select(ctx, task)
	if err != nil {
		return fmt.Errorf("agent selection: %w", err)
	}

	if cfg.Agent != "" {
		result.Role = domain.AgentRole(cfg.Agent)
		result.Reason = fmt.Sprintf("manually overridden to %s", cfg.Agent)
	}

	analyzer := agent.NewComplexity()
	complexity := analyzer.Analyze(task)

	fmt.Printf("✅ Agent Selection:\n\n")
	fmt.Printf("  Role:       %s\n", result.Role)
	fmt.Printf("  Model:      %s\n", result.Model)
	fmt.Printf("  Complexity: %s (score: %d)\n", complexity.Level, complexity.Score)
	fmt.Printf("  Reason:     %s\n", result.Reason)

	if len(result.Skills) > 0 {
		fmt.Printf("  Skills:     %s\n", strings.Join(result.Skills, ", "))
	}

	if cfg.DryRun {
		fmt.Printf("\n⏸️  Dry run mode - execution skipped\n")
		return nil
	}

	fmt.Printf("\n⚠️  Execution engine not yet implemented\n")
	fmt.Printf("This command currently only performs agent selection.\n")
	fmt.Printf("Full execution will be available after the execution engine is complete.\n")

	return nil
}

func parseTaskType(t string) agent.TaskType {
	switch strings.ToLower(t) {
	case "feature":
		return agent.TaskTypeFeature
	case "bugfix", "bug":
		return agent.TaskTypeBugFix
	case "refactor":
		return agent.TaskTypeRefactor
	case "test":
		return agent.TaskTypeTest
	case "documentation", "docs":
		return agent.TaskTypeDocumentation
	case "architecture", "arch":
		return agent.TaskTypeArchitecture
	default:
		return agent.TaskTypeFeature
	}
}

type mockRegistry struct{}

func (m *mockRegistry) MatchForContext(ctx domain.SkillContext) []string {
	return []string{}
}
