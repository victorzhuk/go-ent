package execution

import (
	"fmt"
	"path/filepath"
)

// NewTaskContext creates a new TaskContext with the given project path.
func NewTaskContext(projectPath string) *TaskContext {
	return &TaskContext{
		ProjectPath: projectPath,
		Files:       []string{},
	}
}

// WithChange sets the change ID.
func (tc *TaskContext) WithChange(changeID string) *TaskContext {
	tc.ChangeID = changeID
	return tc
}

// WithTask sets the task ID.
func (tc *TaskContext) WithTask(taskID string) *TaskContext {
	tc.TaskID = taskID
	return tc
}

// WithWorkflow sets the workflow ID.
func (tc *TaskContext) WithWorkflow(workflowID string) *TaskContext {
	tc.WorkflowID = workflowID
	return tc
}

// WithFiles sets the relevant files.
func (tc *TaskContext) WithFiles(files []string) *TaskContext {
	tc.Files = files
	return tc
}

// AddFile adds a file to the context.
func (tc *TaskContext) AddFile(file string) *TaskContext {
	tc.Files = append(tc.Files, file)
	return tc
}

// HasFiles returns true if context has files.
func (tc *TaskContext) HasFiles() bool {
	return len(tc.Files) > 0
}

// RelativePath converts an absolute path to a path relative to project root.
func (tc *TaskContext) RelativePath(absPath string) (string, error) {
	return filepath.Rel(tc.ProjectPath, absPath)
}

// AbsolutePath converts a relative path to an absolute path from project root.
func (tc *TaskContext) AbsolutePath(relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(tc.ProjectPath, relPath)
}

// String returns a string representation of the context.
func (tc *TaskContext) String() string {
	if tc.ChangeID != "" && tc.TaskID != "" {
		return fmt.Sprintf("%s/%s", tc.ChangeID, tc.TaskID)
	}
	if tc.ChangeID != "" {
		return tc.ChangeID
	}
	return tc.ProjectPath
}
