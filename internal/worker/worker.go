package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/opencode"
)

func (w *Worker) Start(ctx context.Context, configPath string) error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Status != StatusIdle {
		return fmt.Errorf("worker %s: cannot start, current status: %s", w.ID, w.Status)
	}

	w.configPath = configPath
	w.LastHealthCheck = time.Now()
	w.LastOutputTime = time.Now()
	w.Health = HealthHealthy

	switch w.Method {
	case config.MethodACP:
		if err := w.startACP(ctx); err != nil {
			return fmt.Errorf("start acp: %w", err)
		}
	case config.MethodCLI:
		w.Status = StatusRunning
	case config.MethodAPI:
		return fmt.Errorf("worker %s: API method not implemented", w.ID)
	default:
		return fmt.Errorf("worker %s: unknown communication method: %s", w.ID, w.Method)
	}

	return nil
}

func (w *Worker) startACP(ctx context.Context) error {
	acpCfg := opencode.Config{
		ConfigPath: w.configPath,
		ClientName: "go-ent-worker",
		ClientVer:  "1.0.0",
	}

	client, err := opencode.NewACPClient(ctx, acpCfg)
	if err != nil {
		w.UpdateHealth(HealthUnhealthy, "failed to create ACP client")
		return fmt.Errorf("create ACP client: %w", err)
	}

	if err := client.Initialize(ctx); err != nil {
		client.Close()
		w.UpdateHealth(HealthUnhealthy, "failed to initialize ACP client")
		return fmt.Errorf("initialize ACP client: %w", err)
	}

	if _, err := client.SessionNew(ctx, w.Provider, w.Model, nil); err != nil {
		client.Close()
		w.UpdateHealth(HealthUnhealthy, "failed to create ACP session")
		return fmt.Errorf("create ACP session: %w", err)
	}

	w.acpClient = client
	w.Status = StatusRunning
	w.RecordOutput()

	return nil
}

func (w *Worker) Stop() error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Status == StatusIdle || w.Status == StatusCompleted ||
		w.Status == StatusFailed || w.Status == StatusCancelled {
		return nil
	}

	if w.acpClient != nil && w.acpClient.IsInitialized() && w.acpClient.SessionID() != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = w.acpClient.SessionCancel(ctx, "worker_stop")
	}

	if w.acpClient != nil {
		_ = w.acpClient.Close()
		w.acpClient = nil
	}

	if w.cancel != nil {
		w.cancel()
	}

	if w.cmd != nil && w.cmd.Process != nil {
		if err := w.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("kill worker process: %w", err)
		}
	}

	w.Status = StatusCancelled
	w.Health = HealthUnknown

	return nil
}

func (w *Worker) SendPrompt(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	w.Mutex.Lock()
	if w.Status != StatusRunning {
		w.Mutex.Unlock()
		return "", fmt.Errorf("worker %s: cannot send prompt, current status: %s", w.ID, w.Status)
	}
	w.Mutex.Unlock()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	switch w.Method {
	case config.MethodACP:
		return w.sendACP(ctx, prompt)
	case config.MethodCLI:
		return w.sendCLI(ctx, prompt)
	case config.MethodAPI:
		return "", fmt.Errorf("worker %s: API method not implemented", w.ID)
	default:
		return "", fmt.Errorf("worker %s: unknown communication method: %s", w.ID, w.Method)
	}
}

func (w *Worker) sendACP(ctx context.Context, prompt string) (string, error) {
	if w.acpClient == nil {
		w.Status = StatusFailed
		w.UpdateHealth(HealthUnhealthy, "ACP client not initialized")
		return "", fmt.Errorf("worker %s: ACP client not initialized", w.ID)
	}

	result, err := w.acpClient.SessionPrompt(ctx, prompt, nil, nil)
	if err != nil {
		w.Status = StatusFailed
		w.UpdateHealth(HealthUnhealthy, "ACP prompt failed")
		return "", fmt.Errorf("worker %s: send prompt: %w", w.ID, err)
	}

	w.Output = fmt.Sprintf("Prompt ID: %s, Status: %s", result.PromptID, result.Status)
	w.RecordOutput()

	if result.Status == "executing" {
		go w.monitorACPCaptions(ctx, "")
	}

	return w.Output, nil
}

func (w *Worker) monitorACPCaptions(ctx context.Context, promptID string) {
	updates := w.acpClient.Updates()

	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				w.Mutex.Lock()
				if w.Status == StatusRunning {
					w.Status = StatusCompleted
				}
				w.Mutex.Unlock()
				return
			}

			if update.SessionID == w.acpClient.SessionID() && (promptID == "" || update.PromptID == promptID) {
				w.handleUpdate(update)
			}
		}
	}
}

