package workspace

type Workspace struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description,omitempty"`
	Path        string       `yaml:"-"`
	Projects    []ProjectRef `yaml:"-"`
}

type ProjectRef struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Description string `yaml:"description,omitempty"`
}

type WorkspaceRef struct {
	Name string `yaml:"workspace"`
}

type WorkspaceRegistry struct {
	Workspaces map[string]string `yaml:"workspaces"`
}

type ProjectsRegistry struct {
	Projects []ProjectRef `yaml:"projects"`
}
