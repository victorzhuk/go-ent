package worker

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/opencode"
	"github.com/victorzhuk/go-ent/internal/openspec"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthUnknown   HealthStatus = "unknown"
	HealthTimeout   HealthStatus = "timeout"
)

func (h HealthStatus) String() string {
	if h == "" {
		return "unknown"
	}
	return string(h)
}

func (h HealthStatus) Valid() bool {
	switch h {
	case HealthHealthy, HealthUnhealthy, HealthUnknown, HealthTimeout:
		return true
	default:
		return false
	}
}

type WorkerStatus string

const (
	StatusIdle      WorkerStatus = "idle"
	StatusRunning   WorkerStatus = "running"
	StatusCompleted WorkerStatus = "completed"
	StatusFailed    WorkerStatus = "failed"
	StatusCancelled WorkerStatus = "cancelled"
)

func (s WorkerStatus) String() string {
	if s == "" {
		return "unknown"
	}
	return string(s)
}

func (s WorkerStatus) Valid() bool {
	switch s {
	case StatusIdle, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

type ProviderConfig struct {
	Name    string
	BaseURL string
	APIKey  string
	Models  []string
	Enabled bool
}

type Worker struct {
	ID        string
	Provider  string
	Model     string
	Method    config.CommunicationMethod
	Status    WorkerStatus
	Task      *execution.Task
	StartedAt time.Time
	Output    string
	Mutex     sync.Mutex

	cmd        *exec.Cmd
	cancel     context.CancelFunc
	configPath string

	Health           HealthStatus
	LastHealthCheck  time.Time
	LastOutputTime   time.Time
	HealthCheckCount int
	UnhealthySince   time.Time
	RetryCount       int

	acpClient *opencode.ACPClient
}

type ResultAggregator struct {
	result map[string]string
}

func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{
		result: make(map[string]string),
	}
}

type WorkerPool struct {
	maxConcurrency int
	running        int
	mu             sync.Mutex
	cond           *sync.Cond
}

func NewWorkerPool(maxConcurrency int) *WorkerPool {
	p := &WorkerPool{
		maxConcurrency: maxConcurrency,
	}
	p.cond = sync.NewCond(&p.mu)
	return p
}

func (p *WorkerPool) Acquire(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.running >= p.maxConcurrency {
		select {
		case <-ctx.Done():
			return fmt.Errorf("acquire: %w", ctx.Err())
		default:
			p.cond.Wait()
		}
	}

	p.running++
	return nil
}

func (p *WorkerPool) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.running--
	p.cond.Signal()
}

func (p *WorkerPool) Running() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

type SpawnRequest struct {
	WorkerID           string
	Provider           string
	Model              string
	Method             config.CommunicationMethod
	Task               *execution.Task
	Timeout            time.Duration
	Metadata           map[string]interface{}
	OpenCodeConfigPath string
}

type WorkerManager struct {
	workers       map[string]*Worker
	configs       map[string]ProviderConfig
	pool          *WorkerPool
	aggregator    *ResultAggregator
	taskTracker   *openspec.TaskTracker
	registryStore *spec.RegistryStore
	hooks         *HookChain
	mu            sync.RWMutex
	logger        *slog.Logger
}

func NewWorkerManager(taskTracker *openspec.TaskTracker, registryStore *spec.RegistryStore) *WorkerManager {
	pool := NewWorkerPool(10)

	return &WorkerManager{
		workers:       make(map[string]*Worker),
		configs:       make(map[string]ProviderConfig),
		pool:          pool,
		aggregator:    NewResultAggregator(),
		taskTracker:   taskTracker,
		registryStore: registryStore,
		hooks:         NewHookChain(),
		logger:        slog.Default(),
	}
}

func NewWorkerManagerWithoutTracking() *WorkerManager {
	return NewWorkerManager(nil, nil)
}

