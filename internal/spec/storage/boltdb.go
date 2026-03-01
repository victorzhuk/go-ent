package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"

	"github.com/victorzhuk/go-ent/internal/spec"
	"github.com/victorzhuk/go-ent/internal/spec/parser"
)

const (
	changesBucket  = "changes"
	tasksBucket    = "tasks"
	depsBucket     = "deps"
	blockersBucket = "blockers"
	runtimeBucket  = "runtime"
	metaBucket     = "meta"

	boltOpenTimeout = 5 * time.Second
	filePerms       = 0o600
	dirPerms        = 0o750
)

type BoltStore struct {
	db     *bbolt.DB
	path   string
	mu     sync.RWMutex
	parser parser.TaskParser
}

func NewBoltStore(rootPath string) (*BoltStore, error) {
	return NewBoltStoreWithParser(rootPath, parser.NewTaskParser())
}

func NewBoltStoreWithParser(rootPath string, p parser.TaskParser) (*BoltStore, error) {
	cachePath := filepath.Join(rootPath, ".cache", "openspec.db")

	if err := os.MkdirAll(filepath.Dir(cachePath), dirPerms); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	db, err := bbolt.Open(cachePath, filePerms, &bbolt.Options{Timeout: boltOpenTimeout})
	if err != nil {
		return nil, fmt.Errorf("open boltdb: %w", err)
	}

	store := &BoltStore{
		db:     db,
		path:   cachePath,
		parser: p,
	}

	if err := store.initBuckets(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return store, nil
}

func (s *BoltStore) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	s.db, err = bbolt.Open(s.path, filePerms, &bbolt.Options{Timeout: boltOpenTimeout})
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

func (s *BoltStore) view(bucket string, fn func(*bbolt.Bucket) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("%w: %s", ErrBucketNotFound, bucket)
		}
		return fn(b)
	})
}

func (s *BoltStore) update(bucket string, fn func(*bbolt.Bucket) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("%w: %s", ErrBucketNotFound, bucket)
		}
		return fn(b)
	})
}

