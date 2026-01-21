package aggregator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/memory"
	"github.com/victorzhuk/go-ent/internal/openspec"
)

type MergeStrategy string

const (
	MergeFirstSuccess      MergeStrategy = "first_success"
	MergeLastSuccess       MergeStrategy = "last_success"
	MergeConcat            MergeStrategy = "concat"
	MergeJSON              MergeStrategy = "json_merge"
	MergeByPriority        MergeStrategy = "priority"
	MergePreferredProvider MergeStrategy = "preferred_provider"
)

type MergeConfig struct {
	Strategy   MergeStrategy
	Priority   []string
	Preferred  string
	OnConflict string
}

type MergeDecision struct {
	Strategy       MergeStrategy
	SelectedWorker string
	SkippedWorkers []string
	Reason         string
	Timestamp      time.Time
}

type MergedOutput struct {
	Content       string
	SourceWorkers []string
	Metadata      map[string]interface{}
	Decisions     []MergeDecision
	Errors        []string
}

type FileEdit struct {
	WorkerID  string
	FilePath  string
	StartTime time.Time
	EndTime   time.Time
	Operation string
}

type Conflict struct {
	FilePath   string
	Workers    []string
	Resolution string
	DetectedAt time.Time
}

type WorkerResult struct {
	WorkerID      string
	Provider      string
	Model         string
	Status        string
	Output        string
	Error         string
	StartTime     time.Time
	EndTime       time.Time
	Metadata      map[string]string
	FileEdits     []FileEdit
	HasConflicts  bool
	ConflictCount int
	Cost          float64
	OutputSize    int
}

type AggregatedResult struct {
	Results        map[string]*WorkerResult
	CompletedCount int
	FailedCount    int
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	SuccessRate    float64
	Conflicts      []Conflict
	ConflictCount  int
	MergedOutput   *MergedOutput
	MergeConfig    *MergeConfig
}

type ProviderStats struct {
	Name           string
	TasksCompleted int
	TasksFailed    int
	SuccessRate    float64
	AvgDuration    time.Duration
	TotalCost      float64
	CostPerTask    float64
	OutputSize     int
}

type ExecutionSummary struct {
	StartTime      time.Time
	EndTime        time.Time
	TotalDuration  time.Duration
	TotalTasks     int
	WorkersUsed    int
	Providers      map[string]*ProviderStats
	ProviderCosts  map[string]*ProviderCosts
	WorkerCosts    map[string]*WorkerCost
	OverallSuccess float64
	TotalCost      float64
	Conflicts      []Conflict
	ConflictCount  int
	MergeDecisions []MergeDecision
	MergeStrategy  MergeStrategy
	CostTracking   *config.CostTrackingConfig
}

type WorkerCost struct {
	WorkerID       string
	Provider       string
	Model          string
	Method         string
	TaskCount      int
	TotalCost      float64
	AvgCostPerTask float64
	StartTime      time.Time
	EndTime        time.Time
}

type ProviderCosts struct {
	Provider        string
	TotalCost       float64
	TaskCount       int
	AvgCostPerTask  float64
	Budget          float64
	BudgetUsed      float64
	BudgetRemaining float64
	BudgetExceeded  bool
	Currency        string
}

type Aggregator struct {
	results        map[string]*WorkerResult
	completed      []string
	failed         []string
	expected       map[string]bool
	fileEdits      map[string][]FileEdit
	conflicts      []Conflict
	conflictCount  int
	resolution     string
	mergeConfig    *MergeConfig
	mergedOutput   *MergedOutput
	mergeDecisions []MergeDecision
	workerCosts    map[string]*WorkerCost
	providerCosts  map[string]*ProviderCosts
	costConfig     *config.CostTrackingConfig
	taskTracker    *openspec.TaskTracker
	memory         *memory.MemoryStore
	mu             sync.RWMutex
	startTime      time.Time
	timeout        time.Duration
}

func NewAggregator(timeout time.Duration, mergeConfig *MergeConfig, taskTracker *openspec.TaskTracker, memoryStore *memory.MemoryStore) *Aggregator {
	if mergeConfig == nil {
		mergeConfig = &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
	}
	if memoryStore == nil {
		memoryStore = memory.NewMemoryStore()
	}
	return &Aggregator{
		results:        make(map[string]*WorkerResult),
		completed:      make([]string, 0),
		failed:         make([]string, 0),
		expected:       make(map[string]bool),
		fileEdits:      make(map[string][]FileEdit),
		conflicts:      make([]Conflict, 0),
		conflictCount:  0,
		resolution:     "last_write",
		mergeConfig:    mergeConfig,
		mergeDecisions: make([]MergeDecision, 0),
		workerCosts:    make(map[string]*WorkerCost),
		providerCosts:  make(map[string]*ProviderCosts),
		taskTracker:    taskTracker,
		memory:         memoryStore,
		startTime:      time.Now(),
		timeout:        timeout,
	}
}

func NewAggregatorWithDefaults(taskTracker *openspec.TaskTracker) *Aggregator {
	return NewAggregator(5*time.Minute, nil, taskTracker, memory.NewMemoryStore())
}

func NewAggregatorWithoutTracking(timeout time.Duration, mergeConfig *MergeConfig) *Aggregator {
	return NewAggregator(timeout, mergeConfig, nil, nil)
}

func (a *Aggregator) TrackFileEdit(edit *FileEdit) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.fileEdits[edit.FilePath] = append(a.fileEdits[edit.FilePath], *edit)

	conflict := a.detectConflict(edit)
	if conflict != nil {
		a.conflicts = append(a.conflicts, *conflict)
		a.conflictCount++
		slog.Warn("conflict detected",
			"file", conflict.FilePath,
			"workers", conflict.Workers,
			"resolution", conflict.Resolution,
		)
	}
}

func (a *Aggregator) detectConflict(edit *FileEdit) *Conflict {
	edits, exists := a.fileEdits[edit.FilePath]
	if !exists {
		return nil
	}

	for _, existingEdit := range edits {
		if existingEdit.WorkerID == edit.WorkerID {
			continue
		}

		if a.timeOverlaps(existingEdit.StartTime, existingEdit.EndTime, edit.StartTime, edit.EndTime) {
			return &Conflict{
				FilePath:   edit.FilePath,
				Workers:    []string{existingEdit.WorkerID, edit.WorkerID},
				Resolution: a.resolution,
				DetectedAt: time.Now(),
			}
		}
	}

	return nil
}