func (m *WorkerManager) RegisterHook(hookType HookType, hook HookFunc, filter *HookFilter) {
	m.hooks.Add(hookType, hook, filter)
	m.logger.Debug("hook registered",
		"hook_type", hookType,
		"filter", filter,
	)
}

func (m *WorkerManager) Spawn(ctx context.Context, req SpawnRequest) (string, error) {
	if !req.Method.Valid() {
		return "", fmt.Errorf("invalid communication method: %s", req.Method)
	}

	workerID := req.WorkerID
	if workerID == "" {
		workerID = uuid.Must(uuid.NewV7()).String()
	}

	worker := &Worker{
		ID:         workerID,
		Provider:   req.Provider,
		Model:      req.Model,
		Method:     req.Method,
		Status:     StatusIdle,
		Task:       req.Task,
		StartedAt:  time.Now(),
		configPath: req.OpenCodeConfigPath,
	}

	hookCtx := &HookContext{
		HookType: HookPreSpawn,
		Task:     req.Task,
		Worker:   worker,
	}

	if err := m.ExecuteHooks(ctx, HookPreSpawn, hookCtx); err != nil {
		m.logger.Warn("pre_spawn hook failed",
			"worker_id", workerID,
			"error", err,
		)
		return "", fmt.Errorf("pre_spawn hook failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workers[workerID]; exists {
		return "", fmt.Errorf("worker %s already exists", workerID)
	}

	m.workers[workerID] = worker

	m.logger.Debug("spawned worker",
		"worker_id", workerID,
		"provider", req.Provider,
		"model", req.Model,
		"method", req.Method,
		"config_path", req.OpenCodeConfigPath,
	)

	if m.taskTracker != nil && req.Task != nil {
		taskID := m.taskTracker.ExtractTaskID(req.Task.Description)
		if !taskID.IsZero() {
			if err := m.taskTracker.MarkInProgress(taskID); err != nil {
				m.logger.Warn("failed to mark task in progress",
					"task_id", taskID.String(),
					"error", err,
				)
			} else {
				m.logger.Info("marked task in progress",
					"task_id", taskID.String(),
					"worker_id", workerID,
				)
			}
		}
	}

	return workerID, nil
}

func (m *WorkerManager) Get(workerID string) *Worker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.workers[workerID]
}

func (m *WorkerManager) ExecuteHooks(ctx context.Context, hookType HookType, hookCtx *HookContext) error {
	return m.hooks.Execute(ctx, hookType, hookCtx)
}

func (m *WorkerManager) Cancel(ctx context.Context, workerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	worker, exists := m.workers[workerID]
	if !exists {
		return fmt.Errorf("worker %s not found", workerID)
	}

	worker.Mutex.Lock()
	worker.Status = StatusCancelled
	worker.Mutex.Unlock()

	if worker.Method == config.MethodACP && worker.Status == StatusRunning {
		m.pool.Release()
	}

	m.logger.Debug("cancelled worker", "worker_id", workerID)

	return nil
}

func (m *WorkerManager) List(statusFilter ...WorkerStatus) []*Worker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(statusFilter) == 0 {
		result := make([]*Worker, 0, len(m.workers))
		for _, w := range m.workers {
			result = append(result, w)
		}
		return result
	}

	var result []*Worker
	filter := statusFilter[0]

	for _, w := range m.workers {
		if w.Status == filter {
			result = append(result, w)
		}
	}

	return result
}

func (m *WorkerManager) Cleanup(maxAge ...time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	age := 1 * time.Hour
	if len(maxAge) > 0 {
		age = maxAge[0]
	}

	now := time.Now()
	var toDelete []string

	for id, worker := range m.workers {
		if !worker.Status.Valid() {
			continue
		}

		if worker.Status == StatusCompleted ||
			worker.Status == StatusFailed ||
			worker.Status == StatusCancelled {
			if now.Sub(worker.StartedAt) > age {
				toDelete = append(toDelete, id)
			}
		}
	}

	for _, id := range toDelete {
		delete(m.workers, id)
	}

	m.logger.Debug("cleaned up workers", "count", len(toDelete))

	return len(toDelete)
}

