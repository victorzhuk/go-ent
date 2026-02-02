package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

const (
	changesBucket  = "changes"
	tasksBucket    = "tasks"
	depsBucket     = "deps"
	blockersBucket = "blockers"
	runtimeBucket  = "runtime"
	metaBucket     = "meta"
)

type ChangeMetadata struct {
	ID         string       `json:"id"`
	Title      string       `json:"title"`
	Status     ChangeStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	TaskCount  int          `json:"task_count"`
	Completed  int          `json:"completed"`
	InProgress int          `json:"in_progress"`
	Blocked    int          `json:"blocked"`
}

type Task struct {
	ID         string     `json:"id"`
	ChangeID   string     `json:"change_id"`
	TaskNum    string     `json:"task_num"`
	Content    string     `json:"content"`
	Status     TaskStatus `json:"status"`
	Priority   int        `json:"priority"`
	DependsOn  []string   `json:"depends_on"`
	SourceLine int        `json:"source_line"`
	SyncedAt   time.Time  `json:"synced_at"`
}

type DependencyInfo struct {
	TaskID        string   `json:"task_id"`
	DependsOn     []string `json:"depends_on"`
	DependedBy    []string `json:"depended_by"`
	IsBlocked     bool     `json:"is_blocked"`
	BlockingTasks []string `json:"blocking_tasks"`
}

