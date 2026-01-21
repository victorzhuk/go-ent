package tools

import (
	"context"
	"testing"
	"time"

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/worker"
)

func TestWorkerOutput(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	tests := []struct {
		name        string
		setup       func() string
		input       WorkerOutputInput
		wantError   bool
		errorPrefix string
	}{
		{
			name: "completed worker with output",
			setup: func() string {
				task := execution.NewTask("test task")
				spawnReq := worker.SpawnRequest{
					Method:   config.MethodCLI,
					Provider: "glm",
					Model:    "glm-4",
					Task:     task,
				}
				workerID, err := manager.Spawn(context.Background(), spawnReq)
				if err != nil {
					t.Fatalf("failed to spawn worker: %v", err)
				}
				w := manager.Get(workerID)
				if w == nil {
					t.Fatal("worker is nil after spawn")
				}
				w.Mutex.Lock()
				w.Output = "test output\nmore output"
				w.Status = worker.StatusCompleted
				w.Mutex.Unlock()
				return workerID
			},
			input:     WorkerOutputInput{},
			wantError: false,
		},
		{
			name: "empty worker id",
			setup: func() string {
				return ""
			},
			input:       WorkerOutputInput{},
			wantError:   true,
			errorPrefix: "worker_id is required",
		},
		{
			name: "worker not found",
			setup: func() string {
				return "nonexistent-id"
			},
			input: WorkerOutputInput{
				WorkerID: "nonexistent-id",
			},
			wantError:   true,
			errorPrefix: "get output:",
		},
		{
			name: "empty worker id",
			setup: func() string {
				return ""
			},
			input:       WorkerOutputInput{},
			wantError:   true,
			errorPrefix: "worker_id is required",
		},
		{
			name: "running worker with no output",
			setup: func() string {
				task := execution.NewTask("test task")
				spawnReq := worker.SpawnRequest{
					Method:   config.MethodCLI,
					Provider: "glm",
					Model:    "glm-4",
					Task:     task,
				}
				workerID, err := manager.Spawn(context.Background(), spawnReq)
				if err != nil {
					t.Fatalf("failed to spawn worker: %v", err)
				}
				return workerID
			},
			input:     WorkerOutputInput{},
			wantError: false,
		},
		{
			name: "failed worker with output",
			setup: func() string {
				task := execution.NewTask("test task")
				spawnReq := worker.SpawnRequest{
					Method:   config.MethodCLI,
					Provider: "glm",
					Model:    "glm-4",
					Task:     task,
				}
				workerID, err := manager.Spawn(context.Background(), spawnReq)
				if err != nil {
					t.Fatalf("failed to spawn worker: %v", err)
				}
				w := manager.Get(workerID)
				if w == nil {
					t.Fatal("worker is nil after spawn")
				}
				w.Mutex.Lock()
				w.Output = "error occurred"
				w.Status = worker.StatusFailed
				w.Mutex.Unlock()
				return workerID
			},
			input:     WorkerOutputInput{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workerID := tt.setup()
			if workerID != "" {
				tt.input.WorkerID = workerID
			}

			ctx := context.Background()
			result, _, err := handler(ctx, nil, tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorPrefix != "" && len(err.Error()) >= len(tt.errorPrefix) && err.Error()[:len(tt.errorPrefix)] != tt.errorPrefix {
					t.Errorf("error prefix mismatch: got %s, want %s", err.Error()[:len(tt.errorPrefix)], tt.errorPrefix)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result.Content) == 0 {
				t.Error("expected content in result")
			}
		})
	}
}

func TestWorkerOutputWithFilter(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "INFO: Starting\nERROR: Something went wrong\nDEBUG: Done\nINFO: Finished"
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	tests := []struct {
		name       string
		pattern    string
		wantOutput string
	}{
		{
			name:       "filters by error pattern",
			pattern:    "ERROR:.*",
			wantOutput: "ERROR: Something went wrong",
		},
		{
			name:       "filters multiple matches",
			pattern:    "INFO:.*",
			wantOutput: "INFO: Starting\nINFO: Finished",
		},
		{
			name:       "empty pattern returns all output",
			pattern:    "",
			wantOutput: "INFO: Starting\nERROR: Something went wrong\nDEBUG: Done\nINFO: Finished",
		},
		{
			name:       "no matches returns empty",
			pattern:    "WARN:.*",
			wantOutput: "",
		},
		{
			name:       "case-insensitive filter",
			pattern:    "(?i)error:.*",
			wantOutput: "ERROR: Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := WorkerOutputInput{
				WorkerID: workerID,
				Filter:   tt.pattern,
			}

			ctx := context.Background()
			result, responseAny, err := handler(ctx, nil, input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Content) == 0 {
				t.Error("expected content in result")
			}

			response, ok := responseAny.(WorkerOutputResponse)
			if !ok {
				t.Fatal("expected WorkerOutputResponse")
			}

			if response.Output != tt.wantOutput {
				t.Errorf("expected output %q, got %q", tt.wantOutput, response.Output)
			}

			if tt.pattern != "" {
				if response.LineCount != 4 {
					t.Errorf("expected total line count 4, got %d", response.LineCount)
				}
			}
		})
	}
}

