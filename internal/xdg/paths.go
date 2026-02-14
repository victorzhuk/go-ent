package xdg

import (
	"os"
	"path/filepath"
)

const appName = "go-ent"

func ConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, appName)
	}
	return filepath.Join(homeDir(), ".config", appName)
}

func DataDir() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return filepath.Join(v, appName)
	}
	return filepath.Join(homeDir(), ".local", "share", appName)
}

func CacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, appName)
	}
	return filepath.Join(homeDir(), ".cache", appName)
}

func LegacyDir() string {
	return filepath.Join(homeDir(), ".go-ent")
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return os.TempDir()
}
