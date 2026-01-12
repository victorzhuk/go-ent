package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

const (
	BucketTasks    = "tasks"
	BucketChanges  = "changes"
	BucketDeps     = "deps"
	BucketBlocking = "blocking"
	BucketMeta     = "meta"
)

type BoltStore struct {
	path string
	db   *bbolt.DB
}

func NewBoltStore(path string) (*BoltStore, error) {
	if path == "" {
		return nil, errors.New("bolt store path cannot be empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	store := &BoltStore{path: path, db: db}
	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return store, nil
}

func (s *BoltStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *BoltStore) initBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{BucketTasks, BucketChanges, BucketDeps, BucketBlocking, BucketMeta}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
}

func (s *BoltStore) GetTask(id TaskID) (*RegistryTask, error) {
	var task RegistryTask
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketTasks))
		data := b.Get([]byte(id.String()))
		if data == nil {
			return fmt.Errorf("task not found: %s", id)
		}
		return json.Unmarshal(data, &task)
	})
	return &task, err
}

func (s *BoltStore) UpdateTask(task *RegistryTask) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketTasks))
		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("marshal task: %w", err)
		}
		if err := b.Put([]byte(task.ID.String()), data); err != nil {
			return fmt.Errorf("put task: %w", err)
		}

		return s.updateChangeSummary(tx, task.ID.ChangeID)
	})
}

func (s *BoltStore) ListTasks(filter TaskFilter) ([]RegistryTask, error) {
	var tasks []RegistryTask
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketTasks))
		return b.ForEach(func(k, v []byte) error {
			var task RegistryTask
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}
			if matches(&filter, &task) {
				tasks = append(tasks, task)
			}
			return nil
		})
	})
	return tasks, err
}

func (s *BoltStore) NextTasks(count int) ([]RegistryTask, error) {
	if count <= 0 {
		count = 1
	}

	tasks, err := s.ListTasks(TaskFilter{
		Status:    RegStatusPending,
		Unblocked: true,
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return priorityValue(tasks[i].Priority) < priorityValue(tasks[j].Priority)
		}
		return tasks[i].ID.String() < tasks[j].ID.String()
	})

	if len(tasks) > count {
		tasks = tasks[:count]
	}
	return tasks, nil
}

func (s *BoltStore) AddDependency(from, to TaskID) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		deps := tx.Bucket([]byte(BucketDeps))
		blocking := tx.Bucket([]byte(BucketBlocking))

		depsList, _ := s.getDeps(tx, from)
		for _, existing := range depsList {
			if existing.String() == to.String() {
				return nil
			}
		}
		depsList = append(depsList, to)

		data, err := json.Marshal(depsList)
		if err != nil {
			return err
		}
		if err := deps.Put([]byte(from.String()), data); err != nil {
			return err
		}

		blockingList, _ := s.getBlocking(tx, to)
		blockingList = append(blockingList, from)
		data, err = json.Marshal(blockingList)
		if err != nil {
			return err
		}
		return blocking.Put([]byte(to.String()), data)
	})
}

func (s *BoltStore) RemoveDependency(from, to TaskID) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		deps := tx.Bucket([]byte(BucketDeps))
		blocking := tx.Bucket([]byte(BucketBlocking))

		depsList, _ := s.getDeps(tx, from)
		filtered := make([]TaskID, 0, len(depsList))
		for _, dep := range depsList {
			if dep.String() != to.String() {
				filtered = append(filtered, dep)
			}
		}
		data, err := json.Marshal(filtered)
		if err != nil {
			return err
		}
		if err := deps.Put([]byte(from.String()), data); err != nil {
			return err
		}

		blockingList, _ := s.getBlocking(tx, to)
		filtered = make([]TaskID, 0, len(blockingList))
		for _, blocker := range blockingList {
			if blocker.String() != from.String() {
				filtered = append(filtered, blocker)
			}
		}
		data, err = json.Marshal(filtered)
		if err != nil {
			return err
		}
		return blocking.Put([]byte(to.String()), data)
	})
}

func (s *BoltStore) GetBlockers(id TaskID) ([]TaskID, error) {
	var blockers []TaskID
	err := s.db.View(func(tx *bbolt.Tx) error {
		depsList, err := s.getDeps(tx, id)
		if err != nil {
			return err
		}

		tasks := tx.Bucket([]byte(BucketTasks))
		for _, dep := range depsList {
			data := tasks.Get([]byte(dep.String()))
			if data != nil {
				var task RegistryTask
				if err := json.Unmarshal(data, &task); err == nil {
					if task.Status != RegStatusCompleted {
						blockers = append(blockers, dep)
					}
				}
			}
		}
		return nil
	})
	return blockers, err
}

func (s *BoltStore) GetChangeSummary(changeID string) (*ChangeSummary, error) {
	var summary ChangeSummary
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketChanges))
		data := b.Get([]byte(changeID))
		if data == nil {
			return fmt.Errorf("change not found: %s", changeID)
		}
		return json.Unmarshal(data, &summary)
	})
	return &summary, err
}

