package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type ToolHandler func(ctx context.Context, params json.RawMessage) (any, error)

type ClientRequestHandler struct {
	handlers map[string]ToolHandler
	mu       sync.RWMutex
	logger   *slog.Logger
}

func NewClientRequestHandler(logger *slog.Logger) *ClientRequestHandler {
	h := &ClientRequestHandler{
		handlers: make(map[string]ToolHandler),
		logger:   logger,
	}
	h.registerHandlers()
	return h
}

func (h *ClientRequestHandler) registerHandlers() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers["fs/read_text_file"] = h.handleReadTextFile
	h.handlers["fs/write_text_file"] = h.handleWriteTextFile
	h.handlers["fs/list_directory"] = h.handleListDirectory
	h.handlers["terminal/exec"] = h.handleTerminalExec
	h.handlers["terminal/write_input"] = h.handleTerminalWriteInput
}

func (h *ClientRequestHandler) HandleRequest(ctx context.Context, method string, params json.RawMessage) (any, error) {
	h.mu.RLock()
	handler, ok := h.handlers[method]
	h.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", method)
	}

	return handler(ctx, params)
}

type ReadTextFileParams struct {
	Path string `json:"path"`
}

type ReadTextFileResult struct {
	Content string `json:"content"`
}

func (h *ClientRequestHandler) handleReadTextFile(ctx context.Context, params json.RawMessage) (any, error) {
	var p ReadTextFileParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if p.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	cleanPath, err := sanitizePath(p.Path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return ReadTextFileResult{
		Content: string(content),
	}, nil
}

type WriteTextFileParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteTextFileResult struct {
	Path string `json:"path"`
}

func (h *ClientRequestHandler) handleWriteTextFile(ctx context.Context, params json.RawMessage) (any, error) {
	var p WriteTextFileParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if p.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	cleanPath, err := sanitizePath(p.Path)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	if err := os.WriteFile(cleanPath, []byte(p.Content), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return WriteTextFileResult{
		Path: cleanPath,
	}, nil
}

type ListDirectoryParams struct {
	Path string `json:"path"`
}

type ListDirectoryResult struct {
	Entries []DirEntry `json:"entries"`
}

type DirEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

func (h *ClientRequestHandler) handleListDirectory(ctx context.Context, params json.RawMessage) (any, error) {
	var p ListDirectoryParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	path := p.Path
	if path == "" {
		path = "."
	}

	cleanPath, err := sanitizePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	result := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, DirEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}

	return ListDirectoryResult{
		Entries: result,
	}, nil
}

type TerminalExecParams struct {
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Directory string   `json:"directory,omitempty"`
	Env       []string `json:"env,omitempty"`
}

type TerminalExecResult struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

func (h *ClientRequestHandler) handleTerminalExec(ctx context.Context, params json.RawMessage) (any, error) {
	var p TerminalExecParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if p.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	if err := validateCommand(p.Command); err != nil {
		return nil, err
	}

	if err := validateCommandArgs(p.Args); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, p.Command, p.Args...)

	if p.Directory != "" {
		cleanDir, err := sanitizePath(p.Directory)
		if err != nil {
			return nil, err
		}
		cmd.Dir = cleanDir
	}

	if len(p.Env) > 0 {
		cmd.Env = append(os.Environ(), p.Env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := TerminalExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return nil, fmt.Errorf("exec command: %w", err)
	}

	result.ExitCode = 0
	return result, nil
}

type TerminalWriteInputParams struct {
	PID   string `json:"pid"`
	Input string `json:"input"`
}

type TerminalWriteInputResult struct {
	Written int `json:"written"`
}

func (h *ClientRequestHandler) handleTerminalWriteInput(ctx context.Context, params json.RawMessage) (any, error) {
	var p TerminalWriteInputParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if p.PID == "" {
		return nil, fmt.Errorf("pid is required")
	}

	if p.Input == "" {
		return TerminalWriteInputResult{
			Written: 0,
		}, nil
	}

	return nil, fmt.Errorf("terminal/write_input not implemented - interactive sessions not supported")
}

func sanitizePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)

	parts := strings.Split(cleanPath, string(filepath.Separator))
	depth := 0
	for _, part := range parts {
		if part == ".." {
			depth--
			if depth < 0 {
				return "", fmt.Errorf("path escapes root directory: %s", path)
			}
		} else if part != "" && part != "." {
			depth++
		}
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	return absPath, nil
}

var dangerousCommands = map[string]bool{
	"rm":        true,
	"rmdir":     true,
	"dd":        true,
	"mkfs":      true,
	"format":    true,
	"fdisk":     true,
	"shutdown":  true,
	"reboot":    true,
	"poweroff":  true,
	"halt":      true,
	"init":      true,
	"systemctl": true,
	"service":   true,
}

func validateCommand(cmd string) error {
	base := filepath.Base(cmd)

	if dangerousCommands[base] {
		return fmt.Errorf("dangerous command blocked: %s", cmd)
	}

	if strings.Contains(cmd, "&&") || strings.Contains(cmd, "||") || strings.Contains(cmd, ";") {
		return fmt.Errorf("command chaining not allowed")
	}

	if strings.Contains(cmd, "$(") || strings.Contains(cmd, "`") {
		return fmt.Errorf("command substitution not allowed")
	}

	return nil
}

func validateCommandArgs(args []string) error {
	for _, arg := range args {
		if strings.Contains(arg, "&&") || strings.Contains(arg, "||") || strings.Contains(arg, ";") {
			return fmt.Errorf("command chaining not allowed")
		}
		if strings.Contains(arg, "$(") || strings.Contains(arg, "`") {
			return fmt.Errorf("command substitution not allowed")
		}
	}
	return nil
}
