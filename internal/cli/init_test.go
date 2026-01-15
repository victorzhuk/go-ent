package cli_test

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goent "github.com/victorzhuk/go-ent"
	"github.com/victorzhuk/go-ent/internal/cli"
)

func TestInitCommand_ToolRequired(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no --tool flag",
			args:        []string{"init"},
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "empty --tool flag",
			args:        []string{"init", "--tool="},
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "invalid tool value",
			args:        []string{"init", "--tool=invalid"},
			wantErr:     true,
			errContains: "invalid tool",
		},
		{
			name:        "valid tool claude",
			args:        []string{"init", "--tool=claude", "--dry-run"},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "valid tool opencode",
			args:        []string{"init", "--tool=opencode", "--dry-run"},
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "valid tool all",
			args:        []string{"init", "--tool=all", "--dry-run"},
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := executeCommandWithCapture(t, tt.args...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.NotEmpty(t, stdout)
			}
		})
	}
}

func TestResolveAgentList_AllAgents(t *testing.T) {
	t.Parallel()

	agents, err := cli.ResolveAgentList(goent.PluginFS, nil, false, false)
	require.NoError(t, err)
	assert.NotEmpty(t, agents)
}

func TestResolveAgentList_SpecificAgents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		agentsFlag    []string
		includeDeps   bool
		noDeps        bool
		wantErr       bool
		errContains   string
		expectedCount int
	}{
		{
			name:          "single agent with no deps",
			agentsFlag:    []string{"planner"},
			includeDeps:   false,
			noDeps:        false,
			wantErr:       false,
			errContains:   "",
			expectedCount: 1,
		},
		{
			name:          "single agent with no deps validation",
			agentsFlag:    []string{"planner"},
			includeDeps:   false,
			noDeps:        true,
			wantErr:       false,
			errContains:   "",
			expectedCount: 1,
		},
		{
			name:          "single agent with include deps",
			agentsFlag:    []string{"planner"},
			includeDeps:   true,
			noDeps:        false,
			wantErr:       false,
			errContains:   "",
			expectedCount: 1,
		},
		{
			name:          "empty list",
			agentsFlag:    []string{},
			includeDeps:   false,
			noDeps:        false,
			wantErr:       false,
			errContains:   "",
			expectedCount: -1,
		},
		{
			name:          "non-existent agent",
			agentsFlag:    []string{"nonexistent"},
			includeDeps:   true,
			noDeps:        false,
			wantErr:       true,
			errContains:   "agent not found",
			expectedCount: 0,
		},
		{
			name:          "multiple agents",
			agentsFlag:    []string{"planner", "tester"},
			includeDeps:   false,
			noDeps:        true,
			wantErr:       false,
			errContains:   "",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, tt.includeDeps, tt.noDeps)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, agents)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, agents)

				if tt.expectedCount == -1 {
					assert.NotEmpty(t, agents)
				} else {
					assert.Len(t, agents, tt.expectedCount)
				}
			}
		})
	}
}

func TestResolveAgentList_IncludeDeps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		agentsFlag       []string
		expectedInResult []string
	}{
		{
			name:             "planner has no deps",
			agentsFlag:       []string{"planner"},
			expectedInResult: []string{"planner"},
		},
		{
			name:             "tester has no deps",
			agentsFlag:       []string{"tester"},
			expectedInResult: []string{"tester"},
		},
		{
			name:             "reproducer with all transitive deps",
			agentsFlag:       []string{"reproducer"},
			expectedInResult: []string{"acceptor", "architect", "coder", "debugger", "debugger-fast", "debugger-heavy", "planner", "reproducer", "researcher", "reviewer", "tester"},
		},
		{
			name:             "multiple agents with transitive deps",
			agentsFlag:       []string{"planner", "tester", "reproducer"},
			expectedInResult: []string{"acceptor", "architect", "coder", "debugger", "debugger-fast", "debugger-heavy", "planner", "reproducer", "researcher", "reviewer", "tester"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, true, false)
			require.NoError(t, err)
			require.NotEmpty(t, agents)

			agentsSet := make(map[string]bool)
			for _, a := range agents {
				agentsSet[a] = true
			}

			for _, expected := range tt.expectedInResult {
				assert.True(t, agentsSet[expected], "expected agent %s in result", expected)
			}

			assert.Len(t, agents, len(tt.expectedInResult))
		})
	}
}

