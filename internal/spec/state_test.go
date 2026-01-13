package spec

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single dependency",
			input: "<!-- depends: 1.1 -->",
			want:  []string{"1.1"},
		},
		{
			name:  "multiple dependencies",
			input: "<!-- depends: 1.1, 1.2, 1.3 -->",
			want:  []string{"1.1", "1.2", "1.3"},
		},
		{
			name:  "no dependencies",
			input: "no deps here",
			want:  nil,
		},
		{
			name:  "empty comment",
			input: "<!-- depends: -->",
			want:  []string{},
		},
		{
			name:  "extra spaces",
			input: "<!-- depends:  1.1 ,  1.2  -->",
			want:  []string{"1.1", "1.2"},
		},
		{
			name:  "missing closing bracket",
			input: "<!-- depends: 1.1, 1.2",
			want:  nil,
		},
		{
			name:  "malformed comment no colon",
			input: "<!-- depends 1.1 -->",
			want:  nil,
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "only whitespace deps",
			input: "<!-- depends:   -->",
			want:  []string{},
		},
		{
			name:  "mixed spacing",
			input: "<!-- depends:1.1,1.2-->",
			want:  []string{"1.1", "1.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParseDependencies(tt.input)

			assert.Equal(t, tt.want, got, "ParseDependencies(%q)", tt.input)
		})
	}
}

func TestGenerateChangeState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		changeID     string
		summary      *ChangeSummary
		tasks        []RegistryTask
		wantErr      bool
		wantProgress ProgressInfo
		wantCurrent  *TaskInfo
	}{
		{
			name:     "zero tasks",
			changeID: "test-001",
			summary: &ChangeSummary{
				ID:        "test-001",
				Title:     "Test Change",
				Completed: 0,
				Total:     0,
				Blocked:   0,
			},
			tasks: []RegistryTask{},
			wantProgress: ProgressInfo{
				Completed: 0,
				Total:     0,
				Percent:   0,
			},
		},
		{
			name:     "all pending tasks",
			changeID: "test-001",
			summary: &ChangeSummary{
				ID:        "test-001",
				Title:     "Test Change",
				Completed: 0,
				Total:     3,
				Blocked:   1,
			},
			tasks: []RegistryTask{
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "1"},
					Content:    "Task 1",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{},
					SourceLine: 10,
				},
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "2"},
					Content:    "Task 2",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{{ChangeID: "test-001", TaskNum: "1"}},
					SourceLine: 15,
				},
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "3"},
					Content:    "Task 3",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{},
					SourceLine: 20,
				},
			},
			wantProgress: ProgressInfo{
				Completed: 0,
				Total:     3,
				Percent:   0,
			},
			wantCurrent: &TaskInfo{
				ID:      TaskID{ChangeID: "test-001", TaskNum: "1"},
				Content: "Task 1",
				Line:    10,
				Status:  RegStatusPending,
			},
		},
		{
			name:     "in-progress task",
			changeID: "test-001",
			summary: &ChangeSummary{
				ID:        "test-001",
				Title:     "Test Change",
				Completed: 1,
				Total:     3,
				Blocked:   0,
			},
			tasks: []RegistryTask{
				{
					ID:          TaskID{ChangeID: "test-001", TaskNum: "1"},
					Content:     "Task 1",
					Status:      RegStatusCompleted,
					CompletedAt: &[]time.Time{time.Now()}[0],
					SourceLine:  10,
				},
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "2"},
					Content:    "Task 2",
					Status:     RegStatusInProgress,
					SourceLine: 15,
				},
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "3"},
					Content:    "Task 3",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{},
					SourceLine: 20,
				},
			},
			wantProgress: ProgressInfo{
				Completed: 1,
				Total:     3,
				Percent:   33,
			},
			wantCurrent: &TaskInfo{
				ID:      TaskID{ChangeID: "test-001", TaskNum: "2"},
				Content: "Task 2",
				Line:    15,
				Status:  RegStatusInProgress,
			},
		},
		{
			name:     "all completed",
			changeID: "test-001",
			summary: &ChangeSummary{
				ID:        "test-001",
				Title:     "Test Change",
				Completed: 2,
				Total:     2,
				Blocked:   0,
			},
			tasks: []RegistryTask{
				{
					ID:          TaskID{ChangeID: "test-001", TaskNum: "1"},
					Content:     "Task 1",
					Status:      RegStatusCompleted,
					CompletedAt: &[]time.Time{time.Now()}[0],
					SourceLine:  10,
				},
				{
					ID:          TaskID{ChangeID: "test-001", TaskNum: "2"},
					Content:     "Task 2",
					Status:      RegStatusCompleted,
					CompletedAt: &[]time.Time{time.Now()}[0],
					SourceLine:  15,
				},
			},
			wantProgress: ProgressInfo{
				Completed: 2,
				Total:     2,
				Percent:   100,
			},
			wantCurrent: nil,
		},
		{
			name:     "with blockers",
			changeID: "test-001",
			summary: &ChangeSummary{
				ID:        "test-001",
				Title:     "Test Change",
				Completed: 0,
				Total:     3,
				Blocked:   1,
			},
			tasks: []RegistryTask{
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "1"},
					Content:    "Task 1",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{},
					SourceLine: 10,
				},
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "2"},
					Content:    "Task 2",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{{ChangeID: "test-001", TaskNum: "1"}},
					SourceLine: 15,
				},
				{
					ID:         TaskID{ChangeID: "test-001", TaskNum: "3"},
					Content:    "Task 3",
					Status:     RegStatusPending,
					BlockedBy:  []TaskID{},
					SourceLine: 20,
				},
			},
			wantProgress: ProgressInfo{
				Completed: 0,
				Total:     3,
				Percent:   0,
			},
			wantCurrent: &TaskInfo{
				ID:      TaskID{ChangeID: "test-001", TaskNum: "1"},
				Content: "Task 1",
				Line:    10,
				Status:  RegStatusPending,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bolt := setupBoltStore(t)
			store := NewStore(t.TempDir())

			s := NewStateStore(store, bolt)

			for i := range tt.tasks {
				require.NoError(t, bolt.UpdateTask(&tt.tasks[i]))

				for _, dep := range tt.tasks[i].DependsOn {
					require.NoError(t, bolt.AddDependency(tt.tasks[i].ID, dep))
				}
			}

			require.NoError(t, bolt.UpdateChange(*tt.summary))

			got, err := s.GenerateChangeState(tt.changeID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.changeID, got.ID)
			assert.Equal(t, tt.wantProgress.Completed, got.Progress.Completed)
			assert.Equal(t, tt.wantProgress.Total, got.Progress.Total)
			assert.Equal(t, tt.wantProgress.Percent, got.Progress.Percent)

			if tt.wantCurrent == nil {
				assert.Nil(t, got.CurrentTask)
			} else {
				require.NotNil(t, got.CurrentTask)
				assert.Equal(t, tt.wantCurrent.ID, got.CurrentTask.ID)
				assert.Equal(t, tt.wantCurrent.Content, got.CurrentTask.Content)
				assert.Equal(t, tt.wantCurrent.Line, got.CurrentTask.Line)
				assert.Equal(t, tt.wantCurrent.Status, got.CurrentTask.Status)
			}
		})
	}
}

