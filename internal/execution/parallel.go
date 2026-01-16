package execution

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// ParallelStrategy executes independent tasks concurrently with dependency management.
type ParallelStrategy struct {
	maxConcurrency int
}

// NewParallelStrategy creates a new parallel execution strategy.
func NewParallelStrategy(maxConcurrency int) *ParallelStrategy {
	if maxConcurrency <= 0 {
		maxConcurrency = 4 // Default concurrency
	}
	return &ParallelStrategy{
		maxConcurrency: maxConcurrency,
	}
}

// Name returns the strategy identifier.
func (p *ParallelStrategy) Name() domain.ExecutionStrategy {
	return domain.ExecutionStrategyParallel
}

// ParallelTask represents a task in the dependency graph.
type ParallelTask struct {
	ID          string
	Description string
	Agent       domain.AgentRole
	Model       string
	DependsOn   []string
	Skills      []string
	Metadata    map[string]interface{}
}

// Execute runs tasks in parallel respecting dependencies.
func (p *ParallelStrategy) Execute(ctx context.Context, engine *Engine, task *Task) (*Result, error) {
	// Extract parallel tasks from metadata
	tasks, ok := task.Metadata["parallel_tasks"].([]ParallelTask)
	if !ok || len(tasks) == 0 {
		return nil, fmt.Errorf("no parallel tasks provided in metadata")
	}

	// Build dependency graph
	graph := p.buildDependencyGraph(tasks)

	// Topological sort to get execution order
	sorted, err := p.topologicalSort(graph)
	if err != nil {
		return nil, fmt.Errorf("dependency cycle detected: %w", err)
	}

	// Execute tasks respecting dependencies
	results := make(map[string]*Result)
	var mu sync.Mutex
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(p.maxConcurrency)

	// Process tasks in waves (by dependency level)
	for _, taskID := range sorted {
		parallelTask := p.findTask(tasks, taskID)
		if parallelTask == nil {
			continue
		}

		eg.Go(func() error {
			// Wait for dependencies
			if err := p.waitForDependencies(egCtx, parallelTask, results, &mu); err != nil {
				return err
			}

			// Build execution request
			req := &Request{
				Task:     parallelTask.Description,
				Agent:    parallelTask.Agent,
				Model:    parallelTask.Model,
				Skills:   parallelTask.Skills,
				Strategy: domain.ExecutionStrategyParallel,
				Budget:   task.Budget,
				Context:  task.Context,
				Metadata: parallelTask.Metadata,
			}

			// Add dependency outputs to context
			mu.Lock()
			if len(parallelTask.DependsOn) > 0 {
				if req.Metadata == nil {
					req.Metadata = make(map[string]interface{})
				}
				depOutputs := make(map[string]string)
				for _, depID := range parallelTask.DependsOn {
					if depResult, ok := results[depID]; ok {
						depOutputs[depID] = depResult.Output
					}
				}
				req.Metadata["dependency_outputs"] = depOutputs
			}
			mu.Unlock()

			// Select runtime
			runtime := task.ForceRuntime
			if runtime == "" {
				runtime = engine.selectRuntime(egCtx)
			}

			// Get runner
			runner, err := engine.getRunner(runtime)
			if err != nil {
				return fmt.Errorf("get runner for task %s: %w", taskID, err)
			}

			// Check budget
			if task.Budget != nil {
				estimate := NewCostEstimate(parallelTask.Model, 2000, 1000)
				if err := engine.budget.Check(egCtx, estimate, task.Budget); err != nil {
					return fmt.Errorf("budget check for task %s: %w", taskID, err)
				}
			}

			// Execute
			result, err := runner.Execute(egCtx, req)
			if err != nil {
				return fmt.Errorf("execution of task %s: %w", taskID, err)
			}

			// Record result
			mu.Lock()
			results[taskID] = result
			mu.Unlock()

			// Record spending
			if result.Success {
				cost := CalculateCost(parallelTask.Model, result.TokensIn, result.TokensOut)
				result.Cost = cost

				taskKey := fmt.Sprintf("%s-%s", task.Context.ChangeID, taskID)
				engine.budget.Record(taskKey, result.TokensIn, result.TokensOut, cost)
			}

			return nil
		})
	}

	// Wait for all tasks
	if err := eg.Wait(); err != nil {
		return p.aggregateResults(results, sorted, false, err.Error()), nil
	}

	return p.aggregateResults(results, sorted, true, ""), nil
}

// buildDependencyGraph creates adjacency list representation.
func (p *ParallelStrategy) buildDependencyGraph(tasks []ParallelTask) map[string][]string {
	graph := make(map[string][]string)
	for _, t := range tasks {
		if _, exists := graph[t.ID]; !exists {
			graph[t.ID] = []string{}
		}
		for _, dep := range t.DependsOn {
			graph[dep] = append(graph[dep], t.ID)
		}
	}
	return graph
}

// topologicalSort returns tasks in dependency order.
func (p *ParallelStrategy) topologicalSort(graph map[string][]string) ([]string, error) {
	// Calculate in-degrees
	inDegree := make(map[string]int)
	for node := range graph {
		if _, ok := inDegree[node]; !ok {
			inDegree[node] = 0
		}
		for _, neighbor := range graph[node] {
			inDegree[neighbor]++
		}
	}

	// Find nodes with no dependencies
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	var sorted []string
	for len(queue) > 0 {
		// Pop from queue
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		// Reduce in-degree for neighbors
		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycle
	if len(sorted) != len(graph) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return sorted, nil
}

// waitForDependencies blocks until all dependencies complete.
func (p *ParallelStrategy) waitForDependencies(ctx context.Context, task *ParallelTask, results map[string]*Result, mu *sync.Mutex) error {
	if len(task.DependsOn) == 0 {
		return nil
	}

	// Poll for dependencies
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			mu.Lock()
			allReady := true
			var failedDep string
			for _, depID := range task.DependsOn {
				if result, ok := results[depID]; ok {
					if !result.Success {
						failedDep = depID
						allReady = false
						break
					}
				} else {
					allReady = false
				}
			}
			mu.Unlock()

			if failedDep != "" {
				return fmt.Errorf("dependency %s failed", failedDep)
			}
			if allReady {
				return nil
			}

			// Brief sleep before next check
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// findTask locates a task by ID.
func (p *ParallelStrategy) findTask(tasks []ParallelTask, id string) *ParallelTask {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
}

// aggregateResults combines all task results.
func (p *ParallelStrategy) aggregateResults(results map[string]*Result, order []string, success bool, errMsg string) *Result {
	var output strings.Builder
	var totalTokensIn, totalTokensOut int
	var totalCost float64
	var adjustments []string

	for _, taskID := range order {
		if result, ok := results[taskID]; ok {
			output.WriteString(fmt.Sprintf("=== Task %s ===\n%s\n\n", taskID, result.Output))
			totalTokensIn += result.TokensIn
			totalTokensOut += result.TokensOut
			totalCost += result.Cost
			adjustments = append(adjustments, result.Adjustments...)
		}
	}

	return &Result{
		Success:     success,
		Output:      output.String(),
		Error:       errMsg,
		TokensIn:    totalTokensIn,
		TokensOut:   totalTokensOut,
		Cost:        totalCost,
		Adjustments: adjustments,
		Metadata: map[string]interface{}{
			"tasks_completed": len(results),
			"execution_order": order,
		},
	}
}

// CanHandle checks if this strategy can handle the task.
func (p *ParallelStrategy) CanHandle(task *Task) bool {
	// Parallel strategy requires parallel_tasks in metadata
	_, ok := task.Metadata["parallel_tasks"].([]ParallelTask)
	return ok
}