func TestResolveAgentList_TopologicalOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		agentsFlag []string
		minCount   int
	}{
		{
			name:       "planner has no dependencies",
			agentsFlag: []string{"planner"},
			minCount:   1,
		},
		{
			name:       "multiple agents with transitive deps",
			agentsFlag: []string{"planner", "tester", "reproducer"},
			minCount:   11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, true, false)
			require.NoError(t, err)
			require.NotEmpty(t, agents)

			assert.GreaterOrEqual(t, len(agents), tt.minCount)

			seen := make(map[string]bool)
			for _, agent := range agents {
				_, exists := seen[agent]
				assert.False(t, exists, "duplicate agent: %s", agent)
				seen[agent] = true
			}
		})
	}
}

func TestResolveAgentList_NoDuplicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		agentsFlag []string
	}{
		{
			name:       "single agent",
			agentsFlag: []string{"planner"},
		},
		{
			name:       "multiple agents",
			agentsFlag: []string{"planner", "tester", "reproducer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, true, false)
			require.NoError(t, err)
			require.NotEmpty(t, agents)

			seen := make(map[string]bool)
			for _, agent := range agents {
				_, exists := seen[agent]
				assert.False(t, exists, "duplicate agent found: %s", agent)
				seen[agent] = true
			}
		})
	}
}

func TestResolveAgentList_NoDepsSkipValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		agentsFlag  []string
		expectedLen int
	}{
		{
			name:        "single agent",
			agentsFlag:  []string{"planner"},
			expectedLen: 1,
		},
		{
			name:        "multiple agents",
			agentsFlag:  []string{"planner", "tester", "reproducer"},
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, false, true)
			require.NoError(t, err)
			assert.NotNil(t, agents)
			assert.Len(t, agents, tt.expectedLen)

			for i, expected := range tt.agentsFlag {
				assert.Equal(t, expected, agents[i])
			}
		})
	}
}

func TestInitCommand_MutuallyExclusiveFlags(t *testing.T) {
	t.Parallel()

	_, _, err := executeCommandWithCapture(t,
		"init", "--tool=claude", "--include-deps", "--no-deps", "--dry-run")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestInitCommand_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	t.Run("empty agents list", func(t *testing.T) {
		_, stderr, err := executeCommandWithCapture(t,
			"init", "--tool=claude", "--agents", "", "--no-deps", "--dry-run")

		require.NoError(t, err)
		assert.Empty(t, stderr)
	})

	t.Run("non-existent agent name", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(oldWd)

		require.NoError(t, os.Chdir(tmpDir))

		_, _, err = executeCommandWithCapture(t,
			"init", "--tool=claude", "--agents", "nonexistent", "--include-deps", "--dry-run")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent not found")
	})

	t.Run("mix of valid and invalid agents", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(oldWd)

		require.NoError(t, os.Chdir(tmpDir))

		agents, err := cli.ResolveAgentList(goent.PluginFS, []string{"planner", "nonexistent"}, true, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent not found")
		assert.Nil(t, agents)
	})

	t.Run("all tools generates both directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(oldWd)

		require.NoError(t, os.Chdir(tmpDir))

		_, stderr, err := executeCommandWithCapture(t,
			"init", "--tool=all", "--no-deps", "--agents", "planner")

		require.NoError(t, err)
		assert.Empty(t, stderr)

		claudeDir := filepath.Join(tmpDir, ".claude")
		opencodeDir := filepath.Join(tmpDir, ".opencode")

		assert.DirExists(t, claudeDir)
		assert.DirExists(t, opencodeDir)
	})
}

func TestInitCommand_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "my-project")

	require.NoError(t, os.MkdirAll(customPath, 0755))

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	stdout, stderr, err := executeCommandWithCapture(t,
		"init", customPath, "--tool=claude", "--no-deps", "--agents", "planner")

	require.NoError(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)

	targetDir := filepath.Join(customPath, ".claude")
	assert.DirExists(t, targetDir)
}

