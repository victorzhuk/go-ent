package marketplace

//nolint:gosec // test file with necessary file operations

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstaller_CreatesStruct(t *testing.T) {
	t.Parallel()

	mockClient := &Client{}
	tempDir := t.TempDir()

	installer := NewInstaller(mockClient, tempDir)

	assert.NotNil(t, installer)
	assert.Equal(t, mockClient, installer.client)
	assert.Equal(t, tempDir, installer.pluginsDir)
}

func TestInstallerVerify_EmptyData(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	err := installer.Verify([]byte{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty plugin data")
}

func TestInstallerVerify_InvalidZip(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	err := installer.Verify([]byte{0x00, 0x00, 0x00, 0x00})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plugin format")
}

func TestInstallerVerify_ValidZip(t *testing.T) {
	t.Parallel()

	zipData := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	err := installer.Verify(zipData)

	assert.NoError(t, err)
}

func TestInstallerExtract_CreatesDirectories(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	_, err := zipWriter.Create("skills/")
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)
	require.NoError(t, err)

	pluginDir := filepath.Join(tempDir, "test-plugin")
	info, err := os.Stat(pluginDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	skillsDir := filepath.Join(pluginDir, "skills")
	info, err = os.Stat(skillsDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInstallerExtract_ExtractsFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	fileContent := []byte("test file content")
	writer, err := zipWriter.Create("file.txt")
	require.NoError(t, err)
	_, err = writer.Write(fileContent)
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)
	require.NoError(t, err)

	filePath := filepath.Join(tempDir, "test-plugin", "file.txt")
	//nolint:G304 // controlled test path
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, fileContent, content)
}

func TestInstallerExtract_UnsafePath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	writer, err := zipWriter.Create("../../test.txt")
	require.NoError(t, err)
	_, _ = writer.Write([]byte("test"))
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe path")
}

func TestInstallerExtract_AbsolutePath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	writer, err := zipWriter.Create("../../etc/passwd")
	require.NoError(t, err)
	_, _ = writer.Write([]byte("test"))

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe path")
}

func TestInstallerExtract_WithManifest(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	manifestContent := []byte("name: test-plugin\nversion: 1.0.0\n")
	writer, err := zipWriter.Create("plugin.yaml")
	require.NoError(t, err)
	_, err = writer.Write(manifestContent)
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)
	require.NoError(t, err)

	manifestPath := filepath.Join(tempDir, "test-plugin", "plugin.yaml")
	//nolint:G304 // controlled test path
	content, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, manifestContent, content)
}

func TestInstallerExtract_NestedDirectories(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	fileContent := []byte("nested file content")
	writer, err := zipWriter.Create("deep/nested/path/file.txt")
	require.NoError(t, err)
	_, err = writer.Write(fileContent)
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)
	require.NoError(t, err)

	filePath := filepath.Join(tempDir, "test-plugin", "deep", "nested", "path", "file.txt")
	//nolint:G304 // controlled test path
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, fileContent, content)
}

func TestInstallerValidate_ManifestExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	pluginDir := filepath.Join(tempDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	require.NoError(t, os.WriteFile(manifestPath, []byte("name: test\n"), 0600))

	err := installer.Validate("test-plugin")

	assert.NoError(t, err)
}

func TestInstallerValidate_ManifestMissing(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	pluginDir := filepath.Join(tempDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	err := installer.Validate("test-plugin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin.yaml not found")
}

func TestInstallerValidate_PluginNotExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	err := installer.Validate("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin.yaml not found")
}

func TestInstallerUninstall_RemovesDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	pluginDir := filepath.Join(tempDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	require.NoError(t, os.WriteFile(manifestPath, []byte("name: test\n"), 0600))

	err := installer.Uninstall("test-plugin")
	require.NoError(t, err)

	_, err = os.Stat(pluginDir)
	assert.True(t, os.IsNotExist(err))
}

func TestInstallerUninstall_NonExistent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	err := installer.Uninstall("nonexistent")

	assert.NoError(t, err)
}

