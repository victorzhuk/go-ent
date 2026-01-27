package cli

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootpkg "github.com/victorzhuk/go-ent"
)

func TestGetAgentPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tool     string
		prefix   string
		agent    string
		expected string
	}{
		{
			name:     "Claude Code with default prefix",
			tool:     "claude",
			prefix:   "ent",
			agent:    "coder",
			expected: filepath.Join(".claude", "agents", "ent", "coder.md"),
		},
		{
			name:     "Claude Code with custom prefix",
			tool:     "claude",
			prefix:   "myproject",
			agent:    "architect",
			expected: filepath.Join(".claude", "agents", "myproject", "architect.md"),
		},
		{
			name:     "OpenCode with default prefix",
			tool:     "opencode",
			prefix:   "ent",
			agent:    "coder",
			expected: filepath.Join(".opencode", "agent", "coder.md"),
		},
		{
			name:     "OpenCode with custom prefix (no effect)",
			tool:     "opencode",
			prefix:   "myproject",
			agent:    "architect",
			expected: filepath.Join(".opencode", "agent", "architect.md"),
		},
		{
			name:     "Unsupported tool returns empty",
			tool:     "cursor",
			prefix:   "ent",
			agent:    "coder",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getAgentPath(tt.tool, tt.prefix, tt.agent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadAgents(t *testing.T) {
	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	agents, err := loadAgents()
	require.NoError(t, err)
	assert.Greater(t, len(agents), 0)

	coder, ok := agents["coder"]
	require.True(t, ok, "coder agent should exist")
	assert.Equal(t, "coder", coder.Name)
	assert.Equal(t, "Go developer. Implements features, writes code.", coder.Description)
	assert.Equal(t, "execution", coder.Role)
	assert.Equal(t, "standard", coder.Complexity)
	assert.Contains(t, coder.Skills, "go-code")
	assert.Contains(t, coder.Tools, "Read")
}

func TestLoadPrompts(t *testing.T) {
	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	prompts, err := loadPrompts()
	require.NoError(t, err)
	assert.Greater(t, len(prompts), 0)

	coderPrompt, ok := prompts["coder"]
	require.True(t, ok, "coder prompt should exist")
	assert.Contains(t, coderPrompt, "You are a senior Go backend developer")
	assert.Contains(t, coderPrompt, "Implement features from tasks.md")
}

func TestLoadShared(t *testing.T) {
	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	shared, err := loadShared()
	require.NoError(t, err)
	assert.NotEmpty(t, shared)

	expectedContent := []string{
		"# Principal Hierarchy for Constitutional AI",
		"# Judgment Guidance for Constitutional AI",
		"# OpenSpec Workflow",
		"# Code Conventions",
		"# Agent Handoffs",
		"# Tooling Reference",
	}

	for _, content := range expectedContent {
		assert.Contains(t, shared, content, "shared content should include section: %s", content)
	}
}

func TestLoadTemplate(t *testing.T) {
	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	t.Run("Claude template", func(t *testing.T) {

		tpl, err := loadTemplate("claude")
		require.NoError(t, err)
		assert.NotNil(t, tpl)
	})

	t.Run("OpenCode template", func(t *testing.T) {

		tpl, err := loadTemplate("opencode")
		require.NoError(t, err)
		assert.NotNil(t, tpl)
	})

	t.Run("Unsupported tool", func(t *testing.T) {

		tpl, err := loadTemplate("cursor")
		assert.Error(t, err)
		assert.Nil(t, tpl)
		assert.Contains(t, err.Error(), "unsupported tool")
	})
}

func TestRenderAgent(t *testing.T) {
	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	meta := &agentMeta{
		Name:        "test",
		Description: "Test agent",
		Model:       "main",
		Color:       "#FF0000",
		Role:        "execution",
		Complexity:  "standard",
		Skills:      []string{"go-code"},
		Tools:       []string{"Read", "Write"},
	}

	tpl, err := loadTemplate("claude")
	require.NoError(t, err)

	prompt := "Test prompt content\n"
	shared := "Shared content\n"

	result, err := renderAgent(meta, prompt, shared, tpl)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.Contains(t, result, "name: test")
	assert.Contains(t, result, "description: \"Test agent\"")
	assert.Contains(t, result, "model: main")
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "Shared content")
	assert.Contains(t, result, "Test prompt content")
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Create new file", func(t *testing.T) {

		testPath := filepath.Join(tmpDir, "new", "file.md")
		content := "Test content"

		err := writeFile(testPath, content, false, false)
		require.NoError(t, err)

		data, err := os.ReadFile(testPath)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("Overwrite existing file with force", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "force", "file.md")
		err := os.MkdirAll(filepath.Dir(testPath), 0750)
		require.NoError(t, err)
		err = os.WriteFile(testPath, []byte("old"), 0600)
		require.NoError(t, err)

		err = writeFile(testPath, "new", true, false)
		require.NoError(t, err)

		data, err := os.ReadFile(testPath)
		require.NoError(t, err)
		assert.Equal(t, "new", string(data))
	})

	t.Run("Error on existing file without force", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "error", "file.md")
		err := os.MkdirAll(filepath.Dir(testPath), 0750)
		require.NoError(t, err)
		err = os.WriteFile(testPath, []byte("existing"), 0600)
		require.NoError(t, err)

		err = writeFile(testPath, "new", false, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file already exists")
	})

	t.Run("Dry run does not create file", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "dryrun", "file.md")
		content := "Test content"

		err := writeFile(testPath, content, false, true)
		require.NoError(t, err)

		_, err = os.Stat(testPath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestCopyCommands(t *testing.T) {
	t.Run("Copy commands to Claude Code", func(t *testing.T) {

		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		var fs embed.FS = rootpkg.PluginFS
		SetPluginFS(fs)

		err = copyCommands("claude", "ent", false, false)
		require.NoError(t, err)

		cmdDir := filepath.Join(tmpDir, ".claude", "commands", "ent")
		entries, err := os.ReadDir(cmdDir)
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0)

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				path := filepath.Join(cmdDir, entry.Name())
				data, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.NotEmpty(t, string(data))
			}
		}
	})

	t.Run("Copy commands to OpenCode", func(t *testing.T) {

		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		var fs embed.FS = rootpkg.PluginFS
		SetPluginFS(fs)

		err = copyCommands("opencode", "ent", false, false)
		require.NoError(t, err)

		cmdDir := filepath.Join(tmpDir, ".opencode", "commands", "ent")
		entries, err := os.ReadDir(cmdDir)
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0)
	})
}