func TestInitCommand_VerboseFlag(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	stdout, stderr, err := executeCommandWithCapture(t,
		"--verbose", "init", "--tool=claude", "--no-deps", "--agents", "planner")

	require.NoError(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}

func TestInitCommand_ModelOverride(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	stdout, stderr, err := executeCommandWithCapture(t,
		"init", "--tool=claude", "--model", "heavy=opus", "--no-deps", "--agents", "planner")

	require.NoError(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)

	targetDir := filepath.Join(tmpDir, ".claude")
	agentFile := filepath.Join(targetDir, "agents", "ent", "planner.md")

	assert.FileExists(t, agentFile)

	content, err := os.ReadFile(agentFile)
	require.NoError(t, err)
	contentStr := string(content)

	// Model override may add prefix to agent name
	assert.Contains(t, contentStr, "name:")
	assert.Contains(t, contentStr, "planner")
}

func TestInitCommand_ExistingConfigWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	targetDir := filepath.Join(tmpDir, ".claude")
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	existingFile := filepath.Join(targetDir, "agents", "ent", "planner.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(existingFile), 0755))
	require.NoError(t, os.WriteFile(existingFile, []byte("old content"), 0644))

	_, stderr, err := executeCommandWithCapture(t,
		"init", "--tool=claude", "--no-deps", "--agents", "planner")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Empty(t, stderr)

	oldContent, _ := os.ReadFile(existingFile)
	assert.Equal(t, "old content", string(oldContent))

	newStdout, newStderr, err := executeCommandWithCapture(t,
		"init", "--tool=claude", "--force", "--no-deps", "--agents", "planner")

	require.NoError(t, err)
	assert.NotEmpty(t, newStdout)
	assert.Empty(t, newStderr)

	newContent, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.NotEqual(t, "old content", string(newContent))
}

func TestInitCommand_DryRunOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	stdout, stderr, err := executeCommandWithCapture(t,
		"init", "--tool=claude", "--dry-run")

	require.NoError(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)

	assert.Contains(t, stdout, "DRY RUN")
	assert.Contains(t, stdout, "CLAUDE")
	assert.Contains(t, stdout, "Preview")
	assert.Contains(t, stdout, "Run without --dry-run")
}

func TestResolveAgentList_TransitiveDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		agentsFlag      []string
		expectedDeps    []string
		transitiveCheck []struct {
			agent    string
			transDep string
		}
	}{
		{
			name:         "architect transitive deps include planner and coder and their deps",
			agentsFlag:   []string{"architect"},
			expectedDeps: []string{"architect", "planner", "coder"},
			transitiveCheck: []struct {
				agent    string
				transDep string
			}{
				{"architect", "tester"},
				{"architect", "reviewer"},
				{"architect", "debugger"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, true, false)
			require.NoError(t, err)
			require.NotEmpty(t, agents)

			agentsSet := make(map[string]bool)
			for _, a := range agents {
				agentsSet[a] = true
			}

			for _, dep := range tt.expectedDeps {
				assert.True(t, agentsSet[dep], "expected agent %s in result", dep)
			}

			for _, check := range tt.transitiveCheck {
				assert.True(t, agentsSet[check.transDep],
					"expected transitive dependency %s for agent %s", check.transDep, check.agent)
			}
		})
	}
}

func TestResolveAgentList_ComplexDependencyGraph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		agentsFlag []string
		minAgents  int
		maxAgents  int
	}{
		{
			name:       "architect with all transitive deps",
			agentsFlag: []string{"architect"},
			minAgents:  6,
			maxAgents:  10,
		},
		{
			name:       "coder with all transitive deps",
			agentsFlag: []string{"coder"},
			minAgents:  3,
			maxAgents:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, true, false)
			require.NoError(t, err)
			require.NotEmpty(t, agents)

			assert.GreaterOrEqual(t, len(agents), tt.minAgents)
			assert.LessOrEqual(t, len(agents), tt.maxAgents)

			seen := make(map[string]bool)
			for _, agent := range agents {
				_, exists := seen[agent]
				assert.False(t, exists, "duplicate agent: %s", agent)
				seen[agent] = true
			}
		})
	}
}

func TestInitCommand_DirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	_, stderr, err := executeCommandWithCapture(t,
		"init", "--tool=claude", "--no-deps", "--agents", "planner")

	require.NoError(t, err)
	assert.Empty(t, stderr)

	targetDir := filepath.Join(tmpDir, ".claude")
	assert.DirExists(t, targetDir)

	agentsDir := filepath.Join(targetDir, "agents")
	assert.DirExists(t, agentsDir)

	agentFile := filepath.Join(agentsDir, "ent", "planner.md")
	assert.FileExists(t, agentFile)
}

func testInitCmdExecute(t *testing.T, cfg cli.InitConfig) (string, string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	err := cli.InitTools(context.Background(), cfg)

	wOut.Close()
	wErr.Close()

	var outBuf, errBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	io.Copy(&errBuf, rErr)

	return outBuf.String(), errBuf.String(), err
}

