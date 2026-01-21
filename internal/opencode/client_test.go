package opencode

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRequestHandler_HandleReadTextFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello world"), 0644))

	tests := []struct {
		name    string
		params  json.RawMessage
		want    any
		wantErr bool
	}{
		{
			name:   "valid read",
			params: json.RawMessage(`{"path": "` + testFile + `"}`),
			want: ReadTextFileResult{
				Content: "hello world",
			},
			wantErr: false,
		},
		{
			name:    "missing path",
			params:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "file not found",
			params:  json.RawMessage(`{"path": "/nonexistent/file.txt"}`),
			wantErr: true,
		},
		{
			name:    "path traversal blocked",
			params:  json.RawMessage(`{"path": "../../../etc/passwd"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientRequestHandler(nil)

			result, err := h.HandleRequest(context.Background(), "fs/read_text_file", tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				wantBytes, _ := json.Marshal(tt.want)
				gotBytes, _ := json.Marshal(result)
				assert.JSONEq(t, string(wantBytes), string(gotBytes))
			}
		})
	}
}

func TestClientRequestHandler_HandleWriteTextFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		params  json.RawMessage
		wantErr bool
	}{
		{
			name: "valid write",
			params: json.RawMessage(`{
				"path": "` + filepath.Join(tmpDir, "new.txt") + `",
				"content": "test content"
			}`),
			wantErr: false,
		},
		{
			name: "write with subdirs",
			params: json.RawMessage(`{
				"path": "` + filepath.Join(tmpDir, "subdir", "nested.txt") + `",
				"content": "nested content"
			}`),
			wantErr: false,
		},
		{
			name:    "missing path",
			params:  json.RawMessage(`{"content": "test"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientRequestHandler(nil)

			result, err := h.HandleRequest(context.Background(), "fs/write_text_file", tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				res, ok := result.(WriteTextFileResult)
				require.True(t, ok)

				content, err := os.ReadFile(res.Path)
				require.NoError(t, err)
				assert.NotEmpty(t, content)
			}
		})
	}
}

func TestClientRequestHandler_HandleListDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	testSubdir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(testSubdir, 0755))

	tests := []struct {
		name    string
		params  json.RawMessage
		want    ListDirectoryResult
		wantErr bool
	}{
		{
			name:    "list current dir",
			params:  json.RawMessage(`{}`),
			wantErr: false,
		},
		{
			name:    "list specific dir",
			params:  json.RawMessage(`{"path": "` + tmpDir + `"}`),
			wantErr: false,
		},
		{
			name:    "invalid path",
			params:  json.RawMessage(`{"path": "/nonexistent"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientRequestHandler(nil)

			result, err := h.HandleRequest(context.Background(), "fs/list_directory", tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				res, ok := result.(ListDirectoryResult)
				require.True(t, ok)
				assert.NotEmpty(t, res.Entries)
			}
		})
	}
}

func TestClientRequestHandler_HandleTerminalExec(t *testing.T) {
	tests := []struct {
		name    string
		params  json.RawMessage
		wantErr bool
	}{
		{
			name: "echo command",
			params: json.RawMessage(`{
				"command": "echo",
				"args": ["hello"]
			}`),
			wantErr: false,
		},
		{
			name: "ls command",
			params: json.RawMessage(`{
				"command": "ls",
				"args": ["-la"]
			}`),
			wantErr: false,
		},
		{
			name:    "missing command",
			params:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name: "dangerous command blocked",
			params: json.RawMessage(`{
				"command": "rm",
				"args": ["-rf", "/"]
			}`),
			wantErr: true,
		},
		{
			name: "command chaining blocked",
			params: json.RawMessage(`{
				"command": "echo",
				"args": ["test", "&&", "rm", "-rf", "/"]
			}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewClientRequestHandler(nil)

			result, err := h.HandleRequest(context.Background(), "terminal/exec", tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				res, ok := result.(TerminalExecResult)
				require.True(t, ok)
				assert.NotEqual(t, -1, res.ExitCode)
			}
		})
	}
}

func TestClientRequestHandler_HandleTerminalWriteInput(t *testing.T) {
	h := NewClientRequestHandler(nil)

	result, err := h.HandleRequest(context.Background(), "terminal/write_input", json.RawMessage(`{
		"pid": "12345",
		"input": "test input"
	}`))

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/tmp/test.txt",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "./test.txt",
			wantErr: false,
		},
		{
			name:    "path traversal blocked",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "complex traversal",
			path:    "/tmp/../../../etc/passwd",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sanitizePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantErr bool
	}{
		{
			name:    "safe command",
			cmd:     "echo",
			wantErr: false,
		},
		{
			name:    "safe command with path",
			cmd:     "/bin/ls",
			wantErr: false,
		},
		{
			name:    "dangerous command rm",
			cmd:     "rm",
			wantErr: true,
		},
		{
			name:    "dangerous command dd",
			cmd:     "dd",
			wantErr: true,
		},
		{
			name:    "command chaining",
			cmd:     "echo && rm -rf /",
			wantErr: true,
		},
		{
			name:    "command substitution",
			cmd:     "echo $(cat /etc/passwd)",
			wantErr: true,
		},
		{
			name:    "backtick substitution",
			cmd:     "echo `whoami`",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientRequestHandler_UnknownMethod(t *testing.T) {
	h := NewClientRequestHandler(nil)

	result, err := h.HandleRequest(context.Background(), "unknown/method", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown tool")
}

func TestJSONRPCResponse_Marshal(t *testing.T) {
	resp := jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      "test-id",
		Result:  json.RawMessage(`{"success": true}`),
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var got jsonrpcResponse
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, resp.JSONRPC, got.JSONRPC)
	assert.Equal(t, resp.ID, got.ID)
	assert.JSONEq(t, string(resp.Result), string(got.Result))
}

func TestJSONRPCResponse_Error(t *testing.T) {
	resp := jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      "test-id",
		Error: &jsonrpcError{
			Code:    -32602,
			Message: "Invalid params",
			Data:    "field missing",
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var got jsonrpcResponse
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, resp.JSONRPC, got.JSONRPC)
	assert.Equal(t, resp.ID, got.ID)
	assert.NotNil(t, got.Error)
	assert.Equal(t, resp.Error.Code, got.Error.Code)
	assert.Equal(t, resp.Error.Message, got.Error.Message)
}