func (s *BoltStore) GetTask(changeID, taskNum string) (*spec.Task, error) {
	var task *spec.Task
	err := s.view(tasksBucket, func(b *bbolt.Bucket) error {
		data := b.Get([]byte(taskKey(changeID, taskNum)))
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

func (s *BoltStore) PutTask(task *spec.Task) error {
	if task == nil {
		return fmt.Errorf("%w: nil task", ErrInvalidInput)
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	err = s.update(tasksBucket, func(b *bbolt.Bucket) error {
		return b.Put([]byte(taskKey(task.ChangeID, task.TaskNum)), data)
	})
	if err != nil {
		return fmt.Errorf("put task %s/%s: %w", task.ChangeID, task.TaskNum, err)
	}
	return nil
}

func (s *BoltStore) DeleteTask(changeID, taskNum string) error {
	err := s.update(tasksBucket, func(b *bbolt.Bucket) error {
		return b.Delete([]byte(taskKey(changeID, taskNum)))
	})
	if err != nil {
		return fmt.Errorf("delete task %s/%s: %w", changeID, taskNum, err)
	}
	return nil
}

func (s *BoltStore) GetChange(changeID string) (*spec.ChangeMetadata, error) {
	var change *spec.ChangeMetadata
	err := s.view(changesBucket, func(b *bbolt.Bucket) error {
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

func (s *BoltStore) PutChange(change *spec.ChangeMetadata) error {
	if change == nil {
		return fmt.Errorf("%w: nil change", ErrInvalidInput)
	}

	data, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("marshal change: %w", err)
	}

	err = s.update(changesBucket, func(b *bbolt.Bucket) error {
		return b.Put([]byte(change.ID), data)
	})
	if err != nil {
		return fmt.Errorf("put change %s: %w", change.ID, err)
	}
	return nil
}

func (s *BoltStore) GetDeps(changeID, taskNum string) (*spec.DependencyInfo, error) {
	var deps *spec.DependencyInfo
	err := s.view(depsBucket, func(b *bbolt.Bucket) error {
		data := b.Get([]byte(taskKey(changeID, taskNum)))
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

func (s *BoltStore) PutDeps(deps *spec.DependencyInfo) error {
	if deps == nil {
		return fmt.Errorf("%w: nil deps", ErrInvalidInput)
	}

	data, err := json.Marshal(deps)
	if err != nil {
		return fmt.Errorf("marshal deps: %w", err)
	}

	err = s.update(depsBucket, func(b *bbolt.Bucket) error {
		return b.Put([]byte(deps.TaskID), data)
	})
	if err != nil {
		return fmt.Errorf("put deps: %w", err)
	}
	return nil
}

func (s *BoltStore) GetRuntimeState(changeID, taskNum string) (*spec.RuntimeState, error) {
	var state *spec.RuntimeState
	err := s.view(runtimeBucket, func(b *bbolt.Bucket) error {
		data := b.Get([]byte(taskKey(changeID, taskNum)))
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

func (s *BoltStore) PutRuntimeState(state *spec.RuntimeState) error {
	if state == nil {
		return fmt.Errorf("%w: nil runtime state", ErrInvalidInput)
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal runtime state: %w", err)
	}

	err = s.update(runtimeBucket, func(b *bbolt.Bucket) error {
		return b.Put([]byte(state.TaskID), data)
	})
	if err != nil {
		return fmt.Errorf("put runtime state %s: %w", state.TaskID, err)
	}
	return nil
}

func (s *BoltStore) GetSyncMeta() (*spec.SyncMeta, error) {
	var meta *spec.SyncMeta
	err := s.view(metaBucket, func(b *bbolt.Bucket) error {
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

func (s *BoltStore) PutSyncMeta(meta *spec.SyncMeta) error {
	if meta == nil {
		return fmt.Errorf("%w: nil sync meta", ErrInvalidInput)
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal sync meta: %w", err)
	}

	err = s.update(metaBucket, func(b *bbolt.Bucket) error {
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

func TaskKey(changeID, taskNum string) string {
	return taskKey(changeID, taskNum)
}

func (s *BoltStore) DeleteChange(changeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		changeBucket := tx.Bucket([]byte(changesBucket))
		if changeBucket != nil {
			if err := changeBucket.Delete([]byte(changeID)); err != nil {
				return fmt.Errorf("delete change %s: %w", changeID, err)
			}
		}

		tasksB := tx.Bucket([]byte(tasksBucket))
		if tasksB != nil {
			c := tasksB.Cursor()
			prefix := []byte(changeID + ":")
			for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
				if err := c.Delete(); err != nil {
					return fmt.Errorf("delete task %s: %w", string(k), err)
				}
			}
		}

		depsB := tx.Bucket([]byte(depsBucket))
		if depsB != nil {
			taskPrefix := changeID + ":"
			c := depsB.Cursor()
			for k, _ := c.Seek([]byte(taskPrefix)); k != nil && bytes.HasPrefix(k, prefix(taskPrefix)); k, _ = c.Next() {
				if err := c.Delete(); err != nil {
					return fmt.Errorf("delete deps %s: %w", string(k), err)
				}
			}
		}

		return nil
	})
}

func prefix(s string) []byte {
	return []byte(s)
}

func (s *BoltStore) RebuildFromMarkdown(rootPath string) error {
	openspecDir := filepath.Join(rootPath, "openspec", "changes")

	storedMeta, err := s.GetSyncMeta()
	if err != nil {
		storedMeta = &spec.SyncMeta{FileMTimes: make(map[string]int64)}
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

		meta := &spec.SyncMeta{
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

func (s *BoltStore) walkOpenspecDirWithMTimes(openspecDir string, storedMTimes map[string]int64) ([]*spec.ChangeMetadata, []*spec.Task, []*spec.DependencyInfo, map[string]int64, error) {
	var changes []*spec.ChangeMetadata
	var tasks []*spec.Task
	var deps []*spec.DependencyInfo
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

		tasksInfo, err := os.Stat(tasksPath)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("stat tasks %s: %w", changeID, err)
		}
		tasksMTime := tasksInfo.ModTime().Unix()
		fileMTimes[tasksPath] = tasksMTime

		fileUnchanged := false
		if storedMTimes != nil {
			if storedMTime, exists := storedMTimes[tasksPath]; exists && storedMTime == tasksMTime {
				fileUnchanged = true
			}
		}

		var changeTasks []*spec.Task
		var title string

		if fileUnchanged {
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

		var status spec.ChangeStatus
		switch {
		case completed == len(changeTasks) && len(changeTasks) > 0:
			status = spec.StatusApproved
		case completed > 0:
			status = spec.StatusActive
		default:
			status = spec.StatusDraft
		}

		change := &spec.ChangeMetadata{
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
	for i, line := range strings.Split(lines, "\n") {
		if i > 10 {
			break
		}
		if len(line) > 0 && line[0] == '#' {
			for len(line) > 0 && (line[0] == '#' || line[0] == ' ') {
				line = line[1:]
			}
			return strings.TrimSpace(line), nil
		}
	}

	return "", fmt.Errorf("no title found in proposal")
}

func (s *BoltStore) ParseTasksFile(path string, changeID string, fileMTimes *map[string]int64) ([]*spec.Task, error) {
	parserTasks, err := s.parser.ParseTasksFile(path, changeID)
	if err != nil {
		return nil, fmt.Errorf("parse tasks file: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat tasks: %w", err)
	}
	(*fileMTimes)[path] = info.ModTime().Unix()

	tasks := make([]*spec.Task, len(parserTasks))
	for i, pt := range parserTasks {
		tasks[i] = &spec.Task{
			ID:         pt.ID,
			ChangeID:   pt.ChangeID,
			TaskNum:    pt.TaskNum,
			Content:    pt.Content,
			Status:     specTaskStatus(pt.Status),
			Priority:   pt.Priority,
			DependsOn:  pt.DependsOn,
			SourceLine: pt.SourceLine,
			SyncedAt:   pt.SyncedAt,
		}
	}

	return tasks, nil
}

func specTaskStatus(s parser.TaskStatus) spec.TaskStatus {
	switch s {
	case parser.TaskPending:
		return spec.TaskPending
	case parser.TaskInProgress:
		return spec.TaskInProgress
	case parser.TaskCompleted:
		return spec.TaskCompleted
	default:
		return spec.TaskPending
	}
}

func (s *BoltStore) CountTaskStatuses(tasks []*spec.Task) (completed, inProgress, blocked int) {
	for _, task := range tasks {
		switch task.Status {
		case spec.TaskCompleted:
			completed++
		case spec.TaskInProgress:
			inProgress++
		}
	}
	return completed, inProgress, blocked
}

func (s *BoltStore) BuildDependencies(tasks []*spec.Task) []*spec.DependencyInfo {
	taskMap := make(map[string]*spec.Task)
	for _, task := range tasks {
		taskMap[task.TaskNum] = task
	}

	var deps []*spec.DependencyInfo
	for _, task := range tasks {
		dep := &spec.DependencyInfo{
			TaskID:    task.ID,
			DependsOn: task.DependsOn,
		}

		var blockingTasks []string
		for _, depNum := range task.DependsOn {
			if depTask, exists := taskMap[depNum]; exists {
				dep.DependedBy = append(dep.DependedBy, depTask.ID)
				if depTask.Status != spec.TaskCompleted {
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

func (s *BoltStore) ListAllChanges() ([]*spec.ChangeMetadata, error) {
	var changes []*spec.ChangeMetadata
	err := s.view(changesBucket, func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			var change spec.ChangeMetadata
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

func (s *BoltStore) ListTasks(changeID, status string) ([]*spec.Task, error) {
	var tasks []*spec.Task
	err := s.view(tasksBucket, func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			var task spec.Task
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

func (s *BoltStore) GetTasksByChange(changeID string) ([]*spec.Task, error) {
	return s.ListTasks(changeID, "")
}
