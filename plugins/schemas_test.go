package schemas_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentSchema(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join("..", "plugins", "go-ent", "schemas", "agent.schema.json")

	schemaData, err := os.ReadFile(schemaPath) // #nosec G304 -- test file
	require.NoError(t, err)

	var schema map[string]any
	err = json.Unmarshal(schemaData, &schema)
	require.NoError(t, err)

	assert.Equal(t, "Claude Code Agent Schema", schema["title"])
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema["$schema"])

	required, ok := schema["required"].([]any)
	require.True(t, ok)
	assert.Contains(t, required, "name")
	assert.Contains(t, required, "description")

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	modelProp, ok := props["model"].(map[string]any)
	require.True(t, ok)
	modelEnum, ok := modelProp["enum"].([]any)
	require.True(t, ok)
	assert.Contains(t, modelEnum, "sonnet")
	assert.Contains(t, modelEnum, "haiku")
	assert.Contains(t, modelEnum, "opus")
	assert.Contains(t, modelEnum, "inherit")
}

func TestAgentSchemaStructure(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join("..", "plugins", "go-ent", "schemas", "agent.schema.json")

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
	assert.Equal(t, "^[a-z][a-z0-9-]*$", nameProps["pattern"])

	skills, ok := props["skills"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", skills["type"])
	assert.True(t, skills["uniqueItems"].(bool))

	color, ok := props["color"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "^#[0-9A-Fa-f]{6}$", color["pattern"])

	role, ok := props["role"].(map[string]any)
	require.True(t, ok)
	roleEnum, ok := role["enum"].([]any)
	require.True(t, ok)
	assert.Contains(t, roleEnum, "planning")
	assert.Contains(t, roleEnum, "execution")

	complexity, ok := props["complexity"].(map[string]any)
	require.True(t, ok)
	complexityEnum, ok := complexity["enum"].([]any)
	require.True(t, ok)
	assert.Contains(t, complexityEnum, "light")
	assert.Contains(t, complexityEnum, "standard")
	assert.Contains(t, complexityEnum, "heavy")
}