func TestGenerateRootState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		changes         []ChangeSummary
		tasks           []RegistryTask
		wantChangeCount int
		wantTaskCount   int
	}{
		{
			name:            "no changes",
			changes:         []ChangeSummary{},
			tasks:           []RegistryTask{},
			wantChangeCount: 0,
			wantTaskCount:   0,
		},
		{
			name: "changes but no next tasks",
			changes: []ChangeSummary{
				{
					ID:        "change-001",
					Title:     "First Change",
					Completed: 5,
					Total:     5,
					Blocked:   0,
				},
			},
			tasks: []RegistryTask{
				{ID: TaskID{ChangeID: "change-001", TaskNum: "1"}, Content: "Task 1", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "2"}, Content: "Task 2", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "3"}, Content: "Task 3", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "4"}, Content: "Task 4", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "5"}, Content: "Task 5", Status: RegStatusCompleted, Priority: PriorityMedium},
			},
			wantChangeCount: 1,
			wantTaskCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bolt := setupBoltStore(t)
			store := NewStore(t.TempDir())

			s := NewStateStore(store, bolt)

			for i := range tt.tasks {
				require.NoError(t, bolt.UpdateTask(&tt.tasks[i]))

				for _, dep := range tt.tasks[i].DependsOn {
					require.NoError(t, bolt.AddDependency(tt.tasks[i].ID, dep))
				}
			}

			for i := range tt.changes {
				require.NoError(t, bolt.UpdateChange(tt.changes[i]))
			}

			got, err := s.GenerateRootState()

			require.NoError(t, err)
			assert.Len(t, got.ActiveChanges, tt.wantChangeCount)
			assert.Len(t, got.RecommendedTasks, tt.wantTaskCount)
			assert.False(t, got.Updated.IsZero())

			for i, change := range got.ActiveChanges {
				assert.Equal(t, tt.changes[i].ID, change.ID)
				assert.Equal(t, tt.changes[i].Completed, change.Completed)
				assert.Equal(t, tt.changes[i].Total, change.Total)
				assert.Equal(t, tt.changes[i].Blocked, change.Blocked)
			}
		})
	}
}