func (w *Worker) handleUpdate(update opencode.SessionUpdateNotification) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	switch update.Type {
	case "output":
		if update.Data != "" {
			w.Output += update.Data
			w.recordOutputLocked()
		}

	case "progress":
		if update.Message != "" {
			w.Output += fmt.Sprintf("\n[Progress: %.1f%%] %s", update.Progress*100, update.Message)
			w.recordOutputLocked()
		}

	case "tool":
		w.Output += fmt.Sprintf("\n[Tool: %s] %s", update.Tool, update.Status)
		w.recordOutputLocked()

	case "complete":
		w.Status = StatusCompleted
		w.Output += "\n[Complete]"
		w.recordOutputLocked()

	case "error":
		w.Status = StatusFailed
		if update.Error != "" {
			w.Output += fmt.Sprintf("\n[Error] %s", update.Error)
		}
		w.updateHealthLocked(HealthUnhealthy, "prompt execution failed")

	case "cancelled":
		w.Status = StatusCancelled
		w.Output += "\n[Cancelled]"

	default:
		slog.Debug("unknown update type", "type", update.Type)
	}
}

func (w *Worker) recordOutputLocked() {
	w.LastOutputTime = time.Now()
	w.Health = HealthHealthy
	w.UnhealthySince = time.Time{}
}

func (w *Worker) updateHealthLocked(newStatus HealthStatus, reason string) {
	oldStatus := w.Health
	w.Health = newStatus

	if oldStatus != newStatus {
		w.HealthCheckCount++

		if newStatus != HealthHealthy && w.UnhealthySince.IsZero() {
			w.UnhealthySince = time.Now()
		} else if newStatus == HealthHealthy {
			w.UnhealthySince = time.Time{}
		}

		if reason != "" {
			slog.Debug("worker health changed",
				"worker_id", w.ID,
				"old_health", oldStatus,
				"new_health", newStatus,
				"reason", reason,
			)
		}
	}
}

func (w *Worker) sendCLI(ctx context.Context, prompt string) (string, error) {
	args := []string{"run"}
	if w.Model != "" {
		args = append(args, "--model", w.Model)
	}
	args = append(args, "--prompt", prompt)

	cmd := exec.CommandContext(ctx, "opencode", args...) // #nosec G204 -- controlled binary path

	if w.configPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("OPENCODE_CONFIG=%s", w.configPath))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		w.Status = StatusFailed
		w.UpdateHealth(HealthUnhealthy, "CLI execution failed")
		return string(output), fmt.Errorf("opencode run: %w", err)
	}

	w.Output = string(output)
	w.Status = StatusCompleted
	w.RecordOutput()

	return w.Output, nil
}

func (w *Worker) GetStatus() WorkerStatus {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Method == config.MethodACP && w.acpClient != nil {
		if !w.acpClient.IsInitialized() {
			w.Status = StatusFailed
			w.Health = HealthUnhealthy
		}
	}

	if w.cmd != nil && w.cmd.Process != nil {
		state, err := w.cmd.Process.Wait()
		if err == nil && !state.Success() {
			w.Status = StatusFailed
			w.Health = HealthUnhealthy
		}
	}

	return w.Status
}

func (w *Worker) CheckHealth(ctx context.Context, timeout time.Duration) HealthStatus {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.LastHealthCheck = time.Now()

	if w.Status != StatusRunning {
		w.Health = HealthUnknown
		return HealthUnknown
	}

	if time.Since(w.LastOutputTime) > timeout {
		w.Health = HealthTimeout
		if w.UnhealthySince.IsZero() {
			w.UnhealthySince = time.Now()
		}
		return HealthTimeout
	}

	if w.Method == config.MethodACP {
		if w.acpClient == nil || !w.acpClient.IsInitialized() {
			w.Health = HealthUnhealthy
			if w.UnhealthySince.IsZero() {
				w.UnhealthySince = time.Now()
			}
			return HealthUnhealthy
		}
	} else {
		if w.cmd == nil || w.cmd.Process == nil {
			w.Health = HealthUnhealthy
			if w.UnhealthySince.IsZero() {
				w.UnhealthySince = time.Now()
			}
			return HealthUnhealthy
		}

		state, err := w.cmd.Process.Wait()
		if err == nil {
			if !state.Success() {
				w.Health = HealthUnhealthy
				if w.UnhealthySince.IsZero() {
					w.UnhealthySince = time.Now()
				}
				return HealthUnhealthy
			}
		}
	}

	w.Health = HealthHealthy
	w.UnhealthySince = time.Time{}
	return HealthHealthy
}

func (w *Worker) IsHealthy() bool {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	return w.Health == HealthHealthy && w.Status == StatusRunning
}

func (w *Worker) IsTimedOut(timeout time.Duration) bool {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Status != StatusRunning {
		return false
	}

	return time.Since(w.LastOutputTime) > timeout
}

func (w *Worker) RecordOutput() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	w.recordOutputLocked()
}

func (w *Worker) UpdateHealth(newStatus HealthStatus, reason string) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	w.updateHealthLocked(newStatus, reason)
}

func (w *Worker) ResetRetryCount() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.RetryCount = 0
}

func (w *Worker) IncrementRetryCount() int {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	w.RetryCount++
	return w.RetryCount
}

func (w *Worker) ShouldRetry(maxRetries int) bool {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	return w.RetryCount < maxRetries
}

func (w *Worker) StreamUpdates() <-chan opencode.SessionUpdateNotification {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.acpClient == nil {
		return nil
	}

	return w.acpClient.Updates()
}
