package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig_WithPrompts(t *testing.T) {
	content := `name: go-basic
category: go
description: Basic Go skill template
author: OpenSpec Team
version: 1.0.0
prompts:
  - key: skill_name
    prompt: "What is the skill name?"
    default: ""
    required: true
  - key: skill_description
    prompt: "What does this skill do?"
    default: "A new skill"
    required: true
  - key: skill_version
    prompt: "What version?"
    default: "1.0.0"
    required: false`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Equal(t, "go-basic", result.Name)
	assert.Equal(t, "go", result.Category)
	assert.Equal(t, "Basic Go skill template", result.Description)
	assert.Equal(t, "OpenSpec Team", result.Author)
	assert.Equal(t, "1.0.0", result.Version)

	assert.Len(t, result.Prompts, 3)
	assert.Equal(t, "skill_name", result.Prompts[0].Key)
	assert.Equal(t, "What is the skill name?", result.Prompts[0].Prompt)
	assert.Equal(t, "", result.Prompts[0].Default)
	assert.True(t, result.Prompts[0].Required)

	assert.Equal(t, "skill_description", result.Prompts[1].Key)
	assert.Equal(t, "What does this skill do?", result.Prompts[1].Prompt)
	assert.Equal(t, "A new skill", result.Prompts[1].Default)
	assert.True(t, result.Prompts[1].Required)

	assert.Equal(t, "skill_version", result.Prompts[2].Key)
	assert.Equal(t, "What version?", result.Prompts[2].Prompt)
	assert.Equal(t, "1.0.0", result.Prompts[2].Default)
	assert.False(t, result.Prompts[2].Required)
}

func TestParseConfig_EmptyPrompts(t *testing.T) {
	content := `name: go-basic
category: go
description: Basic Go skill template
prompts: []`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Empty(t, result.Prompts)
}

func TestParseConfig_MissingPrompts(t *testing.T) {
	content := `name: go-basic
category: go
description: Basic Go skill template`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Nil(t, result.Prompts)
}

func TestParseConfig_Validate_MissingName(t *testing.T) {
	content := `category: go
description: Basic Go skill template`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "required field 'name' is empty")
}

func TestParseConfig_Validate_MissingCategory(t *testing.T) {
	content := `name: go-basic
description: Basic Go skill template`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "required field 'category' is empty")
}

func TestParseConfig_Validate_EmptyName(t *testing.T) {
	content := `name: ""
category: go
description: Basic Go skill template`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "required field 'name' is empty")
}

func TestParseConfig_Validate_EmptyCategory(t *testing.T) {
	content := `name: go-basic
category: ""
description: Basic Go skill template`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "required field 'category' is empty")
}

func TestTemplateConfig_Validate_Valid(t *testing.T) {
	cfg := &TemplateConfig{
		Name:        "go-basic",
		Category:    "go",
		Description: "Basic Go skill template",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestTemplateConfig_Validate_OptionalFields(t *testing.T) {
	cfg := &TemplateConfig{
		Name:     "go-basic",
		Category: "go",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestParseConfig_NonExistentFile(t *testing.T) {
	nonExistentPath := "/tmp/non-existent-config-12345.yaml"

	result, err := ParseConfig(nonExistentPath)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "read config file")
}

func TestParseConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `name: go-basic
category: go
description: This is invalid
  indentation: bad
yaml: [unclosed`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(invalidYAML), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "parse yaml")
}

func TestParseConfig_MinimalConfig(t *testing.T) {
	content := `name: go-basic
category: go`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Equal(t, "go-basic", result.Name)
	assert.Equal(t, "go", result.Category)
	assert.Empty(t, result.Description)
	assert.Empty(t, result.Author)
	assert.Empty(t, result.Version)
	assert.Nil(t, result.Prompts)
}

func TestParseConfig_AllFields(t *testing.T) {
	content := `name: go-complete
category: go
description: Complete Go skill template
author: OpenSpec Team
version: 2.0.0
prompts:
  - key: skill_name
    prompt: "Skill name?"
    default: "my-skill"
    required: true
  - key: skill_description
    prompt: "Description?"
    default: ""
    required: false`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Equal(t, "go-complete", result.Name)
	assert.Equal(t, "go", result.Category)
	assert.Equal(t, "Complete Go skill template", result.Description)
	assert.Equal(t, "OpenSpec Team", result.Author)
	assert.Equal(t, "2.0.0", result.Version)
	assert.Len(t, result.Prompts, 2)
}

func TestTemplateConfig_Validate_MissingName(t *testing.T) {
	cfg := &TemplateConfig{
		Category:    "go",
		Description: "Basic Go skill template",
	}

	err := cfg.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required field 'name' is empty")
}

func TestTemplateConfig_Validate_MissingCategory(t *testing.T) {
	cfg := &TemplateConfig{
		Name:        "go-basic",
		Description: "Basic Go skill template",
	}

	err := cfg.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required field 'category' is empty")
}

func TestParseConfig_SinglePrompt(t *testing.T) {
	content := `name: go-basic
category: go
prompts:
  - key: skill_name
    prompt: "What is the skill name?"
    required: true`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Len(t, result.Prompts, 1)
	assert.Equal(t, "skill_name", result.Prompts[0].Key)
	assert.Equal(t, "What is the skill name?", result.Prompts[0].Prompt)
	assert.Empty(t, result.Prompts[0].Default)
	assert.True(t, result.Prompts[0].Required)
}

func TestParseConfig_MultiplePrompts(t *testing.T) {
	content := `name: go-complete
category: go
prompts:
  - key: prompt1
    prompt: "First prompt?"
    default: "default1"
    required: true
  - key: prompt2
    prompt: "Second prompt?"
    required: false
  - key: prompt3
    prompt: "Third prompt?"
    default: "default3"
    required: true
  - key: prompt4
    prompt: "Fourth prompt?"
    required: false`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseConfig(path)

	require.NoError(t, err)
	assert.Len(t, result.Prompts, 4)

	assert.Equal(t, "prompt1", result.Prompts[0].Key)
	assert.Equal(t, "default1", result.Prompts[0].Default)
	assert.True(t, result.Prompts[0].Required)

	assert.Equal(t, "prompt2", result.Prompts[1].Key)
	assert.Empty(t, result.Prompts[1].Default)
	assert.False(t, result.Prompts[1].Required)

	assert.Equal(t, "prompt3", result.Prompts[2].Key)
	assert.Equal(t, "default3", result.Prompts[2].Default)
	assert.True(t, result.Prompts[2].Required)

	assert.Equal(t, "prompt4", result.Prompts[3].Key)
	assert.Empty(t, result.Prompts[3].Default)
	assert.False(t, result.Prompts[3].Required)
}