func (m *WorkerManager) SetWorkerStatus(workerID string, status WorkerStatus) {
	m.mu.RLock()
	worker, exists := m.workers[workerID]
	m.mu.RUnlock()

	if !exists {
		panic(fmt.Sprintf("worker %s not found", workerID))
	}

	worker.Mutex.Lock()
	oldStatus := worker.Status
	worker.Status = status
	worker.Mutex.Unlock()

	if m.taskTracker != nil && worker.Task != nil {
		taskID := m.taskTracker.ExtractTaskID(worker.Task.Description)
		if !taskID.IsZero() {
			switch status {
			case StatusCompleted:
				if oldStatus != StatusCompleted {
					if err := m.taskTracker.MarkCompleted(taskID); err != nil {
						m.logger.Warn("failed to mark task completed",
							"task_id", taskID.String(),
							"error", err,
						)
					} else {
						m.logger.Info("marked task completed",
							"task_id", taskID.String(),
							"worker_id", workerID,
						)
					}
				}
			case StatusFailed:
				if oldStatus != StatusFailed {
					if err := m.taskTracker.MarkFailed(taskID, "Worker failed"); err != nil {
						m.logger.Warn("failed to mark task failed",
							"task_id", taskID.String(),
							"error", err,
						)
					} else {
						m.logger.Info("marked task failed",
							"task_id", taskID.String(),
							"worker_id", workerID,
						)
					}
				}
			}
		}
	}
}

func (m *WorkerManager) GetStatus(workerID string) (WorkerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	worker, exists := m.workers[workerID]
	if !exists {
		return WorkerStatus(""), fmt.Errorf("worker %s not found", workerID)
	}

	return worker.Status, nil
}

type PromptRequest struct {
	WorkerID     string
	Prompt       string
	ContextFiles []string
	Tools        []string
}

type PromptResponse struct {
	PromptID string
	Status   string
	Result   string
}

