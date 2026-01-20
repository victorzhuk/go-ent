package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/victorzhuk/go-ent/internal/spec"
	"gopkg.in/yaml.v3"
)

const helpText = `Migrate registry from YAML to BoltDB.

Usage:
  migrate-registry [options]

Options:
  -registry string
        path to registry.yaml file (default "openspec/registry.yaml")
  -bolt string
        path to BoltDB database (default "openspec/registry.db")
  -dry-run
        preview changes without writing to BoltDB
  -verbose
        verbose output

Examples:
  # Preview migration
  migrate-registry --dry-run

  # Perform migration
  migrate-registry

  # Use custom paths
  migrate-registry --registry /path/to/registry.yaml --bolt /path/to/registry.db
`

type yamlTime time.Time

func (t *yamlTime) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	layouts := []string{
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999999999",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, s)
		if err == nil {
			*t = yamlTime(parsed)
			return nil
		}
	}
	return fmt.Errorf("invalid time format: %s", s)
}

func (t yamlTime) Time() time.Time {
	return time.Time(t)
}

var (
	showHelp     = flag.Bool("help", false, "show help text")
	registryPath = flag.String("registry", "openspec/registry.yaml", "path to registry.yaml file")
	boltPath     = flag.String("bolt", "openspec/registry.db", "path to BoltDB database")
	dryRun       = flag.Bool("dry-run", false, "preview changes without writing to BoltDB")
	verbose      = flag.Bool("verbose", false, "verbose output")
)

type yamlRegistry struct {
	Version  string                        `yaml:"version"`
	SyncedAt yamlTime                      `yaml:"synced_at"`
	Changes  map[string]spec.ChangeSummary `yaml:"changes"`
	Archived map[string]spec.ChangeSummary `yaml:"archived"`
	Tasks    []yamlTask                    `yaml:"tasks"`
}

type yamlTask struct {
	_           struct{}                `yaml:"id"`
	ChangeID    string                  `yaml:"change_id"`
	TaskNum     string                  `yaml:"task_num"`
	Content     string                  `yaml:"content"`
	Status      spec.RegistryTaskStatus `yaml:"status"`
	Priority    spec.TaskPriority       `yaml:"priority"`
	DependsOn   []spec.TaskID           `yaml:"depends_on,omitempty"`
	BlockedBy   []spec.TaskID           `yaml:"blocked_by,omitempty"`
	Assignee    string                  `yaml:"assignee,omitempty"`
	Session     string                  `yaml:"session,omitempty"`
	StartedAt   *time.Time              `yaml:"started_at,omitempty"`
	CompletedAt *time.Time              `yaml:"completed_at,omitempty"`
	Notes       string                  `yaml:"notes,omitempty"`
	SourceLine  int                     `yaml:"source_line"`
	SyncedAt    yamlTime                `yaml:"synced_at"`
}

func (yt *yamlTask) toRegistryTask() spec.RegistryTask {
	return spec.RegistryTask{
		ID: spec.TaskID{
			ChangeID: yt.ChangeID,
			TaskNum:  yt.TaskNum,
		},
		Content:     yt.Content,
		Status:      yt.Status,
		Priority:    yt.Priority,
		DependsOn:   yt.DependsOn,
		BlockedBy:   yt.BlockedBy,
		Assignee:    yt.Assignee,
		Session:     yt.Session,
		StartedAt:   yt.StartedAt,
		CompletedAt: yt.CompletedAt,
		Notes:       yt.Notes,
		SourceLine:  yt.SourceLine,
		SyncedAt:    yt.SyncedAt.Time(),
	}
}

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, helpText) }
	flag.Parse()

	if *showHelp {
		fmt.Print(helpText)
		os.Exit(0)
	}

	ctx := context.Background()

	logger := setupLogger(*verbose)
	slog.SetDefault(logger)

	if *dryRun {
		logger.Info("dry-run mode: no changes will be written")
	}

	reg, err := loadYAML(*registryPath)
	if err != nil {
		logger.Error("load registry", "error", err)
		os.Exit(1)
	}

	logger.Info("loaded registry", "version", reg.Version, "changes", len(reg.Changes), "tasks", len(reg.Tasks), "synced_at", reg.SyncedAt.Time())

	if !*dryRun {
		store, err := spec.NewBoltStore(*boltPath)
		if err != nil {
			logger.Error("open bolt store", "error", err)
			os.Exit(1)
		}

		if err := migrate(ctx, store, reg); err != nil {
			logger.Error("migrate", "error", err)
			if closeErr := store.Close(); closeErr != nil {
				logger.Error("close store", "error", closeErr)
			}
			os.Exit(1)
		}

		if err := store.Close(); err != nil {
			logger.Error("close store", "error", err)
		}
	} else {
		dryRunMigrate(reg)
	}

	logger.Info("migration complete")
}

