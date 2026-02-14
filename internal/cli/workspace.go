package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/workspace"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces for multi-project development",
		Long: `Workspaces allow sharing skills, specs, and configuration across multiple projects.

A workspace has two parts:
  - Shared directory (version-controlled): skills/ and openspec/
  - Per-user XDG state: config, project registry, and cache`,
	}

	cmd.AddCommand(newWorkspaceInitCmd())
	cmd.AddCommand(newWorkspaceAddCmd())
	cmd.AddCommand(newWorkspaceListCmd())
	cmd.AddCommand(newWorkspaceSetCmd())
	cmd.AddCommand(newWorkspaceInfoCmd())
	cmd.AddCommand(newWorkspaceSyncCmd())

	return cmd
}

func newWorkspaceInitCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "init <path>",
		Short: "Initialize a new workspace",
		Long: `Creates a workspace at the given path with shared skills/ and openspec/ directories.

Registers the workspace in the XDG workspace registry and creates per-user
workspace config and project registry.

Examples:
  ent workspace init /path/to/workspace --name=enterprise-apps
  ent workspace init . --name=my-team`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceInit(args[0], name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Workspace name (defaults to directory name)")

	return cmd
}

func runWorkspaceInit(path, name string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if name == "" {
		name = filepath.Base(absPath)
	}

	reg, err := workspace.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if _, exists := reg.Workspaces[name]; exists {
		return fmt.Errorf("%w: %s", workspace.ErrWorkspaceExists, name)
	}

	for _, dir := range []string{
		filepath.Join(absPath, "skills"),
		filepath.Join(absPath, "openspec", "specs"),
		filepath.Join(absPath, "openspec", "changes"),
	} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	projectYAML := fmt.Sprintf("name: %s\ndescription: Workspace shared specs\n", name)
	projectPath := filepath.Join(absPath, "openspec", "project.yaml")
	if err := os.WriteFile(projectPath, []byte(projectYAML), 0o600); err != nil {
		return fmt.Errorf("write project.yaml: %w", err)
	}

	reg.Workspaces[name] = absPath
	if err := workspace.SaveRegistry(reg); err != nil {
		return fmt.Errorf("save registry: %w", err)
	}

	if err := workspace.SaveWorkspaceConfig(name, workspace.DefaultWorkspaceConfig()); err != nil {
		return fmt.Errorf("save workspace config: %w", err)
	}

	if err := workspace.SaveProjectsRegistry(name, &workspace.ProjectsRegistry{}); err != nil {
		return fmt.Errorf("save projects registry: %w", err)
	}

	fmt.Printf("Workspace '%s' initialized at %s\n", name, absPath)
	return nil
}

func newWorkspaceAddCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "add [project-path]",
		Short: "Add a project to a workspace",
		Long: `Links a project to a workspace. Creates .go-ent/workspace.yaml in the project
and registers it in the workspace's project registry.

The project path defaults to the current directory.

Examples:
  ent workspace add --workspace=enterprise-apps
  ent workspace add /path/to/project --workspace=my-team`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := "."
			if len(args) > 0 {
				projectPath = args[0]
			}
			return runWorkspaceAdd(projectPath, wsName)
		},
	}

	cmd.Flags().StringVar(&wsName, "workspace", "", "Workspace name (required)")
	_ = cmd.MarkFlagRequired("workspace")

	return cmd
}

func runWorkspaceAdd(projectPath, wsName string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if _, err := workspace.ResolveWorkspace(wsName); err != nil {
		return err
	}

	projects, err := workspace.LoadProjectsRegistry(wsName)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}

	projectName := filepath.Base(absPath)
	for _, p := range projects.Projects {
		if p.Path == absPath {
			return fmt.Errorf("%w: %s", workspace.ErrProjectExists, projectName)
		}
	}

	projects.Projects = append(projects.Projects, workspace.ProjectRef{
		Name: projectName,
		Path: absPath,
	})

	if err := workspace.SaveProjectsRegistry(wsName, projects); err != nil {
		return fmt.Errorf("save projects: %w", err)
	}

	if err := workspace.WriteWorkspaceRef(absPath, wsName); err != nil {
		return fmt.Errorf("write workspace ref: %w", err)
	}

	fmt.Printf("Added project '%s' to workspace '%s'\n", projectName, wsName)
	return nil
}

func newWorkspaceListCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces or projects in a workspace",
		Long: `Without --workspace, lists all registered workspaces.
With --workspace, lists projects in that workspace.

Examples:
  ent workspace list
  ent workspace list --workspace=enterprise-apps`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceList(wsName)
		},
	}

	cmd.Flags().StringVar(&wsName, "workspace", "", "Show projects for this workspace")

	return cmd
}

func runWorkspaceList(wsName string) error {
	if wsName != "" {
		return listWorkspaceProjects(wsName)
	}
	return listWorkspaces()
}

func listWorkspaces() error {
	reg, err := workspace.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if len(reg.Workspaces) == 0 {
		fmt.Println("No workspaces registered")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tPATH")
	_, _ = fmt.Fprintln(w, "----\t----")

	for name, path := range reg.Workspaces {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", name, path)
	}

	return w.Flush()
}

func listWorkspaceProjects(wsName string) error {
	projects, err := workspace.LoadProjectsRegistry(wsName)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}

	if len(projects.Projects) == 0 {
		fmt.Printf("No projects in workspace '%s'\n", wsName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tPATH\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "----\t----\t-----------")

	for _, p := range projects.Projects {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Path, p.Description)
	}

	return w.Flush()
}

func newWorkspaceSetCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set workspace configuration values",
		Long: `Sets configuration values in the workspace config.

Supported keys:
  models.<name>   Set model mapping (e.g., models.fast=haiku)

Examples:
  ent workspace set models.fast haiku --workspace=enterprise-apps`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceSet(args[0], args[1], wsName)
		},
	}

	cmd.Flags().StringVar(&wsName, "workspace", "", "Workspace name")

	return cmd
}

func runWorkspaceSet(key, value, wsName string) error {
	if wsName == "" {
		wsName, _ = workspace.DetectWorkspace(".")
		if wsName == "" {
			return fmt.Errorf("no workspace specified and none detected in current directory")
		}
	}

	cfg, err := workspace.LoadWorkspaceConfig(wsName)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid key format: use section.field (e.g., models.fast)")
	}

	switch parts[0] {
	case "models":
		validModels := map[string]bool{"fast": true, "main": true, "heavy": true}
		if !validModels[parts[1]] {
			return fmt.Errorf("invalid model tier %q: must be one of [fast, main, heavy]", parts[1])
		}
		if cfg.Models == nil {
			cfg.Models = make(map[string]string)
		}
		cfg.Models[parts[1]] = value
	default:
		return fmt.Errorf("unknown config section: %s", parts[0])
	}

	if err := workspace.SaveWorkspaceConfig(wsName, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Set %s = %s in workspace '%s'\n", key, value, wsName)
	return nil
}

func newWorkspaceInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show workspace details for current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceInfo()
		},
	}
}

func runWorkspaceInfo() error {
	wsName, err := workspace.DetectWorkspace(".")
	if err != nil {
		return fmt.Errorf("detect workspace: %w", err)
	}

	if wsName == "" {
		fmt.Println("Current project is not part of any workspace")
		return nil
	}

	ws, err := workspace.ResolveWorkspace(wsName)
	if err != nil {
		return fmt.Errorf("resolve workspace: %w", err)
	}

	fmt.Printf("Workspace: %s\n", ws.Name)
	fmt.Printf("Path:      %s\n", ws.Path)
	fmt.Printf("Projects:  %d\n", len(ws.Projects))

	if len(ws.Projects) > 0 {
		fmt.Println()
		for _, p := range ws.Projects {
			fmt.Printf("  - %s (%s)\n", p.Name, p.Path)
		}
	}

	skillDirs := workspace.SkillsDirs(ws)
	if len(skillDirs) > 0 {
		fmt.Printf("\nShared skills: %s\n", skillDirs[0])
	}

	return nil
}

func newWorkspaceSyncCmd() *cobra.Command {
	var projectName string
	var all bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize project specs with workspace",
		Long: `Synchronizes project specs to/from the workspace.

Operations performed:
  1. Index project specs into workspace BoltDB cache
  2. Copy shared workspace specs to projects (with ws- prefix)
  3. Rebuild cross-project dependency links

Examples:
  ent workspace sync
  ent workspace sync --all
  ent workspace sync --project=app-api --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceSync(projectName, all, dryRun)
		},
	}

	cmd.Flags().StringVar(&projectName, "project", "", "Sync only this project")
	cmd.Flags().BoolVar(&all, "all", false, "Sync all projects in workspace")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without writing")

	return cmd
}

func runWorkspaceSync(projectName string, all, dryRun bool) error {
	wsName, err := workspace.DetectWorkspace(".")
	if err != nil {
		return fmt.Errorf("detect workspace: %w", err)
	}

	if wsName == "" {
		return fmt.Errorf("current project is not part of any workspace")
	}

	ws, err := workspace.ResolveWorkspace(wsName)
	if err != nil {
		return err
	}

	projects := ws.Projects
	if !all && projectName != "" {
		var filtered []workspace.ProjectRef
		for _, p := range projects {
			if p.Name == projectName {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("%w: %s", workspace.ErrProjectNotFound, projectName)
		}
		projects = filtered
	} else if !all {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get cwd: %w", err)
		}
		var filtered []workspace.ProjectRef
		for _, p := range ws.Projects {
			if p.Path == cwd {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("current directory not registered in workspace")
		}
		projects = filtered
	}

	for _, p := range projects {
		if dryRun {
			fmt.Printf("[dry-run] Would sync project %s → workspace %s\n", p.Name, ws.Name)
			printSyncPreview(p, ws)
			continue
		}

		fmt.Printf("Syncing project %s → workspace %s...\n", p.Name, ws.Name)

		specs, changes, pulled, err := workspace.SyncProject(ws, p)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		fmt.Printf("  Indexed %d specs, %d active changes\n", specs, changes)
		if pulled > 0 {
			fmt.Printf("  Pulled %d workspace specs\n", pulled)
		}
	}

	fmt.Println("Sync complete")
	return nil
}

func printSyncPreview(p workspace.ProjectRef, ws *workspace.Workspace) {
	specsDir := filepath.Join(p.Path, "openspec", "specs")
	if entries, err := os.ReadDir(specsDir); err == nil {
		specCount := 0
		for _, e := range entries {
			if e.IsDir() {
				specCount++
			}
		}
		fmt.Printf("  Specs to index: %d\n", specCount)
	}

	changesDir := filepath.Join(p.Path, "openspec", "changes")
	if entries, err := os.ReadDir(changesDir); err == nil {
		changeCount := 0
		for _, e := range entries {
			if e.IsDir() && e.Name() != "archive" {
				changeCount++
			}
		}
		fmt.Printf("  Changes to index: %d\n", changeCount)
	}

	wsSpecsDir := filepath.Join(ws.Path, "openspec", "specs")
	if entries, err := os.ReadDir(wsSpecsDir); err == nil {
		wsSpecCount := 0
		for _, e := range entries {
			if e.IsDir() {
				wsSpecCount++
			}
		}
		if wsSpecCount > 0 {
			fmt.Printf("  Workspace specs to pull: %d\n", wsSpecCount)
		}
	}
}