func (a *Aggregator) timeOverlaps(start1, end1, start2, end2 time.Time) bool {
	if end1.IsZero() {
		end1 = time.Now()
	}
	if end2.IsZero() {
		end2 = time.Now()
	}
	return start1.Before(end2) && !end1.Before(start2)
}

func (a *Aggregator) ResolveConflicts() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, conflict := range a.conflicts {
		edits := a.fileEdits[conflict.FilePath]
		if len(edits) < 2 {
			continue
		}

		switch conflict.Resolution {
		case "first_write":
			a.applyFirstWriteResolution(conflict, edits)
		case "last_write":
			a.applyLastWriteResolution(conflict, edits)
		case "merge_attempt":
			a.applyMergeAttemptResolution(conflict, edits)
		}
	}
}

func (a *Aggregator) applyFirstWriteResolution(conflict Conflict, edits []FileEdit) {
	if len(edits) == 0 {
		return
	}

	firstEdit := edits[0]
	firstWorkers := map[string]bool{firstEdit.WorkerID: true}

	for _, result := range a.results {
		if firstWorkers[result.WorkerID] {
			continue
		}
		for i := len(result.FileEdits) - 1; i >= 0; i-- {
			if result.FileEdits[i].FilePath == conflict.FilePath {
				result.FileEdits = append(result.FileEdits[:i], result.FileEdits[i+1:]...)
				result.HasConflicts = true
				result.ConflictCount++
			}
		}
	}
}

func (a *Aggregator) applyLastWriteResolution(conflict Conflict, edits []FileEdit) {
	if len(edits) == 0 {
		return
	}

	lastEdit := edits[len(edits)-1]
	lastWorkers := map[string]bool{lastEdit.WorkerID: true}

	for _, result := range a.results {
		if lastWorkers[result.WorkerID] {
			continue
		}
		for i := len(result.FileEdits) - 1; i >= 0; i-- {
			if result.FileEdits[i].FilePath == conflict.FilePath {
				result.FileEdits = append(result.FileEdits[:i], result.FileEdits[i+1:]...)
				result.HasConflicts = true
				result.ConflictCount++
			}
		}
	}
}

func (a *Aggregator) applyMergeAttemptResolution(conflict Conflict, edits []FileEdit) {
	for _, result := range a.results {
		for _, edit := range result.FileEdits {
			if edit.FilePath == conflict.FilePath {
				result.HasConflicts = true
				result.ConflictCount++
			}
		}
	}
}

func (a *Aggregator) AddResult(workerID string, result *WorkerResult) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	if result.WorkerID != workerID {
		return fmt.Errorf("worker ID mismatch: %s != %s", workerID, result.WorkerID)
	}

	oldResult, exists := a.results[workerID]
	if exists {
		if oldResult.Status == "completed" && !contains(a.completed, workerID) {
			a.completed = append(a.completed, workerID)
		} else if oldResult.Status == "failed" || oldResult.Error != "" {
			if !contains(a.failed, workerID) {
				a.failed = append(a.failed, workerID)
			}
		}
	}

	if result.EndTime.IsZero() {
		result.EndTime = time.Now()
	}

	for _, edit := range result.FileEdits {
		a.fileEdits[edit.FilePath] = append(a.fileEdits[edit.FilePath], edit)
		conflict := a.detectConflict(&edit)
		if conflict != nil {
			a.conflicts = append(a.conflicts, *conflict)
			a.conflictCount++
			slog.Warn("conflict detected",
				"file", conflict.FilePath,
				"workers", conflict.Workers,
				"resolution", conflict.Resolution,
			)
		}
	}

	a.results[workerID] = result

	if result.Status == "completed" {
		if !contains(a.completed, workerID) {
			a.completed = append(a.completed, workerID)

			if a.taskTracker != nil && result.Metadata != nil {
				if taskDesc, ok := result.Metadata["task_description"]; ok {
					taskID := a.taskTracker.ExtractTaskID(taskDesc)
					if !taskID.IsZero() {
						notes := fmt.Sprintf("Completed by worker %s (%s/%s)",
							result.WorkerID,
							result.Provider,
							result.Model,
						)
						if err := a.taskTracker.MarkCompleted(taskID, notes); err != nil {
							slog.Warn("failed to mark task completed in registry",
								"task_id", taskID.String(),
								"error", err,
							)
						}
					}
				}
			}

			a.storePattern(result)
		}
		a.failed = removeFromSlice(a.failed, workerID)
	} else if result.Status == "failed" || result.Error != "" {
		if !contains(a.failed, workerID) {
			a.failed = append(a.failed, workerID)

			if a.taskTracker != nil && result.Metadata != nil {
				if taskDesc, ok := result.Metadata["task_description"]; ok {
					taskID := a.taskTracker.ExtractTaskID(taskDesc)
					if !taskID.IsZero() {
						errorMsg := fmt.Sprintf("Worker %s failed: %s", result.WorkerID, result.Error)
						if err := a.taskTracker.MarkFailed(taskID, errorMsg); err != nil {
							slog.Warn("failed to mark task failed in registry",
								"task_id", taskID.String(),
								"error", err,
							)
						}
					}
				}
			}
		}
		a.completed = removeFromSlice(a.completed, workerID)
	}

	if result.Cost > 0 && result.Provider != "" {
		a.workerCosts[workerID] = &WorkerCost{
			WorkerID:       workerID,
			Provider:       result.Provider,
			Model:          result.Model,
			Method:         result.Metadata["method"],
			TaskCount:      1,
			TotalCost:      result.Cost,
			AvgCostPerTask: result.Cost,
			StartTime:      result.StartTime,
			EndTime:        result.EndTime,
		}

		a.updateProviderCost(result.Provider, result.Cost)
	}

	return nil
}