func TestWriteRootStateMd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		changes    []ChangeSummary
		tasks      []RegistryTask
		nextTasks  []RegistryTask
		wantInFile []string
	}{
		{
			name: "no tasks available",
			changes: []ChangeSummary{
				{
					ID:        "change-001",
					Title:     "First Change",
					Completed: 5,
					Total:     5,
					Blocked:   0,
				},
			},
			tasks: []RegistryTask{
				{ID: TaskID{ChangeID: "change-001", TaskNum: "1"}, Content: "Task 1", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "2"}, Content: "Task 2", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "3"}, Content: "Task 3", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "4"}, Content: "Task 4", Status: RegStatusCompleted, Priority: PriorityMedium},
				{ID: TaskID{ChangeID: "change-001", TaskNum: "5"}, Content: "Task 5", Status: RegStatusCompleted, Priority: PriorityMedium},
			},
			nextTasks: []RegistryTask{},
			wantInFile: []string{
				"# OpenSpec State",
				"## Active Changes",
				"| Change | Progress | Blocked |",
				"## Recommended Next",
				"No unblocked tasks available",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bolt := setupBoltStore(t)
			store := NewStore(t.TempDir())

			s := NewStateStore(store, bolt)

			for i := range tt.tasks {
				require.NoError(t, bolt.UpdateTask(&tt.tasks[i]))

				for _, dep := range tt.tasks[i].DependsOn {
					require.NoError(t, bolt.AddDependency(tt.tasks[i].ID, dep))
				}
			}

			for i := range tt.changes {
				require.NoError(t, bolt.UpdateChange(tt.changes[i]))
			}

			for i := range tt.nextTasks {
				require.NoError(t, bolt.UpdateTask(&tt.nextTasks[i]))
			}

			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "root-state.md")

			err := s.WriteRootStateMd(outputPath)

			require.NoError(t, err)

			content, err := os.ReadFile(outputPath)
			require.NoError(t, err)

			contentStr := string(content)

			for _, want := range tt.wantInFile {
				assert.Contains(t, contentStr, want)
			}
		})
	}
}

func TestParseTasksWithDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		changeID      string
		tasksContent  string
		wantTaskCount int
	}{
		{
			name:          "empty file",
			changeID:      "test-001",
			tasksContent:  ``,
			wantTaskCount: 0,
		},
		{
			name:     "valid tasks without dependencies",
			changeID: "test-001",
			tasksContent: `# Tasks

- [ ] Task 1
- [ ] Task 2
- [x] Task 3
`,
			wantTaskCount: 3,
		},
		{
			name:     "valid tasks with dependencies",
			changeID: "test-001",
			tasksContent: `# Tasks

- [ ] Task 1 <!-- depends: -->
- [ ] Task 2 <!-- depends: 1 -->
- [x] Task 3 <!-- depends: 1, 2 -->
`,
			wantTaskCount: 3,
		},
		{
			name:     "checked and unchecked tasks",
			changeID: "test-001",
			tasksContent: `# Tasks

- [ ] Pending task
- [x] Completed task
- [ ] Another pending
- [X] Also completed (uppercase)
`,
			wantTaskCount: 4,
		},
		{
			name:     "tasks in different sections",
			changeID: "test-001",
			tasksContent: `# Tasks

## Phase 1

- [ ] Task 1
- [ ] Task 2

## Phase 2

- [ ] Task 3
- [x] Task 4
`,
			wantTaskCount: 4,
		},
		{
			name:     "line numbers and content extraction",
			changeID: "test-001",
			tasksContent: `# Header

Some text here.

- [ ] Task at line 4
- [ ] Task at line 5 with extra words
- [x] Completed task at line 6
`,
			wantTaskCount: 3,
		},
		{
			name:     "HTML comment parsing for dependencies",
			changeID: "test-001",
			tasksContent: `# Tasks

- [ ] Task 1 <!-- depends: 2.1 -->
- [ ] Task 2 <!-- depends: 1, 2.2, 3.1 -->
- [ ] Task 3 <!-- depends: 1.1, 1.2 -->
- [x] Task 4 <!-- depends: 1, 2 -->
`,
			wantTaskCount: 4,
		},
		{
			name:     "malformed HTML comments handled",
			changeID: "test-001",
			tasksContent: `# Tasks

- [ ] Task 1 <!-- depends: 2.1
- [ ] Task 2 <!-- depends 1 -->
- [ ] Task 3
- [ ] Task 4 <!-- depends: -->
`,
			wantTaskCount: 4,
		},
		{
			name:     "extra spaces in dependencies",
			changeID: "test-001",
			tasksContent: `# Tasks

- [ ] Task 1 <!-- depends:  1.1 ,  1.2  -->
- [ ] Task 2 <!-- depends: 2.1,2.2 -->
- [ ] Task 3 <!-- depends: 3.1 -->
`,
			wantTaskCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tasksPath := filepath.Join(tmpDir, "tasks.md")

			err := os.WriteFile(tasksPath, []byte(tt.tasksContent), 0644)
			require.NoError(t, err)

			bolt := setupBoltStore(t)
			store := NewStore(t.TempDir())

			s := NewStateStore(store, bolt)

			tasks, err := s.ParseTasksWithDependencies(tt.changeID, tasksPath)

			require.NoError(t, err)
			assert.Len(t, tasks, tt.wantTaskCount)

			for _, task := range tasks {
				assert.Equal(t, tt.changeID, task.ID.ChangeID)
				assert.Equal(t, PriorityMedium, task.Priority)
				assert.False(t, task.SyncedAt.IsZero())
			}
		})
	}
}