func TestCopySkills(t *testing.T) {
	t.Run("Copy skills to Claude Code", func(t *testing.T) {

		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		var fs embed.FS = rootpkg.PluginFS
		SetPluginFS(fs)

		err = copySkills("claude", "ent", false, false)
		require.NoError(t, err)

		skillDir := filepath.Join(tmpDir, ".claude", "skills", "ent")

		var walkAndCount func(dir string) (int, error)
		walkAndCount = func(dir string) (int, error) {
			count := 0
			entries, err := os.ReadDir(dir)
			if err != nil {
				return 0, err
			}

			for _, entry := range entries {
				path := filepath.Join(dir, entry.Name())
				if entry.IsDir() {
					c, err := walkAndCount(path)
					if err != nil {
						return 0, err
					}
					count += c
					continue
				}
				if strings.HasSuffix(entry.Name(), ".md") {
					count++
					data, err := os.ReadFile(path)
					require.NoError(t, err)
					assert.NotEmpty(t, string(data))
				}
			}
			return count, nil
		}

		count, err := walkAndCount(skillDir)
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})
}

func TestInitCmd_ToolParsing(t *testing.T) {
	t.Run("Single tool - claude", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		var fs embed.FS = rootpkg.PluginFS
		SetPluginFS(fs)

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"init", "--tool=claude"})

		err = rootCmd.Execute()
		require.NoError(t, err)

		agentDir := filepath.Join(tmpDir, ".claude", "agents", "ent")
		_, err = os.Stat(agentDir)
		require.NoError(t, err)
	})

	t.Run("Single tool - opencode", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		var fs embed.FS = rootpkg.PluginFS
		SetPluginFS(fs)

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"init", "--tool=opencode"})

		err = rootCmd.Execute()
		require.NoError(t, err)

		agentDir := filepath.Join(tmpDir, ".opencode", "agent")
		_, err = os.Stat(agentDir)
		require.NoError(t, err)
	})

	t.Run("Multiple tools - claude,opencode", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		var fs embed.FS = rootpkg.PluginFS
		SetPluginFS(fs)

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"init", "--tool=claude,opencode"})

		err = rootCmd.Execute()
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, ".claude", "agents", "ent"))
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, ".opencode", "agent"))
		require.NoError(t, err)
	})

	t.Run("Missing --tool flag", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(oldWd) })

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"init"})

		err = rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool")
	})
}

