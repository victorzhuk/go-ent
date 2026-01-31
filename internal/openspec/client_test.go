package openspec

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_List(t *testing.T) {
	t.Parallel()

	client := New("../..")
	ctx := context.Background()

	// Test listing changes
	data, err := client.List(ctx, "changes")
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Parse the result
	items, err := ParseList(data)
	require.NoError(t, err)
	assert.NotNil(t, items)
}

func TestClient_Validate(t *testing.T) {
	t.Parallel()

	client := New("../..")
	ctx := context.Background()

	data, err := client.Validate(ctx, "all")
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Parse the result
	result, err := ParseValidate(data)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
