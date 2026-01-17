package background

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/domain"
)

// Manager manages background agent lifecycle.
type Manager struct {
	mu          sync.RWMutex
	agents      map[string]*Agent
	selector    Selector
	cfg         Config
	shutdownCtx context.Context
	cancel      context.CancelFunc
	onShutdown  []func(context.Context) error

	// Resource tracking
	agentGoroutines map[string]int
}

// Selector defines the interface for selecting agent configuration.
type Selector interface {
	Select(ctx context.Context, task interface{}) (*SelectionResult, error)
}

// TaskType represents the complexity classification of a task.
type TaskType string

const (
	TaskTypeExploration TaskType = "exploration"
	TaskTypeComplexity  TaskType = "complexity"
	TaskTypeCritical    TaskType = "critical"
)

// SelectionResult represents a selected agent configuration.
type SelectionResult struct {
	Role   string
	Model  string
	Skills []string
	Reason string
}

// SpawnOpts holds optional parameters for spawning an agent.
type SpawnOpts struct {
	Role    string
	Model   string
	Timeout int
}

// Config holds background agent configuration.
type Config struct {
	// MaxConcurrent is the maximum number of concurrent agents.
	MaxConcurrent int

	// DefaultRole is the default agent role to use.
	DefaultRole string

	// DefaultModel is the default model to use.
	DefaultModel string

	// Timeout is the maximum execution duration per agent.
	Timeout int

	// ModelTier configures model selection by task complexity.
	ModelTier config.ModelTierConfig

	// Models maps friendly names to actual model IDs.
	Models config.ModelsConfig

	// CleanupInterval is how often to check for old agents to cleanup.
	CleanupInterval time.Duration

	// MaxAgentAge is the maximum age before an agent is considered abandoned.
	MaxAgentAge time.Duration

	// ResourceLimits sets per-agent resource constraints.
	ResourceLimits config.ResourceLimits
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() Config {
	return Config{
		MaxConcurrent:   5,
		DefaultRole:     "developer",
		DefaultModel:    "haiku",
		Timeout:         300,
		CleanupInterval: time.Hour,
		MaxAgentAge:     time.Hour,
		ResourceLimits: config.ResourceLimits{
			MaxMemoryMB:   512,
			MaxGoroutines: 100,
			MaxCPUPercent: 80,
		},
	}
}

// classifyTask analyzes a task string and determines its complexity type.
func classifyTask(task string) TaskType {
	lower := strings.ToLower(task)

	criticalKeywords := []string{
		"critical", "important", "decision", "approve", "security",
		"breaking", "delete", "remove", "dangerous", "production",
	}
	for _, kw := range criticalKeywords {
		if strings.Contains(lower, kw) {
			return TaskTypeCritical
		}
	}

	complexityKeywords := []string{
		"implement", "refactor", "optimize", "design", "architect",
		"solve", "debug", "write", "create", "build", "integrate",
		"migrate", "transform", "restructure",
	}
	for _, kw := range complexityKeywords {
		if strings.Contains(lower, kw) {
			return TaskTypeComplexity
		}
	}

	explorationKeywords := []string{
		"explore", "analyze", "find", "search", "list", "check",
		"investigate", "read", "view", "examine", "review", "inspect",
	}
	for _, kw := range explorationKeywords {
		if strings.Contains(lower, kw) {
			return TaskTypeExploration
		}
	}

	return TaskTypeExploration
}

// detectAction infers the domain action from a task description.
func detectAction(task string) domain.SpecAction {
	lower := strings.ToLower(task)

	if strings.Contains(lower, "research") {
		return domain.SpecActionResearch
	}
	if strings.Contains(lower, "analyze") || strings.Contains(lower, "examine") || strings.Contains(lower, "investigate") {
		return domain.SpecActionAnalyze
	}
	if strings.Contains(lower, "plan") {
		return domain.SpecActionPlan
	}
	if strings.Contains(lower, "design") {
		return domain.SpecActionDesign
	}
	if strings.Contains(lower, "review") {
		return domain.SpecActionReview
	}
	if strings.Contains(lower, "test") || strings.Contains(lower, "verify") {
		return domain.SpecActionVerify
	}
	if strings.Contains(lower, "debug") {
		return domain.SpecActionDebug
	}

	return domain.SpecActionImplement
}

// checkResourceLimits verifies system can accommodate new agent based on configured limits.
func (m *Manager) checkResourceLimits() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.agents) >= m.cfg.MaxConcurrent {
		return fmt.Errorf("max concurrent agents (%d) reached", m.cfg.MaxConcurrent)
	}

	return nil
}

