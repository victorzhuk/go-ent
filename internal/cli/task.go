package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/spec"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage OpenSpec tasks",
		Long:  "List, show, and execute tasks from the OpenSpec registry",
	}

	cmd.AddCommand(newTaskNextCmd())
	cmd.AddCommand(newTaskShowCmd())
	cmd.AddCommand(newTaskRunCmd())
	cmd.AddCommand(newTaskCompleteCmd())
	cmd.AddCommand(newTaskListCmd())

	return cmd
}

func newTaskNextCmd() *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:   "next",
		Short: "Show next unblocked task(s)",
		Long:  "Show the next unblocked task(s) from the registry, sorted by priority",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskNext(cmd.Context(), count)
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 1, "number of tasks to show")
	return cmd
}

func newTaskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <task-id>",
		Short: "Show task details with context",
		Long:  "Show detailed information about a task including state.md context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskShow(cmd.Context(), args[0])
		},
	}
	return cmd
}

func newTaskRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <task-id>",
		Short: "Execute task with agent workflow",
		Long:  "Display agent workflow recommendations for task execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskRun(cmd.Context(), args[0])
		},
	}
	return cmd
}

func newTaskCompleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "complete <task-id>",
		Short: "Mark task as complete in tasks.md and regenerate state",
		Long:  "Update task checkbox in tasks.md, sync BoltDB, and regenerate state.md files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskComplete(cmd.Context(), args[0], yes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}

func newTaskListCmd() *cobra.Command {
	var (
		status    string
		changeID  string
		unblocked bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks with filters",
		Long:  "List tasks from the registry with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskList(cmd.Context(), status, changeID, unblocked)
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status (pending, in_progress, completed, blocked, skipped)")
	cmd.Flags().StringVarP(&changeID, "change", "c", "", "filter by change ID")
	cmd.Flags().BoolVarP(&unblocked, "unblocked", "u", false, "show only unblocked tasks")

	return cmd
}

func runTaskNext(ctx context.Context, count int) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	store := spec.NewStore(cwd)
	registryStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return fmt.Errorf("create registry store: %w", err)
	}
	defer func() { _ = registryStore.Close() }()

	result, err := registryStore.NextTask(count)
	if err != nil {
		return fmt.Errorf("get next task: %w", err)
	}

	if result.Recommended == nil {
		fmt.Println("No unblocked tasks available")
		if result.BlockedCount > 0 {
			fmt.Printf("(%d tasks blocked)\n", result.BlockedCount)
		}
		return nil
	}

	fmt.Printf("══════════════════════════════════════════\n")
	fmt.Printf("RECOMMENDED TASK\n")
	fmt.Printf("══════════════════════════════════════════\n\n")

	printTaskDetails(result.Recommended)

	if result.Reason != "" {
		fmt.Printf("📝 Reason: %s\n\n", result.Reason)
	}

	if len(result.Alternatives) > 0 {
		fmt.Printf("Alternatives:\n")
		for i, alt := range result.Alternatives {
			fmt.Printf("  %d. %s - %s (%s)\n", i+1, alt.ID.String(), alt.Content, alt.Priority)
		}
		fmt.Println()
	}

	fmt.Printf("Run: ent task run %s\n", result.Recommended.ID.String())
	return nil
}

func runTaskShow(ctx context.Context, taskIDStr string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	store := spec.NewStore(cwd)
	registryStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return fmt.Errorf("create registry store: %w", err)
	}
	defer func() { _ = registryStore.Close() }()

	taskID, err := parseTaskID(taskIDStr)
	if err != nil {
		return fmt.Errorf("parse task ID: %w", err)
	}

	task, err := registryStore.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	fmt.Printf("══════════════════════════════════════════\n")
	fmt.Printf("TASK: %s\n", task.ID.String())
	fmt.Printf("══════════════════════════════════════════\n\n")

	printTaskDetails(task)

	printDependencies(registryStore, task)

	printStateContext(store, task.ID.ChangeID)

	return nil
}

