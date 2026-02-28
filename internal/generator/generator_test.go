package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/config"
)

type mockTarget struct {
	name         string
	runtime      string
	outputPath   string
	generateErr  error
	generateFunc func(*AgentSource, *PromptContent) ([]byte, error)
}

func (m *mockTarget) Name() string                                     { return m.name }
func (m *mockTarget) Runtime() string                                  { return m.runtime }
func (m *mockTarget) OutputPath(agentName string) string               { return m.outputPath }
func (m *mockTarget) SkillOutputPath(category, name string) string     { return "" }
func (m *mockTarget) GenerateSkill(skill *SkillSource) ([]byte, error) { return nil, nil }
func (m *mockTarget) Generate(agent *AgentSource, prompts *PromptContent) ([]byte, error) {
	if m.generateFunc != nil {
		return m.generateFunc(agent, prompts)
	}
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return []byte("mock output"), nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		srcDir  string
		cfg     *config.ToolRuntimeConfig
		targets []Target
	}{
		{
			name:    "creates generator with all fields",
			srcDir:  "/src/agents",
			cfg:     config.DefaultToolRuntimeConfig(),
			targets: []Target{NewClaudeTarget("/output")},
		},
		{
			name:    "creates generator with nil config",
			srcDir:  "/src/agents",
			cfg:     nil,
			targets: []Target{NewClaudeTarget("/output")},
		},
		{
			name:    "creates generator with empty targets",
			srcDir:  "/src/agents",
			cfg:     config.DefaultToolRuntimeConfig(),
			targets: []Target{},
		},
		{
			name:    "creates generator with multiple targets",
			srcDir:  "/src/agents",
			cfg:     config.DefaultToolRuntimeConfig(),
			targets: []Target{NewClaudeTarget("/claude"), NewOpenCodeTarget("/opencode")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := New(tt.srcDir, tt.cfg, tt.targets...)

			assert.Equal(t, tt.srcDir, g.SrcDir)
			assert.Equal(t, tt.cfg, g.Config)
			assert.Len(t, g.Targets, len(tt.targets))
		})
	}
}

func TestWrapWithFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		frontmatter  []byte
		content      string
		wantContains []string
	}{
		{
			name:        "wraps simple content",
			frontmatter: []byte("name: test\nmodel: sonnet\n"),
			content:     "This is the agent prompt.",
			wantContains: []string{
				"---\n",
				"name: test",
				"model: sonnet",
				"This is the agent prompt.",
			},
		},
		{
			name:        "wraps content with special characters",
			frontmatter: []byte("name: test-agent\ndescription: \"A test\"\n"),
			content:     "Content with\nmultiple lines\nand special chars: @#$%",
			wantContains: []string{
				"---\n",
				"name: test-agent",
				"description: \"A test\"",
				"Content with",
				"multiple lines",
			},
		},
		{
			name:        "handles empty content",
			frontmatter: []byte("name: empty\n"),
			content:     "",
			wantContains: []string{
				"---\n",
				"name: empty",
			},
		},
		{
			name:        "handles empty frontmatter",
			frontmatter: []byte(""),
			content:     "Just content",
			wantContains: []string{
				"---\n",
				"Just content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := wrapWithFrontmatter(tt.frontmatter, tt.content)

			for _, want := range tt.wantContains {
				assert.Contains(t, string(result), want)
			}

			assert.True(t, bytes.HasPrefix(result, []byte("---\n")))
			assert.Contains(t, string(result), "---\n\n")
		})
	}
}

func TestGenerator_WriteOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "writes to new file",
			path:    "output/test.md",
			data:    []byte("# Test Content"),
			wantErr: false,
		},
		{
			name:    "writes to nested directory",
			path:    "output/nested/deep/test.md",
			data:    []byte("nested content"),
			wantErr: false,
		},
		{
			name:    "overwrites existing file",
			path:    "output/existing.md",
			data:    []byte("new content"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			fullPath := filepath.Join(tmpDir, tt.path)

			g := &Generator{}
			err := g.writeOutput(fullPath, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			data, err := os.ReadFile(fullPath)
			require.NoError(t, err)
			assert.Equal(t, tt.data, data)
		})
	}
}

func TestGenerator_WriteOutput_CreatesDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "deeply", "nested", "dir", "test.md")

	g := &Generator{}
	err := g.writeOutput(nestedPath, []byte("content"))

	require.NoError(t, err)
	assert.FileExists(t, nestedPath)
}