func (a *Aggregator) WaitForAll(timeout time.Duration) (*AggregatedResult, error) {
	if timeout <= 0 {
		timeout = a.timeout
	}

	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	deadline := time.Now().Add(timeout)

	for {
		a.mu.RLock()
		allDone := len(a.results) == len(a.completed)+len(a.failed)
		totalWorkers := len(a.completed) + len(a.failed)
		a.mu.RUnlock()

		if allDone && totalWorkers > 0 {
			break
		}

		if time.Now().After(deadline) {
			a.mu.Lock()
			a.collectTimeouts()
			a.mu.Unlock()
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return a.GetAggregatedResult()
}

func (a *Aggregator) collectTimeouts() {
	for workerID, result := range a.results {
		if result.EndTime.IsZero() {
			result.Status = "timeout"
			result.Error = "worker timed out"
			result.EndTime = time.Now()

			if !contains(a.completed, workerID) && !contains(a.failed, workerID) {
				a.failed = append(a.failed, workerID)
			}
		}
	}
}

func (a *Aggregator) GetResult(workerID string) (*WorkerResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result, exists := a.results[workerID]
	if !exists {
		return nil, fmt.Errorf("result for worker %s not found", workerID)
	}

	return result, nil
}

func (a *Aggregator) AllCompleted() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.expected) == 0 {
		return false
	}

	return len(a.expected) == len(a.completed)+len(a.failed)
}

func (a *Aggregator) FailedWorkers() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]string, len(a.failed))
	copy(result, a.failed)
	return result
}

func (a *Aggregator) GetAggregatedResult() (*AggregatedResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endTime := time.Now()
	duration := endTime.Sub(a.startTime)

	completedCount := len(a.completed)
	failedCount := len(a.failed)
	total := completedCount + failedCount

	var successRate float64
	if total > 0 {
		successRate = float64(completedCount) / float64(total) * 100
	}

	resultsCopy := make(map[string]*WorkerResult, len(a.results))
	for k, v := range a.results {
		resultsCopy[k] = v
	}

	conflictsCopy := make([]Conflict, len(a.conflicts))
	copy(conflictsCopy, a.conflicts)

	configCopy := *a.mergeConfig
	priorityCopy := make([]string, len(a.mergeConfig.Priority))
	copy(priorityCopy, a.mergeConfig.Priority)
	configCopy.Priority = priorityCopy

	return &AggregatedResult{
		Results:        resultsCopy,
		CompletedCount: completedCount,
		FailedCount:    failedCount,
		StartTime:      a.startTime,
		EndTime:        endTime,
		Duration:       duration,
		SuccessRate:    successRate,
		Conflicts:      conflictsCopy,
		ConflictCount:  a.conflictCount,
		MergedOutput:   a.mergedOutput,
		MergeConfig:    &configCopy,
	}, nil
}

func (a *Aggregator) RegisterWorkers(workerIDs []string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, id := range workerIDs {
		a.expected[id] = true
		if _, exists := a.results[id]; !exists {
			a.results[id] = &WorkerResult{
				WorkerID:  id,
				StartTime: time.Now(),
				Status:    "running",
			}
		}
	}
}

func (a *Aggregator) MarkFailed(workerID, errorMsg string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.expected[workerID] = true

	result, exists := a.results[workerID]
	if !exists {
		result = &WorkerResult{
			WorkerID:  workerID,
			StartTime: time.Now(),
		}
		a.results[workerID] = result
	}

	result.Status = "failed"
	result.Error = errorMsg
	result.EndTime = time.Now()

	if !contains(a.completed, workerID) && !contains(a.failed, workerID) {
		a.failed = append(a.failed, workerID)
	}
}

func (a *Aggregator) CompletedWorkers() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]string, len(a.completed))
	copy(result, a.completed)
	return result
}

func (a *Aggregator) SetResolutionStrategy(strategy string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if strategy == "first_write" || strategy == "last_write" || strategy == "merge_attempt" {
		a.resolution = strategy
	}
}

func (a *Aggregator) GetConflicts() []Conflict {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]Conflict, len(a.conflicts))
	copy(result, a.conflicts)
	return result
}

func (a *Aggregator) Merge() (*MergedOutput, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.completed) == 0 {
		return nil, fmt.Errorf("no completed workers to merge")
	}

	output := &MergedOutput{
		SourceWorkers: make([]string, 0),
		Metadata:      make(map[string]interface{}),
		Decisions:     make([]MergeDecision, 0),
		Errors:        make([]string, 0),
	}

	var err error
	switch a.mergeConfig.Strategy {
	case MergeFirstSuccess:
		err = a.mergeFirstSuccess(output)
	case MergeLastSuccess:
		err = a.mergeLastSuccess(output)
	case MergeConcat:
		err = a.mergeConcat(output)
	case MergeJSON:
		err = a.mergeJSON(output)
	case MergeByPriority:
		err = a.mergeByPriority(output)
	case MergePreferredProvider:
		err = a.mergePreferredProvider(output)
	default:
		err = fmt.Errorf("unknown merge strategy: %s", a.mergeConfig.Strategy)
	}

	if err != nil {
		a.logMergeError("merge failed", err)
		output.Errors = append(output.Errors, err.Error())
		return output, err
	}

	a.mergedOutput = output
	a.logMergeSuccess(output)

	return output, nil
}

func (a *Aggregator) mergeFirstSuccess(output *MergedOutput) error {
	successful := a.getSuccessfulResults()
	if len(successful) == 0 {
		return fmt.Errorf("no successful results to merge")
	}

	first := successful[0]
	decision := MergeDecision{
		Strategy:       MergeFirstSuccess,
		SelectedWorker: first.WorkerID,
		SkippedWorkers: a.getSkippedWorkers([]string{first.WorkerID}),
		Reason:         "first successful result selected",
		Timestamp:      time.Now(),
	}

	output.Content = first.Output
	output.SourceWorkers = []string{first.WorkerID}
	output.Metadata["provider"] = first.Provider
	output.Metadata["model"] = first.Model
	output.Decisions = []MergeDecision{decision}

	a.mergeDecisions = append(a.mergeDecisions, decision)

	slog.Info("merge: first_success",
		"worker", first.WorkerID,
		"provider", first.Provider,
		"model", first.Model,
	)

	return nil
}

func (a *Aggregator) mergeLastSuccess(output *MergedOutput) error {
	successful := a.getSuccessfulResults()
	if len(successful) == 0 {
		return fmt.Errorf("no successful results to merge")
	}

	last := successful[len(successful)-1]
	decision := MergeDecision{
		Strategy:       MergeLastSuccess,
		SelectedWorker: last.WorkerID,
		SkippedWorkers: a.getSkippedWorkers([]string{last.WorkerID}),
		Reason:         "last successful result selected",
		Timestamp:      time.Now(),
	}

	output.Content = last.Output
	output.SourceWorkers = []string{last.WorkerID}
	output.Metadata["provider"] = last.Provider
	output.Metadata["model"] = last.Model
	output.Decisions = []MergeDecision{decision}

	a.mergeDecisions = append(a.mergeDecisions, decision)

	slog.Info("merge: last_success",
		"worker", last.WorkerID,
		"provider", last.Provider,
		"model", last.Model,
	)

	return nil
}

