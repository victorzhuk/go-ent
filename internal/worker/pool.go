package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
)

type Pool struct {
	workers             map[string]*Worker
	maxConcurrency      int
	mu                  sync.Mutex
	cond                *sync.Cond
	healthCheckInterval time.Duration
	workerTimeout       time.Duration
	maxRetries          int
	retryDelay          time.Duration
	logger              *slog.Logger
	ctx                 context.Context
	cancel              context.CancelFunc
}

func NewPool(maxConcurrency int, healthCheckInterval, workerTimeout time.Duration, maxRetries int, retryDelay time.Duration) *Pool {
	p := &Pool{
		workers:             make(map[string]*Worker),
		maxConcurrency:      maxConcurrency,
		healthCheckInterval: healthCheckInterval,
		workerTimeout:       workerTimeout,
		maxRetries:          maxRetries,
		retryDelay:          retryDelay,
		logger:              slog.Default(),
	}
	p.cond = sync.NewCond(&p.mu)
	return p
}

func (p *Pool) StartHealthChecker(ctx context.Context) {
	p.mu.Lock()
	p.ctx, p.cancel = ctx, nil
	p.mu.Unlock()

	go p.runHealthChecks()
}

func (p *Pool) StopHealthChecker() {
	p.mu.Lock()
	if p.cancel != nil {
		p.cancel()
	}
	p.mu.Unlock()
}

func (p *Pool) runHealthChecks() {
	ticker := time.NewTicker(p.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.checkAllWorkersHealth()
		}
	}
}

func (p *Pool) checkAllWorkersHealth() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, w := range p.workers {
		w.Mutex.Lock()
		status := w.Status
		w.Mutex.Unlock()

		if status != StatusRunning {
			continue
		}

		ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
		health := w.CheckHealth(ctx, p.workerTimeout)
		cancel()

		if health != HealthHealthy {
			p.logger.Debug("worker unhealthy",
				"worker_id", id,
				"health", health,
				"unhealthy_since", w.UnhealthySince,
			)

			if health == HealthTimeout || health == HealthUnhealthy {
				p.handleUnhealthyWorker(w, id)
			}
		} else {
			w.Mutex.Lock()
			w.ResetRetryCount()
			w.Mutex.Unlock()
		}
	}
}

func (p *Pool) handleUnhealthyWorker(w *Worker, id string) {
	w.Mutex.Lock()
	shouldRetry := w.ShouldRetry(p.maxRetries)
	w.Mutex.Unlock()

	if shouldRetry {
		w.Mutex.Lock()
		retryCount := w.IncrementRetryCount()
		w.Mutex.Unlock()

		p.logger.Debug("retrying unhealthy worker",
			"worker_id", id,
			"retry_count", retryCount,
			"max_retries", p.maxRetries,
		)

		go p.retryWorker(w, id)
	} else {
		p.logger.Debug("terminating unhealthy worker",
			"worker_id", id,
			"health", w.Health,
		)

		if err := p.Terminate(id); err != nil {
			p.logger.Error("failed to terminate unhealthy worker",
				"worker_id", id,
				"error", err,
			)
		}
	}
}

func (p *Pool) retryWorker(w *Worker, id string) {
	time.Sleep(p.retryDelay)

	w.Mutex.Lock()
	status := w.Status
	w.Mutex.Unlock()

	if status == StatusRunning {
		if err := w.Stop(); err != nil {
			p.logger.Error("failed to stop worker for retry",
				"worker_id", id,
				"error", err,
			)
			return
		}

		ctx := context.Background()
		if err := w.Start(ctx, w.configPath); err != nil {
			p.logger.Error("failed to restart worker",
				"worker_id", id,
				"error", err,
			)
		} else {
			p.logger.Debug("worker restarted successfully",
				"worker_id", id,
			)
		}
	}
}

func (p *Pool) Spawn(ctx context.Context, provider, model string, method config.CommunicationMethod, task *execution.Task) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.runningCount() >= p.maxConcurrency {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("spawn worker: %w", ctx.Err())
		default:
			p.cond.Wait()
		}
	}

	id := uuid.Must(uuid.NewV7()).String()
	now := time.Now()
	worker := &Worker{
		ID:              id,
		Provider:        provider,
		Model:           model,
		Method:          method,
		Status:          StatusIdle,
		Task:            task,
		Health:          HealthUnknown,
		LastHealthCheck: now,
		LastOutputTime:  now,
	}

	p.workers[id] = worker

	p.logger.Debug("spawned worker", "worker_id", id, "provider", provider, "model", model)

	return id, nil
}

func (p *Pool) Get(id string) *Worker {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.workers[id]
}

func (p *Pool) List(filterStatus ...WorkerStatus) []*Worker {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(filterStatus) == 0 {
		result := make([]*Worker, 0, len(p.workers))
		for _, w := range p.workers {
			result = append(result, w)
		}
		return result
	}

	var result []*Worker
	filter := filterStatus[0]

	for _, w := range p.workers {
		w.Mutex.Lock()
		status := w.Status
		w.Mutex.Unlock()

		if status == filter {
			result = append(result, w)
		}
	}

	return result
}

func (p *Pool) Terminate(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	worker, exists := p.workers[id]
	if !exists {
		return fmt.Errorf("worker %s not found", id)
	}

	worker.Mutex.Lock()
	if worker.Status == StatusRunning {
		worker.Status = StatusCancelled
	}
	worker.Mutex.Unlock()

	if err := worker.Stop(); err != nil {
		return fmt.Errorf("stop worker %s: %w", id, err)
	}

	p.cond.Signal()

	return nil
}

func (p *Pool) runningCount() int {
	count := 0
	for _, w := range p.workers {
		w.Mutex.Lock()
		status := w.Status
		w.Mutex.Unlock()

		if status == StatusRunning {
			count++
		}
	}
	return count
}

func (p *Pool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := PoolStats{
		Total:          len(p.workers),
		MaxConcurrency: p.maxConcurrency,
	}

	for _, w := range p.workers {
		w.Mutex.Lock()
		status := w.Status
		w.Mutex.Unlock()

		switch status {
		case StatusIdle:
			stats.Idle++
		case StatusRunning:
			stats.Running++
		case StatusCompleted:
			stats.Completed++
		case StatusFailed:
			stats.Failed++
		case StatusCancelled:
			stats.Cancelled++
		}
	}

	return stats
}

type PoolStats struct {
	Total          int
	Idle           int
	Running        int
	Completed      int
	Failed         int
	Cancelled      int
	MaxConcurrency int
}
