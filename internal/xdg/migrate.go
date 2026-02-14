package xdg

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func MigrateIfNeeded() error {
	legacy := LegacyDir()
	if _, err := os.Stat(legacy); os.IsNotExist(err) {
		return nil
	}

	if err := migrateFile(
		filepath.Join(legacy, "models.yaml"),
		filepath.Join(ConfigDir(), "models.yaml"),
	); err != nil {
		return fmt.Errorf("migrate models.yaml: %w", err)
	}

	if err := migrateDir(
		filepath.Join(legacy, "templates", "skills"),
		filepath.Join(DataDir(), "templates", "skills"),
	); err != nil {
		return fmt.Errorf("migrate templates: %w", err)
	}

	return nil
}

func migrateFile(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	data, err := os.ReadFile(src) // #nosec G304
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return fmt.Errorf("create dir for %s: %w", dst, err)
	}

	if err := os.WriteFile(dst, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}

	slog.Warn("migrated from legacy path",
		"from", src,
		"to", dst,
		"hint", "consider removing ~/.go-ent/ after verifying migration",
	)

	return nil
}

func migrateDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", src, err)
	}

	if len(entries) == 0 {
		return nil
	}

	if err := os.MkdirAll(dst, 0o750); err != nil {
		return fmt.Errorf("create dir %s: %w", dst, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := migrateDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := migrateFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	slog.Warn("migrated directory from legacy path",
		"from", src,
		"to", dst,
	)

	return nil
}