func (a *Aggregator) mergeConcat(output *MergedOutput) error {
	successful := a.getSuccessfulResults()
	if len(successful) == 0 {
		return fmt.Errorf("no successful results to merge")
	}

	var builder strings.Builder
	sourceWorkers := make([]string, 0)
	providers := make(map[string]bool)
	models := make(map[string]bool)

	for _, result := range successful {
		builder.WriteString(result.Output)
		builder.WriteString("\n\n")
		sourceWorkers = append(sourceWorkers, result.WorkerID)
		providers[result.Provider] = true
		models[result.Model] = true
	}

	decision := MergeDecision{
		Strategy:       MergeConcat,
		SelectedWorker: "all",
		SkippedWorkers: a.getSkippedWorkers(sourceWorkers),
		Reason:         fmt.Sprintf("concatenated %d worker outputs", len(successful)),
		Timestamp:      time.Now(),
	}

	output.Content = builder.String()
	output.SourceWorkers = sourceWorkers
	output.Metadata["worker_count"] = len(successful)
	output.Decisions = []MergeDecision{decision}

	a.mergeDecisions = append(a.mergeDecisions, decision)

	slog.Info("merge: concat",
		"worker_count", len(successful),
		"total_chars", len(output.Content),
	)

	return nil
}

func (a *Aggregator) mergeJSON(output *MergedOutput) error {
	successful := a.getSuccessfulResults()
	if len(successful) == 0 {
		return fmt.Errorf("no successful results to merge")
	}

	merged := make(map[string]interface{})
	sourceWorkers := make([]string, 0)

	for _, result := range successful {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
			slog.Warn("failed to parse JSON from worker",
				"worker", result.WorkerID,
				"error", err,
			)
			continue
		}

		for k, v := range data {
			if _, exists := merged[k]; exists {
				if a.mergeConfig.OnConflict == "markers" {
					merged[k+"_"+result.WorkerID] = v
				}
			} else {
				merged[k] = v
			}
		}

		sourceWorkers = append(sourceWorkers, result.WorkerID)
	}

	if len(merged) == 0 {
		return fmt.Errorf("no valid JSON to merge")
	}

	mergedJSON, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal merged JSON: %w", err)
	}

	decision := MergeDecision{
		Strategy:       MergeJSON,
		SelectedWorker: fmt.Sprintf("%d workers", len(sourceWorkers)),
		SkippedWorkers: a.getSkippedWorkers(sourceWorkers),
		Reason:         "merged JSON outputs",
		Timestamp:      time.Now(),
	}

	output.Content = string(mergedJSON)
	output.SourceWorkers = sourceWorkers
	output.Metadata["key_count"] = len(merged)
	output.Decisions = []MergeDecision{decision}

	a.mergeDecisions = append(a.mergeDecisions, decision)

	slog.Info("merge: json",
		"worker_count", len(sourceWorkers),
		"key_count", len(merged),
	)

	return nil
}

func (a *Aggregator) mergeByPriority(output *MergedOutput) error {
	if len(a.mergeConfig.Priority) == 0 {
		return fmt.Errorf("priority list is empty")
	}

	priorityMap := make(map[string]int)
	for i, provider := range a.mergeConfig.Priority {
		priorityMap[provider] = i
	}

	successful := a.getSuccessfulResults()
	if len(successful) == 0 {
		return fmt.Errorf("no successful results to merge")
	}

	var selected *WorkerResult
	selectedIndex := -1

	for _, result := range successful {
		if idx, exists := priorityMap[result.Provider]; exists {
			if selectedIndex == -1 || idx < selectedIndex {
				selected = result
				selectedIndex = idx
			}
		}
	}

	if selected == nil {
		return fmt.Errorf("no result from providers in priority list")
	}

	decision := MergeDecision{
		Strategy:       MergeByPriority,
		SelectedWorker: selected.WorkerID,
		SkippedWorkers: a.getSkippedWorkers([]string{selected.WorkerID}),
		Reason:         fmt.Sprintf("selected by priority (%s is #1)", selected.Provider),
		Timestamp:      time.Now(),
	}

	output.Content = selected.Output
	output.SourceWorkers = []string{selected.WorkerID}
	output.Metadata["provider"] = selected.Provider
	output.Metadata["priority"] = selectedIndex
	output.Decisions = []MergeDecision{decision}

	a.mergeDecisions = append(a.mergeDecisions, decision)

	slog.Info("merge: priority",
		"worker", selected.WorkerID,
		"provider", selected.Provider,
		"priority", selectedIndex,
	)

	return nil
}

func (a *Aggregator) mergePreferredProvider(output *MergedOutput) error {
	if a.mergeConfig.Preferred == "" {
		return fmt.Errorf("preferred provider not specified")
	}

	successful := a.getSuccessfulResults()
	if len(successful) == 0 {
		return fmt.Errorf("no successful results to merge")
	}

	var selected *WorkerResult
	for _, result := range successful {
		if result.Provider == a.mergeConfig.Preferred {
			selected = result
			break
		}
	}

	if selected == nil {
		return fmt.Errorf("no result from preferred provider: %s", a.mergeConfig.Preferred)
	}

	decision := MergeDecision{
		Strategy:       MergePreferredProvider,
		SelectedWorker: selected.WorkerID,
		SkippedWorkers: a.getSkippedWorkers([]string{selected.WorkerID}),
		Reason:         fmt.Sprintf("preferred provider: %s", selected.Provider),
		Timestamp:      time.Now(),
	}

	output.Content = selected.Output
	output.SourceWorkers = []string{selected.WorkerID}
	output.Metadata["provider"] = selected.Provider
	output.Metadata["model"] = selected.Model
	output.Decisions = []MergeDecision{decision}

	a.mergeDecisions = append(a.mergeDecisions, decision)

	slog.Info("merge: preferred_provider",
		"worker", selected.WorkerID,
		"provider", selected.Provider,
		"model", selected.Model,
	)

	return nil
}

func (a *Aggregator) getSuccessfulResults() []*WorkerResult {
	results := make([]*WorkerResult, 0)
	for _, workerID := range a.completed {
		if result, exists := a.results[workerID]; exists && result.Error == "" {
			results = append(results, result)
		}
	}
	return results
}

