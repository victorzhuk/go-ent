package schemas_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentMetaSchema(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join("..", "plugins", "go-ent", "schemas", "agent-meta.schema.json")

	schemaData, err := os.ReadFile(schemaPath) // #nosec G304 -- test file
	require.NoError(t, err)

	var schema map[string]any
	err = json.Unmarshal(schemaData, &schema)
	require.NoError(t, err)

	assert.Equal(t, "Agent Metadata Schema", schema["title"])
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema["$schema"])

	required, ok := schema["required"].([]any)
	require.True(t, ok)
	assert.Contains(t, required, "name")
	assert.Contains(t, required, "description")
	assert.Contains(t, required, "model")
	assert.Contains(t, required, "color")
	assert.Contains(t, required, "tags")
	assert.Contains(t, required, "prompts")

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	modelEnum, ok := props["model"].(map[string]any)["enum"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"main", "fast", "heavy"}, modelEnum)

	tagsProps, ok := props["tags"].(map[string]any)["properties"].(map[string]any)
	require.True(t, ok)

	roleEnum, ok := tagsProps["role"].(map[string]any)["enum"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"planning", "execution", "review", "debug", "test"}, roleEnum)

	complexityEnum, ok := tagsProps["complexity"].(map[string]any)["enum"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"light", "standard", "heavy"}, complexityEnum)

	colorPattern, ok := props["color"].(map[string]any)["pattern"].(string)
	require.True(t, ok)
	assert.Equal(t, "^#[0-9A-Fa-f]{6}$", colorPattern)
}

func TestAgentMetaSchemaStructure(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join("..", "plugins", "go-ent", "schemas", "agent-meta.schema.json")

	schemaData, err := os.ReadFile(schemaPath) // #nosec G304 -- test file
	require.NoError(t, err)

	var schema map[string]any
	err = json.Unmarshal(schemaData, &schema)
	require.NoError(t, err)

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	nameProps, ok := props["name"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", nameProps["type"])
	assert.Equal(t, "^#[0-9A-Fa-f]{6}$", props["color"].(map[string]any)["pattern"])

	tags, ok := props["tags"].(map[string]any)
	require.True(t, ok)
	assert.False(t, tags["additionalProperties"].(bool))

	prompts, ok := props["prompts"].(map[string]any)
	require.True(t, ok)
	assert.False(t, prompts["additionalProperties"].(bool))

	tools, ok := props["tools"].(map[string]any)
	require.True(t, ok)
	assert.True(t, tools["uniqueItems"].(bool))

	skills, ok := props["skills"].(map[string]any)
	require.True(t, ok)
	assert.True(t, skills["uniqueItems"].(bool))

	deps, ok := props["dependencies"].(map[string]any)
	require.True(t, ok)
	assert.True(t, deps["uniqueItems"].(bool))
}
