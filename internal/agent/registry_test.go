package agent

//nolint:gosec // test file with necessary file operations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAgent(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		agent       string
		path        string
		wantErr     bool
		errMsg      string
		preRegister bool
	}{
		{
			name: "valid agent",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test-agent.md")
				err := os.WriteFile(path, []byte(`---
name: test-agent
description: "Test agent"
model: gpt-4
---
`), 0600)
				require.NoError(t, err)
				return path
			},
			agent:   "test-agent",
			wantErr: false,
		},
		{
			name: "name mismatch",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test-agent.md")
				err := os.WriteFile(path, []byte(`---
name: different-agent
description: "Test agent"
model: gpt-4
---
`), 0600)
				require.NoError(t, err)
				return path
			},
			agent:   "test-agent",
			wantErr: true,
			errMsg:  "name mismatch",
		},
		{
			name: "invalid path",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/agent.md"
			},
			agent:   "test-agent",
			wantErr: true,
			errMsg:  "open",
		},
		{
			name: "duplicate name",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test-agent.md")
				err := os.WriteFile(path, []byte(`---
name: test-agent
description: "Test agent"
model: gpt-4
---
`), 0600)
				require.NoError(t, err)
				return path
			},
			agent:       "test-agent",
			wantErr:     true,
			errMsg:      "already registered",
			preRegister: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			path := tt.setup(t)

			if tt.preRegister {
				err := r.RegisterAgent(tt.agent, path)
				require.NoError(t, err)
			}

			err := r.RegisterAgent(tt.agent, path)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				meta, err := r.Get(tt.agent)
				assert.NoError(t, err)
				assert.Equal(t, tt.agent, meta.Name)
			}
		})
	}
}

func TestRegistry_UnregisterAgent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-agent.md")
	err := os.WriteFile(path, []byte(`---
name: test-agent
description: "Test agent"
model: gpt-4
---
`), 0600)
	require.NoError(t, err)

	r := NewRegistry()
	err = r.RegisterAgent("test-agent", path)
	require.NoError(t, err)

	err = r.UnregisterAgent("test-agent")
	assert.NoError(t, err)

	_, err = r.Get("test-agent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_UnregisterAgentNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.UnregisterAgent("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