type RuntimeState struct {
	TaskID    string     `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Assignee  string     `json:"assignee"`
	Session   string     `json:"session"`
	StartedAt time.Time  `json:"started_at"`
	Notes     []string   `json:"notes"`
}

type SyncMeta struct {
	Version      int              `json:"version"`
	LastFullSync time.Time        `json:"last_full_sync"`
	FileMTimes   map[string]int64 `json:"file_mtimes"`
}

type BoltStore struct {
	db   *bbolt.DB
	path string
	mu   sync.RWMutex
}

func NewBoltStore(rootPath string) (*BoltStore, error) {
	cachePath := filepath.Join(rootPath, ".cache", "openspec.db")

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o750); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	db, err := bbolt.Open(cachePath, 0o600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open boltdb: %w", err)
	}

	store := &BoltStore{
		db:   db,
		path: cachePath,
	}

	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return store, nil
}

func (s *BoltStore) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	s.db, err = bbolt.Open(s.path, 0o600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return fmt.Errorf("open boltdb: %w", err)
	}

	if err := s.initBuckets(); err != nil {
		return fmt.Errorf("init buckets: %w", err)
	}

	return nil
}

func (s *BoltStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *BoltStore) initBuckets() error {
	buckets := []string{changesBucket, tasksBucket, depsBucket, blockersBucket, runtimeBucket, metaBucket}

	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
}

func (s *BoltStore) GetTask(changeID, taskNum string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var task *Task
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		key := taskKey(changeID, taskNum)
		data := b.Get([]byte(key))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &task)
	})
	if err != nil {
		return nil, fmt.Errorf("get task %s/%s: %w", changeID, taskNum, err)
	}

	return task, nil
}

func (s *BoltStore) PutTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task == nil {
		return fmt.Errorf("nil task")
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		key := taskKey(task.ChangeID, task.TaskNum)
		return b.Put([]byte(key), data)
	})
	if err != nil {
		return fmt.Errorf("put task %s/%s: %w", task.ChangeID, task.TaskNum, err)
	}

	return nil
}

func (s *BoltStore) DeleteTask(changeID, taskNum string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		key := taskKey(changeID, taskNum)
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("delete task %s/%s: %w", changeID, taskNum, err)
	}

	return nil
}

func (s *BoltStore) GetChange(changeID string) (*ChangeMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var change *ChangeMetadata
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(changesBucket))
		if b == nil {
			return fmt.Errorf("changes bucket not found")
		}

		data := b.Get([]byte(changeID))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &change)
	})
	if err != nil {
		return nil, fmt.Errorf("get change %s: %w", changeID, err)
	}

	return change, nil
}

func (s *BoltStore) PutChange(change *ChangeMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if change == nil {
		return fmt.Errorf("nil change")
	}

	data, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("marshal change: %w", err)
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(changesBucket))
		if b == nil {
			return fmt.Errorf("changes bucket not found")
		}

		return b.Put([]byte(change.ID), data)
	})
	if err != nil {
		return fmt.Errorf("put change %s: %w", change.ID, err)
	}

	return nil
}

func (s *BoltStore) GetDeps(changeID, taskNum string) (*DependencyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var deps *DependencyInfo
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(depsBucket))
		if b == nil {
			return fmt.Errorf("deps bucket not found")
		}

		key := taskKey(changeID, taskNum)
		data := b.Get([]byte(key))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &deps)
	})
	if err != nil {
		return nil, fmt.Errorf("get deps %s/%s: %w", changeID, taskNum, err)
	}

	return deps, nil
}

func (s *BoltStore) PutDeps(deps *DependencyInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if deps == nil {
		return fmt.Errorf("nil deps")
	}

	data, err := json.Marshal(deps)
	if err != nil {
		return fmt.Errorf("marshal deps: %w", err)
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(depsBucket))
		if b == nil {
			return fmt.Errorf("deps bucket not found")
		}

		key := deps.TaskID
		return b.Put([]byte(key), data)
	})
	if err != nil {
		return fmt.Errorf("put deps: %w", err)
	}

	return nil
}

func (s *BoltStore) GetRuntimeState(changeID, taskNum string) (*RuntimeState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var state *RuntimeState
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(runtimeBucket))
		if b == nil {
			return fmt.Errorf("runtime bucket not found")
		}

		key := taskKey(changeID, taskNum)
		data := b.Get([]byte(key))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &state)
	})
	if err != nil {
		return nil, fmt.Errorf("get runtime state %s/%s: %w", changeID, taskNum, err)
	}

	return state, nil
}

func (s *BoltStore) PutRuntimeState(state *RuntimeState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state == nil {
		return fmt.Errorf("nil runtime state")
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal runtime state: %w", err)
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(runtimeBucket))
		if b == nil {
			return fmt.Errorf("runtime bucket not found")
		}

		return b.Put([]byte(state.TaskID), data)
	})
	if err != nil {
		return fmt.Errorf("put runtime state %s: %w", state.TaskID, err)
	}

	return nil
}

func (s *BoltStore) GetSyncMeta() (*SyncMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var meta *SyncMeta
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(metaBucket))
		if b == nil {
			return fmt.Errorf("meta bucket not found")
		}

		data := b.Get([]byte("sync_meta"))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &meta)
	})
	if err != nil {
		return nil, fmt.Errorf("get sync meta: %w", err)
	}

	return meta, nil
}

func (s *BoltStore) PutSyncMeta(meta *SyncMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if meta == nil {
		return fmt.Errorf("nil sync meta")
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal sync meta: %w", err)
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(metaBucket))
		if b == nil {
			return fmt.Errorf("meta bucket not found")
		}

		return b.Put([]byte("sync_meta"), data)
	})
	if err != nil {
		return fmt.Errorf("put sync meta: %w", err)
	}

	return nil
}

func taskKey(changeID, taskNum string) string {
	return fmt.Sprintf("%s:%s", changeID, taskNum)
}

func (s *BoltStore) RebuildFromMarkdown(rootPath string) error {
	openspecDir := filepath.Join(rootPath, "openspec", "changes")

	// Get stored mtimes for incremental sync
	storedMeta, err := s.GetSyncMeta()
	if err != nil {
		storedMeta = &SyncMeta{FileMTimes: make(map[string]int64)}
	}
	if storedMeta.FileMTimes == nil {
		storedMeta.FileMTimes = make(map[string]int64)
	}

	changes, tasks, deps, fileMTimes, err := s.walkOpenspecDirWithMTimes(openspecDir, storedMeta.FileMTimes)
	if err != nil {
		return fmt.Errorf("walk openspec dir: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err = s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range []string{changesBucket, tasksBucket, depsBucket} {
			b := tx.Bucket([]byte(bucketName))
			if err := b.ForEach(func(k, v []byte) error {
				return b.Delete(k)
			}); err != nil {
				return fmt.Errorf("clear bucket %s: %w", bucketName, err)
			}
		}

		for _, change := range changes {
			data, err := json.Marshal(change)
			if err != nil {
				return fmt.Errorf("marshal change %s: %w", change.ID, err)
			}
			if err := tx.Bucket([]byte(changesBucket)).Put([]byte(change.ID), data); err != nil {
				return fmt.Errorf("put change %s: %w", change.ID, err)
			}
		}

		for _, task := range tasks {
			data, err := json.Marshal(task)
			if err != nil {
				return fmt.Errorf("marshal task %s: %w", task.ID, err)
			}
			key := taskKey(task.ChangeID, task.TaskNum)
			if err := tx.Bucket([]byte(tasksBucket)).Put([]byte(key), data); err != nil {
				return fmt.Errorf("put task %s: %w", task.ID, err)
			}
		}

		for _, dep := range deps {
			data, err := json.Marshal(dep)
			if err != nil {
				return fmt.Errorf("marshal deps %s: %w", dep.TaskID, err)
			}
			if err := tx.Bucket([]byte(depsBucket)).Put([]byte(dep.TaskID), data); err != nil {
				return fmt.Errorf("put deps %s: %w", dep.TaskID, err)
			}
		}

		meta := &SyncMeta{
			Version:      1,
			LastFullSync: time.Now(),
			FileMTimes:   fileMTimes,
		}
		metaData, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshal sync meta: %w", err)
		}
		if err := tx.Bucket([]byte(metaBucket)).Put([]byte("sync_meta"), metaData); err != nil {
			return fmt.Errorf("put sync meta: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("rebuild database: %w", err)
	}

	return nil
}

func (s *BoltStore) walkOpenspecDir(openspecDir string) ([]*ChangeMetadata, []*Task, []*DependencyInfo, map[string]int64, error) {
	return s.walkOpenspecDirWithMTimes(openspecDir, nil)
}

func (s *BoltStore) walkOpenspecDirWithMTimes(openspecDir string, storedMTimes map[string]int64) ([]*ChangeMetadata, []*Task, []*DependencyInfo, map[string]int64, error) {
	var changes []*ChangeMetadata
	var tasks []*Task
	var deps []*DependencyInfo
	fileMTimes := make(map[string]int64)

	entries, err := os.ReadDir(openspecDir)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("read openspec dir: %w", err)
	}

	parseCount := 0
	skipCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		changeID := entry.Name()
		changeDir := filepath.Join(openspecDir, changeID)

		tasksPath := filepath.Join(changeDir, "tasks.md")
		proposalPath := filepath.Join(changeDir, "proposal.md")

		// Check if tasks.md needs to be reparsed
		tasksInfo, err := os.Stat(tasksPath)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("stat tasks %s: %w", changeID, err)
		}
		tasksMTime := tasksInfo.ModTime().Unix()
		fileMTimes[tasksPath] = tasksMTime

		// Check if file unchanged and we can load from cache
		fileUnchanged := false
		if storedMTimes != nil {
			if storedMTime, exists := storedMTimes[tasksPath]; exists && storedMTime == tasksMTime {
				fileUnchanged = true
			}
		}

		var changeTasks []*Task
		var title string

		if fileUnchanged {
			// Load from existing BoltDB data instead of re-parsing
			existingChange, err := s.GetChange(changeID)
			if err == nil && existingChange != nil {
				title = existingChange.Title
				cachedTasks, err := s.GetTasksByChange(changeID)
				if err == nil && len(cachedTasks) > 0 {
					changeTasks = cachedTasks
					skipCount++
				} else {
					fileUnchanged = false
				}
			} else {
				fileUnchanged = false
			}
		}

		// Parse if file changed or cache miss
		if !fileUnchanged {
			title, err = s.ParseProposalTitle(proposalPath)
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("parse proposal %s: %w", changeID, err)
			}

			changeTasks, err = s.ParseTasksFile(tasksPath, changeID, &fileMTimes)
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("parse tasks %s: %w", changeID, err)
			}
			parseCount++
		}

		completed, inProgress, blocked := s.CountTaskStatuses(changeTasks)

		var status ChangeStatus
		if completed == len(changeTasks) && len(changeTasks) > 0 {
			status = StatusApproved
		} else if completed > 0 {
			status = StatusActive
		} else {
			status = StatusDraft
		}

		change := &ChangeMetadata{
			ID:         changeID,
			Title:      title,
			Status:     status,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			TaskCount:  len(changeTasks),
			Completed:  completed,
			InProgress: inProgress,
			Blocked:    blocked,
		}

		changes = append(changes, change)
		tasks = append(tasks, changeTasks...)

		changeDeps := s.BuildDependencies(changeTasks)
		deps = append(deps, changeDeps...)
	}

	if skipCount > 0 {
		fmt.Fprintf(os.Stderr, "BoltDB sync: parsed %d changes, skipped %d unchanged\n", parseCount, skipCount)
	}

	return changes, tasks, deps, fileMTimes, nil
}

func (s *BoltStore) ParseProposalTitle(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read proposal: %w", err)
	}

	lines := string(data)
	for i, line := range splitLines(lines) {
		if i > 10 {
			break
		}
		if len(line) > 0 && line[0] == '#' {
			return trimLeft(line, "# "), nil
		}
	}

	return "", fmt.Errorf("no title found in proposal")
}

func (s *BoltStore) ParseTasksFile(path string, changeID string, fileMTimes *map[string]int64) ([]*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat tasks: %w", err)
	}
	(*fileMTimes)[path] = info.ModTime().Unix()

	var tasks []*Task
	lines := string(data)
	taskLines := splitLines(lines)

	for i, line := range taskLines {
		task := s.parseTaskLine(line, changeID, i+1)
		if task != nil {
			task.ID = taskKey(changeID, task.TaskNum)
			task.SyncedAt = time.Now()
			task.DependsOn = s.extractDependencies(lines, i)
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

func (s *BoltStore) parseTaskLine(line, changeID string, lineNum int) *Task {
	if len(line) < 2 || (line[0:2] != "- " && line[0:2] != "* ") {
		return nil
	}

	content := trimLeft(line[2:], " ")

	if len(content) < 6 || content[0:3] != "[ ]" && content[0:3] != "[x]" {
		return nil
	}

	completed := content[0:3] == "[x]"
	content = trimLeft(content[3:], " ")

	if len(content) < 2 || content[0:2] != "**" {
		return nil
	}

	content = content[2:]
	endIdx := findIndex(content, "**")
	if endIdx == -1 {
		return nil
	}

	taskNum := content[:endIdx]
	content = trimLeft(content[endIdx+2:], " ")

	status := TaskPending
	if completed {
		status = TaskCompleted
	}

	return &Task{
		ChangeID:   changeID,
		TaskNum:    taskNum,
		Content:    content,
		Status:     status,
		Priority:   0,
		SourceLine: lineNum,
	}
}

func (s *BoltStore) extractDependencies(lines string, taskLineIdx int) []string {
	taskLines := splitLines(lines)
	var deps []string

	for i := taskLineIdx + 1; i < len(taskLines); i++ {
		line := trimLeft(taskLines[i], " \t")
		if len(line) == 0 {
			continue
		}

		if len(line) > 2 && (line[0:2] == "- " || line[0:2] == "* ") {
			break
		}

		if len(line) > 12 && (line[0:13] == "Dependencies:" || line[0:13] == "dependencies:") {
			depContent := trimLeft(line[13:], " ")
			depNums := splitWords(depContent)
			for _, num := range depNums {
				if num != "" {
					deps = append(deps, num)
				}
			}
			break
		}

		if len(line) > 2 && line[0:2] == "# " || line[0:3] == "## " {
			break
		}
	}

	return deps
}

func (s *BoltStore) CountTaskStatuses(tasks []*Task) (completed, inProgress, blocked int) {
	for _, task := range tasks {
		switch task.Status {
		case TaskCompleted:
			completed++
		case TaskInProgress:
			inProgress++
		}
	}
	return completed, inProgress, blocked
}

func (s *BoltStore) BuildDependencies(tasks []*Task) []*DependencyInfo {
	taskMap := make(map[string]*Task)
	for _, task := range tasks {
		taskMap[task.TaskNum] = task
	}

	var deps []*DependencyInfo
	for _, task := range tasks {
		dep := &DependencyInfo{
			TaskID:    task.ID,
			DependsOn: task.DependsOn,
		}

		var blockingTasks []string
		for _, depNum := range task.DependsOn {
			if depTask, exists := taskMap[depNum]; exists {
				dep.DependedBy = append(dep.DependedBy, depTask.ID)
				if depTask.Status != TaskCompleted {
					dep.IsBlocked = true
					blockingTasks = append(blockingTasks, depTask.ID)
				}
			}
		}
		dep.BlockingTasks = blockingTasks

		deps = append(deps, dep)
	}

	return deps
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimLeft(s, cutset string) string {
	for len(s) > 0 {
		found := false
		for _, c := range cutset {
			if len(s) > 0 && rune(s[0]) == c {
				s = s[1:]
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return s
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func splitWords(s string) []string {
	var words []string
	start := -1
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' || s[i] == ' ' || s[i] == '\t' {
			if start != -1 {
				words = append(words, s[start:i])
				start = -1
			}
		} else if start == -1 {
			start = i
		}
	}
	return words
}

func (s *BoltStore) ListAllChanges() ([]*ChangeMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var changes []*ChangeMetadata
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(changesBucket))
		if b == nil {
			return fmt.Errorf("changes bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			var change ChangeMetadata
			if err := json.Unmarshal(v, &change); err != nil {
				return fmt.Errorf("unmarshal change %s: %w", string(k), err)
			}
			changes = append(changes, &change)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("list changes: %w", err)
	}

	return changes, nil
}

func (s *BoltStore) ListTasks(changeID, status string) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*Task
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			var task Task
			if err := json.Unmarshal(v, &task); err != nil {
				return fmt.Errorf("unmarshal task %s: %w", string(k), err)
			}

			if changeID != "" && task.ChangeID != changeID {
				return nil
			}

			if status != "" && string(task.Status) != status {
				return nil
			}

			tasks = append(tasks, &task)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	return tasks, nil
}

func (s *BoltStore) GetTasksByChange(changeID string) ([]*Task, error) {
	return s.ListTasks(changeID, "")
}