func runTaskRun(ctx context.Context, taskIDStr string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	store := spec.NewStore(cwd)
	registryStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return fmt.Errorf("create registry store: %w", err)
	}
	defer func() { _ = registryStore.Close() }()

	taskID, err := parseTaskID(taskIDStr)
	if err != nil {
		return fmt.Errorf("parse task ID: %w", err)
	}

	task, err := registryStore.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	fmt.Printf("══════════════════════════════════════════\n")
	fmt.Printf("TASK EXECUTION: %s\n", task.ID.String())
	fmt.Printf("══════════════════════════════════════════\n\n")

	printTaskDetails(task)

	fmt.Printf("\n══════════════════════════════════════════\n")
	fmt.Printf("AGENT WORKFLOW (ACP)\n")
	fmt.Printf("══════════════════════════════════════════\n\n")

	printAgentWorkflow(task)

	fmt.Printf("\n══════════════════════════════════════════\n")
	fmt.Printf("NEXT STEPS\n")
	fmt.Printf("══════════════════════════════════════════\n\n")

	fmt.Printf("1. Assess complexity and delegate to @ent:task-fast\n")
	fmt.Printf("2. If complex, escalate to @ent:task-heavy for deep analysis\n")
	fmt.Printf("3. Implement with @ent:coder\n")
	if isNonTrivialTask(task) {
		fmt.Printf("4. Request review from @ent:reviewer\n")
	}
	fmt.Printf("5. Validate with @ent:tester\n")
	fmt.Printf("6. Acceptance verification with @ent:acceptor\n")
	fmt.Printf("7. Mark complete: ent task run %s --complete\n", task.ID.String())

	return nil
}