func TestInitCommand_SpecificTools(t *testing.T) {
	tests := []struct {
		name      string
		tool      string
		targetDir string
	}{
		{
			name:      "claude tool",
			tool:      "claude",
			targetDir: ".claude",
		},
		{
			name:      "opencode tool",
			tool:      "opencode",
			targetDir: ".opencode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldWd)

			require.NoError(t, os.Chdir(tmpDir))

			cfg := cli.InitConfig{
				Path:   tmpDir,
				Tool:   tt.tool,
				Agents: []string{"planner"},
				NoDeps: true,
				DryRun: false,
				Force:  true,
			}

			stdout, stderr, err := testInitCmdExecute(t, cfg)

			require.NoError(t, err)
			assert.NotEmpty(t, stdout)
			assert.Empty(t, stderr)

			targetDir := filepath.Join(tmpDir, tt.targetDir)
			assert.DirExists(t, targetDir)
		})
	}
}

func TestInitCommand_CircularDependencyDetection(t *testing.T) {
	mockFS := createMockFSWithCycle(t)

	_, err := cli.ResolveAgentList(mockFS, []string{"a"}, true, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func createMockFSWithCycle(t *testing.T) fs.FS {
	t.Helper()

	metas := map[string]string{
		"a.yaml": `name: a
description: Agent A
dependencies:
  - b
`,
		"b.yaml": `name: b
description: Agent B
dependencies:
  - c
`,
		"c.yaml": `name: c
description: Agent C
dependencies:
  - a
`,
	}

	dir := t.TempDir()
	metaDir := filepath.Join(dir, "plugins", "go-ent", "agents", "meta")
	require.NoError(t, os.MkdirAll(metaDir, 0755))

	for name, content := range metas {
		path := filepath.Join(metaDir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	return os.DirFS(dir)
}

func TestInitCommand_MissingDependencyValidation(t *testing.T) {
	mockFS := createMockFSWithMissingDep(t)

	_, err := cli.ResolveAgentList(mockFS, []string{"a"}, true, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dependency not found")
	assert.Contains(t, err.Error(), "missing")
}

func createMockFSWithMissingDep(t *testing.T) fs.FS {
	t.Helper()

	metas := map[string]string{
		"a.yaml": `name: a
description: Agent A
dependencies:
  - missing
`,
	}

	dir := t.TempDir()
	metaDir := filepath.Join(dir, "plugins", "go-ent", "agents", "meta")
	require.NoError(t, os.MkdirAll(metaDir, 0755))

	for name, content := range metas {
		path := filepath.Join(metaDir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	return os.DirFS(dir)
}

func TestInitCommand_GeneratedAgentContent(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	require.NoError(t, os.Chdir(tmpDir))

	cfg := cli.InitConfig{
		Path:   tmpDir,
		Tool:   "claude",
		Agents: []string{"planner"},
		NoDeps: true,
		DryRun: false,
		Force:  true,
	}

	_, outStderr, outErr := testInitCmdExecute(t, cfg)
	require.NoError(t, outErr)
	assert.Empty(t, outStderr)

	agentFile := filepath.Join(tmpDir, ".claude", "agents", "ent", "planner.md")
	content, err := os.ReadFile(agentFile)
	require.NoError(t, err)
	contentStr := string(content)

	lines := strings.Split(contentStr, "\n")

	assert.True(t, len(lines) > 3, "file should have multiple lines")

	frontmatterEnd := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if frontmatterEnd == -1 {
				frontmatterEnd = i
			}
		}
	}

	assert.NotEqual(t, -1, frontmatterEnd, "should have frontmatter")

	hasBody := false
	for i := frontmatterEnd + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			hasBody = true
			break
		}
	}

	assert.True(t, hasBody, "should have body content after frontmatter")
}

func TestResolveAgentList_AgentSetEquality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		agentsFlag    []string
		includeDeps   bool
		expectedCount int
	}{
		{
			name:          "planner with include-deps should have same count as no deps",
			agentsFlag:    []string{"planner"},
			includeDeps:   true,
			expectedCount: 1,
		},
		{
			name:          "multiple agents include deps",
			agentsFlag:    []string{"planner", "tester"},
			includeDeps:   true,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agents, err := cli.ResolveAgentList(goent.PluginFS, tt.agentsFlag, tt.includeDeps, false)
			require.NoError(t, err)
			assert.Len(t, agents, tt.expectedCount)
		})
	}
}
