package parser

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSingleTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		changeID  string
		lineNum   int
		wantTask  *Task
		wantError string
	}{
		{
			name:     "valid pending task",
			content:  "- [ ] **1.1** Implement parser",
			changeID: "test-change",
			lineNum:  1,
			wantTask: &Task{
				ChangeID:   "test-change",
				TaskNum:    "1.1",
				Content:    "Implement parser",
				Status:     TaskPending,
				Priority:   0,
				SourceLine: 1,
			},
		},
		{
			name:     "valid completed task",
			content:  "- [x] **1.2** Write tests",
			changeID: "test-change",
			lineNum:  5,
			wantTask: &Task{
				ChangeID:   "test-change",
				TaskNum:    "1.2",
				Content:    "Write tests",
				Status:     TaskCompleted,
				Priority:   0,
				SourceLine: 5,
			},
		},
		{
			name:      "invalid format - no prefix",
			content:   "1.1 Implement parser",
			changeID:  "test-change",
			lineNum:   1,
			wantError: "invalid format",
		},
		{
			name:      "invalid format - no checkbox",
			content:   "- **1.1** Implement parser",
			changeID:  "test-change",
			lineNum:   1,
			wantError: "invalid format",
		},
		{
			name:      "invalid format - no task number",
			content:   "- [ ] Implement parser",
			changeID:  "test-change",
			lineNum:   1,
			wantError: "invalid format",
		},
		{
			name:      "invalid format - unclosed bold",
			content:   "- [ ] **1.1 Implement parser",
			changeID:  "test-change",
			lineNum:   1,
			wantError: "invalid format",
		},
		{
			name:     "task with extra whitespace",
			content:  "-  [x]  **1.1**  Implement parser  ",
			changeID: "test-change",
			lineNum:  10,
			wantTask: &Task{
				ChangeID:   "test-change",
				TaskNum:    "1.1",
				Content:    "Implement parser  ",
				Status:     TaskCompleted,
				Priority:   0,
				SourceLine: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewTaskParser()
			task, err := parser.ParseSingleTask(tt.content, tt.changeID, tt.lineNum)

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, tt.wantTask.ChangeID, task.ChangeID)
				assert.Equal(t, tt.wantTask.TaskNum, task.TaskNum)
				assert.Equal(t, tt.wantTask.Content, task.Content)
				assert.Equal(t, tt.wantTask.Status, task.Status)
				assert.Equal(t, tt.wantTask.Priority, task.Priority)
				assert.Equal(t, tt.wantTask.SourceLine, task.SourceLine)
			}
		})
	}
}

func TestParseMultipleTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		changeID  string
		wantTasks int
		wantError string
		verifyIDs []string
	}{
		{
			name: "multiple tasks",
			content: `- [ ] **1.1** Create parser package
- [x] **1.2** Define Task entity
- [ ] **1.3** Implement parsing logic`,
			changeID:  "test-change",
			wantTasks: 3,
			verifyIDs: []string{"1.1", "1.2", "1.3"},
		},
		{
			name: "tasks with dependencies",
			content: `- [ ] **1.1** Create parser
Dependencies: 1.2
- [x] **1.2** Define Task entity
Dependencies: 
- [ ] **1.3** Implement parsing`,
			changeID:  "test-change",
			wantTasks: 3,
			verifyIDs: []string{"1.1", "1.2", "1.3"},
		},
		{
			name:      "empty content",
			content:   "",
			changeID:  "test-change",
			wantError: "empty content",
		},
		{
			name: "no valid tasks",
			content: `Some text
More text
# Header`,
			changeID:  "test-change",
			wantTasks: 0,
		},
		{
			name: "duplicate task IDs",
			content: `- [ ] **1.1** First task
- [x] **1.1** Second task`,
			changeID:  "test-change",
			wantError: "duplicate task ID",
		},
		{
			name: "mixed valid and invalid lines",
			content: `# Task List

- [ ] **1.1** Create parser
Some description text

- [x] **1.2** Define entity

## Phase 2

- [ ] **1.3** Implement`,
			changeID:  "test-change",
			wantTasks: 3,
			verifyIDs: []string{"1.1", "1.2", "1.3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewTaskParser()
			tasks, err := parser.ParseMultipleTasks(tt.content, tt.changeID)

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantError))
			} else {
				require.NoError(t, err)
				assert.Len(t, tasks, tt.wantTasks)

				if len(tt.verifyIDs) > 0 {
					for _, id := range tt.verifyIDs {
						found := false
						for _, task := range tasks {
							if task.TaskNum == id {
								found = true
								break
							}
						}
						assert.True(t, found, "task ID %s not found", id)
					}
				}
			}
		})
	}
}

func TestExtractDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lines    []string
		taskIdx  int
		wantDeps []string
	}{
		{
			name: "single dependency",
			lines: []string{
				"- [ ] **1.1** First task",
				"Dependencies: 1.2",
				"- [ ] **1.2** Second task",
			},
			taskIdx:  0,
			wantDeps: []string{"1.2"},
		},
		{
			name: "multiple dependencies",
			lines: []string{
				"- [ ] **1.1** First task",
				"Dependencies: 1.2, 1.3, 1.4",
				"- [ ] **1.2** Second task",
			},
			taskIdx:  0,
			wantDeps: []string{"1.2", "1.3", "1.4"},
		},
		{
			name: "dependencies with spaces",
			lines: []string{
				"- [ ] **1.1** First task",
				"Dependencies:  1.2  ,  1.3  ",
				"- [ ] **1.2** Second task",
			},
			taskIdx:  0,
			wantDeps: []string{"1.2", "1.3"},
		},
		{
			name: "no dependencies",
			lines: []string{
				"- [ ] **1.1** First task",
				"- [ ] **1.2** Second task",
			},
			taskIdx:  0,
			wantDeps: []string(nil),
		},
		{
			name: "dependencies with tabs",
			lines: []string{
				"- [ ] **1.1** First task",
				"\tDependencies:\t1.2,\t1.3",
				"- [ ] **1.2** Second task",
			},
			taskIdx:  0,
			wantDeps: []string{"1.2", "1.3"},
		},
		{
			name: "stop at next task",
			lines: []string{
				"- [ ] **1.1** First task",
				"Dependencies: 1.2",
				"- [ ] **1.2** Second task",
				"Dependencies: 1.1",
			},
			taskIdx:  0,
			wantDeps: []string{"1.2"},
		},
		{
			name: "case insensitive dependencies",
			lines: []string{
				"- [ ] **1.1** First task",
				"dependencies: 1.2",
				"- [ ] **1.2** Second task",
			},
			taskIdx:  0,
			wantDeps: []string{"1.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deps := extractDependencies(tt.lines, tt.taskIdx)
			assert.Equal(t, tt.wantDeps, deps)
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantStatus TaskStatus
		wantError  bool
	}{
		{
			name:       "pending lowercase",
			input:      "pending",
			wantStatus: TaskPending,
		},
		{
			name:       "pending todo",
			input:      "todo",
			wantStatus: TaskPending,
		},
		{
			name:       "in progress",
			input:      "in_progress",
			wantStatus: TaskInProgress,
		},
		{
			name:       "doing",
			input:      "doing",
			wantStatus: TaskInProgress,
		},
		{
			name:       "active",
			input:      "active",
			wantStatus: TaskInProgress,
		},
		{
			name:       "completed",
			input:      "completed",
			wantStatus: TaskCompleted,
		},
		{
			name:       "done",
			input:      "done",
			wantStatus: TaskCompleted,
		},
		{
			name:       "finished",
			input:      "finished",
			wantStatus: TaskCompleted,
		},
		{
			name:       "with spaces",
			input:      "  completed  ",
			wantStatus: TaskCompleted,
		},
		{
			name:      "invalid status",
			input:     "invalid",
			wantError: true,
		},
		{
			name:      "empty status",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := NormalizeStatus(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidStatus)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus, status)
			}
		})
	}
}

func TestTaskValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		task      *Task
		wantError string
	}{
		{
			name: "valid task",
			task: &Task{
				TaskNum:   "1.1",
				Content:   "Test task",
				Status:    TaskPending,
				DependsOn: []string{"1.2"},
			},
		},
		{
			name: "missing task number",
			task: &Task{
				Content: "Test task",
				Status:  TaskPending,
			},
			wantError: "missing task number",
		},
		{
			name: "invalid task ID with spaces",
			task: &Task{
				TaskNum: "1.1 task",
				Content: "Test task",
				Status:  TaskPending,
			},
			wantError: "invalid task id",
		},
		{
			name: "invalid task ID with special chars",
			task: &Task{
				TaskNum: "1.1@test",
				Content: "Test task",
				Status:  TaskPending,
			},
			wantError: "invalid task id",
		},
		{
			name: "empty task ID",
			task: &Task{
				TaskNum: "",
				Content: "Test task",
				Status:  TaskPending,
			},
			wantError: "missing task number",
		},
		{
			name: "invalid status",
			task: &Task{
				TaskNum: "1.1",
				Content: "Test task",
				Status:  TaskStatus("invalid"),
			},
			wantError: "invalid status",
		},
		{
			name: "invalid dependency format",
			task: &Task{
				TaskNum:   "1.1",
				Content:   "Test task",
				Status:    TaskPending,
				DependsOn: []string{"1.2 task"},
			},
			wantError: "invalid dependency format",
		},
		{
			name: "multiple invalid dependencies",
			task: &Task{
				TaskNum:   "1.1",
				Content:   "Test task",
				Status:    TaskPending,
				DependsOn: []string{"1.2", "1.3@test", "1.4"},
			},
			wantError: "invalid dependency format",
		},
		{
			name: "valid task with dots",
			task: &Task{
				TaskNum:   "1.1.2",
				Content:   "Test task",
				Status:    TaskPending,
				DependsOn: []string{"1.1.1"},
			},
		},
		{
			name: "valid task with hyphens",
			task: &Task{
				TaskNum:   "1-1-2",
				Content:   "Test task",
				Status:    TaskPending,
				DependsOn: []string{"1-1-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.task.Validate()
			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantError))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status TaskStatus
		valid  bool
	}{
		{"pending", TaskPending, true},
		{"in_progress", TaskInProgress, true},
		{"completed", TaskCompleted, true},
		{"invalid", TaskStatus("invalid"), false},
		{"empty", TaskStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, tt.status.Valid())
		})
	}
}

