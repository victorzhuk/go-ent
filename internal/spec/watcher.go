package spec

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.etcd.io/bbolt"
)

const (
	defaultDebounce = 100 * time.Millisecond
)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	debounce  time.Duration
	boltStore *BoltStore
	mu        sync.Mutex
	pending   map[string]bool
	timers    map[string]*time.Timer
	ctx       context.Context
	cancel    context.CancelFunc
	rootPath  string
}

func NewWatcher(boltStore *BoltStore, debounce time.Duration) (*Watcher, error) {
	if debounce == 0 {
		debounce = defaultDebounce
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		fsWatcher: fsWatcher,
		debounce:  debounce,
		boltStore: boltStore,
		pending:   make(map[string]bool),
		timers:    make(map[string]*time.Timer),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (w *Watcher) Start(rootPath string) error {
	w.rootPath = rootPath

	openspecDir := filepath.Join(rootPath, "openspec", "changes")

	if err := filepath.Walk(openspecDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("walk path %s: %w", path, err)
		}

		if info.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			if err := w.fsWatcher.Add(path); err != nil {
				slog.Warn("failed to watch directory", "path", path, "error", err)
			}
			return nil
		}

		if strings.HasSuffix(path, ".md") {
			if err := w.fsWatcher.Add(path); err != nil {
				slog.Warn("failed to watch file", "path", path, "error", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk openspec dir: %w", err)
	}

	go w.eventLoop()

	slog.Info("started file watcher", "path", openspecDir, "debounce", w.debounce)
	return nil
}

func (w *Watcher) Stop() error {
	w.cancel()

	if w.fsWatcher != nil {
		if err := w.fsWatcher.Close(); err != nil {
			slog.Warn("error closing file watcher", "error", err)
		}
	}

	w.mu.Lock()
	for path, timer := range w.timers {
		timer.Stop()
		delete(w.timers, path)
	}
	w.mu.Unlock()

	slog.Info("stopped file watcher")
	return nil
}

func (w *Watcher) eventLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "error", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	if strings.HasPrefix(filepath.Base(event.Name), ".") {
		return
	}

	if !strings.HasSuffix(event.Name, ".md") {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
		if existingTimer, exists := w.timers[event.Name]; exists {
			existingTimer.Stop()
		}

		w.pending[event.Name] = true
		w.timers[event.Name] = time.AfterFunc(w.debounce, func() {
			w.mu.Lock()
			delete(w.timers, event.Name)
			delete(w.pending, event.Name)
			w.mu.Unlock()

			if err := w.syncFile(event.Name); err != nil {
				slog.Error("sync file error", "path", event.Name, "error", err)
			}
		})
	}

	if event.Op&fsnotify.Remove == fsnotify.Remove {
		if existingTimer, exists := w.timers[event.Name]; exists {
			existingTimer.Stop()
			delete(w.timers, event.Name)
		}

		if pending, exists := w.pending[event.Name]; exists && pending {
			delete(w.pending, event.Name)
			return
		}

		changeID := w.extractChangeID(event.Name)
		if changeID != "" {
			if err := w.syncChange(changeID); err != nil {
				slog.Error("sync change after removal", "change", changeID, "error", err)
			}
		}
	}

	if event.Op&fsnotify.Rename == fsnotify.Rename {
		if existingTimer, exists := w.timers[event.Name]; exists {
			existingTimer.Stop()
			delete(w.timers, event.Name)
		}

		changeID := w.extractChangeID(event.Name)
		if changeID != "" {
			if err := w.syncChange(changeID); err != nil {
				slog.Error("sync change after rename", "change", changeID, "error", err)
			}
		}
	}
}

func (w *Watcher) syncFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		changeID := w.extractChangeID(path)
		if changeID != "" {
			return w.syncChange(changeID)
		}
		return nil
	}

	changeID := w.extractChangeID(path)
	if changeID == "" {
		return fmt.Errorf("cannot extract change ID from path: %s", path)
	}

	slog.Debug("syncing file", "path", path, "change", changeID)

	return w.syncChange(changeID)
}

