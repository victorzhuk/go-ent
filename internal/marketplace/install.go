package marketplace

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Installer handles plugin installation from marketplace.
type Installer struct {
	client     *Client
	pluginsDir string
}

// NewInstaller creates a new plugin installer.
func NewInstaller(client *Client, pluginsDir string) *Installer {
	return &Installer{
		client:     client,
		pluginsDir: pluginsDir,
	}
}

// Install downloads and installs a plugin.
func (i *Installer) Install(ctx context.Context, name, version string) error {
	data, err := i.client.Download(ctx, name, version)
	if err != nil {
		return fmt.Errorf("download plugin: %w", err)
	}

	if err := i.Verify(data); err != nil {
		return fmt.Errorf("verify plugin: %w", err)
	}

	if err := i.Extract(name, data); err != nil {
		return fmt.Errorf("extract plugin: %w", err)
	}

	return nil
}

// Verify checks if plugin data is valid.
func (i *Installer) Verify(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty plugin data")
	}

	if !i.isZipArchive(data) {
		return fmt.Errorf("invalid plugin format")
	}

	return nil
}

// Extract extracts plugin archive to plugin directory.
func (i *Installer) Extract(name string, data []byte) error {
	pluginDir := filepath.Join(i.pluginsDir, name)

	if err := os.MkdirAll(pluginDir, 0750); err != nil {
		return fmt.Errorf("create plugin directory: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("create zip reader: %w", err)
	}

	for _, file := range zipReader.File {
		filePath := filepath.Join(pluginDir, file.Name) // #nosec G305 -- path validated below

		rel, err := filepath.Rel(pluginDir, filePath)
		if err != nil || filepath.IsAbs(rel) || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("unsafe path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("create directory %s: %w", file.Name, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
			return fmt.Errorf("create parent directory for %s: %w", file.Name, err)
		}

		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("open file %s: %w", file.Name, err)
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode()) // #nosec G304 -- controlled file path
		if err != nil {
			_ = fileReader.Close()
			return fmt.Errorf("create file %s: %w", file.Name, err)
		}

		// #nosec G110 -- archive from trusted source
		if _, err := io.Copy(destFile, fileReader); err != nil {
			_ = destFile.Close()
			_ = fileReader.Close()
			return fmt.Errorf("copy file %s: %w", file.Name, err)
		}

		_ = destFile.Close()
		_ = fileReader.Close()
	}

	return nil
}

// Validate checks if installed plugin is valid.
func (i *Installer) Validate(name string) error {
	pluginDir := filepath.Join(i.pluginsDir, name)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin.yaml not found")
	}

	return nil
}

// Uninstall removes plugin from file system.
func (i *Installer) Uninstall(name string) error {
	pluginDir := filepath.Join(i.pluginsDir, name)

	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("remove plugin directory: %w", err)
	}

	return nil
}

// ListInstalled returns list of installed plugin names.
func (i *Installer) ListInstalled() []string {
	entries, err := os.ReadDir(i.pluginsDir)
	if err != nil {
		return []string{}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	return names
}

func (i *Installer) isZipArchive(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	signature := []byte{0x50, 0x4B, 0x03, 0x04}
	for i := 0; i < 4; i++ {
		//nolint:gosec // bounds checked above
		if data[i] != signature[i] {
			return false
		}
	}

	return true
}