func TestGenerator_WriteOutput_Permissions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")

	g := &Generator{}
	err := g.writeOutput(filePath, []byte("content"))

	require.NoError(t, err)

	info, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestGenerator_GenerateAgent(t *testing.T) {
	t.Parallel()

	t.Run("generates agent with single target", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "output")

		// Create agent source file in agents/meta directory
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))
		agentYAML := `name: test-agent
description: Test agent
model: main
prompts:
  main: test-agent
`
		require.NoError(t, os.WriteFile(filepath.Join(agentsMetaDir, "test-agent.yaml"), []byte(agentYAML), 0o600))

		// Create prompt file at agents/prompts/agents/ (LoadAgentMetaSource goes up from meta dir)
		promptsDir := filepath.Join(tmpDir, "agents", "prompts", "agents")
		require.NoError(t, os.MkdirAll(promptsDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "test-agent.md"), []byte("Test prompt content."), 0o600))

		mock := &mockTarget{
			name:       "mock",
			runtime:    "mock",
			outputPath: filepath.Join(outputDir, "test-agent.md"),
		}

		g := New(agentsMetaDir, nil, mock)
		err := g.GenerateAgent("test-agent")

		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(outputDir, "test-agent.md"))
	})

	t.Run("generates agent with multiple targets", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))

		agentYAML := `name: multi-agent
description: Multi target agent
model: main
prompts:
  main: multi
`
		require.NoError(t, os.WriteFile(filepath.Join(agentsMetaDir, "multi-agent.yaml"), []byte(agentYAML), 0o600))

		promptsDir := filepath.Join(tmpDir, "agents", "prompts", "agents")
		require.NoError(t, os.MkdirAll(promptsDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "multi.md"), []byte("Multi prompt."), 0o600))

		outputDir1 := filepath.Join(tmpDir, "output1")
		outputDir2 := filepath.Join(tmpDir, "output2")

		g := New(agentsMetaDir, nil,
			&mockTarget{name: "target1", runtime: "mock", outputPath: filepath.Join(outputDir1, "multi-agent.md")},
			&mockTarget{name: "target2", runtime: "mock", outputPath: filepath.Join(outputDir2, "multi-agent.md")},
		)

		err := g.GenerateAgent("multi-agent")
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(outputDir1, "multi-agent.md"))
		assert.FileExists(t, filepath.Join(outputDir2, "multi-agent.md"))
	})

	t.Run("returns error for missing agent", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))

		g := New(agentsMetaDir, nil, &mockTarget{name: "mock", runtime: "mock", outputPath: "output.md"})
		err := g.GenerateAgent("nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load source")
	})
}

func TestGenerator_GenerateAll(t *testing.T) {
	t.Parallel()

	t.Run("generates all agents", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))

		promptsDir := filepath.Join(tmpDir, "agents", "prompts", "agents")
		require.NoError(t, os.MkdirAll(promptsDir, 0o750))

		// Create multiple agent files
		for _, name := range []string{"agent1", "agent2"} {
			agentYAML := `name: ` + name + `
description: ` + name + ` description
model: main
prompts:
  main: ` + name + `
`
			require.NoError(t, os.WriteFile(filepath.Join(agentsMetaDir, name+".yaml"), []byte(agentYAML), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(promptsDir, name+".md"), []byte(name+" prompt."), 0o600))
		}

		outputDir := filepath.Join(tmpDir, "output")
		mock := &mockTarget{
			name:       "mock",
			runtime:    "mock",
			outputPath: filepath.Join(outputDir, "agent.md"),
			generateFunc: func(a *AgentSource, p *PromptContent) ([]byte, error) {
				return []byte(a.Name), nil
			},
		}

		g := New(agentsMetaDir, nil, mock)
		err := g.GenerateAll()

		require.NoError(t, err)
	})

	t.Run("handles empty agents directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))

		g := New(agentsMetaDir, nil, &mockTarget{name: "mock", runtime: "mock", outputPath: "output.md"})
		err := g.GenerateAll()

		require.NoError(t, err)
	})
}

func TestGenerator_GenerateAgent_WithConfig(t *testing.T) {
	t.Parallel()

	t.Run("resolves claude model with config", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))

		agentYAML := `name: config-agent
description: Agent with config
model: main
prompts:
  main: config
`
		require.NoError(t, os.WriteFile(filepath.Join(agentsMetaDir, "config-agent.yaml"), []byte(agentYAML), 0o600))

		promptsDir := filepath.Join(tmpDir, "agents", "prompts", "agents")
		require.NoError(t, os.MkdirAll(promptsDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "config.md"), []byte("Config prompt."), 0o600))

		outputDir := filepath.Join(tmpDir, "output")

		var receivedModel string
		mock := &mockTarget{
			name:       "claude",
			runtime:    "claude",
			outputPath: filepath.Join(outputDir, "config-agent.md"),
			generateFunc: func(a *AgentSource, p *PromptContent) ([]byte, error) {
				receivedModel = a.Model.Claude
				return []byte("output"), nil
			},
		}

		cfg := config.DefaultToolRuntimeConfig()
		g := New(agentsMetaDir, cfg, mock)
		err := g.GenerateAgent("config-agent")

		require.NoError(t, err)
		assert.Equal(t, "claude-sonnet-4-5-20250929", receivedModel)
	})

	t.Run("resolves opencode model with config", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		agentsMetaDir := filepath.Join(tmpDir, "agents", "meta")
		require.NoError(t, os.MkdirAll(agentsMetaDir, 0o750))

		agentYAML := `name: opencode-agent
description: OpenCode agent
model: fast
prompts:
  main: opencode
`
		require.NoError(t, os.WriteFile(filepath.Join(agentsMetaDir, "opencode-agent.yaml"), []byte(agentYAML), 0o600))

		promptsDir := filepath.Join(tmpDir, "agents", "prompts", "agents")
		require.NoError(t, os.MkdirAll(promptsDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(promptsDir, "opencode.md"), []byte("OpenCode prompt."), 0o600))

		outputDir := filepath.Join(tmpDir, "output")

		var receivedModel string
		mock := &mockTarget{
			name:       "opencode",
			runtime:    "opencode",
			outputPath: filepath.Join(outputDir, "opencode-agent.md"),
			generateFunc: func(a *AgentSource, p *PromptContent) ([]byte, error) {
				receivedModel = a.Model.OpenCode
				return []byte("output"), nil
			},
		}

		cfg := config.DefaultToolRuntimeConfig()
		g := New(agentsMetaDir, cfg, mock)
		err := g.GenerateAgent("opencode-agent")

		require.NoError(t, err)
		assert.Equal(t, "zai-coding-plan/glm-4.7-flash", receivedModel)
	})
}