func (s *BoltStore) ListChanges() ([]ChangeSummary, error) {
	var changes []ChangeSummary
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketChanges))
		return b.ForEach(func(k, v []byte) error {
			var summary ChangeSummary
			if err := json.Unmarshal(v, &summary); err != nil {
				return err
			}
			changes = append(changes, summary)
			return nil
		})
	})
	return changes, err
}

func (s *BoltStore) SetSyncedAt(t time.Time) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketMeta))
		data, _ := json.Marshal(t)
		return b.Put([]byte("synced_at"), data)
	})
}

func (s *BoltStore) GetSyncedAt() (time.Time, error) {
	var t time.Time
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketMeta))
		data := b.Get([]byte("synced_at"))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &t)
	})
	return t, err
}

func (s *BoltStore) getDeps(tx *bbolt.Tx, id TaskID) ([]TaskID, error) {
	b := tx.Bucket([]byte(BucketDeps))
	data := b.Get([]byte(id.String()))
	if data == nil {
		return nil, nil
	}
	var deps []TaskID
	err := json.Unmarshal(data, &deps)
	return deps, err
}

func (s *BoltStore) getBlocking(tx *bbolt.Tx, id TaskID) ([]TaskID, error) {
	b := tx.Bucket([]byte(BucketBlocking))
	data := b.Get([]byte(id.String()))
	if data == nil {
		return nil, nil
	}
	var blocking []TaskID
	err := json.Unmarshal(data, &blocking)
	return blocking, err
}

func (s *BoltStore) updateChangeSummary(tx *bbolt.Tx, changeID string) error {
	tasks := tx.Bucket([]byte(BucketTasks))
	changes := tx.Bucket([]byte(BucketChanges))

	var total, completed, inProgress, blocked int
	err := tasks.ForEach(func(k, v []byte) error {
		var task RegistryTask
		if err := json.Unmarshal(v, &task); err != nil {
			return err
		}
		if task.ID.ChangeID == changeID {
			total++
			switch task.Status {
			case RegStatusCompleted:
				completed++
			case RegStatusInProgress:
				inProgress++
			case RegStatusBlocked:
				blocked++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	summary := ChangeSummary{
		ID:         changeID,
		Total:      total,
		Completed:  completed,
		InProgress: inProgress,
		Blocked:    blocked,
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	return changes.Put([]byte(changeID), data)
}

// UpdateChange updates or creates a change summary (public version)
// UpdateChange updates or creates a change summary (public version)
func (s *BoltStore) UpdateChange(summary ChangeSummary) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return s.updateChangeSummary(tx, summary.ID)
	})
}

// ClearTasks removes all tasks from the database (used during full rebuild)
func (s *BoltStore) ClearTasks() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		// Drop and recreate buckets
		if err := tx.DeleteBucket([]byte(BucketTasks)); err != nil && err != bbolt.ErrBucketNotFound {
			return fmt.Errorf("delete tasks bucket: %w", err)
		}
		if err := tx.DeleteBucket([]byte(BucketDeps)); err != nil && err != bbolt.ErrBucketNotFound {
			return fmt.Errorf("delete deps bucket: %w", err)
		}
		if err := tx.DeleteBucket([]byte(BucketBlocking)); err != nil && err != bbolt.ErrBucketNotFound {
			return fmt.Errorf("delete blocking bucket: %w", err)
		}

		// Recreate buckets
		if _, err := tx.CreateBucket([]byte(BucketTasks)); err != nil {
			return fmt.Errorf("create tasks bucket: %w", err)
		}
		if _, err := tx.CreateBucket([]byte(BucketDeps)); err != nil {
			return fmt.Errorf("create deps bucket: %w", err)
		}
		if _, err := tx.CreateBucket([]byte(BucketBlocking)); err != nil {
			return fmt.Errorf("create blocking bucket: %w", err)
		}

		return nil
	})
}

// SetMeta sets a metadata value (generic version of SetSyncedAt)
func (s *BoltStore) SetMeta(key, value string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketMeta))
		if b == nil {
			return fmt.Errorf("meta bucket not found")
		}
		return b.Put([]byte(key), []byte(value))
	})
}

// GetMeta retrieves a metadata value
func (s *BoltStore) GetMeta(key string) (string, error) {
	var value string
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketMeta))
		if b == nil {
			return fmt.Errorf("meta bucket not found")
		}
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key not found")
		}
		value = string(data)
		return nil
	})
	return value, err
}

func matches(f *TaskFilter, task *RegistryTask) bool {
	if f.ChangeID != "" && task.ID.ChangeID != f.ChangeID {
		return false
	}
	if f.Status != "" && task.Status != f.Status {
		return false
	}
	if f.Priority != "" && task.Priority != f.Priority {
		return false
	}
	if f.Assignee != "" && task.Assignee != f.Assignee {
		return false
	}
	if f.Unblocked && len(task.BlockedBy) > 0 {
		return false
	}
	if f.Limit > 0 && len(task.BlockedBy) == 0 {
		return true
	}
	return true
}