func (a *Aggregator) getSkippedWorkers(selected []string) []string {
	skipped := make([]string, 0)
	skippedMap := make(map[string]bool)
	for _, id := range selected {
		skippedMap[id] = true
	}

	for _, workerID := range a.completed {
		if !skippedMap[workerID] {
			skipped = append(skipped, workerID)
		}
	}

	return skipped
}

func (a *Aggregator) logMergeSuccess(output *MergedOutput) {
	slog.Info("merge completed successfully",
		"strategy", a.mergeConfig.Strategy,
		"worker_count", len(output.SourceWorkers),
		"content_length", len(output.Content),
	)
}

func (a *Aggregator) logMergeError(reason string, err error) {
	slog.Error("merge failed",
		"reason", reason,
		"strategy", a.mergeConfig.Strategy,
		"error", err,
	)
}

func (a *Aggregator) GetMergeConfig() *MergeConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()

	configCopy := *a.mergeConfig
	priorityCopy := make([]string, len(a.mergeConfig.Priority))
	copy(priorityCopy, a.mergeConfig.Priority)
	configCopy.Priority = priorityCopy

	return &configCopy
}

func (a *Aggregator) SetMergeConfig(config *MergeConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.mergeConfig = config
}

func (a *Aggregator) GetMergedOutput() *MergedOutput {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.mergedOutput
}

func (a *Aggregator) GetMergeDecisions() []MergeDecision {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]MergeDecision, len(a.mergeDecisions))
	copy(result, a.mergeDecisions)
	return result
}

func (a *Aggregator) TotalWorkers() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.expected)
}

func (a *Aggregator) SetCostTracking(costConfig *config.CostTrackingConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.costConfig = costConfig
}

func (a *Aggregator) GetCostTracking() *config.CostTrackingConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.costConfig
}

func (a *Aggregator) TrackWorkerCost(workerID, provider, model, method string, cost float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	wc, exists := a.workerCosts[workerID]
	if !exists {
		wc = &WorkerCost{
			WorkerID:  workerID,
			Provider:  provider,
			Model:     model,
			Method:    method,
			StartTime: time.Now(),
		}
		a.workerCosts[workerID] = wc
	}

	wc.TotalCost += cost
	wc.TaskCount++
	if wc.TaskCount > 0 {
		wc.AvgCostPerTask = wc.TotalCost / float64(wc.TaskCount)
	}
	wc.EndTime = time.Now()

	slog.Debug("worker cost tracked",
		"worker", workerID,
		"provider", provider,
		"model", model,
		"cost", cost,
		"total_cost", wc.TotalCost,
		"task_count", wc.TaskCount,
	)

	a.updateProviderCost(provider, cost)
}

func (a *Aggregator) updateProviderCost(provider string, cost float64) {
	pc, exists := a.providerCosts[provider]
	if !exists {
		pc = &ProviderCosts{
			Provider: provider,
			Currency: "USD",
		}
		if a.costConfig != nil && a.costConfig.Enabled {
			pc.Budget = a.costConfig.GlobalBudget
		}
		a.providerCosts[provider] = pc
	}

	pc.TotalCost += cost
	pc.TaskCount++
	pc.BudgetUsed = pc.TotalCost
	pc.BudgetRemaining = pc.Budget - pc.BudgetUsed
	pc.BudgetExceeded = pc.BudgetUsed > pc.Budget

	if pc.TaskCount > 0 {
		pc.AvgCostPerTask = pc.TotalCost / float64(pc.TaskCount)
	}

	slog.Debug("provider cost updated",
		"provider", provider,
		"cost", cost,
		"total_cost", pc.TotalCost,
		"task_count", pc.TaskCount,
		"budget_used", pc.BudgetUsed,
		"budget_remaining", pc.BudgetRemaining,
		"budget_exceeded", pc.BudgetExceeded,
	)
}

func (a *Aggregator) GetWorkerCost(workerID string) (*WorkerCost, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	wc, exists := a.workerCosts[workerID]
	if !exists {
		return nil, fmt.Errorf("worker cost not found: %s", workerID)
	}

	wcCopy := *wc
	return &wcCopy, nil
}

func (a *Aggregator) GetAllWorkerCosts() map[string]*WorkerCost {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*WorkerCost, len(a.workerCosts))
	for k, v := range a.workerCosts {
		wcCopy := *v
		result[k] = &wcCopy
	}
	return result
}

func (a *Aggregator) GetProviderCosts(provider string) (*ProviderCosts, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	pc, exists := a.providerCosts[provider]
	if !exists {
		return nil, fmt.Errorf("provider cost not found: %s", provider)
	}

	pcCopy := *pc
	return &pcCopy, nil
}

