package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/execution"
)

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

type CommunicationMethod string

const (
	MethodACP CommunicationMethod = "acp"
	MethodCLI CommunicationMethod = "cli"
	MethodAPI CommunicationMethod = "api"
)

func (m CommunicationMethod) String() string {
	if m == "" {
		return "unknown"
	}
	return string(m)
}

func (m CommunicationMethod) Valid() bool {
	switch m {
	case MethodACP, MethodCLI, MethodAPI:
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
	Method    CommunicationMethod
	Status    WorkerStatus
	Task      *execution.Task
	StartedAt time.Time
	Output    string
	Mutex     sync.Mutex

	cmd        *exec.Cmd
	cancel     context.CancelFunc
	configPath string
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
	WorkerID string
	Provider string
	Model    string
	Method   CommunicationMethod
	Task     *execution.Task
	Timeout  time.Duration
	Metadata map[string]interface{}
}

type WorkerManager struct {
	workers    map[string]*Worker
	configs    map[string]ProviderConfig
	pool       *WorkerPool
	aggregator *ResultAggregator
	mu         sync.RWMutex
	logger     *slog.Logger
}

func NewWorkerManager() *WorkerManager {
	pool := NewWorkerPool(10)

	return &WorkerManager{
		workers:    make(map[string]*Worker),
		configs:    make(map[string]ProviderConfig),
		pool:       pool,
		aggregator: NewResultAggregator(),
		logger:     slog.Default(),
	}
}

func (m *WorkerManager) Spawn(ctx context.Context, req SpawnRequest) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !req.Method.Valid() {
		return "", fmt.Errorf("invalid communication method: %s", req.Method)
	}

	workerID := req.WorkerID
	if workerID == "" {
		workerID = uuid.Must(uuid.NewV7()).String()
	}

	if _, exists := m.workers[workerID]; exists {
		return "", fmt.Errorf("worker %s already exists", workerID)
	}

	worker := &Worker{
		ID:        workerID,
		Provider:  req.Provider,
		Model:     req.Model,
		Method:    req.Method,
		Status:    StatusIdle,
		Task:      req.Task,
		StartedAt: time.Now(),
	}

	m.workers[workerID] = worker

	m.logger.Debug("spawned worker",
		"worker_id", workerID,
		"provider", req.Provider,
		"model", req.Model,
		"method", req.Method,
	)

	return workerID, nil
}

func (m *WorkerManager) Get(workerID string) *Worker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.workers[workerID]
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

	if worker.Method == MethodACP && worker.Status == StatusRunning {
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
	worker.Status = status
	worker.Mutex.Unlock()
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
