package worker

import (
	"context"
	"fmt"
	"os/exec"
)

func (w *Worker) Start(ctx context.Context, configPath string) error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Status != StatusIdle {
		return fmt.Errorf("worker %s: cannot start, current status: %s", w.ID, w.Status)
	}

	w.configPath = configPath

	switch w.Method {
	case MethodACP:
		if err := w.startACP(ctx); err != nil {
			return fmt.Errorf("start acp: %w", err)
		}
	case MethodCLI:
		w.Status = StatusRunning
	case MethodAPI:
		return fmt.Errorf("worker %s: API method not implemented", w.ID)
	default:
		return fmt.Errorf("worker %s: unknown communication method: %s", w.ID, w.Method)
	}

	return nil
}

func (w *Worker) startACP(ctx context.Context) error {
	cmdCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	cmd := exec.CommandContext(cmdCtx, "opencode", "acp") // #nosec G204 -- controlled binary path

	if w.configPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("OPENCODE_CONFIG=%s", w.configPath))
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("exec opencode acp: %w", err)
	}

	w.cmd = cmd
	w.Status = StatusRunning

	return nil
}

func (w *Worker) Stop() error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Status == StatusIdle || w.Status == StatusCompleted ||
		w.Status == StatusFailed || w.Status == StatusCancelled {
		return nil
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

	return nil
}

func (w *Worker) SendPrompt(ctx context.Context, prompt string) (string, error) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.Status != StatusRunning {
		return "", fmt.Errorf("worker %s: cannot send prompt, current status: %s", w.ID, w.Status)
	}

	switch w.Method {
	case MethodACP:
		return "", fmt.Errorf("worker %s: ACP prompt requires ACP client (task 2.1)", w.ID)
	case MethodCLI:
		return w.sendCLI(ctx, prompt)
	case MethodAPI:
		return "", fmt.Errorf("worker %s: API method not implemented", w.ID)
	default:
		return "", fmt.Errorf("worker %s: unknown communication method: %s", w.ID, w.Method)
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
		return string(output), fmt.Errorf("opencode run: %w", err)
	}

	w.Output = string(output)
	w.Status = StatusCompleted

	return w.Output, nil
}

func (w *Worker) GetStatus() WorkerStatus {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.cmd != nil && w.cmd.Process != nil {
		state, err := w.cmd.Process.Wait()
		if err == nil && !state.Success() {
			w.Status = StatusFailed
		}
	}

	return w.Status
}