func runTaskList(ctx context.Context, status, changeID string, unblocked bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	store := spec.NewStore(cwd)
	registryStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return fmt.Errorf("create registry store: %w", err)
	}
	defer func() { _ = registryStore.Close() }()

	filter := spec.TaskFilter{}
	if status != "" {
		filter.Status = spec.RegistryTaskStatus(status)
	}
	if changeID != "" {
		filter.ChangeID = changeID
	}
	if unblocked {
		filter.Unblocked = true
	}

	tasks, err := registryStore.ListTasks(filter)
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	for _, task := range tasks {
		fmt.Printf("  %s - %s\n", task.ID.String(), task.Content)
		fmt.Printf("    Status: %s | Priority: %s\n", task.Status, task.Priority)
		if len(task.DependsOn) > 0 {
			depIDs := make([]string, len(task.DependsOn))
			for i, dep := range task.DependsOn {
				depIDs[i] = dep.String()
			}
			fmt.Printf("    Depends: %s\n", strings.Join(depIDs, ", "))
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d tasks\n", len(tasks))
	return nil
}

func runTaskComplete(ctx context.Context, taskIDStr string, yes bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	store := spec.NewStore(cwd)
	registryStore, err := spec.NewRegistryStore(store)
	if err != nil {
		return fmt.Errorf("create registry store: %w", err)
	}
	defer func() { _ = registryStore.Close() }()

	taskID, err := parseTaskID(taskIDStr)
	if err != nil {
		return fmt.Errorf("parse task ID: %w", err)
	}

	task, err := registryStore.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if task.Status == spec.RegStatusCompleted {
		fmt.Printf("Task %s is already marked as completed\n", taskID.String())
		return nil
	}

	tasksPath := filepath.Join(store.SpecPath(), "changes", taskID.ChangeID, "tasks.md")

	content, err := os.ReadFile(tasksPath) // #nosec G304 -- controlled path from trusted sources
	if err != nil {
		return fmt.Errorf("read tasks.md: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	taskPattern := regexp.MustCompile(`^[-*]\s+\[([ xX])\]\s+([0-9]+\.[0-9]+)\s+(.+)$`)

	var targetLine int
	var oldLine string
	var newLine string
	found := false

	for i, line := range lines {
		matches := taskPattern.FindStringSubmatch(line)
		if len(matches) >= 4 && matches[2] == taskID.TaskNum {
			targetLine = i + 1
			oldLine = line
			found = true

			checkbox := matches[1]
			taskNum := matches[2]
			taskDesc := strings.TrimSpace(matches[3])

			if checkbox == "x" || checkbox == "X" {
				fmt.Printf("Task %s is already marked as completed\n", taskID.String())
				return nil
			}

			completionDate := time.Now().Format("2006-01-02")
			newLine = fmt.Sprintf("- [x] %s %s ✓ %s", taskNum, taskDesc, completionDate)
			break
		}
	}

	if !found {
		return fmt.Errorf("task %s not found in tasks.md", taskID.String())
	}

	fmt.Printf("══════════════════════════════════════════\n")
	fmt.Printf("TASK COMPLETION\n")
	fmt.Printf("══════════════════════════════════════════\n\n")

	fmt.Printf("Task: %s\n", taskID.String())
	fmt.Printf("Content: %s\n\n", task.Content)

	fmt.Printf("File: %s:%d\n\n", tasksPath, targetLine)

	fmt.Printf("BEFORE:\n")
	fmt.Printf("  %s\n\n", oldLine)

	fmt.Printf("AFTER:\n")
	fmt.Printf("  %s\n\n", newLine)

	if !yes {
		fmt.Print("Update tasks.md and regenerate state.md? [y/N]: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() || strings.ToLower(scanner.Text()) != "y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	lines[targetLine-1] = newLine

	originalContent := string(content)
	newContent := strings.Join(lines, "\n")

	if err := os.WriteFile(tasksPath, []byte(newContent), 0644); err != nil { //nolint:gosec
		return fmt.Errorf("write tasks.md: %w", err)
	}

	stateStore := registryStore.StateStore()

	if err := stateStore.SyncFromTasksMd(); err != nil {
		_ = os.WriteFile(tasksPath, []byte(originalContent), 0644) //nolint:gosec
		return fmt.Errorf("sync from tasks.md (rolled back tasks.md): %w", err)
	}

	changeStatePath := filepath.Join(store.SpecPath(), "changes", taskID.ChangeID, "state.md")
	if err := stateStore.WriteChangeStateMd(taskID.ChangeID, changeStatePath); err != nil {
		_ = os.WriteFile(tasksPath, []byte(originalContent), 0644) //nolint:gosec
		return fmt.Errorf("write change state.md (rolled back tasks.md): %w", err)
	}

	rootStatePath := filepath.Join(store.SpecPath(), "state.md")
	if err := stateStore.WriteRootStateMd(rootStatePath); err != nil {
		_ = os.WriteFile(tasksPath, []byte(originalContent), 0644) //nolint:gosec
		return fmt.Errorf("write root state.md (rolled back tasks.md): %w", err)
	}

	fmt.Printf("✓ Task %s marked complete\n", taskID.String())
	fmt.Printf("✓ tasks.md updated\n")
	fmt.Printf("✓ BoltDB synced\n")
	fmt.Printf("✓ state.md regenerated (change and root)\n")

	return nil
}

func printTaskDetails(task *spec.RegistryTask) {
	fmt.Printf("📋 Task: %s\n", task.Content)
	fmt.Printf("   Change: %s\n", task.ID.ChangeID)
	fmt.Printf("   Status: %s\n", task.Status)
	fmt.Printf("   Priority: %s\n", task.Priority)

	if len(task.DependsOn) > 0 {
		fmt.Printf("   Dependencies: %d\n", len(task.DependsOn))
	}

	if len(task.BlockedBy) > 0 {
		fmt.Printf("   Blocked by: %d tasks\n", len(task.BlockedBy))
	}

	if task.Assignee != "" {
		fmt.Printf("   Assignee: %s\n", task.Assignee)
	}

	if task.StartedAt != nil {
		fmt.Printf("   Started: %s\n", task.StartedAt.Format("2006-01-02 15:04"))
	}

	if task.CompletedAt != nil {
		fmt.Printf("   Completed: %s\n", task.CompletedAt.Format("2006-01-02 15:04"))
	}

	if task.Notes != "" {
		fmt.Printf("\n📝 Notes:\n%s\n", task.Notes)
	}

	fmt.Println()
}

func printDependencies(registryStore *spec.RegistryStore, task *spec.RegistryTask) {
	if len(task.DependsOn) == 0 {
		return
	}

	fmt.Printf("\n📦 Dependencies:\n")
	for _, depID := range task.DependsOn {
		dep, err := registryStore.GetTask(depID)
		if err != nil {
			fmt.Printf("  - %s (error loading: %v)\n", depID.String(), err)
			continue
		}

		statusIcon := "✓"
		if dep.Status != spec.RegStatusCompleted {
			statusIcon = "○"
		}

		fmt.Printf("  %s %s - %s (%s)\n", statusIcon, depID.String(), dep.Content, dep.Status)
	}
	fmt.Println()
}

func printStateContext(store *spec.Store, changeID string) {
	statePath := filepath.Join(store.SpecPath(), "changes", changeID, "state.md")

	data, err := os.ReadFile(statePath) // #nosec G304 -- controlled path from trusted sources
	if err != nil {
		return
	}

	fmt.Printf("══════════════════════════════════════════\n")
	fmt.Printf("CHANGE STATE\n")
	fmt.Printf("══════════════════════════════════════════\n\n")

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimPrefix(line, "## ")
			continue
		}
		if section == "Progress" || section == "Current Task" || section == "Blockers" {
			fmt.Printf("%s\n", line)
		}
	}
}

func printAgentWorkflow(task *spec.RegistryTask) {
	fmt.Printf("Phase 1: Assessment\n")
	fmt.Printf("  → @ent:task-fast will assess this task\n\n")

	complexity := assessComplexity(task)
	if complexity > 0.8 {
		fmt.Printf("Phase 2: Deep Analysis (escalated)\n")
		fmt.Printf("  → @ent:task-heavy should handle this\n")
		fmt.Printf("  → Complexity score: %.2f\n\n", complexity)
	}

	fmt.Printf("Phase 3: Implementation\n")
	fmt.Printf("  → @ent:coder will implement this\n\n")

	if isNonTrivialTask(task) {
		fmt.Printf("Phase 4: Review\n")
		fmt.Printf("  → @ent:reviewer should review this\n\n")
	}

	fmt.Printf("Phase 5: Testing\n")
	fmt.Printf("  → @ent:tester will validate this\n\n")

	fmt.Printf("Phase 6: Acceptance\n")
	fmt.Printf("  → @ent:acceptor will verify against spec\n\n")

	fmt.Printf("Phase 7: Complete\n")
	fmt.Printf("  → Update tasks.md and run state_sync\n")
}

func parseTaskID(input string) (spec.TaskID, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return spec.TaskID{}, fmt.Errorf("invalid task ID format: %s (expected change-id/task-num)", input)
	}

	return spec.TaskID{
		ChangeID: parts[0],
		TaskNum:  parts[1],
	}, nil
}

func assessComplexity(task *spec.RegistryTask) float64 {
	score := 0.5

	if task.Priority == spec.PriorityCritical {
		score += 0.2
	}

	if len(task.DependsOn) > 2 {
		score += 0.1 * float64(len(task.DependsOn)-2)
	}

	if len(task.BlockedBy) > 0 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func isNonTrivialTask(task *spec.RegistryTask) bool {
	if task.Priority == spec.PriorityCritical || task.Priority == spec.PriorityHigh {
		return true
	}
	if len(task.DependsOn) > 1 {
		return true
	}
	return false
}
