package workspace

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const workspaceSpecPrefix = "ws-"

func SyncProject(ws *Workspace, project ProjectRef) (specs, changes, pulled int, err error) {
	db, dbErr := OpenDB(ws.Name)
	if dbErr != nil {
		slog.Warn("workspace db unavailable, syncing without index", "error", dbErr)
	}
	if db != nil {
		defer db.Close()
	}

	specs, err = indexProjectSpecs(ws, project, db)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("index specs: %w", err)
	}

	changes, err = indexProjectChanges(ws, project, db)
	if err != nil {
		return specs, 0, 0, fmt.Errorf("index changes: %w", err)
	}

	if db != nil {
		if err := db.PutProject(&ProjectMeta{
			Name:        project.Name,
			Path:        project.Path,
			SyncedAt:    time.Now(),
			SpecCount:   specs,
			ChangeCount: changes,
		}); err != nil {
			slog.Warn("failed to index project", "project", project.Name, "error", err)
		}
	}

	pulled, err = pullWorkspaceSpecs(ws, project)
	if err != nil {
		return specs, changes, 0, fmt.Errorf("pull workspace specs: %w", err)
	}

	return specs, changes, pulled, nil
}

func indexProjectSpecs(ws *Workspace, project ProjectRef, db *WorkspaceDB) (int, error) {
	specsDir := filepath.Join(project.Path, "openspec", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read specs dir: %w", err)
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		specID := e.Name()
		title := ExtractTitle(filepath.Join(specsDir, specID, "spec.md"))

		if db != nil {
			if err := db.PutSpec(&SpecMeta{
				ID:      specID,
				Project: project.Name,
				Title:   title,
				Path:    filepath.Join(specsDir, specID),
			}); err != nil {
				slog.Warn("failed to index spec", "spec", specID, "error", err)
			}
		}

		count++
	}

	return count, nil
}

func indexProjectChanges(ws *Workspace, project ProjectRef, db *WorkspaceDB) (int, error) {
	changesDir := filepath.Join(project.Path, "openspec", "changes")
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read changes dir: %w", err)
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "archive" {
			continue
		}

		changeID := e.Name()
		title := ExtractTitle(filepath.Join(changesDir, changeID, "proposal.md"))

		if db != nil {
			if err := db.PutChange(&ChangeMeta{
				ID:      changeID,
				Project: project.Name,
				Title:   title,
				Status:  "active",
			}); err != nil {
				slog.Warn("failed to index change", "change", changeID, "error", err)
			}
		}

		count++
	}

	return count, nil
}

func pullWorkspaceSpecs(ws *Workspace, project ProjectRef) (int, error) {
	wsSpecsDir := filepath.Join(ws.Path, "openspec", "specs")
	entries, err := os.ReadDir(wsSpecsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read workspace specs: %w", err)
	}

	projectSpecsDir := filepath.Join(project.Path, "openspec", "specs")
	if err := os.MkdirAll(projectSpecsDir, 0o750); err != nil {
		return 0, fmt.Errorf("create project specs dir: %w", err)
	}

	pulled := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		specID := e.Name()
		wsSpecDir := filepath.Join(wsSpecsDir, specID)
		targetName := workspaceSpecPrefix + specID
		if strings.HasPrefix(specID, workspaceSpecPrefix) {
			targetName = specID
		}

		targetDir := filepath.Join(projectSpecsDir, targetName)
		if err := copySpecDir(wsSpecDir, targetDir); err != nil {
			return pulled, fmt.Errorf("copy spec %s: %w", specID, err)
		}
		pulled++
	}

	return pulled, nil
}

func copySpecDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		data, err := os.ReadFile(srcPath) // #nosec G304
		if err != nil {
			return fmt.Errorf("read %s: %w", e.Name(), err)
		}

		if err := os.WriteFile(dstPath, data, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", e.Name(), err)
		}
	}

	return nil
}

func ExtractTitle(path string) string {
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	return ""
}