func TestWorkerOutputWithLimit(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "line 1\nline 2\nline 3\nline 4\nline 5"
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	tests := []struct {
		name      string
		limit     int
		wantCount int
		truncated bool
	}{
		{
			name:      "limit 2 returns first 2 lines",
			limit:     2,
			wantCount: 2,
			truncated: true,
		},
		{
			name:      "limit greater than output count",
			limit:     10,
			wantCount: 5,
			truncated: false,
		},
		{
			name:      "no limit returns all output",
			limit:     0,
			wantCount: 5,
			truncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := WorkerOutputInput{
				WorkerID: workerID,
				Limit:    tt.limit,
			}

			ctx := context.Background()
			result, responseAny, err := handler(ctx, nil, input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Content) == 0 {
				t.Error("expected content in result")
			}

			response, ok := responseAny.(WorkerOutputResponse)
			if !ok {
				t.Fatal("expected WorkerOutputResponse")
			}

			if tt.limit > 0 && tt.limit < 5 {
				if response.Truncated != tt.truncated {
					t.Errorf("expected truncated %v, got %v", tt.truncated, response.Truncated)
				}
			}

			if response.LineCount != 5 {
				t.Errorf("expected total line count 5, got %d", response.LineCount)
			}

			if response.Truncated && tt.limit > 0 {
				if len(response.Output) == 0 {
					t.Error("expected output but got empty")
				}
			}
		})
	}
}

func TestWorkerOutputWithSince(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "initial output"
	w.LastOutputTime = time.Now().Add(-5 * time.Minute)
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	time.Sleep(10 * time.Millisecond)

	w.Mutex.Lock()
	w.Output += "\nrecent output"
	w.LastOutputTime = time.Now()
	w.Mutex.Unlock()

	since := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)

	input := WorkerOutputInput{
		WorkerID: workerID,
		Since:    since,
	}

	ctx := context.Background()
	result, responseAny, err := handler(ctx, nil, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result")
	}

	response, ok := responseAny.(WorkerOutputResponse)
	if !ok {
		t.Fatal("expected WorkerOutputResponse")
	}

	if response.Output == "" {
		t.Error("expected output after timestamp")
	}

	if response.LastUpdated == "" {
		t.Error("expected valid LastUpdated timestamp")
	}

	_, parseErr := time.Parse(time.RFC3339, response.LastUpdated)
	if parseErr != nil {
		t.Errorf("expected valid RFC3339 timestamp, got error: %v", parseErr)
	}
}

func TestWorkerOutputInvalidSince(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "test output"
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	input := WorkerOutputInput{
		WorkerID: workerID,
		Since:    "invalid-timestamp",
	}

	ctx := context.Background()
	result, _, err := handler(ctx, nil, input)

	if err == nil {
		t.Error("expected error for invalid timestamp but got none")
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result even with error")
	}
}

