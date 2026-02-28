package workspace

import "errors"

var (
	ErrNoWorkspace       = errors.New("no workspace configured")
	ErrWorkspaceExists   = errors.New("workspace already exists")
	ErrProjectExists     = errors.New("project already registered in workspace")
	ErrProjectNotFound   = errors.New("project not found in workspace")
	ErrWorkspaceNotFound = errors.New("workspace not found in registry")
)
