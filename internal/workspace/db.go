package workspace

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

const (
	projectsBucket       = "projects"
	projectSpecsBucket   = "project-specs"
	projectChangesBucket = "project-changes"
	dbOpenTimeout        = 1 * time.Second
)

type ProjectMeta struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	SyncedAt    time.Time `json:"synced_at"`
	SpecCount   int       `json:"spec_count"`
	ChangeCount int       `json:"change_count"`
}

type SpecMeta struct {
	ID      string `json:"id"`
	Project string `json:"project"`
	Title   string `json:"title"`
	Path    string `json:"path"`
}

type ChangeMeta struct {
	ID      string `json:"id"`
	Project string `json:"project"`
	Title   string `json:"title"`
	Status  string `json:"status"`
}

type WorkspaceDB struct {
	db   *bbolt.DB
	path string
}

func OpenDB(name string) (*WorkspaceDB, error) {
	dbPath := WorkspaceDBPath(name)

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{Timeout: dbOpenTimeout})
	if err != nil {
		return nil, fmt.Errorf("open workspace db: %w", err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range []string{projectsBucket, projectSpecsBucket, projectChangesBucket} {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return &WorkspaceDB{db: db, path: dbPath}, nil
}

func (w *WorkspaceDB) Close() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}

func (w *WorkspaceDB) PutProject(meta *ProjectMeta) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectsBucket))
		data, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshal project: %w", err)
		}
		return b.Put([]byte(meta.Name), data)
	})
}

func (w *WorkspaceDB) PutSpec(meta *SpecMeta) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectSpecsBucket))
		key := meta.Project + ":" + meta.ID
		data, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshal spec: %w", err)
		}
		return b.Put([]byte(key), data)
	})
}

func (w *WorkspaceDB) PutChange(meta *ChangeMeta) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectChangesBucket))
		key := meta.Project + ":" + meta.ID
		data, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshal change: %w", err)
		}
		return b.Put([]byte(key), data)
	})
}

func (w *WorkspaceDB) ListProjects() ([]ProjectMeta, error) {
	var result []ProjectMeta
	err := w.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectsBucket))
		return b.ForEach(func(k, v []byte) error {
			var meta ProjectMeta
			if err := json.Unmarshal(v, &meta); err != nil {
				slog.Warn("corrupt project entry, skipping", "key", string(k), "error", err)
				return nil
			}
			result = append(result, meta)
			return nil
		})
	})
	return result, err
}

func (w *WorkspaceDB) ListSpecs(project string) ([]SpecMeta, error) {
	var result []SpecMeta
	prefix := project + ":"
	err := w.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectSpecsBucket))
		c := b.Cursor()
		for k, v := c.Seek([]byte(prefix)); k != nil && len(k) >= len(prefix) && string(k[:len(prefix)]) == prefix; k, v = c.Next() {
			var meta SpecMeta
			if err := json.Unmarshal(v, &meta); err != nil {
				slog.Warn("corrupt spec entry, skipping", "key", string(k), "error", err)
				continue
			}
			result = append(result, meta)
		}
		return nil
	})
	return result, err
}