func TestIsValidTaskID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"simple number", "1.1", true},
		{"with dots", "1.1.2.3", true},
		{"with hyphens", "1-1-2", true},
		{"alphanumeric", "task-1", true},
		{"mixed", "1.1-task-2", true},
		{"empty", "", false},
		{"with space", "1.1 task", false},
		{"with special char", "1.1@test", false},
		{"with underscore", "1_1", false},
		{"with plus", "1+1", false},
		{"with slash", "1/1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, isValidTaskID(tt.id))
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	t.Parallel()

	t.Run("ErrInvalidFormat", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("no task prefix")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidFormat, err)

		// Test that we can unwrap and check for ErrInvalidFormat
		assert.ErrorIs(t, wrapped, ErrInvalidFormat)
	})

	t.Run("ErrInvalidTaskID", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("task ID has spaces")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidTaskID, err)

		// Test that we can unwrap and check for ErrInvalidTaskID
		assert.ErrorIs(t, wrapped, ErrInvalidTaskID)
	})

	t.Run("ErrDuplicateID", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("task ID already exists")
		wrapped := fmt.Errorf("%w: %v", ErrDuplicateID, err)

		// Test that we can unwrap and check for ErrDuplicateID
		assert.ErrorIs(t, wrapped, ErrDuplicateID)
	})

	t.Run("ErrInvalidStatus", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("unknown status value")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidStatus, err)

		// Test that we can unwrap and check for ErrInvalidStatus
		assert.ErrorIs(t, wrapped, ErrInvalidStatus)
	})

	t.Run("ErrInvalidYAML", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("YAML parse failed")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidYAML, err)

		// Test that we can unwrap and check for ErrInvalidYAML
		assert.ErrorIs(t, wrapped, ErrInvalidYAML)
	})

	t.Run("ErrEmptyContent", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("content is empty")
		wrapped := fmt.Errorf("%w: %v", ErrEmptyContent, err)

		// Test that we can unwrap and check for ErrEmptyContent
		assert.ErrorIs(t, wrapped, ErrEmptyContent)
	})

	t.Run("ErrMissingTaskNum", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("task number not found")
		wrapped := fmt.Errorf("%w: %v", ErrMissingTaskNum, err)

		// Test that we can unwrap and check for ErrMissingTaskNum
		assert.ErrorIs(t, wrapped, ErrMissingTaskNum)
	})

	t.Run("ErrInvalidDepends", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("dependency format invalid")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidDepends, err)

		// Test that we can unwrap and check for ErrInvalidDepends
		assert.ErrorIs(t, wrapped, ErrInvalidDepends)
	})

	t.Run("wrapped error checking", func(t *testing.T) {
		t.Parallel()

		// Create a wrapped error chain
		baseErr := errors.New("validation failed")
		wrapped := fmt.Errorf("parse task: %w", baseErr)
		fullyWrapped := fmt.Errorf("parse file: %w", wrapped)

		// Test that we can check for base error
		assert.ErrorIs(t, fullyWrapped, baseErr)
	})

	t.Run("error messages", func(t *testing.T) {
		t.Parallel()

		// Test that error messages are lowercase
		assert.Equal(t, "invalid format", ErrInvalidFormat.Error())
		assert.Equal(t, "invalid task ID", ErrInvalidTaskID.Error())
		assert.Equal(t, "duplicate task ID", ErrDuplicateID.Error())
		assert.Equal(t, "invalid status", ErrInvalidStatus.Error())
		assert.Equal(t, "invalid YAML metadata", ErrInvalidYAML.Error())
		assert.Equal(t, "empty content", ErrEmptyContent.Error())
		assert.Equal(t, "missing task number", ErrMissingTaskNum.Error())
		assert.Equal(t, "invalid dependency format", ErrInvalidDepends.Error())

		// Test no trailing punctuation
		assert.False(t, hasTrailingPunctuation(ErrInvalidFormat.Error()))
		assert.False(t, hasTrailingPunctuation(ErrInvalidTaskID.Error()))
		assert.False(t, hasTrailingPunctuation(ErrDuplicateID.Error()))
		assert.False(t, hasTrailingPunctuation(ErrInvalidStatus.Error()))
		assert.False(t, hasTrailingPunctuation(ErrEmptyContent.Error()))
	})

	t.Run("sentinel errors are unique", func(t *testing.T) {
		t.Parallel()

		// Test that each error is unique
		assert.NotEqual(t, ErrInvalidFormat, ErrInvalidTaskID)
		assert.NotEqual(t, ErrInvalidTaskID, ErrDuplicateID)
		assert.NotEqual(t, ErrDuplicateID, ErrInvalidStatus)
		assert.NotEqual(t, ErrInvalidStatus, ErrEmptyContent)
	})
}

func hasTrailingPunctuation(s string) bool {
	return len(s) > 0 && (s[len(s)-1] == '.' || s[len(s)-1] == '!' || s[len(s)-1] == '?')
}