func TestWorkerOutputInvalidRegex(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "some output"
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	input := WorkerOutputInput{
		WorkerID: workerID,
		Filter:   "[invalid(",
	}

	ctx := context.Background()
	result, _, err := handler(ctx, nil, input)

	if err == nil {
		t.Error("expected error for invalid regex but got none")
	}

	if len(result.Content) == 0 {
		t.Error("expected content in result even with error")
	}
}

func TestWorkerOutputResponseStructure(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()
	handler := makeWorkerOutputHandler(manager)

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "test output line 1\ntest output line 2"
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	input := WorkerOutputInput{
		WorkerID: workerID,
		Filter:   "line 1",
		Limit:    10,
	}

	ctx := context.Background()
	result, responseAny, err := handler(ctx, nil, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	response, ok := responseAny.(WorkerOutputResponse)
	if !ok {
		t.Fatal("expected WorkerOutputResponse")
	}

	if response.WorkerID != workerID {
		t.Errorf("expected worker ID %s, got %s", workerID, response.WorkerID)
	}

	if response.Output != "test output line 1" {
		t.Errorf("expected filtered output 'test output line 1', got %q", response.Output)
	}

	if response.LineCount != 2 {
		t.Errorf("expected total line count 2, got %d", response.LineCount)
	}

	if response.LastUpdated == "" {
		t.Error("expected valid LastUpdated timestamp")
	}

	_, parseErr := time.Parse(time.RFC3339, response.LastUpdated)
	if parseErr != nil {
		t.Errorf("expected valid RFC3339 timestamp, got error: %v", parseErr)
	}

	if response.Truncated {
		t.Error("expected not truncated")
	}
}

func TestGetWorkerOutput(t *testing.T) {
	manager := worker.NewWorkerManagerWithoutTracking()

	task := execution.NewTask("test task")
	spawnReq := worker.SpawnRequest{
		Method:   config.MethodCLI,
		Provider: "glm",
		Model:    "glm-4",
		Task:     task,
	}
	workerID, err := manager.Spawn(context.Background(), spawnReq)
	if err != nil {
		t.Fatalf("failed to spawn worker: %v", err)
	}
	w := manager.Get(workerID)
	if w == nil {
		t.Fatal("worker is nil after spawn")
	}
	w.Mutex.Lock()
	w.Output = "line 1\nline 2\nline 3"
	w.Status = worker.StatusCompleted
	w.Mutex.Unlock()

	tests := []struct {
		name       string
		workerID   string
		pattern    string
		limit      int
		wantOutput string
		wantError  bool
	}{
		{
			name:       "completed worker with no filter",
			workerID:   workerID,
			pattern:    "",
			limit:      0,
			wantOutput: "line 1\nline 2\nline 3",
			wantError:  false,
		},
		{
			name:       "completed worker with filter",
			workerID:   workerID,
			pattern:    "line [13]",
			limit:      0,
			wantOutput: "line 1\nline 3",
			wantError:  false,
		},
		{
			name:       "completed worker with limit",
			workerID:   workerID,
			pattern:    "",
			limit:      2,
			wantOutput: "line 1\nline 2",
			wantError:  false,
		},
		{
			name:       "completed worker no matches",
			workerID:   workerID,
			pattern:    "ERROR:.*",
			limit:      0,
			wantOutput: "",
			wantError:  false,
		},
		{
			name:       "worker not found",
			workerID:   "nonexistent-id",
			pattern:    "",
			limit:      0,
			wantOutput: "",
			wantError:  true,
		},
		{
			name:       "invalid regex pattern",
			workerID:   workerID,
			pattern:    "[invalid(",
			limit:      0,
			wantOutput: "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := worker.WorkerOutputRequest{
				WorkerID: tt.workerID,
				Filter:   tt.pattern,
				Limit:    tt.limit,
			}

			resp, err := manager.GetOutput(req)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.Output != tt.wantOutput {
				t.Errorf("expected output %q, got %q", tt.wantOutput, resp.Output)
			}

			if tt.limit > 0 && tt.limit < 3 {
				if !resp.Truncated {
					t.Error("expected truncated to be true")
				}
			} else {
				if resp.Truncated {
					t.Error("expected truncated to be false")
				}
			}
		})
	}
}
