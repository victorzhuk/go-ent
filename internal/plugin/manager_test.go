package plugin

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRegistry struct{}

func (m *mockRegistry) RegisterSkill(name, path string) error {
	return nil
}

func (m *mockRegistry) RegisterAgent(name, path string) error {
	return nil
}

func (m *mockRegistry) UnregisterSkill(name string) error {
	return nil
}

func (m *mockRegistry) UnregisterAgent(name string) error {
	return nil
}

type mockMarketplace struct{}

func (m *mockMarketplace) Download(ctx context.Context, name, version string) ([]byte, error) {
	return nil, nil
}

func TestManager_Initialize_LogsParseFailures(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	invalidPluginDir := filepath.Join(tmpDir, "invalid-plugin")
	require.NoError(t, os.Mkdir(invalidPluginDir, 0750))

	invalidManifestPath := filepath.Join(invalidPluginDir, ManifestFile)
	require.NoError(t, os.WriteFile(invalidManifestPath, []byte("invalid yaml"), 0600))

	validPluginDir := filepath.Join(tmpDir, "valid-plugin")
	require.NoError(t, os.Mkdir(validPluginDir, 0750))

	validManifestPath := filepath.Join(validPluginDir, ManifestFile)
	validManifest := `name: valid-plugin
version: 1.0.0
description: A valid plugin
author: Test

skills:
  - name: test-skill
    path: skills/test.md
`
	require.NoError(t, os.WriteFile(validManifestPath, []byte(validManifest), 0600))

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	logger := slog.New(logHandler)

	m := NewManager(tmpDir, &mockRegistry{}, &mockMarketplace{}, logger)

	err := m.Initialize(context.Background())
	assert.NoError(t, err)

	plugins := m.List()
	assert.Len(t, plugins, 1)
	assert.Equal(t, "valid-plugin", plugins[0].Name)
}

func TestManager_NewManager_DefaultLogger(t *testing.T) {
	t.Parallel()

	m := NewManager("test-dir", &mockRegistry{}, &mockMarketplace{}, nil)

	assert.NotNil(t, m.logger)
	assert.Equal(t, slog.Default(), m.logger)
}

func TestManager_NewManager_CustomLogger(t *testing.T) {
	t.Parallel()

	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewManager("test-dir", &mockRegistry{}, &mockMarketplace{}, customLogger)

	assert.Equal(t, customLogger, m.logger)
}
