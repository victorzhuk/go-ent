package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ArchiveResult contains the result of an archive operation.
type ArchiveResult struct {
	ChangeID     string
	ArchivePath  string
	UpdatedSpecs []string
	DryRun       bool
	Errors       []string
}

// Archiver handles archiving completed changes.
type Archiver struct {
	store *Store
}

// NewArchiver creates a new archiver for the given store.
func NewArchiver(store *Store) *Archiver {
	return &Archiver{store: store}
}

// Archive archives a change and optionally merges deltas into specs.
func (a *Archiver) Archive(changeID string, skipSpecs bool, dryRun bool) (*ArchiveResult, error) {
	result := &ArchiveResult{
		ChangeID: changeID,
		DryRun:   dryRun,
	}

	changePath := filepath.Join(a.store.SpecPath(), "changes", changeID)

	// Verify change exists
	if _, err := os.Stat(changePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("change not found: %s", changeID)
	}

	// Generate archive path with date prefix
	today := time.Now().Format("2006-01-02")
	archiveName := fmt.Sprintf("%s-%s", today, changeID)
	archivePath := filepath.Join(a.store.SpecPath(), "changes", "archive", archiveName)
	result.ArchivePath = archivePath

	// Merge deltas into specs if not skipped
	if !skipSpecs {
		updatedSpecs, err := a.mergeDeltas(changePath, dryRun)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("merge deltas: %v", err))
			return result, nil
		}
		result.UpdatedSpecs = updatedSpecs
	}

	// Move change to archive
	if !dryRun {
		// Ensure archive directory exists
		if err := os.MkdirAll(filepath.Dir(archivePath), 0750); err != nil {
			return nil, fmt.Errorf("create archive dir: %w", err)
		}

		// Move the change directory
		if err := os.Rename(changePath, archivePath); err != nil {
			return nil, fmt.Errorf("move to archive: %w", err)
		}
	}

	return result, nil
}

// mergeDeltas merges all delta specs from change into main specs.
func (a *Archiver) mergeDeltas(changePath string, dryRun bool) ([]string, error) {
	var updatedSpecs []string

	deltaSpecsPath := filepath.Join(changePath, "specs")

	// Check if specs directory exists
	if _, err := os.Stat(deltaSpecsPath); os.IsNotExist(err) {
		return nil, nil // No deltas to merge
	}

	// Walk through delta specs
	err := filepath.Walk(deltaSpecsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Get capability name from path
		relPath, err := filepath.Rel(deltaSpecsPath, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}

		capabilityDir := filepath.Dir(relPath)
		if capabilityDir == "." {
			capabilityDir = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		// Read delta spec
		deltaContent, err := os.ReadFile(path) // #nosec G304 -- controlled file path
		if err != nil {
			return fmt.Errorf("read delta spec: %w", err)
		}

		// Parse delta operations
		delta, err := ParseDeltaSpec(string(deltaContent))
		if err != nil {
			return fmt.Errorf("parse delta spec: %w", err)
		}

		// Read base spec
		baseSpecPath := filepath.Join(a.store.SpecPath(), "specs", capabilityDir, "spec.md")
		var baseContent string

		if _, err := os.Stat(baseSpecPath); os.IsNotExist(err) {
			baseContent = fmt.Sprintf("# %s Specification\n\n", capitalizeFirst(capabilityDir))
		} else {
			content, err := os.ReadFile(baseSpecPath) // #nosec G304 -- controlled file path
			if err != nil {
				return fmt.Errorf("read base spec: %w", err)
			}
			baseContent = string(content)
		}

		// Merge deltas
		merged, err := MergeDeltas(baseContent, delta)
		if err != nil {
			return fmt.Errorf("merge deltas: %w", err)
		}

		// Write updated spec
		if !dryRun {
			if err := os.MkdirAll(filepath.Dir(baseSpecPath), 0750); err != nil {
				return fmt.Errorf("create spec dir: %w", err)
			}
			if err := os.WriteFile(baseSpecPath, []byte(merged), 0600); err != nil {
				return fmt.Errorf("write spec: %w", err)
			}
		}

		updatedSpecs = append(updatedSpecs, capabilityDir)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedSpecs, nil
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 32
	}
	return string(r)
}

// ValidateBeforeArchive validates a change before archiving.
func (a *Archiver) ValidateBeforeArchive(changeID string, strict bool) (*ValidationResult, error) {
	changePath := filepath.Join(a.store.SpecPath(), "changes", changeID)
	validator := NewValidator()
	return validator.ValidateChange(changePath, strict)
}