func TestInitCmd_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"init", "--tool=claude", "--dry-run"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, ".claude"))
	assert.True(t, os.IsNotExist(err), "dry-run should not create files")
}

func TestInitCmd_Force(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	agentDir := filepath.Join(tmpDir, ".claude", "agents", "ent")
	err = os.MkdirAll(agentDir, 0750)
	require.NoError(t, err)

	existingFile := filepath.Join(agentDir, "coder.md")
	err = os.WriteFile(existingFile, []byte("old content"), 0600)
	require.NoError(t, err)

	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"init", "--tool=claude", "--force"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.NotEqual(t, "old content", string(data))
	assert.Contains(t, string(data), "name: coder")
}

func TestInitCmd_CustomPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"init", "--tool=claude", "--prefix=myproject"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	agentDir := filepath.Join(tmpDir, ".claude", "agents", "myproject")
	_, err = os.Stat(agentDir)
	require.NoError(t, err)

	cmdDir := filepath.Join(tmpDir, ".claude", "commands", "myproject")
	_, err = os.Stat(cmdDir)
	require.NoError(t, err)
}

func TestInitCmd_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var fs embed.FS = rootpkg.PluginFS
	SetPluginFS(fs)

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"init", "--tool=claude,opencode"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	agentDir := filepath.Join(tmpDir, ".claude", "agents", "ent")
	entries, err := os.ReadDir(agentDir)
	require.NoError(t, err)
	assert.Greater(t, len(entries), 0)

	coderPath := filepath.Join(agentDir, "coder.md")
	data, err := os.ReadFile(coderPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "name: coder")
	assert.Contains(t, content, "---")
	assert.Contains(t, content, "# Principal Hierarchy for Constitutional AI")

	cmdDir := filepath.Join(tmpDir, ".claude", "commands", "ent")
	cmdEntries, err := os.ReadDir(cmdDir)
	require.NoError(t, err)
	assert.Greater(t, len(cmdEntries), 0)

	skillDir := filepath.Join(tmpDir, ".claude", "skills", "ent")
	skillEntries, err := os.ReadDir(skillDir)
	require.NoError(t, err)
	assert.Greater(t, len(skillEntries), 0)

	opencodeAgentDir := filepath.Join(tmpDir, ".opencode", "agent")
	opencodeEntries, err := os.ReadDir(opencodeAgentDir)
	require.NoError(t, err)
	assert.Greater(t, len(opencodeEntries), 0)

	opencodeCoderPath := filepath.Join(opencodeAgentDir, "coder.md")
	opencodeData, err := os.ReadFile(opencodeCoderPath)
	require.NoError(t, err)
	opencodeContent := string(opencodeData)

	assert.Contains(t, opencodeContent, "name: coder")
	assert.Contains(t, opencodeContent, "---")
}