func (w *Watcher) syncChange(changeID string) error {
	changeDir := filepath.Join(w.rootPath, "openspec", "changes", changeID)

	if _, err := os.Stat(changeDir); os.IsNotExist(err) {
		if err := w.deleteChangeFromDB(changeID); err != nil {
			return fmt.Errorf("delete change %s: %w", changeID, err)
		}
		slog.Info("removed change from database", "change", changeID)
		return nil
	}

	tasksPath := filepath.Join(changeDir, "tasks.md")
	proposalPath := filepath.Join(changeDir, "proposal.md")

	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		slog.Debug("no tasks file for change", "change", changeID)
		return nil
	}

	title, err := w.boltStore.ParseProposalTitle(proposalPath)
	if err != nil {
		slog.Warn("failed to parse proposal title", "change", changeID, "error", err)
		title = changeID
	}

	fileMTimes := make(map[string]int64)
	tasks, err := w.boltStore.ParseTasksFile(tasksPath, changeID, &fileMTimes)
	if err != nil {
		return fmt.Errorf("parse tasks %s: %w", changeID, err)
	}

	completed, inProgress, blocked := w.boltStore.CountTaskStatuses(tasks)

	var status ChangeStatus
	switch {
	case completed == len(tasks) && len(tasks) > 0:
		status = StatusApproved
	case completed > 0:
		status = StatusActive
	default:
		status = StatusDraft
	}

	change := &ChangeMetadata{
		ID:         changeID,
		Title:      title,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TaskCount:  len(tasks),
		Completed:  completed,
		InProgress: inProgress,
		Blocked:    blocked,
	}

	if err := w.boltStore.PutChange(change); err != nil {
		return fmt.Errorf("put change %s: %w", changeID, err)
	}

	for _, task := range tasks {
		task.ID = taskKey(changeID, task.TaskNum)
		task.SyncedAt = time.Now()

		if err := w.boltStore.PutTask(task); err != nil {
			slog.Warn("failed to put task", "task", task.ID, "error", err)
		}
	}

	deps := w.boltStore.BuildDependencies(tasks)
	for _, dep := range deps {
		if err := w.boltStore.PutDeps(dep); err != nil {
			slog.Warn("failed to put dependencies", "task", dep.TaskID, "error", err)
		}
	}

	slog.Debug("synced change", "change", changeID, "tasks", len(tasks))
	return nil
}

func (w *Watcher) deleteChangeFromDB(changeID string) error {
	if err := w.boltStore.db.Update(func(tx *bbolt.Tx) error {
		changeBucket := tx.Bucket([]byte(changesBucket))
		if changeBucket != nil {
			if err := changeBucket.Delete([]byte(changeID)); err != nil {
				return fmt.Errorf("delete change %s: %w", changeID, err)
			}
		}

		tasksBucket := tx.Bucket([]byte(tasksBucket))
		if tasksBucket != nil {
			c := tasksBucket.Cursor()
			prefix := []byte(changeID + ":")
			for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
				if err := c.Delete(); err != nil {
					return fmt.Errorf("delete task %s: %w", string(k), err)
				}
			}
		}

		depsBucket := tx.Bucket([]byte(depsBucket))
		if depsBucket != nil {
			taskPrefix := changeID + ":"
			c := depsBucket.Cursor()
			for k, _ := c.Seek([]byte(taskPrefix)); k != nil && strings.HasPrefix(string(k), taskPrefix); k, _ = c.Next() {
				if err := c.Delete(); err != nil {
					return fmt.Errorf("delete deps %s: %w", string(k), err)
				}
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("delete change from db %s: %w", changeID, err)
	}

	return nil
}

func (w *Watcher) extractChangeID(path string) string {
	path = filepath.Clean(path)
	parts := strings.Split(path, string(filepath.Separator))

	for i, part := range parts {
		if part == "changes" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

func (w *Watcher) Recover(rootPath string) error {
	slog.Warn("attempting to recover from corrupted state, rebuilding database")

	if err := w.boltStore.RebuildFromMarkdown(rootPath); err != nil {
		return fmt.Errorf("rebuild from markdown: %w", err)
	}

	slog.Info("successfully recovered from corrupted state")
	return nil
}
