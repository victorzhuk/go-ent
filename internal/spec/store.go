package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/victorzhuk/go-ent/internal/config"
	"gopkg.in/yaml.v3"
)

type Store struct {
	rootPath string
}

func NewStore(rootPath string) *Store {
	return &Store{rootPath: rootPath}
}

func (s *Store) RootPath() string {
	return s.rootPath
}

func (s *Store) SpecPath() string {
	openspecPath := filepath.Join(s.rootPath, "openspec")
	if _, err := os.Stat(openspecPath); err == nil {
		return openspecPath
	}
	return filepath.Join(s.rootPath, ".spec")
}

func (s *Store) ConfigPath() string {
	return filepath.Join(s.rootPath, ".go-ent", "config.yaml")
}

func (s *Store) AgentsPath() string {
	return filepath.Join(s.rootPath, "plugins", "go-ent", "agents")
}

func (s *Store) SkillsPath() string {
	return filepath.Join(s.rootPath, "plugins", "go-ent", "skills")
}

func (s *Store) Exists() (bool, error) {
	_, err := os.Stat(s.SpecPath())
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat spec folder: %w", err)
}

func (s *Store) Init(project Project) error {
	specPath := s.SpecPath()

	if err := os.MkdirAll(filepath.Join(specPath, "specs"), 0750); err != nil {
		return fmt.Errorf("create specs dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(specPath, "changes"), 0750); err != nil {
		return fmt.Errorf("create changes dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(specPath, "tasks"), 0750); err != nil {
		return fmt.Errorf("create tasks dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(specPath, "changes", "archive"), 0750); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}

	projectPath := filepath.Join(specPath, "project.yaml")
	data, err := yaml.Marshal(project)
	if err != nil {
		return fmt.Errorf("marshal project: %w", err)
	}

	if err := os.WriteFile(projectPath, data, 0600); err != nil {
		return fmt.Errorf("write project.yaml: %w", err)
	}

	return nil
}

func (s *Store) ListSpecs() ([]ListItem, error) {
	specsPath := filepath.Join(s.SpecPath(), "specs")
	entries, err := os.ReadDir(specsPath)
	if err != nil {
		return nil, fmt.Errorf("read specs dir: %w", err)
	}

	items := make([]ListItem, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		specPath := filepath.Join(specsPath, id, "spec.md")

		desc := ""
		// #nosec G304 -- controlled config/template file path
		if data, err := os.ReadFile(specPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") {
					desc = strings.TrimPrefix(line, "# ")
					break
				}
			}
		}

		items = append(items, ListItem{
			ID:          id,
			Name:        id,
			Type:        "spec",
			Path:        specPath,
			Description: desc,
		})
	}

	return items, nil
}

func (s *Store) ListChanges(status string) ([]ListItem, error) {
	changesPath := filepath.Join(s.SpecPath(), "changes")
	entries, err := os.ReadDir(changesPath)
	if err != nil {
		return nil, fmt.Errorf("read changes dir: %w", err)
	}

	items := make([]ListItem, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()

		// Skip the archive directory itself (contains archived changes)
		if id == "archive" {
			continue
		}

		proposalPath := filepath.Join(changesPath, id, "proposal.md")

		desc := ""
		changeStatus := "active"
		// #nosec G304 -- controlled config/template file path
		if data, err := os.ReadFile(proposalPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") {
					desc = strings.TrimPrefix(line, "# ")
				}
				if strings.HasPrefix(line, "Status:") {
					changeStatus = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Status:")))
				}
			}
		}

		if status != "" && status != changeStatus {
			continue
		}

		items = append(items, ListItem{
			ID:          id,
			Name:        id,
			Type:        "change",
			Status:      changeStatus,
			Path:        proposalPath,
			Description: desc,
		})
	}

	return items, nil
}

func (s *Store) ListTasks() ([]ListItem, error) {
	tasksPath := filepath.Join(s.SpecPath(), "tasks")
	entries, err := os.ReadDir(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("read tasks dir: %w", err)
	}

	items := make([]ListItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".md")
		taskPath := filepath.Join(tasksPath, entry.Name())

		desc := ""
		status := "pending"
		// #nosec G304 -- controlled config/template file path
		if data, err := os.ReadFile(taskPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") {
					desc = strings.TrimPrefix(line, "# ")
				}
				if strings.HasPrefix(line, "Status:") {
					status = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Status:")))
				}
			}
		}

		items = append(items, ListItem{
			ID:          id,
			Name:        id,
			Type:        "task",
			Status:      status,
			Path:        taskPath,
			Description: desc,
		})
	}

	return items, nil
}

func (s *Store) ReadFile(path string) (string, error) {
	fullPath := filepath.Join(s.SpecPath(), path)
	data, err := os.ReadFile(fullPath) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

func (s *Store) WriteFile(path, content string) error {
	fullPath := filepath.Join(s.SpecPath(), path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func (s *Store) DeleteFile(path string) error {
	fullPath := filepath.Join(s.SpecPath(), path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete %s: %w", path, err)
	}
	return nil
}

func (s *Store) DeleteDir(path string) error {
	fullPath := filepath.Join(s.SpecPath(), path)
	if err := os.RemoveAll(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete dir %s: %w", path, err)
	}
	return nil
}

func (s *Store) RegistryPath() string {
	return filepath.Join(s.SpecPath(), "registry.yaml")
}

func (s *Store) RegistryExists() bool {
	_, err := os.Stat(s.RegistryPath())
	return err == nil
}

func (s *Store) LoadRegistry() (*Registry, error) {
	data, err := os.ReadFile(s.RegistryPath()) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("read registry.yaml: %w", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("unmarshal registry: %w", err)
	}

	return &reg, nil
}

func (s *Store) SaveRegistry(reg *Registry) error {
	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(s.RegistryPath(), data, 0600); err != nil {
		return fmt.Errorf("write registry.yaml: %w", err)
	}

	return nil
}

func (s *Store) LoadConfig() (*config.Config, error) {
	return config.Load(s.rootPath)
}

func (s *Store) SaveConfig(cfg *config.Config) error {
	cfgPath := s.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return fmt.Errorf("write config.yaml: %w", err)
	}

	return nil
}

// Generic YAML helpers
func loadYAML[T any](path string) (*T, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var obj T
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}

	return &obj, nil
}

func saveYAML[T any](path string, obj *T) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