func TestParseTasksWithDependencies_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		changeID  string
		tasksPath string
		wantErr   string
	}{
		{
			name:      "file not found",
			changeID:  "test-001",
			tasksPath: "/nonexistent/tasks.md",
			wantErr:   "open tasks.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bolt := setupBoltStore(t)
			store := NewStore(t.TempDir())

			s := NewStateStore(store, bolt)

			_, err := s.ParseTasksWithDependencies(tt.changeID, tt.tasksPath)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestStateStore_ParseTasksWithDependencies_ContentTrimming(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.md")

	content := `# Tasks

- [ ]  Task with leading space
- [ ] Task with trailing space  
- [ ]  Task with both spaces  
- [x] Completed task with spaces
- [ ] Task with <!-- depends: 1 --> HTML comment
- [ ] Task with multiple <!-- depends: 1, 2 --> dependencies
`

	err := os.WriteFile(tasksPath, []byte(content), 0644)
	require.NoError(t, err)

	bolt := setupBoltStore(t)
	store := NewStore(t.TempDir())

	s := NewStateStore(store, bolt)

	tasks, err := s.ParseTasksWithDependencies("test-001", tasksPath)

	require.NoError(t, err)
	assert.Len(t, tasks, 6)

	assert.Equal(t, "Task with leading space", tasks[0].Content)
	assert.Equal(t, "Task with trailing space", tasks[1].Content)
	assert.Equal(t, "Task with both spaces", tasks[2].Content)
	assert.Equal(t, "Completed task with spaces", tasks[3].Content)
	assert.Equal(t, "Task with", tasks[4].Content)
	assert.Equal(t, "Task with multiple", tasks[5].Content)

	assert.Equal(t, RegStatusCompleted, tasks[3].Status)

	assert.Len(t, tasks[4].DependsOn, 1)
	assert.Equal(t, "1", tasks[4].DependsOn[0].TaskNum)

	assert.Len(t, tasks[5].DependsOn, 2)
	assert.Equal(t, "1", tasks[5].DependsOn[0].TaskNum)
	assert.Equal(t, "2", tasks[5].DependsOn[1].TaskNum)
}

func TestStateStore_WriteChangeStateMd_RecentActivity(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tasks := []RegistryTask{
		{
			ID:          TaskID{ChangeID: "test-001", TaskNum: "1"},
			Content:     "Task 1",
			Status:      RegStatusCompleted,
			CompletedAt: &[]time.Time{now.Add(-2 * time.Hour)}[0],
			SourceLine:  10,
		},
		{
			ID:          TaskID{ChangeID: "test-001", TaskNum: "2"},
			Content:     "Task 2",
			Status:      RegStatusCompleted,
			CompletedAt: &[]time.Time{now.Add(-1 * time.Hour)}[0],
			SourceLine:  15,
		},
		{
			ID:          TaskID{ChangeID: "test-001", TaskNum: "3"},
			Content:     "Task 3",
			Status:      RegStatusCompleted,
			CompletedAt: &[]time.Time{now}[0],
			SourceLine:  20,
		},
	}

	bolt := setupBoltStore(t)
	store := NewStore(t.TempDir())

	s := NewStateStore(store, bolt)

	summary := &ChangeSummary{
		ID:        "test-001",
		Title:     "Test Change",
		Completed: 3,
		Total:     3,
		Blocked:   0,
	}

	require.NoError(t, bolt.UpdateChange(*summary))

	for i := range tasks {
		require.NoError(t, bolt.UpdateTask(&tasks[i]))
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "state.md")

	err := s.WriteChangeStateMd("test-001", outputPath)

	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)

	assert.Contains(t, contentStr, "## Recent Activity")
	assert.Contains(t, contentStr, "| Task | Action | Time |")
	assert.Contains(t, contentStr, "| T3 | completed |")

	assert.Contains(t, contentStr, "| T1 | completed |")
}

func TestStateStore_WriteChangeStateMd_FileCreateError(t *testing.T) {
	t.Parallel()

	bolt := setupBoltStore(t)
	store := NewStore(t.TempDir())

	s := NewStateStore(store, bolt)

	summary := &ChangeSummary{
		ID:        "test-001",
		Title:     "Test Change",
		Completed: 0,
		Total:     0,
		Blocked:   0,
	}

	require.NoError(t, bolt.UpdateChange(*summary))

	invalidPath := "/nonexistent/path/state.md"

	err := s.WriteChangeStateMd("test-001", invalidPath)

	assert.Error(t, err)
}

func TestParseDependencies_ComplexCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "numeric task numbers",
			input: "<!-- depends: 1, 2, 3 -->",
			want:  []string{"1", "2", "3"},
		},
		{
			name:  "decimal task numbers",
			input: "<!-- depends: 1.1, 1.2, 2.1 -->",
			want:  []string{"1.1", "1.2", "2.1"},
		},
		{
			name:  "mixed format",
			input: "<!-- depends: 1, 2.1, 3.2.1 -->",
			want:  []string{"1", "2.1", "3.2.1"},
		},
		{
			name:  "trailing comma",
			input: "<!-- depends: 1.1, 1.2, -->",
			want:  []string{"1.1", "1.2"},
		},
		{
			name:  "duplicate deps",
			input: "<!-- depends: 1.1, 1.1, 1.2 -->",
			want:  []string{"1.1", "1.1", "1.2"},
		},
		{
			name:  "comment after deps",
			input: "<!-- depends: 1.1, 1.2 --> some text",
			want:  []string{"1.1", "1.2"},
		},
		{
			name:  "multiple comments",
			input: "<!-- depends: 1.1 --> <!-- depends: 1.2 -->",
			want:  []string{"1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParseDependencies(tt.input)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStateStore_WriteRootStateMd_FileCreateError(t *testing.T) {
	t.Parallel()

	bolt := setupBoltStore(t)
	store := NewStore(t.TempDir())

	s := NewStateStore(store, bolt)

	invalidPath := "/nonexistent/path/root-state.md"

	err := s.WriteRootStateMd(invalidPath)

	assert.Error(t, err)
}

func TestGenerateChangeState_ErrorCases(t *testing.T) {
	t.Parallel()

	bolt := setupBoltStore(t)
	store := NewStore(t.TempDir())

	s := NewStateStore(store, bolt)

	t.Run("change not found", func(t *testing.T) {
		t.Parallel()

		_, err := s.GenerateChangeState("nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get change summary")
	})
}

func TestStateStore_ProgressPercentage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		completed int
		total     int
		want      int
	}{
		{"zero total", 0, 0, 0},
		{"all completed", 5, 5, 100},
		{"half completed", 2, 4, 50},
		{"one completed", 1, 3, 33},
		{"partial", 7, 10, 70},
		{"almost done", 9, 10, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bolt := setupBoltStore(t)
			store := NewStore(t.TempDir())

			s := NewStateStore(store, bolt)

			for i := 1; i <= tt.total; i++ {
				task := RegistryTask{
					ID:         TaskID{ChangeID: "test-001", TaskNum: string(rune('0' + i))},
					Content:    "Task " + string(rune('0'+i)),
					Status:     RegStatusPending,
					Priority:   PriorityMedium,
					SourceLine: i * 10,
					SyncedAt:   time.Now(),
				}

				if i <= tt.completed {
					task.Status = RegStatusCompleted
					now := time.Now()
					task.CompletedAt = &now
				}

				require.NoError(t, bolt.UpdateTask(&task))
			}

			summary := &ChangeSummary{
				ID:        "test-001",
				Title:     "Test Change",
				Completed: tt.completed,
				Total:     tt.total,
				Blocked:   0,
			}

			require.NoError(t, bolt.UpdateChange(*summary))

			state, err := s.GenerateChangeState("test-001")

			require.NoError(t, err)
			assert.Equal(t, tt.want, state.Progress.Percent)
		})
	}
}
