package workspace

import (
	"path/filepath"

	"github.com/victorzhuk/go-ent/internal/xdg"
)

func registryPath() string {
	return filepath.Join(xdg.ConfigDir(), "workspaces.yaml")
}

func workspaceDataDir(name string) string {
	return filepath.Join(xdg.DataDir(), "workspaces", name)
}

func workspaceCacheDir(name string) string {
	return filepath.Join(xdg.CacheDir(), "workspaces", name)
}
