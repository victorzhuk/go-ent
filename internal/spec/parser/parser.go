package parser

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type TaskParser interface {
	ParseSingleTask(content string, changeID string, lineNum int) (*Task, error)
	ParseMultipleTasks(content string, changeID string) ([]*Task, error)
	ParseTasksFile(path string, changeID string) ([]*Task, error)
}

type taskParser struct{}

func NewTaskParser() TaskParser {
	return &taskParser{}
}

func (p *taskParser) ParseSingleTask(content string, changeID string, lineNum int) (*Task, error) {
	task, err := parseTaskLine(content, changeID, lineNum)
	if err != nil {
		return nil, fmt.Errorf("parse task line: %w", err)
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	return task, nil
}

func (p *taskParser) ParseMultipleTasks(content string, changeID string) ([]*Task, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	lines := strings.Split(content, "\n")
	var tasks []*Task
	seenIDs := make(map[string]bool)

	for i, line := range lines {
		task, err := parseTaskLine(line, changeID, i+1)
		if err != nil {
			if err == ErrInvalidFormat {
				continue
			}
			return nil, fmt.Errorf("parse line %d: %w", i+1, err)
		}

		if task == nil {
			continue
		}

		taskID := fmt.Sprintf("%s:%s", changeID, task.TaskNum)
		if seenIDs[taskID] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateID, taskID)
		}
		seenIDs[taskID] = true

		task.ID = taskID
		task.SyncedAt = time.Now()
		task.DependsOn = extractDependencies(lines, i)

		if err := task.Validate(); err != nil {
			return nil, fmt.Errorf("validate task %s: %w", taskID, err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (p *taskParser) ParseTasksFile(path string, changeID string) ([]*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	content := string(data)
	return p.ParseMultipleTasks(content, changeID)
}

func parseTaskLine(line, changeID string, lineNum int) (*Task, error) {
	line = strings.TrimRight(line, "\r\n")

	if len(line) < 2 || (line[0:2] != "- " && line[0:2] != "* ") {
		return nil, ErrInvalidFormat
	}

	content := strings.TrimLeft(line[2:], " ")

	if len(content) < 6 || (content[0:3] != "[ ]" && content[0:3] != "[x]") {
		return nil, ErrInvalidFormat
	}

	completed := content[0:3] == "[x]"
	content = strings.TrimLeft(content[3:], " ")

	if len(content) < 2 || content[0:2] != "**" {
		return nil, ErrInvalidFormat
	}

	content = content[2:]
	endIdx := strings.Index(content, "**")
	if endIdx == -1 {
		return nil, ErrInvalidFormat
	}

	taskNum := content[:endIdx]
	content = strings.TrimLeft(content[endIdx+2:], " ")

	status := TaskPending
	if completed {
		status = TaskCompleted
	}

	return &Task{
		ChangeID:   changeID,
		TaskNum:    taskNum,
		Content:    content,
		Status:     status,
		Priority:   0,
		SourceLine: lineNum,
	}, nil
}

func extractDependencies(lines []string, taskLineIdx int) []string {
	var deps []string

	for i := taskLineIdx + 1; i < len(lines); i++ {
		line := strings.TrimLeft(lines[i], " \t")
		if len(line) == 0 {
			continue
		}

		if len(line) > 2 && (line[0:2] == "- " || line[0:2] == "* ") {
			break
		}

		if len(line) > 12 && (line[:13] == "Dependencies:" || line[:13] == "dependencies:") {
			depContent := strings.TrimLeft(line[13:], " ")
			depNums := strings.FieldsFunc(depContent, func(r rune) bool {
				return r == ',' || r == ' ' || r == '\t'
			})
			for _, num := range depNums {
				if num != "" {
					deps = append(deps, num)
				}
			}
			break
		}

		if len(line) > 2 && (line[:2] == "# " || (len(line) > 3 && line[:3] == "## ")) {
			break
		}
	}

	return deps
}