func (a *Aggregator) GetAllProviderCosts() map[string]*ProviderCosts {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*ProviderCosts, len(a.providerCosts))
	for k, v := range a.providerCosts {
		pcCopy := *v
		result[k] = &pcCopy
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func (a *Aggregator) GenerateSummary() *ExecutionSummary {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endTime := time.Now()
	if len(a.completed) > 0 || len(a.failed) > 0 {
		for _, result := range a.results {
			if !result.EndTime.IsZero() && result.EndTime.After(endTime) {
				endTime = result.EndTime
			}
		}
	}

	totalDuration := endTime.Sub(a.startTime)
	totalTasks := len(a.completed) + len(a.failed)
	providers := make(map[string]*ProviderStats)
	providerCosts := make(map[string]*ProviderCosts)
	totalCost := 0.0

	providerTasks := make(map[string][]*WorkerResult)
	for _, result := range a.results {
		if result.Provider == "" {
			continue
		}
		providerTasks[result.Provider] = append(providerTasks[result.Provider], result)
		totalCost += result.Cost
	}

	for provider, results := range providerTasks {
		completed := 0
		failed := 0
		totalDurationProv := time.Duration(0)
		totalCostProv := 0.0
		outputSize := 0

		for _, result := range results {
			if result.Status == "completed" && result.Error == "" {
				completed++
			} else {
				failed++
			}
			if !result.EndTime.IsZero() && !result.StartTime.IsZero() {
				totalDurationProv += result.EndTime.Sub(result.StartTime)
			}
			totalCostProv += result.Cost
			outputSize += result.OutputSize
		}

		var successRate float64
		totalProv := completed + failed
		if totalProv > 0 {
			successRate = float64(completed) / float64(totalProv) * 100
		}

		var avgDuration time.Duration
		if completed > 0 {
			avgDuration = totalDurationProv / time.Duration(completed)
		}

		var costPerTask float64
		if totalProv > 0 {
			costPerTask = totalCostProv / float64(totalProv)
		}

		providers[provider] = &ProviderStats{
			Name:           provider,
			TasksCompleted: completed,
			TasksFailed:    failed,
			SuccessRate:    successRate,
			AvgDuration:    avgDuration,
			TotalCost:      totalCostProv,
			CostPerTask:    costPerTask,
			OutputSize:     outputSize,
		}
	}

	for provider, provCost := range a.providerCosts {
		provCopy := *provCost
		providerCosts[provider] = &provCopy
	}

	workerCosts := make(map[string]*WorkerCost, len(a.workerCosts))
	for workerID, wc := range a.workerCosts {
		wcCopy := *wc
		workerCosts[workerID] = &wcCopy
	}

	var overallSuccess float64
	if totalTasks > 0 {
		overallSuccess = float64(len(a.completed)) / float64(totalTasks) * 100
	}

	mergeDecisions := make([]MergeDecision, len(a.mergeDecisions))
	copy(mergeDecisions, a.mergeDecisions)

	conflicts := make([]Conflict, len(a.conflicts))
	copy(conflicts, a.conflicts)

	var mergeStrategy MergeStrategy
	if a.mergeConfig != nil {
		mergeStrategy = a.mergeConfig.Strategy
	}

	return &ExecutionSummary{
		StartTime:      a.startTime,
		EndTime:        endTime,
		TotalDuration:  totalDuration,
		TotalTasks:     totalTasks,
		WorkersUsed:    len(a.expected),
		Providers:      providers,
		ProviderCosts:  providerCosts,
		WorkerCosts:    workerCosts,
		OverallSuccess: overallSuccess,
		TotalCost:      totalCost,
		Conflicts:      conflicts,
		ConflictCount:  len(a.conflicts),
		MergeDecisions: mergeDecisions,
		MergeStrategy:  mergeStrategy,
		CostTracking:   a.costConfig,
	}
}

func (s *ExecutionSummary) ToMarkdown() string {
	var builder strings.Builder

	builder.WriteString("# Execution Summary\n\n")
	builder.WriteString(fmt.Sprintf("**Start Time:** %s\n", s.StartTime.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("**End Time:** %s\n", s.EndTime.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("**Total Duration:** %s\n", s.TotalDuration))
	builder.WriteString(fmt.Sprintf("**Total Tasks:** %d\n", s.TotalTasks))
	builder.WriteString(fmt.Sprintf("**Workers Used:** %d\n", s.WorkersUsed))
	builder.WriteString(fmt.Sprintf("**Overall Success Rate:** %.2f%%\n", s.OverallSuccess))
	builder.WriteString(fmt.Sprintf("**Total Cost:** $%.4f\n\n", s.TotalCost))

	builder.WriteString("## Worker Cost Breakdown\n\n")
	if len(s.WorkerCosts) > 0 {
		builder.WriteString("| Worker | Provider | Model | Method | Tasks | Total Cost | Avg Cost/Task |\n")
		builder.WriteString("|--------|----------|-------|--------|-------|------------|--------------|\n")
		for _, wc := range s.WorkerCosts {
			builder.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %d | $%.4f | $%.4f |\n",
				wc.WorkerID, wc.Provider, wc.Model, wc.Method, wc.TaskCount, wc.TotalCost, wc.AvgCostPerTask))
		}
		builder.WriteString("\n")
	} else {
		builder.WriteString("No worker cost data available.\n\n")
	}

	builder.WriteString("## Provider Cost Details\n\n")
	if len(s.ProviderCosts) > 0 {
		for provider, pc := range s.ProviderCosts {
			builder.WriteString(fmt.Sprintf("### %s\n\n", provider))
			builder.WriteString(fmt.Sprintf("- **Total Cost:** $%.4f\n", pc.TotalCost))
			builder.WriteString(fmt.Sprintf("- **Task Count:** %d\n", pc.TaskCount))
			builder.WriteString(fmt.Sprintf("- **Average Cost Per Task:** $%.4f\n", pc.AvgCostPerTask))
			builder.WriteString(fmt.Sprintf("- **Budget:** $%.4f\n", pc.Budget))
			builder.WriteString(fmt.Sprintf("- **Budget Used:** $%.4f\n", pc.BudgetUsed))
			builder.WriteString(fmt.Sprintf("- **Budget Remaining:** $%.4f\n", pc.BudgetRemaining))
			if pc.BudgetExceeded {
				builder.WriteString("- **Budget Exceeded:** ⚠️ YES\n")
			} else {
				builder.WriteString("- **Budget Exceeded:** ✅ NO\n")
			}
			builder.WriteString("\n")
		}
	} else {
		builder.WriteString("No provider cost data available.\n\n")
	}

	builder.WriteString("## Provider Statistics\n\n")
	for provider, stats := range s.Providers {
		builder.WriteString(fmt.Sprintf("### %s\n\n", provider))
		builder.WriteString(fmt.Sprintf("- **Tasks Completed:** %d\n", stats.TasksCompleted))
		builder.WriteString(fmt.Sprintf("- **Tasks Failed:** %d\n", stats.TasksFailed))
		builder.WriteString(fmt.Sprintf("- **Success Rate:** %.2f%%\n", stats.SuccessRate))
		builder.WriteString(fmt.Sprintf("- **Average Duration:** %s\n", stats.AvgDuration))
		builder.WriteString(fmt.Sprintf("- **Total Cost:** $%.4f\n", stats.TotalCost))
		builder.WriteString(fmt.Sprintf("- **Cost Per Task:** $%.4f\n", stats.CostPerTask))
		builder.WriteString(fmt.Sprintf("- **Output Size:** %d bytes\n\n", stats.OutputSize))
	}

	builder.WriteString("## Cost Breakdown\n\n")
	builder.WriteString("| Provider | Cost | Percentage |\n")
	builder.WriteString("|----------|------|------------|\n")
	for provider, stats := range s.Providers {
		percentage := 0.0
		if s.TotalCost > 0 {
			percentage = (stats.TotalCost / s.TotalCost) * 100
		}
		builder.WriteString(fmt.Sprintf("| %s | $%.4f | %.2f%% |\n", provider, stats.TotalCost, percentage))
	}
	builder.WriteString("\n")

	builder.WriteString("## Timing Breakdown\n\n")
	builder.WriteString("| Provider | Avg Duration | Tasks |\n")
	builder.WriteString("|----------|--------------|-------|\n")
	for provider, stats := range s.Providers {
		totalTasks := stats.TasksCompleted + stats.TasksFailed
		builder.WriteString(fmt.Sprintf("| %s | %s | %d |\n", provider, stats.AvgDuration, totalTasks))
	}
	builder.WriteString("\n")

	builder.WriteString("## Conflicts\n\n")
	if s.ConflictCount > 0 {
		for _, conflict := range s.Conflicts {
			builder.WriteString(fmt.Sprintf("### Conflict: %s\n\n", conflict.FilePath))
			builder.WriteString(fmt.Sprintf("- **Workers:** %s\n", strings.Join(conflict.Workers, ", ")))
			builder.WriteString(fmt.Sprintf("- **Resolution:** %s\n", conflict.Resolution))
			builder.WriteString(fmt.Sprintf("- **Detected At:** %s\n\n", conflict.DetectedAt.Format(time.RFC3339)))
		}
	} else {
		builder.WriteString("No conflicts detected.\n\n")
	}

	builder.WriteString("## Merge Summary\n\n")
	if len(s.MergeDecisions) > 0 {
		builder.WriteString(fmt.Sprintf("**Merge Strategy:** %s\n\n", s.MergeStrategy))
		for i, decision := range s.MergeDecisions {
			builder.WriteString(fmt.Sprintf("### Decision %d\n\n", i+1))
			builder.WriteString(fmt.Sprintf("- **Strategy:** %s\n", decision.Strategy))
			builder.WriteString(fmt.Sprintf("- **Selected Worker:** %s\n", decision.SelectedWorker))
			if len(decision.SkippedWorkers) > 0 {
				builder.WriteString(fmt.Sprintf("- **Skipped Workers:** %s\n", strings.Join(decision.SkippedWorkers, ", ")))
			}
			builder.WriteString(fmt.Sprintf("- **Reason:** %s\n", decision.Reason))
			builder.WriteString(fmt.Sprintf("- **Timestamp:** %s\n\n", decision.Timestamp.Format(time.RFC3339)))
		}
	} else {
		builder.WriteString("No merge decisions made.\n\n")
	}

	return builder.String()
}

func (s *ExecutionSummary) ToJSON() (string, error) {
	data := struct {
		StartTime      string                        `json:"start_time"`
		EndTime        string                        `json:"end_time"`
		TotalDuration  string                        `json:"total_duration"`
		TotalTasks     int                           `json:"total_tasks"`
		WorkersUsed    int                           `json:"workers_used"`
		Providers      map[string]*ProviderStatsJSON `json:"providers"`
		ProviderCosts  map[string]*ProviderCostsJSON `json:"provider_costs"`
		WorkerCosts    []WorkerCostJSON              `json:"worker_costs"`
		OverallSuccess float64                       `json:"overall_success"`
		TotalCost      float64                       `json:"total_cost"`
		Conflicts      []ConflictJSON                `json:"conflicts"`
		ConflictCount  int                           `json:"conflict_count"`
		MergeDecisions []MergeDecisionJSON           `json:"merge_decisions"`
		MergeStrategy  string                        `json:"merge_strategy"`
		CostTracking   *CostTrackingJSON             `json:"cost_tracking,omitempty"`
	}{
		StartTime:      s.StartTime.Format(time.RFC3339),
		EndTime:        s.EndTime.Format(time.RFC3339),
		TotalDuration:  s.TotalDuration.String(),
		TotalTasks:     s.TotalTasks,
		WorkersUsed:    s.WorkersUsed,
		OverallSuccess: s.OverallSuccess,
		TotalCost:      s.TotalCost,
		ConflictCount:  s.ConflictCount,
		MergeStrategy:  string(s.MergeStrategy),
		Providers:      make(map[string]*ProviderStatsJSON),
		ProviderCosts:  make(map[string]*ProviderCostsJSON),
		WorkerCosts:    make([]WorkerCostJSON, 0, len(s.WorkerCosts)),
		Conflicts:      make([]ConflictJSON, len(s.Conflicts)),
		MergeDecisions: make([]MergeDecisionJSON, len(s.MergeDecisions)),
	}

	if s.CostTracking != nil {
		data.CostTracking = &CostTrackingJSON{
			Enabled:        s.CostTracking.Enabled,
			GlobalBudget:   s.CostTracking.GlobalBudget,
			ResetPeriod:    s.CostTracking.ResetPeriod.String(),
			PersistHistory: s.CostTracking.PersistHistory,
			HistoryFile:    s.CostTracking.HistoryFile,
		}
	}

	for provider, stats := range s.Providers {
		data.Providers[provider] = &ProviderStatsJSON{
			Name:           stats.Name,
			TasksCompleted: stats.TasksCompleted,
			TasksFailed:    stats.TasksFailed,
			SuccessRate:    stats.SuccessRate,
			AvgDuration:    stats.AvgDuration.String(),
			TotalCost:      stats.TotalCost,
			CostPerTask:    stats.CostPerTask,
			OutputSize:     stats.OutputSize,
		}
	}

	for provider, pc := range s.ProviderCosts {
		data.ProviderCosts[provider] = &ProviderCostsJSON{
			Provider:        provider,
			TotalCost:       pc.TotalCost,
			TaskCount:       pc.TaskCount,
			AvgCostPerTask:  pc.AvgCostPerTask,
			Budget:          pc.Budget,
			BudgetUsed:      pc.BudgetUsed,
			BudgetRemaining: pc.BudgetRemaining,
			BudgetExceeded:  pc.BudgetExceeded,
			Currency:        pc.Currency,
		}
	}

	for _, wc := range s.WorkerCosts {
		data.WorkerCosts = append(data.WorkerCosts, WorkerCostJSON{
			WorkerID:       wc.WorkerID,
			Provider:       wc.Provider,
			Model:          wc.Model,
			Method:         wc.Method,
			TaskCount:      wc.TaskCount,
			TotalCost:      wc.TotalCost,
			AvgCostPerTask: wc.AvgCostPerTask,
			StartTime:      wc.StartTime.Format(time.RFC3339),
			EndTime:        wc.EndTime.Format(time.RFC3339),
		})
	}

	for i, conflict := range s.Conflicts {
		data.Conflicts[i] = ConflictJSON{
			FilePath:   conflict.FilePath,
			Workers:    conflict.Workers,
			Resolution: conflict.Resolution,
			DetectedAt: conflict.DetectedAt.Format(time.RFC3339),
		}
	}

	for i, decision := range s.MergeDecisions {
		data.MergeDecisions[i] = MergeDecisionJSON{
			Strategy:       string(decision.Strategy),
			SelectedWorker: decision.SelectedWorker,
			SkippedWorkers: decision.SkippedWorkers,
			Reason:         decision.Reason,
			Timestamp:      decision.Timestamp.Format(time.RFC3339),
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal summary to JSON: %w", err)
	}

	return string(jsonData), nil
}

type ProviderStatsJSON struct {
	Name           string  `json:"name"`
	TasksCompleted int     `json:"tasks_completed"`
	TasksFailed    int     `json:"tasks_failed"`
	SuccessRate    float64 `json:"success_rate"`
	AvgDuration    string  `json:"avg_duration"`
	TotalCost      float64 `json:"total_cost"`
	CostPerTask    float64 `json:"cost_per_task"`
	OutputSize     int     `json:"output_size"`
}

type ConflictJSON struct {
	FilePath   string   `json:"file_path"`
	Workers    []string `json:"workers"`
	Resolution string   `json:"resolution"`
	DetectedAt string   `json:"detected_at"`
}

type MergeDecisionJSON struct {
	Strategy       string   `json:"strategy"`
	SelectedWorker string   `json:"selected_worker"`
	SkippedWorkers []string `json:"skipped_workers"`
	Reason         string   `json:"reason"`
	Timestamp      string   `json:"timestamp"`
}

type CostTrackingJSON struct {
	Enabled        bool    `json:"enabled"`
	GlobalBudget   float64 `json:"global_budget"`
	ResetPeriod    string  `json:"reset_period"`
	PersistHistory bool    `json:"persist_history"`
	HistoryFile    string  `json:"history_file"`
}

type ProviderCostsJSON struct {
	Provider        string  `json:"provider"`
	TotalCost       float64 `json:"total_cost"`
	TaskCount       int     `json:"task_count"`
	AvgCostPerTask  float64 `json:"avg_cost_per_task"`
	Budget          float64 `json:"budget"`
	BudgetUsed      float64 `json:"budget_used"`
	BudgetRemaining float64 `json:"budget_remaining"`
	BudgetExceeded  bool    `json:"budget_exceeded"`
	Currency        string  `json:"currency"`
}

type WorkerCostJSON struct {
	WorkerID       string  `json:"worker_id"`
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	Method         string  `json:"method"`
	TaskCount      int     `json:"task_count"`
	TotalCost      float64 `json:"total_cost"`
	AvgCostPerTask float64 `json:"avg_cost_per_task"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
}

func (a *Aggregator) StorePattern(result *WorkerResult) error {
	pattern := ExtractPattern(result)
	if err := a.memory.Store(pattern); err != nil {
		return fmt.Errorf("store pattern: %w", err)
	}

	slog.Info("pattern stored",
		"worker", result.WorkerID,
		"provider", result.Provider,
		"model", result.Model,
		"task_type", pattern.TaskType,
		"success", pattern.Success,
		"cost", result.Cost,
	)

	return nil
}

func (a *Aggregator) storePattern(result *WorkerResult) {
	pattern := ExtractPattern(result)
	if err := a.memory.Store(pattern); err != nil {
		slog.Warn("failed to store pattern", "error", err)
	} else {
		slog.Debug("pattern stored",
			"worker", result.WorkerID,
			"provider", result.Provider,
			"model", result.Model,
			"task_type", pattern.TaskType,
			"success", pattern.Success,
		)
	}
}

func ExtractPattern(result *WorkerResult) *memory.Pattern {
	pattern := &memory.Pattern{
		Provider:     result.Provider,
		Model:        result.Model,
		Success:      result.Status == "completed",
		Cost:         result.Cost,
		OutputSize:   result.OutputSize,
		ErrorPattern: "",
	}

	if result.Metadata != nil {
		pattern.TaskType = result.Metadata["task_type"]
		if pattern.TaskType == "" {
			pattern.TaskType = result.Metadata["type"]
		}
		if method, ok := result.Metadata["method"]; ok {
			pattern.Method = method
		}
	}

	if pattern.Method == "" {
		if result.Metadata != nil && result.Metadata["method"] != "" {
			pattern.Method = result.Metadata["method"]
		} else {
			pattern.Method = "acp"
		}
	}

	pattern.FileCount = len(result.FileEdits)
	if result.Metadata != nil {
		if contextSize, ok := result.Metadata["context_size"]; ok {
			if _, err := fmt.Sscanf(contextSize, "%d", &pattern.ContextSize); err == nil {
			}
		}
	}

	if !result.StartTime.IsZero() && !result.EndTime.IsZero() {
		pattern.Duration = result.EndTime.Sub(result.StartTime)
	}

	if result.Error != "" {
		pattern.ErrorPattern = classifyError(result.Error)
	}

	return pattern
}

func classifyError(errorMsg string) string {
	errorMsgLower := strings.ToLower(errorMsg)

	if strings.Contains(errorMsgLower, "timeout") || strings.Contains(errorMsgLower, "timed out") {
		return "timeout"
	}
	if strings.Contains(errorMsgLower, "rate limit") || strings.Contains(errorMsgLower, "too many requests") {
		return "rate_limit"
	}
	if strings.Contains(errorMsgLower, "authentication") || strings.Contains(errorMsgLower, "unauthorized") {
		return "auth_error"
	}
	if strings.Contains(errorMsgLower, "context") && strings.Contains(errorMsgLower, "limit") {
		return "context_limit"
	}
	if strings.Contains(errorMsgLower, "memory") && strings.Contains(errorMsgLower, "exceed") {
		return "memory_exceeded"
	}
	if strings.Contains(errorMsgLower, "connection") || strings.Contains(errorMsgLower, "network") {
		return "network_error"
	}

	return "unknown_error"
}

func (a *Aggregator) GetMemoryStore() *memory.MemoryStore {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.memory
}

func (a *Aggregator) SetMemoryStore(memoryStore *memory.MemoryStore) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.memory = memoryStore
}