func TestInstallerListInstalled_ReturnsNames(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	plugin1Dir := filepath.Join(tempDir, "plugin1")
	require.NoError(t, os.Mkdir(plugin1Dir, 0750))

	plugin2Dir := filepath.Join(tempDir, "plugin2")
	require.NoError(t, os.Mkdir(plugin2Dir, 0750))

	filePath := filepath.Join(tempDir, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test"), 0600))

	plugins := installer.ListInstalled()

	require.Len(t, plugins, 2)
	assert.Contains(t, plugins, "plugin1")
	assert.Contains(t, plugins, "plugin2")
}

func TestInstallerListInstalled_Empty(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	plugins := installer.ListInstalled()

	assert.Empty(t, plugins)
}

func TestInstallerListInstalled_WithFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	filePath := filepath.Join(tempDir, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test"), 0600))

	plugins := installer.ListInstalled()

	assert.Empty(t, plugins)
}

func TestIsZipArchive_ValidSignature(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	result := installer.isZipArchive([]byte{0x50, 0x4B, 0x03, 0x04})

	assert.True(t, result)
}

func TestIsZipArchive_InvalidSignature(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	result := installer.isZipArchive([]byte{0x00, 0x00, 0x00, 0x00})

	assert.False(t, result)
}

func TestIsZipArchive_TooShort(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	result := installer.isZipArchive([]byte{0x50, 0x4B})

	assert.False(t, result)
}

func TestIsZipArchive_Empty(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	result := installer.isZipArchive([]byte{})

	assert.False(t, result)
}

func TestInstallerInstall_Success(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	manifestContent := []byte("name: test-plugin\nversion: 1.0.0\n")
	writer, err := zipWriter.Create("plugin.yaml")
	require.NoError(t, err)
	_, err = writer.Write(manifestContent)
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())
	zipData := buf.Bytes()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plugins/test-plugin/versions/1.0.0/download", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipData)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	client := NewClientWithURL(server.URL)
	installer := NewInstaller(client, tempDir)

	err = installer.Install(context.Background(), "test-plugin", "1.0.0")

	require.NoError(t, err)

	manifestPath := filepath.Join(tempDir, "test-plugin", "plugin.yaml")
	_, err = os.Stat(manifestPath)
	assert.NoError(t, err)
}

func TestInstallerInstall_DownloadError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("plugin not found"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	client := NewClientWithURL(server.URL)
	installer := NewInstaller(client, tempDir)

	err := installer.Install(context.Background(), "test-plugin", "1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download plugin")
}

func TestInstallerInstall_VerifyError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{0x00, 0x00, 0x00, 0x00})
	}))
	defer server.Close()

	tempDir := t.TempDir()
	client := NewClientWithURL(server.URL)
	installer := NewInstaller(client, tempDir)

	err := installer.Install(context.Background(), "test-plugin", "1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verify plugin")
}

func TestInstallerExtract_Overwrite(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	pluginDir := filepath.Join(tempDir, "test-plugin")
	require.NoError(t, os.Mkdir(pluginDir, 0750))

	existingFile := filepath.Join(pluginDir, "existing.txt")
	require.NoError(t, os.WriteFile(existingFile, []byte("old content"), 0600))

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	newContent := []byte("new content")
	writer, err := zipWriter.Create("existing.txt")
	require.NoError(t, err)
	_, err = writer.Write(newContent)
	require.NoError(t, err)

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err = installer.Extract("test-plugin", zipData)
	require.NoError(t, err)

	//nolint:G304 // controlled test path
	content, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.Equal(t, newContent, content)
}

func TestInstallerExtract_MultipleFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	installer := NewInstaller(&Client{}, tempDir)

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	files := map[string]string{
		"file1.txt":        "content 1",
		"file2.txt":        "content 2",
		"subdir/file3.txt": "content 3",
	}

	for path, content := range files {
		writer, err := zipWriter.Create(path)
		require.NoError(t, err)
		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, zipWriter.Close())

	zipData := buf.Bytes()

	err := installer.Extract("test-plugin", zipData)
	require.NoError(t, err)

	for path, expectedContent := range files {
		filePath := filepath.Join(tempDir, "test-plugin", path)
		//nolint:G304 // controlled test path
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, expectedContent, string(content))
	}
}