func loadYAML(path string) (*yamlRegistry, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var reg yamlRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	return &reg, nil
}

func migrate(ctx context.Context, store *spec.BoltStore, reg *yamlRegistry) error {
	stats := struct {
		changes int
		tasks   int
		deps    int
		errors  []string
	}{}

	changeIDs := make(map[string]bool)

	for id, summary := range reg.Changes {
		if err := store.UpdateChange(summary); err != nil {
			return fmt.Errorf("update change %s: %w", id, err)
		}
		changeIDs[id] = true
		stats.changes++
		slog.Debug("migrated change", "id", id, "status", summary.Status, "total", summary.Total)
	}

	for _, t := range reg.Tasks {
		task := t.toRegistryTask()
		if err := store.UpdateTask(&task); err != nil {
			stats.errors = append(stats.errors, fmt.Sprintf("task %s: %v", task.ID.String(), err))
			slog.Warn("failed to migrate task", "id", task.ID.String(), "error", err)
			continue
		}
		stats.tasks++
		slog.Debug("migrated task", "id", task.ID.String(), "status", task.Status)
	}

	for _, t := range reg.Tasks {
		task := t.toRegistryTask()
		for _, dep := range task.DependsOn {
			if !changeIDs[dep.ChangeID] {
				slog.Warn("skipping dependency: task change not found", "from", task.ID.String(), "to", dep.String())
				continue
			}
			if err := store.AddDependency(task.ID, dep); err != nil {
				stats.errors = append(stats.errors, fmt.Sprintf("dependency %s -> %s: %v", task.ID.String(), dep.String(), err))
				slog.Warn("failed to add dependency", "from", task.ID.String(), "to", dep.String(), "error", err)
				continue
			}
			stats.deps++
			slog.Debug("added dependency", "from", task.ID.String(), "to", dep.String())
		}
	}

	if !reg.SyncedAt.Time().IsZero() {
		if err := store.SetSyncedAt(reg.SyncedAt.Time()); err != nil {
			slog.Warn("failed to set synced_at", "error", err)
		} else {
			slog.Debug("set synced_at", "time", reg.SyncedAt.Time())
		}
	}

	slog.Info("migration summary", "changes", stats.changes, "tasks", stats.tasks, "dependencies", stats.deps, "errors", len(stats.errors))
	for _, e := range stats.errors {
		slog.Error("migration error", "detail", e)
	}

	return nil
}

func dryRunMigrate(reg *yamlRegistry) {
	slog.Info("dry-run summary", "changes", len(reg.Changes), "archived", len(reg.Archived), "tasks", len(reg.Tasks))

	for id, summary := range reg.Changes {
		slog.Info("change", "id", id, "status", summary.Status, "total", summary.Total, "completed", summary.Completed, "in_progress", summary.InProgress, "blocked", summary.Blocked)
	}

	for i, t := range reg.Tasks {
		if i < 5 || *verbose {
			task := t.toRegistryTask()
			slog.Info("task", "index", i, "id", task.ID.String(), "change_id", t.ChangeID, "task_num", t.TaskNum, "status", task.Status, "priority", task.Priority)
		}
	}
}

func setupLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