func (m *WorkerManager) SendPrompt(ctx context.Context, req PromptRequest) (*PromptResponse, error) {
	m.mu.RLock()
	worker, exists := m.workers[req.WorkerID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("worker %s not found", req.WorkerID)
	}

	if !worker.Status.Valid() {
		return nil, fmt.Errorf("worker %s has invalid status: %s", req.WorkerID, worker.Status)
	}

	if worker.Status != StatusIdle && worker.Status != StatusRunning {
		return nil, fmt.Errorf("worker %s cannot accept prompts (status: %s)", req.WorkerID, worker.Status)
	}

	if worker.Method != config.MethodACP {
		return nil, fmt.Errorf("worker %s does not support prompting (method: %s)", req.WorkerID, worker.Method)
	}

	hookCtx := &HookContext{
		HookType: HookPrePrompt,
		Task:     worker.Task,
		Worker:   worker,
	}

	if err := m.ExecuteHooks(ctx, HookPrePrompt, hookCtx); err != nil {
		m.logger.Warn("pre_prompt hook failed",
			"worker_id", req.WorkerID,
			"error", err,
		)
		return nil, fmt.Errorf("pre_prompt hook failed: %w", err)
	}

	worker.Mutex.Lock()
	if worker.acpClient == nil {
		worker.Mutex.Unlock()
		return nil, fmt.Errorf("worker %s has no active ACP session", req.WorkerID)
	}

	if !worker.acpClient.IsInitialized() {
		worker.Mutex.Unlock()
		return nil, fmt.Errorf("worker %s ACP client not initialized", req.WorkerID)
	}

	if worker.acpClient.SessionStatus() == "" {
		worker.Mutex.Unlock()
		return nil, fmt.Errorf("worker %s has no active session", req.WorkerID)
	}

	worker.Status = StatusRunning
	worker.Mutex.Unlock()

	var context []opencode.MessageContext
	if len(req.ContextFiles) > 0 {
		for _, file := range req.ContextFiles {
			context = append(context, opencode.MessageContext{
				Role:    "file",
				Content: file,
			})
		}
	}

	options := make(map[string]any)
	if len(req.Tools) > 0 {
		options["tools"] = req.Tools
	}

	result, err := worker.acpClient.SessionPrompt(ctx, req.Prompt, context, options)

	postHookCtx := &HookContext{
		Task:   worker.Task,
		Worker: worker,
	}

	if err != nil {
		worker.Mutex.Lock()
		worker.Status = StatusFailed
		worker.Mutex.Unlock()

		postHookCtx.HookType = HookPostFail
		postHookCtx.Error = err

		if hookErr := m.ExecuteHooks(ctx, HookPostFail, postHookCtx); hookErr != nil {
			m.logger.Warn("post_fail hook failed",
				"worker_id", req.WorkerID,
				"error", hookErr,
			)
		}

		if m.taskTracker != nil && worker.Task != nil {
			taskID := m.taskTracker.ExtractTaskID(worker.Task.Description)
			if !taskID.IsZero() {
				if trackErr := m.taskTracker.MarkFailed(taskID, err.Error()); trackErr != nil {
					m.logger.Warn("failed to mark task failed",
						"task_id", taskID.String(),
						"error", trackErr,
					)
				}
			}
		}

		return nil, fmt.Errorf("send prompt: %w", err)
	}

	postHookCtx.HookType = HookPostComplete
	postHookCtx.Result = &WorkerResult{
		WorkerID: req.WorkerID,
		Success:  true,
	}

	if hookErr := m.ExecuteHooks(ctx, HookPostComplete, postHookCtx); hookErr != nil {
		m.logger.Warn("post_complete hook failed",
			"worker_id", req.WorkerID,
			"error", hookErr,
		)
	}

	response := &PromptResponse{
		PromptID: result.PromptID,
		Status:   result.Status,
	}

	m.logger.Debug("prompt sent",
		"worker_id", req.WorkerID,
		"prompt_id", result.PromptID,
		"status", result.Status,
	)

	return response, nil
}

type WorkerOutputRequest struct {
	WorkerID string
	Since    time.Time
	Filter   string
	Limit    int
}

type WorkerOutputResponse struct {
	WorkerID    string
	Output      string
	LineCount   int
	LastUpdated time.Time
	Truncated   bool
}

func (m *WorkerManager) GetOutput(req WorkerOutputRequest) (*WorkerOutputResponse, error) {
	m.mu.RLock()
	worker, exists := m.workers[req.WorkerID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("worker %s not found", req.WorkerID)
	}

	worker.Mutex.Lock()
	defer worker.Mutex.Unlock()

	output := worker.Output
	lines := splitLines(output)
	lineCount := len(lines)

	var filteredOutput string
	truncated := false

	filteredLines := lines

	if !req.Since.IsZero() {
		sinceLines := []string{}
		for _, line := range filteredLines {
			if worker.LastOutputTime.After(req.Since) || worker.LastOutputTime.Equal(req.Since) {
				sinceLines = append(sinceLines, line)
			}
		}
		filteredLines = sinceLines
	}

	if req.Filter != "" {
		re, err := regexp.Compile(req.Filter)
		if err != nil {
			return nil, fmt.Errorf("compile regex: %w", err)
		}
		filterLines := []string{}
		for _, line := range filteredLines {
			if re.MatchString(line) {
				filterLines = append(filterLines, line)
			}
		}
		filteredLines = filterLines
	}

	if req.Limit > 0 && len(filteredLines) > req.Limit {
		filteredLines = filteredLines[:req.Limit]
		truncated = true
	}

	filteredOutput = joinLines(filteredLines)

	return &WorkerOutputResponse{
		WorkerID:    worker.ID,
		Output:      filteredOutput,
		LineCount:   lineCount,
		LastUpdated: worker.LastOutputTime,
		Truncated:   truncated,
	}, nil
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := []string{}
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