// toAgentTask converts a background task string to an agent Task struct.
func toAgentTask(task string) agent.Task {
	lower := strings.ToLower(task)

	taskType := agent.TaskTypeFeature
	taskTypeBg := classifyTask(task)

	switch taskTypeBg {
	case TaskTypeCritical:
		taskType = agent.TaskTypeBugFix
	case TaskTypeComplexity:
		if strings.Contains(lower, "architecture") || strings.Contains(lower, "design") {
			taskType = agent.TaskTypeArchitecture
		} else {
			taskType = agent.TaskTypeFeature
		}
	case TaskTypeExploration:
		taskType = agent.TaskTypeDocumentation
	}

	action := detectAction(task)
	phase := action.Phase()

	return agent.Task{
		Description: task,
		Type:        taskType,
		Action:      action,
		Phase:       phase,
		Files:       []string{},
		Metadata:    make(map[string]interface{}),
	}
}

// fromAgentSelection converts an agent selection result to background selection result.
func fromAgentSelection(sr *agent.SelectionResult) *SelectionResult {
	return &SelectionResult{
		Role:   sr.Role.String(),
		Model:  sr.Model,
		Skills: sr.Skills,
		Reason: sr.Reason,
	}
}

// AgentSelectorAdapter wraps agent.Selector to implement background.Selector.
type AgentSelectorAdapter struct {
	selector *agent.Selector
}

// NewAgentSelectorAdapter creates a new adapter for agent.Selector.
func NewAgentSelectorAdapter(selector *agent.Selector) *AgentSelectorAdapter {
	return &AgentSelectorAdapter{selector: selector}
}

// Select implements background.Selector interface.
func (a *AgentSelectorAdapter) Select(ctx context.Context, task interface{}) (*SelectionResult, error) {
	agentTask, ok := task.(agent.Task)
	if !ok {
		return nil, fmt.Errorf("expected agent.Task, got %T", task)
	}

	result, err := a.selector.Select(ctx, agentTask)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return fromAgentSelection(result), nil
}

// selectModelByTier selects the appropriate model based on task type.
func selectModelByTier(taskType TaskType, modelTier config.ModelTierConfig) string {
	switch taskType {
	case TaskTypeCritical:
		return modelTier.Critical
	case TaskTypeComplexity:
		return modelTier.Complexity
	case TaskTypeExploration:
		return modelTier.Exploration
	default:
		return ""
	}
}

// NewManager creates a new background agent manager.
func NewManager(selector Selector, cfg Config) *Manager {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 5
	}
	if cfg.DefaultRole == "" {
		cfg.DefaultRole = "developer"
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "haiku"
	}

	shutdownCtx, cancel := context.WithCancel(context.Background())

	return &Manager{
		agents:          make(map[string]*Agent),
		selector:        selector,
		cfg:             cfg,
		shutdownCtx:     shutdownCtx,
		cancel:          cancel,
		onShutdown:      make([]func(context.Context) error, 0),
		agentGoroutines: make(map[string]int),
	}
}

// Spawn creates and starts a new background agent.
func (m *Manager) Spawn(ctx context.Context, task string, opts SpawnOpts) (*Agent, error) {
	if task == "" {
		return nil, fmt.Errorf("task required")
	}

	if err := m.checkResourceLimits(); err != nil {
		return nil, fmt.Errorf("resource limits: %w", err)
	}

	id := uuid.New().String()

	role := m.cfg.DefaultRole
	model := m.cfg.DefaultModel

	if m.selector != nil {
		agentTask := toAgentTask(task)
		result, err := m.selector.Select(ctx, agentTask)
		if err == nil && result != nil && result.Model != "" {
			role = result.Role
			model = result.Model
		}
	}

	if model == m.cfg.DefaultModel && m.cfg.Models != nil {
		taskType := classifyTask(task)
		modelKey := selectModelByTier(taskType, m.cfg.ModelTier)

		if modelKey != "" {
			if modelID, ok := m.cfg.Models[modelKey]; ok {
				model = modelID
			}
		}
	}

	if opts.Role != "" {
		role = opts.Role
	}

	if opts.Model != "" {
		model = opts.Model
	}

	agent, err := NewAgent(id, role, model, task)
	if err != nil {
		return nil, fmt.Errorf("new agent: %w", err)
	}

	m.mu.Lock()
	m.agents[id] = agent
	m.agentGoroutines[id] = 0
	m.mu.Unlock()

	go m.runAgent(ctx, agent)

	return agent, nil
}

