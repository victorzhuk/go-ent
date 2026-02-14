package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func DetectWorkspace(projectPath string) (string, error) {
	refPath := filepath.Join(projectPath, ".go-ent", "workspace.yaml")

	data, err := os.ReadFile(refPath) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read workspace ref: %w", err)
	}

	var ref WorkspaceRef
	if err := yaml.Unmarshal(data, &ref); err != nil {
		return "", fmt.Errorf("parse workspace ref: %w", err)
	}

	return ref.Name, nil
}

func ResolveWorkspace(name string) (*Workspace, error) {
	if name == "" {
		return nil, ErrNoWorkspace
	}

	reg, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	wsPath, ok := reg.Workspaces[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, name)
	}

	projects, err := LoadProjectsRegistry(name)
	if err != nil {
		return nil, fmt.Errorf("load projects: %w", err)
	}

	return &Workspace{
		Name:     name,
		Path:     wsPath,
		Projects: projects.Projects,
	}, nil
}

func DetectAndResolve(projectPath string) (*Workspace, error) {
	name, err := DetectWorkspace(projectPath)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, nil
	}

	ws, err := ResolveWorkspace(name)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func WriteWorkspaceRef(projectPath, workspaceName string) error {
	ref := WorkspaceRef{Name: workspaceName}

	data, err := yaml.Marshal(ref)
	if err != nil {
		return fmt.Errorf("marshal workspace ref: %w", err)
	}

	dir := filepath.Join(projectPath, ".go-ent")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create .go-ent dir: %w", err)
	}

	path := filepath.Join(dir, "workspace.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write workspace.yaml: %w", err)
	}

	return nil
}

func SkillsDirs(ws *Workspace) []string {
	if ws == nil {
		return nil
	}

	skillsDir := filepath.Join(ws.Path, "skills")
	if _, err := os.Stat(skillsDir); err == nil {
		return []string{skillsDir}
	}

	return nil
}

func WorkspaceDBPath(name string) string {
	return filepath.Join(workspaceCacheDir(name), "workspace.db")
}