// runAgent executes the agent and updates its status.
// TODO: Implement actual task execution logic.
func (m *Manager) runAgent(ctx context.Context, agent *Agent) {
	agent.Start()

	defer func() {
		if r := recover(); r != nil {
			agent.Fail(fmt.Errorf("panic: %v", r))
		}
	}()

	agentCtx, cancel := context.WithTimeout(ctx, time.Duration(m.cfg.Timeout)*time.Second)
	defer cancel()

	// TODO: Replace with actual task execution
	// For now, simulate work by waiting briefly then completing
	select {
	case <-agentCtx.Done():
		if m.shutdownCtx.Err() != nil {
			agent.Kill()
		} else if agentCtx.Err() == context.DeadlineExceeded {
			agent.Fail(fmt.Errorf("timeout after %ds", m.cfg.Timeout))
		} else {
			agent.Complete("task executed")
		}
	case <-time.After(10 * time.Millisecond):
		// Simulate quick task completion for now
		agent.Complete("task executed")
	}
}

// Get retrieves an agent by ID.
func (m *Manager) Get(id string) (*Agent, error) {
	m.mu.RLock()
	agent, exists := m.agents[id]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrAgentNotFound
	}

	return agent, nil
}

// List returns all agents, optionally filtered by status.
func (m *Manager) List(status Status) []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agents []*Agent
	for _, agent := range m.agents {
		agent.mu.RLock()
		match := status == "" || agent.Status == status
		agent.mu.RUnlock()
		if match {
			agents = append(agents, agent)
		}
	}
	return agents
}

// Kill terminates a running agent.
func (m *Manager) Kill(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[id]
	if !exists {
		return ErrAgentNotFound
	}

	agent.mu.RLock()
	agentStatus := agent.Status
	agent.mu.RUnlock()

	if agentStatus != StatusRunning {
		return fmt.Errorf("agent not running: %s", agentStatus)
	}

	agent.Kill()
	return nil
}

// Cleanup removes completed or failed agents.
func (m *Manager) Cleanup(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int
	for id, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		agent.mu.RUnlock()

		if agentStatus == StatusCompleted || agentStatus == StatusFailed || agentStatus == StatusKilled {
			delete(m.agents, id)
			delete(m.agentGoroutines, id)
			count++
		}
	}
	return count
}

// Count returns the total number of agents.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.agents)
}

// CountByStatus returns the number of agents with the given status.
func (m *Manager) CountByStatus(status Status) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int
	for _, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		agent.mu.RUnlock()

		if agentStatus == status {
			count++
		}
	}
	return count
}

// IncrementGoroutines increments goroutine count for an agent.
func (m *Manager) IncrementGoroutines(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	limit := m.cfg.ResourceLimits.MaxGoroutines
	if limit > 0 {
		current := m.agentGoroutines[agentID]
		if current >= limit {
			return fmt.Errorf("goroutine limit (%d) exceeded for agent %s", limit, agentID)
		}
		m.agentGoroutines[agentID] = current + 1
	}
	return nil
}

// DecrementGoroutines decrements goroutine count for an agent.
func (m *Manager) DecrementGoroutines(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if current, ok := m.agentGoroutines[agentID]; ok && current > 0 {
		m.agentGoroutines[agentID] = current - 1
	}
}

// GetResourceUsage returns current resource usage.
func (m *Manager) GetResourceUsage() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	usage := make(map[string]int)
	for id, count := range m.agentGoroutines {
		usage[id] = count
	}
	return usage
}

// Shutdown gracefully stops all running agents and cleans up.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		agent.mu.RUnlock()

		if agentStatus == StatusRunning {
			agent.Kill()
		}
	}

	m.agents = make(map[string]*Agent)
	m.agentGoroutines = make(map[string]int)

	m.cancel()

	for _, hook := range m.onShutdown {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("shutdown hook failed: %w", err)
		}
	}

	return nil
}

// OnShutdown registers a hook to be called during shutdown.
func (m *Manager) OnShutdown(hook func(context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onShutdown = append(m.onShutdown, hook)
}

// CleanupOld removes agents older than maxAge.
func (m *Manager) CleanupOld(ctx context.Context, maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var count int

	for id, agent := range m.agents {
		agent.mu.RLock()
		agentStatus := agent.Status
		createdAt := agent.CreatedAt
		agent.mu.RUnlock()

		age := now.Sub(createdAt)
		if (agentStatus == StatusCompleted || agentStatus == StatusFailed || agentStatus == StatusKilled) && age > maxAge {
			delete(m.agents, id)
			delete(m.agentGoroutines, id)
			count++
		}
	}

	return count
}

// StartCleanupRoutine starts a background goroutine that periodically cleans up old agents.
func (m *Manager) StartCleanupRoutine(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(m.cfg.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-m.shutdownCtx.Done():
				return
			case <-ticker.C:
				cleaned := m.CleanupOld(ctx, m.cfg.MaxAgentAge)
				if cleaned > 0 {
					fmt.Printf("cleaned up %d old agents\n", cleaned)
				}
			}
		}
	}()
}
